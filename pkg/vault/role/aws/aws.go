package aws

import (
	"encoding/json"

	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	appcat "kmodules.xyz/custom-resources/apis/appcatalog/v1alpha1"
	appcat_cs "kmodules.xyz/custom-resources/client/clientset/versioned/typed/appcatalog/v1alpha1"
	api "kubevault.dev/operator/apis/engine/v1alpha1"
	"kubevault.dev/operator/pkg/vault"
	"kubevault.dev/operator/pkg/vault/role"
)

const DefaultAWSPath = "aws"

type AWSRoleInterface interface {
	role.RoleInterface

	// Enable enables aws secret engine
	EnableAWS() error

	// IsAWSEnabled checks whether aws is enabled or not
	IsAWSEnabled() (bool, error)

	// DeleteRole deletes role
	DeleteRole(name string) error
}

func NewAWSRole(kClient kubernetes.Interface, appClient appcat_cs.AppcatalogV1alpha1Interface, role *api.AWSRole) (AWSRoleInterface, error) {
	vClient, err := vault.NewClient(kClient, appClient, role.Spec.VaultRef)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create vault api client")
	}

	awsPath, err := GetAWSPath(appClient, role.Spec.VaultRef)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get aws path")
	}
	return &AWSRole{
		kubeClient:  kClient,
		vaultClient: vClient,
		awsRole:     role,
		awsPath:     awsPath,
	}, nil

}

// If aws path does not exist, then use default aws path
func GetAWSPath(c appcat_cs.AppcatalogV1alpha1Interface, ref *appcat.AppReference) (string, error) {
	vApp, err := c.AppBindings(ref.Namespace).Get(ref.Name, metav1.GetOptions{})
	if err != nil {
		return "", err
	}

	var cf struct {
		AWSPath string `json:"awsPath,omitempty"`
	}

	if vApp.Spec.Parameters != nil {
		err := json.Unmarshal(vApp.Spec.Parameters.Raw, &cf)
		if err != nil {
			return "", err
		}
	}

	if cf.AWSPath != "" {
		return cf.AWSPath, nil
	}
	return DefaultAWSPath, nil
}
