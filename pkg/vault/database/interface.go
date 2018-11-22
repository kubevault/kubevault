package database

import (
	"github.com/kubevault/operator/pkg/vault"
	rbacv1 "k8s.io/api/rbac/v1"
)

type DatabaseRoleInterface interface {
	RoleInterface

	// EnableDatabase enables database secret engine
	EnableDatabase() error

	// IsDatabaseEnabled checks whether database is enabled or not
	IsDatabaseEnabled() (bool, error)

	// DeleteRole deletes role
	DeleteRole(name string) error
}

type RoleInterface interface {
	// CreateConfig creates database configuration
	CreateConfig() error

	// CreateRole creates role
	CreateRole() error
}

type DatabaseCredentialManager interface {
	// Gets credential from vault
	GetCredential() (*vault.DatabaseCredential, error)

	// Creates a kubernetes secret containing postgres credential
	CreateSecret(name string, namespace string, credential *vault.DatabaseCredential) error

	// Creates kubernetes role
	CreateRole(name string, namespace string, secretName string) error

	// Creates kubernetes role binding
	CreateRoleBinding(name string, namespace string, roleName string, subjects []rbacv1.Subject) error

	IsLeaseExpired(leaseID string) (bool, error)

	RevokeLease(leaseID string) error
}
