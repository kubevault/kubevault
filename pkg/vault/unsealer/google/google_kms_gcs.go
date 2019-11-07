/*
Copyright The KubeVault Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package google

import (
	"fmt"
	"path/filepath"

	api "kubevault.dev/operator/apis/kubevault/v1alpha1"
	"kubevault.dev/operator/pkg/vault/util"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	rbac "k8s.io/api/rbac/v1"
	core_util "kmodules.xyz/client-go/core/v1"
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
		if c.Name == util.VaultUnsealerContainerName {
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
		pt.Spec.Volumes = core_util.UpsertVolume(pt.Spec.Volumes, corev1.Volume{
			Name: GoogleCredentialVolume,
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: o.CredentialSecret,
				},
			},
		})

		cont.VolumeMounts = core_util.UpsertVolumeMount(cont.VolumeMounts, corev1.VolumeMount{
			Name:      GoogleCredentialVolume,
			MountPath: filepath.Dir(GoogleCredentialFile),
		})

		cont.Env = core_util.UpsertEnvVars(cont.Env, corev1.EnvVar{
			Name:  GoogleCredentialEnv,
			Value: GoogleCredentialFile,
		})
	}

	pt.Spec.Containers = core_util.UpsertContainer(pt.Spec.Containers, cont)
	return nil
}

// GetRBAC returns required rbac roles
func (o *Options) GetRBAC(prefix, namespace string) []rbac.Role {
	return nil
}
