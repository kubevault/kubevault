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

package policy

import (
	api "kubevault.dev/operator/apis/policy/v1alpha1"
	"kubevault.dev/operator/pkg/vault"

	vaultapi "github.com/hashicorp/vault/api"
	"github.com/pkg/errors"
	"k8s.io/client-go/kubernetes"
	appcat "kmodules.xyz/custom-resources/apis/appcatalog/v1alpha1"
	appcat_cs "kmodules.xyz/custom-resources/client/clientset/versioned/typed/appcatalog/v1alpha1"
)

type Policy interface {
	EnsurePolicy(name string, policy string) error
	DeletePolicy(name string) error
}

type vPolicy struct {
	client *vaultapi.Client
}

func NewPolicyClientForVault(kc kubernetes.Interface, appc appcat_cs.AppcatalogV1alpha1Interface, p *api.VaultPolicy) (Policy, error) {
	if p == nil {
		return nil, errors.New("VaultPolicy is nil")
	}
	vAppRef := &appcat.AppReference{
		Namespace: p.Namespace,
		Name:      p.Spec.VaultRef.Name,
	}
	vc, err := vault.NewClient(kc, appc, vAppRef)
	if err != nil {
		return nil, err
	}

	return &vPolicy{
		client: vc,
	}, nil
}

// EnsurePolicy creates or updates the policy
// it's safe to call multiple times.
// https://www.vaultproject.io/api/system/policy.html#create-update-policy
func (v *vPolicy) EnsurePolicy(name string, policy string) error {
	return v.client.Sys().PutPolicy(name, policy)
}

// Delete deletes the policy
func (v *vPolicy) DeletePolicy(name string) error {
	return v.client.Sys().DeletePolicy(name)
}
