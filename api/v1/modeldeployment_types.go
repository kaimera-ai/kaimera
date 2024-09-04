package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ModelDeploymentSpec defines the desired state of ModelDeployment
type ModelDeploymentSpec struct {
	ModelName          string            `json:"modelName,omitempty"`
	NodeSelectorLabels map[string]string `json:"nodeSelectorLabels,omitempty"`
	Replicas           int32             `json:"replicas,omitempty"`
	Runtime            string            `json:"runtime,omitempty"`
	MaxModelLength     int32             `json:"maxModelLength,omitempty"`
	Provider           string            `json:"provider,omitempty"`
}

// ModelDeploymentStatus defines the observed state of ModelDeployment
type ModelDeploymentStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// ModelDeployment is the Schema for the modeldeployments API
type ModelDeployment struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ModelDeploymentSpec   `json:"spec,omitempty"`
	Status ModelDeploymentStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ModelDeploymentList contains a list of ModelDeployment
type ModelDeploymentList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ModelDeployment `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ModelDeployment{}, &ModelDeploymentList{})
}
