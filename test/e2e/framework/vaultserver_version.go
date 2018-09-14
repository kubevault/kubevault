package framework

import (
	api "github.com/kubevault/operator/apis/core/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (f *Framework) CreateVaultserverVersion() error {
	v := &api.VaultserverVersion{
		ObjectMeta: metav1.ObjectMeta{
			Name: vaultVersion,
		},
		Spec: api.VaultserverVersionSpec{
			Version: vaultVersion,
			Vault: api.VaultserverVersionVault{
				Image: "vault:0.11.1",
			},
			Unsealer: api.VaultserverVersionUnsealer{
				Image: "nightfury1204/vault-unsealer:canary",
			},
		},
	}
	_, err := f.VaultServerClient.CoreV1alpha1().VaultserverVersions().Create(v)
	return err
}

func (f *Framework) DeleteVaultserverVersion() error {
	return f.VaultServerClient.CoreV1alpha1().VaultserverVersions().Delete(vaultVersion, deleteInForeground())
}
