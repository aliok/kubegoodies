package configmappropagation

import "k8s.io/apimachinery/pkg/types"

func SetPropagationAnnotation(annotations map[string]string, srcNamespace string, srcName string) {
	annotations[PropagationAnnotationNamespaceKey] = srcNamespace
	annotations[PropagationAnnotationNameKey] = srcName
}

func GetPropagationAnnotation(annotations map[string]string) *types.NamespacedName {
	ns, ok := annotations[PropagationAnnotationNamespaceKey]
	if !ok {
		return nil
	}
	name, ok := annotations[PropagationAnnotationNameKey]
	if !ok {
		return nil
	}

	return &types.NamespacedName{
		Namespace: ns,
		Name:      name,
	}
}
