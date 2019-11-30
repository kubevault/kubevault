/*
Copyright The KubeVault Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	ResourceKindAWSRole = "AWSRole"
	ResourceAWSRole     = "awsrole"
	ResourceAWSRoles    = "awsroles"
)

// AWSRole

// +genclient
// +k8s:openapi-gen=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// +kubebuilder:object:root=true
// +kubebuilder:resource:path=awsroles,singular=awsrole,categories={vault,appscode,all}
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.phase"
type AWSRole struct {
	metav1.TypeMeta   `json:",inline,omitempty"`
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
	Spec              AWSRoleSpec   `json:"spec,omitempty" protobuf:"bytes,2,opt,name=spec"`
	Status            AWSRoleStatus `json:"status,omitempty" protobuf:"bytes,3,opt,name=status"`
}

type AWSCredentialType string

const (
	AWSCredentialIAMUser         AWSCredentialType = "iam_user"
	AWSCredentialAssumedRole     AWSCredentialType = "assumed_role"
	AWSCredentialFederationToken AWSCredentialType = "federation_token"
)

// AWSRoleSpec contains connection information, AWS role info, etc
// More info: https://www.vaultproject.io/api/secret/aws/index.html#parameters-3
type AWSRoleSpec struct {
	// VaultRef is the name of a AppBinding referencing to a Vault Server
	VaultRef core.LocalObjectReference `json:"vaultRef" protobuf:"bytes,1,opt,name=vaultRef"`

	// Path defines the path of the AWS secret engine
	// default: aws
	// More info: https://www.vaultproject.io/docs/auth/aws.html#via-the-cli
	// +optional
	Path string `json:"path,omitempty" protobuf:"bytes,2,opt,name=path"`

	// Specifies the type of credential to be used when retrieving credentials from the role
	CredentialType AWSCredentialType `json:"credentialType" protobuf:"bytes,3,opt,name=credentialType,casttype=AWSCredentialType"`

	// Specifies the ARNs of the AWS roles this Vault role is allowed to assume.
	// Required when credential_type is assumed_role and prohibited otherwise
	RoleARNs []string `json:"roleARNs,omitempty" protobuf:"bytes,4,rep,name=roleARNs"`

	// Specifies the ARNs of the AWS managed policies to be attached to IAM users when they are requested.
	// Valid only when credential_type is iam_user. When credential_type is iam_user,
	// at least one of policy_arns or policy_document must be specified.
	PolicyARNs []string `json:"policyARNs,omitempty" protobuf:"bytes,5,rep,name=policyARNs"`

	// The IAM policy document for the role. The behavior depends on the credential type.
	// With iam_user, the policy document will be attached to the IAM user generated and
	// augment the permissions the IAM user has. With assumed_role and federation_token,
	// the policy document will act as a filter on what the credentials can do.
	// +optional
	PolicyDocument string `json:"policyDocument,omitempty" protobuf:"bytes,6,opt,name=policyDocument"`

	// Specifies the IAM policy in JSON format.
	// +optional
	// +kubebuilder:validation:EmbeddedResource
	// +kubebuilder:pruning:PreserveUnknownFields
	Policy *runtime.RawExtension `json:"policy,omitempty" protobuf:"bytes,7,opt,name=policy"`

	// The default TTL for STS credentials. When a TTL is not specified when STS credentials are requested,
	// and a default TTL is specified on the role, then this default TTL will be used.
	// Valid only when credential_type is one of assumed_role or federation_token
	DefaultSTSTTL string `json:"defaultSTSTTL,omitempty" protobuf:"bytes,8,opt,name=defaultSTSTTL"`

	// The max allowed TTL for STS credentials (credentials TTL are capped to max_sts_ttl).
	// Valid only when credential_type is one of assumed_role or federation_token
	MaxSTSTTL string `json:"maxSTSTTL,omitempty" protobuf:"bytes,9,opt,name=maxSTSTTL"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:object:root=true

type AWSRoleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	// Items is a list of AWSRole objects
	Items []AWSRole `json:"items,omitempty" protobuf:"bytes,2,rep,name=items"`
}

const (
	AWSCredentialAccessKeyKey = "access_key"
	AWSCredentialSecretKeyKey = "secret_key"
)

type AWSRolePhase string

type AWSRoleStatus struct {
	Phase AWSRolePhase `json:"phase,omitempty" protobuf:"bytes,1,opt,name=phase,casttype=AWSRolePhase"`

	// ObservedGeneration is the most recent generation observed for this AWSRole. It corresponds to the
	// AWSRole's generation, which is updated on mutation by the API Server.
	ObservedGeneration int64 `json:"observedGeneration,omitempty" protobuf:"varint,2,opt,name=observedGeneration"`

	// Represents the latest available observations of a AWSRole current state.
	Conditions []AWSRoleCondition `json:"conditions,omitempty" protobuf:"bytes,3,rep,name=conditions"`
}

// AWSRoleCondition describes the state of a AWSRole at a certain point.
type AWSRoleCondition struct {
	// Type of AWSRole condition.
	Type string `json:"type,omitempty" protobuf:"bytes,1,opt,name=type"`

	// Status of the condition, one of True, False, Unknown.
	Status core.ConditionStatus `json:"status,omitempty" protobuf:"bytes,2,opt,name=status,casttype=k8s.io/api/core/v1.ConditionStatus"`

	// The reason for the condition's.
	Reason string `json:"reason,omitempty" protobuf:"bytes,3,opt,name=reason"`

	// A human readable message indicating details about the transition.
	Message string `json:"message,omitempty" protobuf:"bytes,4,opt,name=message"`
}
