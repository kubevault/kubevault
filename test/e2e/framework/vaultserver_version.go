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
	"fmt"

	api "kubevault.dev/operator/apis/catalog/v1alpha1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (f *Framework) CreateVaultserverVersion() error {
	v := &api.VaultServerVersion{
		ObjectMeta: metav1.ObjectMeta{
			Name: vaultVersion,
		},
		Spec: api.VaultServerVersionSpec{
			Version: vaultVersion,
			Vault: api.VaultServerVersionVault{
				Image: "vault:1.2.3",
			},
			Unsealer: api.VaultServerVersionUnsealer{
				Image: fmt.Sprintf("%s/%s", DockerRegistry, UnsealerImage),
			},
			Exporter: api.VaultServerVersionExporter{
				Image: "kubevault/vault-exporter:0.1.0",
			},
		},
	}
	_, err := f.CSClient.CatalogV1alpha1().VaultServerVersions().Create(v)
	return err
}

func (f *Framework) DeleteVaultserverVersion() error {
	return f.CSClient.CatalogV1alpha1().VaultServerVersions().Delete(vaultVersion, deleteInForeground())
}
