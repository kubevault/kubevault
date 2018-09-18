package azure

import (
	"fmt"
	"strings"

	api "github.com/kubevault/operator/apis/core/v1alpha1"
	corev1 "k8s.io/api/core/v1"
)

const (
	AzureAccountKeyEnv = "AZURE_ACCOUNT_KEY"
)

var azureStorageFmt = `
storage "azure" {
%s
}
`

type Options struct {
	api.AzureSpec
}

func NewOptions(s api.AzureSpec) (*Options, error) {
	return &Options{
		s,
	}, nil
}

func (o *Options) Apply(pt *corev1.PodTemplateSpec) error {
	if o.AccountKeySecret != "" {
		pt.Spec.Containers[0].Env = append(pt.Spec.Containers[0].Env, corev1.EnvVar{
			Name: AzureAccountKeyEnv,
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: o.AccountKeySecret,
					},
					Key: "account_key",
				},
			},
		})
	}
	return nil
}

func (o *Options) GetSecrets(namespace string) ([]corev1.Secret, error) {
	return nil, nil
}

// vault doc: https://www.vaultproject.io/docs/configuration/storage/google-cloud-storage.html
//
// GetGcsConfig creates gcs storae config from GcsSpec
func (o *Options) GetStorageConfig() (string, error) {
	params := []string{}
	if o.AccountName != "" {
		params = append(params, fmt.Sprintf(`accountName = "%s"`, o.AccountName))
	}
	if o.Container != "" {
		params = append(params, fmt.Sprintf(`container = "%s"`, o.Container))
	}
	if o.MaxParallel != 0 {
		params = append(params, fmt.Sprintf(`max_parallel = %d`, o.MaxParallel))
	}

	storageCfg := fmt.Sprintf(azureStorageFmt, strings.Join(params, "\n"))
	return storageCfg, nil
}
