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

func NewPolicyBindingClient(c cs.Interface, kc kubernetes.Interface, p *api.VaultPolicyBinding) (PolicyBinding, error) {
	if len(p.Spec.Policies) == 0 {
		return nil, errors.New(".spec.policies must be non empty")
	}

	vPlcy, err := c.PolicyV1alpha1().VaultPolicies(p.Namespace).Get(p.Spec.Policies[0], metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	vc, err := vault.NewClient(kc, p.Namespace, vPlcy.Spec.Vault)
	if err != nil {
		return nil, err
	}
	pb := &PBind{
		vClient:      vc,
		policies:     p.Spec.Policies,
		saNames:      p.Spec.ServiceAccountNames,
		saNamespaces: p.Spec.ServiceAccountNamespaces,
		ttl:          p.Spec.TTL,
		maxTTL:       p.Spec.MaxTTL,
		period:       p.Spec.Period,
	}
	return pb, nil
}

type PBind struct {
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
func (p *PBind) Ensure(name string) error {
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
func (p *PBind) Delete(name string) error {
	path := fmt.Sprintf("/v1/auth/kubernetes/role/%s", name)
	req := p.vClient.NewRequest("DELETE", path)
	_, err := p.vClient.RawRequest(req)
	if err != nil {
		return err
	}
	return nil
}
