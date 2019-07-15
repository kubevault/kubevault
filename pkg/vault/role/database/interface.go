package database

import (
	vaultapi "github.com/hashicorp/vault/api"
	rbacv1 "k8s.io/api/rbac/v1"
	"kubevault.dev/operator/pkg/vault/role"
)

type DatabaseRoleInterface interface {
	role.RoleInterface

	// EnableDatabase enables database secret engine
	EnableDatabase() error

	// IsDatabaseEnabled checks whether database is enabled or not
	IsDatabaseEnabled() (bool, error)

	// DeleteRole deletes role
	DeleteRole(name string) error
}

type DatabaseCredentialManager interface {
	// Gets credential from vault
	GetCredential() (*vaultapi.Secret, error)

	// Creates a kubernetes secret containing postgres credential
	CreateSecret(name string, namespace string, credential *vaultapi.Secret) error

	// Creates kubernetes role
	CreateRole(name string, namespace string, secretName string) error

	// Creates kubernetes role binding
	CreateRoleBinding(name string, namespace string, roleName string, subjects []rbacv1.Subject) error

	IsLeaseExpired(leaseID string) (bool, error)

	RevokeLease(leaseID string) error
}
