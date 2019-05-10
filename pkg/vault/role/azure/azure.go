package azure

import (
	"encoding/json"

	api "github.com/kubevault/operator/apis/engine/v1alpha1"
	"github.com/kubevault/operator/pkg/vault"
	"github.com/kubevault/operator/pkg/vault/role"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	appcat "kmodules.xyz/custom-resources/apis/appcatalog/v1alpha1"
	appcat_cs "kmodules.xyz/custom-resources/client/clientset/versioned/typed/appcatalog/v1alpha1"
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
	vClient, err := vault.NewClient(kClient, appClient, role.Spec.AuthManagerRef)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create vault api client")
	}

	azurePath, err := GetAzurePath(appClient, role.Spec.AuthManagerRef)
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
func GetAzurePath(c appcat_cs.AppcatalogV1alpha1Interface, ref *appcat.AppReference) (string, error) {
	vApp, err := c.AppBindings(ref.Namespace).Get(ref.Name, metav1.GetOptions{})
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
