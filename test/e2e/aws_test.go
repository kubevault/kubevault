/*
Copyright AppsCode Inc. and Contributors

Licensed under the AppsCode Community License 1.0.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    https://github.com/appscode/licenses/raw/1.0.0/AppsCode-Community-1.0.0.md

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

	api "kubevault.dev/apimachinery/apis/engine/v1alpha1"
	"kubevault.dev/operator/pkg/controller"
	"kubevault.dev/operator/test/e2e/framework"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	core "k8s.io/api/core/v1"
	v1 "k8s.io/api/rbac/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kmapi "kmodules.xyz/client-go/api/v1"
	awsconsts "kmodules.xyz/constants/aws"
)

var _ = Describe("AWS Secret Engine", func() {

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
		IsAWSRoleCreated = func(name, namespace string) {
			By(fmt.Sprintf("Checking whether AWSRole:(%s/%s) role is created", namespace, name))
			Eventually(func() bool {
				_, err := f.CSClient.EngineV1alpha1().AWSRoles(namespace).Get(context.TODO(), name, metav1.GetOptions{})
				return err == nil
			}, timeOut, pollingInterval).Should(BeTrue(), "AWSRole is created")
		}
		IsAWSRoleDeleted = func(name, namespace string) {
			By(fmt.Sprintf("Checking whether AWSRole:(%s/%s) is deleted", namespace, name))
			Eventually(func() bool {
				_, err := f.CSClient.EngineV1alpha1().AWSRoles(namespace).Get(context.TODO(), name, metav1.GetOptions{})
				return kerrors.IsNotFound(err)
			}, timeOut, pollingInterval).Should(BeTrue(), "AWSRole is deleted")
		}
		IsAWSRoleSucceeded = func(name, namespace string) {
			By(fmt.Sprintf("Checking whether AWSRole:(%s/%s) is succeeded", namespace, name))
			Eventually(func() bool {
				r, err := f.CSClient.EngineV1alpha1().AWSRoles(namespace).Get(context.TODO(), name, metav1.GetOptions{})
				if err == nil {
					return r.Status.Phase == controller.AWSRolePhaseSuccess
				}
				return false
			}, timeOut, pollingInterval).Should(BeTrue(), "AWSRole status is succeeded")

		}

		IsAWSRoleFailed = func(name, namespace string) {
			By(fmt.Sprintf("Checking whether AWSRole:(%s/%s) is failed", namespace, name))
			Eventually(func() bool {
				r, err := f.CSClient.EngineV1alpha1().AWSRoles(namespace).Get(context.TODO(), name, metav1.GetOptions{})
				if err == nil {
					return r.Status.Phase != controller.AWSRolePhaseSuccess && len(r.Status.Conditions) != 0
				}
				return false
			}, timeOut, pollingInterval).Should(BeTrue(), "AWSRole status is failed")
		}
		IsAWSAccessKeyRequestCreated = func(name, namespace string) {
			By(fmt.Sprintf("Checking whether AWSAccessKeyRequest:(%s/%s) is created", namespace, name))
			Eventually(func() bool {
				_, err := f.CSClient.EngineV1alpha1().AWSAccessKeyRequests(namespace).Get(context.TODO(), name, metav1.GetOptions{})
				return err == nil
			}, timeOut, pollingInterval).Should(BeTrue(), "AWSAccessKeyRequest is created")
		}
		IsAWSAccessKeyRequestDeleted = func(name, namespace string) {
			By(fmt.Sprintf("Checking whether AWSAccessKeyRequest:(%s/%s) is deleted", namespace, name))
			Eventually(func() bool {
				_, err := f.CSClient.EngineV1alpha1().AWSAccessKeyRequests(namespace).Get(context.TODO(), name, metav1.GetOptions{})
				return kerrors.IsNotFound(err)
			}, timeOut, pollingInterval).Should(BeTrue(), "AWSAccessKeyRequest is deleted")
		}
		IsAWSAKRConditionApproved = func(name, namespace string) {
			By("Checking whether AWSAccessKeyRequestConditions-> Type: Approved")
			Eventually(func() bool {
				crd, err := f.CSClient.EngineV1alpha1().AWSAccessKeyRequests(namespace).Get(context.TODO(), name, metav1.GetOptions{})
				if err == nil {
					return kmapi.IsConditionTrue(crd.Status.Conditions, kmapi.ConditionRequestApproved)
				}
				return false
			}, timeOut, pollingInterval).Should(BeTrue(), "Conditions-> Type : Approved")
		}
		IsAWSAKRConditionDenied = func(name, namespace string) {
			By("Checking whether AWSAccessKeyRequestConditions-> Type: Denied")
			Eventually(func() bool {
				crd, err := f.CSClient.EngineV1alpha1().AWSAccessKeyRequests(namespace).Get(context.TODO(), name, metav1.GetOptions{})
				if err == nil {
					return kmapi.IsConditionTrue(crd.Status.Conditions, kmapi.ConditionRequestDenied)
				}
				return false
			}, timeOut, pollingInterval).Should(BeTrue(), "Conditions-> Type: Denied")
		}
		IsAWSAccessKeySecretCreated = func(name, namespace string) {
			By("Checking whether AWSAccessKeySecret is created")
			Eventually(func() bool {
				crd, err := f.CSClient.EngineV1alpha1().AWSAccessKeyRequests(namespace).Get(context.TODO(), name, metav1.GetOptions{})
				if err == nil && crd.Status.Secret != nil {
					_, err2 := f.KubeClient.CoreV1().Secrets(namespace).Get(context.TODO(), crd.Status.Secret.Name, metav1.GetOptions{})
					return err2 == nil
				}
				return false
			}, timeOut, pollingInterval).Should(BeTrue(), "AWSAccessKeySecret is created")
		}
		IsAWSAccessKeySecretDeleted = func(secretName, namespace string) {
			By("Checking whether AWSAccessKeySecret is deleted")
			Eventually(func() bool {
				_, err2 := f.KubeClient.CoreV1().Secrets(namespace).Get(context.TODO(), secretName, metav1.GetOptions{})
				return kerrors.IsNotFound(err2)
			}, timeOut, pollingInterval).Should(BeTrue(), "AWSAccessKeySecret is deleted")
		}
	)

	BeforeEach(func() {
		f = root.Invoke()
		if !framework.SelfHostedOperator {
			Skip("Skipping AWS secret engine tests because the operator isn't running inside cluster")
		}
	})

	AfterEach(func() {
		time.Sleep(20 * time.Second)
	})

	Describe("AWSRole", func() {

		var (
			awsCredentials core.Secret
			awsRole        api.AWSRole
			awsSE          api.SecretEngine
		)

		const (
			awsCredSecret   = "aws-cred-3224"
			awsRoleName     = "my-aws-roleset-4325"
			awsSecretEngine = "my-aws-secretengine-3423423"
		)

		BeforeEach(func() {
			credentials := awsconsts.CredentialsFromEnv()
			if len(credentials) == 0 {
				Skip("Skipping aws secret engine tests, empty env")
			}
			awsCredentials = core.Secret{
				TypeMeta: metav1.TypeMeta{},
				ObjectMeta: metav1.ObjectMeta{
					Name:      awsCredSecret,
					Namespace: f.Namespace(),
				},
				Data: credentials,
			}
			_, err := f.KubeClient.CoreV1().Secrets(f.Namespace()).Create(context.TODO(), &awsCredentials, metav1.CreateOptions{})
			Expect(err).NotTo(HaveOccurred(), "Create aws credentials secret")

			awsRole = api.AWSRole{
				ObjectMeta: metav1.ObjectMeta{
					Name:      awsRoleName,
					Namespace: f.Namespace(),
				},
				Spec: api.AWSRoleSpec{
					VaultRef: core.LocalObjectReference{
						Name: f.VaultAppRef.Name,
					},
					CredentialType: api.AWSCredentialIAMUser,
					PolicyDocument: `
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
				},
			}

			awsSE = api.SecretEngine{
				ObjectMeta: metav1.ObjectMeta{
					Name:      awsSecretEngine,
					Namespace: f.Namespace(),
				},
				Spec: api.SecretEngineSpec{
					VaultRef: core.LocalObjectReference{
						Name: f.VaultAppRef.Name,
					},
					Path: "aws",
					SecretEngineConfiguration: api.SecretEngineConfiguration{
						AWS: &api.AWSConfiguration{
							CredentialSecret: awsCredSecret,
							Region:           "us-west-2",
							LeaseConfig: &api.LeaseConfig{
								Lease:    "1h",
								LeaseMax: "1h",
							},
						},
					},
				},
			}
		})

		AfterEach(func() {
			err := f.KubeClient.CoreV1().Secrets(f.Namespace()).Delete(context.TODO(), awsCredSecret, metav1.DeleteOptions{})
			Expect(err).NotTo(HaveOccurred(), "Delete AWS credentials secret")
		})

		Context("Create AWSRole", func() {
			var p api.AWSRole
			var se api.SecretEngine

			BeforeEach(func() {
				p = awsRole
				se = awsSE
			})

			AfterEach(func() {
				By("Deleting AWSRole...")
				err := f.CSClient.EngineV1alpha1().AWSRoles(awsRole.Namespace).Delete(context.TODO(), p.Name, metav1.DeleteOptions{})
				Expect(err).NotTo(HaveOccurred(), "Delete AWSRole")

				IsAWSRoleDeleted(p.Name, p.Namespace)

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

				By("Creating AWSRole...")
				_, err = f.CSClient.EngineV1alpha1().AWSRoles(p.Namespace).Create(context.TODO(), &p, metav1.CreateOptions{})
				Expect(err).NotTo(HaveOccurred(), "Create AWSRole")

				IsAWSRoleCreated(p.Name, p.Namespace)
				IsAWSRoleSucceeded(p.Name, p.Namespace)
			})

		})

		Context("Create AWSRole without enabling secretEngine", func() {
			var p api.AWSRole

			BeforeEach(func() {
				p = awsRole
			})

			AfterEach(func() {
				By("Deleting AWSRole...")
				err := f.CSClient.EngineV1alpha1().AWSRoles(awsRole.Namespace).Delete(context.TODO(), p.Name, metav1.DeleteOptions{})
				Expect(err).NotTo(HaveOccurred(), "Delete AWSRole")

				IsAWSRoleDeleted(p.Name, p.Namespace)

			})

			It("Should be failed making AWSRole", func() {

				By("Creating AWSRole...")
				_, err := f.CSClient.EngineV1alpha1().AWSRoles(p.Namespace).Create(context.TODO(), &p, metav1.CreateOptions{})
				Expect(err).NotTo(HaveOccurred(), "Create AWSRole")

				IsAWSRoleCreated(p.Name, p.Namespace)
				IsAWSRoleFailed(p.Name, p.Namespace)
			})
		})

	})

	Describe("AWSAccessKeyRequest", func() {

		var (
			awsCredentials core.Secret
			awsRole        api.AWSRole
			awsSE          api.SecretEngine
			awsAKR         api.AWSAccessKeyRequest
		)

		const (
			awsCredSecret   = "aws-cred-3224"
			awsRoleName     = "my-aws-roleset-4325"
			awsSecretEngine = "my-aws-secretengine-3423423"
			awsAKRName      = "my-aws-token-2345"
		)

		BeforeEach(func() {
			credentials := awsconsts.CredentialsFromEnv()
			if len(credentials) == 0 {
				Skip("Skipping aws secret engine tests, empty env")
			}
			awsCredentials = core.Secret{
				TypeMeta: metav1.TypeMeta{},
				ObjectMeta: metav1.ObjectMeta{
					Name:      awsCredSecret,
					Namespace: f.Namespace(),
				},
				Data: credentials,
			}
			_, err := f.KubeClient.CoreV1().Secrets(f.Namespace()).Create(context.TODO(), &awsCredentials, metav1.CreateOptions{})
			Expect(err).NotTo(HaveOccurred(), "Create aws credentials secret")

			awsSE = api.SecretEngine{
				ObjectMeta: metav1.ObjectMeta{
					Name:      awsSecretEngine,
					Namespace: f.Namespace(),
				},
				Spec: api.SecretEngineSpec{
					VaultRef: core.LocalObjectReference{
						Name: f.VaultAppRef.Name,
					},
					Path: "aws",
					SecretEngineConfiguration: api.SecretEngineConfiguration{
						AWS: &api.AWSConfiguration{
							CredentialSecret: awsCredSecret,
							Region:           "us-west-2",
							LeaseConfig: &api.LeaseConfig{
								Lease:    "1h",
								LeaseMax: "1h",
							},
						},
					},
				},
			}
			_, err = f.CSClient.EngineV1alpha1().SecretEngines(awsSE.Namespace).Create(context.TODO(), &awsSE, metav1.CreateOptions{})
			Expect(err).NotTo(HaveOccurred(), "Create aws SecretEngine")
			IsSecretEngineCreated(awsSE.Name, awsSE.Namespace)

			awsRole = api.AWSRole{
				ObjectMeta: metav1.ObjectMeta{
					Name:      awsRoleName,
					Namespace: f.Namespace(),
				},
				Spec: api.AWSRoleSpec{
					VaultRef: core.LocalObjectReference{
						Name: f.VaultAppRef.Name,
					},
					CredentialType: api.AWSCredentialIAMUser,
					PolicyDocument: `
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
				},
			}

			awsAKR = api.AWSAccessKeyRequest{
				ObjectMeta: metav1.ObjectMeta{
					Name:      awsAKRName,
					Namespace: f.Namespace(),
				},
				Spec: api.AWSAccessKeyRequestSpec{
					RoleRef: api.RoleRef{
						Name:      awsRoleName,
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
			err := f.KubeClient.CoreV1().Secrets(f.Namespace()).Delete(context.TODO(), awsCredSecret, metav1.DeleteOptions{})
			Expect(err).NotTo(HaveOccurred(), "Delete AWS credentials secret")

			err = f.CSClient.EngineV1alpha1().SecretEngines(awsSE.Namespace).Delete(context.TODO(), awsSE.Name, metav1.DeleteOptions{})
			Expect(err).NotTo(HaveOccurred(), "Delete AWS SecretEngine")
			IsSecretEngineDeleted(awsSE.Name, awsSE.Namespace)
		})

		Context("Create, Approve, Deny AWSAccessKeyRequests", func() {
			BeforeEach(func() {
				_, err := f.CSClient.EngineV1alpha1().AWSRoles(awsRole.Namespace).Create(context.TODO(), &awsRole, metav1.CreateOptions{})
				Expect(err).NotTo(HaveOccurred(), "Create AWSRole")

				IsAWSRoleCreated(awsRole.Name, awsRole.Namespace)
				IsAWSRoleSucceeded(awsRole.Name, awsRole.Namespace)

			})

			AfterEach(func() {
				err := f.CSClient.EngineV1alpha1().AWSAccessKeyRequests(awsAKR.Namespace).Delete(context.TODO(), awsAKR.Name, metav1.DeleteOptions{})
				Expect(err).NotTo(HaveOccurred(), "Delete AWSAccessKeyRequest")
				IsAWSAccessKeyRequestDeleted(awsAKR.Name, awsAKR.Namespace)

				err = f.CSClient.EngineV1alpha1().AWSRoles(awsRole.Namespace).Delete(context.TODO(), awsRole.Name, metav1.DeleteOptions{})
				Expect(err).NotTo(HaveOccurred(), "Delete AWSRole")
				IsAWSRoleDeleted(awsRole.Name, awsRole.Namespace)
			})

			It("Should be successful, Create AWSAccessKeyRequest", func() {
				_, err := f.CSClient.EngineV1alpha1().AWSAccessKeyRequests(awsAKR.Namespace).Create(context.TODO(), &awsAKR, metav1.CreateOptions{})
				Expect(err).NotTo(HaveOccurred(), "Create AWSAccessKeyRequest")

				IsAWSAccessKeyRequestCreated(awsAKR.Name, awsAKR.Namespace)
			})

			It("Should be successful, Condition approved", func() {
				By("Creating AWSAccessKeyRequest...")
				r, err := f.CSClient.EngineV1alpha1().AWSAccessKeyRequests(awsAKR.Namespace).Create(context.TODO(), &awsAKR, metav1.CreateOptions{})
				Expect(err).NotTo(HaveOccurred(), "Create AWSAccessKeyRequest")

				IsAWSAccessKeyRequestCreated(awsAKR.Name, awsAKR.Namespace)

				By("Updating AWS AccessKeyRequest status...")
				err = f.UpdateAWSAccessKeyRequestStatus(&api.AWSAccessKeyRequestStatus{
					Conditions: []kmapi.Condition{
						{
							Type:               kmapi.ConditionRequestApproved,
							Status:             core.ConditionTrue,
							LastTransitionTime: metav1.Now(),
						},
					},
					Phase: api.RequestStatusPhaseApproved,
				}, r)
				Expect(err).NotTo(HaveOccurred(), "Update conditions: Approved")
				IsAWSAKRConditionApproved(awsAKR.Name, awsAKR.Namespace)
			})

			It("Should be successful, Condition denied", func() {
				By("Creating AWSAccessKeyRequest...")
				r, err := f.CSClient.EngineV1alpha1().AWSAccessKeyRequests(awsAKR.Namespace).Create(context.TODO(), &awsAKR, metav1.CreateOptions{})
				Expect(err).NotTo(HaveOccurred(), "Create AWSAccessKeyRequest")

				IsAWSAccessKeyRequestCreated(awsAKR.Name, awsAKR.Namespace)

				By("Updating AWS AccessKeyRequest status...")
				err = f.UpdateAWSAccessKeyRequestStatus(&api.AWSAccessKeyRequestStatus{
					Conditions: []kmapi.Condition{
						{
							Type:               kmapi.ConditionRequestDenied,
							Status:             core.ConditionTrue,
							LastTransitionTime: metav1.Now(),
						},
					},
					Phase: api.RequestStatusPhaseDenied,
				}, r)
				Expect(err).NotTo(HaveOccurred(), "Update conditions: Denied")

				IsAWSAKRConditionDenied(awsAKR.Name, awsAKR.Namespace)
			})
		})

		Context("Create iam_secret", func() {
			var (
				secretName string
			)

			BeforeEach(func() {

				By("Creating AWSRole...")
				r, err := f.CSClient.EngineV1alpha1().AWSRoles(awsRole.Namespace).Create(context.TODO(), &awsRole, metav1.CreateOptions{})
				Expect(err).NotTo(HaveOccurred(), "Create AWSRole")

				IsAWSRoleSucceeded(r.Name, r.Namespace)

			})

			AfterEach(func() {
				By("Deleting AWS accesskeyrequest...")
				err := f.CSClient.EngineV1alpha1().AWSAccessKeyRequests(awsAKR.Namespace).Delete(context.TODO(), awsAKR.Name, metav1.DeleteOptions{})
				Expect(err).NotTo(HaveOccurred(), "Delete AWSAccessKeyRequest")

				IsAWSAccessKeyRequestDeleted(awsAKR.Name, awsAKR.Namespace)
				IsAWSAccessKeySecretDeleted(secretName, awsAKR.Namespace)

				By("Deleting AWSRole...")
				err = f.CSClient.EngineV1alpha1().AWSRoles(awsRole.Namespace).Delete(context.TODO(), awsRole.Name, metav1.DeleteOptions{})
				Expect(err).NotTo(HaveOccurred(), "Delete AWSRole")

				IsAWSRoleDeleted(awsRole.Name, awsRole.Namespace)
			})

			It("Should be successful, Create Access Key Secret", func() {
				By("Creating AWS accessKeyRequest...")
				r, err := f.CSClient.EngineV1alpha1().AWSAccessKeyRequests(awsAKR.Namespace).Create(context.TODO(), &awsAKR, metav1.CreateOptions{})
				Expect(err).NotTo(HaveOccurred(), "Create AWSAccessKeyRequest")

				IsAWSAccessKeyRequestCreated(awsAKR.Name, awsAKR.Namespace)

				By("Updating AWS AccessKeyRequest status...")
				err = f.UpdateAWSAccessKeyRequestStatus(&api.AWSAccessKeyRequestStatus{
					Conditions: []kmapi.Condition{
						{
							Type:               kmapi.ConditionRequestApproved,
							Status:             core.ConditionTrue,
							LastTransitionTime: metav1.Now(),
						},
					},
					Phase: api.RequestStatusPhaseApproved,
				}, r)

				Expect(err).NotTo(HaveOccurred(), "Update conditions: Approved")
				IsAWSAKRConditionApproved(awsAKR.Name, awsAKR.Namespace)

				IsAWSAccessKeySecretCreated(awsAKR.Name, awsAKR.Namespace)

				d, err := f.CSClient.EngineV1alpha1().AWSAccessKeyRequests(awsAKR.Namespace).Get(context.TODO(), awsAKR.Name, metav1.GetOptions{})
				Expect(err).NotTo(HaveOccurred(), "Get AWSAccessKeyRequest")
				if d.Status.Secret != nil {
					secretName = d.Status.Secret.Name
				}
			})
		})

	})
})
