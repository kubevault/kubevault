package v1alpha1

import (
	"github.com/appscode/go/encoding/json/types"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	appcat "kmodules.xyz/custom-resources/apis/appcatalog/v1alpha1"
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

	// Vault contains the reference of kmodules.xyz/custom-resources/apis/appcatalog/v1alpha1.AppBinding
	// which contains information to communicate with vault
	VaultAppRef *appcat.AppReference `json:"vaultAppRef"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type VaultPolicyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []VaultPolicy `json:"items,omitempty"`
}

type PolicyStatus string

const (
	PolicySuccess PolicyStatus = "Success"
	PolicyFailed  PolicyStatus = "Failed"
)

type VaultPolicyStatus struct {
	// observedGeneration is the most recent generation observed for this resource. It corresponds to the
	// resource's generation, which is updated on mutation by the API Server.
	// +optional
	ObservedGeneration *types.IntHash `json:"observedGeneration,omitempty"`

	// Status indicates whether the policy successfully applied in vault or not or in progress
	// +optional
	Status PolicyStatus `json:"status,omitempty"`

	// Represents the latest available observations of a VaultPolicy.
	// +optional
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
	// +optional
	Type PolicyConditionType `json:"type,omitempty"`

	// Status of the condition, one of True, False, Unknown.
	// +optional
	Status core.ConditionStatus `json:"status,omitempty"`

	// The reason for the condition's.
	// +optional
	Reason string `json:"reason,omitempty"`

	// A human readable message indicating details about the transition.
	// +optional
	Message string `json:"message,omitempty"`
}
