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

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

const (
	ResourceKindVaultServerConfiguration = "VaultServerConfiguration"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// VaultServerConfiguration defines a Vault Server configuration.
type VaultServerConfiguration struct {
	// +optional
	metav1.TypeMeta `json:",inline,omitempty"`

	// Specifies the path which is used for authentication by this AppBinding.
	// If vault server is provisioned by KubeVault, this is usually `kubernetes`.
	Path string `json:"path" protobuf:"bytes,1,opt,name=path"`

	// Specifies the service account name
	// +optional
	ServiceAccountName string `json:"serviceAccountName,omitempty" protobuf:"bytes,2,opt,name=serviceAccountName"`

	// Specifies the service account name for token reviewer
	// It has system:auth-delegator permission
	// It's jwt token is used on vault kubernetes auth config
	// +optional
	TokenReviewerServiceAccountName string `json:"tokenReviewerServiceAccountName,omitempty" protobuf:"bytes,3,opt,name=tokenReviewerServiceAccountName"`

	// Specifies the vault role name for policy controller
	// It has permission to create policy in vault
	// +optional
	PolicyControllerRole string `json:"policyControllerRole,omitempty" protobuf:"bytes,4,opt,name=policyControllerRole"`

	// Specifies the vault role name for auth controller
	// It has permission to enable/disable auth method in vault
	// +optional
	AuthMethodControllerRole string `json:"authMethodControllerRole,omitempty" protobuf:"bytes,5,opt,name=authMethodControllerRole"`

	// Specifies to use pod service account for vault csi driver
	// +optional
	UsePodServiceAccountForCSIDriver bool `json:"usePodServiceAccountForCsiDriver,omitempty" protobuf:"varint,6,opt,name=usePodServiceAccountForCsiDriver"`
}
