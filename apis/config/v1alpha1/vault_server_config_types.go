package v1alpha1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

const (
	ResourceKindVaultServerConfiguration = "VaultServerConfiguration"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// VaultServerConfiguration defines a Vault Server configuration.
type VaultServerConfiguration struct {
	metav1.TypeMeta `json:",inline,omitempty"`

	// Specifies the service account name for token reviewer
	// It has system:auth-delegator permission
	// It's jwt token is used on vault kubernetes auth config
	TokenReviewerServiceAccountName string `json:"tokenReviewerServiceAccountName,omitempty"`

	// Specifies the service account name for policy controller
	// It has permission to create policy in vault
	PolicyControllerServiceAccountName string `json:"policyControllerServiceAccountName,omitempty"`

	// Specifies to use pod service account for vault csi driver
	UsePodServiceAccountForCSIDriver bool `json:"usePodServiceAccountForCsiDriver,omitempty"`

	// Specifies the path where kubernetes auth is enabled
	// default : kubernetes
	AuthPath string `json:"authPath,omitempty"`
}
