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
)

// Options represents the instance of the Raft storage.
type Options struct {
	api.RaftSpec
	Replicas int32
}

// configTemplate is the template used to produce the Vault configuration
const configTemplate = `
storage "raft" {
  path = "{{ .RaftSpec.Path }}"

  {{- for $n := range (iter .Replicas) }}
  retry_join {
    leader_api_addr = "http://vault-{{ $n }}.vault-internal:8200"
  }
  {{- end }}
}
`

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
	return errors.New("not implemented error")
}

// GetStorageConfig ...
func (o *Options) GetStorageConfig() (string, error) {
	t, err := template.New("config").Parse(configTemplate)
	if err != nil {
		return "", errors.Wrap(err, "compile storage template failed")
	}

	// https://github.com/bradfitz/iter
	t = t.Funcs(map[string]interface{}{
		"iter": func(n int) []struct{} {
			return make([]struct{}, n)
		},
	})

	buf := bytes.NewBuffer(make([]byte, 1024))
	if err := t.Execute(buf, o); err != nil {
		return "", errors.Wrap(err, "execute storage template failed")
	}

	return buf.String(), nil
}
