package gcp

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

const DefaultGCPPath = "gcp"

type GCPRoleInterface interface {
	role.RoleInterface

	// Enable enables gcp secret engine
	EnableGCP() error

	// IsGCPEnabled checks whether gcp is enabled or not
	IsGCPEnabled() (bool, error)

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

	gcpPath, err := GetGCPPath(appClient, vAppRef)
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
func GetGCPPath(c appcat_cs.AppcatalogV1alpha1Interface, vAppRef *appcat.AppReference) (string, error) {
	vApp, err := c.AppBindings(vAppRef.Namespace).Get(vAppRef.Name, metav1.GetOptions{})
	if err != nil {
		return "", err
	}

	var cf struct {
		GCPPath string `json:"gcpPath,omitempty"`
	}

	if vApp.Spec.Parameters != nil {
		err := json.Unmarshal(vApp.Spec.Parameters.Raw, &cf)
		if err != nil {
			return "", err
		}
	}

	if cf.GCPPath != "" {
		return cf.GCPPath, nil
	}
	return DefaultGCPPath, nil
}
