/*
Copyright The KubeVault Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package e2e_test

import (
	"context"
	"fmt"
	"time"

	api "kubevault.dev/operator/apis/engine/v1alpha1"
	"kubevault.dev/operator/pkg/controller"
	"kubevault.dev/operator/test/e2e/framework"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	core "k8s.io/api/core/v1"
	v1 "k8s.io/api/rbac/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kmapi "kmodules.xyz/client-go/api/v1"
	azureconsts "kmodules.xyz/constants/azure"
)

var _ = Describe("Azure Secret Engine", func() {

	var f *framework.Invocation

	var (
		IsSecretEngineCreated = func(name, namespace string) {
			By(fmt.Sprintf("Checking whether SecretEngine:(%s/%s) is created", namespace, name))
			Eventually(func() bool {
				_, err := f.CSClient.EngineV1alpha1().SecretEngines(namespace).Get(name, metav1.GetOptions{})
				return err == nil
			}, timeOut, pollingInterval).Should(BeTrue(), "SecretEngine is created")
		}
		IsSecretEngineDeleted = func(name, namespace string) {
			By(fmt.Sprintf("Checking whether SecretEngine:(%s/%s) is deleted", namespace, name))
			Eventually(func() bool {
				_, err := f.CSClient.EngineV1alpha1().SecretEngines(namespace).Get(name, metav1.GetOptions{})
				return kerrors.IsNotFound(err)
			}, timeOut, pollingInterval).Should(BeTrue(), "SecretEngine is deleted")
		}
		IsSecretEngineSucceeded = func(name, namespace string) {
			By(fmt.Sprintf("Checking whether SecretEngine:(%s/%s) is succeeded", namespace, name))
			Eventually(func() bool {
				r, err := f.CSClient.EngineV1alpha1().SecretEngines(namespace).Get(name, metav1.GetOptions{})
				return err == nil && r.Status.Phase == controller.SecretEnginePhaseSuccess

			}, timeOut, pollingInterval).Should(BeTrue(), "SecretEngine status is succeeded")

		}
		IsAzureRoleCreated = func(name, namespace string) {
			By(fmt.Sprintf("Checking whether AzureRole:(%s/%s) role is created", namespace, name))
			Eventually(func() bool {
				_, err := f.CSClient.EngineV1alpha1().AzureRoles(namespace).Get(name, metav1.GetOptions{})
				return err == nil
			}, timeOut, pollingInterval).Should(BeTrue(), "AzureRole is created")
		}
		IsAzureRoleDeleted = func(name, namespace string) {
			By(fmt.Sprintf("Checking whether AzureRole:(%s/%s) is deleted", namespace, name))
			Eventually(func() bool {
				_, err := f.CSClient.EngineV1alpha1().AzureRoles(namespace).Get(name, metav1.GetOptions{})
				return kerrors.IsNotFound(err)
			}, timeOut, pollingInterval).Should(BeTrue(), "AzureRole is deleted")
		}
		IsAzureRoleSucceeded = func(name, namespace string) {
			By(fmt.Sprintf("Checking whether AzureRole:(%s/%s) is succeeded", namespace, name))
			Eventually(func() bool {
				r, err := f.CSClient.EngineV1alpha1().AzureRoles(namespace).Get(name, metav1.GetOptions{})
				return err == nil && r.Status.Phase == controller.AzureRolePhaseSuccess

			}, timeOut, pollingInterval).Should(BeTrue(), "AzureRole status is succeeded")

		}

		IsAzureRoleFailed = func(name, namespace string) {
			By(fmt.Sprintf("Checking whether AzureRole:(%s/%s) is failed", namespace, name))
			Eventually(func() bool {
				r, err := f.CSClient.EngineV1alpha1().AzureRoles(namespace).Get(name, metav1.GetOptions{})
				return err == nil && r.Status.Phase != controller.AzureRolePhaseSuccess && len(r.Status.Conditions) != 0

			}, timeOut, pollingInterval).Should(BeTrue(), "AzureRole status is failed")
		}
		IsAzureAccessKeyRequestCreated = func(name, namespace string) {
			By(fmt.Sprintf("Checking whether AzureAccessKeyRequest:(%s/%s) is created", namespace, name))
			Eventually(func() bool {
				_, err := f.CSClient.EngineV1alpha1().AzureAccessKeyRequests(namespace).Get(name, metav1.GetOptions{})
				return err == nil
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
						if value.Type == kmapi.ConditionRequestApproved {
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
						if value.Type == kmapi.ConditionRequestDenied {
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
					_, err2 := f.KubeClient.CoreV1().Secrets(namespace).Get(context.TODO(), crd.Status.Secret.Name, metav1.GetOptions{})
					return err2 == nil
				}
				return false
			}, timeOut, pollingInterval).Should(BeTrue(), "AzureAccessKeySecret is created")
		}
		IsAzureAccessKeySecretDeleted = func(secretName, namespace string) {
			By("Checking whether AzureAccessKeySecret is deleted")
			Eventually(func() bool {
				_, err2 := f.KubeClient.CoreV1().Secrets(namespace).Get(context.TODO(), secretName, metav1.GetOptions{})
				return kerrors.IsNotFound(err2)
			}, timeOut, pollingInterval).Should(BeTrue(), "AzureAccessKeySecret is deleted")
		}
	)

	BeforeEach(func() {
		f = root.Invoke()
		if !framework.SelfHostedOperator {
			Skip("Skipping AzureRole test because the operator isn't running inside cluster")
		}
		// vault server creates appBinding, vault policy, and policy binding
		time.Sleep(20 * time.Second)
	})

	AfterEach(func() {
		time.Sleep(20 * time.Second)
	})

	Describe("AzureRole", func() {

		var (
			azureCredentials core.Secret
			azureRole        api.AzureRole
			azureSE          api.SecretEngine
		)

		const (
			azureCredSecret   = "azure-cred-3224"
			azureRoleName     = "my-azure-role-4325"
			azureSecretEngine = "my-azure-secretengine-3423423"
		)

		BeforeEach(func() {
			credentials := azureconsts.CredentialsFromEnv()
			if len(credentials) == 0 {
				Skip("Skipping azure secret engine tests, empty env")
			}

			azureCredentials = core.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      azureCredSecret,
					Namespace: f.Namespace(),
				},
				Data: credentials,
			}
			_, err := f.KubeClient.CoreV1().Secrets(f.Namespace()).Create(context.TODO(), &azureCredentials, metav1.CreateOptions{})
			Expect(err).NotTo(HaveOccurred(), "Create azure credentials secret")

			azureRole = api.AzureRole{
				ObjectMeta: metav1.ObjectMeta{
					Name:      azureRoleName,
					Namespace: f.Namespace(),
				},
				Spec: api.AzureRoleSpec{
					VaultRef: core.LocalObjectReference{
						Name: f.VaultAppRef.Name,
					},
					ApplicationObjectID: "c1cb042d-96d7-423a-8dba-243c2e5010d3",
				},
			}

			azureSE = api.SecretEngine{
				ObjectMeta: metav1.ObjectMeta{
					Name:      azureSecretEngine,
					Namespace: f.Namespace(),
				},
				Spec: api.SecretEngineSpec{
					VaultRef: core.LocalObjectReference{
						Name: f.VaultAppRef.Name,
					},
					Path: "azure",
					SecretEngineConfiguration: api.SecretEngineConfiguration{
						Azure: &api.AzureConfiguration{
							CredentialSecret: azureCredSecret,
						},
					},
				},
			}
		})

		AfterEach(func() {
			err := f.KubeClient.CoreV1().Secrets(f.Namespace()).Delete(context.TODO(), azureCredSecret, metav1.DeleteOptions{})
			Expect(err).NotTo(HaveOccurred(), "Delete Azure credentials secret")
		})

		Context("Create AzureRole", func() {
			var p api.AzureRole
			var se api.SecretEngine

			BeforeEach(func() {
				p = azureRole
				se = azureSE
			})

			AfterEach(func() {
				By("Deleting AzureRole...")
				err := f.CSClient.EngineV1alpha1().AzureRoles(azureRole.Namespace).Delete(p.Name, metav1.DeleteOptions{})
				Expect(err).NotTo(HaveOccurred(), "Delete AzureRole")

				IsAzureRoleDeleted(p.Name, p.Namespace)

				By("Deleting SecretEngine...")
				err = f.CSClient.EngineV1alpha1().SecretEngines(se.Namespace).Delete(se.Name, metav1.DeleteOptions{})
				Expect(err).NotTo(HaveOccurred(), "Delete Secret engine")

				IsSecretEngineDeleted(se.Name, se.Namespace)
			})

			It("Should be successful", func() {
				By("Creating SecretEngine...")
				_, err := f.CSClient.EngineV1alpha1().SecretEngines(se.Namespace).Create(&se)
				Expect(err).NotTo(HaveOccurred(), "Create SecretEngine")

				IsSecretEngineCreated(se.Name, se.Namespace)
				IsSecretEngineSucceeded(se.Name, se.Namespace)

				By("Creating AzureRole...")
				_, err = f.CSClient.EngineV1alpha1().AzureRoles(p.Namespace).Create(&p)
				Expect(err).NotTo(HaveOccurred(), "Create AzureRole")

				IsAzureRoleCreated(p.Name, p.Namespace)
				IsAzureRoleSucceeded(p.Name, p.Namespace)
			})

		})

		Context("Create AzureRole without enabling secretEngine", func() {
			var p api.AzureRole

			BeforeEach(func() {
				p = azureRole
			})

			AfterEach(func() {
				By("Deleting AzureRole...")
				err := f.CSClient.EngineV1alpha1().AzureRoles(azureRole.Namespace).Delete(p.Name, metav1.DeleteOptions{})
				Expect(err).NotTo(HaveOccurred(), "Delete AzureRole")

				IsAzureRoleDeleted(p.Name, p.Namespace)

			})

			It("Should be failed making AzureRole", func() {

				By("Creating AzureRole...")
				_, err := f.CSClient.EngineV1alpha1().AzureRoles(p.Namespace).Create(&p)
				Expect(err).NotTo(HaveOccurred(), "Create AzureRole")

				IsAzureRoleCreated(p.Name, p.Namespace)
				IsAzureRoleFailed(p.Name, p.Namespace)
			})
		})

	})

	Describe("AzureAccessKeyRequest", func() {

		var (
			azureCredentials core.Secret
			azureRole        api.AzureRole
			azureSE          api.SecretEngine
			azureAKR         api.AzureAccessKeyRequest
		)

		const (
			azureCredSecret   = "azure-cred-3224"
			azureRoleName     = "my-azure-roleset-4325"
			azureSecretEngine = "my-azure-secretengine-3423423"
			azureAKRName      = "my-azure-token-2345"
		)

		BeforeEach(func() {
			credentials := azureconsts.CredentialsFromEnv()
			if len(credentials) == 0 {
				Skip("Skipping azure secret engine tests, empty env")
			}

			azureCredentials = core.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      azureCredSecret,
					Namespace: f.Namespace(),
				},
				Data: credentials,
			}
			_, err := f.KubeClient.CoreV1().Secrets(f.Namespace()).Create(context.TODO(), &azureCredentials, metav1.CreateOptions{})
			Expect(err).NotTo(HaveOccurred(), "Create azure credentials secret")

			azureSE = api.SecretEngine{
				ObjectMeta: metav1.ObjectMeta{
					Name:      azureSecretEngine,
					Namespace: f.Namespace(),
				},
				Spec: api.SecretEngineSpec{
					VaultRef: core.LocalObjectReference{
						Name: f.VaultAppRef.Name,
					},
					Path: "azure",
					SecretEngineConfiguration: api.SecretEngineConfiguration{
						Azure: &api.AzureConfiguration{
							CredentialSecret: azureCredSecret,
						},
					},
				},
			}
			_, err = f.CSClient.EngineV1alpha1().SecretEngines(azureSE.Namespace).Create(&azureSE)
			Expect(err).NotTo(HaveOccurred(), "Create azure SecretEngine")
			IsSecretEngineCreated(azureSE.Name, azureSE.Namespace)

			azureRole = api.AzureRole{
				ObjectMeta: metav1.ObjectMeta{
					Name:      azureRoleName,
					Namespace: f.Namespace(),
				},
				Spec: api.AzureRoleSpec{
					VaultRef: core.LocalObjectReference{
						Name: f.VaultAppRef.Name,
					},
					ApplicationObjectID: "c1cb042d-96d7-423a-8dba-243c2e5010d3",
				},
			}

			azureAKR = api.AzureAccessKeyRequest{
				ObjectMeta: metav1.ObjectMeta{
					Name:      azureAKRName,
					Namespace: f.Namespace(),
				},
				Spec: api.AzureAccessKeyRequestSpec{
					RoleRef: api.RoleRef{
						Name:      azureRoleName,
						Namespace: f.Namespace(),
					},
					Subjects: []v1.Subject{
						{
							Kind:      "ServiceAccount",
							Name:      "sa",
							Namespace: "demo",
						},
					},
				},
			}
		})

		AfterEach(func() {
			err := f.KubeClient.CoreV1().Secrets(f.Namespace()).Delete(context.TODO(), azureCredSecret, metav1.DeleteOptions{})
			Expect(err).NotTo(HaveOccurred(), "Delete azure credentials secret")

			err = f.CSClient.EngineV1alpha1().SecretEngines(azureSE.Namespace).Delete(azureSE.Name, metav1.DeleteOptions{})
			Expect(err).NotTo(HaveOccurred(), "Delete azure SecretEngine")
			IsSecretEngineDeleted(azureSE.Name, azureSE.Namespace)
		})

		Context("Create, Approve, Deny AzureAccessKeyRequests", func() {
			BeforeEach(func() {
				_, err := f.CSClient.EngineV1alpha1().AzureRoles(azureRole.Namespace).Create(&azureRole)
				Expect(err).NotTo(HaveOccurred(), "Create AzureRole")

				IsAzureRoleCreated(azureRole.Name, azureRole.Namespace)
				IsAzureRoleSucceeded(azureRole.Name, azureRole.Namespace)

			})

			AfterEach(func() {
				err := f.CSClient.EngineV1alpha1().AzureAccessKeyRequests(azureAKR.Namespace).Delete(azureAKR.Name, metav1.DeleteOptions{})
				Expect(err).NotTo(HaveOccurred(), "Delete azureAccessKeyRequest")
				IsAzureAccessKeyRequestDeleted(azureAKR.Name, azureAKR.Namespace)

				err = f.CSClient.EngineV1alpha1().AzureRoles(azureRole.Namespace).Delete(azureRole.Name, metav1.DeleteOptions{})
				Expect(err).NotTo(HaveOccurred(), "Delete azureRole")
				IsAzureRoleDeleted(azureRole.Name, azureRole.Namespace)
			})

			It("Should be successful, Create AzureAccessKeyRequest", func() {
				_, err := f.CSClient.EngineV1alpha1().AzureAccessKeyRequests(azureAKR.Namespace).Create(&azureAKR)
				Expect(err).NotTo(HaveOccurred(), "Create AzureAccessKeyRequest")

				IsAzureAccessKeyRequestCreated(azureAKR.Name, azureAKR.Namespace)
			})

			It("Should be successful, Condition approved", func() {
				By("Creating AzureAccessKeyRequest...")
				r, err := f.CSClient.EngineV1alpha1().AzureAccessKeyRequests(azureAKR.Namespace).Create(&azureAKR)
				Expect(err).NotTo(HaveOccurred(), "Create AzureAccessKeyRequest")

				IsAzureAccessKeyRequestCreated(azureAKR.Name, azureAKR.Namespace)

				By("Updating Azure AccessKeyRequest status...")
				err = f.UpdateAzureAccessKeyRequestStatus(&api.AzureAccessKeyRequestStatus{
					Conditions: []kmapi.Condition{
						{
							Type:               kmapi.ConditionRequestApproved,
							LastTransitionTime: metav1.Now(),
						},
					},
				}, r)
				Expect(err).NotTo(HaveOccurred(), "Update conditions: Approved")
				IsAzureAKRConditionApproved(azureAKR.Name, azureAKR.Namespace)
			})

			It("Should be successful, Condition denied", func() {
				By("Creating AzureAccessKeyRequest...")
				r, err := f.CSClient.EngineV1alpha1().AzureAccessKeyRequests(azureAKR.Namespace).Create(&azureAKR)
				Expect(err).NotTo(HaveOccurred(), "Create AzureAccessKeyRequest")

				IsAzureAccessKeyRequestCreated(azureAKR.Name, azureAKR.Namespace)

				By("Updating Azure AccessKeyRequest status...")
				err = f.UpdateAzureAccessKeyRequestStatus(&api.AzureAccessKeyRequestStatus{
					Conditions: []kmapi.Condition{
						{
							Type:               kmapi.ConditionRequestDenied,
							LastTransitionTime: metav1.Now(),
						},
					},
				}, r)
				Expect(err).NotTo(HaveOccurred(), "Update conditions: Denied")

				IsAzureAKRConditionDenied(azureAKR.Name, azureAKR.Namespace)
			})
		})

		Context("Generate azure service principals from Vault", func() {
			var (
				secretName string
			)

			BeforeEach(func() {

				By("Creating AzureRole...")
				r, err := f.CSClient.EngineV1alpha1().AzureRoles(azureRole.Namespace).Create(&azureRole)
				Expect(err).NotTo(HaveOccurred(), "Create AzureRole")

				IsAzureRoleSucceeded(r.Name, r.Namespace)

			})

			AfterEach(func() {
				By("Deleting azure accesskeyrequest...")
				err := f.CSClient.EngineV1alpha1().AzureAccessKeyRequests(azureAKR.Namespace).Delete(azureAKR.Name, metav1.DeleteOptions{})
				Expect(err).NotTo(HaveOccurred(), "Delete AzureAccessKeyRequest")

				IsAzureAccessKeyRequestDeleted(azureAKR.Name, azureAKR.Namespace)
				IsAzureAccessKeySecretDeleted(secretName, azureAKR.Namespace)

				By("Deleting azureRole...")
				err = f.CSClient.EngineV1alpha1().AzureRoles(azureRole.Namespace).Delete(azureRole.Name, metav1.DeleteOptions{})
				Expect(err).NotTo(HaveOccurred(), "Delete AzureRole")

				IsAzureRoleDeleted(azureRole.Name, azureRole.Namespace)
			})

			It("Should be successful, Create Access Key Secret", func() {
				By("Creating Azure accessKeyRequest...")
				r, err := f.CSClient.EngineV1alpha1().AzureAccessKeyRequests(azureAKR.Namespace).Create(&azureAKR)
				Expect(err).NotTo(HaveOccurred(), "Create AzureAccessKeyRequest")

				IsAzureAccessKeyRequestCreated(azureAKR.Name, azureAKR.Namespace)

				By("Updating Azure AccessKeyRequest status...")
				err = f.UpdateAzureAccessKeyRequestStatus(&api.AzureAccessKeyRequestStatus{
					Conditions: []kmapi.Condition{
						{
							Type:               kmapi.ConditionRequestApproved,
							LastTransitionTime: metav1.Now(),
						},
					},
				}, r)

				Expect(err).NotTo(HaveOccurred(), "Update conditions: Approved")
				IsAzureAKRConditionApproved(azureAKR.Name, azureAKR.Namespace)

				IsAzureAccessKeySecretCreated(azureAKR.Name, azureAKR.Namespace)

				d, err := f.CSClient.EngineV1alpha1().AzureAccessKeyRequests(azureAKR.Namespace).Get(azureAKR.Name, metav1.GetOptions{})
				Expect(err).NotTo(HaveOccurred(), "Get AzureAccessKeyRequest")
				if d.Status.Secret != nil {
					secretName = d.Status.Secret.Name
				}
			})
		})

	})
})
