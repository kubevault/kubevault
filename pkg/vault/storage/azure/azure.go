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
package azure

import (
	"fmt"
	"strings"

	api "kubevault.dev/operator/apis/kubevault/v1alpha1"

	core "k8s.io/api/core/v1"
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

func (o *Options) Apply(pt *core.PodTemplateSpec) error {
	if o.AccountKeySecret != "" {
		pt.Spec.Containers[0].Env = append(pt.Spec.Containers[0].Env, core.EnvVar{
			Name: AzureAccountKeyEnv,
			ValueFrom: &core.EnvVarSource{
				SecretKeyRef: &core.SecretKeySelector{
					LocalObjectReference: core.LocalObjectReference{
						Name: o.AccountKeySecret,
					},
					Key: "account_key",
				},
			},
		})
	}
	return nil
}

func (o *Options) GetSecrets(namespace string) ([]core.Secret, error) {
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
