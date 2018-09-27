package v1alpha1

import (
	"github.com/appscode/go/encoding/json/types"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	ResourceKindVaultPolicy = "VaultPolicy"
	ResourceVaultPolicy     = "vaultpolicy"
	ResourceVaultPolicies   = "vaultpolicies"
)

// +genclient
// +k8s:openapi-gen=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type VaultPolicy struct {
	metav1.TypeMeta   `json:",inline,omitempty"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              VaultPolicySpec   `json:"spec,omitempty"`
	Status            VaultPolicyStatus `json:"status,omitempty"`
}

type VaultPolicySpec struct {
	// Policy specifies the vault policy in hcl format.
	// For example:
	// path "secret/*" {
	//   capabilities = ["create", "read", "update", "delete", "list"]
	// }
	Policy string `json:"policy"`

	// Vault contains the information that necessary to talk with vault
	Vault *Vault `json:"vault"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type VaultPolicyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []VaultPolicy `json:"items,omitempty"`
}

type PolicyStatus string

const (
	PolicySuccess    PolicyStatus = "Success"
	PolicyFailed     PolicyStatus = "Failed"
	PolicyProcessing PolicyStatus = "Processing"
)

type VaultPolicyStatus struct {
	// observedGeneration is the most recent generation observed for this resource. It corresponds to the
	// resource's generation, which is updated on mutation by the API Server.
	// +optional
	ObservedGeneration *types.IntHash `json:"observedGeneration,omitempty"`

	// Status indicates whether the policy successfully applied in vault or not or in progress
	Status PolicyStatus `json:"status,omitempty"`

	// Represents the latest available observations of a VaultPolicy.
	Conditions []PolicyCondition `json:"conditions,omitempty"`
}

type PolicyConditionType string

// These are valid conditions of a VaultPolicy.
const (
	PolicyConditionFailure PolicyConditionType = "Failure"
)

// PolicyCondition describes the state of a VaultPolicy at a certain point.
type PolicyCondition struct {
	// Type of PolicyCondition condition.
	Type PolicyConditionType `json:"type,omitempty"`

	// Status of the condition, one of True, False, Unknown.
	Status core.ConditionStatus `json:"status,omitempty"`

	// The reason for the condition's.
	Reason string `json:"reason,omitempty"`

	// A human readable message indicating details about the transition.
	Message string `json:"message,omitempty"`
}

// Vault contains the information that necessary to talk with vault
type Vault struct {
	// Specifies the address of the vault server, e.g:'http://127.0.0.1:8200'
	Address string `json:"address"`

	// Name of the secret containing the vault token
	// access permission:
	// secret data:
	//	- token:<value>
	TokenSecret string `json:"tokenSecret"`

	// To skip tls verification for vault server
	SkipTLSVerification bool `json:"skipTLSVerification,omitempty"`

	// Name of the secret containing the ca cert to verify vault server
	// secret data:
	//	- ca.crt:<value>
	ServerCASecret string `json:"server_ca_secret,omitempty"`

	// Name of the secret containing the client.srt and client.key
	// secret data:
	//	- client.crt: <value>
	//	- client.srt: <value>
	ClientTLSSecret string `json:"clientTLSSecret,omitempty"`
}
