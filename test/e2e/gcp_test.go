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
	core "k8s.io/api/core/v1"
	rbac "k8s.io/api/rbac/v1"
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
		IsGCPAccessKeyRequestCreated = func(name, namespace string) {
			By(fmt.Sprintf("Checking whether GCPAccessKeyRequest:(%s/%s) is created", namespace, name))
			Eventually(func() bool {
				_, err := f.CSClient.EngineV1alpha1().GCPAccessKeyRequests(namespace).Get(name, metav1.GetOptions{})
				if err == nil {
					return true
				}
				return false
			}, timeOut, pollingInterval).Should(BeTrue(), "GCPAccessKeyRequest is created")
		}
		IsGCPAccessKeyRequestDeleted = func(name, namespace string) {
			By(fmt.Sprintf("Checking whether GCPAccessKeyRequest:(%s/%s) is deleted", namespace, name))
			Eventually(func() bool {
				_, err := f.CSClient.EngineV1alpha1().GCPAccessKeyRequests(namespace).Get(name, metav1.GetOptions{})
				return kerrors.IsNotFound(err)
			}, timeOut, pollingInterval).Should(BeTrue(), "GCPAccessKeyRequest is deleted")
		}
		IsGCPAKRConditionApproved = func(name, namespace string) {
			By(fmt.Sprintf("Checking whether GCPAccessKeyRequestConditions-> Type: Approved"))
			Eventually(func() bool {
				crd, err := f.CSClient.EngineV1alpha1().GCPAccessKeyRequests(namespace).Get(name, metav1.GetOptions{})
				if err == nil {
					for _, value := range crd.Status.Conditions {
						if value.Type == api.AccessApproved {
							return true
						}
					}
				}
				return false
			}, timeOut, pollingInterval).Should(BeTrue(), "Conditions-> Type : Approved")
		}
		IsGCPAKRConditionDenied = func(name, namespace string) {
			By(fmt.Sprintf("Checking whether GCPAccessKeyRequestConditions-> Type: Denied"))
			Eventually(func() bool {
				crd, err := f.CSClient.EngineV1alpha1().GCPAccessKeyRequests(namespace).Get(name, metav1.GetOptions{})
				if err == nil {
					for _, value := range crd.Status.Conditions {
						if value.Type == api.AccessDenied {
							return true
						}
					}
				}
				return false
			}, timeOut, pollingInterval).Should(BeTrue(), "Conditions-> Type: Denied")
		}
		IsGCPAccessKeySecretCreated = func(name, namespace string) {
			By("Checking whether GCPAccessKeySecret is created")
			Eventually(func() bool {
				crd, err := f.CSClient.EngineV1alpha1().GCPAccessKeyRequests(namespace).Get(name, metav1.GetOptions{})
				if err == nil && crd.Status.Secret != nil {
					_, err2 := f.KubeClient.CoreV1().Secrets(namespace).Get(crd.Status.Secret.Name, metav1.GetOptions{})
					return err2 == nil
				}
				return false
			}, timeOut, pollingInterval).Should(BeTrue(), "GCPAccessKeySecret is created")
		}
		IsGCPAccessKeySecretDeleted = func(secretName, namespace string) {
			By("Checking whether GCPAccessKeySecret is deleted")
			Eventually(func() bool {
				_, err2 := f.KubeClient.CoreV1().Secrets(namespace).Get(secretName, metav1.GetOptions{})
				return kerrors.IsNotFound(err2)
			}, timeOut, pollingInterval).Should(BeTrue(), "GCPAccessKeySecret is deleted")
		}
	)

	Describe("GCPRole", func() {
		var (
			gcpCredentials core.Secret
			gcpRole        api.GCPRole
		)

		const (
			gcpCredSecret = "gcp-cred-3224"
			gcpRoleName   = "my-gcp-roleset-4325"
		)

		BeforeEach(func() {

			credentialAddr := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
			jsonBytes, err := ioutil.ReadFile(credentialAddr)
			Expect(err).NotTo(HaveOccurred(), "Parse gcp credentials")

			gcpCredentials = core.Secret{
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

	Describe("GCPAccessKeyRequest", func() {
		var (
			gcpCredentials core.Secret
			gcpRole        api.GCPRole
			gcpAKReq       api.GCPAccessKeyRequest
		)
		const (
			gcpCredSecret = "gcp-cred-2343"
			gcpRoleName   = "gcp-token-roleset-23432"
			gcpAKReqName  = "gcp-akr-324432"
		)

		BeforeEach(func() {
			credentialAddr := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
			jsonBytes, err := ioutil.ReadFile(credentialAddr)
			Expect(err).NotTo(HaveOccurred(), "Parse gcp credentials")

			gcpCredentials = core.Secret{
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

			gcpAKReq = api.GCPAccessKeyRequest{
				ObjectMeta: metav1.ObjectMeta{
					Name:      gcpAKReqName,
					Namespace: f.Namespace(),
				},
				Spec: api.GCPAccessKeyRequestSpec{
					RoleRef: api.RoleReference{
						Name:      gcpRoleName,
						Namespace: f.Namespace(),
					},
					Subjects: []rbac.Subject{
						{
							Kind:      rbac.ServiceAccountKind,
							Name:      "sa-5576",
							Namespace: f.Namespace(),
						},
					},
				},
			}
		})

		AfterEach(func() {
			err := f.KubeClient.CoreV1().Secrets(f.Namespace()).Delete(gcpCredSecret, &metav1.DeleteOptions{})
			Expect(err).NotTo(HaveOccurred(), "Delete gcp credentials secret")
		})

		Context("Create, Approve, Deny GCPAccessKeyRequests", func() {
			BeforeEach(func() {
				r, err := f.CSClient.EngineV1alpha1().GCPRoles(gcpRole.Namespace).Create(&gcpRole)
				Expect(err).NotTo(HaveOccurred(), "Create GCPRole")

				IsVaultGCPRoleCreated(r.RoleName())
				IsGCPRoleSucceeded(r.Name, r.Namespace)
			})

			AfterEach(func() {
				err := f.CSClient.EngineV1alpha1().GCPAccessKeyRequests(gcpAKReq.Namespace).Delete(gcpAKReq.Name, &metav1.DeleteOptions{})
				Expect(err).NotTo(HaveOccurred(), "Delete GCPAccessKeyRequest")

				IsGCPAccessKeyRequestDeleted(gcpAKReq.Name, gcpAKReq.Namespace)

				err = f.CSClient.EngineV1alpha1().GCPRoles(gcpRole.Namespace).Delete(gcpRole.Name, &metav1.DeleteOptions{})
				Expect(err).NotTo(HaveOccurred(), "Delete GCPRole")

				IsGCPRoleDeleted(gcpRole.Name, gcpRole.Namespace)
				IsVaultGCPRoleDeleted(gcpRole.RoleName())
			})

			It("Should be successful, Create GCPAccessKeyRequest", func() {
				_, err := f.CSClient.EngineV1alpha1().GCPAccessKeyRequests(gcpAKReq.Namespace).Create(&gcpAKReq)
				Expect(err).NotTo(HaveOccurred(), "Create GCPAccessKeyRequest")

				IsGCPAccessKeyRequestCreated(gcpAKReq.Name, gcpAKReq.Namespace)
			})

			It("Should be successful, Condition approved", func() {
				r, err := f.CSClient.EngineV1alpha1().GCPAccessKeyRequests(gcpAKReq.Namespace).Create(&gcpAKReq)
				Expect(err).NotTo(HaveOccurred(), "Create GCPAccessKeyRequest")

				IsGCPAccessKeyRequestCreated(gcpAKReq.Name, gcpAKReq.Namespace)

				err = f.UpdateGCPAccessKeyRequestStatus(&api.GCPAccessKeyRequestStatus{
					Conditions: []api.GCPAccessKeyRequestCondition{
						{
							Type:           api.AccessApproved,
							LastUpdateTime: metav1.Now(),
						},
					},
				}, r)
				Expect(err).NotTo(HaveOccurred(), "Update conditions: Approved")

				IsGCPAKRConditionApproved(gcpAKReq.Name, gcpAKReq.Namespace)
			})

			It("Should be successful, Condition denied", func() {
				r, err := f.CSClient.EngineV1alpha1().GCPAccessKeyRequests(gcpAKReq.Namespace).Create(&gcpAKReq)
				Expect(err).NotTo(HaveOccurred(), "Create GCPAccessKeyRequest")

				IsGCPAccessKeyRequestCreated(gcpAKReq.Name, gcpAKReq.Namespace)

				err = f.UpdateGCPAccessKeyRequestStatus(&api.GCPAccessKeyRequestStatus{
					Conditions: []api.GCPAccessKeyRequestCondition{
						{
							Type:           api.AccessDenied,
							LastUpdateTime: metav1.Now(),
						},
					},
				}, r)
				Expect(err).NotTo(HaveOccurred(), "Update conditions: Denied")

				IsGCPAKRConditionDenied(gcpAKReq.Name, gcpAKReq.Namespace)
			})
		})

		Context("Create secret where SecretType is access_token", func() {
			var (
				secretName string
			)

			BeforeEach(func() {
				gcpRole.Spec.SecretType = api.GCPSecretAccessToken
				gcpAKReq.Status.Conditions = []api.GCPAccessKeyRequestCondition{
					{
						Type: api.AccessApproved,
					},
				}
				r, err := f.CSClient.EngineV1alpha1().GCPRoles(gcpRole.Namespace).Create(&gcpRole)
				Expect(err).NotTo(HaveOccurred(), "Create GCPRole")

				IsVaultGCPRoleCreated(r.RoleName())
				IsGCPRoleSucceeded(r.Name, r.Namespace)

			})

			AfterEach(func() {
				err := f.CSClient.EngineV1alpha1().GCPAccessKeyRequests(gcpAKReq.Namespace).Delete(gcpAKReq.Name, &metav1.DeleteOptions{})
				Expect(err).NotTo(HaveOccurred(), "Delete GCPAccessKeyRequest")

				IsGCPAccessKeyRequestDeleted(gcpAKReq.Name, gcpAKReq.Namespace)
				IsGCPAccessKeySecretDeleted(secretName, gcpAKReq.Namespace)

				err = f.CSClient.EngineV1alpha1().GCPRoles(gcpRole.Namespace).Delete(gcpRole.Name, &metav1.DeleteOptions{})
				Expect(err).NotTo(HaveOccurred(), "Delete GCPRole")

				IsGCPRoleDeleted(gcpRole.Name, gcpRole.Namespace)
				IsVaultGCPRoleDeleted(gcpRole.RoleName())
			})

			It("Should be successful, Create Access Key Secret", func() {
				_, err := f.CSClient.EngineV1alpha1().GCPAccessKeyRequests(gcpAKReq.Namespace).Create(&gcpAKReq)
				Expect(err).NotTo(HaveOccurred(), "Create GCPAccessKeyRequest")

				IsGCPAccessKeyRequestCreated(gcpAKReq.Name, gcpAKReq.Namespace)
				IsGCPAccessKeySecretCreated(gcpAKReq.Name, gcpAKReq.Namespace)

				d, err := f.CSClient.EngineV1alpha1().GCPAccessKeyRequests(gcpAKReq.Namespace).Get(gcpAKReq.Name, metav1.GetOptions{})
				Expect(err).NotTo(HaveOccurred(), "Get GCPAccessKeyRequest")
				if d.Status.Secret != nil {
					secretName = d.Status.Secret.Name
				}
			})
		})

		Context("Create secret where SecretType is service_account_key", func() {
			var (
				secretName string
			)

			BeforeEach(func() {
				gcpRole.Spec.SecretType = api.GCPSecretServiceAccountKey
				gcpAKReq.Status.Conditions = []api.GCPAccessKeyRequestCondition{
					{
						Type: api.AccessApproved,
					},
				}
				r, err := f.CSClient.EngineV1alpha1().GCPRoles(gcpRole.Namespace).Create(&gcpRole)
				Expect(err).NotTo(HaveOccurred(), "Create GCPRole")

				IsVaultGCPRoleCreated(r.RoleName())
				IsGCPRoleSucceeded(r.Name, r.Namespace)

			})

			AfterEach(func() {
				err := f.CSClient.EngineV1alpha1().GCPAccessKeyRequests(gcpAKReq.Namespace).Delete(gcpAKReq.Name, &metav1.DeleteOptions{})
				Expect(err).NotTo(HaveOccurred(), "Delete GCPAccessKeyRequest")

				IsGCPAccessKeyRequestDeleted(gcpAKReq.Name, gcpAKReq.Namespace)
				IsGCPAccessKeySecretDeleted(secretName, gcpAKReq.Namespace)

				err = f.CSClient.EngineV1alpha1().GCPRoles(gcpRole.Namespace).Delete(gcpRole.Name, &metav1.DeleteOptions{})
				Expect(err).NotTo(HaveOccurred(), "Delete GCPRole")

				IsGCPRoleDeleted(gcpRole.Name, gcpRole.Namespace)
				IsVaultGCPRoleDeleted(gcpRole.RoleName())
			})

			It("Should be successful, Create Access Key Secret", func() {
				_, err := f.CSClient.EngineV1alpha1().GCPAccessKeyRequests(gcpAKReq.Namespace).Create(&gcpAKReq)
				Expect(err).NotTo(HaveOccurred(), "Create GCPAccessKeyRequest")

				IsGCPAccessKeyRequestCreated(gcpAKReq.Name, gcpAKReq.Namespace)
				IsGCPAccessKeySecretCreated(gcpAKReq.Name, gcpAKReq.Namespace)

				d, err := f.CSClient.EngineV1alpha1().GCPAccessKeyRequests(gcpAKReq.Namespace).Get(gcpAKReq.Name, metav1.GetOptions{})
				Expect(err).NotTo(HaveOccurred(), "Get GCPAccessKeyRequest")
				if d.Status.Secret != nil {
					secretName = d.Status.Secret.Name
				}
			})
		})
	})
})
