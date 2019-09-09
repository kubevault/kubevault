package gcp

import (
	"github.com/pkg/errors"
	"k8s.io/client-go/kubernetes"
	appcat "kmodules.xyz/custom-resources/apis/appcatalog/v1alpha1"
	appcat_cs "kmodules.xyz/custom-resources/client/clientset/versioned/typed/appcatalog/v1alpha1"
	api "kubevault.dev/operator/apis/engine/v1alpha1"
	"kubevault.dev/operator/pkg/vault"
	"kubevault.dev/operator/pkg/vault/role"
)

const DefaultGCPPath = "gcp"

type GCPRoleInterface interface {
	role.RoleInterface

	// DeleteRole deletes role
	DeleteRole(name string) error
}

func NewGCPRole(kClient kubernetes.Interface, appClient appcat_cs.AppcatalogV1alpha1Interface, role *api.GCPRole) (GCPRoleInterface, error) {
	vAppRef := &appcat.AppReference{
		Namespace: role.Namespace,
		Name:      role.Spec.VaultRef.Name,
	}
	vClient, err := vault.NewClient(kClient, appClient, vAppRef)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create vault api client")
	}

	gcpPath, err := GetGCPPath(role)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get gcp path")
	}
	return &GCPRole{
		kubeClient:  kClient,
		vaultClient: vClient,
		gcpRole:     role,
		gcpPath:     gcpPath,
	}, nil
}

// If gcp path does not exist, then use default gcp path
func GetGCPPath(role *api.GCPRole) (string, error) {
	if role.Spec.Path != "" {
		return role.Spec.Path, nil
	}
	return DefaultGCPPath, nil
}
