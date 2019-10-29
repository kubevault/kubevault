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
	"path/filepath"

	api "kubevault.dev/operator/apis/kubevault/v1alpha1"
	"kubevault.dev/operator/pkg/vault/util"

	"github.com/pkg/errors"
	core "k8s.io/api/core/v1"
	rbac "k8s.io/api/rbac/v1"
	core_util "kmodules.xyz/client-go/core/v1"
)

const (
	ModeAzureKeyVault = "azure-key-vault"
)

type Options struct {
	api.AzureKeyVault
}

func NewOptions(s api.AzureKeyVault) (*Options, error) {
	return &Options{
		s,
	}, nil
}

func (o *Options) Apply(pt *core.PodTemplateSpec) error {
	if pt == nil {
		return errors.New("podTempleSpec is nil")
	}

	var args []string
	var cont core.Container

	for _, c := range pt.Spec.Containers {
		if c.Name == util.VaultUnsealerContainerName {
			cont = c
		}
	}

	args = append(args, fmt.Sprintf("--mode=%s", ModeAzureKeyVault))
	if o.VaultBaseUrl != "" {
		args = append(args, fmt.Sprintf("--azure.vault-base-url=%s", o.VaultBaseUrl))
	}
	if o.TenantID != "" {
		args = append(args, fmt.Sprintf("--azure.tenant-id=%s", o.TenantID))
	}
	if o.Cloud != "" {
		args = append(args, fmt.Sprintf("--azure.cloud=%s", o.Cloud))
	}
	if o.UseManagedIdentity {
		args = append(args, fmt.Sprintf("--azure.use-managed-identity=true"))
	}

	var envs []core.EnvVar

	if o.AADClientSecret != "" {
		envs = append(envs, core.EnvVar{
			Name: "AZURE_CLIENT_ID",
			ValueFrom: &core.EnvVarSource{
				SecretKeyRef: &core.SecretKeySelector{
					LocalObjectReference: core.LocalObjectReference{
						Name: o.AADClientSecret,
					},
					Key: "client-id",
				},
			},
		}, core.EnvVar{
			Name: "AZURE_CLIENT_SECRET",
			ValueFrom: &core.EnvVarSource{
				SecretKeyRef: &core.SecretKeySelector{
					LocalObjectReference: core.LocalObjectReference{
						Name: o.AADClientSecret,
					},
					Key: "client-secret",
				},
			},
		})
	}

	if o.ClientCertSecret != "" {
		envs = append(envs, core.EnvVar{
			Name: "AZURE_CLIENT_CERT_PASSWORD",
			ValueFrom: &core.EnvVarSource{
				SecretKeyRef: &core.SecretKeySelector{
					LocalObjectReference: core.LocalObjectReference{
						Name: o.ClientCertSecret,
					},
					Key: "client-cert-password",
				},
			},
		})

		volumeName := "azure-client-cert"

		pt.Spec.Volumes = core_util.UpsertVolume(pt.Spec.Volumes, core.Volume{
			Name: volumeName,
			VolumeSource: core.VolumeSource{
				Secret: &core.SecretVolumeSource{
					SecretName: o.ClientCertSecret,
					Items: []core.KeyToPath{
						{
							Key:  "client-cert",
							Path: "client.crt",
						},
					},
				},
			},
		})

		certFilePath := "/etc/vault/unsealer/azure/cert/client.crt"

		cont.VolumeMounts = core_util.UpsertVolumeMount(cont.VolumeMounts, core.VolumeMount{
			Name:      volumeName,
			MountPath: filepath.Dir(certFilePath),
		})

		args = append(args, fmt.Sprintf("--azure.client-cert-path=%s", certFilePath))
	}

	cont.Args = append(cont.Args, args...)
	cont.Env = core_util.UpsertEnvVars(cont.Env, envs...)
	pt.Spec.Containers = core_util.UpsertContainer(pt.Spec.Containers, cont)
	return nil
}

// GetRBAC returns required rbac roles
func (o *Options) GetRBAC(prefix, namespace string) []rbac.Role {
	return nil
}
