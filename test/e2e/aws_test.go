package e2e_test

import (
	"fmt"
	"time"

	api "github.com/kubevault/operator/apis/secretengine/v1alpha1"
	"github.com/kubevault/operator/pkg/controller"
	"github.com/kubevault/operator/pkg/vault"
	"github.com/kubevault/operator/test/e2e/framework"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	core "k8s.io/api/core/v1"
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
				_, err := f.CSClient.SecretengineV1alpha1().AWSRoles(namespace).Get(name, metav1.GetOptions{})
				return err == nil
			}, timeOut, pollingInterval).Should(BeTrue(), "Is AWS role created")
		}

		IsAWSRoleDeleted = func(name, namespace string) {
			By(fmt.Sprintf("Checking Is AWSRole(%s/%s) deleted", namespace, name))
			Eventually(func() bool {
				_, err := f.CSClient.SecretengineV1alpha1().AWSRoles(namespace).Get(name, metav1.GetOptions{})
				return kerrors.IsNotFound(err)
			}, timeOut, pollingInterval).Should(BeTrue(), "Is AWSRole role deleted")
		}

		IsAWSRoleSucceeded = func(name, namespace string) {
			By(fmt.Sprintf("Checking Is AWSRole(%s/%s) succeeded", namespace, name))
			Eventually(func() bool {
				r, err := f.CSClient.SecretengineV1alpha1().AWSRoles(namespace).Get(name, metav1.GetOptions{})
				if err == nil {
					return r.Status.Phase == controller.AWSRolePhaseSuccess
				}
				return false
			}, timeOut, pollingInterval).Should(BeTrue(), "Is AWSRole role succeeded")
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
				err := f.CSClient.SecretengineV1alpha1().AWSRoles(p.Namespace).Delete(p.Name, &metav1.DeleteOptions{})
				Expect(err).NotTo(HaveOccurred(), "Delete AWSRole")

				IsAWSRoleDeleted(p.Name, p.Namespace)
				IsVaultAWSRoleDeleted(p.RoleName())
			})

			It("should be successful", func() {
				_, err := f.CSClient.SecretengineV1alpha1().AWSRoles(awsRole.Namespace).Create(&p)
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

				_, err := f.CSClient.SecretengineV1alpha1().AWSRoles(awsRole.Namespace).Create(&p)
				Expect(err).NotTo(HaveOccurred(), "Create AWSRole")

				IsAWSRoleCreated(p.Name, p.Namespace)
			})

			It("should be successful", func() {
				err := f.CSClient.SecretengineV1alpha1().AWSRoles(p.Namespace).Delete(p.Name, &metav1.DeleteOptions{})
				Expect(err).NotTo(HaveOccurred(), "Delete AWSRole")

				IsAWSRoleDeleted(p.Name, p.Namespace)
			})
		})

	})

})
