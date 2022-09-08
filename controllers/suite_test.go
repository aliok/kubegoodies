/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"fmt"
	kubegoodiesv1 "github.com/aliok/kubegoodies/api/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/e2e-framework/klient/conf"
	"sigs.k8s.io/e2e-framework/klient/k8s"
	"sigs.k8s.io/e2e-framework/klient/wait/conditions"
	"testing"

	"sigs.k8s.io/e2e-framework/klient/wait"
	"sigs.k8s.io/e2e-framework/pkg/env"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"
)

func Hello(name string) string {
	return fmt.Sprintf("Hello %s", name)
}

// TestHello shows an example of a test environment
// that uses a simple setup to assess a feature (test)
// in a test function directly (outside of test suite TestMain)
func TestHello(t *testing.T) {
	e := env.NewWithConfig(envconf.New())
	feat := features.New("Hello Feature").
		WithLabel("type", "simple").
		Assess("test message", func(ctx context.Context, t *testing.T, _ *envconf.Config) context.Context {
			result := Hello("foo")
			if result != "Hello foo" {
				t.Error("unexpected message")
			}
			return ctx
		})

	e.Test(t, feat.Feature())
}

// The following shows an example of a simple
// test function that uses feature with a setup
// step.
func TestHello_WithSetup(t *testing.T) {
	kubeconfigpath := conf.ResolveKubeConfigFile()
	cfg := envconf.NewWithKubeConfig(kubeconfigpath)
	e := env.NewWithConfig(cfg)
	feat := features.New("Hello Feature").
		// TODO: need the label?
		WithLabel("type", "simple").
		Setup(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			err := kubegoodiesv1.AddToScheme(cfg.Client().Resources().GetScheme())
			if err != nil {
				t.Fatal(err)
			}
			return ctx
		}).
		Setup(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			propagation := kubegoodiesv1.ConfigMapPropagation{
				//TypeMeta: metav1.TypeMeta{
				//	// TODO: are these needed?
				//	//Kind:       "",
				//	//APIVersion: "",
				//},
				ObjectMeta: metav1.ObjectMeta{
					Name: "test", // TODO: randomize name
				},
				Spec: kubegoodiesv1.ConfigMapPropagationSpec{
					Source: kubegoodiesv1.PropagationSource{
						Namespace: "default",
						Names:     []string{"src-by-name-1"},
					},
					Target: kubegoodiesv1.PropagationTarget{
						Namespaces: []string{"ns1"},
					},
				},
			}
			if err := cfg.Client().Resources().Create(ctx, &propagation); err != nil {
				t.Fatalf("failed to create ConfigMapPropagation: %v", err)
			}

			return ctx
		}).
		Assess("propagation ready", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			propagation := kubegoodiesv1.ConfigMapPropagation{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test", // TODO: randomize name
				},
			}

			// TODO: maybe override the default timeout of 5 mins with something shorter
			err := wait.For(conditions.New(cfg.Client().Resources()).ResourceMatch(&propagation, func(obj k8s.Object) bool {
				prop := obj.(*kubegoodiesv1.ConfigMapPropagation)
				for _, cond := range prop.Status.Conditions {
					if cond.Type == kubegoodiesv1.ConfigMapPropagationConditionTypeReady {
						return cond.Status == metav1.ConditionTrue
					}
				}
				return false
			}))
			if err != nil {
				t.Errorf("propagation is not ready: %v", err)
			}

			return ctx
		}).Feature()

	e.Test(t, feat)
}
