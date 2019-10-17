package v1alpha1

import (
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	ResourceKindGCPRole = "GCPRole"
	ResourceGCPRole     = "gcprole"
	ResourceGCPRoles    = "gcproles"
)

// +genclient
// +k8s:openapi-gen=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// +kubebuilder:object:root=true
// +kubebuilder:resource:path=gcprole,singular=gcproles,categories={vault,appscode,all}
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.phase"
type GCPRole struct {
	metav1.TypeMeta   `json:",inline,omitempty"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              GCPRoleSpec   `json:"spec,omitempty"`
	Status            GCPRoleStatus `json:"status,omitempty"`
}

type GCPSecretType string

const (
	GCPSecretAccessToken       GCPSecretType = "access_token"
	GCPSecretServiceAccountKey GCPSecretType = "service_account_key"
)

// GCPRoleSpec contains connection information, GCP role info, etc
// More info: https://www.vaultproject.io/api/secret/gcp/index.html#parameters
type GCPRoleSpec struct {
	// VaultRef is the name of a AppBinding referencing to a Vault Server
	VaultRef core.LocalObjectReference `json:"vaultRef"`

	// Path defines the path of the Google Cloud secret engine
	// default: gcp
	// More info: https://www.vaultproject.io/docs/auth/gcp.html#via-the-cli-helper
	// +optional
	Path string `json:"path,omitempty"`

	// Specifies the type of secret generated for this role set
	SecretType GCPSecretType `json:"secretType"`

	// Name of the GCP project that this roleset's service account will belong to.
	// Cannot be updated.
	Project string `json:"project"`

	// Bindings configuration string (expects HCL or JSON format in raw
	// or base64-encoded string)
	Bindings string `json:"bindings"`

	// List of OAuth scopes to assign to access_token secrets generated
	// under this role set (access_token role sets only)
	// +optional
	TokenScopes []string `json:"tokenScopes,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:object:root=true

type GCPRoleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	// Items is a list of GCPRole objects
	Items []GCPRole `json:"items,omitempty"`
}

const (
	GCPSACredentialJson = "sa.json"
)

type GCPRolePhase string

type GCPRoleStatus struct {
	Phase GCPRolePhase `json:"phase,omitempty"`

	// ObservedGeneration is the most recent generation observed for this GCPRole. It corresponds to the
	// GCPRole's generation, which is updated on mutation by the API Server.
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`

	// Represents the latest available observations of a GCPRole current state.
	Conditions []GCPRoleCondition `json:"conditions,omitempty"`
}

// GCPRoleCondition describes the state of a GCPRole at a certain point.
type GCPRoleCondition struct {
	// Type of GCPRole condition.
	Type string `json:"type,omitempty"`

	// Status of the condition, one of True, False, Unknown.
	Status core.ConditionStatus `json:"status,omitempty"`

	// The reason for the condition's.
	Reason string `json:"reason,omitempty"`

	// A human readable message indicating details about the transition.
	Message string `json:"message,omitempty"`
}
