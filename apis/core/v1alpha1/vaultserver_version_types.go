package v1alpha1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

const (
	ResourceKindVaultserverVersion = "VaultserverVersion"
	ResourceVaultserverVersion     = "vaultserverversion"
	ResourceVaultserverVersions    = "vaultserverversions"
)

// +genclient
// +genclient:nonNamespaced
// +genclient:skipVerbs=updateStatus
// +k8s:openapi-gen=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// VaultserverVersion defines a vaultserver version.
type VaultserverVersion struct {
	metav1.TypeMeta   `json:",inline,omitempty"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              VaultserverVersionSpec `json:"spec,omitempty"`
}

// VaultserverVersionSpec is the spec for postgres version
type VaultserverVersionSpec struct {
	// Version
	Version string `json:"version"`
	// Vault Image
	Vault VaultserverVersionVault `json:"vault"`
	// Unsealer Image
	Unsealer VaultserverVersionUnsealer `json:"unsealer"`
	// Deprecated versions usable but regarded as obsolete and best avoided, typically due to having been superseded.
	// +optional
	Deprecated bool `json:"deprecated,omitempty"`
}

// VaultserverVersionVault is the vault image
type VaultserverVersionVault struct {
	Image string `json:"image"`
}

// VaultserverVersionUnsealer is the image for the vault unsealer
type VaultserverVersionUnsealer struct {
	Image string `json:"image"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// VaultserverVersionList is a list of VaultserverVersions
type VaultserverVersionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	// Items is a list of VaultserverVersion CRD objects
	Items []VaultserverVersion `json:"items,omitempty"`
}
