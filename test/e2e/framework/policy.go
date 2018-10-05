package framework

import (
	"github.com/appscode/go/crypto/rand"
	api "github.com/kubevault/operator/apis/policy/v1alpha1"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	appcat "kmodules.xyz/custom-resources/apis/appcatalog/v1alpha1"
)

func (f *Invocation) VaultPolicy(policy, vAddr, tokenSecret string) *api.VaultPolicy {
	return &api.VaultPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      rand.WithUniqSuffix("vault-policy"),
			Namespace: f.namespace,
			Labels: map[string]string{
				"test": f.app,
			},
		},
		Spec: api.VaultPolicySpec{
			VaultAppRef: &appcat.AppReference{},
			Policy:      policy,
		},
	}
}

func (f *Framework) CreateVaultPolicy(obj *api.VaultPolicy) (*api.VaultPolicy, error) {
	return f.CSClient.PolicyV1alpha1().VaultPolicies(obj.Namespace).Create(obj)
}

func (f *Framework) GetVaultPolicy(obj *api.VaultPolicy) (*api.VaultPolicy, error) {
	return f.CSClient.PolicyV1alpha1().VaultPolicies(obj.Namespace).Get(obj.Name, metav1.GetOptions{})
}

func (f *Framework) UpdateVaultPolicy(obj *api.VaultPolicy) (*api.VaultPolicy, error) {
	return f.CSClient.PolicyV1alpha1().VaultPolicies(obj.Namespace).Update(obj)
}

func (f *Framework) DeleteVaultPolicy(meta metav1.ObjectMeta) error {
	return f.CSClient.PolicyV1alpha1().VaultPolicies(meta.Namespace).Delete(meta.Name, deleteInBackground())
}

func (f *Framework) EventuallyVaultPolicy(meta metav1.ObjectMeta) GomegaAsyncAssertion {
	return Eventually(func() *api.VaultPolicy {
		obj, err := f.CSClient.PolicyV1alpha1().VaultPolicies(meta.Namespace).Get(meta.Name, metav1.GetOptions{})
		Expect(err).NotTo(HaveOccurred())
		return obj
	})
}
