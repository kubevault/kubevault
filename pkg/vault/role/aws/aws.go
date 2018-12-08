package aws

import (
	"encoding/json"

	api "github.com/kubevault/operator/apis/secretengine/v1alpha1"
	"github.com/kubevault/operator/pkg/vault"
	"github.com/kubevault/operator/pkg/vault/role"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	appcat "kmodules.xyz/custom-resources/apis/appcatalog/v1alpha1"
	appcat_cs "kmodules.xyz/custom-resources/client/clientset/versioned/typed/appcatalog/v1alpha1"
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
	vClient, err := vault.NewClient(kClient, appClient, role.Spec.AuthManagerRef)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create vault api client")
	}

	awsPath, err := GetAWSPath(appClient, role.Spec.AuthManagerRef)
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

// If aws path does not exist, then use default database path
func GetAWSPath(c appcat_cs.AppcatalogV1alpha1Interface, ref *appcat.AppReference) (string, error) {
	vApp, err := c.AppBindings(ref.Namespace).Get(ref.Name, metav1.GetOptions{})
	if err != nil {
		return "", err
	}

	var cf struct {
		AWSPath string `json:"aws_path,omitempty"`
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
