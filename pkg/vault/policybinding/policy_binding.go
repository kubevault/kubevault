package policybinding

import (
	"fmt"

	vautlapi "github.com/hashicorp/vault/api"
	api "github.com/kubevault/operator/apis/policy/v1alpha1"
	cs "github.com/kubevault/operator/client/clientset/versioned"
	"github.com/kubevault/operator/pkg/vault"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type PolicyBinding interface {
	// create or update policy binding
	Ensure(name string) error
	// delete policy binding
	Delete(name string) error
}

func NewPolicyBindingClient(c cs.Interface, kc kubernetes.Interface, pBind *api.VaultPolicyBinding) (PolicyBinding, error) {
	if pBind == nil {
		return nil, errors.New("VaultPolicyBinding is nil")
	}
	if len(pBind.Spec.Policies) == 0 {
		return nil, errors.New(".spec.policies must be non empty")
	}

	pb := &pBinding{
		saNames:      pBind.Spec.ServiceAccountNames,
		saNamespaces: pBind.Spec.ServiceAccountNamespaces,
		ttl:          pBind.Spec.TTL,
		maxTTL:       pBind.Spec.MaxTTL,
		period:       pBind.Spec.Period,
	}

	var vaultCfg *api.Vault
	// check whether VaultPolicy exists
	for _, pName := range pBind.Spec.Policies {
		plcy, err := c.PolicyV1alpha1().VaultPolicies(pBind.Namespace).Get(pName, metav1.GetOptions{})
		if err != nil {
			return nil, errors.Wrap(err, "for .spec.policies")
		}
		if vaultCfg == nil {
			// take vault connection config from policy
			vaultCfg = plcy.Spec.Vault
		}
		// add vault policy name
		// VaultPolicy.OffshootName() is used to create policy in vault
		pb.policies = append(pb.policies, plcy.OffshootName())
	}

	var err error
	pb.vClient, err = vault.NewClient(kc, pBind.Namespace, vaultCfg)
	if err != nil {
		return nil, err
	}
	return pb, nil
}

type pBinding struct {
	vClient      *vautlapi.Client
	policies     []string
	saNames      []string
	saNamespaces []string
	ttl          string
	maxTTL       string
	period       string
}

// create or update policy binding
// it's safe to call it multiple times
func (p *pBinding) Ensure(name string) error {
	path := fmt.Sprintf("/v1/auth/kubernetes/role/%s", name)
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
	path := fmt.Sprintf("/v1/auth/kubernetes/role/%s", name)
	req := p.vClient.NewRequest("DELETE", path)
	_, err := p.vClient.RawRequest(req)
	if err != nil {
		return err
	}
	return nil
}
