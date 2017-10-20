package controller

import (
	"fmt"

	"github.com/appscode/go/log"
	stringz "github.com/appscode/go/strings"
	v1u "github.com/appscode/kutil/core/v1"
	extu "github.com/appscode/kutil/extensions/v1beta1"
	"github.com/golang/glog"
	"github.com/hashicorp/vault/api"
	apiv1 "k8s.io/api/core/v1"
	extensions "k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/runtime"
)

func (c *VaultController) runReplicaSetWatcher() {
	for c.processNextReplicaSet() {
	}
}

func (c *VaultController) processNextReplicaSet() bool {
	// Wait until there is a new item in the working queue
	key, quit := c.rsQueue.Get()
	if quit {
		return false
	}
	// Tell the queue that we are done with processing this key. This unblocks the key for other workers
	// This allows safe parallel processing because two deployments with the same key are never processed in
	// parallel.
	defer c.rsQueue.Done(key)

	// Invoke the method containing the business logic
	err := c.runReplicaSetInitializer(key.(string))
	if err == nil {
		// Forget about the #AddRateLimited history of the key on every successful synchronization.
		// This ensures that future processing of updates for this key is not delayed because of
		// an outdated error history.
		c.rsQueue.Forget(key)
		return true
	}
	log.Errorln("Failed to process ReplicaSet %v. Reason: %s", key, err)

	// This controller retries 5 times if something goes wrong. After that, it stops trying.
	if c.rsQueue.NumRequeues(key) < c.options.MaxNumRequeues {
		glog.Infof("Error syncing deployment %v: %v", key, err)

		// Re-enqueue the key rate limited. Based on the rate limiter on the
		// queue and the re-enqueue history, the key will be processed later again.
		c.rsQueue.AddRateLimited(key)
		return true
	}

	c.rsQueue.Forget(key)
	// Report to an external entity that, even after several retries, we could not successfully process this key
	runtime.HandleError(err)
	glog.Infof("Dropping deployment %q out of the queue: %v", key, err)
	return true
}

// syncToStdout is the business logic of the controller. In this controller it simply prints
// information about the deployment to stdout. In case an error happened, it has to simply return the error.
// The retry logic should not be part of the business logic.
func (c *VaultController) runReplicaSetInitializer(key string) error {
	obj, exists, err := c.rsIndexer.GetByKey(key)
	if err != nil {
		glog.Errorf("Fetching object with key %s from store failed with %v", key, err)
		return err
	}

	if !exists {
		// Below we will warm up our cache with a ReplicaSet, so that we will see a delete for one d
		fmt.Printf("ReplicaSet %s does not exist anymore\n", key)
	} else {
		d := obj.(*extensions.ReplicaSet)
		fmt.Printf("Sync/Add/Update for ReplicaSet %s\n", d.GetName())

		if d.DeletionTimestamp != nil {
			if v1u.HasFinalizer(d.ObjectMeta, "finalizer.kubernetes.io/vault") ||
				v1u.HasFinalizer(d.ObjectMeta, "initializer.kubernetes.io/vault") {
				d, err = extu.PatchReplicaSet(c.k8sClient, d, func(in *extensions.ReplicaSet) *extensions.ReplicaSet {
					in.ObjectMeta = v1u.RemoveFinalizer(in.ObjectMeta, "finalizer.kubernetes.io/vault")
					return in
				})
				return err
			}
		} else if d.GetInitializers() != nil {
			pendingInitializers := d.GetInitializers().Pending
			if pendingInitializers[0].Name == "vault.initializer.kubernetes.io" {
				serviceAccountName := stringz.Val(d.Spec.Template.Spec.ServiceAccountName, "default")

				sa, err := c.k8sClient.CoreV1().ServiceAccounts(d.Namespace).Get(serviceAccountName, metav1.GetOptions{})
				if err != nil {
					return err
				}

				var vaultSecret *apiv1.Secret
				if secretName, found := GetString(sa.Annotations, "vaultproject.io/secret.name"); !found {
					return fmt.Errorf("missing vault secret annotation for service account %s", serviceAccountName)
				} else {
					vaultSecret, err = c.k8sClient.CoreV1().Secrets(d.Namespace).Get(secretName, metav1.GetOptions{})
					if err != nil {
						return err
					}
				}

				d, err = extu.PatchReplicaSet(c.k8sClient, d, func(in *extensions.ReplicaSet) *extensions.ReplicaSet {
					in.ObjectMeta = v1u.RemoveNextInitializer(in.ObjectMeta)
					in.ObjectMeta = v1u.AddFinalizer(in.ObjectMeta, "finalizer.kubernetes.io/vault")

					volSrc := apiv1.SecretVolumeSource{
						SecretName: vaultSecret.Name,
						Items: []apiv1.KeyToPath{
							{
								Key:  api.EnvVaultAddress,
								Path: "vault-addr",
								// Mode:
							},
							{
								Key:  api.EnvVaultToken,
								Path: "token",
								// Mode:
							},
							{
								Key:  "VAULT_TOKEN_ACCESSOR",
								Path: "token-accessor",
								// Mode:
							},
							{
								Key:  "LEASE_DURATION",
								Path: "lease-duration",
								// Mode:
							},
							{
								Key:  "RENEWABLE",
								Path: "renewable",
								// Mode:
							},
						},
						// DefaultMode
					}
					if _, found := vaultSecret.Data[api.EnvVaultCACert]; found {
						volSrc.Items = append(volSrc.Items, apiv1.KeyToPath{
							Key:  api.EnvVaultCACert,
							Path: "ca.crt",
							// Mode:
						})
					}
					in.Spec.Template.Spec.Volumes = v1u.UpsertVolume(in.Spec.Template.Spec.Volumes, apiv1.Volume{
						Name: vaultSecret.Name,
						VolumeSource: apiv1.VolumeSource{
							Secret: &volSrc,
						},
					})
					for ci, c := range in.Spec.Template.Spec.Containers {
						c.Env = v1u.UpsertEnvVar(c.Env, apiv1.EnvVar{
							Name: api.EnvVaultAddress,
							ValueFrom: &apiv1.EnvVarSource{
								SecretKeyRef: &apiv1.SecretKeySelector{
									LocalObjectReference: apiv1.LocalObjectReference{
										Name: vaultSecret.Name,
									},
									Key: api.EnvVaultAddress,
								},
							},
						})
						if _, found := vaultSecret.Data[api.EnvVaultCACert]; found {
							c.Env = v1u.UpsertEnvVar(c.Env, apiv1.EnvVar{
								Name:  api.EnvVaultCAPath,
								Value: "/var/run/secrets/vaultproject.io/approle/ca.crt",
							})
						}
						in.Spec.Template.Spec.Containers[ci].Env = c.Env

						in.Spec.Template.Spec.Containers[ci].VolumeMounts = v1u.UpsertVolumeMount(c.VolumeMounts, apiv1.VolumeMount{
							Name:      vaultSecret.Name,
							MountPath: "/var/run/secrets/vaultproject.io/approle",
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
