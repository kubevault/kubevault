/*
Copyright AppsCode Inc. and Contributors

Licensed under the AppsCode Community License 1.0.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    https://github.com/appscode/licenses/raw/1.0.0/AppsCode-Community-1.0.0.md

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

	api "kubevault.dev/apimachinery/apis/kubevault/v1alpha1"

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
			Backend: api.BackendStorageSpec{
				Raft: &api.RaftSpec{
					Path:      "/test",
					RetryJoin: []api.RetryJoinSpec{},
				},
			},
		},
	}

	opts, err := NewOptions(kfake, vaultServer)
	assert.Nil(t, err)

	pt := &core.PodTemplateSpec{
		Spec: core.PodSpec{
			Containers: []core.Container{
				{
					Name: "vault",
					Env: []core.EnvVar{
						{
							Name:  "VAULT_API_ADDR",
							Value: "https://vault.default.svc:8200",
						},
						{
							Name:  "VAULT_CLUSTER_ADDR",
							Value: "https://vault.default.svc:8201",
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
				Value: "https://$(HOSTNAME).vault-internal:8201",
			},
		}
		got := pt.Spec.Containers[0].Env[:2]
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
			Backend: api.BackendStorageSpec{
				Raft: &api.RaftSpec{
					Path:      "/test",
					RetryJoin: []api.RetryJoinSpec{},
				},
			},
		},
	}

	opts, err := NewOptions(kfake, vaultServer)
	assert.Nil(t, err)

	out := `
storage "raft" {
path = "/test"

retry_join {
 leader_api_addr = "http://-0.internal..svc:8200"
}


retry_join {
 leader_api_addr = "http://-1.internal..svc:8200"
}


retry_join {
 leader_api_addr = "http://-2.internal..svc:8200"
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
