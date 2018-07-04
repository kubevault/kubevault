package postgresql

import (
	"fmt"
	"strings"

	api "github.com/kubevault/operator/apis/core/v1alpha1"
	corev1 "k8s.io/api/core/v1"
)

var postgresqlStorageFmt = `
storage "postgresql" {
%s
}
`

type Options struct {
	api.PostgreSQLSpec
}

func NewOptions(s api.PostgreSQLSpec) (*Options, error) {
	return &Options{
		s,
	}, nil
}

func (o *Options) Apply(pt *corev1.PodTemplateSpec) error {
	return nil
}

// vault doc: https://www.vaultproject.io/docs/configuration/storage/postgresql.html
//
// GetGcsConfig creates postgresql storae config from GcsSpec
func (o *Options) GetStorageConfig() (string, error) {
	params := []string{}
	if o.ConnectionUrl != "" {
		params = append(params, fmt.Sprintf(`connection_url = "%s"`, o.ConnectionUrl))
	}
	if o.Table != "" {
		params = append(params, fmt.Sprintf(`table = "%s"`, o.Table))
	}
	if o.MaxParallel != 0 {
		params = append(params, fmt.Sprintf(`max_parallel = "%d"`, o.MaxParallel))
	}

	storageCfg := fmt.Sprintf(postgresqlStorageFmt, strings.Join(params, "\n"))
	return storageCfg, nil
}
