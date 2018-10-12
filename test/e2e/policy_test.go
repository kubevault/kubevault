package e2e_test

import (
	"fmt"
	"time"

	api "github.com/kubevault/operator/apis/policy/v1alpha1"
	"github.com/kubevault/operator/pkg/vault"
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
			By(fmt.Sprintf("checking policy(%s) exists in vault", p.PolicyName()))
			Eventually(func() bool {
				vc, err := vault.NewClient(f.KubeClient, f.AppcatClient, f.VaultAppRef)
				if err != nil {
					return false
				}
				_, err = vc.Sys().GetPolicy(p.PolicyName())
				return err == nil
			}, timeOut, pollingInterval).Should(BeTrue(), fmt.Sprintf("policy(%s) should exists in vault", p.PolicyName()))
		}
		IsPolicyUpdatedInVault = func(p *api.VaultPolicy, plcy string) {
			By(fmt.Sprintf("checking policy(%s) exists in vault", p.PolicyName()))
			Eventually(func() bool {
				vc, err := vault.NewClient(f.KubeClient, f.AppcatClient, f.VaultAppRef)
				if err != nil {
					return false
				}
				p, err := vc.Sys().GetPolicy(p.PolicyName())
				return err == nil && p == plcy
			}, timeOut, pollingInterval).Should(BeTrue(), fmt.Sprintf("policy(%s) should exists in vault", p.PolicyName()))
		}
		IsVaultPolicyBindingDeleted = func(name, namespace string) {
			By(fmt.Sprintf("Waiting for VaultPolicyBinding (%s/%s) to delete", namespace, name))
			Eventually(func() bool {
				_, err := f.CSClient.PolicyV1alpha1().VaultPolicyBindings(namespace).Get(name, metav1.GetOptions{})
				return kerr.IsNotFound(err)
			}, timeOut, pollingInterval).Should(BeTrue(), fmt.Sprintf("VaultPolicyBinding (%s/%s) should not exists", namespace, name))
		}
		IsVaultPolicyBindingSucceeded = func(name, namespace string) {
			By(fmt.Sprintf("Waiting for VaultPolicyBinding (%s/%s) to success", namespace, name))
			Eventually(func() bool {
				pb, err := f.CSClient.PolicyV1alpha1().VaultPolicyBindings(namespace).Get(name, metav1.GetOptions{})
				if err != nil {
					return false
				}
				return pb.Status.Status == api.PolicyBindingSuccess
			}, timeOut, pollingInterval).Should(BeTrue(), fmt.Sprintf("VaultPolicyBinding (%s/%s) should succeed", namespace, name))
		}

		IsVaultPolicyBindingFailed = func(name, namespace string) {
			By(fmt.Sprintf("Waiting for VaultPolicyBinding (%s/%s) to fail", namespace, name))
			Eventually(func() bool {
				pb, err := f.CSClient.PolicyV1alpha1().VaultPolicyBindings(namespace).Get(name, metav1.GetOptions{})
				if err != nil {
					return false
				}
				return pb.Status.Status != api.PolicyBindingSuccess
			}, timeOut, pollingInterval).Should(BeTrue(), fmt.Sprintf("VaultPolicyBinding (%s/%s) should fail", namespace, name))
		}
	)

	BeforeEach(func() {
		f = root.Invoke()
		// enable kubernetes auth
		vc, err := vault.NewClient(f.KubeClient, f.AppcatClient, f.VaultAppRef)
		Expect(err).NotTo(HaveOccurred(), "create vault client")
		Expect(framework.EnsureKubernetesAuth(vc)).NotTo(HaveOccurred(), "ensure kubernetes auth")

	})
	AfterEach(func() {
		time.Sleep(5 * time.Second)
	})

	Describe("Create, Update and Delete VaultPolicy", func() {
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

	Describe("Create, Update and Delete VaultPolicyBinding", func() {
		var (
			vPolicy *api.VaultPolicy
		)

		BeforeEach(func() {
			vPolicy = f.VaultPolicy("{}", f.VaultAppRef)
			_, err := f.CreateVaultPolicy(vPolicy)
			Expect(err).NotTo(HaveOccurred())
			IsPolicyExistInVault(vPolicy)
		})
		AfterEach(func() {
			Expect(f.DeleteVaultPolicy(vPolicy.ObjectMeta)).NotTo(HaveOccurred())
			IsVaultPolicyDeleted(vPolicy.Name, vPolicy.Namespace)
		})

		Context("Create", func() {
			var (
				vPBind *api.VaultPolicyBinding
			)

			BeforeEach(func() {
				vPBind = f.VaultPolicyBinding([]string{vPolicy.Name}, []string{"test"}, []string{"test"})
			})
			AfterEach(func() {
				Expect(f.DeleteVaultPolicyBinding(vPBind.ObjectMeta)).NotTo(HaveOccurred())
				IsVaultPolicyBindingDeleted(vPBind.Name, vPBind.Namespace)
			})

			It("should be successful", func() {
				_, err := f.CreateVaultPolicyBinding(vPBind)
				Expect(err).NotTo(HaveOccurred())
				IsVaultPolicyBindingSucceeded(vPBind.Name, vPBind.Namespace)
			})
		})

		Context("Update service account names", func() {
			var (
				vPBind *api.VaultPolicyBinding
			)

			BeforeEach(func() {
				vPBind = f.VaultPolicyBinding([]string{vPolicy.Name}, []string{"test"}, []string{"test"})
				_, err := f.CreateVaultPolicyBinding(vPBind)
				Expect(err).NotTo(HaveOccurred())
				IsVaultPolicyBindingSucceeded(vPBind.Name, vPBind.Namespace)
			})
			AfterEach(func() {
				Expect(f.DeleteVaultPolicyBinding(vPBind.ObjectMeta)).NotTo(HaveOccurred())
				IsVaultPolicyBindingDeleted(vPBind.Name, vPBind.Namespace)
			})

			It("should be successful", func() {
				var err error
				vPBind, err = f.GetVaultPolicyBinding(vPBind)
				Expect(err).NotTo(HaveOccurred())

				vPBind.Spec.ServiceAccountNames = []string{"new"}
				vPBind, err = f.UpdateVaultPolicyBinding(vPBind)
				Expect(err).NotTo(HaveOccurred(), "update VaultPolicyBinding")
				// wait to apply the changes
				time.Sleep(time.Second * 3)
				IsVaultPolicyBindingSucceeded(vPBind.Name, vPBind.Namespace)
			})
		})

		Context("Delete invalid VaultPolicyBinding", func() {
			var (
				vPBind *api.VaultPolicyBinding
			)

			BeforeEach(func() {
				vPBind = f.VaultPolicyBinding([]string{"invalid"}, []string{"test"}, []string{"test"})
				_, err := f.CreateVaultPolicyBinding(vPBind)
				Expect(err).NotTo(HaveOccurred())
				IsVaultPolicyBindingFailed(vPBind.Name, vPBind.Namespace)
			})

			It("should be successful", func() {
				Expect(f.DeleteVaultPolicyBinding(vPBind.ObjectMeta)).NotTo(HaveOccurred(), "delete VaultPolicyBinding")
				IsVaultPolicyBindingDeleted(vPBind.Name, vPBind.Namespace)
			})
		})
	})
})
