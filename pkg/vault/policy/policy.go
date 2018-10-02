package policy

import (
	vaultapi "github.com/hashicorp/vault/api"
	api "github.com/kubevault/operator/apis/policy/v1alpha1"
	"github.com/kubevault/operator/pkg/vault"
	"github.com/pkg/errors"
	"k8s.io/client-go/kubernetes"
)

type Policy interface {
	EnsurePolicy(name string, policy string) error
	DeletePolicy(name string) error
}

type vPolicy struct {
	client *vaultapi.Client
}

func NewPolicyClientForVault(kc kubernetes.Interface, p *api.VaultPolicy) (Policy, error) {
	if p == nil {
		return nil, errors.New(".spec.vault is nil")
	}

	vc, err := vault.NewClient(kc, p.Namespace, p.Spec.Vault)
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
