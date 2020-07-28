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

package raft

import (
	"bytes"
	"text/template"

	api "kubevault.dev/operator/apis/kubevault/v1alpha1"

	"github.com/pkg/errors"
	core "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	core_util "kmodules.xyz/client-go/core/v1"
)

// Options represents the instance of the Raft storage.
type Options struct {
	api.RaftSpec
	Replicas int32
}

const (
	// VaultRaftVolumeName is the name of the backend volume created.
	VaultRaftVolumeName = "vault-raft-backend"

	// configTemplate is the template used to produce the Vault configuration
	configTemplate = `
storage "raft" {
  path = "{{ .RaftSpec.Path }}"
{{ range $i, $_ := (iter .Replicas) }}
  retry_join {
    leader_api_addr         = "https://vault-{{ $i }}.vault-internal:8200"
    leader_ca_cert_file     = "/etc/vault/tls/cacert.crt"
    leader_client_cert_file = "/etc/vault/tls/tls.crt"
    leader_client_key_file  = "/etc/vault/tls/tls.key"
  }
{{ end -}}
}
`
)

// NewOptions instanciate the Raft storage.
func NewOptions(kubeClient kubernetes.Interface, vaultServer *api.VaultServer, rs api.RaftSpec) (*Options, error) {
	o := &Options{
		RaftSpec: rs,
		Replicas: 1,
	}

	if vaultServer.Spec.Replicas != nil {
		o.Replicas = *vaultServer.Spec.Replicas
	}

	return o, nil
}

// Apply ...
func (o *Options) Apply(pt *core.PodTemplateSpec) error {
	if o.Path == "" {
		return errors.New("path is empty")
	}

	// Change the environments variables
	pt.Spec.Containers[0].Env = core_util.UpsertEnvVars(
		pt.Spec.Containers[0].Env,
		core.EnvVar{
			Name:  "VAULT_API_ADDR",
			Value: "https://$(POD_IP):8200",
		},
		core.EnvVar{
			Name:  "VAULT_CLUSTER_ADDR",
			Value: "https://$(HOSTNAME).vault-internal:8200",
		},
	)

	// Configure the volume
	pt.Spec.Volumes = append(pt.Spec.Volumes, core.Volume{
		Name: VaultRaftVolumeName,
		VolumeSource: core.VolumeSource{
			EmptyDir: &core.EmptyDirVolumeSource{},
		},
	})

	pt.Spec.Containers[0].VolumeMounts = append(pt.Spec.Containers[0].VolumeMounts, core.VolumeMount{
		Name:      VaultRaftVolumeName,
		MountPath: o.Path,
	})

	return nil
}

// GetStorageConfig ...
func (o *Options) GetStorageConfig() (string, error) {

	// https://github.com/bradfitz/iter
	t := template.New("config").Funcs(map[string]interface{}{
		"iter": func(n int32) []struct{} {
			return make([]struct{}, n)
		},
	})

	if _, err := t.Parse(configTemplate); err != nil {
		return "", errors.Wrap(err, "compile storage template failed")
	}

	buf := bytes.NewBuffer([]byte{})
	if err := t.Execute(buf, o); err != nil {
		return "", errors.Wrap(err, "execute storage template failed")
	}

	return buf.String(), nil
}
