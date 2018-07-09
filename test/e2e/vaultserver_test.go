package e2e_test

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"time"

	api "github.com/kubevault/operator/apis/core/v1alpha1"
	"github.com/kubevault/operator/pkg/controller"
	"github.com/kubevault/operator/pkg/vault/util"
	"github.com/kubevault/operator/test/e2e/framework"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
)

const (
	timeOut         = 20 * time.Minute
	pollingInterval = 10 * time.Second
)

var _ = Describe("VaultServer", func() {
	var (
		f *framework.Invocation
		// vs *api.VaultServer
	)

	BeforeEach(func() {
		f = root.Invoke()
	})
	AfterEach(func() {
		time.Sleep(30 * time.Second)
	})

	var (
		backendInmem = api.BackendStorageSpec{
			Inmem: true,
		}
	)

	var (
		checkForVaultTLSSecretCreated = func(name, namespace string) {
			By(fmt.Sprintf("Waiting for vault tls secret (%s/%s) to create", namespace, name))
			Eventually(func() bool {
				sr, err := f.KubeClient.CoreV1().Secrets(namespace).Get(name, metav1.GetOptions{})
				if err == nil {
					if _, ok := sr.Data[controller.CaCertName]; !ok {
						return false
					}
					if _, ok := sr.Data[controller.ServerCertName]; !ok {
						return false
					}
					if _, ok := sr.Data[controller.ServerkeyName]; !ok {
						return false
					}
					return true
				}
				return false
			}, timeOut, pollingInterval).Should(BeTrue())
		}

		checkForSecretCreated = func(name, namespace string) {
			By(fmt.Sprintf("Waiting for vault tls secret (%s/%s) to create", namespace, name))
			Eventually(func() bool {
				_, err := f.KubeClient.CoreV1().Secrets(namespace).Get(name, metav1.GetOptions{})
				if err == nil {
					return true
				}
				return false
			}, timeOut, pollingInterval).Should(BeTrue())
		}

		checkForSecretDeleted = func(name, namespace string) {
			By(fmt.Sprintf("Waiting for secret (%s/%s) to delete", namespace, name))
			Eventually(func() bool {
				_, err := f.KubeClient.CoreV1().Secrets(namespace).Get(name, metav1.GetOptions{})
				return kerrors.IsNotFound(err)
			}, timeOut, pollingInterval).Should(BeTrue())
		}

		checkForVaultConfigMapCreated = func(name, namespace string) {
			By(fmt.Sprintf("Waiting for vault configMap (%s/%s) to create", namespace, name))
			Eventually(func() bool {
				cm, err := f.KubeClient.CoreV1().ConfigMaps(namespace).Get(name, metav1.GetOptions{})
				if err == nil {
					if _, ok := cm.Data[filepath.Base(util.VaultConfigFile)]; !ok {
						return false
					}
					return true
				}
				return false
			}, timeOut, pollingInterval).Should(BeTrue())
		}

		checkForVaultConfigMapDeleted = func(name, namespace string) {
			By(fmt.Sprintf("Waiting for vault configMap (%s/%s) to delete", namespace, name))
			Eventually(func() bool {
				_, err := f.KubeClient.CoreV1().ConfigMaps(namespace).Get(name, metav1.GetOptions{})
				return kerrors.IsNotFound(err)
			}, timeOut, pollingInterval).Should(BeTrue())
		}

		checkForVaultDeploymentCreatedOrUpdated = func(name, namespace string, vs *api.VaultServer) {
			By(fmt.Sprintf("Waiting for vault deployment (%s/%s) to create/update", namespace, name))
			Eventually(func() bool {
				d, err := f.KubeClient.AppsV1beta1().Deployments(namespace).Get(name, metav1.GetOptions{})
				if err == nil {
					return *d.Spec.Replicas == vs.Spec.Nodes
				}
				return false
			}, timeOut, pollingInterval).Should(BeTrue())
		}

		checkForVaultDeploymentDeleted = func(name, namespace string) {
			By(fmt.Sprintf("Waiting for vault deployment (%s/%s) to delete", namespace, name))
			Eventually(func() bool {
				_, err := f.KubeClient.AppsV1beta1().Deployments(namespace).Get(name, metav1.GetOptions{})
				return kerrors.IsNotFound(err)
			}, timeOut, pollingInterval).Should(BeTrue())
		}

		checkForVaultServerCreated = func(name, namespace string) {
			By(fmt.Sprintf("Waiting for vault server (%s/%s) to create", namespace, name))
			Eventually(func() bool {
				_, err := f.VaultServerClient.CoreV1alpha1().VaultServers(namespace).Get(name, metav1.GetOptions{})
				return err == nil
			}, timeOut, pollingInterval).Should(BeTrue())
		}

		checkForVaultServerDeleted = func(name, namespace string) {
			By(fmt.Sprintf("Waiting for vault server (%s/%s) to delete", namespace, name))
			Eventually(func() bool {
				_, err := f.VaultServerClient.CoreV1alpha1().VaultServers(namespace).Get(name, metav1.GetOptions{})
				return kerrors.IsNotFound(err)
			}, timeOut, pollingInterval).Should(BeTrue())
		}

		shouldCreateVaultServer = func(vs *api.VaultServer) {
			By("Creating vault server")
			_, err := f.CreateVaultServer(vs)
			Expect(err).NotTo(HaveOccurred())

			checkForVaultServerCreated(vs.Name, vs.Namespace)
			checkForVaultTLSSecretCreated(controller.VaultTlsSecretName, vs.Namespace)
			checkForVaultConfigMapCreated(util.ConfigMapNameForVault(vs), vs.Namespace)
			checkForVaultDeploymentCreatedOrUpdated(vs.Name, vs.Namespace, vs)
		}

		shouldUpdateVaultServerReplica = func(replicas int32, vs *api.VaultServer) {
			By("Getting current vault server")
			vs, err := f.GetVaultServer(vs)
			Expect(err).NotTo(HaveOccurred())

			By("Updating replica number")
			vs.Spec.Nodes = replicas
			_, err = f.UpdateVaultServer(vs)
			Expect(err).NotTo(HaveOccurred())

			checkForVaultDeploymentCreatedOrUpdated(vs.Name, vs.Namespace, vs)
		}

		checkForVaultIsUnsealed = func(vs *api.VaultServer) {
			By("Checking whether vault is unsealed")
			Eventually(func() bool {
				v, err := f.VaultServerClient.CoreV1alpha1().VaultServers(vs.Namespace).Get(vs.Name, metav1.GetOptions{})
				if err == nil {
					if len(v.Status.VaultStatus.Unsealed) == int(vs.Spec.Nodes) {
						By(fmt.Sprintf("Unseal-pods: %v", v.Status.VaultStatus.Unsealed))
						return true
					}
				}
				return false
			}, timeOut, pollingInterval).Should(BeTrue())
		}
	)

	Describe("Creating vault server for", func() {
		var (
			vs *api.VaultServer
		)

		Context("inmem backend", func() {
			BeforeEach(func() {
				vs = f.VaultServer(3, backendInmem)
			})

			AfterEach(func() {
				err := f.DeleteVaultServer(vs.ObjectMeta)
				Expect(err).NotTo(HaveOccurred())

				checkForVaultServerDeleted(vs.Name, vs.Namespace)
				checkForSecretDeleted(controller.VaultTlsSecretName, vs.Namespace)
				checkForVaultConfigMapDeleted(util.ConfigMapNameForVault(vs), vs.Namespace)
				checkForVaultDeploymentDeleted(vs.Name, vs.Namespace)
			})

			It("should create vault server", func() {
				shouldCreateVaultServer(vs)
			})
		})

		Context("etcd backend", func() {

			BeforeEach(func() {
				url, err := f.DeployEtcd()
				Expect(err).NotTo(HaveOccurred())

				etcd := api.BackendStorageSpec{
					Etcd: &api.EtcdSpec{
						EtcdApi: "v3",
						Address: url,
					},
				}

				vs = f.VaultServer(3, etcd)
			})

			AfterEach(func() {
				err := f.DeleteEtcd()
				Expect(err).NotTo(HaveOccurred())

				err = f.DeleteVaultServer(vs.ObjectMeta)
				Expect(err).NotTo(HaveOccurred())

				checkForVaultServerDeleted(vs.Name, vs.Namespace)
				checkForSecretDeleted(controller.VaultTlsSecretName, vs.Namespace)
				checkForVaultConfigMapDeleted(util.ConfigMapNameForVault(vs), vs.Namespace)
				checkForVaultDeploymentDeleted(vs.Name, vs.Namespace)
			})

			It("should create vault server", func() {
				shouldCreateVaultServer(vs)
			})
		})

		Context("gcs backend", func() {
			const (
				secretName = "google-cred"
			)
			BeforeEach(func() {
				credFile := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
				data, err := ioutil.ReadFile(credFile)
				Expect(err).NotTo(HaveOccurred())

				sr := corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      secretName,
						Namespace: f.Namespace(),
					},
					Data: map[string][]byte{
						"sa.json": data,
					},
				}

				Expect(f.CreateSecret(sr)).NotTo(HaveOccurred())

				gcs := api.BackendStorageSpec{
					Gcs: &api.GcsSpec{
						Bucket:           "vault-test-bucket",
						CredentialSecret: secretName,
					},
				}

				vs = f.VaultServer(3, gcs)
			})

			AfterEach(func() {
				Expect(f.DeleteSecret(secretName, vs.Namespace)).NotTo(HaveOccurred())
				checkForSecretDeleted(secretName, vs.Namespace)

				Expect(f.DeleteVaultServer(vs.ObjectMeta)).NotTo(HaveOccurred())

				checkForVaultServerDeleted(vs.Name, vs.Namespace)
				checkForSecretDeleted(controller.VaultTlsSecretName, vs.Namespace)
				checkForVaultConfigMapDeleted(util.ConfigMapNameForVault(vs), vs.Namespace)
				checkForVaultDeploymentDeleted(vs.Name, vs.Namespace)
			})

			It("should create vault server", func() {
				shouldCreateVaultServer(vs)
			})
		})

		Context("s3 backend", func() {
			const (
				awsCredSecret = "test-aws-cred"
			)

			BeforeEach(func() {
				s3 := api.BackendStorageSpec{
					S3: &api.S3Spec{
						Bucket:           "test-vault-s3",
						Region:           "us-wes-1",
						CredentialSecret: awsCredSecret,
					},
				}

				vs = f.VaultServer(1, s3)

				sr := corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      awsCredSecret,
						Namespace: vs.Namespace,
					},
					Data: map[string][]byte{
						"access_key": []byte(os.Getenv("AWS_ACCESS_KEY_ID")),
						"secret_key": []byte(os.Getenv("AWS_SECRET_ACCESS_KEY")),
					},
				}

				Expect(f.CreateSecret(sr)).NotTo(HaveOccurred())
			})

			AfterEach(func() {
				Expect(f.DeleteSecret(awsCredSecret, vs.Namespace)).NotTo(HaveOccurred())

				checkForSecretDeleted(awsCredSecret, vs.Namespace)

				Expect(f.DeleteVaultServer(vs.ObjectMeta)).NotTo(HaveOccurred())

				checkForVaultServerDeleted(vs.Name, vs.Namespace)
				checkForSecretDeleted(controller.VaultTlsSecretName, vs.Namespace)
				checkForVaultConfigMapDeleted(util.ConfigMapNameForVault(vs), vs.Namespace)
				checkForVaultDeploymentDeleted(vs.Name, vs.Namespace)
			})

			It("should create vault server", func() {
				shouldCreateVaultServer(vs)
			})
		})
	})

	Describe("Updating vault server replica for", func() {
		Context("inmem backend", func() {
			var (
				vs *api.VaultServer
			)

			BeforeEach(func() {
				vs = f.VaultServer(3, backendInmem)
				shouldCreateVaultServer(vs)
			})
			AfterEach(func() {
				f.DeleteVaultServer(vs.ObjectMeta)

				checkForVaultServerDeleted(vs.Name, vs.Namespace)
				checkForSecretDeleted(controller.VaultTlsSecretName, vs.Namespace)
				checkForVaultConfigMapDeleted(util.ConfigMapNameForVault(vs), vs.Namespace)
				checkForVaultDeploymentDeleted(vs.Name, vs.Namespace)
			})

			It("should update replica number to 1", func() {
				shouldUpdateVaultServerReplica(1, vs)
			})

			It("should update replica number to 5", func() {
				shouldUpdateVaultServerReplica(4, vs)
			})
		})
	})

	Describe("Deleting vault resources", func() {
		Context("using inmem as backend", func() {
			var (
				err error
				vs  *api.VaultServer
			)

			BeforeEach(func() {
				vs = f.VaultServer(3, backendInmem)
				shouldCreateVaultServer(vs)
			})
			AfterEach(func() {
				f.DeleteVaultServer(vs.ObjectMeta)

				checkForVaultServerDeleted(vs.Name, vs.Namespace)
				checkForSecretDeleted(controller.VaultTlsSecretName, vs.Namespace)
				checkForVaultConfigMapDeleted(util.ConfigMapNameForVault(vs), vs.Namespace)
				checkForVaultDeploymentDeleted(vs.Name, vs.Namespace)
			})

			It("should keep the number of pods same as specification, after deleting some pods", func() {
				Eventually(func() bool {
					vs, err = f.VaultServerClient.CoreV1alpha1().VaultServers(vs.Namespace).Get(vs.Name, metav1.GetOptions{})
					if kerrors.IsNotFound(err) {
						return false
					} else {
						return len(vs.Status.UpdatedNodes) == int(vs.Spec.Nodes)
					}
				}, timeOut, pollingInterval).Should(BeTrue())

				p := rand.Int() % int(vs.Spec.Nodes)

				err = f.DeletePod(vs.Status.UpdatedNodes[p], vs.Namespace)
				Expect(err).NotTo(HaveOccurred())

				Eventually(func() bool {
					pods, err := f.KubeClient.CoreV1().Pods(vs.Namespace).List(metav1.ListOptions{
						LabelSelector: labels.SelectorFromSet(util.LabelsForVault(vs.Name)).String(),
					})
					if kerrors.IsNotFound(err) {
						return false
					} else {
						return len(pods.Items) == int(vs.Spec.Nodes)
					}
				}, timeOut, pollingInterval).Should(BeTrue())
			})

		})
	})

	Describe("Vault status monitor", func() {
		Context("using inmem as backend", func() {
			var (
				err error
				vs  *api.VaultServer
			)

			BeforeEach(func() {
				vs = f.VaultServer(3, backendInmem)
				shouldCreateVaultServer(vs)
			})
			AfterEach(func() {
				f.DeleteVaultServer(vs.ObjectMeta)

				checkForVaultServerDeleted(vs.Name, vs.Namespace)
				checkForSecretDeleted(controller.VaultTlsSecretName, vs.Namespace)
				checkForVaultConfigMapDeleted(util.ConfigMapNameForVault(vs), vs.Namespace)
				checkForVaultDeploymentDeleted(vs.Name, vs.Namespace)
			})

			It("status should contain 3 updated pods and 3 sealed pods", func() {
				Eventually(func() bool {
					vs, err = f.VaultServerClient.CoreV1alpha1().VaultServers(vs.Namespace).Get(vs.Name, metav1.GetOptions{})
					if kerrors.IsNotFound(err) {
						return false
					} else {
						return !vs.Status.Initialized &&
							len(vs.Status.UpdatedNodes) == 3 &&
							len(vs.Status.VaultStatus.Sealed) == 3
					}
				}, timeOut, pollingInterval).Should(BeTrue())
			})
		})
	})

	Describe("Vault unsealer using kubernetes secret", func() {
		var (
			vs     *api.VaultServer
			unseal *api.UnsealerSpec
		)

		const (
			secretName = "test-vault-keys"
		)

		BeforeEach(func() {
			unseal = &api.UnsealerSpec{
				SecretShares:    4,
				SecretThreshold: 2,
				InsecureTLS:     true,
				Mode: api.ModeSpec{
					KubernetesSecret: &api.KubernetesSecretSpec{
						SecretName: secretName,
					},
				},
			}
		})

		AfterEach(func() {
			f.DeleteVaultServer(vs.ObjectMeta)
			err := f.DeleteSecret(secretName, vs.Namespace)
			Expect(err).NotTo(HaveOccurred())

			checkForVaultServerDeleted(vs.Name, vs.Namespace)
			checkForSecretDeleted(controller.VaultTlsSecretName, vs.Namespace)
			checkForVaultConfigMapDeleted(util.ConfigMapNameForVault(vs), vs.Namespace)
			checkForVaultDeploymentDeleted(vs.Name, vs.Namespace)
		})

		Context("using inmem backend", func() {
			BeforeEach(func() {
				vs = f.VaultServerWithUnsealer(1, backendInmem, *unseal)
			})

			It("vault should be unsealed", func() {
				shouldCreateVaultServer(vs)

				checkForSecretCreated(secretName, vs.Namespace)
				checkForVaultIsUnsealed(vs)
			})
		})

		Context("using etcd backend", func() {
			BeforeEach(func() {
				url, err := f.DeployEtcd()
				Expect(err).NotTo(HaveOccurred())

				etcd := api.BackendStorageSpec{
					Etcd: &api.EtcdSpec{
						EtcdApi: "v3",
						Address: url,
					},
				}

				vs = f.VaultServerWithUnsealer(1, etcd, *unseal)
			})

			AfterEach(func() {
				err := f.DeleteEtcd()
				Expect(err).NotTo(HaveOccurred())
			})

			It("vault should be unsealed", func() {
				shouldCreateVaultServer(vs)

				checkForSecretCreated(secretName, vs.Namespace)
				checkForVaultIsUnsealed(vs)
			})
		})
	})

	Describe("unsealing using google kms gcs", func() {
		var (
			vs *api.VaultServer
		)

		Context("using gcs backend", func() {
			const (
				secretName = "google-cred"
			)
			BeforeEach(func() {
				credFile := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
				data, err := ioutil.ReadFile(credFile)
				Expect(err).NotTo(HaveOccurred())

				sr := corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      secretName,
						Namespace: f.Namespace(),
					},
					Data: map[string][]byte{
						"sa.json": data,
					},
				}

				Expect(f.CreateSecret(sr)).NotTo(HaveOccurred())
				gcs := api.BackendStorageSpec{
					Gcs: &api.GcsSpec{
						Bucket:           "vault-test-bucket",
						CredentialSecret: secretName,
					},
				}

				unsealer := api.UnsealerSpec{
					SecretShares:    4,
					SecretThreshold: 2,
					InsecureTLS:     true,
					Mode: api.ModeSpec{
						GoogleKmsGcs: &api.GoogleKmsGcsSpec{
							Bucket:           "vault-test-bucket",
							CredentialSecret: secretName,
							KmsCryptoKey:     "vault-init",
							KmsKeyRing:       "vault-key-ring",
							KmsLocation:      "global",
							KmsProject:       "tigerworks-kube",
						},
					},
				}

				vs = f.VaultServerWithUnsealer(1, gcs, unsealer)
			})

			AfterEach(func() {
				Expect(f.DeleteSecret(secretName, vs.Namespace)).NotTo(HaveOccurred())
				checkForSecretDeleted(secretName, vs.Namespace)

				Expect(f.DeleteVaultServer(vs.ObjectMeta)).NotTo(HaveOccurred())

				checkForVaultServerDeleted(vs.Name, vs.Namespace)
				checkForSecretDeleted(controller.VaultTlsSecretName, vs.Namespace)
				checkForVaultConfigMapDeleted(util.ConfigMapNameForVault(vs), vs.Namespace)
				checkForVaultDeploymentDeleted(vs.Name, vs.Namespace)
			})

			It("should create vault server", func() {
				shouldCreateVaultServer(vs)

				checkForVaultIsUnsealed(vs)
			})
		})
	})

	Describe("unsealing using aws kms ssm", func() {
		var (
			vs *api.VaultServer
		)

		const (
			awsCredSecret = "test-aws-cred"
		)

		Context("using s3 backend", func() {
			BeforeEach(func() {
				s3 := api.BackendStorageSpec{
					S3: &api.S3Spec{
						Bucket:           "test-vault-s3",
						Region:           "us-west-1",
						CredentialSecret: awsCredSecret,
					},
				}

				unsealer := api.UnsealerSpec{
					SecretShares:    4,
					SecretThreshold: 2,
					InsecureTLS:     true,
					Mode: api.ModeSpec{
						AwsKmsSsm: &api.AwsKmsSsmSpec{
							KmsKeyID:         "65ed2c85-4915-4e82-be47-d56ccaa8019b",
							Region:           "us-west-1",
							CredentialSecret: awsCredSecret,
						},
					},
				}

				vs = f.VaultServerWithUnsealer(1, s3, unsealer)

				sr := corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      awsCredSecret,
						Namespace: vs.Namespace,
					},
					Data: map[string][]byte{
						"access_key": []byte(os.Getenv("AWS_ACCESS_KEY_ID")),
						"secret_key": []byte(os.Getenv("AWS_SECRET_ACCESS_KEY")),
					},
				}

				Expect(f.CreateSecret(sr)).NotTo(HaveOccurred())
			})

			AfterEach(func() {
				Expect(f.DeleteSecret(awsCredSecret, vs.Namespace)).NotTo(HaveOccurred())
				checkForSecretDeleted(awsCredSecret, vs.Namespace)

				Expect(f.DeleteVaultServer(vs.ObjectMeta)).NotTo(HaveOccurred())

				checkForVaultServerDeleted(vs.Name, vs.Namespace)
				checkForSecretDeleted(controller.VaultTlsSecretName, vs.Namespace)
				checkForVaultConfigMapDeleted(util.ConfigMapNameForVault(vs), vs.Namespace)
				checkForVaultDeploymentDeleted(vs.Name, vs.Namespace)
			})

			It("should create vault server", func() {
				shouldCreateVaultServer(vs)

				checkForVaultIsUnsealed(vs)
			})
		})
	})

	Describe("unsealing using azure key vault", func() {
		var (
			vs *api.VaultServer
		)

		const (
			azureCredSecret = "test-azure-cred"
		)

		Context("using azure storage backend", func() {
			BeforeEach(func() {
				var (
					clientID     = os.Getenv("AZURE_CLIENT_ID")
					clientSecret = os.Getenv("AZURE_CLIENT_SECRET")
					tenantID     = os.Getenv("AZURE_TENANT_ID")
					accountName  = os.Getenv("AZURE_STORAGE_ACCOUNT_NAME")
					accountKey   = os.Getenv("AZURE_STORAGE_ACCOUNT_KEY")
				)

				Expect(clientID != "").To(BeTrue())
				Expect(clientSecret != "").To(BeTrue())
				Expect(tenantID != "").To(BeTrue())
				Expect(accountName != "").To(BeTrue())
				Expect(accountKey != "").To(BeTrue())

				sr := corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      azureCredSecret,
						Namespace: f.Namespace(),
					},
					Data: map[string][]byte{
						"client-id":     []byte(clientID),
						"client-secret": []byte(clientSecret),
					},
				}

				Expect(f.CreateSecret(sr)).NotTo(HaveOccurred())

				azure := api.BackendStorageSpec{
					Azure: &api.AzureSpec{
						AccountName: accountName,
						AccountKey:  accountKey,
						Container:   "vault",
					},
				}

				unsealer := api.UnsealerSpec{
					SecretShares:    4,
					SecretThreshold: 2,
					InsecureTLS:     true,
					Mode: api.ModeSpec{
						AzureKeyVault: &api.AzureKeyVault{
							VaultBaseUrl:    "https://vault-test-1204.vault.azure.net/",
							TenantID:        tenantID,
							AADClientSecret: azureCredSecret,
						},
					},
				}

				vs = f.VaultServerWithUnsealer(1, azure, unsealer)
			})

			AfterEach(func() {
				Expect(f.DeleteSecret(azureCredSecret, vs.Namespace)).NotTo(HaveOccurred())
				checkForSecretDeleted(azureCredSecret, vs.Namespace)

				Expect(f.DeleteVaultServer(vs.ObjectMeta)).NotTo(HaveOccurred())

				checkForVaultServerDeleted(vs.Name, vs.Namespace)
				checkForSecretDeleted(controller.VaultTlsSecretName, vs.Namespace)
				checkForVaultConfigMapDeleted(util.ConfigMapNameForVault(vs), vs.Namespace)
				checkForVaultDeploymentDeleted(vs.Name, vs.Namespace)
			})

			It("should create vault server", func() {
				shouldCreateVaultServer(vs)

				checkForVaultIsUnsealed(vs)
			})
		})
	})

	Describe("using postgerSQL backend", func() {
		Context("using unsealer kubernetes secret", func() {
			var (
				vs *api.VaultServer
			)

			const (
				k8sSecretName       = "k8s-postgres-vault-keys"
				connectionUrlSecret = "postgresql-conn-url"
			)
			BeforeEach(func() {
				url, err := f.DeployPostgresSQL()
				Expect(err).NotTo(HaveOccurred())

				sr := corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      connectionUrlSecret,
						Namespace: f.Namespace(),
					},
					Data: map[string][]byte{
						"connection_url": []byte(fmt.Sprintf("postgres://postgres:root@%s/database?sslmode=disable", url)),
					},
				}

				Expect(f.CreateSecret(sr)).NotTo(HaveOccurred())

				postgres := api.BackendStorageSpec{
					PostgreSQL: &api.PostgreSQLSpec{
						ConnectionUrlSecret: connectionUrlSecret,
					},
				}

				unsealer := api.UnsealerSpec{
					SecretShares:    4,
					SecretThreshold: 2,
					InsecureTLS:     true,
					Mode: api.ModeSpec{
						KubernetesSecret: &api.KubernetesSecretSpec{
							SecretName: k8sSecretName,
						},
					},
				}

				vs = f.VaultServerWithUnsealer(1, postgres, unsealer)
			})

			AfterEach(func() {

				Expect(f.DeletePostgresSQL()).NotTo(HaveOccurred())

				Expect(f.DeleteSecret(k8sSecretName, vs.Namespace)).NotTo(HaveOccurred())
				checkForSecretDeleted(k8sSecretName, vs.Namespace)

				Expect(f.DeleteSecret(connectionUrlSecret, vs.Namespace)).NotTo(HaveOccurred())
				checkForSecretDeleted(connectionUrlSecret, vs.Namespace)

				Expect(f.DeleteVaultServer(vs.ObjectMeta)).NotTo(HaveOccurred())

				checkForVaultServerDeleted(vs.Name, vs.Namespace)
				checkForSecretDeleted(controller.VaultTlsSecretName, vs.Namespace)
				checkForVaultConfigMapDeleted(util.ConfigMapNameForVault(vs), vs.Namespace)
				checkForVaultDeploymentDeleted(vs.Name, vs.Namespace)
			})

			It("should create vault server", func() {
				shouldCreateVaultServer(vs)

				checkForVaultIsUnsealed(vs)
			})
		})
	})

	Describe("using mySQL backend", func() {
		Context("using unsealer kubernetes secret", func() {
			var (
				vs *api.VaultServer
			)

			const (
				k8sSecretName   = "k8s-mysql-vault-keys"
				mysqlCredSecret = "mysql-cred-1234"
			)
			BeforeEach(func() {
				url, err := f.DeployMySQL()
				Expect(err).NotTo(HaveOccurred())

				sr := corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      mysqlCredSecret,
						Namespace: f.Namespace(),
					},
					Data: map[string][]byte{
						"username": []byte("root"),
						"password": []byte("root"),
					},
				}

				Expect(f.CreateSecret(sr)).NotTo(HaveOccurred())

				mysql := api.BackendStorageSpec{
					MySQL: &api.MySQLSpec{
						Address:              url,
						UserCredentialSecret: mysqlCredSecret,
					},
				}

				unsealer := api.UnsealerSpec{
					SecretShares:    4,
					SecretThreshold: 2,
					InsecureTLS:     true,
					Mode: api.ModeSpec{
						KubernetesSecret: &api.KubernetesSecretSpec{
							SecretName: k8sSecretName,
						},
					},
				}

				vs = f.VaultServerWithUnsealer(1, mysql, unsealer)
			})

			AfterEach(func() {

				Expect(f.DeleteMySQL()).NotTo(HaveOccurred())

				Expect(f.DeleteSecret(k8sSecretName, vs.Namespace)).NotTo(HaveOccurred())
				checkForSecretDeleted(k8sSecretName, vs.Namespace)
				Expect(f.DeleteSecret(mysqlCredSecret, vs.Namespace)).NotTo(HaveOccurred())
				checkForSecretDeleted(mysqlCredSecret, vs.Namespace)

				Expect(f.DeleteVaultServer(vs.ObjectMeta)).NotTo(HaveOccurred())

				checkForVaultServerDeleted(vs.Name, vs.Namespace)
				checkForSecretDeleted(controller.VaultTlsSecretName, vs.Namespace)
				checkForVaultConfigMapDeleted(util.ConfigMapNameForVault(vs), vs.Namespace)
				checkForVaultDeploymentDeleted(vs.Name, vs.Namespace)
			})

			It("should create vault server", func() {
				shouldCreateVaultServer(vs)

				checkForVaultIsUnsealed(vs)
			})
		})
	})

	Describe("using File system backend", func() {
		Context("using unsealer kubernetes secret", func() {
			var (
				vs *api.VaultServer
			)

			const (
				k8sSecretName = "k8s-file-vault-keys"
			)
			BeforeEach(func() {
				file := api.BackendStorageSpec{
					File: &api.FileSpec{
						Path: "/etc/data",
					},
				}

				unsealer := api.UnsealerSpec{
					SecretShares:    4,
					SecretThreshold: 2,
					InsecureTLS:     true,
					Mode: api.ModeSpec{
						KubernetesSecret: &api.KubernetesSecretSpec{
							SecretName: k8sSecretName,
						},
					},
				}

				vs = f.VaultServerWithUnsealer(1, file, unsealer)
			})

			AfterEach(func() {
				Expect(f.DeleteSecret(k8sSecretName, vs.Namespace)).NotTo(HaveOccurred())
				checkForSecretDeleted(k8sSecretName, vs.Namespace)

				Expect(f.DeleteVaultServer(vs.ObjectMeta)).NotTo(HaveOccurred())

				checkForVaultServerDeleted(vs.Name, vs.Namespace)
				checkForSecretDeleted(controller.VaultTlsSecretName, vs.Namespace)
				checkForVaultConfigMapDeleted(util.ConfigMapNameForVault(vs), vs.Namespace)
				checkForVaultDeploymentDeleted(vs.Name, vs.Namespace)
			})

			It("should create vault server", func() {
				shouldCreateVaultServer(vs)

				checkForVaultIsUnsealed(vs)
			})
		})
	})

	Describe("unsealing using kubernetes secret", func() {
		var (
			vs *api.VaultServer
		)

		const (
			awsCredSecret = "test-aws-cred-123"
			region =  "us-west-1"
			table = "vault-dynamodb-test-1234"
			readCapacity = 5
			writeCapacity = 5
		)

		Context("using dynamoDB backend", func() {
			BeforeEach(func() {
				if !f.RunDynamoDBTest {
					Skip("dynamodb test is skipped")
				}

				sr := corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      awsCredSecret,
						Namespace: f.Namespace(),
					},
					Data: map[string][]byte{
						"access_key": []byte(os.Getenv("AWS_ACCESS_KEY_ID")),
						"secret_key": []byte(os.Getenv("AWS_SECRET_ACCESS_KEY")),
					},
				}

				Expect(f.CreateSecret(sr)).NotTo(HaveOccurred())

				Expect(f.DynamoDBCreateTable(region, table, readCapacity,writeCapacity)).NotTo(HaveOccurred())

				dynamodb := api.BackendStorageSpec{
					DynamoDB: &api.DynamoDBSpec{
						Region: region,
						CredentialSecret: awsCredSecret,
						Table: table,
						ReadCapacity: readCapacity,
						WriteCapacity: writeCapacity,
					},
				}

				unsealer := api.UnsealerSpec{
					SecretShares:    4,
					SecretThreshold: 2,
					InsecureTLS:     true,
					Mode: api.ModeSpec{
						KubernetesSecret: &api.KubernetesSecretSpec{
							SecretName: "k8s-dynamodb-keys-1234",
						},
					},
				}

				vs = f.VaultServerWithUnsealer(1,dynamodb, unsealer)
			})

			AfterEach(func() {
				Expect(f.DynamoDBDeleteTable(region, table)).NotTo(HaveOccurred())

				Expect(f.DeleteSecret(awsCredSecret, vs.Namespace)).NotTo(HaveOccurred())
				checkForSecretDeleted(awsCredSecret, vs.Namespace)

				Expect(f.DeleteVaultServer(vs.ObjectMeta)).NotTo(HaveOccurred())

				checkForVaultServerDeleted(vs.Name, vs.Namespace)
				checkForSecretDeleted(controller.VaultTlsSecretName, vs.Namespace)
				checkForVaultConfigMapDeleted(util.ConfigMapNameForVault(vs), vs.Namespace)
				checkForVaultDeploymentDeleted(vs.Name, vs.Namespace)
			})

			It("should create vault server", func() {
				shouldCreateVaultServer(vs)

				checkForVaultIsUnsealed(vs)
			})
		})
	})
})
