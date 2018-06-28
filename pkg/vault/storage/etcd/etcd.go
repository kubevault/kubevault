package etcd

import (
	"fmt"
	"path/filepath"
	"strings"

	api "github.com/kubevault/operator/apis/core/v1alpha1"
	"github.com/kubevault/operator/pkg/vault/util"
	corev1 "k8s.io/api/core/v1"
)

const (
	// TLS related file name for etcd
	EtcdTLSAssetDir    = "/etc/vault/storage/etcd/tls/"
	EtcdClientCaName   = "etcd-ca.crt"
	EtcdClientCertName = "etcd-client.crt"
	EtcdClientKeyName  = "etcd-client.key"
)

var etcdStorageFmt = `
storage "etcd" {
%s
}
`

type Options struct {
	api.EtcdSpec
}

func NewOptions(s api.EtcdSpec) (*Options, error) {
	return &Options{
		s,
	}, nil
}

// Apply will do:
// - If TLSSecretName is provided, then add volume for etcd tls
// - If CredentialSecretName is provided, then set environment variable
func (o *Options) Apply(pt *corev1.PodTemplateSpec) error {
	etcdTLSAssetVolume := "vault-etcd-tls"
	if o.TLSSecretName != "" {
		// mount tls secret
		pt.Spec.Volumes = append(pt.Spec.Volumes, corev1.Volume{
			Name: etcdTLSAssetVolume,
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: o.TLSSecretName,
				},
			},
		})

		pt.Spec.Containers[0].VolumeMounts = append(pt.Spec.Containers[0].VolumeMounts, corev1.VolumeMount{
			Name:      etcdTLSAssetVolume,
			MountPath: util.EtcdTLSAssetDir,
			ReadOnly:  true,
		})
	}

	if o.CredentialSecretName != "" {
		// set env variable ETCD_USERNAME and ETCD_PASSWORD
		pt.Spec.Containers[0].Env = append(pt.Spec.Containers[0].Env,
			corev1.EnvVar{
				Name: "ETCD_USERNAME",
				ValueFrom: &corev1.EnvVarSource{
					SecretKeyRef: &corev1.SecretKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: o.CredentialSecretName,
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
							Name: o.CredentialSecretName,
						},
						Key: "password",
					},
				},
			},
		)
	}

	return nil
}

// vault doc: https://www.vaultproject.io/docs/configuration/storage/etcd.html
//
// Note:
// - Secret `TLSSecretName` mounted in `EtcdTLSAssetDir`
// - Secret `CredentialSecret` will be used as environment variable
//
// GetStorageConfig creates etcd storage config from EtcdSpec
func (o *Options) GetStorageConfig() (string, error) {
	params := []string{}
	if o.Address != "" {
		params = append(params, fmt.Sprintf(`address = "%s"`, o.Address))
	}
	if o.EtcdApi != "" {
		params = append(params, fmt.Sprintf(`etcd_api = "%s"`, o.EtcdApi))
	}
	if o.Path != "" {
		params = append(params, fmt.Sprintf(`path = "%s"`, o.Path))
	}
	if o.DiscoverySrv != "" {
		params = append(params, fmt.Sprintf(`discovery_srv = "%s"`, o.DiscoverySrv))
	}
	if o.HAEnable {
		params = append(params, fmt.Sprintf(`ha_enable = "true"`))
	} else {
		params = append(params, fmt.Sprintf(`ha_enable = "false"`))
	}
	if o.Sync {
		params = append(params, fmt.Sprintf(`sync = "true"`))
	} else {
		params = append(params, fmt.Sprintf(`sync = "false"`))
	}
	if o.TLSSecretName != "" {
		params = append(params, fmt.Sprintf(`tls_ca_file = "%s"`, filepath.Join(EtcdTLSAssetDir, EtcdClientCaName)),
			fmt.Sprintf(`tls_cert_file = "%s"`, filepath.Join(EtcdTLSAssetDir, EtcdClientCertName)),
			fmt.Sprintf(`tls_key_file = "%s"`, filepath.Join(EtcdTLSAssetDir, EtcdClientKeyName)))
	}

	storageCfg := fmt.Sprintf(etcdStorageFmt, strings.Join(params, "\n"))
	return storageCfg, nil
}
