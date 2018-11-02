package v1alpha1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

const (
	ResourceKindVaultServerConfiguration = "VaultServerConfiguration"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// VaultServerConfiguration defines a Vault Server configuration.
type VaultServerConfiguration struct {
	// +optional
	metav1.TypeMeta `json:",inline,omitempty"`

	// Specifies the service account name
	// +optional
	ServiceAccountName string `json:"serviceAccountName,omitempty"`

	// Specifies the service account name for token reviewer
	// It has system:auth-delegator permission
	// It's jwt token is used on vault kubernetes auth config
	// +optional
	TokenReviewerServiceAccountName string `json:"tokenReviewerServiceAccountName,omitempty"`

	// Specifies the vault role name for policy controller
	// It has permission to create policy in vault
	// +optional
	PolicyControllerRole string `json:"policyControllerRole,omitempty"`

	// Specifies the vault role name for auth controller
	// It has permission to enable/disable auth method in vault
	// +optional
	AuthMethodControllerRole string `json:"authMethodControllerRole,omitempty"`

	// Specifies to use pod service account for vault csi driver
	// +optional
	UsePodServiceAccountForCSIDriver bool `json:"usePodServiceAccountForCsiDriver,omitempty"`

	// Specifies the path where kubernetes auth is enabled
	// default : kubernetes
	// +optional
	AuthPath string `json:"authPath,omitempty"`
}
