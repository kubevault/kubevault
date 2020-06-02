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

package mysql

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	api "kubevault.dev/operator/apis/kubevault/v1alpha1"

	"github.com/pkg/errors"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const (
	MySQLTLSCAFile = "/etc/vault/mysql/certs/ca.crt"
)

var mysqlStorageFmt = `
storage "mysql" {
%s
}
`

type Options struct {
	api.MySQLSpec
	Username string
	Password string
}

func NewOptions(kubeClient kubernetes.Interface, namespace string, s api.MySQLSpec) (*Options, error) {
	var (
		username string
		password string
	)

	if s.UserCredentialSecret != "" {
		sr, err := kubeClient.CoreV1().Secrets(namespace).Get(context.TODO(), s.UserCredentialSecret, metav1.GetOptions{})
		if err != nil {
			return nil, errors.Wrapf(err, "failed to get user credential secret(%s)", s.UserCredentialSecret)
		}

		if value, ok := sr.Data["username"]; ok {
			username = string(value)
		} else {
			return nil, errors.Errorf("username not found in secret(%s)", s.UserCredentialSecret)
		}

		if value, ok := sr.Data["password"]; ok {
			password = string(value)
		} else {
			return nil, errors.Errorf("password not found in secret(%s)", s.UserCredentialSecret)
		}
	}

	return &Options{
		s,
		username,
		password,
	}, nil
}

func (o *Options) Apply(pt *core.PodTemplateSpec) error {
	if o.TLSCASecret != "" {
		pt.Spec.Volumes = append(pt.Spec.Volumes, core.Volume{
			Name: "mysql-tls",
			VolumeSource: core.VolumeSource{
				Secret: &core.SecretVolumeSource{
					SecretName: o.TLSCASecret,
					Items: []core.KeyToPath{
						{
							Key:  "tls_ca_file",
							Path: "ca.crt",
						},
					},
				},
			},
		})

		pt.Spec.Containers[0].VolumeMounts = append(pt.Spec.Containers[0].VolumeMounts, core.VolumeMount{
			Name:      "mysql-tls",
			MountPath: filepath.Dir(MySQLTLSCAFile),
		})
	}

	return nil
}

// vault doc: https://www.vaultproject.io/docs/configuration/storage/google-cloud-storage.html
//
// GetGcsConfig creates gcs storae config from GcsSpec
func (o *Options) GetStorageConfig() (string, error) {
	params := []string{}
	if o.Address != "" {
		params = append(params, fmt.Sprintf(`address = "%s"`, o.Address))
	}
	if o.Database != "" {
		params = append(params, fmt.Sprintf(`database = "%s"`, o.Database))
	}
	if o.Table != "" {
		params = append(params, fmt.Sprintf(`table = "%s"`, o.Table))
	}
	if o.TLSCASecret != "" {
		params = append(params, fmt.Sprintf(`tls_ca_file = "%s"`, MySQLTLSCAFile))
	}
	if o.Username != "" {
		params = append(params, fmt.Sprintf(`username = "%s"`, o.Username))
	}
	if o.Password != "" {
		params = append(params, fmt.Sprintf(`password = "%s"`, o.Password))
	}
	if o.MaxParallel != 0 {
		params = append(params, fmt.Sprintf(`max_parallel = "%d"`, o.MaxParallel))
	}

	storageCfg := fmt.Sprintf(mysqlStorageFmt, strings.Join(params, "\n"))
	return storageCfg, nil
}
