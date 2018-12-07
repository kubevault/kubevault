package e2e_test

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"time"

	rand_util "github.com/appscode/go/crypto/rand"
	"github.com/golang/glog"
	cleanhttp "github.com/hashicorp/go-cleanhttp"
	vaultapi "github.com/hashicorp/vault/api"
	api "github.com/kubevault/operator/apis/kubevault/v1alpha1"
	"github.com/kubevault/operator/pkg/controller"
	"github.com/kubevault/operator/pkg/vault/util"
	"github.com/kubevault/operator/test/e2e/framework"
	"github.com/ncw/swift"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	core "k8s.io/api/core/v1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	ofst "kmodules.xyz/offshoot-api/api/v1"
)

const (
	timeOut         = 5 * time.Minute
	pollingInterval = 10 * time.Second
)

var _ = Describe("VaultServer", func() {
	var (
		f *framework.Invocation
	)

	BeforeEach(func() {
		f = root.Invoke()
	})
	AfterEach(func() {
		time.Sleep(10 * time.Second)
	})

	var (
		backendInmem = api.BackendStorageSpec{
			Inmem: &api.InmemSpec{},
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
			}, timeOut, pollingInterval).Should(BeTrue(), fmt.Sprintf("vault tls secret (%s/%s) should exists", namespace, name))
		}

		checkForSecretCreated = func(name, namespace string) {
			By(fmt.Sprintf("Waiting for vault secret (%s/%s) to create", namespace, name))
			Eventually(func() bool {
				_, err := f.KubeClient.CoreV1().Secrets(namespace).Get(name, metav1.GetOptions{})
				if err == nil {
					return true
				}
				return false
			}, timeOut, pollingInterval).Should(BeTrue(), fmt.Sprintf("secret (%s/%s) should exists", namespace, name))
		}

		checkForSecretDeleted = func(name, namespace string) {
			By(fmt.Sprintf("Waiting for secret (%s/%s) to delete", namespace, name))
			Eventually(func() bool {
				_, err := f.KubeClient.CoreV1().Secrets(namespace).Get(name, metav1.GetOptions{})
				return kerr.IsNotFound(err)
			}, timeOut, pollingInterval).Should(BeTrue(), fmt.Sprintf("secret (%s/%s) should be deleted", namespace, name))
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
			}, timeOut, pollingInterval).Should(BeTrue(), fmt.Sprintf("configMap (%s/%s) should exists", namespace, name))
		}

		checkForVaultConfigMapDeleted = func(name, namespace string) {
			By(fmt.Sprintf("Waiting for vault configMap (%s/%s) to delete", namespace, name))
			Eventually(func() bool {
				_, err := f.KubeClient.CoreV1().ConfigMaps(namespace).Get(name, metav1.GetOptions{})
				return kerr.IsNotFound(err)
			}, timeOut, pollingInterval).Should(BeTrue(), fmt.Sprintf("configMap (%s/%s) should not exists", namespace, name))
		}

		checkForVaultDeploymentCreatedOrUpdated = func(name, namespace string, vs *api.VaultServer) {
			By(fmt.Sprintf("Waiting for vault deployment (%s/%s) to create/update", namespace, name))
			Eventually(func() bool {
				d, err := f.KubeClient.AppsV1().Deployments(namespace).Get(name, metav1.GetOptions{})
				if err == nil {
					return *d.Spec.Replicas == vs.Spec.Nodes
				}
				return false
			}, timeOut, pollingInterval).Should(BeTrue(), fmt.Sprintf("deployment (%s/%s) replicas should be equal to v.spec.nodes", namespace, name))
		}

		checkForVaultDeploymentDeleted = func(name, namespace string) {
			By(fmt.Sprintf("Waiting for vault deployment (%s/%s) to delete", namespace, name))
			Eventually(func() bool {
				_, err := f.KubeClient.AppsV1().Deployments(namespace).Get(name, metav1.GetOptions{})
				return kerr.IsNotFound(err)
			}, timeOut, pollingInterval).Should(BeTrue(), fmt.Sprintf("deployment (%s/%s) should not exists", namespace, name))
		}

		checkForVaultServerCreated = func(name, namespace string) {
			By(fmt.Sprintf("Waiting for vault server (%s/%s) to create", namespace, name))
			Eventually(func() bool {
				_, err := f.CSClient.KubevaultV1alpha1().VaultServers(namespace).Get(name, metav1.GetOptions{})
				return err == nil
			}, timeOut, pollingInterval).Should(BeTrue(), fmt.Sprintf("vaultserver (%s/%s) should exists", namespace, name))
		}

		checkForVaultServerDeleted = func(name, namespace string) {
			By(fmt.Sprintf("Waiting for vault server (%s/%s) to delete", namespace, name))
			Eventually(func() bool {
				_, err := f.CSClient.KubevaultV1alpha1().VaultServers(namespace).Get(name, metav1.GetOptions{})
				return kerr.IsNotFound(err)
			}, timeOut, pollingInterval).Should(BeTrue(), fmt.Sprintf("vaultserver (%s/%s) should not exists", namespace, name))
		}

		checkForAppBindingCreated = func(name, namespace string) {
			By(fmt.Sprintf("Waiting for AppBinding (%s/%s) to create", namespace, name))
			Eventually(func() bool {
				_, err := f.AppcatClient.AppBindings(namespace).Get(name, metav1.GetOptions{})
				return err == nil
			}, timeOut, pollingInterval).Should(BeTrue(), fmt.Sprintf("AppBinding (%s/%s) should exists", namespace, name))
		}

		checkForAppBindingDeleted = func(name, namespace string) {
			By(fmt.Sprintf("Waiting for AppBinding (%s/%s) to delete", namespace, name))
			Eventually(func() bool {
				_, err := f.AppcatClient.AppBindings(namespace).Get(name, metav1.GetOptions{})
				return kerr.IsNotFound(err)
			}, timeOut, pollingInterval).Should(BeTrue(), fmt.Sprintf("AppBinding (%s/%s) should not exists", namespace, name))
		}

		shouldCreateVaultServer = func(vs *api.VaultServer) {
			By("Creating vault server")
			_, err := f.CreateVaultServer(vs)
			Expect(err).NotTo(HaveOccurred())

			checkForVaultServerCreated(vs.Name, vs.Namespace)
			checkForVaultTLSSecretCreated(vs.TLSSecretName(), vs.Namespace)
			checkForVaultConfigMapCreated(vs.ConfigMapName(), vs.Namespace)
			checkForVaultDeploymentCreatedOrUpdated(vs.Name, vs.Namespace, vs)
			checkForAppBindingCreated(vs.Name, vs.Namespace)
			By("vault server created")
		}

		shouldUpdateVaultServerReplica = func(replicas int32, vs *api.VaultServer) {
			By("Getting current vault server")
			vs, err := f.GetVaultServer(vs)
			Expect(err).NotTo(HaveOccurred())

			By("Updating replica number")
			vs.Spec.Nodes = replicas
			vs, err = f.UpdateVaultServer(vs)
			Expect(err).NotTo(HaveOccurred())

			time.Sleep(3 * time.Second)
			By("Getting update vault server")
			vs, err = f.GetVaultServer(vs)
			Expect(err).NotTo(HaveOccurred())
			Expect(vs.Spec.Nodes == replicas).To(BeTrue(), "should match replicas")

			checkForVaultDeploymentCreatedOrUpdated(vs.Name, vs.Namespace, vs)
		}

		checkForVaultIsUnsealed = func(vs *api.VaultServer) {
			By("Checking whether vault is unsealed")
			Eventually(func() bool {
				v, err := f.CSClient.KubevaultV1alpha1().VaultServers(vs.Namespace).Get(vs.Name, metav1.GetOptions{})
				if err == nil {
					if len(v.Status.VaultStatus.Unsealed) == int(vs.Spec.Nodes) {
						By(fmt.Sprintf("Unseal-pods: %v", v.Status.VaultStatus.Unsealed))
						return true
					}
				}
				return false
			}, timeOut, pollingInterval).Should(BeTrue(), fmt.Sprintf("number of unseal pods should be equal to v.spec.nodes"))
		}

		checkForVaultServerCleanup = func(vs *api.VaultServer) {
			checkForVaultServerDeleted(vs.Name, vs.Namespace)
			checkForSecretDeleted(vs.TLSSecretName(), vs.Namespace)
			checkForVaultConfigMapDeleted(vs.ConfigMapName(), vs.Namespace)
			checkForVaultDeploymentDeleted(vs.Name, vs.Namespace)
			checkForAppBindingDeleted(vs.Name, vs.Namespace)
		}

		checkForIsVaultHAEnabled = func(vs *api.VaultServer) {
			By(fmt.Sprintf("Checking is HA enabled in vault(%s/%s)", vs.Namespace, vs.Name))
			Eventually(func() bool {
				v, err := f.GetVaultServer(vs)
				if err == nil {
					nodeIP, err := f.GetNodePortIP(v.OffshootSelectors())
					if err == nil {
						var url string
						srv, err := f.KubeClient.CoreV1().Services(v.Namespace).Get(v.OffshootName(), metav1.GetOptions{})
						if err == nil {
							for _, p := range srv.Spec.Ports {
								if p.Port == controller.VaultClientPort {
									url = fmt.Sprintf("https://%s:%d", nodeIP, p.NodePort)
									break
								}
							}
							if url != "" {
								By(fmt.Sprintf("vault url: %s", url))
								cfg := vaultapi.DefaultConfig()
								cfg.Address = url
								cfg.ConfigureTLS(&vaultapi.TLSConfig{
									Insecure: true,
								})
								vc, err := vaultapi.NewClient(cfg)
								if err == nil {
									status, err := vc.Sys().Leader()
									if err == nil {
										return status.HAEnabled
									} else {
										glog.Errorln(err)
									}
								} else {
									glog.Errorln(err)
								}
							}
						} else {
							glog.Errorln(err)
						}
					} else {
						glog.Errorln(err)
					}
				} else {
					glog.Errorln(err)
				}
				return false
			}, timeOut, pollingInterval).Should(BeTrue(), fmt.Sprintf("HA should be enabled in vault server (%s/%s)", vs.Namespace, vs.Name))

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
				checkForVaultServerCleanup(vs)
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
				checkForVaultServerCleanup(vs)
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

				sr := core.Secret{
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
				checkForVaultServerCleanup(vs)
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

				sr := core.Secret{
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
				checkForVaultServerCleanup(vs)
			})

			It("should create vault server", func() {
				shouldCreateVaultServer(vs)
			})
		})
	})

	Describe("Updating vault server replica for", func() {
		Context("inmem backend", func() {
			var (
				vs             *api.VaultServer
				vaultKeySecret string
			)

			BeforeEach(func() {
				vaultKeySecret = rand_util.WithUniqSuffix("v-key")
				unseal := api.UnsealerSpec{
					SecretShares:          4,
					SecretThreshold:       2,
					InsecureSkipTLSVerify: true,
					Mode: api.ModeSpec{
						KubernetesSecret: &api.KubernetesSecretSpec{
							SecretName: vaultKeySecret,
						},
					},
				}
				vs = f.VaultServerWithUnsealer(3, backendInmem, unseal)
				shouldCreateVaultServer(vs)
			})
			AfterEach(func() {
				f.DeleteVaultServer(vs.ObjectMeta)
				checkForVaultServerCleanup(vs)

				Expect(f.DeleteSecret(vaultKeySecret, vs.Namespace)).NotTo(HaveOccurred(), "delete vault key secret")
				checkForSecretDeleted(vaultKeySecret, vs.Namespace)
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
				vs             *api.VaultServer
				vaultKeySecret string
			)

			BeforeEach(func() {
				vaultKeySecret = rand_util.WithUniqSuffix("v-key")
				unsealer := api.UnsealerSpec{
					SecretShares:          4,
					SecretThreshold:       2,
					InsecureSkipTLSVerify: true,
					Mode: api.ModeSpec{
						KubernetesSecret: &api.KubernetesSecretSpec{
							SecretName: vaultKeySecret,
						},
					},
				}

				vs = f.VaultServerWithUnsealer(1, backendInmem, unsealer)

				shouldCreateVaultServer(vs)
			})
			AfterEach(func() {
				f.DeleteVaultServer(vs.ObjectMeta)
				checkForVaultServerCleanup(vs)

				f.DeleteSecret(vaultKeySecret, vs.Namespace)
				checkForSecretDeleted(vaultKeySecret, vs.Namespace)
			})

			It("should keep the number of pods same as specification, after deleting some pods", func() {

				pods, err := f.KubeClient.CoreV1().Pods(vs.Namespace).List(metav1.ListOptions{
					LabelSelector: labels.SelectorFromSet(vs.OffshootSelectors()).String(),
				})
				Expect(err).NotTo(HaveOccurred(), "list vault pods")

				Eventually(func() bool {
					pods, err := f.KubeClient.CoreV1().Pods(vs.Namespace).List(metav1.ListOptions{
						LabelSelector: labels.SelectorFromSet(vs.OffshootSelectors()).String(),
					})
					if kerr.IsNotFound(err) {
						return false
					} else {
						return len(pods.Items) == int(vs.Spec.Nodes)
					}
				}, timeOut, pollingInterval).Should(BeTrue(), "number of pods should be equal to v.spce.nodes")

				p := rand.Int() % int(len(pods.Items))

				err = f.DeletePod(pods.Items[p].Name, vs.Namespace)
				Expect(err).NotTo(HaveOccurred())
				Eventually(func() bool {
					pods, err := f.KubeClient.CoreV1().Pods(vs.Namespace).List(metav1.ListOptions{
						LabelSelector: labels.SelectorFromSet(vs.OffshootSelectors()).String(),
					})
					if kerr.IsNotFound(err) {
						return false
					} else {
						return len(pods.Items) == int(vs.Spec.Nodes)
					}
				}, timeOut, pollingInterval).Should(BeTrue(), "number of pods should be equal to v.spce.nodes")
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
				unsealer := api.UnsealerSpec{
					SecretShares:          4,
					SecretThreshold:       2,
					InsecureSkipTLSVerify: true,
					Mode: api.ModeSpec{
						KubernetesSecret: &api.KubernetesSecretSpec{
							SecretName: "k8s-inmem-keys-123411",
						},
					},
				}

				vs = f.VaultServerWithUnsealer(1, backendInmem, unsealer)
				shouldCreateVaultServer(vs)
			})
			AfterEach(func() {
				f.DeleteVaultServer(vs.ObjectMeta)
				checkForVaultServerCleanup(vs)

				f.DeleteSecret("k8s-inmem-keys-123411", vs.Namespace)
				checkForSecretDeleted("k8s-inmem-keys-123411", vs.Namespace)
			})

			It("status should contain 1 updated pods and 1 unseal pods", func() {
				Eventually(func() bool {
					vs, err = f.CSClient.KubevaultV1alpha1().VaultServers(vs.Namespace).Get(vs.Name, metav1.GetOptions{})
					if kerr.IsNotFound(err) {
						return false
					} else {
						return vs.Status.Initialized &&
							len(vs.Status.UpdatedNodes) == 1 &&
							len(vs.Status.VaultStatus.Unsealed) == 1
					}
				}, timeOut, pollingInterval).Should(BeTrue(), "status should contain 1 updated pods and 1 unseal pods")
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
				SecretShares:          4,
				SecretThreshold:       2,
				InsecureSkipTLSVerify: true,
				Mode: api.ModeSpec{
					KubernetesSecret: &api.KubernetesSecretSpec{
						SecretName: secretName,
					},
				},
			}
		})

		AfterEach(func() {
			f.DeleteVaultServer(vs.ObjectMeta)
			checkForVaultServerCleanup(vs)

			err := f.DeleteSecret(secretName, vs.Namespace)
			Expect(err).NotTo(HaveOccurred())
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

		Context("using swift backend", func() {
			const swiftCredSecret = "swift-user-cred"
			var (
				username  = os.Getenv("OS_USERNAME")
				password  = os.Getenv("OS_PASSWORD")
				authUrl   = os.Getenv("OS_AUTH_URL")
				region    = os.Getenv("OS_REGION_NAME")
				tenant    = os.Getenv("OS_TENANT_NAME")
				container = "vault-test"
			)

			BeforeEach(func() {
				if username == "" || password == "" || authUrl == "" || region == "" || tenant == "" {
					Skip("OS_USERNAME or OS_PASSWORD or OS_AUTH_URL or OS_REGION_NAME or OS_TENANT_NAME  are not provided")
				}

				cleaner := swift.Connection{
					UserName:  username,
					ApiKey:    password,
					AuthUrl:   authUrl,
					Tenant:    tenant,
					Region:    region,
					Transport: cleanhttp.DefaultPooledTransport(),
				}
				Expect(cleaner.Authenticate()).NotTo(HaveOccurred())

				// clean all the objects in swift storage
				newObjects, err := cleaner.ObjectNamesAll(container, nil)
				Expect(err).NotTo(HaveOccurred())
				for _, o := range newObjects {
					err := cleaner.ObjectDelete(container, o)
					Expect(err).NotTo(HaveOccurred())
				}

				sr := core.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      swiftCredSecret,
						Namespace: f.Namespace(),
					},
					Data: map[string][]byte{
						"username": []byte(username),
						"password": []byte(password),
					},
				}
				_, err = f.KubeClient.CoreV1().Secrets(sr.Namespace).Create(&sr)
				Expect(err).NotTo(HaveOccurred())

				swift := api.BackendStorageSpec{
					Swift: &api.SwiftSpec{
						AuthUrl:          authUrl,
						Container:        container,
						CredentialSecret: swiftCredSecret,
						Region:           region,
						Tenant:           tenant,
					},
				}
				vs = f.VaultServerWithUnsealer(1, swift, *unseal)
			})

			AfterEach(func() {
				err := f.DeleteSecret(swiftCredSecret, f.Namespace())
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

				sr := core.Secret{
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
					SecretShares:          4,
					SecretThreshold:       2,
					InsecureSkipTLSVerify: true,
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
				checkForVaultServerCleanup(vs)
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
					SecretShares:          4,
					SecretThreshold:       2,
					InsecureSkipTLSVerify: true,
					Mode: api.ModeSpec{
						AwsKmsSsm: &api.AwsKmsSsmSpec{
							KmsKeyID:         "65ed2c85-4915-4e82-be47-d56ccaa8019b",
							Region:           "us-west-1",
							CredentialSecret: awsCredSecret,
						},
					},
				}

				vs = f.VaultServerWithUnsealer(1, s3, unsealer)

				sr := core.Secret{
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
				checkForVaultServerCleanup(vs)
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
			azureCredSecret  = "test-azure-cred"
			azureAcKeySecret = "test-azure-ac-key"
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

				sr := core.Secret{
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

				acSr := core.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      azureAcKeySecret,
						Namespace: f.Namespace(),
					},
					Data: map[string][]byte{
						"account_key": []byte(accountKey),
					},
				}
				Expect(f.CreateSecret(acSr)).NotTo(HaveOccurred())

				azure := api.BackendStorageSpec{
					Azure: &api.AzureSpec{
						AccountName:      accountName,
						AccountKeySecret: azureAcKeySecret,
						Container:        "vault",
					},
				}

				unsealer := api.UnsealerSpec{
					SecretShares:          4,
					SecretThreshold:       2,
					InsecureSkipTLSVerify: true,
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
				Expect(f.DeleteSecret(azureAcKeySecret, vs.Namespace)).NotTo(HaveOccurred())
				checkForSecretDeleted(azureAcKeySecret, vs.Namespace)

				Expect(f.DeleteVaultServer(vs.ObjectMeta)).NotTo(HaveOccurred())
				checkForVaultServerCleanup(vs)
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

				sr := core.Secret{
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
					SecretShares:          4,
					SecretThreshold:       2,
					InsecureSkipTLSVerify: true,
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
				checkForVaultServerCleanup(vs)
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
				url, err := f.DeployMySQLForVault()
				Expect(err).NotTo(HaveOccurred())

				sr := core.Secret{
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
					SecretShares:          4,
					SecretThreshold:       2,
					InsecureSkipTLSVerify: true,
					Mode: api.ModeSpec{
						KubernetesSecret: &api.KubernetesSecretSpec{
							SecretName: k8sSecretName,
						},
					},
				}

				vs = f.VaultServerWithUnsealer(1, mysql, unsealer)
			})

			AfterEach(func() {

				Expect(f.DeleteMySQLForVault()).NotTo(HaveOccurred())

				Expect(f.DeleteSecret(k8sSecretName, vs.Namespace)).NotTo(HaveOccurred())
				checkForSecretDeleted(k8sSecretName, vs.Namespace)
				Expect(f.DeleteSecret(mysqlCredSecret, vs.Namespace)).NotTo(HaveOccurred())
				checkForSecretDeleted(mysqlCredSecret, vs.Namespace)

				Expect(f.DeleteVaultServer(vs.ObjectMeta)).NotTo(HaveOccurred())
				checkForVaultServerCleanup(vs)
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
					SecretShares:          4,
					SecretThreshold:       2,
					InsecureSkipTLSVerify: true,
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
				checkForVaultServerCleanup(vs)
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
			region        = "us-west-1"
			table         = "vault-dynamodb-test-1234"
			readCapacity  = 5
			writeCapacity = 5
		)

		Context("using dynamoDB backend", func() {
			BeforeEach(func() {
				if !f.RunDynamoDBTest {
					Skip("dynamodb test is skipped")
				}

				sr := core.Secret{
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

				Expect(f.DynamoDBCreateTable(region, table, readCapacity, writeCapacity)).NotTo(HaveOccurred())

				dynamodb := api.BackendStorageSpec{
					DynamoDB: &api.DynamoDBSpec{
						Region:           region,
						CredentialSecret: awsCredSecret,
						Table:            table,
						ReadCapacity:     readCapacity,
						WriteCapacity:    writeCapacity,
					},
				}

				unsealer := api.UnsealerSpec{
					SecretShares:          4,
					SecretThreshold:       2,
					InsecureSkipTLSVerify: true,
					Mode: api.ModeSpec{
						KubernetesSecret: &api.KubernetesSecretSpec{
							SecretName: "k8s-dynamodb-keys-1234",
						},
					},
				}

				vs = f.VaultServerWithUnsealer(1, dynamodb, unsealer)
			})

			AfterEach(func() {
				Expect(f.DynamoDBDeleteTable(region, table)).NotTo(HaveOccurred())

				Expect(f.DeleteSecret(awsCredSecret, vs.Namespace)).NotTo(HaveOccurred())
				checkForSecretDeleted(awsCredSecret, vs.Namespace)

				Expect(f.DeleteVaultServer(vs.ObjectMeta)).NotTo(HaveOccurred())
				checkForVaultServerCleanup(vs)
			})

			It("should create vault server", func() {
				shouldCreateVaultServer(vs)

				checkForVaultIsUnsealed(vs)
			})
		})
	})

	Describe("Vault HA cluster", func() {
		var (
			vs     *api.VaultServer
			unseal *api.UnsealerSpec
		)

		const (
			secretName = "test-vault-keys"
		)

		BeforeEach(func() {
			unseal = &api.UnsealerSpec{
				SecretShares:          4,
				SecretThreshold:       2,
				InsecureSkipTLSVerify: true,
				Mode: api.ModeSpec{
					KubernetesSecret: &api.KubernetesSecretSpec{
						SecretName: secretName,
					},
				},
			}
		})

		AfterEach(func() {
			f.DeleteVaultServer(vs.ObjectMeta)
			checkForVaultServerCleanup(vs)

			err := f.DeleteSecret(secretName, vs.Namespace)
			Expect(err).NotTo(HaveOccurred())
		})

		Context("using etcd backend", func() {
			BeforeEach(func() {
				url, err := f.DeployEtcd()
				Expect(err).NotTo(HaveOccurred())

				etcd := api.BackendStorageSpec{
					Etcd: &api.EtcdSpec{
						EtcdApi:  "v3",
						Address:  url,
						HAEnable: true,
					},
				}

				vs = f.VaultServerWithUnsealer(1, etcd, *unseal)
				vs.Spec.ServiceTemplate = ofst.ServiceTemplateSpec{
					Spec: ofst.ServiceSpec{
						Type: core.ServiceTypeNodePort,
					},
				}
			})

			AfterEach(func() {
				err := f.DeleteEtcd()
				Expect(err).NotTo(HaveOccurred())
			})

			It("vault should be unsealed", func() {
				shouldCreateVaultServer(vs)

				checkForSecretCreated(secretName, vs.Namespace)
				checkForVaultIsUnsealed(vs)
				checkForIsVaultHAEnabled(vs)
			})
		})
	})
})
