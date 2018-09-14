package google

import (
	"fmt"
	"path/filepath"

	kutilcorev1 "github.com/appscode/kutil/core/v1"
	api "github.com/kubevault/operator/apis/core/v1alpha1"
	"github.com/kubevault/operator/pkg/vault/util"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	rbac "k8s.io/api/rbac/v1"
)

const (
	ModeGoogleCloudKmsGCS  = "google-cloud-kms-gcs"
	GoogleCredentialFile   = "/etc/vault/unsealer/google/creds/sa.json"
	GoogleCredentialEnv    = "GOOGLE_APPLICATION_CREDENTIALS"
	GoogleCredentialVolume = "vault-unsealer-google-credential"
)

type Options struct {
	api.GoogleKmsGcsSpec
}

func NewOptions(s api.GoogleKmsGcsSpec) (*Options, error) {
	return &Options{
		s,
	}, nil
}

func (o *Options) Apply(pt *corev1.PodTemplateSpec) error {
	if pt == nil {
		return errors.New("podTempleSpec is nil")
	}

	var args []string
	var cont corev1.Container

	for _, c := range pt.Spec.Containers {
		if c.Name == util.VaultUnsealerContainerName() {
			cont = c
		}
	}

	args = append(args, fmt.Sprintf("--mode=%s", ModeGoogleCloudKmsGCS))
	if o.Bucket != "" {
		args = append(args, fmt.Sprintf("--google.storage-bucket=%s", o.Bucket))
	}
	if o.KmsProject != "" {
		args = append(args, fmt.Sprintf("--google.kms-project=%s", o.KmsProject))
	}
	if o.KmsLocation != "" {
		args = append(args, fmt.Sprintf("--google.kms-location=%s", o.KmsLocation))
	}
	if o.KmsKeyRing != "" {
		args = append(args, fmt.Sprintf("--google.kms-key-ring=%s", o.KmsKeyRing))
	}
	if o.KmsCryptoKey != "" {
		args = append(args, fmt.Sprintf("--google.kms-crypto-key=%s", o.KmsCryptoKey))
	}

	cont.Args = append(cont.Args, args...)

	if o.CredentialSecret != "" {
		pt.Spec.Volumes = kutilcorev1.UpsertVolume(pt.Spec.Volumes, corev1.Volume{
			Name: GoogleCredentialVolume,
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: o.CredentialSecret,
				},
			},
		})

		cont.VolumeMounts = kutilcorev1.UpsertVolumeMount(cont.VolumeMounts, corev1.VolumeMount{
			Name:      GoogleCredentialVolume,
			MountPath: filepath.Dir(GoogleCredentialFile),
		})

		cont.Env = kutilcorev1.UpsertEnvVars(cont.Env, corev1.EnvVar{
			Name:  GoogleCredentialEnv,
			Value: GoogleCredentialFile,
		})
	}

	pt.Spec.Containers = kutilcorev1.UpsertContainer(pt.Spec.Containers, cont)
	return nil
}

// GetRBAC returns required rbac roles
func (o *Options) GetRBAC(namespace string) []rbac.Role {
	return nil
}
