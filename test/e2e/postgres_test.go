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
	api "kubedb.dev/apimachinery/apis/authorization/v1alpha1"
	"kubevault.dev/operator/pkg/vault"
	"kubevault.dev/operator/test/e2e/framework"
)

var _ = Describe("Postgres role and role binding", func() {

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

		IsPostgresRoleCreated = func(name, namespace string) {
			By(fmt.Sprintf("Checking Is PostgresRole(%s/%s) created", namespace, name))
			Eventually(func() bool {
				_, err := f.DBClient.AuthorizationV1alpha1().PostgresRoles(namespace).Get(name, metav1.GetOptions{})
				return err == nil
			}, timeOut, pollingInterval).Should(BeTrue(), "Is PostgresRole role created")
		}

		IsPostgresRoleDeleted = func(name, namespace string) {
			By(fmt.Sprintf("Checking Is PostgresRole(%s/%s) deleted", namespace, name))
			Eventually(func() bool {
				_, err := f.DBClient.AuthorizationV1alpha1().PostgresRoles(namespace).Get(name, metav1.GetOptions{})
				return kerrors.IsNotFound(err)
			}, timeOut, pollingInterval).Should(BeTrue(), "Is PostgresRole role deleted")
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

	Describe("PostgresRole", func() {
		var (
			pgRole api.PostgresRole
		)

		BeforeEach(func() {
			pgRole = api.PostgresRole{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "pg-role-test1",
					Namespace: f.Namespace(),
				},
				Spec: api.PostgresRoleSpec{
					AuthManagerRef: f.VaultAppRef,
					DatabaseRef: &core.LocalObjectReference{
						Name: f.PostgresAppRef.Name,
					},
					CreationStatements: []string{
						"CREATE ROLE \"{{name}}\" WITH LOGIN PASSWORD '{{password}}' VALID UNTIL '{{expiration}}';",
						"GRANT SELECT ON ALL TABLES IN SCHEMA public TO \"{{name}}\";",
					},
					MaxTTL:     "1h",
					DefaultTTL: "300",
				},
			}
		})

		Context("Create PostgresRole", func() {
			var p api.PostgresRole

			BeforeEach(func() {
				p = pgRole
			})

			AfterEach(func() {
				err := f.DBClient.AuthorizationV1alpha1().PostgresRoles(p.Namespace).Delete(p.Name, &metav1.DeleteOptions{})
				Expect(err).NotTo(HaveOccurred(), "Delete PostgresRole")

				IsPostgresRoleDeleted(p.Name, p.Namespace)
				IsVaultDatabaseRoleDeleted(p.RoleName())
			})

			It("should be successful", func() {
				_, err := f.DBClient.AuthorizationV1alpha1().PostgresRoles(pgRole.Namespace).Create(&p)
				Expect(err).NotTo(HaveOccurred(), "Create PostgresRole")

				IsVaultDatabaseConfigCreated(p.Spec.DatabaseRef.Name)
				IsVaultDatabaseRoleCreated(p.RoleName())
			})
		})

		Context("Delete PostgresRole, invalid vault address", func() {
			var p api.PostgresRole

			BeforeEach(func() {
				p = pgRole
				p.Spec.AuthManagerRef = &appcat.AppReference{
					Name:      "invalid",
					Namespace: f.Namespace(),
				}

				_, err := f.DBClient.AuthorizationV1alpha1().PostgresRoles(pgRole.Namespace).Create(&p)
				Expect(err).NotTo(HaveOccurred(), "Create PostgresRole")

				IsPostgresRoleCreated(p.Name, p.Namespace)
			})

			It("should be successful", func() {
				err := f.DBClient.AuthorizationV1alpha1().PostgresRoles(p.Namespace).Delete(p.Name, &metav1.DeleteOptions{})
				Expect(err).NotTo(HaveOccurred(), "Delete PostgresRole")

				IsPostgresRoleDeleted(p.Name, p.Namespace)
			})
		})

	})

	Describe("DatabaseAccessRequest", func() {
		var (
			pRole  api.PostgresRole
			dbAreq api.DatabaseAccessRequest
		)

		BeforeEach(func() {
			pRole = api.PostgresRole{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "m-role-test1",
					Namespace: f.Namespace(),
				},
				Spec: api.PostgresRoleSpec{
					AuthManagerRef: f.VaultAppRef,
					DatabaseRef: &core.LocalObjectReference{
						Name: f.PostgresAppRef.Name,
					},
					CreationStatements: []string{
						"CREATE ROLE \"{{name}}\" WITH LOGIN PASSWORD '{{password}}' VALID UNTIL '{{expiration}}';",
						"GRANT SELECT ON ALL TABLES IN SCHEMA public TO \"{{name}}\";",
					},
					MaxTTL:     "1h",
					DefaultTTL: "300",
				},
			}

			dbAreq = api.DatabaseAccessRequest{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "postgres-cred-1123",
					Namespace: f.Namespace(),
				},
				Spec: api.DatabaseAccessRequestSpec{
					RoleRef: api.RoleReference{
						Kind:      api.ResourceKindPostgresRole,
						Name:      pRole.Name,
						Namespace: pRole.Namespace,
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
				_, err := f.DBClient.AuthorizationV1alpha1().PostgresRoles(pRole.Namespace).Create(&pRole)
				Expect(err).NotTo(HaveOccurred(), "Create PostgresRole")

				IsVaultDatabaseConfigCreated(pRole.Spec.DatabaseRef.Name)
				IsVaultDatabaseRoleCreated(pRole.RoleName())
			})

			AfterEach(func() {
				err := f.DBClient.AuthorizationV1alpha1().DatabaseAccessRequests(dbAreq.Namespace).Delete(dbAreq.Name, &metav1.DeleteOptions{})
				Expect(err).NotTo(HaveOccurred(), "Delete DatabaseAccessRequest")

				IsDatabaseAccessRequestDeleted(dbAreq.Name, dbAreq.Namespace)

				err = f.DBClient.AuthorizationV1alpha1().PostgresRoles(pRole.Namespace).Delete(pRole.Name, &metav1.DeleteOptions{})
				Expect(err).NotTo(HaveOccurred(), "Delete PostgresRole")

				IsPostgresRoleDeleted(pRole.Name, pRole.Namespace)
				IsVaultDatabaseRoleDeleted(pRole.RoleName())
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
