package v1alpha1

import (
	"github.com/appscode/go/encoding/json/types"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	ResourceKindAzureRole = "AzureRole"
	ResourceAzureRole     = "azurerole"
	ResourceAzureRoles    = "azureroles"
)

// +genclient
// +k8s:openapi-gen=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// +kubebuilder:object:root=true
// +kubebuilder:resource:path=azureroles,singular=azurerole,categories={vault,appscode,all}
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.phase"
type AzureRole struct {
	metav1.TypeMeta   `json:",inline,omitempty"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              AzureRoleSpec   `json:"spec,omitempty"`
	Status            AzureRoleStatus `json:"status,omitempty"`
}

type AzureSecretType string

const (
	AzureClientSecret   = "client-secret"
	AzureSubscriptionID = "subscription-id"
	AzureTenantID       = "tenant-id"
	AzureClientID       = "client-id"
)

// AzureRoleSpec contains connection information, Azure role info, etc
// More info: https://www.vaultproject.io/api/secret/azure/index.html#create-update-role
type AzureRoleSpec struct {
	// VaultRef is the name of a AppBinding referencing to a Vault Server
	VaultRef core.LocalObjectReference `json:"vaultRef"`

	// Path defines the path of the Azure secret engine
	// default: azure
	// More info: https://www.vaultproject.io/docs/auth/azure.html#via-the-cli
	// +optional
	Path string `json:"path,omitempty"`

	// List of Azure roles to be assigned to the generated service principal.
	// The array must be in JSON format, properly escaped as a string
	AzureRoles string `json:"azureRoles,omitempty"`

	// Application Object ID for an existing service principal
	// that will be used instead of creating dynamic service principals.
	// If present, azure_roles will be ignored.
	ApplicationObjectID string `json:"applicationObjectID, omitempty"`

	// Specifies the default TTL for service principals generated using this role.
	// Accepts time suffixed strings ("1h") or an integer number of seconds.
	// Defaults to the system/engine default TTL time.
	TTL string `json:"ttl, omitempty"`

	// Specifies the maximum TTL for service principals
	// generated using this role. Accepts time suffixed strings ("1h")
	// or an integer number of seconds. Defaults to the system/engine max TTL time.
	MaxTTL string `json:"maxTTL, omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:object:root=true

type AzureRoleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata, omitempty"`

	// Items is a list of AzureRole objects
	Items []AzureRole `json:"items, omitempty"`
}

type AzureRolePhase string

type AzureRoleStatus struct {
	Phase AzureRolePhase `json:"phase,omitempty"`

	// observedGeneration is the most recent generation observed for this AzureRole. It corresponds to the
	// AzureRole's generation, which is updated on mutation by the API Server.
	ObservedGeneration *types.IntHash `json:"observedGeneration,omitempty"`

	// Represents the latest available observations of a AzureRole current state.
	Conditions []AzureRoleCondition `json:"conditions,omitempty"`
}

// AzureRoleCondition describes the state of a AzureRole at a certain point.
type AzureRoleCondition struct {
	// Type of AzureRole condition.
	Type string `json:"type,omitempty"`

	// Status of the condition, one of True, False, Unknown.
	Status core.ConditionStatus `json:"status,omitempty"`

	// The reason for the condition's.
	Reason string `json:"reason,omitempty"`

	// A human readable message indicating details about the transition.
	Message string `json:"message,omitempty"`
}
