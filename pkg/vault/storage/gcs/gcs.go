/*
Copyright AppsCode Inc. and Contributors

Licensed under the AppsCode Community License 1.0.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    https://github.com/appscode/licenses/raw/1.0.0/AppsCode-Community-1.0.0.md

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package gcs

import (
	"fmt"
	"path/filepath"
	"strings"

	api "kubevault.dev/apimachinery/apis/kubevault/v1alpha1"

	core "k8s.io/api/core/v1"
)

var gcsStorageFmt = `
storage "gcs" {
%s
}
`

const (
	GoogleCredentialFile   = "/etc/vault/storage/gcs/creds/sa.json"
	GoogleCredentialEnv    = "GOOGLE_APPLICATION_CREDENTIALS"
	GoogleCredentialVolume = "vault-google-credential"
)

type Options struct {
	api.GcsSpec
}

func NewOptions(s api.GcsSpec) (*Options, error) {
	return &Options{
		s,
	}, nil
}

func (o *Options) Apply(pt *core.PodTemplateSpec) error {
	if o.CredentialSecret != "" {
		pt.Spec.Volumes = append(pt.Spec.Volumes, core.Volume{
			Name: GoogleCredentialVolume,
			VolumeSource: core.VolumeSource{
				Secret: &core.SecretVolumeSource{
					SecretName: o.CredentialSecret,
				},
			},
		})

		pt.Spec.Containers[0].VolumeMounts = append(pt.Spec.Containers[0].VolumeMounts, core.VolumeMount{
			Name:      GoogleCredentialVolume,
			MountPath: filepath.Dir(GoogleCredentialFile),
		})

		pt.Spec.Containers[0].Env = append(pt.Spec.Containers[0].Env, core.EnvVar{
			Name:  GoogleCredentialEnv,
			Value: GoogleCredentialFile,
		})
	}
	return nil
}

// vault doc: https://www.vaultproject.io/docs/configuration/storage/google-cloud-storage.html
//
//  GetStorageConfig creates gcs storae config from GcsSpec
func (o *Options) GetStorageConfig() (string, error) {
	params := []string{}
	if o.Bucket != "" {
		params = append(params, fmt.Sprintf(`bucket = "%s"`, o.Bucket))
	}
	if o.HAEnabled {
		params = append(params, `ha_enabled = "true"`)
	}
	if o.ChunkSize != "" {
		params = append(params, fmt.Sprintf(`chunk_size = "%s"`, o.ChunkSize))
	}
	if o.MaxParallel != 0 {
		params = append(params, fmt.Sprintf(`max_parallel = %d`, o.MaxParallel))
	}

	storageCfg := fmt.Sprintf(gcsStorageFmt, strings.Join(params, "\n"))
	return storageCfg, nil
}
