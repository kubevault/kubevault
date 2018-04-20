package v1alpha1

import (
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	ResourceKindVaultServer     = "VaultServer"
	ResourceSingularVaultServer = "vaultserver"
	ResourcePluralVaultServer   = "vaultservers"
)

// +genclient
// +genclient:skipVerbs=updateStatus
// +k8s:openapi-gen=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type VaultServer struct {
	metav1.TypeMeta   `json:",inline,omitempty"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              VaultServerSpec `json:"spec,omitempty"`
}

type VaultServerSpec struct {
	Selector metav1.LabelSelector `json:"selector,omitempty"`
	Schedule string               `json:"schedule,omitempty"`
	// Pod volumes to mount into the sidecar container's filesystem.
	VolumeMounts []core.VolumeMount `json:"volumeMounts,omitempty"`
	// Compute Resources required by the sidecar container.
	Resources core.ResourceRequirements `json:"resources,omitempty"`
	//Indicates that the VaultServer is paused from taking backup. Default value is 'false'
	// +optional
	Paused bool `json:"paused,omitempty"`
	// ImagePullSecrets is an optional list of references to secrets in the same namespace to use for pulling any of the images used by this PodSpec.
	// If specified, these secrets will be passed to individual puller implementations for them to use. For example,
	// in the case of docker, only DockerConfig type secrets are honored.
	// More info: https://kubernetes.io/docs/concepts/containers/images#specifying-imagepullsecrets-on-a-pod
	// +optional
	ImagePullSecrets []core.LocalObjectReference `json:"imagePullSecrets,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type VaultServerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []VaultServer `json:"items,omitempty"`
}
