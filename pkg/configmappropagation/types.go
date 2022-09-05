package configmappropagation

const (
	PropagationAnnotationNamespaceKey = "kubegoodies-configmap-propagation-source-namespace"
	PropagationAnnotationNameKey      = "kubegoodies-configmap-propagation-source-name"
)

type Request struct {
	SourceNamespace string
	SourceName      string
	TargetNamespace string
	TargetName      string
	//  TODO: mod?
}

type Result string
