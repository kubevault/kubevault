package e2e_test

import (
	"fmt"
	"time"

	api "github.com/kubevault/operator/apis/policy/v1alpha1"
	"github.com/kubevault/operator/test/e2e/framework"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("VaultPolicy", func() {
	var (
		f *framework.Invocation
	)

	var (
		IsVaultPolicyDeleted = func(name, namespace string) {
			By(fmt.Sprintf("Waiting for VaultPolicy (%s/%s) to delete", namespace, name))
			Eventually(func() bool {
				_, err := f.CSClient.PolicyV1alpha1().VaultPolicies(namespace).Get(name, metav1.GetOptions{})
				return kerr.IsNotFound(err)
			}, timeOut, pollingInterval).Should(BeTrue(), fmt.Sprintf("VaultPolicy (%s/%s) should not exists", namespace, name))
		}
		IsPolicyExistInVault = func(p *api.VaultPolicy) {
			By(fmt.Sprintf("checking policy(%s) exists in vault", p.Name))
			Eventually(func() bool {
				sr, err := f.KubeClient.CoreV1().Secrets(p.Namespace).Get(p.Spec.Vault.TokenSecret, metav1.GetOptions{})
				if err != nil {
					return false
				}
				vc, err := framework.GetVaultClient(p.Spec.Vault.Address, string(sr.Data["token"]))
				if err != nil {
					return false
				}
				_, err = vc.Sys().GetPolicy(p.Name)
				return err == nil
			}, timeOut, pollingInterval).Should(BeTrue(), fmt.Sprintf("policy(%s) should exists in vault", p.Name))
		}
		IsPolicyUpdatedInVault = func(p *api.VaultPolicy, plcy string) {
			By(fmt.Sprintf("checking policy(%s) exists in vault", p.Name))
			Eventually(func() bool {
				sr, err := f.KubeClient.CoreV1().Secrets(p.Namespace).Get(p.Spec.Vault.TokenSecret, metav1.GetOptions{})
				if err != nil {
					return false
				}
				vc, err := framework.GetVaultClient(p.Spec.Vault.Address, string(sr.Data["token"]))
				if err != nil {
					return false
				}
				p, err := vc.Sys().GetPolicy(p.Name)
				By(p)
				return err == nil && p == plcy
			}, timeOut, pollingInterval).Should(BeTrue(), fmt.Sprintf("policy(%s) should exists in vault", p.Name))
		}
	)

	BeforeEach(func() {
		f = root.Invoke()
	})
	AfterEach(func() {
		time.Sleep(5 * time.Second)
	})

	Describe("Create and Update VaultPolicy", func() {
		Context("Create", func() {
			var (
				vPolicy *api.VaultPolicy
			)

			BeforeEach(func() {
				plcy := "{}"
				vPolicy = f.VaultPolicy(plcy, f.VaultUrl, framework.VaultTokenSecret)
			})
			AfterEach(func() {
				Expect(f.DeleteVaultPolicy(vPolicy.ObjectMeta)).NotTo(HaveOccurred())
				IsVaultPolicyDeleted(vPolicy.Name, vPolicy.Namespace)
			})

			It("should be successful", func() {
				_, err := f.CreateVaultPolicy(vPolicy)
				Expect(err).NotTo(HaveOccurred())
				IsPolicyExistInVault(vPolicy)
			})
		})

		Context("Update", func() {
			var (
				vPolicy *api.VaultPolicy
				err     error
			)

			BeforeEach(func() {
				plcy := `path "secret/*" {
			   		capabilities = ["create", "read", "update", "delete", "list"]
	 			}`
				vPolicy = f.VaultPolicy(plcy, f.VaultUrl, framework.VaultTokenSecret)
				_, err := f.CreateVaultPolicy(vPolicy)
				Expect(err).NotTo(HaveOccurred())
				IsPolicyExistInVault(vPolicy)
			})
			AfterEach(func() {
				Expect(f.DeleteVaultPolicy(vPolicy.ObjectMeta)).NotTo(HaveOccurred())
				IsVaultPolicyDeleted(vPolicy.Name, vPolicy.Namespace)
			})

			It("should be successful", func() {
				vPolicy, err = f.GetVaultPolicy(vPolicy)
				Expect(err).NotTo(HaveOccurred())

				plcy := "{}"
				vPolicy.Spec.Policy = plcy
				_, err := f.UpdateVaultPolicy(vPolicy)
				Expect(err).NotTo(HaveOccurred())
				IsPolicyUpdatedInVault(vPolicy, plcy)
			})
		})

		Context("Delete contianing invalid vauld address", func() {
			var (
				vPolicy *api.VaultPolicy
			)

			BeforeEach(func() {
				plcy := `path "secret/*" {
			   		capabilities = ["create", "read", "update", "delete", "list"]
	 			}`
				vPolicy = f.VaultPolicy(plcy, "https://invalid.com:8200", framework.VaultTokenSecret)
				_, err := f.CreateVaultPolicy(vPolicy)
				Expect(err).NotTo(HaveOccurred())
				time.Sleep(1 * time.Second)
			})

			It("should be successful", func() {
				Expect(f.DeleteVaultPolicy(vPolicy.ObjectMeta)).NotTo(HaveOccurred())
				IsVaultPolicyDeleted(vPolicy.Name, vPolicy.Namespace)
			})
		})
	})
})
