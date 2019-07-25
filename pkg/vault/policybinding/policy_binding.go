package policybinding

import (
	"fmt"

	vaultapi "github.com/hashicorp/vault/api"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	appcat "kmodules.xyz/custom-resources/apis/appcatalog/v1alpha1"
	appcat_cs "kmodules.xyz/custom-resources/client/clientset/versioned/typed/appcatalog/v1alpha1"
	api "kubevault.dev/operator/apis/policy/v1alpha1"
	cs "kubevault.dev/operator/client/clientset/versioned"
	"kubevault.dev/operator/pkg/vault"
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

	pb := &pBinding{
		saNames:      pBind.Spec.ServiceAccountNames,
		saNamespaces: pBind.Spec.ServiceAccountNamespaces,
		ttl:          pBind.Spec.TTL,
		maxTTL:       pBind.Spec.MaxTTL,
		period:       pBind.Spec.Period,
		path:         pBind.Spec.AuthPath,
	}
	pb.setDefaults()

	var vaultRef *appcat.AppReference
	// check whether VaultPolicy exists
	for _, pName := range pBind.Spec.Policies {
		plcy, err := c.PolicyV1alpha1().VaultPolicies(pBind.Namespace).Get(pName, metav1.GetOptions{})
		if err != nil {
			return nil, errors.Wrap(err, "for .spec.policies")
		}
		if vaultRef == nil {
			// take vault connection reference from policy
			vaultRef = plcy.Spec.Ref
		} else {
			// all policy should refer the same vault
			vr := plcy.Spec.Ref
			if vr == nil || vr.Name != vaultRef.Name || vr.Namespace != vaultRef.Namespace {
				return nil, errors.New("all policy should refer the same vault")
			}
		}
		// add vault policy name
		// VaultPolicy.PolicyName() is used to create policy in vault
		pb.policies = append(pb.policies, plcy.PolicyName())
	}

	var err error
	pb.vClient, err = vault.NewClient(kc, appc, vaultRef)
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

func (p *pBinding) setDefaults() {
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
