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
	ResourceKindGCPRole = "GCPRole"
	ResourceGCPRole     = "gcprole"
	ResourceGCPRoles    = "gcproles"
)

// +genclient
// +k8s:openapi-gen=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// +kubebuilder:object:root=true
// +kubebuilder:resource:path=gcproles,singular=gcprole,categories={vault,appscode,all}
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.phase"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
type GCPRole struct {
	metav1.TypeMeta   `json:",inline,omitempty"`
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
	Spec              GCPRoleSpec   `json:"spec,omitempty" protobuf:"bytes,2,opt,name=spec"`
	Status            GCPRoleStatus `json:"status,omitempty" protobuf:"bytes,3,opt,name=status"`
}

// +kubebuilder:validation:Enum=access_token;service_account_key
type GCPSecretType string

const (
	GCPSecretAccessToken       GCPSecretType = "access_token"
	GCPSecretServiceAccountKey GCPSecretType = "service_account_key"
)

// GCPRoleSpec contains connection information, GCP role info, etc
// More info: https://www.vaultproject.io/api/secret/gcp/index.html#parameters
type GCPRoleSpec struct {
	// VaultRef is the name of a AppBinding referencing to a Vault Server
	VaultRef core.LocalObjectReference `json:"vaultRef" protobuf:"bytes,1,opt,name=vaultRef"`

	// Path defines the path of the Google Cloud secret engine
	// default: gcp
	// More info: https://www.vaultproject.io/docs/auth/gcp.html#via-the-cli-helper
	// +optional
	Path string `json:"path,omitempty" protobuf:"bytes,2,opt,name=path"`

	// Specifies the type of secret generated for this role set
	SecretType GCPSecretType `json:"secretType" protobuf:"bytes,3,opt,name=secretType,casttype=GCPSecretType"`

	// Name of the GCP project that this roleset's service account will belong to.
	// Cannot be updated.
	Project string `json:"project" protobuf:"bytes,4,opt,name=project"`

	// Bindings configuration string (expects HCL or JSON format in raw
	// or base64-encoded string)
	Bindings string `json:"bindings" protobuf:"bytes,5,opt,name=bindings"`

	// List of OAuth scopes to assign to access_token secrets generated
	// under this role set (access_token role sets only)
	// +optional
	TokenScopes []string `json:"tokenScopes,omitempty" protobuf:"bytes,6,rep,name=tokenScopes"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:object:root=true

type GCPRoleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	// Items is a list of GCPRole objects
	Items []GCPRole `json:"items,omitempty" protobuf:"bytes,2,rep,name=items"`
}

const (
	GCPSACredentialJson = "sa.json"
)

type GCPRolePhase string

type GCPRoleStatus struct {
	Phase GCPRolePhase `json:"phase,omitempty" protobuf:"bytes,1,opt,name=phase,casttype=GCPRolePhase"`

	// ObservedGeneration is the most recent generation observed for this GCPRole. It corresponds to the
	// GCPRole's generation, which is updated on mutation by the API Server.
	ObservedGeneration int64 `json:"observedGeneration,omitempty" protobuf:"varint,2,opt,name=observedGeneration"`

	// Represents the latest available observations of a GCPRole current state.
	Conditions []kmapi.Condition `json:"conditions,omitempty" protobuf:"bytes,3,rep,name=conditions"`
}
