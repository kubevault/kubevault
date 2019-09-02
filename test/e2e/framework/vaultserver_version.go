package framework

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	api "kubevault.dev/operator/apis/catalog/v1alpha1"
)

func (f *Framework) CreateVaultserverVersion() error {
	v := &api.VaultServerVersion{
		ObjectMeta: metav1.ObjectMeta{
			Name: vaultVersion,
		},
		Spec: api.VaultServerVersionSpec{
			Version: vaultVersion,
			Vault: api.VaultServerVersionVault{
				Image: "vault:1.2.0",
			},
			Unsealer: api.VaultServerVersionUnsealer{
				Image: "kamolhasan/vault-unsealer:new-unsealer_linux_amd64",
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
