package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// RoleRef contains information that points to the role being used
type RoleRef struct {
	// APIGroup is the group for the resource being referenced
	APIGroup string `json:"apiGroup"`
	// Kind is the type of resource being referenced
	Kind string `json:"kind"`
	// Name is the name of resource being referenced
	Name string `json:"name"`
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
