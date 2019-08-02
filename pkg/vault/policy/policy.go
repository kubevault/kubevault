package policy

import (
	vaultapi "github.com/hashicorp/vault/api"
	"github.com/pkg/errors"
	"k8s.io/client-go/kubernetes"
	appcat "kmodules.xyz/custom-resources/apis/appcatalog/v1alpha1"
	appcat_cs "kmodules.xyz/custom-resources/client/clientset/versioned/typed/appcatalog/v1alpha1"
	api "kubevault.dev/operator/apis/policy/v1alpha1"
	"kubevault.dev/operator/pkg/vault"
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
