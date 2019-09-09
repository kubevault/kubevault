package azure

import (
	"fmt"

	vaultapi "github.com/hashicorp/vault/api"
	"github.com/pkg/errors"
	"k8s.io/client-go/kubernetes"
	api "kubevault.dev/operator/apis/engine/v1alpha1"
)

type AzureRole struct {
	azureRole   *api.AzureRole
	vaultClient *vaultapi.Client
	kubeClient  kubernetes.Interface
	azurePath   string // Specifies the path where azure is enabled
}

// ref:
//	- https://www.vaultproject.io/api/secret/azure/index.html#create-update-role

// Creates role
func (a *AzureRole) CreateRole() error {
	if a.vaultClient == nil {
		return errors.New("vault client is nil")
	}
	if a.azureRole == nil {
		return errors.New("AzureRole is nil")
	}
	if a.azurePath == "" {
		return errors.New("azure engine path is empty")
	}

	path := fmt.Sprintf("/v1/%s/roles/%s", a.azurePath, a.azureRole.RoleName())
	req := a.vaultClient.NewRequest("POST", path)

	roleSpec := a.azureRole.Spec
	payload := map[string]interface{}{}

	if roleSpec.AzureRoles != "" {
		payload["azure_roles"] = roleSpec.AzureRoles
	}

	if roleSpec.ApplicationObjectID != "" {
		payload["application_object_id"] = roleSpec.ApplicationObjectID
	}

	if roleSpec.TTL != "" {
		payload["ttl"] = roleSpec.TTL
	}

	if roleSpec.MaxTTL != "" {
		payload["max_ttl"] = roleSpec.MaxTTL
	}

	if err := req.SetJSONBody(payload); err != nil {
		return errors.Wrap(err, "failed to load payload in azure create role request")
	}

	_, err := a.vaultClient.RawRequest(req)
	if err != nil {
		return errors.Wrap(err, "failed to create azure role")
	}
	return nil
}

// DeleteRole deletes role
// It's safe to call multiple time. It doesn't give
// error even if respective role doesn't exist
func (a *AzureRole) DeleteRole(name string) error {
	path := fmt.Sprintf("/v1/%s/roles/%s", a.azurePath, name)
	req := a.vaultClient.NewRequest("DELETE", path)

	_, err := a.vaultClient.RawRequest(req)
	if err != nil {
		return errors.Wrapf(err, "failed to delete azure role %s", name)
	}
	return nil
}
