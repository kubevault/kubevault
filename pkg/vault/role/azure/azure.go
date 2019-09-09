package azure

import (
	"github.com/pkg/errors"
	"k8s.io/client-go/kubernetes"
	appcat "kmodules.xyz/custom-resources/apis/appcatalog/v1alpha1"
	appcat_cs "kmodules.xyz/custom-resources/client/clientset/versioned/typed/appcatalog/v1alpha1"
	api "kubevault.dev/operator/apis/engine/v1alpha1"
	"kubevault.dev/operator/pkg/vault"
	"kubevault.dev/operator/pkg/vault/role"
)

const DefaultAzurePath = "azure"

type AzureRoleInterface interface {
	role.RoleInterface

	// DeleteRole deletes role
	DeleteRole(name string) error
}

func NewAzureRole(kClient kubernetes.Interface, appClient appcat_cs.AppcatalogV1alpha1Interface, role *api.AzureRole) (AzureRoleInterface, error) {
	vAppRef := &appcat.AppReference{
		Namespace: role.Namespace,
		Name:      role.Spec.VaultRef.Name,
	}

	vClient, err := vault.NewClient(kClient, appClient, vAppRef)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create vault api client")
	}

	azurePath, err := GetAzurePath(role)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get azure path")
	}
	return &AzureRole{
		kubeClient:  kClient,
		vaultClient: vClient,
		azureRole:   role,
		azurePath:   azurePath,
	}, nil
}

// If azure path does not exist, then use default azure path
func GetAzurePath(role *api.AzureRole) (string, error) {
	if role.Spec.Path != "" {
		return role.Spec.Path, nil
	}
	return DefaultAzurePath, nil
}
