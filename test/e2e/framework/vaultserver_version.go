/*
Copyright AppsCode Inc. and Contributors

Licensed under the AppsCode Community License 1.0.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    https://github.com/appscode/licenses/raw/1.0.0/AppsCode-Community-1.0.0.md

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

	api "kubevault.dev/apimachinery/apis/catalog/v1alpha1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	meta_util "kmodules.xyz/client-go/meta"
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
				Image: fmt.Sprintf("%s/%s", DockerRegistry, ExporterImage),
			},
		},
	}
	_, err := f.CSClient.CatalogV1alpha1().VaultServerVersions().Create(context.TODO(), v, metav1.CreateOptions{})
	return err
}

func (f *Framework) DeleteVaultserverVersion() error {
	return f.CSClient.CatalogV1alpha1().VaultServerVersions().Delete(context.TODO(), vaultVersion, meta_util.DeleteInForeground())
}
