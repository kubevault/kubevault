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

package credential

import (
	"kubevault.dev/operator/pkg/vault/secret"

	vaultapi "github.com/hashicorp/vault/api"
	rbac "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type CredentialManager interface {
	// Gets credential from vault
	GetCredential() (*vaultapi.Secret, error)

	// Creates a kubernetes secret containing postgres credential
	CreateSecret(name string, namespace string, credential *vaultapi.Secret) error

	// Creates kubernetes role
	CreateRole(name string, namespace string, secretName string) error

	// Creates kubernetes role binding
	CreateRoleBinding(name string, namespace string, roleName string, subjects []rbac.Subject) error

	IsLeaseExpired(leaseID string) (bool, error)

	RevokeLease(leaseID string) error
}

type SecretEngine interface {
	secret.SecretGetter
	ParseCredential(secret *vaultapi.Secret) (map[string][]byte, error)
	GetOwnerReference() metav1.OwnerReference
}
