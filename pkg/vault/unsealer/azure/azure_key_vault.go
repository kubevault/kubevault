package azure

import (
	"fmt"
	"path/filepath"

	api "github.com/kubevault/operator/apis/core/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	rbac "k8s.io/api/rbac/v1"
)

const (
	ModeAzureKeyVault = "azure-key-vault"
)

type Options struct {
	api.AzureKeyVault
}

func NewOptions(s api.AzureKeyVault) (*Options, error) {
	return &Options{
		s,
	}, nil
}

func (o *Options) Apply(pt *corev1.PodTemplateSpec, cont *corev1.Container) error {
	var args []string

	args = append(args, fmt.Sprintf("--mode=%s", ModeAzureKeyVault))

	if o.VaultBaseUrl != "" {
		args = append(args, fmt.Sprintf("--azure.vault-base-url=%s", o.VaultBaseUrl))
	}
	if o.TenantID != "" {
		args = append(args, fmt.Sprintf("--azure.tenant-id=%s", o.TenantID))
	}
	if o.Cloud != "" {
		args = append(args, fmt.Sprintf("--azure.cloud=%s", o.Cloud))
	}
	if o.UseManagedIdentity == true {
		args = append(args, fmt.Sprintf("--azure.use-managed-identity=true"))
	}

	var envs []corev1.EnvVar

	if o.AADClientSecret != "" {
		envs = append(envs, corev1.EnvVar{
			Name: "AZURE_CLIENT_ID",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: o.AADClientSecret,
					},
					Key: "client-id",
				},
			},
		}, corev1.EnvVar{
			Name: "AZURE_CLIENT_SECRET",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: o.AADClientSecret,
					},
					Key: "client-secret",
				},
			},
		})
	}

	if o.ClientCertSecret != "" {
		envs = append(envs, corev1.EnvVar{
			Name: "AZURE_CLIENT_CERT_PASSWORD",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: o.ClientCertSecret,
					},
					Key: "client-cert-password",
				},
			},
		})

		volumeName := "azure-client-cert"

		pt.Spec.Volumes = append(pt.Spec.Volumes, corev1.Volume{
			Name: volumeName,
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: o.ClientCertSecret,
					Items: []corev1.KeyToPath{
						{
							Key:  "client-cert",
							Path: "client.crt",
						},
					},
				},
			},
		})

		certFilePath := "/etc/vault/unsealer/azure/cert/client.crt"

		cont.VolumeMounts = append(cont.VolumeMounts, corev1.VolumeMount{
			Name:      volumeName,
			MountPath: filepath.Dir(certFilePath),
		})

		args = append(args, fmt.Sprintf("--azure.client-cert-path=%s", certFilePath))
	}

	cont.Args = append(cont.Args, args...)

	cont.Env = append(cont.Env, envs...)

	return nil
}

// GetRBAC returns required rbac roles
func (o *Options) GetRBAC(namespace string) []rbac.Role {
	return nil
}
