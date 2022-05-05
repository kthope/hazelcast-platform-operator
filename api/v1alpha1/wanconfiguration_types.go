package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// WanConfigurationSpec defines the desired state of WanConfiguration
type WanConfigurationSpec struct {
	// MapResourceName is the name of Map custom resource which WAN
	// replication will be applied to
	// +kubebuilder:validation:MinLength:=1
	MapResourceName string `json:"mapResourceName"`

	// ClusterName is the clusterName field of the target Hazelcast resource
	// +kubebuilder:validation:MinLength:=1
	TargetClusterName string `json:"targetClusterName"`

	// Endpoints is the target cluster endpoints
	Endpoints string `json:"endpoints"`
}

// WanConfigurationStatus defines the observed state of WanConfiguration
type WanConfigurationStatus struct {
	// State represents the current Hazelcast WAN Replication state
	State string `json:"state,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// WanConfiguration is the Schema for the wanconfigurations API
type WanConfiguration struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   WanConfigurationSpec   `json:"spec,omitempty"`
	Status WanConfigurationStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// WanConfigurationList contains a list of WanConfiguration
type WanConfigurationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []WanConfiguration `json:"items"`
}

func init() {
	SchemeBuilder.Register(&WanConfiguration{}, &WanConfigurationList{})
}
