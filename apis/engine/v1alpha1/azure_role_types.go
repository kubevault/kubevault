package v1alpha1

import (
	"github.com/appscode/go/encoding/json/types"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	appcat "kmodules.xyz/custom-resources/apis/appcatalog/v1alpha1"
)

const (
	ResourceKindAzureRole = "AzureRole"
	ResourceAzureRole     = "azurerole"
	ResourceAzureRoles    = "azureroles"
)

// +genclient
// +k8s:openapi-gen=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// AzureRole
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
type AzureRoleSpec struct {
	AuthManagerRef *appcat.AppReference `json:"authManagerRef,omitempty"`
	Config         *AzureConfig         `json:"config"`

	// ref:
	// - https://www.vaultproject.io/api/secret/azure/index.html#parameters

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

type AzureRoleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata, omitempty"`

	// Items is a list of AzureRole objects
	Items []AzureRole `json:"items, omitempty"`
}

// ref:
//	- https://www.vaultproject.io/api/secret/azure/index.html#configure-access

// AzureConfig contains information to communicate with Azure
type AzureConfig struct {

	// Specifies the secret name containing Azure credentials
	// secret.Data:
	// 	- subscription-id: <value>, The subscription id for the Azure Active Directory.
	//	- tenant-id: <value>, The tenant id for the Azure Active Directory.
	//	- client-id: <value>, The OAuth2 client id to connect to Azure.
	//	- client-secret: <value>, The OAuth2 client secret to connect to Azure.
	// +required
	CredentialSecret string `json:"credentialSecret"`

	// The Azure environment.
	// If not specified, Vault will use Azure Public Cloud.
	// +optional
	Environment string `json:"environment, omitempty"`
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
