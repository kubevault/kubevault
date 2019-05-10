package azure

import (
	"fmt"

	vaultapi "github.com/hashicorp/vault/api"
	api "github.com/kubevault/operator/apis/engine/v1alpha1"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type AzureRole struct {
	azureRole   *api.AzureRole
	vaultClient *vaultapi.Client
	kubeClient  kubernetes.Interface
	azurePath   string // Specifies the path where azure is enabled
}

// checks whether azure is enabled or not
func (a *AzureRole) IsAzureEnabled() (bool, error) {
	mnt, err := a.vaultClient.Sys().ListMounts()
	if err != nil {
		return false, errors.Wrap(err, "failed to list mounted secrets engines")
	}

	mntPath := a.azurePath + "/"
	for k := range mnt {
		if k == mntPath {
			return true, nil
		}
	}
	return false, nil
}

// EnableAzure enables azure secret engine
// It first checks whether azure is enabled or not
func (a *AzureRole) EnableAzure() error {
	enabled, err := a.IsAzureEnabled()
	if err != nil {
		return err
	}

	if enabled {
		return nil
	}

	err = a.vaultClient.Sys().Mount(a.azurePath, &vaultapi.MountInput{
		Type: "azure",
	})
	if err != nil {
		return err
	}
	return nil
}

// ref:
//	- https://www.vaultproject.io/api/secret/azure/index.html#configure-access

// Writes the config file to specified path
func (a *AzureRole) CreateConfig() error {
	if a.vaultClient == nil {
		return errors.New("vault client is nil")
	}
	if a.azureRole == nil {
		return errors.New("AzureRole is nil")
	}
	if a.azurePath == "" {
		return errors.New("azure engine path is empty")
	}

	path := fmt.Sprintf("/v1/%s/config", a.azurePath)
	req := a.vaultClient.NewRequest("POST", path)

	payload := map[string]interface{}{}
	cfg := a.azureRole.Spec.Config
	if cfg == nil {
		return errors.New("azure secret engine config is nil")
	}

	if cfg.SubscriptionID != "" {
		payload["subscription_id"] = cfg.SubscriptionID
	} else {
		return errors.New("azure secret engine configuration failed: subscription id missing")
	}

	if cfg.TenantID != "" {
		payload["tenant_id"] = cfg.TenantID
	} else {
		return errors.New("azure secret engine configuration failed: tenant id missing")
	}

	if cfg.ClientID != "" {
		payload["client_id"] = cfg.ClientID
	}

	if cfg.Environment != "" {
		payload["environment"] = cfg.Environment
	}
	if cfg.ClientSecret != "" {
		sr, err := a.kubeClient.CoreV1().Secrets(a.azureRole.Namespace).Get(cfg.ClientSecret, metav1.GetOptions{})
		if err != nil {
			return errors.Wrap(err, "failed to get azure client secret")
		}

		if val, ok := sr.Data[api.AzureClientSecret]; ok {
			payload["client_secret"] = string(val)
		}

	}

	if err := req.SetJSONBody(payload); err != nil {
		return errors.Wrap(err, "failed to load payload in config create request")
	}

	_, err := a.vaultClient.RawRequest(req)
	if err != nil {
		return errors.Wrap(err, "failed to create azure config")
	}
	return nil
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
