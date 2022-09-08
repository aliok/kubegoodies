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

	"github.com/aliok/kubegoodies/pkg/configmappropagation"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	corev1 "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	kubegoodiesv1 "github.com/aliok/kubegoodies/api/v1"

	"github.com/hashicorp/go-multierror"
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
	meta.SetStatusCondition(&pr.Status.Conditions, metav1.Condition{
		Type:    kubegoodiesv1.ConfigMapPropagationConditionTypeCollectedExecutionRequests,
		Status:  metav1.ConditionTrue,
		Reason:  "CollectedExecutionRequests",
		Message: fmt.Sprintf("Collected %d execution requests for ConfigMapPropagation %s/%s", len(executionReqs), pr.Namespace, pr.Name),
	})

	// recreate the status array so that we create it from scratch
	// TODO: shall we create the items in advance with status=Unknown?
	var itemStatuses []kubegoodiesv1.PropagationStatus

	var errs error

	for _, executionReq := range executionReqs {
		// TODO: set status condition for each execution request
		if err := configmappropagation.Execute(ctx, r.Client, &executionReq); err != nil {
			// TODO: do not return here, but continue with the rest of the requests
			logger.Error(err, "unable to execute configmap propagation request", "request", executionReq)
			errs = multierror.Append(errs, fmt.Errorf("error executing request %v: %v", executionReq, err))

			itemStatuses = append(itemStatuses, kubegoodiesv1.PropagationStatus{
				SourceNamespace: executionReq.SourceNamespace,
				SourceName:      executionReq.SourceName,
				TargetNamespace: executionReq.TargetNamespace,
				TargetName:      executionReq.TargetName,
				Status:          metav1.ConditionFalse,
				Reason:          "PropagationFailed",
				Message:         fmt.Sprintf("error executing request %v", err),
			})
		} else {
			itemStatuses = append(itemStatuses, kubegoodiesv1.PropagationStatus{
				SourceNamespace: executionReq.SourceNamespace,
				SourceName:      executionReq.SourceName,
				TargetNamespace: executionReq.TargetNamespace,
				TargetName:      executionReq.TargetName,
				Status:          metav1.ConditionTrue,
				Reason:          "PropagationSucceeded",
				Message:         fmt.Sprintf("Propagated"),
			})
		}
	}

	pr.Status.PropagationStatus = itemStatuses

	if errs != nil {
		// TODO, update status before returning?
		return ctrl.Result{}, errs
	}

	meta.SetStatusCondition(&pr.Status.Conditions, metav1.Condition{
		Type:    kubegoodiesv1.ConfigMapPropagationConditionTypeReady,
		Status:  metav1.ConditionTrue,
		Reason:  "Ready",
		Message: fmt.Sprintf("ConfigMapPropagation %s/%s is ready", pr.Namespace, pr.Name),
	})

	if err := r.Status().Update(ctx, &pr); err != nil {
		logger.Error(err, "unable to update ConfigMapPropagation status")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ConfigMapPropagationReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&kubegoodiesv1.ConfigMapPropagation{}).
		Complete(r)
}
