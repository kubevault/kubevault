package controller

import (
	"fmt"

	stringz "github.com/appscode/go/strings"
	batch_util "github.com/appscode/kutil/batch/v1"
	core_util "github.com/appscode/kutil/core/v1"
	"github.com/appscode/kutil/tools/queue"
	"github.com/golang/glog"
	"github.com/hashicorp/vault/api"
	batch "k8s.io/api/batch/v1"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (c *VaultController) initJobWatcher() {
	c.jobInformer = c.informerFactory.Batch().V1().Jobs().Informer()
	c.jobQueue = queue.New("Job", c.options.MaxNumRequeues, c.options.NumThreads, c.runJobInitializer)
	c.jobInformer.AddEventHandler(queue.DefaultEventHandler(c.jobQueue.GetQueue()))
	// c.jobLister = c.informerFactory.Batch().V1().Jobs().Lister()
}

func (c *VaultController) runJobInitializer(key string) error {
	obj, exists, err := c.jobInformer.GetIndexer().GetByKey(key)
	if err != nil {
		glog.Errorf("Fetching object with key %s from store failed with %v", key, err)
		return err
	}

	if !exists {
		// Below we will warm up our cache with a Job, so that we will see a delete for one d
		fmt.Printf("Job %s does not exist anymore\n", key)
	} else {
		dp := obj.(*batch.Job)
		fmt.Printf("Sync/Add/Update for Job %s\n", dp.GetName())

		if dp.DeletionTimestamp != nil {
			if core_util.HasFinalizer(dp.ObjectMeta, "finalizer.kubernetes.io/vault") ||
				core_util.HasFinalizer(dp.ObjectMeta, "initializer.kubernetes.io/vault") {
				dp, _, err = batch_util.PatchJob(c.k8sClient, dp, func(in *batch.Job) *batch.Job {
					in.ObjectMeta = core_util.RemoveFinalizer(in.ObjectMeta, "finalizer.kubernetes.io/vault")
					return in
				})
				return err
			}
		} else if dp.GetInitializers() != nil {
			pendingInitializers := dp.GetInitializers().Pending
			if pendingInitializers[0].Name == "vault.initializer.kubernetes.io" {
				serviceAccountName := stringz.Val(dp.Spec.Template.Spec.ServiceAccountName, "default")

				sa, err := c.k8sClient.CoreV1().ServiceAccounts(dp.Namespace).Get(serviceAccountName, metav1.GetOptions{})
				if err != nil {
					return err
				}

				var vaultSecret *core.Secret
				if secretName, found := GetString(sa.Annotations, "vaultproject.io/secret.name"); !found {
					return fmt.Errorf("missing vault secret annotation for service account %s", serviceAccountName)
				} else {
					vaultSecret, err = c.k8sClient.CoreV1().Secrets(dp.Namespace).Get(secretName, metav1.GetOptions{})
					if err != nil {
						return err
					}
				}

				dp, _, err = batch_util.PatchJob(c.k8sClient, dp, func(in *batch.Job) *batch.Job {
					in.ObjectMeta = core_util.RemoveNextInitializer(in.ObjectMeta)
					in.ObjectMeta = core_util.AddFinalizer(in.ObjectMeta, "finalizer.kubernetes.io/vault")

					volSrc := core.SecretVolumeSource{
						SecretName: vaultSecret.Name,
						Items: []core.KeyToPath{
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
						volSrc.Items = append(volSrc.Items, core.KeyToPath{
							Key:  api.EnvVaultCACert,
							Path: "ca.crt",
							// Mode:
						})
					}
					in.Spec.Template.Spec.Volumes = core_util.UpsertVolume(in.Spec.Template.Spec.Volumes, core.Volume{
						Name: vaultSecret.Name,
						VolumeSource: core.VolumeSource{
							Secret: &volSrc,
						},
					})
					for ci, c := range in.Spec.Template.Spec.Containers {
						c.Env = core_util.UpsertEnvVars(c.Env, core.EnvVar{
							Name: api.EnvVaultAddress,
							ValueFrom: &core.EnvVarSource{
								SecretKeyRef: &core.SecretKeySelector{
									LocalObjectReference: core.LocalObjectReference{
										Name: vaultSecret.Name,
									},
									Key: api.EnvVaultAddress,
								},
							},
						})
						if _, found := vaultSecret.Data[api.EnvVaultCACert]; found {
							c.Env = core_util.UpsertEnvVars(c.Env, core.EnvVar{
								Name:  api.EnvVaultCAPath,
								Value: "/var/run/secrets/vaultproject.io/approle/ca.crt",
							})
						}
						in.Spec.Template.Spec.Containers[ci].Env = c.Env

						in.Spec.Template.Spec.Containers[ci].VolumeMounts = core_util.UpsertVolumeMount(c.VolumeMounts, core.VolumeMount{
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
