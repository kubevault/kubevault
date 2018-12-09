package e2e_test

import (
	"fmt"
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

var _ = Describe("AWS role", func() {

	var f *framework.Invocation

	BeforeEach(func() {
		f = root.Invoke()

	})

	AfterEach(func() {
		time.Sleep(20 * time.Second)
	})

	var (
		// vault related
		IsVaultAWSRoleCreated = func(name string) {
			By(fmt.Sprintf("Checking Is vault aws role created"))
			cl, err := vault.NewClient(f.KubeClient, f.AppcatClient, f.VaultAppRef)
			Expect(err).NotTo(HaveOccurred(), "Get vault client")

			req := cl.NewRequest("GET", fmt.Sprintf("/v1/aws/roles/%s", name))
			Eventually(func() bool {
				_, err := cl.RawRequest(req)
				return err == nil
			}, timeOut, pollingInterval).Should(BeTrue(), "Is vault aws role created")
		}

		IsVaultAWSRoleDeleted = func(name string) {
			By(fmt.Sprintf("Checking Is vault aws role deleted"))
			cl, err := vault.NewClient(f.KubeClient, f.AppcatClient, f.VaultAppRef)
			Expect(err).NotTo(HaveOccurred(), "Get vault client")

			req := cl.NewRequest("GET", fmt.Sprintf("/v1/aws/roles/%s", name))
			Eventually(func() bool {
				_, err := cl.RawRequest(req)
				return err != nil
			}, timeOut, pollingInterval).Should(BeTrue(), "Is vault aws role deleted")
		}

		IsAWSRoleCreated = func(name, namespace string) {
			By(fmt.Sprintf("Checking Is AWSRole(%s/%s) created", namespace, name))
			Eventually(func() bool {
				_, err := f.CSClient.EngineV1alpha1().AWSRoles(namespace).Get(name, metav1.GetOptions{})
				return err == nil
			}, timeOut, pollingInterval).Should(BeTrue(), "Is AWS role created")
		}

		IsAWSRoleDeleted = func(name, namespace string) {
			By(fmt.Sprintf("Checking Is AWSRole(%s/%s) deleted", namespace, name))
			Eventually(func() bool {
				_, err := f.CSClient.EngineV1alpha1().AWSRoles(namespace).Get(name, metav1.GetOptions{})
				return kerrors.IsNotFound(err)
			}, timeOut, pollingInterval).Should(BeTrue(), "Is AWSRole role deleted")
		}

		IsAWSRoleSucceeded = func(name, namespace string) {
			By(fmt.Sprintf("Checking Is AWSRole(%s/%s) succeeded", namespace, name))
			Eventually(func() bool {
				r, err := f.CSClient.EngineV1alpha1().AWSRoles(namespace).Get(name, metav1.GetOptions{})
				if err == nil {
					return r.Status.Phase == controller.AWSRolePhaseSuccess
				}
				return false
			}, timeOut, pollingInterval).Should(BeTrue(), "Is AWSRole role succeeded")
		}

		IsAWSAccessKeyRequestCreated = func(name, namespace string) {
			By(fmt.Sprintf("Checking Is AWSAccessKeyRequest(%s/%s) created", namespace, name))
			Eventually(func() bool {
				_, err := f.CSClient.EngineV1alpha1().AWSAccessKeyRequests(namespace).Get(name, metav1.GetOptions{})
				return err == nil
			}, timeOut, pollingInterval).Should(BeTrue(), "Is AWSAccessKeyRequest created")
		}

		IsAWSAccessKeyRequestDeleted = func(name, namespace string) {
			By(fmt.Sprintf("Checking Is AWSAccessKeyRequest(%s/%s) deleted", namespace, name))
			Eventually(func() bool {
				_, err := f.CSClient.EngineV1alpha1().AWSAccessKeyRequests(namespace).Get(name, metav1.GetOptions{})
				return kerrors.IsNotFound(err)
			}, timeOut, pollingInterval).Should(BeTrue(), "Is AWSAccessKeyRequest deleted")
		}

		IsAWSAccessKeyRequestApproved = func(name, namespace string) {
			By(fmt.Sprintf("Checking Is AWSAccessKeyRequest(%s/%s) apporved", namespace, name))
			Eventually(func() bool {
				d, err := f.CSClient.EngineV1alpha1().AWSAccessKeyRequests(namespace).Get(name, metav1.GetOptions{})
				if err == nil {
					return d.Status.Lease != nil
				}
				return false
			}, timeOut, pollingInterval).Should(BeTrue(), "Is AWSAccessKeyRequest approved")
		}
		IsAWSAccessKeyRequestDenied = func(name, namespace string) {
			By(fmt.Sprintf("Checking Is AWSAccessKeyRequest(%s/%s) denied", namespace, name))
			Eventually(func() bool {
				d, err := f.CSClient.EngineV1alpha1().AWSAccessKeyRequests(namespace).Get(name, metav1.GetOptions{})
				if err == nil {
					for _, c := range d.Status.Conditions {
						if c.Type == api.AccessDenied {
							return true
						}
					}
				}
				return false
			}, timeOut, pollingInterval).Should(BeTrue(), "Is AWSAccessKeyRequest denied")
		}
	)

	Describe("AWSRole", func() {
		var (
			awsRole api.AWSRole
		)

		const awsCredSecret = "aws-cred-12235"

		BeforeEach(func() {

			_, err := f.KubeClient.CoreV1().Secrets(f.Namespace()).Create(&core.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: f.Namespace(),
					Name:      awsCredSecret,
				},
				Data: map[string][]byte{
					"access_key": []byte("id"),
					"secret_key": []byte("secret"),
				},
			})
			Expect(err).NotTo(HaveOccurred(), "create aws secret")

			awsRole = api.AWSRole{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "m-role-test1",
					Namespace: f.Namespace(),
				},
				Spec: api.AWSRoleSpec{
					AuthManagerRef: f.VaultAppRef,
					Policy: `
						{
							  "Version": "2012-10-17",
							  "Statement": [
								{
								  "Effect": "Allow",
								  "Action": "ec2:*",
								  "Resource": "*"
								}
							  ]
							}
						`,
					Config: &api.AWSConfig{
						CredentialSecret: awsCredSecret,
						Region:           "us-east-1",
						LeaseConfig: &api.LeaseConfig{
							LeaseMax: "1h",
							Lease:    "1h",
						},
					},
				},
			}
		})

		AfterEach(func() {
			err := f.KubeClient.CoreV1().Secrets(f.Namespace()).Delete(awsCredSecret, &metav1.DeleteOptions{})
			Expect(err).NotTo(HaveOccurred(), "delete aws secret")
		})

		Context("Create AWSRole", func() {
			var p api.AWSRole

			BeforeEach(func() {
				p = awsRole
			})

			AfterEach(func() {
				err := f.CSClient.EngineV1alpha1().AWSRoles(p.Namespace).Delete(p.Name, &metav1.DeleteOptions{})
				Expect(err).NotTo(HaveOccurred(), "Delete AWSRole")

				IsAWSRoleDeleted(p.Name, p.Namespace)
				IsVaultAWSRoleDeleted(p.RoleName())
			})

			It("should be successful", func() {
				_, err := f.CSClient.EngineV1alpha1().AWSRoles(awsRole.Namespace).Create(&p)
				Expect(err).NotTo(HaveOccurred(), "Create AWSole")

				IsVaultAWSRoleCreated(p.RoleName())
				IsAWSRoleSucceeded(p.Name, p.Namespace)
			})
		})

		Context("Delete AWSRole, invalid vault address", func() {
			var p api.AWSRole

			BeforeEach(func() {
				p = awsRole
				p.Spec.AuthManagerRef = &appcat.AppReference{
					Name:      "invalid",
					Namespace: f.Namespace(),
				}

				_, err := f.CSClient.EngineV1alpha1().AWSRoles(awsRole.Namespace).Create(&p)
				Expect(err).NotTo(HaveOccurred(), "Create AWSRole")

				IsAWSRoleCreated(p.Name, p.Namespace)
			})

			It("should be successful", func() {
				err := f.CSClient.EngineV1alpha1().AWSRoles(p.Namespace).Delete(p.Name, &metav1.DeleteOptions{})
				Expect(err).NotTo(HaveOccurred(), "Delete AWSRole")

				IsAWSRoleDeleted(p.Name, p.Namespace)
			})
		})

	})

	Describe("AWSAccessKeyRequest", func() {
		var (
			awsRole  api.AWSRole
			awsAKReq api.AWSAccessKeyRequest
		)

		const awsCredSecret = "aws-cred-12255"

		BeforeEach(func() {
			_, err := f.KubeClient.CoreV1().Secrets(f.Namespace()).Create(&core.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: f.Namespace(),
					Name:      awsCredSecret,
				},
				Data: map[string][]byte{
					"access_key": []byte(os.Getenv("AWS_ACCESS_KEY_ID")),
					"secret_key": []byte(os.Getenv("AWS_SECRET_ACCESS_KEY")),
				},
			})
			Expect(err).NotTo(HaveOccurred(), "create aws secret")

			awsRole = api.AWSRole{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "m-role-test1",
					Namespace: f.Namespace(),
				},
				Spec: api.AWSRoleSpec{
					AuthManagerRef: f.VaultAppRef,
					Policy: `
						{
							  "Version": "2012-10-17",
							  "Statement": [
								{
								  "Effect": "Allow",
								  "Action": "ec2:*",
								  "Resource": "*"
								}
							  ]
							}
						`,
					Config: &api.AWSConfig{
						CredentialSecret: awsCredSecret,
						Region:           "us-east-1",
						LeaseConfig: &api.LeaseConfig{
							LeaseMax: "1h",
							Lease:    "1h",
						},
					},
				},
			}

			awsAKReq = api.AWSAccessKeyRequest{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "aws-access-1123",
					Namespace: f.Namespace(),
				},
				Spec: api.AWSAccessKeyRequestSpec{
					RoleRef: api.RoleReference{
						Name:      awsRole.Name,
						Namespace: awsRole.Namespace,
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

		AfterEach(func() {
			err := f.KubeClient.CoreV1().Secrets(f.Namespace()).Delete(awsCredSecret, &metav1.DeleteOptions{})
			Expect(err).NotTo(HaveOccurred(), "delete aws secret")
		})

		Context("Create, Approve, Deny AWSAccessKeyRequest", func() {
			BeforeEach(func() {
				r, err := f.CSClient.EngineV1alpha1().AWSRoles(awsRole.Namespace).Create(&awsRole)
				Expect(err).NotTo(HaveOccurred(), "Create AWSRole")

				IsVaultAWSRoleCreated(r.RoleName())
				IsAWSRoleSucceeded(r.Name, r.Namespace)
			})

			AfterEach(func() {
				err := f.CSClient.EngineV1alpha1().AWSAccessKeyRequests(awsAKReq.Namespace).Delete(awsAKReq.Name, &metav1.DeleteOptions{})
				Expect(err).NotTo(HaveOccurred(), "Delete AWSAccessKeyRequest")

				IsAWSAccessKeyRequestDeleted(awsAKReq.Name, awsAKReq.Namespace)

				err = f.CSClient.EngineV1alpha1().AWSRoles(awsRole.Namespace).Delete(awsRole.Name, &metav1.DeleteOptions{})
				Expect(err).NotTo(HaveOccurred(), "Delete AWSRole")

				IsAWSRoleDeleted(awsRole.Name, awsRole.Namespace)
				IsVaultAWSRoleDeleted(awsRole.RoleName())
			})

			It("create should be successful", func() {
				_, err := f.CSClient.EngineV1alpha1().AWSAccessKeyRequests(awsAKReq.Namespace).Create(&awsAKReq)
				Expect(err).NotTo(HaveOccurred(), "Create AWSAccessKeyRequest")

				IsAWSAccessKeyRequestCreated(awsAKReq.Name, awsAKReq.Namespace)
			})

			It("approve should be successful", func() {
				d, err := f.CSClient.EngineV1alpha1().AWSAccessKeyRequests(awsAKReq.Namespace).Create(&awsAKReq)
				Expect(err).NotTo(HaveOccurred(), "Create AWSAccessKeyRequest")
				IsAWSAccessKeyRequestCreated(awsAKReq.Name, awsAKReq.Namespace)

				err = f.UpdateAWSAccessKeyRequestStatus(&api.AWSAccessKeyRequestStatus{
					Conditions: []api.AWSAccessKeyRequestCondition{
						{
							Type:           api.AccessApproved,
							LastUpdateTime: metav1.Now(),
						},
					},
				}, d)
				Expect(err).NotTo(HaveOccurred(), "Approve AWSAccessKeyRequest")

				IsAWSAccessKeyRequestApproved(awsAKReq.Name, awsAKReq.Namespace)
			})

			It("deny should be successful", func() {
				d, err := f.CSClient.EngineV1alpha1().AWSAccessKeyRequests(awsAKReq.Namespace).Create(&awsAKReq)
				Expect(err).NotTo(HaveOccurred(), "Create AWSAccessKeyRequest")
				IsAWSAccessKeyRequestCreated(awsAKReq.Name, awsAKReq.Namespace)

				err = f.UpdateAWSAccessKeyRequestStatus(&api.AWSAccessKeyRequestStatus{
					Conditions: []api.AWSAccessKeyRequestCondition{
						{
							Type:           api.AccessDenied,
							LastUpdateTime: metav1.Now(),
						},
					},
				}, d)
				Expect(err).NotTo(HaveOccurred(), "Deny AWSAccessKeyRequest")

				IsAWSAccessKeyRequestDenied(awsAKReq.Name, awsAKReq.Namespace)
			})
		})
	})
})
