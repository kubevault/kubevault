package gcs

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	api "github.com/kube-vault/operator/apis/core/v1alpha1"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var gcsStorageFmt = `
storage "gcs" {
%s
}
`

const (
	GoogleCredentialFile   = "/etc/vault/storage/gcs/creds/credential.json"
	GoogleCredentialEnv    = "GOOGLE_APPLICATION_CREDENTIALS"
	GoogleCredentialSecret = "vault-google-credential"
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

func (o *Options) Apply(pt *corev1.PodTemplateSpec) error {
	if o.CredentialPath != "" {
		pt.Spec.Volumes = append(pt.Spec.Volumes, corev1.Volume{
			Name: GoogleCredentialVolume,
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: GoogleCredentialSecret,
				},
			},
		})

		pt.Spec.Containers[0].VolumeMounts = append(pt.Spec.Containers[0].VolumeMounts, corev1.VolumeMount{
			Name:      GoogleCredentialVolume,
			MountPath: filepath.Dir(GoogleCredentialFile),
		})

		pt.Spec.Containers[0].Env = append(pt.Spec.Containers[0].Env, corev1.EnvVar{
			Name:  GoogleCredentialEnv,
			Value: GoogleCredentialFile,
		})
	}
	return nil
}

func (o *Options) GetSecrets(namespace string) ([]corev1.Secret, error) {
	var secrets []corev1.Secret
	if o.CredentialPath != "" {
		data, err := ioutil.ReadFile(o.CredentialPath)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to read google credential file(%s)", o.CredentialPath)
		}

		secrets = append(secrets, corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name: GoogleCredentialSecret,
			},
			Data: map[string][]byte{
				filepath.Base(GoogleCredentialFile): data,
			},
		})
	}
	return secrets, nil
}

// vault doc: https://www.vaultproject.io/docs/configuration/storage/google-cloud-storage.html
//
// GetGcsConfig creates gcs storae config from GcsSpec
func (o *Options) GetStorageConfig() (string, error) {
	params := []string{}
	if o.Bucket != "" {
		params = append(params, fmt.Sprintf(`bucket = "%s"`, o.Bucket))
	}
	if o.HAEnabled == true {
		params = append(params, fmt.Sprintf(`ha_enabled = "true"`))
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
