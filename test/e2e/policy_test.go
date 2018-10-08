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
	appcat "kmodules.xyz/custom-resources/apis/appcatalog/v1alpha1"
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
			By(fmt.Sprintf("checking policy(%s) exists in vault", p.OffshootName()))
			Eventually(func() bool {
				vApp, err := f.GetAppBinding(p.Spec.VaultAppRef.Name, p.Spec.VaultAppRef.Namespace)
				if err != nil {
					return false
				}
				sr, err := f.KubeClient.CoreV1().Secrets(p.Namespace).Get(vApp.Spec.Secret.Name, metav1.GetOptions{})
				if err != nil {
					return false
				}
				var addr string
				cfg := vApp.Spec.ClientConfig
				if cfg.URL != nil {
					addr = *cfg.URL
				} else {
					Expect(len(cfg.Ports) == 1).NotTo(BeTrue(), "number of port is zero or more than one")
					addr = fmt.Sprintf("%s.%s.svc:%d", cfg.Service.Name, vApp.Namespace, cfg.Ports[0].Port)
				}
				vc, err := framework.GetVaultClient(addr, string(sr.Data["token"]))
				if err != nil {
					return false
				}
				_, err = vc.Sys().GetPolicy(p.OffshootName())
				return err == nil
			}, timeOut, pollingInterval).Should(BeTrue(), fmt.Sprintf("policy(%s) should exists in vault", p.OffshootName()))
		}
		IsPolicyUpdatedInVault = func(p *api.VaultPolicy, plcy string) {
			By(fmt.Sprintf("checking policy(%s) exists in vault", p.OffshootName()))
			Eventually(func() bool {
				vApp, err := f.GetAppBinding(p.Spec.VaultAppRef.Name, p.Spec.VaultAppRef.Namespace)
				if err != nil {
					return false
				}
				sr, err := f.KubeClient.CoreV1().Secrets(p.Namespace).Get(vApp.Spec.Secret.Name, metav1.GetOptions{})
				if err != nil {
					return false
				}
				var addr string
				cfg := vApp.Spec.ClientConfig
				if cfg.URL != nil {
					addr = *cfg.URL
				} else {
					Expect(len(cfg.Ports) == 1).NotTo(BeTrue(), "number of port is zero or more than one")
					addr = fmt.Sprintf("%s.%s.svc:%d", cfg.Service.Name, vApp.Namespace, cfg.Ports[0].Port)
				}
				vc, err := framework.GetVaultClient(addr, string(sr.Data["token"]))
				if err != nil {
					return false
				}
				p, err := vc.Sys().GetPolicy(p.OffshootName())
				return err == nil && p == plcy
			}, timeOut, pollingInterval).Should(BeTrue(), fmt.Sprintf("policy(%s) should exists in vault", p.OffshootName()))
		}
	)

	BeforeEach(func() {
		f = root.Invoke()
	})
	AfterEach(func() {
		time.Sleep(5 * time.Second)
	})

	FDescribe("Create and Update VaultPolicy", func() {
		Context("Create", func() {
			var (
				vPolicy *api.VaultPolicy
			)

			BeforeEach(func() {
				plcy := "{}"
				vPolicy = f.VaultPolicy(plcy, f.VaultAppRef)
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
				vPolicy = f.VaultPolicy(plcy, f.VaultAppRef)
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

		Context("Delete, containing invalid vault address", func() {
			var (
				vPolicy *api.VaultPolicy
			)

			BeforeEach(func() {
				plcy := `path "secret/*" {
			   		capabilities = ["create", "read", "update", "delete", "list"]
	 			}`
				vPolicy = f.VaultPolicy(plcy, &appcat.AppReference{
					Name:      "invalid",
					Namespace: f.Namespace(),
				})
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
