package framework

import (
	api "github.com/kubevault/operator/apis/kubevault/v1alpha1"
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
				Image: "vault:0.11.1",
			},
			Unsealer: api.VaultServerVersionUnsealer{
				Image: "nightfury1204/vault-unsealer:canary",
			},
		},
	}
	_, err := f.VaultServerClient.KubevaultV1alpha1().VaultServerVersions().Create(v)
	return err
}

func (f *Framework) DeleteVaultserverVersion() error {
	return f.VaultServerClient.KubevaultV1alpha1().VaultServerVersions().Delete(vaultVersion, deleteInForeground())
}
