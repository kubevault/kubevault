/*
Copyright The KubeVault Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package gcp

import (
	"fmt"

	api "kubevault.dev/operator/apis/engine/v1alpha1"

	vaultapi "github.com/hashicorp/vault/api"
	"github.com/pkg/errors"
	"k8s.io/client-go/kubernetes"
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
