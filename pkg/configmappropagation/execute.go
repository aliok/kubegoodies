package configmappropagation

import (
	"context"
	"fmt"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func Execute(ctx context.Context, cl client.Client, req *Request) error {
	// TODO: what about labels and annotations?
	// TODO: some way of filtering out stuff?
	// TODO: set an annotation like "github.com/aliok/bla: DO NOT EDIT. THIS CONFIGMAP IS PROPAGATED FROM namespace/foo"

	logger := log.FromContext(ctx)

	logger.Info("propagating", "request", req)

	if req.SourceNamespace == "" {
		return fmt.Errorf("sourceNamespace cannot be empty")
	}

	if req.SourceName == "" {
		return fmt.Errorf("sourceName cannot be empty")
	}

	if req.TargetNamespace == "" {
		return fmt.Errorf("targetNamespace cannot be empty")
	}

	if req.TargetName == "" {
		req.TargetName = req.SourceName
	}

	var sourceCm corev1.ConfigMap
	err := cl.Get(ctx, types.NamespacedName{Namespace: req.SourceNamespace, Name: req.SourceName}, &sourceCm)

	if err != nil && !apierrors.IsNotFound(err) {
		return fmt.Errorf("error getting the source configmap: %v", err)
	}

	sourceExists := err == nil && sourceCm.DeletionTimestamp == nil

	if !sourceExists {
		targetCm := &corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Namespace: req.TargetNamespace, Name: req.TargetName}}
		if err := cl.Delete(ctx, targetCm); client.IgnoreNotFound(err) != nil {
			return fmt.Errorf("error deleting the target configmap: %v", err)
		}
		return nil
	}

	// clone informer's copy
	var annotations = make(map[string]string, len(sourceCm.Annotations))
	for k, v := range sourceCm.Annotations {
		annotations[k] = v
	}

	// set our custom annotation
	SetPropagationAnnotation(annotations, req.SourceNamespace, req.SourceName)

	targetCm := &corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Namespace: req.TargetNamespace, Name: req.TargetName}}

	op, err := controllerutil.CreateOrPatch(ctx, cl, targetCm, func() error {
		targetCm.Annotations = annotations // copy source annotations and add our annotation
		targetCm.Labels = sourceCm.Labels  // copy source labels
		targetCm.OwnerReferences = nil     // cannot set cross namespace ownerRef to source configMap

		targetCm.Immutable = sourceCm.Immutable
		targetCm.Data = sourceCm.Data
		targetCm.BinaryData = sourceCm.BinaryData

		return nil
	})

	if err != nil {
		return fmt.Errorf("error applying the target configmap: %v", err)
	}

	logger.Info("propagated", "request", req, "operation", op)

	return nil
}
