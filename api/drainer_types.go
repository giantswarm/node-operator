package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	DrainerConfigStatusStatusTrue = "True"
)

const (
	DrainerConfigStatusTypeDrained = "Drained"
)

const (
	DrainerConfigStatusTypeTimeout = "Timeout"
)

const (
	kindDrainerConfig = "DrainerConfig"
)

func NewDrainerTypeMeta() metav1.TypeMeta {
	return metav1.TypeMeta{
		APIVersion: version,
		Kind:       kindDrainerConfig,
	}
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:storageversion
// +kubebuilder:subresource:status
// +kubebuilder:resource:categories=common;giantswarm
// +k8s:openapi-gen=true

type DrainerConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              DrainerConfigSpec `json:"spec"`
	// +kubebuilder:validation:Optional
	Status DrainerConfigStatus `json:"status"`
}

// +k8s:openapi-gen=true
type DrainerConfigSpec struct {
	Guest         DrainerConfigSpecGuest         `json:"guest"`
	VersionBundle DrainerConfigSpecVersionBundle `json:"versionBundle"`
}

// +k8s:openapi-gen=true
type DrainerConfigSpecGuest struct {
	Cluster DrainerConfigSpecGuestCluster `json:"cluster"`
	Node    DrainerConfigSpecGuestNode    `json:"node"`
}

// +k8s:openapi-gen=true
type DrainerConfigSpecGuestCluster struct {
	API DrainerConfigSpecGuestClusterAPI `json:"api"`
	// ID is the workload cluster ID of which a node should be drained.
	ID string `json:"id"`
}

// +k8s:openapi-gen=true
type DrainerConfigSpecGuestClusterAPI struct {
	// Endpoint is the workload cluster API endpoint.
	Endpoint string `json:"endpoint"`
}

// +k8s:openapi-gen=true
type DrainerConfigSpecGuestNode struct {
	// Name is the identifier of the workload cluster's master and worker nodes. In
	// Kubernetes/Kubectl they are represented as node names. The names are manage
	// in an abstracted way because of provider specific differences.
	//
	//     AWS: EC2 instance DNS.
	//     Azure: VM name.
	//     KVM: host cluster pod name.
	//
	Name string `json:"name"`
}

// +k8s:openapi-gen=true
type DrainerConfigSpecVersionBundle struct {
	Version string `json:"version"`
}

// +k8s:openapi-gen=true
type DrainerConfigStatus struct {
	Conditions []DrainerConfigStatusCondition `json:"conditions"`
}

// DrainerConfigStatusCondition expresses a condition in which a node may is.
// +k8s:openapi-gen=true
type DrainerConfigStatusCondition struct {
	// LastHeartbeatTime is the last time we got an update on a given condition.
	LastHeartbeatTime metav1.Time `json:"lastHeartbeatTime"`
	// LastTransitionTime is the last time the condition transitioned from one
	// status to another.
	LastTransitionTime metav1.Time `json:"lastTransitionTime"`
	// Status may be True, False or Unknown.
	Status string `json:"status"`
	// Type may be Pending, Ready, Draining, Drained.
	Type string `json:"type"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type DrainerConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []DrainerConfig `json:"items"`
}
