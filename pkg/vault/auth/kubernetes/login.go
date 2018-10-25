package kubernetes

import (
	"encoding/json"
	"fmt"

	vaultapi "github.com/hashicorp/vault/api"
	config "github.com/kubevault/operator/apis/config/v1alpha1"
	"github.com/kubevault/operator/pkg/vault/auth/types"
	vaultuitl "github.com/kubevault/operator/pkg/vault/util"
	"github.com/pkg/errors"
	core "k8s.io/api/core/v1"
	appcat "kmodules.xyz/custom-resources/apis/appcatalog/v1alpha1"
)

type auth struct {
	vClient *vaultapi.Client
	jwt     string
	role    string
	path    string
}

func New(vApp *appcat.AppBinding, secret *core.Secret) (*auth, error) {
	if vApp.Spec.Parameters == nil {
		return nil, errors.New("parameters are not provided")
	}

	cfg, err := vaultuitl.VaultConfigFromAppBinding(vApp)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create vault config from AppBinding")
	}

	vc, err := vaultapi.NewClient(cfg)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create vault client")
	}

	jwt, ok := secret.Data[core.ServiceAccountTokenKey]
	if !ok {
		return nil, errors.New("jwt is missing")
	}

	var cf config.KubernetesAuthConfiguration
	raw, err := vaultuitl.UnQuoteJson(string(vApp.Spec.Parameters.Raw))
	if err != nil {
		return nil, errors.Wrap(err, "failed to unquote json")
	}
	err = json.Unmarshal([]byte(raw), &cf)
	if err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal parameters")
	}
	cf.SetDefaults()

	return &auth{
		vClient: vc,
		jwt:     string(jwt),
		role:    cf.Role,
		path:    cf.AuthPath,
	}, nil
}

// Login will log into vault and return client token
func (a *auth) Login() (string, error) {
	path := fmt.Sprintf("/v1/auth/%s/login", a.path)
	req := a.vClient.NewRequest("POST", path)
	payload := map[string]interface{}{
		"jwt":  a.jwt,
		"role": a.role,
	}
	if err := req.SetJSONBody(payload); err != nil {
		return "", err
	}

	resp, err := a.vClient.RawRequest(req)
	if err != nil {
		return "", err
	}

	var loginResp types.AuthLoginResponse
	defer resp.Body.Close()
	err = json.NewDecoder(resp.Body).Decode(&loginResp)
	if err != nil {
		return "", err
	}
	return loginResp.Auth.ClientToken, nil
}
