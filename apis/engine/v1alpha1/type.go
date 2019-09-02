package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// RoleRef contains information that points to the role being used
type RoleRef struct {
	// APIGroup is the group for the resource being referenced
	APIGroup string `json:"apiGroup,omitempty"`
	// Kind is the type of resource being referenced
	Kind string `json:"kind,omitempty"`
	// Name is the name of resource being referenced
	Name string `json:"name"`
	// Namespace is the namespace of the resource being referenced
	Namespace string `json:"namespace"`
}

type RequestConditionType string

// These are the possible conditions for a certificate request.
const (
	AccessApproved RequestConditionType = "Approved"
	AccessDenied   RequestConditionType = "Denied"
)

// Lease contains lease info
type Lease struct {
	// lease id
	ID string `json:"id,omitempty"`

	// lease duration
	Duration metav1.Duration `json:"duration,omitempty"`

	// Specifies whether this lease is renewable
	Renewable bool `json:"renewable,omitempty"`
}
