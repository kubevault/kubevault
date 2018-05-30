package e2e_test

import (
	"path/filepath"
	"time"

	"fmt"
	"math/rand"

	api "github.com/kube-vault/operator/apis/core/v1alpha1"
	"github.com/kube-vault/operator/pkg/controller"
	"github.com/kube-vault/operator/pkg/util"
	"github.com/kube-vault/operator/test/e2e/framework"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
)

const (
	timeOut         = 10 * time.Minute
	pollingInterval = 10 * time.Second
)

var _ = Describe("VaultServer", func() {
	var (
		f  *framework.Invocation
		vs *api.VaultServer
	)

	BeforeEach(func() {
		f = root.Invoke()
	})
	AfterEach(func() {
		time.Sleep(30 * time.Second)
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
			}, timeOut, pollingInterval).Should(BeTrue())
		}
		checkForVaultTLSSecretDeleted = func(name, namespace string) {
			By(fmt.Sprintf("Waiting for vault tls secret (%s/%s) to delete", namespace, name))
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

		checkForVaultDeploymentCreatedOrUpdated = func(name, namespace string) {
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
				_, err := f.VaultServerClient.VaultV1alpha1().VaultServers(namespace).Get(name, metav1.GetOptions{})
				return err == nil
			}, timeOut, pollingInterval).Should(BeTrue())
		}
		checkForVaultServerDeleted = func(name, namespace string) {
			By(fmt.Sprintf("Waiting for vault server (%s/%s) to delete", namespace, name))
			Eventually(func() bool {
				_, err := f.VaultServerClient.VaultV1alpha1().VaultServers(namespace).Get(name, metav1.GetOptions{})
				return kerrors.IsNotFound(err)
			}, timeOut, pollingInterval).Should(BeTrue())
		}

		shouldCreateVaultServer = func() {
			By("Creating vault server")
			_, err := f.CreateVaultServer(vs)
			Expect(err).NotTo(HaveOccurred())

			checkForVaultServerCreated(vs.Name, vs.Namespace)
			checkForVaultTLSSecretCreated(controller.VaultTlsSecretName, vs.Namespace)
			checkForVaultConfigMapCreated(util.ConfigMapNameForVault(vs), vs.Namespace)
			checkForVaultDeploymentCreatedOrUpdated(vs.Name, vs.Namespace)
		}

		shouldUpdateVaultServerReplica = func(replicas int32) {
			By("Getting current vault server")
			vs, err := f.GetVaultServer(vs)
			Expect(err).NotTo(HaveOccurred())

			By("Updating replica number")
			vs.Spec.Nodes = replicas
			_, err = f.UpdateVaultServer(vs)
			Expect(err).NotTo(HaveOccurred())

			checkForVaultDeploymentCreatedOrUpdated(vs.Name, vs.Namespace)
		}
	)

	Describe("Creating vault server for", func() {
		Context("inmem backend", func() {
			BeforeEach(func() {
				vs = f.VaultServer(3, backendInmem)
			})
			AfterEach(func() {
				f.DeleteVaultServer(vs.ObjectMeta)

				checkForVaultServerDeleted(vs.Name, vs.Namespace)
				checkForVaultTLSSecretDeleted(controller.VaultTlsSecretName, vs.Namespace)
				checkForVaultConfigMapDeleted(util.ConfigMapNameForVault(vs), vs.Namespace)
				checkForVaultDeploymentDeleted(vs.Name, vs.Namespace)
			})

			It("should create vault server", shouldCreateVaultServer)
		})
	})

	Describe("Updating vault server replica for", func() {
		Context("inmem backend", func() {
			BeforeEach(func() {
				vs = f.VaultServer(3, backendInmem)
				shouldCreateVaultServer()
			})
			AfterEach(func() {
				f.DeleteVaultServer(vs.ObjectMeta)

				checkForVaultServerDeleted(vs.Name, vs.Namespace)
				checkForVaultTLSSecretDeleted(controller.VaultTlsSecretName, vs.Namespace)
				checkForVaultConfigMapDeleted(util.ConfigMapNameForVault(vs), vs.Namespace)
				checkForVaultDeploymentDeleted(vs.Name, vs.Namespace)
			})

			It("should update replica number to 1", func() {
				shouldUpdateVaultServerReplica(1)
			})

			It("should update replica number to 5", func() {
				shouldUpdateVaultServerReplica(4)
			})
		})
	})

	Describe("Deleting vault resources", func() {
		Context("using inmem as backend", func() {
			var (
				err error
			)

			BeforeEach(func() {
				vs = f.VaultServer(3, backendInmem)
				shouldCreateVaultServer()
			})
			AfterEach(func() {
				f.DeleteVaultServer(vs.ObjectMeta)

				checkForVaultServerDeleted(vs.Name, vs.Namespace)
				checkForVaultTLSSecretDeleted(controller.VaultTlsSecretName, vs.Namespace)
				checkForVaultConfigMapDeleted(util.ConfigMapNameForVault(vs), vs.Namespace)
				checkForVaultDeploymentDeleted(vs.Name, vs.Namespace)
			})

			It("should keep the number of pods same as specification, after deleting some pods", func() {
				Eventually(func() bool {
					vs, err = f.VaultServerClient.VaultV1alpha1().VaultServers(vs.Namespace).Get(vs.Name, metav1.GetOptions{})
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
			)

			BeforeEach(func() {
				vs = f.VaultServer(3, backendInmem)
				shouldCreateVaultServer()
			})
			AfterEach(func() {
				f.DeleteVaultServer(vs.ObjectMeta)

				checkForVaultServerDeleted(vs.Name, vs.Namespace)
				checkForVaultTLSSecretDeleted(controller.VaultTlsSecretName, vs.Namespace)
				checkForVaultConfigMapDeleted(util.ConfigMapNameForVault(vs), vs.Namespace)
				checkForVaultDeploymentDeleted(vs.Name, vs.Namespace)
			})

			It("status should contain 3 updated pods and 3 sealed pods", func() {
				Eventually(func() bool {
					vs, err = f.VaultServerClient.VaultV1alpha1().VaultServers(vs.Namespace).Get(vs.Name, metav1.GetOptions{})
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

})
