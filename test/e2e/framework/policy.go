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

package framework

import (
	"context"

	api "kubevault.dev/operator/apis/policy/v1alpha1"

	. "github.com/onsi/gomega"
	"gomodules.xyz/x/crypto/rand"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	meta_util "kmodules.xyz/client-go/meta"
	appcat "kmodules.xyz/custom-resources/apis/appcatalog/v1alpha1"
)

func (f *Invocation) VaultPolicy(policy string, ref *appcat.AppReference) *api.VaultPolicy {
	return &api.VaultPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      rand.WithUniqSuffix("vault-policy"),
			Namespace: f.namespace,
			Labels: map[string]string{
				"test": f.app,
			},
		},
		Spec: api.VaultPolicySpec{
			VaultRef: core.LocalObjectReference{
				Name: ref.Name,
			},
			PolicyDocument: policy,
		},
	}
}

func (f *Framework) CreateVaultPolicy(obj *api.VaultPolicy) (*api.VaultPolicy, error) {
	return f.CSClient.PolicyV1alpha1().VaultPolicies(obj.Namespace).Create(context.TODO(), obj, metav1.CreateOptions{})
}

func (f *Framework) GetVaultPolicy(obj *api.VaultPolicy) (*api.VaultPolicy, error) {
	return f.CSClient.PolicyV1alpha1().VaultPolicies(obj.Namespace).Get(context.TODO(), obj.Name, metav1.GetOptions{})
}

func (f *Framework) UpdateVaultPolicy(obj *api.VaultPolicy) (*api.VaultPolicy, error) {
	return f.CSClient.PolicyV1alpha1().VaultPolicies(obj.Namespace).Update(context.TODO(), obj, metav1.UpdateOptions{})
}

func (f *Framework) DeleteVaultPolicy(meta metav1.ObjectMeta) error {
	return f.CSClient.PolicyV1alpha1().VaultPolicies(meta.Namespace).Delete(context.TODO(), meta.Name, meta_util.DeleteInBackground())
}

func (f *Framework) EventuallyVaultPolicy(meta metav1.ObjectMeta) GomegaAsyncAssertion {
	return Eventually(func() *api.VaultPolicy {
		obj, err := f.CSClient.PolicyV1alpha1().VaultPolicies(meta.Namespace).Get(context.TODO(), meta.Name, metav1.GetOptions{})
		Expect(err).NotTo(HaveOccurred())
		return obj
	})
}
