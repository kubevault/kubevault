package gcp

import (
	"fmt"

	vaultapi "github.com/hashicorp/vault/api"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	api "kubevault.dev/operator/apis/engine/v1alpha1"
)

type GCPRole struct {
	gcpRole     *api.GCPRole
	vaultClient *vaultapi.Client
	kubeClient  kubernetes.Interface
	gcpPath     string // Specifies the path where gcp is enabled
}

const (
	GCPSecretType       string = "secret_type"
	GCPOAuthTokenScopes string = "token_scopes"
)

// checks whether gcp is enabled or not
func (a *GCPRole) IsGCPEnabled() (bool, error) {
	mnt, err := a.vaultClient.Sys().ListMounts()
	if err != nil {
		return false, errors.Wrap(err, "failed to list mounted secrets engines")
	}

	mntPath := a.gcpPath + "/"
	for k := range mnt {
		if k == mntPath {
			return true, nil
		}
	}
	return false, nil
}

// EnableGCP enables gcp secret engine
// It first checks whether gcp is enabled or not
func (a *GCPRole) EnableGCP() error {
	enabled, err := a.IsGCPEnabled()
	if err != nil {
		return err
	}

	if enabled {
		return nil
	}

	err = a.vaultClient.Sys().Mount(a.gcpPath, &vaultapi.MountInput{
		Type: "gcp",
	})
	if err != nil {
		return err
	}
	return nil
}

// https://www.vaultproject.io/api/secret/gcp/index.html#write-config
// Writes the config file to specified path
func (a *GCPRole) CreateConfig() error {
	if a.vaultClient == nil {
		return errors.New("vault client is nil")
	}
	if a.gcpRole == nil {
		return errors.New("GCPRole is nil")
	}
	if a.gcpPath == "" {
		return errors.New("gcp engine path is empty")
	}

	path := fmt.Sprintf("/v1/%s/config", a.gcpPath)
	req := a.vaultClient.NewRequest("POST", path)

	payload := map[string]interface{}{}
	cfg := a.gcpRole.Spec.Config
	if cfg == nil {
		return errors.New("gcp secret engine config is nil")
	}
	if cfg.TTL != "" {
		payload["ttl"] = cfg.TTL
	}
	if cfg.MaxTTL != "" {
		payload["max_ttl"] = cfg.MaxTTL
	}

	if cfg.CredentialSecret != "" {
		sr, err := a.kubeClient.CoreV1().Secrets(a.gcpRole.Namespace).Get(cfg.CredentialSecret, metav1.GetOptions{})
		if err != nil {
			return errors.Wrap(err, "failed to get gcp credential secret")
		}

		if val, ok := sr.Data[api.GCPSACredentialJson]; ok {
			payload["credentials"] = string(val)
		}

	}

	if err := req.SetJSONBody(payload); err != nil {
		return errors.Wrap(err, "failed to load payload in config create request")
	}

	_, err := a.vaultClient.RawRequest(req)
	if err != nil {
		return errors.Wrap(err, "failed to create gcp config")
	}
	return nil
}

// Links:
// - https://www.vaultproject.io/api/secret/gcp/index.html#create-update-roleset
// Creates roleset
func (a *GCPRole) CreateRole() error {
	if a.vaultClient == nil {
		return errors.New("vault client is nil")
	}
	if a.gcpRole == nil {
		return errors.New("GCPRole is nil")
	}
	if a.gcpPath == "" {
		return errors.New("gcp engine path is empty")
	}

	path := fmt.Sprintf("/v1/%s/roleset/%s", a.gcpPath, a.gcpRole.RoleName())
	req := a.vaultClient.NewRequest("POST", path)

	roleSpec := a.gcpRole.Spec
	payload := map[string]interface{}{
		"project":  roleSpec.Project,
		"bindings": roleSpec.Bindings,
	}
	if roleSpec.SecretType != "" {
		payload[GCPSecretType] = roleSpec.SecretType
	}

	if roleSpec.TokenScopes != nil {
		payload[GCPOAuthTokenScopes] = roleSpec.TokenScopes
	}

	if err := req.SetJSONBody(payload); err != nil {
		return errors.Wrap(err, "failed to load payload in gcp create role request")
	}

	_, err := a.vaultClient.RawRequest(req)
	if err != nil {
		return errors.Wrap(err, "failed to create gcp role")
	}
	return nil
}

// DeleteRole deletes role
// It's safe to call multiple time. It doesn't give
// error even if respective role doesn't exist
func (a *GCPRole) DeleteRole(name string) error {
	path := fmt.Sprintf("/v1/%s/roleset/%s", a.gcpPath, name)
	req := a.vaultClient.NewRequest("DELETE", path)

	_, err := a.vaultClient.RawRequest(req)
	if err != nil {
		return errors.Wrapf(err, "failed to delete gcp role %s", name)
	}
	return nil
}
