package v1alpha1

import (
	core "k8s.io/api/core/v1"
	rbac "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	ResourceKindAWSAccessKeyRequest = "AWSAccessKeyRequest"
	ResourceAWSAccessKeyRequest     = "awsaccesskeyrequest"
	ResourceAWSAccessKeyRequests    = "awsaccesskeyrequests"
)

// AWSAccessKeyRequest

// +genclient
// +k8s:openapi-gen=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// +kubebuilder:object:root=true
// +kubebuilder:resource:path=awsaccesskeyrequests,singular=awsaccesskeyrequest,categories={vault,appscode,all}
// +kubebuilder:subresource:status
type AWSAccessKeyRequest struct {
	metav1.TypeMeta   `json:",inline,omitempty"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              AWSAccessKeyRequestSpec   `json:"spec,omitempty"`
	Status            AWSAccessKeyRequestStatus `json:"status,omitempty"`
}

// https://www.vaultproject.io/api/secret/aws/index.html#parameters-6
// AWSAccessKeyRequestSpec contains information to request for vault aws credential
type AWSAccessKeyRequestSpec struct {
	// Contains vault aws role info
	RoleRef RoleReference `json:"roleRef"`

	Subjects []rbac.Subject `json:"subjects"`

	// The ARN of the role to assume if credential_type on the Vault role is assumed_role.
	// Must match one of the allowed role ARNs in the Vault role. Optional if the Vault role
	// only allows a single AWS role ARN; required otherwise.
	RoleARN string `json:"roleARN,omitempty"`

	// Specifies the TTL for the use of the STS token. This is specified as a string with a duration suffix.
	// Valid only when credential_type is assumed_role or federation_token. When not specified,
	// the default_sts_ttl set for the role will be used. If that is also not set, then the default value of
	// 3600s will be used. AWS places limits on the maximum TTL allowed. See the AWS documentation on the
	// DurationSeconds parameter for AssumeRole (for assumed_role credential types) and
	// GetFederationToken (for federation_token credential types) for more details.
	TTL string `json:"ttl,omitempty"`

	// If true, '/aws/sts' endpoint will be used to retrieve credential
	// Otherwise, '/aws/creds' endpoint will be used to retrieve credential
	UseSTS bool `json:"useSTS,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:object:root=true

type AWSAccessKeyRequestList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	// Items is a list of AWSAccessKeyRequest objects
	Items []AWSAccessKeyRequest `json:"items,omitempty"`
}

type AWSAccessKeyRequestStatus struct {
	// Conditions applied to the request, such as approval or denial.
	// +optional
	Conditions []AWSAccessKeyRequestCondition `json:"conditions,omitempty"`

	// Name of the secret containing AWSCredential AWSCredentials
	Secret *core.LocalObjectReference `json:"secret,omitempty"`

	// Contains lease info
	Lease *Lease `json:"lease,omitempty"`
}

type AWSAccessKeyRequestCondition struct {
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
