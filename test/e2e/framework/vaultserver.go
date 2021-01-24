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
	"fmt"

	api "kubevault.dev/operator/apis/kubevault/v1alpha1"
	patchutil "kubevault.dev/operator/client/clientset/versioned/typed/kubevault/v1alpha1/util"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"gomodules.xyz/x/crypto/rand"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	meta_util "kmodules.xyz/client-go/meta"
)

const (
	vaultVersion = "test-v1.2.3"
)

func (f *Invocation) VaultServer(replicas int32, bs api.BackendStorageSpec) *api.VaultServer {
	return &api.VaultServer{
		ObjectMeta: metav1.ObjectMeta{
			Name:      rand.WithUniqSuffix("vault-test"),
			Namespace: f.namespace,
			Labels: map[string]string{
				"test": f.app,
			},
		},
		Spec: api.VaultServerSpec{
			Replicas: &replicas,
			Version:  vaultVersion,
			Backend:  bs,
		},
	}
}

func (f *Invocation) VaultServerWithUnsealer(replicas int32, bs api.BackendStorageSpec, us api.UnsealerSpec) *api.VaultServer {
	vs := f.VaultServer(replicas, bs)
	vs.Spec.Unsealer = &us
	return vs
}

func (f *Framework) CreateVaultServer(obj *api.VaultServer) (*api.VaultServer, error) {
	return f.CSClient.KubevaultV1alpha1().VaultServers(obj.Namespace).Create(context.TODO(), obj, metav1.CreateOptions{})
}

func (f *Framework) GetVaultServer(obj *api.VaultServer) (*api.VaultServer, error) {
	return f.CSClient.KubevaultV1alpha1().VaultServers(obj.Namespace).Get(context.TODO(), obj.Name, metav1.GetOptions{})
}

func (f *Framework) UpdateVaultServer(obj *api.VaultServer) (*api.VaultServer, error) {
	in, err := f.GetVaultServer(obj)
	if err != nil {
		return nil, err
	}

	vs, _, err := patchutil.PatchVaultServer(context.TODO(), f.CSClient.KubevaultV1alpha1(), in, func(vs *api.VaultServer) *api.VaultServer {
		vs.Spec = obj.Spec
		By(fmt.Sprint(vs.Spec))
		return vs
	}, metav1.PatchOptions{})
	return vs, err
}

func (f *Framework) DeleteVaultServerObj(obj *api.VaultServer) error {
	err := f.CSClient.EngineV1alpha1().SecretEngines(obj.Namespace).Delete(context.TODO(), obj.Name, metav1.DeleteOptions{})
	if kerr.IsNotFound(err) {
		return nil
	}
	return err
}

func (f *Framework) DeleteVaultServer(meta metav1.ObjectMeta) error {
	err := f.CSClient.KubevaultV1alpha1().VaultServers(meta.Namespace).Delete(context.TODO(), meta.Name, meta_util.DeleteInBackground())
	if kerr.IsNotFound(err) {
		return nil
	}
	return err
}

func (f *Framework) EventuallyVaultServer(meta metav1.ObjectMeta) GomegaAsyncAssertion {
	return Eventually(func() *api.VaultServer {
		obj, err := f.CSClient.KubevaultV1alpha1().VaultServers(meta.Namespace).Get(context.TODO(), meta.Name, metav1.GetOptions{})
		Expect(err).NotTo(HaveOccurred())
		return obj
	})
}
