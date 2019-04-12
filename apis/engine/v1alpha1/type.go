package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type RoleReference struct {
	// Name of the object being referenced.
	Name string `json:"name"`

	// Namespace of the referenced object.
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
