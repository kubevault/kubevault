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
	"kmodules.xyz/custom-resources/apis/appcatalog/v1alpha1"
)

var _ = Describe("Postgres Secret Engine", func() {

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
				if err == nil {
					return r.Status.Phase == controller.SecretEnginePhaseSuccess
				}
				return false
			}, timeOut, pollingInterval).Should(BeTrue(), "SecretEngine status is succeeded")

		}
		IsPostgresRoleCreated = func(name, namespace string) {
			By(fmt.Sprintf("Checking whether PostgresRole:(%s/%s) role is created", namespace, name))
			Eventually(func() bool {
				_, err := f.CSClient.EngineV1alpha1().PostgresRoles(namespace).Get(name, metav1.GetOptions{})
				return err == nil
			}, timeOut, pollingInterval).Should(BeTrue(), "PostgresRole is created")
		}
		IsPostgresRoleDeleted = func(name, namespace string) {
			By(fmt.Sprintf("Checking whether PostgresRole:(%s/%s) is deleted", namespace, name))
			Eventually(func() bool {
				_, err := f.CSClient.EngineV1alpha1().PostgresRoles(namespace).Get(name, metav1.GetOptions{})
				return kerrors.IsNotFound(err)
			}, timeOut, pollingInterval).Should(BeTrue(), "PostgresRole is deleted")
		}
		IsPostgresRoleSucceeded = func(name, namespace string) {
			By(fmt.Sprintf("Checking whether PostgresRole:(%s/%s) is succeeded", namespace, name))
			Eventually(func() bool {
				r, err := f.CSClient.EngineV1alpha1().PostgresRoles(namespace).Get(name, metav1.GetOptions{})
				if err == nil {
					return r.Status.Phase == controller.PostgresRolePhaseSuccess
				}
				return false
			}, timeOut, pollingInterval).Should(BeTrue(), "PostgresRole status is succeeded")

		}

		IsPostgresRoleFailed = func(name, namespace string) {
			By(fmt.Sprintf("Checking whether PostgresRole:(%s/%s) is failed", namespace, name))
			Eventually(func() bool {
				r, err := f.CSClient.EngineV1alpha1().PostgresRoles(namespace).Get(name, metav1.GetOptions{})
				if err == nil {
					return r.Status.Phase != controller.PostgresRolePhaseSuccess && len(r.Status.Conditions) != 0
				}
				return false
			}, timeOut, pollingInterval).Should(BeTrue(), "PostgresRole status is failed")
		}
		IsDatabaseAccessRequestCreated = func(name, namespace string) {
			By(fmt.Sprintf("Checking whether DatabaseAccessRequest:(%s/%s) is created", namespace, name))
			Eventually(func() bool {
				_, err := f.CSClient.EngineV1alpha1().DatabaseAccessRequests(namespace).Get(name, metav1.GetOptions{})
				return err == nil
			}, timeOut, pollingInterval).Should(BeTrue(), "DatabaseAccessRequest is created")
		}
		IsDatabaseAccessRequestDeleted = func(name, namespace string) {
			By(fmt.Sprintf("Checking whether DatabaseAccessRequest:(%s/%s) is deleted", namespace, name))
			Eventually(func() bool {
				_, err := f.CSClient.EngineV1alpha1().DatabaseAccessRequests(namespace).Get(name, metav1.GetOptions{})
				return kerrors.IsNotFound(err)
			}, timeOut, pollingInterval).Should(BeTrue(), "DatabaseAccessRequest is deleted")
		}
		IsPostgresAKRConditionApproved = func(name, namespace string) {
			By(fmt.Sprintf("Checking whether DatabaseAccessRequestConditions-> Type: Approved"))
			Eventually(func() bool {
				crd, err := f.CSClient.EngineV1alpha1().DatabaseAccessRequests(namespace).Get(name, metav1.GetOptions{})
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
		IsPostgresAKRConditionDenied = func(name, namespace string) {
			By(fmt.Sprintf("Checking whether DatabaseAccessRequestConditions-> Type: Denied"))
			Eventually(func() bool {
				crd, err := f.CSClient.EngineV1alpha1().DatabaseAccessRequests(namespace).Get(name, metav1.GetOptions{})
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
		IsPostgresAccessKeySecretCreated = func(name, namespace string) {
			By("Checking whether PostgresAccessKeySecret is created")
			Eventually(func() bool {
				crd, err := f.CSClient.EngineV1alpha1().DatabaseAccessRequests(namespace).Get(name, metav1.GetOptions{})
				if err == nil && crd.Status.Secret != nil {
					_, err2 := f.KubeClient.CoreV1().Secrets(namespace).Get(crd.Status.Secret.Name, metav1.GetOptions{})
					return err2 == nil
				}
				return false
			}, timeOut, pollingInterval).Should(BeTrue(), "PostgresAccessKeySecret is created")
		}
		IsPostgresAccessKeySecretDeleted = func(secretName, namespace string) {
			By("Checking whether PostgresAccessKeySecret is deleted")
			Eventually(func() bool {
				_, err2 := f.KubeClient.CoreV1().Secrets(namespace).Get(secretName, metav1.GetOptions{})
				return kerrors.IsNotFound(err2)
			}, timeOut, pollingInterval).Should(BeTrue(), "PostgresAccessKeySecret is deleted")
		}
	)

	BeforeEach(func() {
		f = root.Invoke()
		if !framework.SelfHostedOperator {
			Skip("Skipping Postgres secret engine tests because the operator isn't running inside cluster")
		}
	})

	AfterEach(func() {
		time.Sleep(20 * time.Second)
	})

	Describe("PostgresRole", func() {

		var (
			postgresRole api.PostgresRole
			postgresSE   api.SecretEngine
		)

		const (
			postgresRoleName     = "my-postgres-role-4325"
			postgresSecretEngine = "my-postgres-secretengine-3423423"
		)

		BeforeEach(func() {

			postgresRole = api.PostgresRole{
				ObjectMeta: metav1.ObjectMeta{
					Name:      postgresRoleName,
					Namespace: f.Namespace(),
				},
				Spec: api.PostgresRoleSpec{
					VaultRef: core.LocalObjectReference{
						Name: f.VaultAppRef.Name,
					},
					DatabaseRef: f.PostgresAppRef,

					CreationStatements: []string{
						"CREATE ROLE \"{{name}}\" WITH LOGIN PASSWORD '{{password}}' VALID UNTIL '{{expiration}}';",
						"GRANT SELECT ON ALL TABLES IN SCHEMA public TO \"{{name}}\";",
					},
					MaxTTL:     "1h",
					DefaultTTL: "300",
				},
			}

			postgresSE = api.SecretEngine{
				ObjectMeta: metav1.ObjectMeta{
					Name:      postgresSecretEngine,
					Namespace: f.Namespace(),
				},
				Spec: api.SecretEngineSpec{
					VaultRef: core.LocalObjectReference{
						Name: f.VaultAppRef.Name,
					},
					Path: "database",
					SecretEngineConfiguration: api.SecretEngineConfiguration{
						Postgres: &api.PostgresConfiguration{
							DatabaseRef: v1alpha1.AppReference{
								Name:      f.PostgresAppRef.Name,
								Namespace: f.PostgresAppRef.Namespace,
							},
						},
					},
				},
			}
		})

		Context("Create PostgresRole", func() {
			var p api.PostgresRole
			var se api.SecretEngine

			BeforeEach(func() {
				p = postgresRole
				se = postgresSE
			})

			AfterEach(func() {
				By("Deleting PostgresRole...")
				err := f.CSClient.EngineV1alpha1().PostgresRoles(postgresRole.Namespace).Delete(p.Name, &metav1.DeleteOptions{})
				Expect(err).NotTo(HaveOccurred(), "Delete PostgresRole")

				IsPostgresRoleDeleted(p.Name, p.Namespace)

				By("Deleting SecretEngine...")
				err = f.CSClient.EngineV1alpha1().SecretEngines(se.Namespace).Delete(se.Name, &metav1.DeleteOptions{})
				Expect(err).NotTo(HaveOccurred(), "Delete Secret engine")

				IsSecretEngineDeleted(se.Name, se.Namespace)
			})

			It("Should be successful", func() {
				By("Creating SecretEngine...")
				_, err := f.CSClient.EngineV1alpha1().SecretEngines(se.Namespace).Create(&se)
				Expect(err).NotTo(HaveOccurred(), "Create SecretEngine")

				IsSecretEngineCreated(se.Name, se.Namespace)
				IsSecretEngineSucceeded(se.Name, se.Namespace)

				By("Creating PostgresRole...")
				_, err = f.CSClient.EngineV1alpha1().PostgresRoles(p.Namespace).Create(&p)
				Expect(err).NotTo(HaveOccurred(), "Create PostgresRole")

				IsPostgresRoleCreated(p.Name, p.Namespace)
				IsPostgresRoleSucceeded(p.Name, p.Namespace)
			})

		})

		Context("Create PostgresRole without enabling secretEngine", func() {
			var p api.PostgresRole

			BeforeEach(func() {
				p = postgresRole
			})

			AfterEach(func() {
				By("Deleting PostgresRole...")
				err := f.CSClient.EngineV1alpha1().PostgresRoles(postgresRole.Namespace).Delete(p.Name, &metav1.DeleteOptions{})
				Expect(err).NotTo(HaveOccurred(), "Delete PostgresRole")

				IsPostgresRoleDeleted(p.Name, p.Namespace)

			})

			It("Should be failed making PostgresRole", func() {

				By("Creating PostgresRole...")
				_, err := f.CSClient.EngineV1alpha1().PostgresRoles(p.Namespace).Create(&p)
				Expect(err).NotTo(HaveOccurred(), "Create PostgresRole")

				IsPostgresRoleCreated(p.Name, p.Namespace)
				IsPostgresRoleFailed(p.Name, p.Namespace)
			})
		})

	})

	Describe("DatabaseAccessRequest", func() {

		var (
			postgresRole api.PostgresRole
			postgresSE   api.SecretEngine
			postgresAKR  api.DatabaseAccessRequest
		)

		const (
			postgresRoleName     = "my-postgres-role-4325"
			postgresSecretEngine = "my-postgres-secretengine-3423423"
			postgresAKRName      = "my-postgres-token-2345"
		)

		BeforeEach(func() {

			postgresSE = api.SecretEngine{
				ObjectMeta: metav1.ObjectMeta{
					Name:      postgresSecretEngine,
					Namespace: f.Namespace(),
				},
				Spec: api.SecretEngineSpec{
					VaultRef: core.LocalObjectReference{
						Name: f.VaultAppRef.Name,
					},
					Path: "database",
					SecretEngineConfiguration: api.SecretEngineConfiguration{
						Postgres: &api.PostgresConfiguration{
							DatabaseRef: v1alpha1.AppReference{
								Name:      f.PostgresAppRef.Name,
								Namespace: f.PostgresAppRef.Namespace,
							},
						},
					},
				},
			}
			_, err := f.CSClient.EngineV1alpha1().SecretEngines(postgresSE.Namespace).Create(&postgresSE)
			Expect(err).NotTo(HaveOccurred(), "Create Postgres SecretEngine")
			IsSecretEngineCreated(postgresSE.Name, postgresSE.Namespace)

			postgresRole = api.PostgresRole{
				ObjectMeta: metav1.ObjectMeta{
					Name:      postgresRoleName,
					Namespace: f.Namespace(),
				},
				Spec: api.PostgresRoleSpec{
					VaultRef: core.LocalObjectReference{
						Name: f.VaultAppRef.Name,
					},
					DatabaseRef: f.PostgresAppRef,

					CreationStatements: []string{
						"CREATE ROLE \"{{name}}\" WITH LOGIN PASSWORD '{{password}}' VALID UNTIL '{{expiration}}';",
						"GRANT SELECT ON ALL TABLES IN SCHEMA public TO \"{{name}}\";",
					},
					MaxTTL:     "1h",
					DefaultTTL: "300",
				},
			}

			postgresAKR = api.DatabaseAccessRequest{
				ObjectMeta: metav1.ObjectMeta{
					Name:      postgresAKRName,
					Namespace: f.Namespace(),
				},
				Spec: api.DatabaseAccessRequestSpec{
					RoleRef: api.RoleRef{
						Kind:      api.ResourceKindPostgresRole,
						Name:      postgresRoleName,
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
			err := f.CSClient.EngineV1alpha1().SecretEngines(postgresSE.Namespace).Delete(postgresSE.Name, &metav1.DeleteOptions{})
			Expect(err).NotTo(HaveOccurred(), "Delete Postgres SecretEngine")
			IsSecretEngineDeleted(postgresSE.Name, postgresSE.Namespace)
		})

		Context("Create, Approve, Deny DatabaseAccessRequests", func() {
			BeforeEach(func() {
				_, err := f.CSClient.EngineV1alpha1().PostgresRoles(postgresRole.Namespace).Create(&postgresRole)
				Expect(err).NotTo(HaveOccurred(), "Create PostgresRole")

				IsPostgresRoleCreated(postgresRole.Name, postgresRole.Namespace)
				IsPostgresRoleSucceeded(postgresRole.Name, postgresRole.Namespace)

			})

			AfterEach(func() {
				err := f.CSClient.EngineV1alpha1().DatabaseAccessRequests(postgresAKR.Namespace).Delete(postgresAKR.Name, &metav1.DeleteOptions{})
				Expect(err).NotTo(HaveOccurred(), "Delete DatabaseAccessRequest")
				IsDatabaseAccessRequestDeleted(postgresAKR.Name, postgresAKR.Namespace)

				err = f.CSClient.EngineV1alpha1().PostgresRoles(postgresRole.Namespace).Delete(postgresRole.Name, &metav1.DeleteOptions{})
				Expect(err).NotTo(HaveOccurred(), "Delete PostgresRole")
				IsPostgresRoleDeleted(postgresRole.Name, postgresRole.Namespace)
			})

			It("Should be successful, Create DatabaseAccessRequest", func() {
				_, err := f.CSClient.EngineV1alpha1().DatabaseAccessRequests(postgresAKR.Namespace).Create(&postgresAKR)
				Expect(err).NotTo(HaveOccurred(), "Create DatabaseAccessRequest")

				IsDatabaseAccessRequestCreated(postgresAKR.Name, postgresAKR.Namespace)
			})

			It("Should be successful, Condition approved", func() {
				By("Creating DatabaseAccessRequest...")
				r, err := f.CSClient.EngineV1alpha1().DatabaseAccessRequests(postgresAKR.Namespace).Create(&postgresAKR)
				Expect(err).NotTo(HaveOccurred(), "Create DatabaseAccessRequest")

				IsDatabaseAccessRequestCreated(postgresAKR.Name, postgresAKR.Namespace)

				By("Updating Postgres AccessKeyRequest status...")
				err = f.UpdateDatabaseAccessRequestStatus(&api.DatabaseAccessRequestStatus{
					Conditions: []api.DatabaseAccessRequestCondition{
						{
							Type:           api.AccessApproved,
							LastUpdateTime: metav1.Now(),
						},
					},
				}, r)
				Expect(err).NotTo(HaveOccurred(), "Update conditions: Approved")
				IsPostgresAKRConditionApproved(postgresAKR.Name, postgresAKR.Namespace)
			})

			It("Should be successful, Condition denied", func() {
				By("Creating DatabaseAccessRequest...")
				r, err := f.CSClient.EngineV1alpha1().DatabaseAccessRequests(postgresAKR.Namespace).Create(&postgresAKR)
				Expect(err).NotTo(HaveOccurred(), "Create DatabaseAccessRequest")

				IsDatabaseAccessRequestCreated(postgresAKR.Name, postgresAKR.Namespace)

				By("Updating Postgres AccessKeyRequest status...")
				err = f.UpdateDatabaseAccessRequestStatus(&api.DatabaseAccessRequestStatus{
					Conditions: []api.DatabaseAccessRequestCondition{
						{
							Type:           api.AccessDenied,
							LastUpdateTime: metav1.Now(),
						},
					},
				}, r)
				Expect(err).NotTo(HaveOccurred(), "Update conditions: Denied")

				IsPostgresAKRConditionDenied(postgresAKR.Name, postgresAKR.Namespace)
			})
		})

		Context("Create database access secret", func() {
			var (
				secretName string
			)

			BeforeEach(func() {

				By("Creating PostgresRole...")
				r, err := f.CSClient.EngineV1alpha1().PostgresRoles(postgresRole.Namespace).Create(&postgresRole)
				Expect(err).NotTo(HaveOccurred(), "Create PostgresRole")

				IsPostgresRoleSucceeded(r.Name, r.Namespace)

			})

			AfterEach(func() {
				By("Deleting Postgres accesskeyrequest...")
				err := f.CSClient.EngineV1alpha1().DatabaseAccessRequests(postgresAKR.Namespace).Delete(postgresAKR.Name, &metav1.DeleteOptions{})
				Expect(err).NotTo(HaveOccurred(), "Delete DatabaseAccessRequest")

				IsDatabaseAccessRequestDeleted(postgresAKR.Name, postgresAKR.Namespace)
				IsPostgresAccessKeySecretDeleted(secretName, postgresAKR.Namespace)

				By("Deleting PostgresRole...")
				err = f.CSClient.EngineV1alpha1().PostgresRoles(postgresRole.Namespace).Delete(postgresRole.Name, &metav1.DeleteOptions{})
				Expect(err).NotTo(HaveOccurred(), "Delete PostgresRole")

				IsPostgresRoleDeleted(postgresRole.Name, postgresRole.Namespace)
			})

			It("Should be successful, Create Access Key Secret", func() {
				By("Creating Postgres accessKeyRequest...")
				r, err := f.CSClient.EngineV1alpha1().DatabaseAccessRequests(postgresAKR.Namespace).Create(&postgresAKR)
				Expect(err).NotTo(HaveOccurred(), "Create DatabaseAccessRequest")

				IsDatabaseAccessRequestCreated(postgresAKR.Name, postgresAKR.Namespace)

				By("Updating Postgres AccessKeyRequest status...")
				err = f.UpdateDatabaseAccessRequestStatus(&api.DatabaseAccessRequestStatus{
					Conditions: []api.DatabaseAccessRequestCondition{
						{
							Type:           api.AccessApproved,
							LastUpdateTime: metav1.Now(),
						},
					},
				}, r)

				Expect(err).NotTo(HaveOccurred(), "Update conditions: Approved")
				IsPostgresAKRConditionApproved(postgresAKR.Name, postgresAKR.Namespace)

				IsPostgresAccessKeySecretCreated(postgresAKR.Name, postgresAKR.Namespace)

				d, err := f.CSClient.EngineV1alpha1().DatabaseAccessRequests(postgresAKR.Namespace).Get(postgresAKR.Name, metav1.GetOptions{})
				Expect(err).NotTo(HaveOccurred(), "Get DatabaseAccessRequest")
				if d.Status.Secret != nil {
					secretName = d.Status.Secret.Name
				}
			})
		})

	})
})
