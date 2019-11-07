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

package policybinding

import (
	"fmt"

	api "kubevault.dev/operator/apis/policy/v1alpha1"
	cs "kubevault.dev/operator/client/clientset/versioned"
	"kubevault.dev/operator/pkg/vault"

	vaultapi "github.com/hashicorp/vault/api"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	appcat "kmodules.xyz/custom-resources/apis/appcatalog/v1alpha1"
	appcat_cs "kmodules.xyz/custom-resources/client/clientset/versioned/typed/appcatalog/v1alpha1"
)

type PolicyBinding interface {
	// create or update policy binding
	Ensure(name string) error
	// delete policy binding
	Delete(name string) error
}

func NewPolicyBindingClient(c cs.Interface, appc appcat_cs.AppcatalogV1alpha1Interface, kc kubernetes.Interface, pBind *api.VaultPolicyBinding) (PolicyBinding, error) {
	if pBind == nil {
		return nil, errors.New("VaultPolicyBinding is nil")
	}
	if len(pBind.Spec.Policies) == 0 {
		return nil, errors.New(".spec.policies must be non empty")
	}
	pb := &pBinding{}
	if pBind.Spec.Kubernetes != nil {
		pb.saNames = pBind.Spec.Kubernetes.ServiceAccountNames
		pb.saNamespaces = pBind.Spec.Kubernetes.ServiceAccountNamespaces
		pb.ttl = pBind.Spec.Kubernetes.TTL
		pb.maxTTL = pBind.Spec.Kubernetes.MaxTTL
		pb.period = pBind.Spec.Kubernetes.Period
		pb.path = pBind.Spec.Kubernetes.Path
		pb.setKubernetesDefaults()
	}

	// check whether VaultPolicy exists
	for _, pIdentifier := range pBind.Spec.Policies {
		var policyName string
		if pIdentifier.Ref != "" {
			// pIdentifier.Ref species the policy crd name
			policy, err := c.PolicyV1alpha1().VaultPolicies(pBind.Namespace).Get(pIdentifier.Ref, metav1.GetOptions{})
			if err != nil {
				return nil, errors.Wrap(err, "for .spec.policies")
			}
			policyName = policy.PolicyName()
		} else {
			// pIdentifier.Name specifies the vault policy name
			// If anyone wants to access a policy crd through this field
			// they need to follow the naming format: k8s.<cluster_name>.<namespace_name>.<policy_name>
			policyName = pIdentifier.Name
		}

		pb.policies = append(pb.policies, policyName)
	}
	var err error
	if pBind.Spec.VaultRef.Name == "" {
		return nil, errors.New("spec.vaultRef must not be empty")
	}

	vAppRef := &appcat.AppReference{
		Namespace: pBind.Namespace,
		Name:      pBind.Spec.VaultRef.Name,
	}
	pb.vClient, err = vault.NewClient(kc, appc, vAppRef)
	if err != nil {
		return nil, err
	}
	return pb, nil
}

type pBinding struct {
	vClient      *vaultapi.Client
	policies     []string
	saNames      []string
	saNamespaces []string
	ttl          string
	maxTTL       string
	period       string
	path         string
}

func (p *pBinding) setKubernetesDefaults() {
	if p.path == "" {
		p.path = "kubernetes"
	}
}

// create or update policy binding
// it's safe to call it multiple times
func (p *pBinding) Ensure(name string) error {
	path := fmt.Sprintf("/v1/auth/%s/role/%s", p.path, name)
	req := p.vClient.NewRequest("POST", path)
	payload := map[string]interface{}{
		"bound_service_account_names":      p.saNames,
		"bound_service_account_namespaces": p.saNamespaces,
		"policies":                         p.policies,
		"ttl":                              p.ttl,
		"max_ttl":                          p.maxTTL,
		"period":                           p.period,
	}

	err := req.SetJSONBody(payload)
	if err != nil {
		return err
	}

	_, err = p.vClient.RawRequest(req)
	if err != nil {
		return err
	}
	return nil
}

// delete policy binding
// it's safe to call it, even if 'name' doesn't exist in vault
func (p *pBinding) Delete(name string) error {
	path := fmt.Sprintf("/v1/auth/%s/role/%s", p.path, name)
	req := p.vClient.NewRequest("DELETE", path)
	_, err := p.vClient.RawRequest(req)
	if err != nil {
		return err
	}
	return nil
}
