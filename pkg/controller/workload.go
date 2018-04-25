package controller

import (
	"fmt"

	stringz "github.com/appscode/go/strings"
	wapi "github.com/appscode/kubernetes-webhook-util/apis/workload/v1"
	core_util "github.com/appscode/kutil/core/v1"
	"github.com/hashicorp/vault/api"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	finalizerName = "vault"
)

func (c *VaultController) mutateWorkload(w *wapi.Workload) error {
	fmt.Printf("Sync/Add/Update for Deployment %s\n", w.GetName())

	w.Object = nil

	if w.DeletionTimestamp != nil {
		if core_util.HasFinalizer(w.ObjectMeta, finalizerName) {
			w.ObjectMeta = core_util.RemoveFinalizer(w.ObjectMeta, finalizerName)
			return nil
		}
	} else {
		serviceAccountName := stringz.Val(w.Spec.Template.Spec.ServiceAccountName, "default")

		sa, err := c.kubeClient.CoreV1().ServiceAccounts(w.Namespace).Get(serviceAccountName, metav1.GetOptions{})
		if err != nil {
			return err
		}

		var vaultSecret *core.Secret
		if secretName, found := GetString(sa.Annotations, "vaultproject.io/secret.name"); !found {
			return fmt.Errorf("missing vault secret annotation for service account %s", serviceAccountName)
		} else {
			vaultSecret, err = c.kubeClient.CoreV1().Secrets(w.Namespace).Get(secretName, metav1.GetOptions{})
			if err != nil {
				return err
			}
		}

		w.ObjectMeta = core_util.RemoveNextInitializer(w.ObjectMeta)
		w.ObjectMeta = core_util.AddFinalizer(w.ObjectMeta, finalizerName)

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
		w.Spec.Template.Spec.Volumes = core_util.UpsertVolume(w.Spec.Template.Spec.Volumes, core.Volume{
			Name: vaultSecret.Name,
			VolumeSource: core.VolumeSource{
				Secret: &volSrc,
			},
		})
		for ci, c := range w.Spec.Template.Spec.Containers {
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
			w.Spec.Template.Spec.Containers[ci].Env = c.Env

			w.Spec.Template.Spec.Containers[ci].VolumeMounts = core_util.UpsertVolumeMount(c.VolumeMounts, core.VolumeMount{
				Name:      vaultSecret.Name,
				MountPath: "/var/run/secrets/vaultproject.io/approle",
				ReadOnly:  true,
			})
		}
	}
	return nil
}
