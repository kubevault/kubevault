/*
Copyright AppsCode Inc. and Contributors

Licensed under the AppsCode Community License 1.0.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    https://github.com/appscode/licenses/raw/1.0.0/AppsCode-Community-1.0.0.md

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package userpass

import (
	"encoding/json"
	"fmt"

	vsapi "kubevault.dev/apimachinery/apis/kubevault/v1alpha1"
	"kubevault.dev/operator/pkg/vault/auth/types"
	authtype "kubevault.dev/operator/pkg/vault/auth/types"
	vaultuitl "kubevault.dev/operator/pkg/vault/util"

	vaultapi "github.com/hashicorp/vault/api"
	"github.com/pkg/errors"
	core "k8s.io/api/core/v1"
)

type auth struct {
	vClient *vaultapi.Client
	user    string
	pass    string
	path    string
}

func New(authInfo *authtype.AuthInfo) (*auth, error) {
	if authInfo == nil {
		return nil, errors.New("authentication information is empty")
	}
	if authInfo.VaultApp == nil {
		return nil, errors.New("AppBinding is empty")
	}

	cfg, err := vaultuitl.VaultConfigFromAppBinding(authInfo.VaultApp)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create vault config from AppBinding")
	}

	vc, err := vaultapi.NewClient(cfg)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create vault client")
	}

	if authInfo.Secret == nil {
		return nil, errors.New("authentication secret is missing")
	}
	secret := authInfo.Secret
	user, ok := secret.Data[core.BasicAuthUsernameKey]
	if !ok {
		return nil, errors.New("username is missing")
	}
	pass, ok := secret.Data[core.BasicAuthPasswordKey]
	if !ok {
		return nil, errors.New("password is missing")
	}

	authPath := string(vsapi.AuthTypeUserPass)
	if authInfo.Path != "" {
		authPath = authInfo.Path
	}

	return &auth{
		vClient: vc,
		user:    string(user),
		pass:    string(pass),
		path:    authPath,
	}, nil
}

// Login will log into vault and return client token
func (a *auth) Login() (string, error) {
	path := fmt.Sprintf("/v1/auth/%s/login/%s", a.path, a.user)
	req := a.vClient.NewRequest("POST", path)
	payload := map[string]interface{}{
		"password": a.pass,
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
