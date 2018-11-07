package v1alpha1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

const (
	ResourceKindVaultServerVersion = "VaultServerVersion"
	ResourceVaultServerVersion     = "vaultserverversion"
	ResourceVaultServerVersions    = "vaultserverversions"
)

// +genclient
// +genclient:nonNamespaced
// +genclient:skipVerbs=updateStatus
// +k8s:openapi-gen=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// VaultServerVersion defines a vaultserver version.
type VaultServerVersion struct {
	metav1.TypeMeta   `json:",inline,omitempty"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              VaultServerVersionSpec `json:"spec,omitempty"`
}

// VaultServerVersionSpec is the spec for postgres version
type VaultServerVersionSpec struct {
	// Version
	Version string `json:"version"`
	// Vault Image
	Vault VaultServerVersionVault `json:"vault"`
	// Unsealer Image
	Unsealer VaultServerVersionUnsealer `json:"unsealer"`
	// Exporter Image
	Exporter VaultServerVersionExporter `json:"exporter"`
	// Deprecated versions usable but regarded as obsolete and best avoided, typically due to having been superseded.
	// +optional
	Deprecated bool `json:"deprecated,omitempty"`
}

// VaultServerVersionVault is the vault image
type VaultServerVersionVault struct {
	Image string `json:"image"`
}

// VaultServerVersionUnsealer is the image for the vault unsealer
type VaultServerVersionUnsealer struct {
	Image string `json:"image"`
}

// VaultServerVersionExporter is the image for the vault exporter
type VaultServerVersionExporter struct {
	Image string `json:"image"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// VaultServerVersionList is a list of VaultserverVersions
type VaultServerVersionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	// Items is a list of VaultServerVersion CRD objects
	Items []VaultServerVersion `json:"items,omitempty"`
}
