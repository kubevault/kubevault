package controller

import (
	api "github.com/soter/vault-operator/apis/vault/v1alpha1"
	"github.com/soter/vault-operator/pkg/util"
	corev1 "k8s.io/api/core/v1"
)

// configureVaultStorage mounts the volume, set environment variable for vault storage configuration
func configureVaultStorage(v *api.VaultServer, pt *corev1.PodTemplateSpec) {
	storage := v.Spec.BackendStorage
	if storage.Etcd != nil {
		configureForEtcd(v.Spec.BackendStorage.Etcd, pt)
	}
}

// configureForEtcd will do:
// - If TLSSecretName is provided, then add volume for etcd tls
// - If CredentialSecretName is provided, then set environment variable
func configureForEtcd(etcd *api.EtcdSpec, pt *corev1.PodTemplateSpec) {
	etcdTLSAssetVolume := "vault-etcd-tls"
	if etcd.TLSSecretName != "" {
		// mount tls secret
		pt.Spec.Volumes = append(pt.Spec.Volumes, corev1.Volume{
			Name: etcdTLSAssetVolume,
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: etcd.TLSSecretName,
				},
			},
		})

		pt.Spec.Containers[0].VolumeMounts = append(pt.Spec.Containers[0].VolumeMounts, corev1.VolumeMount{
			Name:      etcdTLSAssetVolume,
			MountPath: util.EtcdTLSAssetDir,
		})
	}

	if etcd.CredentialSecretName != "" {
		// set env variable ETCD_USERNAME and ETCD_PASSWORD
		pt.Spec.Containers[0].Env = append(pt.Spec.Containers[0].Env,
			corev1.EnvVar{
				Name: "ETCD_USERNAME",
				ValueFrom: &corev1.EnvVarSource{
					SecretKeyRef: &corev1.SecretKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: etcd.CredentialSecretName,
						},
						Key: "username",
					},
				},
			},
			corev1.EnvVar{
				Name: "ETCD_PASSWORD",
				ValueFrom: &corev1.EnvVarSource{
					SecretKeyRef: &corev1.SecretKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: etcd.CredentialSecretName,
						},
						Key: "password",
					},
				},
			},
		)
	}
}
