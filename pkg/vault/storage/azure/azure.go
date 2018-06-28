package azure

import (
	"fmt"
	"strings"

	api "github.com/kubevault/operator/apis/core/v1alpha1"
	corev1 "k8s.io/api/core/v1"
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
	if o.AccountKey != "" {
		params = append(params, fmt.Sprintf(`accountKey = "%s"`, o.AccountKey))
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
