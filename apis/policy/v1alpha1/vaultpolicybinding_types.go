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
	kmapi "kmodules.xyz/client-go/api/v1"
)

const (
	ResourceKindVaultPolicyBinding = "VaultPolicyBinding"
	ResourceVaultPolicyBinding     = "vaultpolicybinding"
	ResourceVaultPolicyBindings    = "vaultpolicybindings"
)

// VaultPolicyBinding binds a list of Vault server policies with Vault users authenticated by various auth methods.
// Currently VaultPolicyBinding only supports users authenticated via Kubernetes auth method.

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
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
	Spec              VaultPolicyBindingSpec   `json:"spec,omitempty" protobuf:"bytes,2,opt,name=spec"`
	Status            VaultPolicyBindingStatus `json:"status,omitempty" protobuf:"bytes,3,opt,name=status"`
}

// links: https://www.vaultproject.io/api/auth/kubernetes/index.html#parameters-1
type VaultPolicyBindingSpec struct {
	// VaultRef is the name of a AppBinding referencing to a Vault Server
	VaultRef core.LocalObjectReference `json:"vaultRef" protobuf:"bytes,1,opt,name=vaultRef"`

	// VaultRoleName is the role name which will be bound of the policies
	// This defaults to following format: k8s.${cluster}.${metadata.namespace}.${metadata.name}
	// xref: https://www.vaultproject.io/api/auth/kubernetes/index.html#create-role
	// +optional
	VaultRoleName string `json:"vaultRoleName,omitempty" protobuf:"bytes,2,opt,name=vaultRoleName"`

	// Policies is a list of Vault policy identifiers.
	Policies []PolicyIdentifier `json:"policies" protobuf:"bytes,3,rep,name=policies"`

	// SubjectRef refers to Vault users who will be granted policies.
	SubjectRef `json:"subjectRef" protobuf:"bytes,4,opt,name=subjectRef"`
}

type PolicyIdentifier struct {
	// Name is a Vault server policy name. This name should be returned by `vault read sys/policy` command.
	// More info: https://www.vaultproject.io/docs/concepts/policies.html#listing-policies
	Name string `json:"name,omitempty" protobuf:"bytes,1,opt,name=name"`

	// Ref is name of a VaultPolicy crd object. Actual vault policy name is spec.vaultRoleName field.
	// More info: https://www.vaultproject.io/docs/concepts/policies.html#listing-policies
	Ref string `json:"ref,omitempty" protobuf:"bytes,2,opt,name=ref"`
}

type SubjectRef struct {
	// Kubernetes refers to Vault users who are authenticated via Kubernetes auth method
	// More info: https://www.vaultproject.io/docs/auth/kubernetes.html#configuration
	Kubernetes *KubernetesSubjectRef `json:"kubernetes,omitempty" protobuf:"bytes,1,opt,name=kubernetes"`
	// More info: https://www.vaultproject.io/docs/auth/approle#configuration
	AppRole *AppRoleSubjectRef `json:"appRole,omitempty" protobuf:"bytes,2,opt,name=appRole"`
}

// More info: https://www.vaultproject.io/api/auth/kubernetes/index.html#create-role
type KubernetesSubjectRef struct {
	// Specifies the path where kubernetes auth is enabled
	// default : kubernetes
	// +optional
	Path string `json:"path,omitempty" protobuf:"bytes,1,opt,name=path"`

	// Name of the role
	Name string `json:"name,omitempty" protobuf:"bytes,2,rep,name=name"`

	// Specifies the names of the service account to bind with policy
	ServiceAccountNames []string `json:"serviceAccountNames" protobuf:"bytes,3,rep,name=serviceAccountNames"`

	// Specifies the namespaces of the service account
	ServiceAccountNamespaces []string `json:"serviceAccountNamespaces" protobuf:"bytes,4,rep,name=serviceAccountNamespaces"`

	//Specifies the TTL period of tokens issued using this role in seconds.
	// +optional
	TTL string `json:"ttl,omitempty" protobuf:"bytes,5,opt,name=ttl"`

	//Specifies the maximum allowed lifetime of tokens issued in seconds using this role.
	// +optional
	MaxTTL string `json:"maxTTL,omitempty" protobuf:"bytes,6,opt,name=maxTTL"`

	// If set, indicates that the token generated using this role should never expire.
	// The token should be renewed within the duration specified by this value.
	// At each renewal, the token's TTL will be set to the value of this parameter.
	// +optional
	Period string `json:"period,omitempty" protobuf:"bytes,7,opt,name=period"`
}

// More info: https://www.vaultproject.io/api-docs/auth/approle#create-update-approle
type AppRoleSubjectRef struct {
	// Specifies the path where approle auth is enabled
	// default : approle
	// +optional
	Path string `json:"path,omitempty" protobuf:"bytes,1,opt,name=path"`

	// RoleName is the Name of the AppRole
	// This defaults to following format: k8s.${cluster}.${metadata.namespace}.${metadata.name}
	RoleName string `json:"roleName,omitempty" protobuf:"bytes,2,opt,name=roleName"`

	// Require secret_id to be presented when logging in using this AppRole.
	BindSecretID bool `json:"bindSecretID" protobuf:"bytes,3,opt,name=bindSecretID"`

	// List of CIDR blocks; if set, specifies blocks of IP addresses which can perform the login operation.
	SecretIDBoundCidrs []string `json:"secretIdBoundCidrs,omitempty" protobuf:"bytes,4,opt,name=secretIdBoundCidrs"`

	// Number of times any particular SecretID can be used to fetch a token from this AppRole, after which the SecretID will expire. A value of zero will allow unlimited uses.
	SecretIDNumUses int64 `json:"secretIdNumUses,omitempty" protobuf:"bytes,5,opt,name=secretIdNumUses"`

	// Duration in either an integer number of seconds (3600) or an integer time unit (60m) after which any SecretID expires.
	SecretIDTTL string `json:"secretIdTTL,omitempty" protobuf:"bytes,6,opt,name=secretIdTTL"`

	// If set, the secret IDs generated using this role will be cluster local. This can only be set during role creation and once set, it can't be reset later.
	EnableLocalSecretIDs bool `json:"enableLocalSecretIds,omitempty" protobuf:"bytes,7,opt,name=enableLocalSecretIds"`

	// The incremental lifetime for generated tokens. This current value of this will be referenced at renewal time.
	TokenTTL int64 `json:"tokenTTL,omitempty" protobuf:"bytes,8,opt,name=tokenTTL"`

	// The maximum lifetime for generated tokens. This current value of this will be referenced at renewal time.
	TokenMaxTTL int64 `json:"tokenMaxTTL,omitempty" protobuf:"bytes,9,opt,name=tokentokenMaxTTL_max_ttl"`

	// List of policies to encode onto generated tokens. Depending on the auth method, this list may be supplemented by user/group/other values.
	TokenPolicies []string `json:"tokenPolicies,omitempty" protobuf:"bytes,10,opt,name=tokenPolicies"`

	// List of CIDR blocks; if set, specifies blocks of IP addresses which can authenticate successfully, and ties the resulting token to these blocks as well.
	TokenBoundCidrs []string `json:"tokenBoundCidrs,omitempty" protobuf:"bytes,11,opt,name=tokenBoundCidrs"`

	// If set, will encode an explicit max TTL onto the token. This is a hard cap even if token_ttl and token_max_ttl would otherwise allow a renewal.
	TokenExplicitMaxTTL int64 `json:"tokenExplicitMaxTTL,omitempty" protobuf:"bytes,12,opt,name=tokenExplicitMaxTTL"`

	// If set, the default policy will not be set on generated tokens; otherwise it will be added to the policies set in token_policies.
	TokenNoDefaultPolicy bool `json:"tokenNoDefaultPolicy,omitempty" protobuf:"bytes,13,opt,name=tokenNoDefaultPolicy"`

	// The maximum number of times a generated token may be used (within its lifetime); 0 means unlimited.
	TokenNumUses int64 `json:"tokenNumUses,omitempty" protobuf:"bytes,14,opt,name=tokenNumUses"`

	// The period, if any, to set on the token.
	TokenPeriod int64 `json:"tokenPeriod,omitempty" protobuf:"bytes,15,opt,name=tokenPeriod"`

	// The type of token that should be generated. Can be service, batch, or default to use the mount's tuned default (which unless changed will be service tokens). For token store roles, there are two additional possibilities: default-service and default-batch which specify the type to return unless the client requests a different type at generation time.
	TokenType string `json:"tokenType,omitempty" protobuf:"bytes,16,opt,name=tokenType"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:object:root=true

type VaultPolicyBindingList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
	Items           []VaultPolicyBinding `json:"items,omitempty" protobuf:"bytes,2,rep,name=items"`
}

// ServiceAccountReference contains name and namespace of the service account
type ServiceAccountReference struct {
	Name      string `json:"name" protobuf:"bytes,1,opt,name=name"`
	Namespace string `json:"namespace" protobuf:"bytes,2,opt,name=namespace"`
}

// +kubebuilder:validation:Enum=Success;Failed
type PolicyBindingPhase string

const (
	PolicyBindingSuccess PolicyBindingPhase = "Success"
	PolicyBindingFailed  PolicyBindingPhase = "Failed"
)

type VaultPolicyBindingStatus struct {
	// ObservedGeneration is the most recent generation observed for this resource. It corresponds to the
	// resource's generation, which is updated on mutation by the API Server.
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty" protobuf:"varint,1,opt,name=observedGeneration"`

	// Phase indicates whether successfully bind the policy to service account in vault or not or in progress
	// +optional
	Phase PolicyBindingPhase `json:"phase,omitempty" protobuf:"bytes,2,opt,name=phase,casttype=PolicyBindingPhase"`

	// Represents the latest available observations of a VaultPolicyBinding.
	// +optional
	Conditions []kmapi.Condition `json:"conditions,omitempty" protobuf:"bytes,3,rep,name=conditions"`
}
