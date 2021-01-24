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

package serviceaccount

import (
	"encoding/json"
	"fmt"
	"time"

	vsapi "kubevault.dev/apimachinery/apis/kubevault/v1alpha1"
	sa_util "kubevault.dev/operator/pkg/util"
	"kubevault.dev/operator/pkg/vault/auth/types"
	authtype "kubevault.dev/operator/pkg/vault/auth/types"
	vaultuitl "kubevault.dev/operator/pkg/vault/util"

	vaultapi "github.com/hashicorp/vault/api"
	"github.com/pkg/errors"
	core "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
)

const (
	timeout      = 30 * time.Second
	timeInterval = 2 * time.Second
)

type auth struct {
	vClient *vaultapi.Client
	jwt     string
	role    string
	path    string
}

func New(kc kubernetes.Interface, authInfo *authtype.AuthInfo) (*auth, error) {
	if authInfo == nil {
		return nil, errors.New("authentication information is empty")
	}

	if authInfo.VaultApp == nil {
		return nil, errors.New("AppBinding is empty")
	}

	vApp := authInfo.VaultApp
	cfg, err := vaultuitl.VaultConfigFromAppBinding(vApp)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create vault config from AppBinding")
	}

	vc, err := vaultapi.NewClient(cfg)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create vault client")
	}
	if authInfo.ServiceAccountRef == nil {
		return nil, errors.New("service account reference is empty")
	}

	sa := authInfo.ServiceAccountRef
	if sa.Name == "" || sa.Namespace == "" {
		return nil, errors.New("name or namespace is missing in service account reference")
	}

	secret, err := sa_util.TryGetJwtTokenSecretNameFromServiceAccount(kc, sa.Name, sa.Namespace, timeInterval, timeout)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get jwt token secret of service account %s/%s", sa.Namespace, sa.Name)
	}

	jwt, ok := secret.Data[core.ServiceAccountTokenKey]
	if !ok {
		return nil, errors.New("jwt is missing")
	}

	if authInfo.Path == "" {
		authInfo.Path = string(vsapi.AuthTypeKubernetes)
	}
	if authInfo.VaultRole == "" {
		return nil, errors.Wrap(err, "VaultRole is empty")
	}

	return &auth{
		vClient: vc,
		jwt:     string(jwt),
		role:    authInfo.VaultRole,
		path:    authInfo.Path,
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
