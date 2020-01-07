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
	// +optional
	Path string `json:"path,omitempty" protobuf:"bytes,1,opt,name=path"`

	// Specifies the vault role name for policy controller
	// It has permission to create policy in vault
	// +optional
	VaultRole string `json:"vaultRole,omitempty" protobuf:"bytes,2,opt,name=vaultRole"`

	// Specifies the Kubernetes authentication information
	// +optional
	Kubernetes *KubernetesAuthConfig `json:"kubernetes,omitempty" protobuf:"bytes,3,opt,name=kubernetes"`

	// Specifies the Azure authentication information
	// +optional
	Azure *AzureAuthConfig `json:"azure,omitempty" protobuf:"bytes,4,opt,name=azure"`

	// Specifies the AWS authentication information
	// +optional
	AWS *AWSAuthConfig `json:"aws,omitempty" protobuf:"bytes,5,opt,name=aws"`
}

// KubernetesAuthConfiguration contains necessary information for
// performing Kubernetes authentication to the Vault server.
type KubernetesAuthConfig struct {
	// Specifies the service account name
	ServiceAccountName string `json:"serviceAccountName" protobuf:"bytes,1,opt,name=serviceAccountName"`

	// Specifies the service account name for token reviewer
	// It has system:auth-delegator permission
	// It's jwt token is used on vault kubernetes auth config
	// +optional
	TokenReviewerServiceAccountName string `json:"tokenReviewerServiceAccountName,omitempty" protobuf:"bytes,2,opt,name=tokenReviewerServiceAccountName"`

	// Specifies to use pod service account for vault csi driver
	// +optional
	UsePodServiceAccountForCSIDriver bool `json:"usePodServiceAccountForCSIDriver,omitempty" protobuf:"varint,3,opt,name=usePodServiceAccountForCSIDriver"`
}

// AzureAuthConfig contains necessary information for
// performing Azure authentication to the Vault server.
type AzureAuthConfig struct {
	// Specifies the subscription ID for the machine
	// that generated the MSI token.
	// +optional
	SubscriptionID string `json:"subscriptionID,omitempty" protobuf:"bytes,1,opt,name=subscriptionID"`

	// Specifies the resource group for the machine
	// that generated the MSI token.
	// +optional
	ResourceGroupName string `json:"resourceGroupName,omitempty" protobuf:"bytes,2,opt,name=resourceGroupName"`

	// Specifies the virtual machine name for the machine
	// that generated the MSI token. If VmssName is provided,
	// this value is ignored.
	// +optional
	VmName string `json:"vmName,omitempty" protobuf:"bytes,3,opt,name=vmName"`

	// Specifies the virtual machine scale set name
	// for the machine that generated the MSI token.
	// +optional
	VmssName string `json:"vmssName,omitempty" protobuf:"bytes,4,opt,name=vmssName"`
}

// AWSAuthConfig contains necessary information for
// performing AWS authentication to the Vault server.
type AWSAuthConfig struct {
	// Specifies the header value that required
	// if X-Vault-AWS-IAM-Server-ID Header is set in Vault.
	// +optional
	HeaderValue string `json:"headerValue,omitempty" protobuf:"bytes,1,opt,name=headerValue"`
}
