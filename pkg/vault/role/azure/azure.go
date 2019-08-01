package azure

import (
	"encoding/json"

	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	appcat_cs "kmodules.xyz/custom-resources/client/clientset/versioned/typed/appcatalog/v1alpha1"
	api "kubevault.dev/operator/apis/engine/v1alpha1"
	"kubevault.dev/operator/pkg/vault"
	"kubevault.dev/operator/pkg/vault/role"
)

const DefaultAzurePath = "azure"

type AzureRoleInterface interface {
	role.RoleInterface

	// Enable enables azure secret engine
	EnableAzure() error

	// IsAzureEnabled checks whether azure is enabled or not
	IsAzureEnabled() (bool, error)

	// DeleteRole deletes role
	DeleteRole(name string) error
}

func NewAzureRole(kClient kubernetes.Interface, appClient appcat_cs.AppcatalogV1alpha1Interface, role *api.AzureRole) (AzureRoleInterface, error) {
	vClient, err := vault.NewClient(kClient, appClient, role.Namespace, role.Spec.VaultRef)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create vault api client")
	}

	azurePath, err := GetAzurePath(appClient, role)
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
func GetAzurePath(c appcat_cs.AppcatalogV1alpha1Interface, role *api.AzureRole) (string, error) {
	vApp, err := c.AppBindings(role.Namespace).Get(role.Spec.VaultRef.Name, metav1.GetOptions{})
	if err != nil {
		return "", err
	}

	var cf struct {
		AzurePath string `json:"azurePath,omitempty"`
	}

	if vApp.Spec.Parameters != nil {
		err := json.Unmarshal(vApp.Spec.Parameters.Raw, &cf)
		if err != nil {
			return "", err
		}
	}

	if cf.AzurePath != "" {
		return cf.AzurePath, nil
	}
	return DefaultAzurePath, nil
}
