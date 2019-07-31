package e2e_test

import (
	"fmt"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	core "k8s.io/api/core/v1"
	rbac "k8s.io/api/rbac/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	appcat "kmodules.xyz/custom-resources/apis/appcatalog/v1alpha1"
	api "kubevault.dev/operator/apis/engine/v1alpha1"
	"kubevault.dev/operator/pkg/vault"
	"kubevault.dev/operator/test/e2e/framework"
)

var _ = Describe("Mongodb role", func() {

	var f *framework.Invocation

	BeforeEach(func() {
		f = root.Invoke()

	})

	AfterEach(func() {
		time.Sleep(20 * time.Second)
	})

	var (
		// vault related
		IsVaultDatabaseConfigCreated = func(name string) {
			By(fmt.Sprintf("Checking Is vault database config created"))
			cl, err := vault.NewClient(f.KubeClient, f.AppcatClient, f.VaultAppRef)
			Expect(err).NotTo(HaveOccurred(), "Get vault client")

			req := cl.NewRequest("GET", fmt.Sprintf("/v1/database/config/%s", name))
			Eventually(func() bool {
				_, err := cl.RawRequest(req)
				return err == nil
			}, timeOut, pollingInterval).Should(BeTrue(), "Is vault database config created")
		}

		IsVaultDatabaseRoleCreated = func(name string) {
			By(fmt.Sprintf("Checking Is vault database role created"))
			cl, err := vault.NewClient(f.KubeClient, f.AppcatClient, f.VaultAppRef)
			Expect(err).NotTo(HaveOccurred(), "Get vault client")

			req := cl.NewRequest("GET", fmt.Sprintf("/v1/database/roles/%s", name))
			Eventually(func() bool {
				_, err := cl.RawRequest(req)
				return err == nil
			}, timeOut, pollingInterval).Should(BeTrue(), "Is vault database role created")
		}

		IsVaultDatabaseRoleDeleted = func(name string) {
			By(fmt.Sprintf("Checking Is vault database role deleted"))
			cl, err := vault.NewClient(f.KubeClient, f.AppcatClient, f.VaultAppRef)
			Expect(err).NotTo(HaveOccurred(), "Get vault client")

			req := cl.NewRequest("GET", fmt.Sprintf("/v1/database/roles/%s", name))
			Eventually(func() bool {
				_, err := cl.RawRequest(req)
				return err != nil
			}, timeOut, pollingInterval).Should(BeTrue(), "Is vault database role deleted")
		}

		IsMongoDBRoleCreated = func(name, namespace string) {
			By(fmt.Sprintf("Checking Is MongoDBRole(%s/%s) created", namespace, name))
			Eventually(func() bool {
				_, err := f.DBClient.AuthorizationV1alpha1().MongoDBRoles(namespace).Get(name, metav1.GetOptions{})
				return err == nil
			}, timeOut, pollingInterval).Should(BeTrue(), "Is Mongodb role created")
		}

		IsMongoDBRoleDeleted = func(name, namespace string) {
			By(fmt.Sprintf("Checking Is MongoDBRole(%s/%s) deleted", namespace, name))
			Eventually(func() bool {
				_, err := f.DBClient.AuthorizationV1alpha1().MongoDBRoles(namespace).Get(name, metav1.GetOptions{})
				return kerrors.IsNotFound(err)
			}, timeOut, pollingInterval).Should(BeTrue(), "Is MongoDBRole role deleted")
		}

		IsDatabaseAccessRequestCreated = func(name, namespace string) {
			By(fmt.Sprintf("Checking Is DatabaseAccessRequest(%s/%s) created", namespace, name))
			Eventually(func() bool {
				_, err := f.DBClient.AuthorizationV1alpha1().DatabaseAccessRequests(namespace).Get(name, metav1.GetOptions{})
				return err == nil
			}, timeOut, pollingInterval).Should(BeTrue(), "Is DatabaseAccessRequest created")
		}

		IsDatabaseAccessRequestDeleted = func(name, namespace string) {
			By(fmt.Sprintf("Checking Is DatabaseAccessRequest(%s/%s) deleted", namespace, name))
			Eventually(func() bool {
				_, err := f.DBClient.AuthorizationV1alpha1().DatabaseAccessRequests(namespace).Get(name, metav1.GetOptions{})
				return kerrors.IsNotFound(err)
			}, timeOut, pollingInterval).Should(BeTrue(), "Is DatabaseAccessRequest deleted")
		}

		IsDatabaseAccessRequestApproved = func(name, namespace string) {
			By(fmt.Sprintf("Checking Is DatabaseAccessRequest(%s/%s) apporved", namespace, name))
			Eventually(func() bool {
				d, err := f.DBClient.AuthorizationV1alpha1().DatabaseAccessRequests(namespace).Get(name, metav1.GetOptions{})
				if err == nil {
					return d.Status.Lease != nil
				}
				return false
			}, timeOut, pollingInterval).Should(BeTrue(), "Is DatabaseAccessRequest approved")
		}
		IsDatabaseAccessRequestDenied = func(name, namespace string) {
			By(fmt.Sprintf("Checking Is DatabaseAccessRequest(%s/%s) denied", namespace, name))
			Eventually(func() bool {
				d, err := f.DBClient.AuthorizationV1alpha1().DatabaseAccessRequests(namespace).Get(name, metav1.GetOptions{})
				if err == nil {
					for _, c := range d.Status.Conditions {
						if c.Type == api.AccessDenied {
							return true
						}
					}
				}
				return false
			}, timeOut, pollingInterval).Should(BeTrue(), "Is DatabaseAccessRequest denied")
		}
	)

	Describe("MongoDBRole", func() {
		var (
			mRole api.MongoDBRole
		)

		BeforeEach(func() {
			mRole = api.MongoDBRole{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "m-role-test1",
					Namespace: f.Namespace(),
				},
				Spec: api.MongoDBRoleSpec{
					AuthManagerRef: f.VaultAppRef,
					DatabaseRef: &core.LocalObjectReference{
						Name: f.MongoAppRef.Name,
					},
					CreationStatements: []string{
						"{ \"db\": \"admin\", \"roles\": [{ \"role\": \"readWrite\" }, {\"role\": \"read\", \"db\": \"foo\"}] }",
					},
					MaxTTL:     "1h",
					DefaultTTL: "300",
				},
			}
		})

		Context("Create MongoDBRole", func() {
			var p api.MongoDBRole

			BeforeEach(func() {
				p = mRole
			})

			AfterEach(func() {
				err := f.DBClient.AuthorizationV1alpha1().MongoDBRoles(p.Namespace).Delete(p.Name, &metav1.DeleteOptions{})
				Expect(err).NotTo(HaveOccurred(), "Delete MongoDBRole")

				IsMongoDBRoleDeleted(p.Name, p.Namespace)
				IsVaultDatabaseRoleDeleted(p.RoleName())
			})

			It("should be successful", func() {
				_, err := f.DBClient.AuthorizationV1alpha1().MongoDBRoles(mRole.Namespace).Create(&p)
				Expect(err).NotTo(HaveOccurred(), "Create Mongodbole")

				IsVaultDatabaseConfigCreated(p.Spec.DatabaseRef.Name)
				IsVaultDatabaseRoleCreated(p.RoleName())
			})
		})

		Context("Delete MongoDBRole, invalid vault address", func() {
			var p api.MongoDBRole

			BeforeEach(func() {
				p = mRole
				p.Spec.AuthManagerRef = &appcat.AppReference{
					Name:      "invalid",
					Namespace: f.Namespace(),
				}

				_, err := f.DBClient.AuthorizationV1alpha1().MongoDBRoles(mRole.Namespace).Create(&p)
				Expect(err).NotTo(HaveOccurred(), "Create MongoDBRole")

				IsMongoDBRoleCreated(p.Name, p.Namespace)
			})

			It("should be successful", func() {
				err := f.DBClient.AuthorizationV1alpha1().MongoDBRoles(p.Namespace).Delete(p.Name, &metav1.DeleteOptions{})
				Expect(err).NotTo(HaveOccurred(), "Delete MongoDBRole")

				IsMongoDBRoleDeleted(p.Name, p.Namespace)
			})
		})

	})

	Describe("DatabaseAccessRequest", func() {
		var (
			mRole  api.MongoDBRole
			dbAreq api.DatabaseAccessRequest
		)

		BeforeEach(func() {
			mRole = api.MongoDBRole{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "m-role-test1",
					Namespace: f.Namespace(),
				},
				Spec: api.MongoDBRoleSpec{
					AuthManagerRef: f.VaultAppRef,
					DatabaseRef: &core.LocalObjectReference{
						Name: f.MongoAppRef.Name,
					},
					CreationStatements: []string{
						"{ \"db\": \"admin\", \"roles\": [{ \"role\": \"readWrite\" }, {\"role\": \"read\", \"db\": \"foo\"}] }",
					},
					MaxTTL:     "1h",
					DefaultTTL: "300",
				},
			}

			dbAreq = api.DatabaseAccessRequest{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "mongo-cred-1123",
					Namespace: f.Namespace(),
				},
				Spec: api.DatabaseAccessRequestSpec{
					RoleRef: api.RoleReference{
						Kind:      api.ResourceKindMongoDBRole,
						Name:      mRole.Name,
						Namespace: mRole.Namespace,
					},
					Subjects: []rbac.Subject{
						{
							Kind:      rbac.ServiceAccountKind,
							Name:      "sa",
							Namespace: f.Namespace(),
						},
					},
				},
			}
		})

		Context("Create, Approve, Deny DatabaseAccessRequest", func() {
			BeforeEach(func() {
				_, err := f.DBClient.AuthorizationV1alpha1().MongoDBRoles(mRole.Namespace).Create(&mRole)
				Expect(err).NotTo(HaveOccurred(), "Create Mongodbole")

				IsVaultDatabaseConfigCreated(mRole.Spec.DatabaseRef.Name)
				IsVaultDatabaseRoleCreated(mRole.RoleName())
			})

			AfterEach(func() {
				err := f.DBClient.AuthorizationV1alpha1().DatabaseAccessRequests(dbAreq.Namespace).Delete(dbAreq.Name, &metav1.DeleteOptions{})
				Expect(err).NotTo(HaveOccurred(), "Delete DatabaseAccessRequest")

				IsDatabaseAccessRequestDeleted(dbAreq.Name, dbAreq.Namespace)

				err = f.DBClient.AuthorizationV1alpha1().MongoDBRoles(mRole.Namespace).Delete(mRole.Name, &metav1.DeleteOptions{})
				Expect(err).NotTo(HaveOccurred(), "Delete MongoDBRole")

				IsMongoDBRoleDeleted(mRole.Name, mRole.Namespace)
				IsVaultDatabaseRoleDeleted(mRole.RoleName())
			})

			It("create should be successful", func() {
				_, err := f.DBClient.AuthorizationV1alpha1().DatabaseAccessRequests(dbAreq.Namespace).Create(&dbAreq)
				Expect(err).NotTo(HaveOccurred(), "Create DatabaseAccessRequest")

				IsDatabaseAccessRequestCreated(dbAreq.Name, dbAreq.Namespace)
			})

			It("approve should be successful", func() {
				d, err := f.DBClient.AuthorizationV1alpha1().DatabaseAccessRequests(dbAreq.Namespace).Create(&dbAreq)
				Expect(err).NotTo(HaveOccurred(), "Create DatabaseAccessRequest")
				IsDatabaseAccessRequestCreated(dbAreq.Name, dbAreq.Namespace)

				err = f.UpdateDatabaseAccessRequestStatus(&api.DatabaseAccessRequestStatus{
					Conditions: []api.DatabaseAccessRequestCondition{
						{
							Type:           api.AccessApproved,
							LastUpdateTime: metav1.Now(),
						},
					},
				}, d)
				Expect(err).NotTo(HaveOccurred(), "Approve DatabaseAccessRequest")

				IsDatabaseAccessRequestApproved(dbAreq.Name, dbAreq.Namespace)
			})

			It("deny should be successful", func() {
				d, err := f.DBClient.AuthorizationV1alpha1().DatabaseAccessRequests(dbAreq.Namespace).Create(&dbAreq)
				Expect(err).NotTo(HaveOccurred(), "Create DatabaseAccessRequest")
				IsDatabaseAccessRequestCreated(dbAreq.Name, dbAreq.Namespace)

				err = f.UpdateDatabaseAccessRequestStatus(&api.DatabaseAccessRequestStatus{
					Conditions: []api.DatabaseAccessRequestCondition{
						{
							Type:           api.AccessDenied,
							LastUpdateTime: metav1.Now(),
						},
					},
				}, d)
				Expect(err).NotTo(HaveOccurred(), "Deny DatabaseAccessRequest")

				IsDatabaseAccessRequestDenied(dbAreq.Name, dbAreq.Namespace)
			})
		})
	})

})
