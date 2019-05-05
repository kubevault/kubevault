package e2e_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"time"

	api "github.com/kubevault/operator/apis/engine/v1alpha1"
	"github.com/kubevault/operator/pkg/controller"
	"github.com/kubevault/operator/pkg/vault"
	"github.com/kubevault/operator/test/e2e/framework"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	appcat "kmodules.xyz/custom-resources/apis/appcatalog/v1alpha1"
)

var _ = Describe("GCP Role", func() {

	var f *framework.Invocation

	BeforeEach(func() {
		f = root.Invoke()
	})

	AfterEach(func() {
		time.Sleep(20 * time.Second)
	})

	var (
		IsVaultGCPRoleCreated = func(name string) {
			By("Checking whether vault gcp role is created")
			cl, err := vault.NewClient(f.KubeClient, f.AppcatClient, f.VaultAppRef)
			Expect(err).NotTo(HaveOccurred(), "To get vault client")

			req := cl.NewRequest("GET", fmt.Sprintf("/v1/gcp/roleset/%s", name))
			Eventually(func() bool {
				_, err := cl.RawRequest(req)
				return err == nil
			}, timeOut, pollingInterval).Should(BeTrue(), "Vault gcp role is created")

		}

		IsVaultGCPRoleDeleted = func(name string) {
			By("Checking whether vault gcp role is deleted")
			cl, err := vault.NewClient(f.KubeClient, f.AppcatClient, f.VaultAppRef)
			Expect(err).NotTo(HaveOccurred(), "To get vault client")

			req := cl.NewRequest("GET", fmt.Sprintf("/v1/gcp/roleset/%s", name))
			Eventually(func() bool {
				_, err := cl.RawRequest(req)
				return err != nil
			}, timeOut, pollingInterval).Should(BeTrue(), "Vault gcp role is deleted")

		}

		IsGCPRoleCreated = func(name, namespace string) {
			By(fmt.Sprintf("Checking whether GCPRole:(%s/%s) role is created", namespace, name))
			Eventually(func() bool {
				_, err := f.CSClient.EngineV1alpha1().GCPRoles(namespace).Get(name, metav1.GetOptions{})
				return err == nil
			}, timeOut, pollingInterval).Should(BeTrue(), "GCPRole is created")
		}

		IsGCPRoleDeleted = func(name, namespace string) {
			By(fmt.Sprintf("Checking whether GCPRole:(%s/%s) is created", namespace, name))
			Eventually(func() bool {
				_, err := f.CSClient.EngineV1alpha1().GCPRoles(namespace).Get(name, metav1.GetOptions{})
				return kerrors.IsNotFound(err)
			}, timeOut, pollingInterval).Should(BeTrue(), "GCPRole is deleted")
		}

		IsGCPRoleSucceeded = func(name, namespace string) {
			By(fmt.Sprintf("Checking whether GCPRole:(%s/%s) is succeeded", namespace, name))
			Eventually(func() bool {
				r, err := f.CSClient.EngineV1alpha1().GCPRoles(namespace).Get(name, metav1.GetOptions{})
				if err == nil {
					return r.Status.Phase == controller.GCPRolePhaseSuccess
				}
				return false
			}, timeOut, pollingInterval).Should(BeTrue(), "GCPRole status is succeeded")

		}
	)

	Describe("GCPRole", func() {
		var (
			gcpCredentials corev1.Secret
			gcpRole        api.GCPRole
		)

		const (
			gcpCredSecret = "gcp-cred"
			gcpRoleName   = "my-gcp-roleset"
		)

		BeforeEach(func() {

			credentialAddr := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
			jsonBytes, err := ioutil.ReadFile(credentialAddr)
			Expect(err).NotTo(HaveOccurred(), "Parse gcp credentials")

			gcpCredentials = corev1.Secret{
				TypeMeta: metav1.TypeMeta{},
				ObjectMeta: metav1.ObjectMeta{
					Name:      gcpCredSecret,
					Namespace: f.Namespace(),
				},
				Data: map[string][]byte{
					"sa.json": jsonBytes,
				},
			}
			_, err = f.KubeClient.CoreV1().Secrets(f.Namespace()).Create(&gcpCredentials)
			Expect(err).NotTo(HaveOccurred(), "Create gcp credentials secret")

			gcpRole = api.GCPRole{
				ObjectMeta: metav1.ObjectMeta{
					Name:      gcpRoleName,
					Namespace: f.Namespace(),
				},
				Spec: api.GCPRoleSpec{
					AuthManagerRef: f.VaultAppRef,
					Config: &api.GCPConfig{
						CredentialSecret: gcpCredSecret,
					},
					SecretType: "access_token",
					Project:    "ackube",
					Bindings: ` resource "//cloudresourcemanager.googleapis.com/projects/ackube" {
					roles = ["roles/viewer"]
				}`,
					TokenScopes: []string{"https://www.googleapis.com/auth/cloud-platform"},
				},
			}
		})

		AfterEach(func() {
			err := f.KubeClient.CoreV1().Secrets(f.Namespace()).Delete(gcpCredSecret, &metav1.DeleteOptions{})
			Expect(err).NotTo(HaveOccurred(), "Delete gcp credentials secret")
		})

		Context("Create GCPRole", func() {
			var p api.GCPRole

			BeforeEach(func() {
				p = gcpRole
			})

			AfterEach(func() {
				err := f.CSClient.EngineV1alpha1().GCPRoles(gcpRole.Namespace).Delete(p.Name, &metav1.DeleteOptions{})
				Expect(err).NotTo(HaveOccurred(), "Delete GCPRole")

				IsVaultGCPRoleDeleted(p.RoleName())
				IsGCPRoleDeleted(p.Name, p.Namespace)
			})

			It("Should be successful", func() {
				_, err := f.CSClient.EngineV1alpha1().GCPRoles(p.Namespace).Create(&p)
				Expect(err).NotTo(HaveOccurred(), "Create GCPRole")

				IsGCPRoleCreated(p.Name, p.Namespace)
				IsVaultGCPRoleCreated(p.RoleName())
				IsGCPRoleSucceeded(p.Name, p.Namespace)
			})

		})

		Context("Create GCPRole with invalid vault AppReference", func() {
			var p api.GCPRole

			BeforeEach(func() {
				p = gcpRole
				p.Spec.AuthManagerRef = &appcat.AppReference{
					Namespace: gcpRole.Namespace,
					Name:      "invalid",
				}
			})

			AfterEach(func() {
				err := f.CSClient.EngineV1alpha1().GCPRoles(gcpRole.Namespace).Delete(p.Name, &metav1.DeleteOptions{})
				Expect(err).NotTo(HaveOccurred(), "Delete GCPRole")

				IsVaultGCPRoleDeleted(p.RoleName())
				IsGCPRoleDeleted(p.Name, p.Namespace)
			})

			It("Should be successful", func() {
				_, err := f.CSClient.EngineV1alpha1().GCPRoles(p.Namespace).Create(&p)
				Expect(err).NotTo(HaveOccurred(), "Create GCPRole")

				IsGCPRoleCreated(p.Name, p.Namespace)
				IsVaultGCPRoleDeleted(p.RoleName())
			})
		})

	})
})
