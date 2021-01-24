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
	"context"

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
	Ensure(pBind *api.VaultPolicyBinding) error
	// delete policy binding
	Delete(pBind *api.VaultPolicyBinding) error
}

func NewPolicyBindingClient(c cs.Interface, appc appcat_cs.AppcatalogV1alpha1Interface, kc kubernetes.Interface, pBind *api.VaultPolicyBinding) (PolicyBinding, error) {
	if pBind == nil {
		return nil, errors.New("VaultPolicyBinding is nil")
	}
	if len(pBind.Spec.Policies) == 0 {
		return nil, errors.New(".spec.policies must be non empty")
	}

	if pBind.Spec.SubjectRef.Kubernetes == nil && pBind.Spec.SubjectRef.AppRole == nil {
		return nil, errors.New(".spec.policies.subjectRef must be non empty")
	}

	// set def values
	pBind.SetDefaults()

	pb := &pBinding{}
	// kubernetes auth
	if pBind.Spec.Kubernetes != nil {
		pb.authKubernetes = &pBindingKubernetes{
			SaNames:      pBind.Spec.Kubernetes.ServiceAccountNames,
			SaNamespaces: pBind.Spec.Kubernetes.ServiceAccountNamespaces,
			TokenTTL:     pBind.Spec.Kubernetes.TTL,
			TokenMaxTTL:  pBind.Spec.Kubernetes.MaxTTL,
			TokenPeriod:  pBind.Spec.Kubernetes.Period,
			name:         pBind.Spec.Kubernetes.Name,
			path:         pBind.Spec.Kubernetes.Path,
		}
	}

	if pBind.Spec.AppRole != nil {
		pb.authAppRole = &pBindingAppRole{
			BindSecretID:         pBind.Spec.AppRole.BindSecretID,
			SecretIDBoundCidrs:   pBind.Spec.AppRole.SecretIDBoundCidrs,
			SecretIDNumUses:      pBind.Spec.AppRole.SecretIDNumUses,
			SecretIDTTL:          pBind.Spec.AppRole.SecretIDTTL,
			EnableLocalSecretIDs: pBind.Spec.AppRole.EnableLocalSecretIDs,
			TokenTTL:             pBind.Spec.AppRole.TokenTTL,
			TokenMaxTTL:          pBind.Spec.AppRole.TokenMaxTTL,
			TokenBoundCidrs:      pBind.Spec.AppRole.TokenBoundCidrs,
			TokenExplicitMaxTTL:  pBind.Spec.AppRole.TokenExplicitMaxTTL,
			TokenNoDefaultPolicy: pBind.Spec.AppRole.TokenNoDefaultPolicy,
			TokenNumUses:         pBind.Spec.AppRole.TokenNumUses,
			TokenPeriod:          pBind.Spec.AppRole.TokenPeriod,
			TokenType:            pBind.Spec.AppRole.TokenType,
			roleName:             pBind.Spec.AppRole.RoleName,
			path:                 pBind.Spec.AppRole.Path,
		}
	}

	// check whether VaultPolicy exists
	for _, pIdentifier := range pBind.Spec.Policies {
		var policyName string
		if pIdentifier.Ref != "" {
			// pIdentifier.Ref species the policy crd name
			policy, err := c.PolicyV1alpha1().VaultPolicies(pBind.Namespace).Get(context.TODO(), pIdentifier.Ref, metav1.GetOptions{})
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
	vClient        *vaultapi.Client
	policies       []string
	authKubernetes *pBindingKubernetes
	authAppRole    *pBindingAppRole
}

type pBindingKubernetes struct {
	path          string
	name          string
	SaNames       []string `json:"bound_service_account_names,omitempty"`
	SaNamespaces  []string `json:"bound_service_account_namespaces,omitempty"`
	TokenTTL      string   `json:"token_ttl,omitempty"`
	TokenPolicies []string `json:"token_policies,omitempty"`
	TokenMaxTTL   string   `json:"token_max_ttl,omitempty"`
	TokenPeriod   string   `json:"token_period,omitempty"`
}

type pBindingAppRole struct {
	path                 string
	roleName             string
	BindSecretID         bool     `json:"bind_secret_id"`
	SecretIDBoundCidrs   []string `json:"secret_id_bound_cidrs,omitempty"`
	SecretIDNumUses      int64    `json:"secret_id_num_uses,omitempty"`
	SecretIDTTL          string   `json:"secret_id_ttl,omitempty"`
	EnableLocalSecretIDs bool     `json:"enable_local_secret_ids,omitempty"`
	TokenTTL             int64    `json:"token_ttl,omitempty"`
	TokenMaxTTL          int64    `json:"token_max_ttl,omitempty"`
	TokenPolicies        []string `json:"token_policies,omitempty"`
	TokenBoundCidrs      []string `json:"token_bound_cidrs,omitempty"`
	TokenExplicitMaxTTL  int64    `json:"token_explicit_max_ttl,omitempty"`
	TokenNoDefaultPolicy bool     `json:"token_no_default_policy,omitempty"`
	TokenNumUses         int64    `json:"token_num_uses"`
	TokenPeriod          int64    `json:"token_period,omitempty"`
	TokenType            string   `json:"token_type,omitempty"`
}

// create or update policy binding
// it's safe to call it multiple times
func (p *pBinding) Ensure(pBind *api.VaultPolicyBinding) error {
	// kubernetes auth
	if p.authKubernetes != nil {
		path := pBind.GeneratePath(p.authKubernetes.name, p.authKubernetes.path)
		p.authKubernetes.TokenPolicies = p.policies
		payload, err := pBind.GeneratePayload(p.authKubernetes)
		if err != nil {
			return err
		}
		_, err = p.vClient.Logical().Write(path, payload)
		if err != nil {
			return err
		}
	}
	// approle auth
	if p.authAppRole != nil {
		path := pBind.GeneratePath(p.authAppRole.roleName, p.authAppRole.path)
		p.authAppRole.TokenPolicies = p.policies
		payload, err := pBind.GeneratePayload(p.authAppRole)
		if err != nil {
			return err
		}
		_, err = p.vClient.Logical().Write(path, payload)
		if err != nil {
			return err
		}
	}
	return nil
}

// delete policy binding
// it's safe to call it, even if 'name' doesn't exist in vault
func (p *pBinding) Delete(pBind *api.VaultPolicyBinding) error {
	if p.authKubernetes != nil {
		path := pBind.GeneratePath(p.authKubernetes.name, p.authKubernetes.path)
		_, err := p.vClient.Logical().Delete(path)
		if err != nil {
			return err
		}
	}
	// approle auth
	if p.authAppRole != nil {
		path := pBind.GeneratePath(p.authAppRole.roleName, p.authAppRole.path)
		_, err := p.vClient.Logical().Delete(path)
		if err != nil {
			return err
		}
	}
	return nil
}
