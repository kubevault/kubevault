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

package dynamodb

import (
	"fmt"
	"strings"

	api "kubevault.dev/operator/apis/kubevault/v1alpha1"

	core "k8s.io/api/core/v1"
)

var dynamodbStorageFmt = `
storage "dynamodb" {
%s
}
`

type Options struct {
	api.DynamoDBSpec
}

func NewOptions(s api.DynamoDBSpec) (*Options, error) {
	return &Options{
		s,
	}, nil
}

func (o *Options) Apply(pt *core.PodTemplateSpec) error {
	envs := []core.EnvVar{}

	if o.CredentialSecret != "" {
		envs = append(envs, core.EnvVar{
			Name: "AWS_ACCESS_KEY_ID",
			ValueFrom: &core.EnvVarSource{
				SecretKeyRef: &core.SecretKeySelector{
					LocalObjectReference: core.LocalObjectReference{
						Name: o.CredentialSecret,
					},
					Key: "access_key",
				},
			},
		}, core.EnvVar{
			Name: "AWS_SECRET_ACCESS_KEY",
			ValueFrom: &core.EnvVarSource{
				SecretKeyRef: &core.SecretKeySelector{
					LocalObjectReference: core.LocalObjectReference{
						Name: o.CredentialSecret,
					},
					Key: "secret_key",
				},
			},
		})
	}
	if o.SessionTokenSecret != "" {
		envs = append(envs, core.EnvVar{
			Name: "AWS_SESSION_TOKEN",
			ValueFrom: &core.EnvVarSource{
				SecretKeyRef: &core.SecretKeySelector{
					LocalObjectReference: core.LocalObjectReference{
						Name: o.SessionTokenSecret,
					},
					Key: "session_token",
				},
			},
		})
	}

	pt.Spec.Containers[0].Env = append(pt.Spec.Containers[0].Env, envs...)
	return nil
}

// vault doc: https://www.vaultproject.io/docs/configuration/storage/google-cloud-storage.html
//
// GetStorageConfig creates gcs storae config from GcsSpec
func (o *Options) GetStorageConfig() (string, error) {
	params := []string{}
	if o.EndPoint != "" {
		params = append(params, fmt.Sprintf(`endpoint = "%s"`, o.EndPoint))
	}
	if o.HaEnabled {
		params = append(params, fmt.Sprintf(`ha_enabled = "true"`))
	}
	if o.Region != "" {
		params = append(params, fmt.Sprintf(`region = "%s"`, o.Region))
	}
	if o.ReadCapacity != 0 {
		params = append(params, fmt.Sprintf(`read_capacity = %d`, o.ReadCapacity))
	}
	if o.WriteCapacity != 0 {
		params = append(params, fmt.Sprintf(`write_capacity = %d`, o.WriteCapacity))
	}
	if o.Table != "" {
		params = append(params, fmt.Sprintf(`table = "%s"`, o.Table))
	}
	if o.MaxParallel != 0 {
		params = append(params, fmt.Sprintf(`max_parallel = %d`, o.MaxParallel))
	}

	storageCfg := fmt.Sprintf(dynamodbStorageFmt, strings.Join(params, "\n"))
	return storageCfg, nil
}
