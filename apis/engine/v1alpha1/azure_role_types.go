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
)

const (
	ResourceKindAzureRole = "AzureRole"
	ResourceAzureRole     = "azurerole"
	ResourceAzureRoles    = "azureroles"
)

// +genclient
// +k8s:openapi-gen=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// +kubebuilder:object:root=true
// +kubebuilder:resource:path=azureroles,singular=azurerole,categories={vault,appscode,all}
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.phase"
type AzureRole struct {
	metav1.TypeMeta   `json:",inline,omitempty"`
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
	Spec              AzureRoleSpec   `json:"spec,omitempty" protobuf:"bytes,2,opt,name=spec"`
	Status            AzureRoleStatus `json:"status,omitempty" protobuf:"bytes,3,opt,name=status"`
}

type AzureSecretType string

const (
	AzureClientSecret   = "client-secret"
	AzureSubscriptionID = "subscription-id"
	AzureTenantID       = "tenant-id"
	AzureClientID       = "client-id"
)

// AzureRoleSpec contains connection information, Azure role info, etc
// More info: https://www.vaultproject.io/api/secret/azure/index.html#create-update-role
type AzureRoleSpec struct {
	// VaultRef is the name of a AppBinding referencing to a Vault Server
	VaultRef core.LocalObjectReference `json:"vaultRef" protobuf:"bytes,1,opt,name=vaultRef"`

	// Path defines the path of the Azure secret engine
	// default: azure
	// More info: https://www.vaultproject.io/docs/auth/azure.html#via-the-cli
	// +optional
	Path string `json:"path,omitempty" protobuf:"bytes,2,opt,name=path"`

	// List of Azure roles to be assigned to the generated service principal.
	// The array must be in JSON format, properly escaped as a string
	AzureRoles string `json:"azureRoles,omitempty" protobuf:"bytes,3,opt,name=azureRoles"`

	// Application Object ID for an existing service principal
	// that will be used instead of creating dynamic service principals.
	// If present, azure_roles will be ignored.
	ApplicationObjectID string `json:"applicationObjectID,omitempty" protobuf:"bytes,4,opt,name=applicationObjectID"`

	// Specifies the default TTL for service principals generated using this role.
	// Accepts time suffixed strings ("1h") or an integer number of seconds.
	// Defaults to the system/engine default TTL time.
	TTL string `json:"ttl,omitempty" protobuf:"bytes,5,opt,name=ttl"`

	// Specifies the maximum TTL for service principals
	// generated using this role. Accepts time suffixed strings ("1h")
	// or an integer number of seconds. Defaults to the system/engine max TTL time.
	MaxTTL string `json:"maxTTL,omitempty" protobuf:"bytes,6,opt,name=maxTTL"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:object:root=true

type AzureRoleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	// Items is a list of AzureRole objects
	Items []AzureRole `json:"items,omitempty" protobuf:"bytes,2,rep,name=items"`
}

type AzureRolePhase string

type AzureRoleStatus struct {
	Phase AzureRolePhase `json:"phase,omitempty" protobuf:"bytes,1,opt,name=phase,casttype=AzureRolePhase"`

	// ObservedGeneration is the most recent generation observed for this AzureRole. It corresponds to the
	// AzureRole's generation, which is updated on mutation by the API Server.
	ObservedGeneration int64 `json:"observedGeneration,omitempty" protobuf:"varint,2,opt,name=observedGeneration"`

	// Represents the latest available observations of a AzureRole current state.
	Conditions []AzureRoleCondition `json:"conditions,omitempty" protobuf:"bytes,3,rep,name=conditions"`
}

// AzureRoleCondition describes the state of a AzureRole at a certain point.
type AzureRoleCondition struct {
	// Type of AzureRole condition.
	Type string `json:"type,omitempty" protobuf:"bytes,1,opt,name=type"`

	// Status of the condition, one of True, False, Unknown.
	Status core.ConditionStatus `json:"status,omitempty" protobuf:"bytes,2,opt,name=status,casttype=k8s.io/api/core/v1.ConditionStatus"`

	// The reason for the condition's.
	Reason string `json:"reason,omitempty" protobuf:"bytes,3,opt,name=reason"`

	// A human readable message indicating details about the transition.
	Message string `json:"message,omitempty" protobuf:"bytes,4,opt,name=message"`
}
