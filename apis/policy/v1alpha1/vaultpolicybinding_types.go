package v1alpha1

import (
	"github.com/appscode/go/encoding/json/types"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	ResourceKindVaultPolicyBinding = "VaultPolicyBinding"
	ResourceVaultPolicyBinding     = "vaultpolicybinding"
	ResourceVaultPolicyBindings    = "vaultpolicybindings"
)

// +genclient
// +k8s:openapi-gen=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// +kubebuilder:object:root=true
// +kubebuilder:resource:path=vaultpolicybindings,singular=vaultpolicybinding,shortName=vpb,categories={vault,appscode,all}
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.phase"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
type VaultPolicyBinding struct {
	metav1.TypeMeta   `json:",inline,omitempty"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              VaultPolicyBindingSpec   `json:"spec,omitempty"`
	Status            VaultPolicyBindingStatus `json:"status,omitempty"`
}

// links: https://www.vaultproject.io/api/auth/kubernetes/index.html#parameters-1
type VaultPolicyBindingSpec struct {
	// +optional
	RoleName string `json:"roleName,omitempty"`

	// Specifies the path where kubernetes auth is enabled
	// default : kubernetes
	// +optional
	AuthPath string `json:"authPath,omitempty"`

	// Specifies the names of the VaultPolicy
	Policies []string `json:"policies"`

	// Specifies the names of the service account to bind with policy
	ServiceAccountNames []string `json:"serviceAccountNames"`

	// Specifies the namespaces of the service account
	ServiceAccountNamespaces []string `json:"serviceAccountNamespaces"`

	//Specifies the TTL period of tokens issued using this role in seconds.
	// +optional
	TTL string `json:"ttl,omitempty"`

	//Specifies the maximum allowed lifetime of tokens issued in seconds using this role.
	// +optional
	MaxTTL string `json:"maxTTL,omitempty"`

	// If set, indicates that the token generated using this role should never expire.
	// The token should be renewed within the duration specified by this value.
	// At each renewal, the token's TTL will be set to the value of this parameter.
	// +optional
	Period string `json:"period,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:object:root=true

type VaultPolicyBindingList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []VaultPolicyBinding `json:"items,omitempty"`
}

// ServiceAccountReference contains name and namespace of the service account
type ServiceAccountReference struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}

type PolicyBindingPhase string

const (
	PolicyBindingSuccess PolicyBindingPhase = "Success"
	PolicyBindingFailed  PolicyBindingPhase = "Failed"
)

type VaultPolicyBindingStatus struct {
	// observedGeneration is the most recent generation observed for this resource. It corresponds to the
	// resource's generation, which is updated on mutation by the API Server.
	// +optional
	ObservedGeneration *types.IntHash `json:"observedGeneration,omitempty"`

	// Phase indicates whether successfully bind the policy to service account in vault or not or in progress
	// +optional
	Phase PolicyBindingPhase `json:"phase,omitempty"`

	// Represents the latest available observations of a VaultPolicyBinding.
	// +optional
	Conditions []PolicyBindingCondition `json:"conditions,omitempty"`
}

type PolicyBindingConditionType string

// These are valid conditions of a VaultPolicyBinding.
const (
	PolicyBindingConditionFailure PolicyBindingConditionType = "Failure"
)

// PolicyBindingCondition describes the state of a VaultPolicyBinding at a certain point.
type PolicyBindingCondition struct {
	// Type of PolicyBindingCondition condition.
	// +optional
	Type PolicyBindingConditionType `json:"type,omitempty"`

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
