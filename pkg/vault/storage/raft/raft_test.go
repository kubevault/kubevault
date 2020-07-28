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
	"fmt"
	"testing"

	api "kubevault.dev/operator/apis/kubevault/v1alpha1"

	"github.com/stretchr/testify/assert"
	core "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestOptions_Apply(t *testing.T) {
	kfake := fake.NewSimpleClientset()

	three := int32(3)
	vaultServer := &api.VaultServer{
		Spec: api.VaultServerSpec{
			Replicas: &three,
		},
	}

	opts, err := NewOptions(kfake, vaultServer, api.RaftSpec{
		Path: "/test",
	})
	assert.Nil(t, err)

	pt := &core.PodTemplateSpec{
		Spec: core.PodSpec{
			Containers: []core.Container{
				{
					Name: "vault",
					Env: []core.EnvVar{
						{
							Name:  "VAULT_API_ADDR",
							Value: "https://$(POD_IP):8200",
						},
						{
							Name:  "VAULT_CLUSTER_ADDR",
							Value: "https://vault-0.vault-internal:8200",
						},
					},
				},
			},
		},
	}

	t.Run("raft storage config", func(t *testing.T) {
		err := opts.Apply(pt)
		assert.Nil(t, err)

		env := []core.EnvVar{
			{
				Name:  "VAULT_API_ADDR",
				Value: "https://$(POD_IP):8200",
			},
			{
				Name:  "VAULT_CLUSTER_ADDR",
				Value: "https://vault-0.vault-internal:8200",
			},
		}
		got := pt.Spec.Containers[0].Env
		if !assert.Equal(t, env, got) {
			fmt.Println("expected:", env)
			fmt.Println("got:", got)
		}
	})
}

func TestOptions_GetStorageConfig(t *testing.T) {
	kfake := fake.NewSimpleClientset()

	three := int32(3)
	vaultServer := &api.VaultServer{
		Spec: api.VaultServerSpec{
			Replicas: &three,
			TLS: &api.TLSPolicy{
				CABundle: []byte("DUMMY CACERT"),
			},
		},
	}

	opts, err := NewOptions(kfake, vaultServer, api.RaftSpec{
		Path: "/test",
	})
	assert.Nil(t, err)

	out := `
storage "raft" {
  path = "/test"
  retry_join {
    leader_api_addr         = "https://vault-0.vault-internal:8200"
    leader_ca_cert          = "DUMMY CACERT"
    leader_client_cert_file = "/etc/vault/tls/tls.crt"
    leader_client_key_file  = "/etc/vault/tls/tls.key"
  }
  retry_join {
    leader_api_addr         = "https://vault-1.vault-internal:8200"
    leader_ca_cert          = "DUMMY CACERT"
    leader_client_cert_file = "/etc/vault/tls/tls.crt"
    leader_client_key_file  = "/etc/vault/tls/tls.key"
  }
  retry_join {
    leader_api_addr         = "https://vault-2.vault-internal:8200"
    leader_ca_cert          = "DUMMY CACERT"
    leader_client_cert_file = "/etc/vault/tls/tls.crt"
    leader_client_key_file  = "/etc/vault/tls/tls.key"
  }
}
`

	t.Run("raft storage config", func(t *testing.T) {
		got, err := opts.GetStorageConfig()
		assert.Nil(t, err)

		if !assert.Equal(t, out, got) {
			fmt.Println("expected:", out)
			fmt.Println("got:", got)
		}
	})
}
