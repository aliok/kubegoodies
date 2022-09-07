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

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ConfigMapPropagationSpec defines the desired state of ConfigMapPropagation
type ConfigMapPropagationSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// +kubebuilder:validation:Required
	Source PropagationSource `json:"source"`

	// +kubebuilder:validation:Required
	Target PropagationTarget `json:"target"`
}

// +kubebuilder:validation:MinProperties=2
type PropagationSource struct {
	// Namespaces is a list of namespaces to watch for configmaps.
	// Type * to watch all namespaces.
	// +kubebuilder:validation:MinLength=1
	Namespace string `json:"namespace"`

	// Names is the list of configmaps to propagate.
	// Either specify Names or ObjectSelector.
	// +kubebuilder:validation:Optional
	Names []string `json:"names,omitempty"`

	// ObjectSelector is a selector to filter configmaps to propagate.
	// Either specify Names or ObjectSelector.
	// +kubebuilder:validation:Optional
	ObjectSelector *metav1.LabelSelector `json:"objectSelector,omitempty"`
}

type PropagationTarget struct {
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinItems=1
	Namespaces []string `json:"namespaces"`
}

// ConfigMapPropagationStatus defines the observed state of ConfigMapPropagation
type ConfigMapPropagationStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Conditions represent the latest available observations of an object's state
	// +kubebuilder:validation:Optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// PropagationStatus is the list of status of each propagation.
	// +kubebuilder:validation:Optional
	PropagationStatus []PropagationStatus `json:"propagationStatus,omitempty"`
}

type PropagationStatus struct {

	// SourceNamespace is the namespace of the source configmap.
	// +kubebuilder:validation:Required
	SourceNamespace string `json:"sourceNamespace"`

	// SourceName is the name of the source configmap.
	// +kubebuilder:validation:Required
	SourceName string `json:"sourceName"`

	// TargetNamespace is the namespace of the target configmap.
	// +kubebuilder:validation:Required
	TargetNamespace string `json:"targetNamespace"`

	// TargetName is the name of the target configmap.
	// +kubebuilder:validation:Required
	TargetName string `json:"targetName"`

	// Status is the status of the propagation. One of True, False, Unknown.
	// +kubebuilder:validation:Required
	Status metav1.ConditionStatus `json:"status"`

	// Reason contains a programmatic identifier indicating the reason for the condition's last transition.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MaxLength=1024
	// +kubebuilder:validation:MinLength=1
	Reason string `json:"reason"`

	// Message is a human readable message indicating details about the transition.
	// This may be an empty string.
	// +required
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MaxLength=32768
	Message string `json:"message"`
}

const (
	// ConfigMapPropagationConditionTypeReady is set when the ConfigMapPropagation is ready.
	ConfigMapPropagationConditionTypeReady = "Ready"

	// ConfigMapPropagationConditionTypeCollectedExecutionRequests is set when the ConfigMapPropagation has collected all execution requests.
	ConfigMapPropagationConditionTypeCollectedExecutionRequests = "CollectedExecutionRequests"
)

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// ConfigMapPropagation is the Schema for the configmappropagations API
type ConfigMapPropagation struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ConfigMapPropagationSpec   `json:"spec,omitempty"`
	Status ConfigMapPropagationStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ConfigMapPropagationList contains a list of ConfigMapPropagation
type ConfigMapPropagationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ConfigMapPropagation `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ConfigMapPropagation{}, &ConfigMapPropagationList{})
}
