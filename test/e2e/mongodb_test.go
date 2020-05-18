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
	kmapi "kmodules.xyz/client-go/api/v1"
	"kmodules.xyz/custom-resources/apis/appcatalog/v1alpha1"
)

var _ = Describe("MongoDB Secret Engine", func() {

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
		IsMongoDBRoleCreated = func(name, namespace string) {
			By(fmt.Sprintf("Checking whether MongoDBRole:(%s/%s) role is created", namespace, name))
			Eventually(func() bool {
				_, err := f.CSClient.EngineV1alpha1().MongoDBRoles(namespace).Get(name, metav1.GetOptions{})
				return err == nil
			}, timeOut, pollingInterval).Should(BeTrue(), "MongoDBRole is created")
		}
		IsMongoDBRoleDeleted = func(name, namespace string) {
			By(fmt.Sprintf("Checking whether MongoDBRole:(%s/%s) is deleted", namespace, name))
			Eventually(func() bool {
				_, err := f.CSClient.EngineV1alpha1().MongoDBRoles(namespace).Get(name, metav1.GetOptions{})
				return kerrors.IsNotFound(err)
			}, timeOut, pollingInterval).Should(BeTrue(), "MongoDBRole is deleted")
		}
		IsMongoDBRoleSucceeded = func(name, namespace string) {
			By(fmt.Sprintf("Checking whether MongoDBRole:(%s/%s) is succeeded", namespace, name))
			Eventually(func() bool {
				r, err := f.CSClient.EngineV1alpha1().MongoDBRoles(namespace).Get(name, metav1.GetOptions{})
				if err == nil {
					return r.Status.Phase == controller.MongoDBRolePhaseSuccess
				}
				return false
			}, timeOut, pollingInterval).Should(BeTrue(), "MongoDBRole status is succeeded")

		}

		IsMongoDBRoleFailed = func(name, namespace string) {
			By(fmt.Sprintf("Checking whether MongoDBRole:(%s/%s) is failed", namespace, name))
			Eventually(func() bool {
				r, err := f.CSClient.EngineV1alpha1().MongoDBRoles(namespace).Get(name, metav1.GetOptions{})
				if err == nil {
					return r.Status.Phase != controller.MongoDBRolePhaseSuccess && len(r.Status.Conditions) != 0
				}
				return false
			}, timeOut, pollingInterval).Should(BeTrue(), "MongoDBRole status is failed")
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
		IsMongoDBAKRConditionApproved = func(name, namespace string) {
			By(fmt.Sprintf("Checking whether DatabaseAccessRequestConditions-> Type: Approved"))
			Eventually(func() bool {
				crd, err := f.CSClient.EngineV1alpha1().DatabaseAccessRequests(namespace).Get(name, metav1.GetOptions{})
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
		IsMongoDBAKRConditionDenied = func(name, namespace string) {
			By(fmt.Sprintf("Checking whether DatabaseAccessRequestConditions-> Type: Denied"))
			Eventually(func() bool {
				crd, err := f.CSClient.EngineV1alpha1().DatabaseAccessRequests(namespace).Get(name, metav1.GetOptions{})
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
		IsMongoDBAccessKeySecretCreated = func(name, namespace string) {
			By("Checking whether MongoDBAccessKeySecret is created")
			Eventually(func() bool {
				crd, err := f.CSClient.EngineV1alpha1().DatabaseAccessRequests(namespace).Get(name, metav1.GetOptions{})
				if err == nil && crd.Status.Secret != nil {
					_, err2 := f.KubeClient.CoreV1().Secrets(namespace).Get(crd.Status.Secret.Name, metav1.GetOptions{})
					return err2 == nil
				}
				return false
			}, timeOut, pollingInterval).Should(BeTrue(), "MongoDBAccessKeySecret is created")
		}
		IsMongoDBAccessKeySecretDeleted = func(secretName, namespace string) {
			By("Checking whether MongoDBAccessKeySecret is deleted")
			Eventually(func() bool {
				_, err2 := f.KubeClient.CoreV1().Secrets(namespace).Get(secretName, metav1.GetOptions{})
				return kerrors.IsNotFound(err2)
			}, timeOut, pollingInterval).Should(BeTrue(), "MongoDBAccessKeySecret is deleted")
		}
	)

	BeforeEach(func() {
		f = root.Invoke()
		if !framework.SelfHostedOperator {
			Skip("Skipping MongoDB secret engine tests because the operator isn't running inside cluster")
		}
	})

	AfterEach(func() {
		time.Sleep(20 * time.Second)
	})

	Describe("MongoDBRole", func() {

		var (
			mongoDBRole api.MongoDBRole
			mongoDBSE   api.SecretEngine
		)

		const (
			mongoDBRoleName     = "my-mongo-role-4325"
			mongoDBSecretEngine = "my-mongo-secretengine-3423423"
		)

		BeforeEach(func() {

			mongoDBRole = api.MongoDBRole{
				ObjectMeta: metav1.ObjectMeta{
					Name:      mongoDBRoleName,
					Namespace: f.Namespace(),
				},
				Spec: api.MongoDBRoleSpec{
					VaultRef: core.LocalObjectReference{
						Name: f.VaultAppRef.Name,
					},
					DatabaseRef: f.MongoAppRef,
					CreationStatements: []string{
						"{ \"db\": \"admin\", \"roles\": [{ \"role\": \"readWrite\" }, {\"role\": \"read\", \"db\": \"foo\"}] }",
					},
					MaxTTL:     "1h",
					DefaultTTL: "300",
				},
			}

			mongoDBSE = api.SecretEngine{
				ObjectMeta: metav1.ObjectMeta{
					Name:      mongoDBSecretEngine,
					Namespace: f.Namespace(),
				},
				Spec: api.SecretEngineSpec{
					VaultRef: core.LocalObjectReference{
						Name: f.VaultAppRef.Name,
					},
					Path: "database",
					SecretEngineConfiguration: api.SecretEngineConfiguration{
						MongoDB: &api.MongoDBConfiguration{
							DatabaseRef: v1alpha1.AppReference{
								Name:      f.MongoAppRef.Name,
								Namespace: f.MongoAppRef.Namespace,
							},
						},
					},
				},
			}
		})

		Context("Create MongoDBRole", func() {
			var p api.MongoDBRole
			var se api.SecretEngine

			BeforeEach(func() {
				p = mongoDBRole
				se = mongoDBSE
			})

			AfterEach(func() {
				By("Deleting MongoDBRole...")
				err := f.CSClient.EngineV1alpha1().MongoDBRoles(mongoDBRole.Namespace).Delete(p.Name, &metav1.DeleteOptions{})
				Expect(err).NotTo(HaveOccurred(), "Delete MongoDBRole")

				IsMongoDBRoleDeleted(p.Name, p.Namespace)

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

				By("Creating MongoDBRole...")
				_, err = f.CSClient.EngineV1alpha1().MongoDBRoles(p.Namespace).Create(&p)
				Expect(err).NotTo(HaveOccurred(), "Create MongoDBRole")

				IsMongoDBRoleCreated(p.Name, p.Namespace)
				IsMongoDBRoleSucceeded(p.Name, p.Namespace)
			})

		})

		Context("Create MongoDBRole without enabling secretEngine", func() {
			var p api.MongoDBRole

			BeforeEach(func() {
				p = mongoDBRole
			})

			AfterEach(func() {
				By("Deleting MongoDBRole...")
				err := f.CSClient.EngineV1alpha1().MongoDBRoles(mongoDBRole.Namespace).Delete(p.Name, &metav1.DeleteOptions{})
				Expect(err).NotTo(HaveOccurred(), "Delete MongoDBRole")

				IsMongoDBRoleDeleted(p.Name, p.Namespace)

			})

			It("Should be failed making MongoDBRole", func() {

				By("Creating MongoDBRole...")
				_, err := f.CSClient.EngineV1alpha1().MongoDBRoles(p.Namespace).Create(&p)
				Expect(err).NotTo(HaveOccurred(), "Create MongoDBRole")

				IsMongoDBRoleCreated(p.Name, p.Namespace)
				IsMongoDBRoleFailed(p.Name, p.Namespace)
			})
		})

	})

	Describe("DatabaseAccessRequest", func() {

		var (
			mongoDBRole api.MongoDBRole
			mongoDBSE   api.SecretEngine
			mongoDBAKR  api.DatabaseAccessRequest
		)

		const (
			mongoDBRoleName     = "my-mongo-role-4325"
			mongoDBSecretEngine = "my-mongo-secretengine-3423423"
			mongoDBAKRName      = "my-mongo-token-2345"
		)

		BeforeEach(func() {

			mongoDBSE = api.SecretEngine{
				ObjectMeta: metav1.ObjectMeta{
					Name:      mongoDBSecretEngine,
					Namespace: f.Namespace(),
				},
				Spec: api.SecretEngineSpec{
					VaultRef: core.LocalObjectReference{
						Name: f.VaultAppRef.Name,
					},
					Path: "database",
					SecretEngineConfiguration: api.SecretEngineConfiguration{
						MongoDB: &api.MongoDBConfiguration{
							DatabaseRef: v1alpha1.AppReference{
								Name:      f.MongoAppRef.Name,
								Namespace: f.MongoAppRef.Namespace,
							},
						},
					},
				},
			}
			_, err := f.CSClient.EngineV1alpha1().SecretEngines(mongoDBSE.Namespace).Create(&mongoDBSE)
			Expect(err).NotTo(HaveOccurred(), "Create mongoDB SecretEngine")
			IsSecretEngineCreated(mongoDBSE.Name, mongoDBSE.Namespace)

			mongoDBRole = api.MongoDBRole{
				ObjectMeta: metav1.ObjectMeta{
					Name:      mongoDBRoleName,
					Namespace: f.Namespace(),
				},
				Spec: api.MongoDBRoleSpec{
					VaultRef: core.LocalObjectReference{
						Name: f.VaultAppRef.Name,
					},
					DatabaseRef: f.MongoAppRef,
					CreationStatements: []string{
						"{ \"db\": \"admin\", \"roles\": [{ \"role\": \"readWrite\" }, {\"role\": \"read\", \"db\": \"foo\"}] }",
					},
					MaxTTL:     "1h",
					DefaultTTL: "300",
				},
			}

			mongoDBAKR = api.DatabaseAccessRequest{
				ObjectMeta: metav1.ObjectMeta{
					Name:      mongoDBAKRName,
					Namespace: f.Namespace(),
				},
				Spec: api.DatabaseAccessRequestSpec{
					RoleRef: api.RoleRef{
						Kind:      api.ResourceKindMongoDBRole,
						Name:      mongoDBRoleName,
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
			err := f.CSClient.EngineV1alpha1().SecretEngines(mongoDBSE.Namespace).Delete(mongoDBSE.Name, &metav1.DeleteOptions{})
			Expect(err).NotTo(HaveOccurred(), "Delete mongoDB SecretEngine")
			IsSecretEngineDeleted(mongoDBSE.Name, mongoDBSE.Namespace)
		})

		Context("Create, Approve, Deny DatabaseAccessRequests", func() {
			BeforeEach(func() {
				_, err := f.CSClient.EngineV1alpha1().MongoDBRoles(mongoDBRole.Namespace).Create(&mongoDBRole)
				Expect(err).NotTo(HaveOccurred(), "Create MongoDBRole")

				IsMongoDBRoleCreated(mongoDBRole.Name, mongoDBRole.Namespace)
				IsMongoDBRoleSucceeded(mongoDBRole.Name, mongoDBRole.Namespace)

			})

			AfterEach(func() {
				err := f.CSClient.EngineV1alpha1().DatabaseAccessRequests(mongoDBAKR.Namespace).Delete(mongoDBAKR.Name, &metav1.DeleteOptions{})
				Expect(err).NotTo(HaveOccurred(), "Delete DatabaseAccessRequest")
				IsDatabaseAccessRequestDeleted(mongoDBAKR.Name, mongoDBAKR.Namespace)

				err = f.CSClient.EngineV1alpha1().MongoDBRoles(mongoDBRole.Namespace).Delete(mongoDBRole.Name, &metav1.DeleteOptions{})
				Expect(err).NotTo(HaveOccurred(), "Delete MongoDBRole")
				IsMongoDBRoleDeleted(mongoDBRole.Name, mongoDBRole.Namespace)
			})

			It("Should be successful, Create DatabaseAccessRequest", func() {
				_, err := f.CSClient.EngineV1alpha1().DatabaseAccessRequests(mongoDBAKR.Namespace).Create(&mongoDBAKR)
				Expect(err).NotTo(HaveOccurred(), "Create DatabaseAccessRequest")

				IsDatabaseAccessRequestCreated(mongoDBAKR.Name, mongoDBAKR.Namespace)
			})

			It("Should be successful, Condition approved", func() {
				By("Creating DatabaseAccessRequest...")
				r, err := f.CSClient.EngineV1alpha1().DatabaseAccessRequests(mongoDBAKR.Namespace).Create(&mongoDBAKR)
				Expect(err).NotTo(HaveOccurred(), "Create DatabaseAccessRequest")

				IsDatabaseAccessRequestCreated(mongoDBAKR.Name, mongoDBAKR.Namespace)

				By("Updating MongoDB AccessKeyRequest status...")
				err = f.UpdateDatabaseAccessRequestStatus(&api.DatabaseAccessRequestStatus{
					Conditions: []kmapi.Condition{
						{
							Type:               kmapi.ConditionRequestApproved,
							LastTransitionTime: metav1.Now(),
						},
					},
				}, r)
				Expect(err).NotTo(HaveOccurred(), "Update conditions: Approved")
				IsMongoDBAKRConditionApproved(mongoDBAKR.Name, mongoDBAKR.Namespace)
			})

			It("Should be successful, Condition denied", func() {
				By("Creating DatabaseAccessRequest...")
				r, err := f.CSClient.EngineV1alpha1().DatabaseAccessRequests(mongoDBAKR.Namespace).Create(&mongoDBAKR)
				Expect(err).NotTo(HaveOccurred(), "Create DatabaseAccessRequest")

				IsDatabaseAccessRequestCreated(mongoDBAKR.Name, mongoDBAKR.Namespace)

				By("Updating MongoDB AccessKeyRequest status...")
				err = f.UpdateDatabaseAccessRequestStatus(&api.DatabaseAccessRequestStatus{
					Conditions: []kmapi.Condition{
						{
							Type:               kmapi.ConditionRequestDenied,
							LastTransitionTime: metav1.Now(),
						},
					},
				}, r)
				Expect(err).NotTo(HaveOccurred(), "Update conditions: Denied")

				IsMongoDBAKRConditionDenied(mongoDBAKR.Name, mongoDBAKR.Namespace)
			})
		})

		Context("Create database access secret", func() {
			var (
				secretName string
			)

			BeforeEach(func() {

				By("Creating MongoDBRole...")
				r, err := f.CSClient.EngineV1alpha1().MongoDBRoles(mongoDBRole.Namespace).Create(&mongoDBRole)
				Expect(err).NotTo(HaveOccurred(), "Create MongoDBRole")

				IsMongoDBRoleSucceeded(r.Name, r.Namespace)

			})

			AfterEach(func() {
				By("Deleting MongoDB accesskeyrequest...")
				err := f.CSClient.EngineV1alpha1().DatabaseAccessRequests(mongoDBAKR.Namespace).Delete(mongoDBAKR.Name, &metav1.DeleteOptions{})
				Expect(err).NotTo(HaveOccurred(), "Delete DatabaseAccessRequest")

				IsDatabaseAccessRequestDeleted(mongoDBAKR.Name, mongoDBAKR.Namespace)
				IsMongoDBAccessKeySecretDeleted(secretName, mongoDBAKR.Namespace)

				By("Deleting MongoDBRole...")
				err = f.CSClient.EngineV1alpha1().MongoDBRoles(mongoDBRole.Namespace).Delete(mongoDBRole.Name, &metav1.DeleteOptions{})
				Expect(err).NotTo(HaveOccurred(), "Delete MongoDBRole")

				IsMongoDBRoleDeleted(mongoDBRole.Name, mongoDBRole.Namespace)
			})

			It("Should be successful, Create Access Key Secret", func() {
				By("Creating MongoDB accessKeyRequest...")
				r, err := f.CSClient.EngineV1alpha1().DatabaseAccessRequests(mongoDBAKR.Namespace).Create(&mongoDBAKR)
				Expect(err).NotTo(HaveOccurred(), "Create DatabaseAccessRequest")

				IsDatabaseAccessRequestCreated(mongoDBAKR.Name, mongoDBAKR.Namespace)

				By("Updating MongoDB AccessKeyRequest status...")
				err = f.UpdateDatabaseAccessRequestStatus(&api.DatabaseAccessRequestStatus{
					Conditions: []kmapi.Condition{
						{
							Type:               kmapi.ConditionRequestApproved,
							LastTransitionTime: metav1.Now(),
						},
					},
				}, r)

				Expect(err).NotTo(HaveOccurred(), "Update conditions: Approved")
				IsMongoDBAKRConditionApproved(mongoDBAKR.Name, mongoDBAKR.Namespace)

				IsMongoDBAccessKeySecretCreated(mongoDBAKR.Name, mongoDBAKR.Namespace)

				d, err := f.CSClient.EngineV1alpha1().DatabaseAccessRequests(mongoDBAKR.Namespace).Get(mongoDBAKR.Name, metav1.GetOptions{})
				Expect(err).NotTo(HaveOccurred(), "Get DatabaseAccessRequest")
				if d.Status.Secret != nil {
					secretName = d.Status.Secret.Name
				}
			})
		})

	})
})
