package credential

import (
	vaultapi "github.com/hashicorp/vault/api"
	"github.com/kubevault/operator/pkg/vault/secret"
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
