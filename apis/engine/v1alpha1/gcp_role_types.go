package v1alpha1

import (
	"github.com/appscode/go/encoding/json/types"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	appcat "kmodules.xyz/custom-resources/apis/appcatalog/v1alpha1"
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
type GCPRoleSpec struct {
	Ref *appcat.AppReference `json:"ref,omitempty"`

	Config *GCPConfig `json:"config"`

	// links:
	// 	- https://www.vaultproject.io/api/secret/gcp/index.html#parameters-1

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

// https://www.vaultproject.io/api/secret/gcp/index.html#write-config
// GCPConfig contains information to communicate with GCP
type GCPConfig struct {
	// Specifies the secret containing GCP credentials
	// secret.Data:
	//	- sa.json
	CredentialSecret string `json:"credentialSecret"`

	// Specifies default config TTL for long-lived credentials
	// (i.e. service account keys).
	// +optional
	TTL string `json:"ttl,omitempty"`

	// Specifies the maximum config TTL for long-lived
	// credentials (i.e. service account keys).
	// +optional
	MaxTTL string `json:"maxTTL,omitempty"`
}

type GCPRolePhase string

type GCPRoleStatus struct {
	Phase GCPRolePhase `json:"phase,omitempty"`

	// observedGeneration is the most recent generation observed for this GCPRole. It corresponds to the
	// GCPRole's generation, which is updated on mutation by the API Server.
	ObservedGeneration *types.IntHash `json:"observedGeneration,omitempty"`

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
