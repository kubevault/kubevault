package file

import (
	"fmt"
	"strings"

	api "github.com/kubevault/operator/apis/kubevault/v1alpha1"
	core "k8s.io/api/core/v1"
)

var fileStorageFmt = `
storage "file" {
%s
}
`

type Options struct {
	api.FileSpec
}

func NewOptions(s api.FileSpec) (*Options, error) {
	return &Options{
		s,
	}, nil
}

func (o *Options) Apply(pt *core.PodTemplateSpec) error {
	return nil
}

// vault doc: https://www.vaultproject.io/docs/configuration/storage/google-cloud-storage.html
//
// GetGcsConfig creates gcs storae config from GcsSpec
func (o *Options) GetStorageConfig() (string, error) {
	params := []string{}
	if o.Path != "" {
		params = append(params, fmt.Sprintf(`path = "%s"`, o.Path))
	}

	storageCfg := fmt.Sprintf(fileStorageFmt, strings.Join(params, "\n"))
	return storageCfg, nil
}
