package framework

import (
	"github.com/appscode/go/crypto/rand"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	api "kubevault.dev/operator/apis/policy/v1alpha1"
)

func (f *Invocation) VaultPolicyBinding(policies, saNames, saNamespaces []string) *api.VaultPolicyBinding {
	return &api.VaultPolicyBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      rand.WithUniqSuffix("v-policy-binding"),
			Namespace: f.namespace,
			Labels: map[string]string{
				"test": f.app,
			},
		},
		Spec: api.VaultPolicyBindingSpec{
			Policies:                 policies,
			ServiceAccountNames:      saNames,
			ServiceAccountNamespaces: saNamespaces,
		},
	}
}

func (f *Framework) CreateVaultPolicyBinding(obj *api.VaultPolicyBinding) (*api.VaultPolicyBinding, error) {
	return f.CSClient.PolicyV1alpha1().VaultPolicyBindings(obj.Namespace).Create(obj)
}

func (f *Framework) GetVaultPolicyBinding(obj *api.VaultPolicyBinding) (*api.VaultPolicyBinding, error) {
	return f.CSClient.PolicyV1alpha1().VaultPolicyBindings(obj.Namespace).Get(obj.Name, metav1.GetOptions{})
}

func (f *Framework) UpdateVaultPolicyBinding(obj *api.VaultPolicyBinding) (*api.VaultPolicyBinding, error) {
	return f.CSClient.PolicyV1alpha1().VaultPolicyBindings(obj.Namespace).Update(obj)
}

func (f *Framework) DeleteVaultPolicyBinding(meta metav1.ObjectMeta) error {
	return f.CSClient.PolicyV1alpha1().VaultPolicyBindings(meta.Namespace).Delete(meta.Name, deleteInBackground())
}

func (f *Framework) EventuallyVaultPolicyBinding(meta metav1.ObjectMeta) GomegaAsyncAssertion {
	return Eventually(func() *api.VaultPolicyBinding {
		obj, err := f.CSClient.PolicyV1alpha1().VaultPolicyBindings(meta.Namespace).Get(meta.Name, metav1.GetOptions{})
		Expect(err).NotTo(HaveOccurred())
		return obj
	})
}
