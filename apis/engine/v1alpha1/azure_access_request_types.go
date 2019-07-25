package v1alpha1

import (
	core "k8s.io/api/core/v1"
	rbac "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	ResourceKindAzureAccessKeyRequest = "AzureAccessKeyRequest"
	ResourceAzureAccessKeyRequest     = "azureaccesskeyrequest"
	ResourceAzureAccessKeyRequests    = "azureaccesskeyrequests"
)

// AzureAccessKeyRequest structure

// +genclient
// +k8s:openapi-gen=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// +kubebuilder:object:root=true
// +kubebuilder:resource:path=azureaccesskeyrequests,singular=azureaccesskeyrequest,categories={vault,appscode,all}
// +kubebuilder:subresource:status
type AzureAccessKeyRequest struct {
	metav1.TypeMeta   `json:",inline,omitempty"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              AzureAccessKeyRequestSpec   `json:"spec,omitempty"`
	Status            AzureAccessKeyRequestStatus `json:"status,omitempty"`
}

// Link:
//		- https://www.vaultproject.io/api/secret/azure/index.html#generate-credentials
//		- https://www.vaultproject.io/docs/secrets/azure/index.html#usage

// AzureAccessKeyRequestSpec contains information to request vault for credentials

type AzureAccessKeyRequestSpec struct {
	// Contains vault azure role info
	// +required
	RoleRef RoleReference `json:"roleRef"`

	// Contains a reference to the object or user identities the role binding is applied to
	// +required
	Subjects []rbac.Subject `json:"subjects"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:object:root=true

type AzureAccessKeyRequestList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	// Items is a list of AzureAccessKeyRequest objects
	Items []AzureAccessKeyRequest `json:"items,omitempty"`
}

type AzureAccessKeyRequestStatus struct {
	// Conditions applied to the request, such as approval or denial.
	// +optional
	Conditions []AzureAccessKeyRequestCondition `json:"conditions,omitempty"`

	// Name of the secret containing AzureCredential
	Secret *core.LocalObjectReference `json:"secret,omitempty"`

	// Contains lease info
	Lease *Lease `json:"lease,omitempty"`
}

type AzureAccessKeyRequestCondition struct {
	// request approval state, currently Approved or Denied.
	Type RequestConditionType `json:"type"`

	// brief reason for the request state
	// +optional
	Reason string `json:"reason,omitempty"`

	// human readable message with details about the request state
	// +optional
	Message string `json:"message,omitempty"`

	// timestamp for the last update to this condition
	// +optional
	LastUpdateTime metav1.Time `json:"lastUpdateTime,omitempty"`
}
