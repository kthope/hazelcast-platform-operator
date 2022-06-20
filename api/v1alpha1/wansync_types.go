package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// WanSyncSpec defines the desired state of WanSync
type WanSyncSpec struct {
	// +kubebuilder:validation:MinLength:=1
	WanConfigurationResource string `json:"wanConfigurationResource,omitempty"`
}

// WanSyncStatus defines the observed state of WanSync
type WanSyncStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// WanSync is the Schema for the wansyncs API
type WanSync struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   WanSyncSpec   `json:"spec,omitempty"`
	Status WanSyncStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// WanSyncList contains a list of WanSync
type WanSyncList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []WanSync `json:"items"`
}

func init() {
	SchemeBuilder.Register(&WanSync{}, &WanSyncList{})
}
