package e2e_test

import (
	"fmt"
	"os"
	"time"

	rbac "k8s.io/api/rbac/v1"
	appcat "kmodules.xyz/custom-resources/apis/appcatalog/v1alpha1"

	api "github.com/kubevault/operator/apis/engine/v1alpha1"
	core "k8s.io/api/core/v1"

	"github.com/kubevault/operator/pkg/controller"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kubevault/operator/pkg/vault"
	"github.com/kubevault/operator/test/e2e/framework"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = FDescribe("Azure Secret Engine", func() {
	var f *framework.Invocation

	BeforeEach(func() {
		f = root.Invoke()
	})

	AfterEach(func() {
		time.Sleep(20 * time.Second)
	})

	var (
		IsVaultAzureRoleCreated = func(name string) {
			By("Checking whether vault azure role is created")
			cl, err := vault.NewClient(f.KubeClient, f.AppcatClient, f.VaultAppRef)
			Expect(err).NotTo(HaveOccurred(), "To get vault client")

			req := cl.NewRequest("GET", fmt.Sprintf("/v1/azure/roles/%s", name))
			Eventually(func() bool {
				_, err := cl.RawRequest(req)
				return err == nil
			}, timeOut, pollingInterval).Should(BeTrue(), "Vault azure role is created")

		}

		IsVaultAzureRoleDeleted = func(name string) {
			By("Checking whether vault azure role is deleted")
			cl, err := vault.NewClient(f.KubeClient, f.AppcatClient, f.VaultAppRef)
			Expect(err).NotTo(HaveOccurred(), "To get vault client")

			req := cl.NewRequest("GET", fmt.Sprintf("/v1/azure/roles/%s", name))
			Eventually(func() bool {
				_, err := cl.RawRequest(req)
				return err != nil
			}, timeOut, pollingInterval).Should(BeTrue(), "Vault azure role is deleted")

		}

		IsAzureRoleCreated = func(name, namespace string) {
			By(fmt.Sprintf("Checking whether AzureRole:(%s/%s) role is created", namespace, name))
			Eventually(func() bool {
				_, err := f.CSClient.EngineV1alpha1().AzureRoles(namespace).Get(name, metav1.GetOptions{})
				return err == nil
			}, timeOut, pollingInterval).Should(BeTrue(), "AzureRole is created")
		}

		IsAzureRoleDeleted = func(name, namespace string) {
			By(fmt.Sprintf("Checking whether AzureRole:(%s/%s) is created", namespace, name))
			Eventually(func() bool {
				_, err := f.CSClient.EngineV1alpha1().AzureRoles(namespace).Get(name, metav1.GetOptions{})
				return kerrors.IsNotFound(err)
			}, timeOut, pollingInterval).Should(BeTrue(), "AzureRole is deleted")
		}

		IsAzureRoleSucceeded = func(name, namespace string) {
			By(fmt.Sprintf("Checking whether AzureRole:(%s/%s) is succeeded", namespace, name))
			Eventually(func() bool {
				r, err := f.CSClient.EngineV1alpha1().AzureRoles(namespace).Get(name, metav1.GetOptions{})
				if err == nil {
					return r.Status.Phase == controller.AzureRolePhaseSuccess
				}
				return false
			}, timeOut, pollingInterval).Should(BeTrue(), "AzureRole status is succeeded")

		}
		IsAzureAccessKeyRequestCreated = func(name, namespace string) {
			By(fmt.Sprintf("Checking whether AzureAccessKeyRequest:(%s/%s) is created", namespace, name))
			Eventually(func() bool {
				_, err := f.CSClient.EngineV1alpha1().AzureAccessKeyRequests(namespace).Get(name, metav1.GetOptions{})
				if err == nil {
					return true
				}
				return false
			}, timeOut, pollingInterval).Should(BeTrue(), "AzureAccessKeyRequest is created")
		}
		IsAzureAccessKeyRequestDeleted = func(name, namespace string) {
			By(fmt.Sprintf("Checking whether AzureAccessKeyRequest:(%s/%s) is deleted", namespace, name))
			Eventually(func() bool {
				_, err := f.CSClient.EngineV1alpha1().AzureAccessKeyRequests(namespace).Get(name, metav1.GetOptions{})
				return kerrors.IsNotFound(err)
			}, timeOut, pollingInterval).Should(BeTrue(), "AzureAccessKeyRequest is deleted")
		}
		IsAzureAKRConditionApproved = func(name, namespace string) {
			By(fmt.Sprintf("Checking whether AzureAccessKeyRequestConditions-> Type: Approved"))
			Eventually(func() bool {
				crd, err := f.CSClient.EngineV1alpha1().AzureAccessKeyRequests(namespace).Get(name, metav1.GetOptions{})
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
		IsAzureAKRConditionDenied = func(name, namespace string) {
			By(fmt.Sprintf("Checking whether AzureAccessKeyRequestConditions-> Type: Denied"))
			Eventually(func() bool {
				crd, err := f.CSClient.EngineV1alpha1().AzureAccessKeyRequests(namespace).Get(name, metav1.GetOptions{})
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
		IsAzureAccessKeySecretCreated = func(name, namespace string) {
			By("Checking whether AzureAccessKeySecret is created")
			Eventually(func() bool {
				crd, err := f.CSClient.EngineV1alpha1().AzureAccessKeyRequests(namespace).Get(name, metav1.GetOptions{})
				if err == nil && crd.Status.Secret != nil {
					_, err2 := f.KubeClient.CoreV1().Secrets(namespace).Get(crd.Status.Secret.Name, metav1.GetOptions{})
					return err2 == nil
				}
				return false
			}, timeOut, pollingInterval).Should(BeTrue(), "AzureAccessKeySecret is created")
		}
		IsAzureAccessKeySecretDeleted = func(secretName, namespace string) {
			By("Checking whether AzureAccessKeySecret is deleted")
			Eventually(func() bool {
				_, err2 := f.KubeClient.CoreV1().Secrets(namespace).Get(secretName, metav1.GetOptions{})
				return kerrors.IsNotFound(err2)
			}, timeOut, pollingInterval).Should(BeTrue(), "AzureAccessKeySecret is deleted")
		}
	)

	Describe("AzureRole", func() {
		var (
			azureCredentials core.Secret
			azureRole        api.AzureRole
		)

		const (
			azureCredSecret = "azure-cred-3224"
			azureRoleName   = "my-azure-role-4325"
		)

		BeforeEach(func() {

			subscriptionID := os.Getenv("AZURE_SUBSCRIPTION_ID")
			tenantID := os.Getenv("AZURE_TENANT_ID")
			clientID := os.Getenv("AZURE_CLIENT_ID")
			clientSecret := os.Getenv("AZURE_CLIENT_SECRET")

			azureCredentials = core.Secret{
				TypeMeta: metav1.TypeMeta{},
				ObjectMeta: metav1.ObjectMeta{
					Name:      azureCredSecret,
					Namespace: f.Namespace(),
				},
				Data: map[string][]byte{
					api.AzureSubscriptionID: []byte(subscriptionID),
					api.AzureTenantID:       []byte(tenantID),
					api.AzureClientID:       []byte(clientID),
					api.AzureClientSecret:   []byte(clientSecret),
				},
			}
			_, err := f.KubeClient.CoreV1().Secrets(f.Namespace()).Create(&azureCredentials)
			Expect(err).NotTo(HaveOccurred(), "Create azure credentials secret")

			azureRole = api.AzureRole{
				ObjectMeta: metav1.ObjectMeta{
					Name:      azureRoleName,
					Namespace: f.Namespace(),
				},
				Spec: api.AzureRoleSpec{
					AuthManagerRef: f.VaultAppRef,
					Config: &api.AzureConfig{
						CredentialSecret: azureCredSecret,
					},
					ApplicationObjectID: "c1cb042d-96d7-423a-8dba-243c2e5010d3",
					TTL:                 "1h",
					MaxTTL:              "1h",
				},
			}
		})

		AfterEach(func() {
			err := f.KubeClient.CoreV1().Secrets(f.Namespace()).Delete(azureCredSecret, &metav1.DeleteOptions{})
			Expect(err).NotTo(HaveOccurred(), "Delete azure credentials secret")
		})

		Context("Create AzureRole", func() {
			var p api.AzureRole

			BeforeEach(func() {
				p = azureRole
			})

			AfterEach(func() {
				err := f.CSClient.EngineV1alpha1().AzureRoles(azureRole.Namespace).Delete(p.Name, &metav1.DeleteOptions{})
				Expect(err).NotTo(HaveOccurred(), "Delete AzureRole")

				IsVaultAzureRoleDeleted(p.RoleName())
				IsAzureRoleDeleted(p.Name, p.Namespace)
			})

			It("Should be successful", func() {
				_, err := f.CSClient.EngineV1alpha1().AzureRoles(p.Namespace).Create(&p)
				Expect(err).NotTo(HaveOccurred(), "Create AzureRole")

				IsAzureRoleCreated(p.Name, p.Namespace)
				IsVaultAzureRoleCreated(p.RoleName())
				IsAzureRoleSucceeded(p.Name, p.Namespace)
			})

		})

		Context("Create AzureRole with invalid vault AppReference", func() {
			var p api.AzureRole

			BeforeEach(func() {
				p = azureRole
				p.Spec.AuthManagerRef = &appcat.AppReference{
					Namespace: azureRole.Namespace,
					Name:      "invalid",
				}
			})

			AfterEach(func() {
				err := f.CSClient.EngineV1alpha1().AzureRoles(azureRole.Namespace).Delete(p.Name, &metav1.DeleteOptions{})
				Expect(err).NotTo(HaveOccurred(), "Delete AzureRole")

				IsVaultAzureRoleDeleted(p.RoleName())
				IsAzureRoleDeleted(p.Name, p.Namespace)
			})

			It("Should be successful", func() {
				_, err := f.CSClient.EngineV1alpha1().AzureRoles(p.Namespace).Create(&p)
				Expect(err).NotTo(HaveOccurred(), "Create AzureRole")

				IsAzureRoleCreated(p.Name, p.Namespace)
				IsVaultAzureRoleDeleted(p.RoleName())
			})
		})

	})

	Describe("AzureAccessKeyRequest", func() {
		var (
			azureCredentials core.Secret
			azureRole        api.AzureRole
			azureAKReq       api.AzureAccessKeyRequest
		)
		const (
			azureCredSecret = "azure-cred-2343"
			azureRoleName   = "azure-role-23432"
			azureAKReqName  = "azure-akr-324432"
		)

		BeforeEach(func() {
			subscriptionID := os.Getenv("AZURE_SUBSCRIPTION_ID")
			tenantID := os.Getenv("AZURE_TENANT_ID")
			clientID := os.Getenv("AZURE_CLIENT_ID")
			clientSecret := os.Getenv("AZURE_CLIENT_SECRET")

			azureCredentials = core.Secret{
				TypeMeta: metav1.TypeMeta{},
				ObjectMeta: metav1.ObjectMeta{
					Name:      azureCredSecret,
					Namespace: f.Namespace(),
				},
				Data: map[string][]byte{
					api.AzureSubscriptionID: []byte(subscriptionID),
					api.AzureTenantID:       []byte(tenantID),
					api.AzureClientID:       []byte(clientID),
					api.AzureClientSecret:   []byte(clientSecret),
				},
			}
			_, err := f.KubeClient.CoreV1().Secrets(f.Namespace()).Create(&azureCredentials)
			Expect(err).NotTo(HaveOccurred(), "Create azure credentials secret")

			azureRole = api.AzureRole{
				ObjectMeta: metav1.ObjectMeta{
					Name:      azureRoleName,
					Namespace: f.Namespace(),
				},
				Spec: api.AzureRoleSpec{
					AuthManagerRef: f.VaultAppRef,
					Config: &api.AzureConfig{
						CredentialSecret: azureCredSecret,
					},
					ApplicationObjectID: "c1cb042d-96d7-423a-8dba-243c2e5010d3",
					TTL:                 "1h",
					MaxTTL:              "1h",
				},
			}

			azureAKReq = api.AzureAccessKeyRequest{
				ObjectMeta: metav1.ObjectMeta{
					Name:      azureAKReqName,
					Namespace: f.Namespace(),
				},
				Spec: api.AzureAccessKeyRequestSpec{
					RoleRef: api.RoleReference{
						Name:      azureRoleName,
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
			err := f.KubeClient.CoreV1().Secrets(f.Namespace()).Delete(azureCredSecret, &metav1.DeleteOptions{})
			Expect(err).NotTo(HaveOccurred(), "Delete azure credentials secret")
		})

		Context("Create, Approve, Deny AzureAccessKeyRequests", func() {
			BeforeEach(func() {
				r, err := f.CSClient.EngineV1alpha1().AzureRoles(azureRole.Namespace).Create(&azureRole)
				Expect(err).NotTo(HaveOccurred(), "Create AzureRole")

				IsVaultAzureRoleCreated(r.RoleName())
				IsAzureRoleSucceeded(r.Name, r.Namespace)
			})

			AfterEach(func() {
				err := f.CSClient.EngineV1alpha1().AzureAccessKeyRequests(azureAKReq.Namespace).Delete(azureAKReq.Name, &metav1.DeleteOptions{})
				Expect(err).NotTo(HaveOccurred(), "Delete AzureAccessKeyRequest")

				IsAzureAccessKeyRequestDeleted(azureAKReq.Name, azureAKReq.Namespace)

				err = f.CSClient.EngineV1alpha1().AzureRoles(azureRole.Namespace).Delete(azureRole.Name, &metav1.DeleteOptions{})
				Expect(err).NotTo(HaveOccurred(), "Delete AzureRole")

				IsAzureRoleDeleted(azureRole.Name, azureRole.Namespace)
				IsVaultAzureRoleDeleted(azureRole.RoleName())
			})

			It("Should be successful, Create AzureAccessKeyRequest", func() {
				_, err := f.CSClient.EngineV1alpha1().AzureAccessKeyRequests(azureAKReq.Namespace).Create(&azureAKReq)
				Expect(err).NotTo(HaveOccurred(), "Create AzureAccessKeyRequest")

				IsAzureAccessKeyRequestCreated(azureAKReq.Name, azureAKReq.Namespace)
			})

			It("Should be successful, Condition approved", func() {
				r, err := f.CSClient.EngineV1alpha1().AzureAccessKeyRequests(azureAKReq.Namespace).Create(&azureAKReq)
				Expect(err).NotTo(HaveOccurred(), "Create AzureAccessKeyRequest")

				IsAzureAccessKeyRequestCreated(azureAKReq.Name, azureAKReq.Namespace)

				err = f.UpdateAzureAccessKeyRequestStatus(&api.AzureAccessKeyRequestStatus{
					Conditions: []api.AzureAccessKeyRequestCondition{
						{
							Type:           api.AccessApproved,
							LastUpdateTime: metav1.Now(),
						},
					},
				}, r)
				Expect(err).NotTo(HaveOccurred(), "Update conditions: Approved")

				IsAzureAKRConditionApproved(azureAKReq.Name, azureAKReq.Namespace)
			})

			It("Should be successful, Condition denied", func() {
				r, err := f.CSClient.EngineV1alpha1().AzureAccessKeyRequests(azureAKReq.Namespace).Create(&azureAKReq)
				Expect(err).NotTo(HaveOccurred(), "Create AzureAccessKeyRequest")

				IsAzureAccessKeyRequestCreated(azureAKReq.Name, azureAKReq.Namespace)

				err = f.UpdateAzureAccessKeyRequestStatus(&api.AzureAccessKeyRequestStatus{
					Conditions: []api.AzureAccessKeyRequestCondition{
						{
							Type:           api.AccessDenied,
							LastUpdateTime: metav1.Now(),
						},
					},
				}, r)
				Expect(err).NotTo(HaveOccurred(), "Update conditions: Denied")

				IsAzureAKRConditionDenied(azureAKReq.Name, azureAKReq.Namespace)
			})
		})

		Context("Create azure secret", func() {
			var (
				secretName string
			)

			BeforeEach(func() {
				azureAKReq.Status.Conditions = []api.AzureAccessKeyRequestCondition{
					{
						Type: api.AccessApproved,
					},
				}
				r, err := f.CSClient.EngineV1alpha1().AzureRoles(azureRole.Namespace).Create(&azureRole)
				Expect(err).NotTo(HaveOccurred(), "Create AzureRole")

				IsVaultAzureRoleCreated(r.RoleName())
				IsAzureRoleSucceeded(r.Name, r.Namespace)
			})

			AfterEach(func() {
				err := f.CSClient.EngineV1alpha1().AzureAccessKeyRequests(azureAKReq.Namespace).Delete(azureAKReq.Name, &metav1.DeleteOptions{})
				Expect(err).NotTo(HaveOccurred(), "Delete AzureAccessKeyRequest")

				IsAzureAccessKeyRequestDeleted(azureAKReq.Name, azureAKReq.Namespace)
				IsAzureAccessKeySecretDeleted(secretName, azureAKReq.Namespace)

				err = f.CSClient.EngineV1alpha1().AzureRoles(azureRole.Namespace).Delete(azureRole.Name, &metav1.DeleteOptions{})
				Expect(err).NotTo(HaveOccurred(), "Delete AzureRole")

				IsAzureRoleDeleted(azureRole.Name, azureRole.Namespace)
				IsVaultAzureRoleDeleted(azureRole.RoleName())
			})

			It("Should be successful, Create Access Key Secret", func() {
				_, err := f.CSClient.EngineV1alpha1().AzureAccessKeyRequests(azureAKReq.Namespace).Create(&azureAKReq)
				Expect(err).NotTo(HaveOccurred(), "Create AzureAccessKeyRequest")

				IsAzureAccessKeyRequestCreated(azureAKReq.Name, azureAKReq.Namespace)
				IsAzureAccessKeySecretCreated(azureAKReq.Name, azureAKReq.Namespace)

				d, err := f.CSClient.EngineV1alpha1().AzureAccessKeyRequests(azureAKReq.Namespace).Get(azureAKReq.Name, metav1.GetOptions{})
				Expect(err).NotTo(HaveOccurred(), "Get AzureAccessKeyRequest")
				if d.Status.Secret != nil {
					secretName = d.Status.Secret.Name
				}
			})
		})
	})
})
