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
	ResourceKindVaultAppRole = "VaultAppRole"
	ResourceVaultAppRole     = "vaultapprole"
	ResourceVaultAppRoles    = "vaultapproles"
)

// +genclient
// +k8s:openapi-gen=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// +kubebuilder:object:root=true
// +kubebuilder:resource:path=vaultapproles,singular=vaultapprole,shortName=vp,categories={vault,approle,appscode,all}
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.phase"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
type VaultAppRole struct {
	metav1.TypeMeta   `json:",inline,omitempty"`
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
	Spec              VaultAppRoleSpec   `json:"spec,omitempty" protobuf:"bytes,2,opt,name=spec"`
	Status            VaultAppRoleStatus `json:"status,omitempty" protobuf:"bytes,3,opt,name=status"`
}

// More info: https://www.vaultproject.io/api-docs/auth/approle
type VaultAppRoleSpec struct {
	// VaultRef is the name of a AppBinding referencing to a Vault Server
	VaultRef core.LocalObjectReference `json:"vaultRef" protobuf:"bytes,1,opt,name=vaultRef"`

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

type VaultAppRoleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
	Items           []VaultAppRole `json:"items,omitempty" protobuf:"bytes,2,rep,name=items"`
}

// +kubebuilder:validation:Enum=Success;Failed
type PolicyPhase string

const (
	AppRoleSuccess PolicyPhase = "Success"
	AppRoleFailed  PolicyPhase = "Failed"
)

type VaultAppRoleStatus struct {
	// ObservedGeneration is the most recent generation observed for this resource. It corresponds to the
	// resource's generation, which is updated on mutation by the API Server.
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty" protobuf:"varint,1,opt,name=observedGeneration"`

	// Phase indicates whether the policy successfully applied in vault or not or in progress
	// +optional
	Phase PolicyPhase `json:"phase,omitempty" protobuf:"bytes,2,opt,name=phase,casttype=PolicyPhase"`

	// Represents the latest available observations of a VaultPolicy.
	// +optional
	Conditions []kmapi.Condition `json:"conditions,omitempty" protobuf:"bytes,3,rep,name=conditions"`
}
