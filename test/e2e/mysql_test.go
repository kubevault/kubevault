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
	"kmodules.xyz/custom-resources/apis/appcatalog/v1alpha1"
)

var _ = Describe("MySQL Secret Engine", func() {

	var f *framework.Invocation

	var (
		IsSecretEngineCreated = func(name, namespace string) {
			By(fmt.Sprintf("Checking whether SecretEngine:(%s/%s) is created", namespace, name))
			Eventually(func() bool {
				_, err := f.CSClient.EngineV1alpha1().SecretEngines(namespace).Get(context.TODO(), name, metav1.GetOptions{})
				return err == nil
			}, timeOut, pollingInterval).Should(BeTrue(), "SecretEngine is created")
		}
		IsSecretEngineDeleted = func(name, namespace string) {
			By(fmt.Sprintf("Checking whether SecretEngine:(%s/%s) is deleted", namespace, name))
			Eventually(func() bool {
				_, err := f.CSClient.EngineV1alpha1().SecretEngines(namespace).Get(context.TODO(), name, metav1.GetOptions{})
				return kerrors.IsNotFound(err)
			}, timeOut, pollingInterval).Should(BeTrue(), "SecretEngine is deleted")
		}
		IsSecretEngineSucceeded = func(name, namespace string) {
			By(fmt.Sprintf("Checking whether SecretEngine:(%s/%s) is succeeded", namespace, name))
			Eventually(func() bool {
				r, err := f.CSClient.EngineV1alpha1().SecretEngines(namespace).Get(context.TODO(), name, metav1.GetOptions{})
				if err == nil {
					return r.Status.Phase == controller.SecretEnginePhaseSuccess
				}
				return false
			}, timeOut, pollingInterval).Should(BeTrue(), "SecretEngine status is succeeded")

		}
		IsMySQLRoleCreated = func(name, namespace string) {
			By(fmt.Sprintf("Checking whether MySQLRole:(%s/%s) role is created", namespace, name))
			Eventually(func() bool {
				_, err := f.CSClient.EngineV1alpha1().MySQLRoles(namespace).Get(context.TODO(), name, metav1.GetOptions{})
				return err == nil
			}, timeOut, pollingInterval).Should(BeTrue(), "MySQLRole is created")
		}
		IsMySQLRoleDeleted = func(name, namespace string) {
			By(fmt.Sprintf("Checking whether MySQLRole:(%s/%s) is deleted", namespace, name))
			Eventually(func() bool {
				_, err := f.CSClient.EngineV1alpha1().MySQLRoles(namespace).Get(context.TODO(), name, metav1.GetOptions{})
				return kerrors.IsNotFound(err)
			}, timeOut, pollingInterval).Should(BeTrue(), "MySQLRole is deleted")
		}
		IsMySQLRoleSucceeded = func(name, namespace string) {
			By(fmt.Sprintf("Checking whether MySQLRole:(%s/%s) is succeeded", namespace, name))
			Eventually(func() bool {
				r, err := f.CSClient.EngineV1alpha1().MySQLRoles(namespace).Get(context.TODO(), name, metav1.GetOptions{})
				if err == nil {
					return r.Status.Phase == controller.MySQLRolePhaseSuccess
				}
				return false
			}, timeOut, pollingInterval).Should(BeTrue(), "MySQLRole status is succeeded")

		}

		IsMySQLRoleFailed = func(name, namespace string) {
			By(fmt.Sprintf("Checking whether MySQLRole:(%s/%s) is failed", namespace, name))
			Eventually(func() bool {
				r, err := f.CSClient.EngineV1alpha1().MySQLRoles(namespace).Get(context.TODO(), name, metav1.GetOptions{})
				if err == nil {
					return r.Status.Phase != controller.MySQLRolePhaseSuccess && len(r.Status.Conditions) != 0
				}
				return false
			}, timeOut, pollingInterval).Should(BeTrue(), "MySQLRole status is failed")
		}
		IsDatabaseAccessRequestCreated = func(name, namespace string) {
			By(fmt.Sprintf("Checking whether DatabaseAccessRequest:(%s/%s) is created", namespace, name))
			Eventually(func() bool {
				_, err := f.CSClient.EngineV1alpha1().DatabaseAccessRequests(namespace).Get(context.TODO(), name, metav1.GetOptions{})
				return err == nil
			}, timeOut, pollingInterval).Should(BeTrue(), "DatabaseAccessRequest is created")
		}
		IsDatabaseAccessRequestDeleted = func(name, namespace string) {
			By(fmt.Sprintf("Checking whether DatabaseAccessRequest:(%s/%s) is deleted", namespace, name))
			Eventually(func() bool {
				_, err := f.CSClient.EngineV1alpha1().DatabaseAccessRequests(namespace).Get(context.TODO(), name, metav1.GetOptions{})
				return kerrors.IsNotFound(err)
			}, timeOut, pollingInterval).Should(BeTrue(), "DatabaseAccessRequest is deleted")
		}
		IsMySQLAKRConditionApproved = func(name, namespace string) {
			By("Checking whether DatabaseAccessRequestConditions-> Type: Approved")
			Eventually(func() bool {
				crd, err := f.CSClient.EngineV1alpha1().DatabaseAccessRequests(namespace).Get(context.TODO(), name, metav1.GetOptions{})
				if err == nil {
					for _, value := range crd.Status.Conditions {
						if value.Type == kmapi.ConditionRequestApproved && value.Status == kmapi.ConditionTrue {
							return true
						}
					}
				}
				return false
			}, timeOut, pollingInterval).Should(BeTrue(), "Conditions-> Type : Approved")
		}
		IsMySQLAKRConditionDenied = func(name, namespace string) {
			By("Checking whether DatabaseAccessRequestConditions-> Type: Denied")
			Eventually(func() bool {
				crd, err := f.CSClient.EngineV1alpha1().DatabaseAccessRequests(namespace).Get(context.TODO(), name, metav1.GetOptions{})
				if err == nil {
					for _, value := range crd.Status.Conditions {
						if value.Type == kmapi.ConditionRequestDenied && value.Status == kmapi.ConditionTrue {
							return true
						}
					}
				}
				return false
			}, timeOut, pollingInterval).Should(BeTrue(), "Conditions-> Type: Denied")
		}
		IsMySQLAccessKeySecretCreated = func(name, namespace string) {
			By("Checking whether MySQLAccessKeySecret is created")
			Eventually(func() bool {
				crd, err := f.CSClient.EngineV1alpha1().DatabaseAccessRequests(namespace).Get(context.TODO(), name, metav1.GetOptions{})
				if err == nil && crd.Status.Secret != nil {
					_, err2 := f.KubeClient.CoreV1().Secrets(namespace).Get(context.TODO(), crd.Status.Secret.Name, metav1.GetOptions{})
					return err2 == nil
				}
				return false
			}, timeOut, pollingInterval).Should(BeTrue(), "MySQLAccessKeySecret is created")
		}
		IsMySQLAccessKeySecretDeleted = func(secretName, namespace string) {
			By("Checking whether MySQLAccessKeySecret is deleted")
			Eventually(func() bool {
				_, err2 := f.KubeClient.CoreV1().Secrets(namespace).Get(context.TODO(), secretName, metav1.GetOptions{})
				return kerrors.IsNotFound(err2)
			}, timeOut, pollingInterval).Should(BeTrue(), "MySQLAccessKeySecret is deleted")
		}
	)

	BeforeEach(func() {
		f = root.Invoke()
		if !framework.SelfHostedOperator {
			Skip("Skipping MySQL secret engine tests because the operator isn't running inside cluster")
		}
	})

	AfterEach(func() {
		time.Sleep(20 * time.Second)
	})

	Describe("MySQLRole", func() {

		var (
			MySQLRole api.MySQLRole
			MySQLSE   api.SecretEngine
		)

		const (
			mysqlRoleName     = "my-mysql-roleset-4325"
			mysqlSecretEngine = "my-mysql-secretengine-3423423"
		)

		BeforeEach(func() {

			MySQLRole = api.MySQLRole{
				ObjectMeta: metav1.ObjectMeta{
					Name:      mysqlRoleName,
					Namespace: f.Namespace(),
				},
				Spec: api.MySQLRoleSpec{
					VaultRef: core.LocalObjectReference{
						Name: f.VaultAppRef.Name,
					},
					DatabaseRef: f.MysqlAppRef,

					CreationStatements: []string{
						"CREATE USER '{{name}}'@'%' IDENTIFIED BY '{{password}}';",
						"GRANT SELECT ON *.* TO '{{name}}'@'%';",
					},
					MaxTTL:     "1h",
					DefaultTTL: "300",
				},
			}

			MySQLSE = api.SecretEngine{
				ObjectMeta: metav1.ObjectMeta{
					Name:      mysqlSecretEngine,
					Namespace: f.Namespace(),
				},
				Spec: api.SecretEngineSpec{
					VaultRef: core.LocalObjectReference{
						Name: f.VaultAppRef.Name,
					},
					Path: "database",
					SecretEngineConfiguration: api.SecretEngineConfiguration{
						MySQL: &api.MySQLConfiguration{
							DatabaseRef: v1alpha1.AppReference{
								Name:      f.MysqlAppRef.Name,
								Namespace: f.MysqlAppRef.Namespace,
							},
						},
					},
				},
			}
		})

		Context("Create MySQLRole", func() {
			var p api.MySQLRole
			var se api.SecretEngine

			BeforeEach(func() {
				p = MySQLRole
				se = MySQLSE
			})

			AfterEach(func() {
				By("Deleting MySQLRole...")
				err := f.CSClient.EngineV1alpha1().MySQLRoles(MySQLRole.Namespace).Delete(context.TODO(), p.Name, metav1.DeleteOptions{})
				Expect(err).NotTo(HaveOccurred(), "Delete MySQLRole")

				IsMySQLRoleDeleted(p.Name, p.Namespace)

				By("Deleting SecretEngine...")
				err = f.CSClient.EngineV1alpha1().SecretEngines(se.Namespace).Delete(context.TODO(), se.Name, metav1.DeleteOptions{})
				Expect(err).NotTo(HaveOccurred(), "Delete Secret engine")

				IsSecretEngineDeleted(se.Name, se.Namespace)
			})

			It("Should be successful", func() {
				By("Creating SecretEngine...")
				_, err := f.CSClient.EngineV1alpha1().SecretEngines(se.Namespace).Create(context.TODO(), &se, metav1.CreateOptions{})
				Expect(err).NotTo(HaveOccurred(), "Create SecretEngine")

				IsSecretEngineCreated(se.Name, se.Namespace)
				IsSecretEngineSucceeded(se.Name, se.Namespace)

				By("Creating MySQLRole...")
				_, err = f.CSClient.EngineV1alpha1().MySQLRoles(p.Namespace).Create(context.TODO(), &p, metav1.CreateOptions{})
				Expect(err).NotTo(HaveOccurred(), "Create MySQLRole")

				IsMySQLRoleCreated(p.Name, p.Namespace)
				IsMySQLRoleSucceeded(p.Name, p.Namespace)
			})

		})

		Context("Create MySQLRole without enabling secretEngine", func() {
			var p api.MySQLRole

			BeforeEach(func() {
				p = MySQLRole
			})

			AfterEach(func() {
				By("Deleting MySQLRole...")
				err := f.CSClient.EngineV1alpha1().MySQLRoles(MySQLRole.Namespace).Delete(context.TODO(), p.Name, metav1.DeleteOptions{})
				Expect(err).NotTo(HaveOccurred(), "Delete MySQLRole")

				IsMySQLRoleDeleted(p.Name, p.Namespace)

			})

			It("Should be failed making MySQLRole", func() {

				By("Creating MySQLRole...")
				_, err := f.CSClient.EngineV1alpha1().MySQLRoles(p.Namespace).Create(context.TODO(), &p, metav1.CreateOptions{})
				Expect(err).NotTo(HaveOccurred(), "Create MySQLRole")

				IsMySQLRoleCreated(p.Name, p.Namespace)
				IsMySQLRoleFailed(p.Name, p.Namespace)
			})
		})

	})

	Describe("DatabaseAccessRequest", func() {

		var (
			MySQLRole api.MySQLRole
			MySQLSE   api.SecretEngine
			MySQLAKR  api.DatabaseAccessRequest
		)

		const (
			mysqlRoleName     = "my-mysql-roleset-4325"
			mysqlSecretEngine = "my-mysql-secretengine-3423423"
			mysqlAKRName      = "my-mysql-token-2345"
		)

		BeforeEach(func() {

			MySQLSE = api.SecretEngine{
				ObjectMeta: metav1.ObjectMeta{
					Name:      mysqlSecretEngine,
					Namespace: f.Namespace(),
				},
				Spec: api.SecretEngineSpec{
					VaultRef: core.LocalObjectReference{
						Name: f.VaultAppRef.Name,
					},
					Path: "database",
					SecretEngineConfiguration: api.SecretEngineConfiguration{
						MySQL: &api.MySQLConfiguration{
							PluginName: "mysql-rds-database-plugin",
							DatabaseRef: v1alpha1.AppReference{
								Name:      f.MysqlAppRef.Name,
								Namespace: f.MysqlAppRef.Namespace,
							},
						},
					},
				},
			}
			_, err := f.CSClient.EngineV1alpha1().SecretEngines(MySQLSE.Namespace).Create(context.TODO(), &MySQLSE, metav1.CreateOptions{})
			Expect(err).NotTo(HaveOccurred(), "Create MySQL SecretEngine")
			IsSecretEngineCreated(MySQLSE.Name, MySQLSE.Namespace)

			MySQLRole = api.MySQLRole{
				ObjectMeta: metav1.ObjectMeta{
					Name:      mysqlRoleName,
					Namespace: f.Namespace(),
				},
				Spec: api.MySQLRoleSpec{
					VaultRef: core.LocalObjectReference{
						Name: f.VaultAppRef.Name,
					},
					DatabaseRef: f.MysqlAppRef,

					CreationStatements: []string{
						"CREATE USER '{{name}}'@'%' IDENTIFIED BY '{{password}}';",
						"GRANT SELECT ON *.* TO '{{name}}'@'%';",
					},
					MaxTTL:     "1h",
					DefaultTTL: "300",
				},
			}

			MySQLAKR = api.DatabaseAccessRequest{
				ObjectMeta: metav1.ObjectMeta{
					Name:      mysqlAKRName,
					Namespace: f.Namespace(),
				},
				Spec: api.DatabaseAccessRequestSpec{
					RoleRef: api.RoleRef{
						Kind:      api.ResourceKindMySQLRole,
						Name:      mysqlRoleName,
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
			err := f.CSClient.EngineV1alpha1().SecretEngines(MySQLSE.Namespace).Delete(context.TODO(), MySQLSE.Name, metav1.DeleteOptions{})
			Expect(err).NotTo(HaveOccurred(), "Delete MySQL SecretEngine")
			IsSecretEngineDeleted(MySQLSE.Name, MySQLSE.Namespace)
		})

		Context("Create, Approve, Deny DatabaseAccessRequests", func() {
			BeforeEach(func() {
				_, err := f.CSClient.EngineV1alpha1().MySQLRoles(MySQLRole.Namespace).Create(context.TODO(), &MySQLRole, metav1.CreateOptions{})
				Expect(err).NotTo(HaveOccurred(), "Create MySQLRole")

				IsMySQLRoleCreated(MySQLRole.Name, MySQLRole.Namespace)
				IsMySQLRoleSucceeded(MySQLRole.Name, MySQLRole.Namespace)

			})

			AfterEach(func() {
				err := f.CSClient.EngineV1alpha1().DatabaseAccessRequests(MySQLAKR.Namespace).Delete(context.TODO(), MySQLAKR.Name, metav1.DeleteOptions{})
				Expect(err).NotTo(HaveOccurred(), "Delete DatabaseAccessRequest")
				IsDatabaseAccessRequestDeleted(MySQLAKR.Name, MySQLAKR.Namespace)

				err = f.CSClient.EngineV1alpha1().MySQLRoles(MySQLRole.Namespace).Delete(context.TODO(), MySQLRole.Name, metav1.DeleteOptions{})
				Expect(err).NotTo(HaveOccurred(), "Delete MySQLRole")
				IsMySQLRoleDeleted(MySQLRole.Name, MySQLRole.Namespace)
			})

			It("Should be successful, Create DatabaseAccessRequest", func() {
				_, err := f.CSClient.EngineV1alpha1().DatabaseAccessRequests(MySQLAKR.Namespace).Create(context.TODO(), &MySQLAKR, metav1.CreateOptions{})
				Expect(err).NotTo(HaveOccurred(), "Create DatabaseAccessRequest")

				IsDatabaseAccessRequestCreated(MySQLAKR.Name, MySQLAKR.Namespace)
			})

			It("Should be successful, Condition approved", func() {
				By("Creating DatabaseAccessRequest...")
				r, err := f.CSClient.EngineV1alpha1().DatabaseAccessRequests(MySQLAKR.Namespace).Create(context.TODO(), &MySQLAKR, metav1.CreateOptions{})
				Expect(err).NotTo(HaveOccurred(), "Create DatabaseAccessRequest")

				IsDatabaseAccessRequestCreated(MySQLAKR.Name, MySQLAKR.Namespace)

				By("Updating MySQL AccessKeyRequest status...")
				err = f.UpdateDatabaseAccessRequestStatus(&api.DatabaseAccessRequestStatus{
					Conditions: []kmapi.Condition{
						{
							Type:               kmapi.ConditionRequestApproved,
							Status:             kmapi.ConditionTrue,
							LastTransitionTime: metav1.Now(),
						},
					},
					Phase: api.RequestStatusPhaseApproved,
				}, r)
				Expect(err).NotTo(HaveOccurred(), "Update conditions: Approved")
				IsMySQLAKRConditionApproved(MySQLAKR.Name, MySQLAKR.Namespace)
			})

			It("Should be successful, Condition denied", func() {
				By("Creating DatabaseAccessRequest...")
				r, err := f.CSClient.EngineV1alpha1().DatabaseAccessRequests(MySQLAKR.Namespace).Create(context.TODO(), &MySQLAKR, metav1.CreateOptions{})
				Expect(err).NotTo(HaveOccurred(), "Create DatabaseAccessRequest")

				IsDatabaseAccessRequestCreated(MySQLAKR.Name, MySQLAKR.Namespace)

				By("Updating MySQL AccessKeyRequest status...")
				err = f.UpdateDatabaseAccessRequestStatus(&api.DatabaseAccessRequestStatus{
					Conditions: []kmapi.Condition{
						{
							Type:               kmapi.ConditionRequestDenied,
							Status:             kmapi.ConditionTrue,
							LastTransitionTime: metav1.Now(),
						},
					},
					Phase: api.RequestStatusPhaseDenied,
				}, r)
				Expect(err).NotTo(HaveOccurred(), "Update conditions: Denied")

				IsMySQLAKRConditionDenied(MySQLAKR.Name, MySQLAKR.Namespace)
			})
		})

		Context("Create database access secret", func() {
			var (
				secretName string
			)

			BeforeEach(func() {

				By("Creating MySQLRole...")
				r, err := f.CSClient.EngineV1alpha1().MySQLRoles(MySQLRole.Namespace).Create(context.TODO(), &MySQLRole, metav1.CreateOptions{})
				Expect(err).NotTo(HaveOccurred(), "Create MySQLRole")

				IsMySQLRoleSucceeded(r.Name, r.Namespace)

			})

			AfterEach(func() {
				By("Deleting MySQL accesskeyrequest...")
				err := f.CSClient.EngineV1alpha1().DatabaseAccessRequests(MySQLAKR.Namespace).Delete(context.TODO(), MySQLAKR.Name, metav1.DeleteOptions{})
				Expect(err).NotTo(HaveOccurred(), "Delete DatabaseAccessRequest")

				IsDatabaseAccessRequestDeleted(MySQLAKR.Name, MySQLAKR.Namespace)
				IsMySQLAccessKeySecretDeleted(secretName, MySQLAKR.Namespace)

				By("Deleting MySQLRole...")
				err = f.CSClient.EngineV1alpha1().MySQLRoles(MySQLRole.Namespace).Delete(context.TODO(), MySQLRole.Name, metav1.DeleteOptions{})
				Expect(err).NotTo(HaveOccurred(), "Delete MySQLRole")

				IsMySQLRoleDeleted(MySQLRole.Name, MySQLRole.Namespace)
			})

			It("Should be successful, Create Access Key Secret", func() {
				By("Creating MySQL accessKeyRequest...")
				r, err := f.CSClient.EngineV1alpha1().DatabaseAccessRequests(MySQLAKR.Namespace).Create(context.TODO(), &MySQLAKR, metav1.CreateOptions{})
				Expect(err).NotTo(HaveOccurred(), "Create DatabaseAccessRequest")

				IsDatabaseAccessRequestCreated(MySQLAKR.Name, MySQLAKR.Namespace)

				By("Updating MySQL AccessKeyRequest status...")
				err = f.UpdateDatabaseAccessRequestStatus(&api.DatabaseAccessRequestStatus{
					Conditions: []kmapi.Condition{
						{
							Type:               kmapi.ConditionRequestApproved,
							Status:             kmapi.ConditionTrue,
							LastTransitionTime: metav1.Now(),
						},
					},
					Phase: api.RequestStatusPhaseApproved,
				}, r)

				Expect(err).NotTo(HaveOccurred(), "Update conditions: Approved")
				IsMySQLAKRConditionApproved(MySQLAKR.Name, MySQLAKR.Namespace)

				IsMySQLAccessKeySecretCreated(MySQLAKR.Name, MySQLAKR.Namespace)

				d, err := f.CSClient.EngineV1alpha1().DatabaseAccessRequests(MySQLAKR.Namespace).Get(context.TODO(), MySQLAKR.Name, metav1.GetOptions{})
				Expect(err).NotTo(HaveOccurred(), "Get DatabaseAccessRequest")
				if d.Status.Secret != nil {
					secretName = d.Status.Secret.Name
				}
			})
		})

	})
})
