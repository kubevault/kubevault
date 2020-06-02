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

package approle

import (
	"fmt"

	api "kubevault.dev/operator/apis/approle/v1alpha1"
	"kubevault.dev/operator/pkg/vault"

	vaultapi "github.com/hashicorp/vault/api"
	"github.com/pkg/errors"
	"k8s.io/client-go/kubernetes"
	appcat "kmodules.xyz/custom-resources/apis/appcatalog/v1alpha1"
	appcat_cs "kmodules.xyz/custom-resources/client/clientset/versioned/typed/appcatalog/v1alpha1"
)

type AppRole interface {
	EnsureAppRole(name string, approle map[string]interface{}) error
	DeleteAppRole(name string) error
}

type vAppRole struct {
	client *vaultapi.Client
}

func NewAppRoleClientForVault(kc kubernetes.Interface, appc appcat_cs.AppcatalogV1alpha1Interface, p *api.VaultAppRole) (AppRole, error) {
	if p == nil {
		return nil, errors.New("VaultAppRole is nil")
	}
	vAppRef := &appcat.AppReference{
		Namespace: p.Namespace,
		Name:      p.Spec.VaultRef.Name,
	}
	vc, err := vault.NewClient(kc, appc, vAppRef)
	if err != nil {
		return nil, err
	}

	return &vAppRole{
		client: vc,
	}, nil
}

// EnsureAppRole creates or updates the approle
// it's safe to call multiple times.
// https://www.vaultproject.io/api-docs/auth/approle#create-custom-approle-secret-id
func (v *vAppRole) EnsureAppRole(name string, payload map[string]interface{}) error {
	_, err := v.client.Logical().Write(fmt.Sprintf("auth/approle/role/%s", name), payload)
	return err
}

// Delete deletes the approle
func (v *vAppRole) DeleteAppRole(name string) error {
	_, err := v.client.Logical().Delete(fmt.Sprintf("auth/approle/role/%s", name))
	return err
}
