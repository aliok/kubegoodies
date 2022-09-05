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
	"github.com/aliok/kubegoodies/pkg/configmappropagation"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	corev1 "k8s.io/api/core/v1"

	kubegoodiesv1 "github.com/aliok/kubegoodies/api/v1"
)

// ConfigMapPropagationReconciler reconciles a ConfigMapPropagation object
type ConfigMapPropagationReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=kubegoodies.aliok.github.com,resources=configmappropagations,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=kubegoodies.aliok.github.com,resources=configmappropagations/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=kubegoodies.aliok.github.com,resources=configmappropagations/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the ConfigMapPropagation object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.11.2/pkg/reconcile
func (r *ConfigMapPropagationReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	var pr kubegoodiesv1.ConfigMapPropagation
	if err := r.Get(ctx, req.NamespacedName, &pr); err != nil {
		// we'll ignore not-found errors, since they can't be fixed by an immediate
		// requeue (we'll need to wait for a new notification), and we can get them
		// on deleted requests.
		logger.Error(err, "unable to fetch ConfigMapPropagation")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	var executionReqs []configmappropagation.Request
	if len(pr.Spec.Source.Names) > 0 {
		for _, srcConfigmapName := range pr.Spec.Source.Names {
			for _, targetNs := range pr.Spec.Target.Namespaces {
				executionReqs = append(executionReqs, configmappropagation.Request{
					SourceNamespace: pr.Spec.Source.Namespace,
					SourceName:      srcConfigmapName,
					TargetNamespace: targetNs,
					TargetName:      srcConfigmapName,
				})
			}
		}
	}

	if pr.Spec.Source.ObjectSelector != nil {
		var srcConfigmapList corev1.ConfigMapList
		if err := r.List(ctx, &srcConfigmapList, client.InNamespace(pr.Spec.Source.Namespace), client.MatchingLabels(pr.Spec.Source.ObjectSelector.MatchLabels)); err != nil {
			logger.Error(err, "unable to list ConfigMaps")
			return ctrl.Result{}, err
		}

		for _, srcConfigmap := range srcConfigmapList.Items {
			for _, targetNs := range pr.Spec.Target.Namespaces {
				executionReqs = append(executionReqs, configmappropagation.Request{
					SourceNamespace: pr.Spec.Source.Namespace,
					SourceName:      srcConfigmap.Name,
					TargetNamespace: targetNs,
					TargetName:      srcConfigmap.Name,
				})
			}
		}
	}

	// TODO: do we need a sanity check for the case where there is a target configmap which is targeted by multiple source configmaps?

	logger.Info("executionReqs", "executionReqs", executionReqs)

	for _, executionReq := range executionReqs {
		if err := configmappropagation.Execute(ctx, r.Client, &executionReq); err != nil {
			// TODO: do not return here, but continue with the rest of the requests
			logger.Error(err, "unable to execute configmap propagation")
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ConfigMapPropagationReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&kubegoodiesv1.ConfigMapPropagation{}).
		Complete(r)
}
