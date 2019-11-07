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

package file

import (
	"fmt"
	"strings"

	api "kubevault.dev/operator/apis/kubevault/v1alpha1"

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
