package v1alpha1

import (
	core "k8s.io/api/core/v1"
	rbac "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	ResourceKindDatabaseAccessRequest = "DatabaseAccessRequest"
	ResourceDatabaseAccessRequest     = "databaseaccessrequest"
	ResourceDatabaseAccessRequests    = "databaseaccessrequests"
)

// +genclient
// +k8s:openapi-gen=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// +kubebuilder:object:root=true
// +kubebuilder:resource:path=databaseaccessrequests,singular=databaseaccessrequest,categories={vault,appscode,all}
// +kubebuilder:subresource:status
type DatabaseAccessRequest struct {
	metav1.TypeMeta   `json:",inline,omitempty"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              DatabaseAccessRequestSpec   `json:"spec,omitempty"`
	Status            DatabaseAccessRequestStatus `json:"status,omitempty"`
}

// DatabaseAccessRequestSpec contains information to request for database credential
type DatabaseAccessRequestSpec struct {
	// Contains vault database role info
	RoleRef RoleRef `json:"roleRef"`

	Subjects []rbac.Subject `json:"subjects"`

	// Specifies the TTL for the leases associated with this role.
	// Accepts time suffixed strings ("1h") or an integer number of seconds.
	// Defaults to roles default TTL time
	TTL string `json:"ttl,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type DatabaseAccessRequestList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	// Items is a list of DatabaseAccessRequest objects
	Items []DatabaseAccessRequest `json:"items,omitempty"`
}

type DatabaseAccessRequestStatus struct {
	// Conditions applied to the request, such as approval or denial.
	// +optional
	Conditions []DatabaseAccessRequestCondition `json:"conditions,omitempty"`

	// Name of the secret containing database credentials
	Secret *core.LocalObjectReference `json:"secret,omitempty"`

	// Contains lease info
	Lease *Lease `json:"lease,omitempty"`
}

type DatabaseAccessRequestCondition struct {
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
