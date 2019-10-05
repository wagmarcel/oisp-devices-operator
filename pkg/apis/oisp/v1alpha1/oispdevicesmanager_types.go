package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	corev1 "k8s.io/api/core/v1"
)

const (
	PhasePending = "PENDING"
	PhaseRunning = "RUNNING"
	PhaseError 	 = "ERROR"
)

// OispDevicesManagerSpec defines the desired state of OispDevicesManager
// +k8s:openapi-gen=true
type OispDevicesManagerSpec struct {
	// Label name to identify oisp-managed nodes
	WatchLabelKey string `json:"watchLabelKey"`
	// Label value
	WatchLabelValue string `json:"watchLabelValue"`
	// annotation name which contains the device resource description
	WatchAnnotationKey string `json:"watchAnnotationKey"`
	// PodTemplate for every plugin deployment
	PodTemplateSpec corev1.PodTemplateSpec `json:"podTemplateSpec"`
}

// OispDevicesManagerStatus defines the observed state of OispDevicesManager
// +k8s:openapi-gen=true
type OispDevicesManagerStatus struct {

	//State of the CRD owning operator
	Phase string `json:"phase,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// OispDevicesManager is the Schema for the oispdevicesmanagers API
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
type OispDevicesManager struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   OispDevicesManagerSpec   `json:"spec,omitempty"`
	Status OispDevicesManagerStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// OispDevicesManagerList contains a list of OispDevicesManager
type OispDevicesManagerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []OispDevicesManager `json:"items"`
}

func init() {
	SchemeBuilder.Register(&OispDevicesManager{}, &OispDevicesManagerList{})
}
