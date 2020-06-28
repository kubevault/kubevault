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
	rbac "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kmapi "kmodules.xyz/client-go/api/v1"
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
// +kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.phase"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
type AWSAccessKeyRequest struct {
	metav1.TypeMeta   `json:",inline,omitempty"`
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
	Spec              AWSAccessKeyRequestSpec   `json:"spec,omitempty" protobuf:"bytes,2,opt,name=spec"`
	Status            AWSAccessKeyRequestStatus `json:"status,omitempty" protobuf:"bytes,3,opt,name=status"`
}

// https://www.vaultproject.io/api/secret/aws/index.html#parameters-6
// AWSAccessKeyRequestSpec contains information to request for vault aws credential
type AWSAccessKeyRequestSpec struct {
	// Contains vault aws role info
	RoleRef RoleRef `json:"roleRef" protobuf:"bytes,1,opt,name=roleRef"`

	Subjects []rbac.Subject `json:"subjects" protobuf:"bytes,2,rep,name=subjects"`

	// The ARN of the role to assume if credential_type on the Vault role is assumed_role.
	// Must match one of the allowed role ARNs in the Vault role. Optional if the Vault role
	// only allows a single AWS role ARN; required otherwise.
	RoleARN string `json:"roleARN,omitempty" protobuf:"bytes,3,opt,name=roleARN"`

	// Specifies the TTL for the use of the STS token. This is specified as a string with a duration suffix.
	// Valid only when credential_type is assumed_role or federation_token. When not specified,
	// the default_sts_ttl set for the role will be used. If that is also not set, then the default value of
	// 3600s will be used. AWS places limits on the maximum TTL allowed. See the AWS documentation on the
	// DurationSeconds parameter for AssumeRole (for assumed_role credential types) and
	// GetFederationToken (for federation_token credential types) for more details.
	TTL string `json:"ttl,omitempty" protobuf:"bytes,4,opt,name=ttl"`

	// If true, '/aws/sts' endpoint will be used to retrieve credential
	// Otherwise, '/aws/creds' endpoint will be used to retrieve credential
	UseSTS bool `json:"useSTS,omitempty" protobuf:"varint,5,opt,name=useSTS"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:object:root=true

type AWSAccessKeyRequestList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	// Items is a list of AWSAccessKeyRequest objects
	Items []AWSAccessKeyRequest `json:"items,omitempty" protobuf:"bytes,2,rep,name=items"`
}

type AWSAccessKeyRequestStatus struct {
	// Specifies the phase of AWSAccessKeyRequestStatus object
	Phase RequestStatusPhase `json:"phase,omitempty" protobuf:"bytes,1,opt,name=phase,casttype=RequestStatusPhase"`

	// Conditions applied to the request, such as approval or denial.
	// +optional
	Conditions []kmapi.Condition `json:"conditions,omitempty" protobuf:"bytes,2,rep,name=conditions"`

	// Name of the secret containing AWSCredential AWSCredentials
	Secret *core.LocalObjectReference `json:"secret,omitempty" protobuf:"bytes,3,opt,name=secret"`

	// Contains lease info
	Lease *Lease `json:"lease,omitempty" protobuf:"bytes,4,opt,name=lease"`

	// observedGeneration is the most recent generation observed for this resource. It corresponds to the
	// resource's generation, which is updated on mutation by the API Server.
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty" protobuf:"varint,5,opt,name=observedGeneration"`
}
