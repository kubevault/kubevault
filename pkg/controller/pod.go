package controller

import (
	"fmt"
	"path"
	"strconv"
	"time"

	stringz "github.com/appscode/go/strings"
	v1_util "github.com/appscode/kutil/core/v1"
	"github.com/golang/glog"
	"github.com/hashicorp/vault/api"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/pkg/api/v1"
)

func (c *VaultController) runPodWatcher() {
	for c.processNextPod() {
	}
}

func (c *VaultController) processNextPod() bool {
	// Wait until there is a new item in the working queue
	key, quit := c.podQueue.Get()
	if quit {
		return false
	}
	// Tell the queue that we are done with processing this key. This unblocks the key for other workers
	// This allows safe parallel processing because two pods with the same key are never processed in
	// parallel.
	defer c.podQueue.Done(key)

	// Invoke the method containing the business logic
	err := c.runPodInitializer(key.(string))
	if err == nil {
		// Forget about the #AddRateLimited history of the key on every successful synchronization.
		// This ensures that future processing of updates for this key is not delayed because of
		// an outdated error history.
		c.podQueue.Forget(key)
		return true
	}

	// This controller retries 5 times if something goes wrong. After that, it stops trying.
	if c.podQueue.NumRequeues(key) < c.options.MaxNumRequeues {
		glog.Infof("Error syncing pod %v: %v", key, err)

		// Re-enqueue the key rate limited. Based on the rate limiter on the
		// queue and the re-enqueue history, the key will be processed later again.
		c.podQueue.AddRateLimited(key)
		return true
	}

	c.podQueue.Forget(key)
	// Report to an external entity that, even after several retries, we could not successfully process this key
	runtime.HandleError(err)
	glog.Infof("Dropping pod %q out of the queue: %v", key, err)
	return true
}

// syncToStdout is the business logic of the controller. In this controller it simply prints
// information about the pod to stdout. In case an error happened, it has to simply return the error.
// The retry logic should not be part of the business logic.
func (c *VaultController) runPodInitializer(key string) error {
	obj, exists, err := c.podIndexer.GetByKey(key)
	if err != nil {
		glog.Errorf("Fetching object with key %s from store failed with %v", key, err)
		return err
	}

	if !exists {
		// Below we will warm up our cache with a Pod, so that we will see a delete for one pod
		fmt.Printf("Pod %s does not exist anymore\n", key)
	} else {
		pod := obj.(*v1.Pod)
		fmt.Printf("Sync/Add/Update for Pod %s\n", pod.GetName())
		if pod.DeletionTimestamp != nil {
			if v1_util.HasFinalizer(pod, "vault.initializer.kubernetes.io") {
				// find the secret, expire and remove the keys
				// v1_util.EnsureVolumeDeleted()

				vi := -1
				var vaultSecret *v1.Secret
				for i, vol := range pod.Spec.Volumes {
					if vol.Secret != nil {
						secret, err := c.k8sClient.CoreV1().Secrets(pod.Namespace).Get(vol.Secret.SecretName, metav1.GetOptions{})
						if err == nil {
							if v, found := GetString(secret.Annotations, "kubernetes.io/service-account.name"); found && v == pod.Spec.ServiceAccountName {
								vi = i
								vaultSecret = secret
								break
							}
						}
					}
				}
				if vi > -1 {
					// revoke token
					if accessor, found := vaultSecret.Data[pod.Name+".a"]; found {
						_, err := c.vaultClient.Logical().Write("/auth/token/revoke-accessor", map[string]interface{}{
							"accessor": string(accessor),
						})
						if err != nil {
							return err
						}
					}

					// remove data from secret
					vaultSecret, err = v1_util.PatchSecret(c.k8sClient, vaultSecret, func(in *v1.Secret) *v1.Secret {
						delete(in.Data, pod.Name+".t")
						delete(in.Data, pod.Name+".a")
						delete(in.Data, pod.Name+".c")
						delete(in.Data, pod.Name+".p")
						return in
					})
					if err != nil {
						return err
					}

					// remove secret
					pod.Spec.Volumes = append(pod.Spec.Volumes[:vi], pod.Spec.Volumes[vi+1:]...)
					for ci, c := range pod.Spec.Containers {
						c.Env = v1_util.EnsureEnvVarDeleted(c.Env, api.EnvVaultAddress)
						c.Env = v1_util.EnsureEnvVarDeleted(c.Env, api.EnvVaultCAPath)
						pod.Spec.Containers[ci].Env = c.Env
						pod.Spec.Containers[ci].VolumeMounts = v1_util.EnsureVolumeMountDeleted(c.VolumeMounts, vaultSecret.Name)
					}
				}
				pod, err = v1_util.PatchPod(c.k8sClient, pod, func(in *v1.Pod) *v1.Pod {
					v1_util.RemoveFinalizer(in, "vault.initializer.kubernetes.io")
					return in
				})
				if err != nil {
					return err
				}
			}
		} else if pod.GetInitializers() != nil {
			pendingInitializers := pod.GetInitializers().Pending
			if pendingInitializers[0].Name == "vault.initializer.kubernetes.io" {
				serviceAccountName := stringz.Val(pod.Spec.ServiceAccountName, "default")
				roleName := pod.Namespace + "." + serviceAccountName

				sa, err := c.k8sClient.CoreV1().ServiceAccounts(pod.Namespace).Get(serviceAccountName, metav1.GetOptions{})
				if err != nil {
					return err
				}

				var vaultSecret *v1.Secret
				if secretName, found := GetString(sa.Annotations, "vaultproject.io/secret.name"); !found {
					return fmt.Errorf("missing vault secret annotation for service account %s", serviceAccountName)
				} else {
					vaultSecret, err = c.k8sClient.CoreV1().Secrets(pod.Namespace).Get(secretName, metav1.GetOptions{})
					if err != nil {
						return err
					}
				}

				var vr *api.Secret
				vr, err = c.vaultClient.Logical().Read(path.Join("auth", c.options.AuthBackend(), "role", roleName, "role-id"))
				if err != nil {
					return err
				}
				roleID := vr.Data["role_id"]

				vr, err = c.vaultClient.Logical().Write(path.Join("auth", c.options.AuthBackend(), "role", roleName, "secret-id"), map[string]interface{}{
					"metadata": map[string]string{
						"host_ip":   pod.Status.HostIP,
						"namespace": pod.Namespace,
						"pod_ip":    pod.Status.PodIP,
						"pod_name":  pod.Name,
						"pod_uid":   string(pod.UID),
					},
				})
				secretID := vr.Data["secret_id"]

				vr, err = c.vaultClient.Logical().Write(path.Join("auth", c.options.AuthBackend(), "login"), map[string]interface{}{
					"role_id":   roleID,
					"secret_id": secretID,
				})
				if vr.WrapInfo == nil {
					return fmt.Errorf("missing wrapped token for role %s", roleName)
				}

				_, err = c.vaultClient.Logical().Write(path.Join("auth", c.options.AuthBackend(), "role", roleName, "secret-id", "destroy"), map[string]interface{}{
					"secret_id": secretID,
				})
				if err != nil {
					return err
				}

				vaultSecret, err = v1_util.PatchSecret(c.k8sClient, vaultSecret, func(in *v1.Secret) *v1.Secret {
					in.Data["VAULT_TOKEN_TTL"] = []byte(strconv.Itoa(vr.WrapInfo.TTL))
					in.Data[pod.Name+".t"] = []byte(vr.WrapInfo.Token)
					in.Data[pod.Name+".a"] = []byte(vr.WrapInfo.WrappedAccessor)
					in.Data[pod.Name+".c"] = []byte(vr.WrapInfo.CreationTime.UTC().Format(time.RFC3339))
					in.Data[pod.Name+".p"] = []byte(vr.WrapInfo.CreationPath)
					return in
				})
				if err != nil {
					return err
				}

				pod, err = v1_util.PatchPod(c.k8sClient, pod, func(in *v1.Pod) *v1.Pod {
					in.Initializers.Pending = in.Initializers.Pending[1:]
					v1_util.AddFinalizer(in, "vault.initializer.kubernetes.io")

					volSrc := v1.SecretVolumeSource{
						SecretName: vaultSecret.Name,
						Items: []v1.KeyToPath{
							{
								Key:  api.EnvVaultAddress,
								Path: "vault-addr",
								// Mode:
							},
							{
								Key:  "VAULT_TOKEN_TTL",
								Path: "vault-token-ttl",
								// Mode:
							},
							{
								Key:  in.Name + ".t",
								Path: "vault-wrapped-token",
								// Mode:
							},
							{
								Key:  in.Name + ".a",
								Path: "vault-token-accessor",
								// Mode:
							},
							{
								Key:  in.Name + ".c",
								Path: "vault-creation-time",
								// Mode:
							},
							{
								Key:  in.Name + ".p",
								Path: "vault-creation-path",
								// Mode:
							},
						},
						// DefaultMode
					}
					if _, found := vaultSecret.Data[api.EnvVaultCACert]; found {
						volSrc.Items = append(volSrc.Items, v1.KeyToPath{
							Key:  api.EnvVaultCACert,
							Path: "ca.crt",
							// Mode:
						})
					}
					in.Spec.Volumes = v1_util.UpsertVolume(in.Spec.Volumes, v1.Volume{
						Name: vaultSecret.Name,
						VolumeSource: v1.VolumeSource{
							Secret: &volSrc,
						},
					})
					for ci, c := range in.Spec.Containers {
						c.Env = v1_util.UpsertEnvVar(c.Env, v1.EnvVar{
							Name: api.EnvVaultAddress,
							ValueFrom: &v1.EnvVarSource{
								SecretKeyRef: &v1.SecretKeySelector{
									LocalObjectReference: v1.LocalObjectReference{
										Name: vaultSecret.Name,
									},
									Key: api.EnvVaultAddress,
								},
							},
						})
						if _, found := vaultSecret.Data[api.EnvVaultCACert]; found {
							c.Env = v1_util.UpsertEnvVar(c.Env, v1.EnvVar{
								Name:  api.EnvVaultCAPath,
								Value: "/var/run/secrets/vaultproject.io/ca.crt",
							})
						}
						in.Spec.Containers[ci].Env = c.Env

						in.Spec.Containers[ci].VolumeMounts = v1_util.UpsertVolumeMount(c.VolumeMounts, v1.VolumeMount{
							Name:      vaultSecret.Name,
							MountPath: "/var/run/secrets/vaultproject.io",
							ReadOnly:  true,
						})
					}

					return in
				})
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}
