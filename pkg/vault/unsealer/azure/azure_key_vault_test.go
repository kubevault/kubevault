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

package azure

import (
	"testing"

	api "kubevault.dev/apimachinery/apis/kubevault/v1alpha1"
	"kubevault.dev/operator/pkg/vault/util"

	"github.com/stretchr/testify/assert"
	core "k8s.io/api/core/v1"
)

func TestOptions_Apply(t *testing.T) {
	expected := []string{
		"--mode=azure-key-vault",
		"--azure.vault-base-url=vault.com",
		"--azure.tenant-id=1234",
		"--azure.cloud=TEST",
		"--azure.use-managed-identity=true",
		"--azure.client-cert-path=/etc/vault/unsealer/azure/cert/client.crt",
	}

	cont := core.Container{
		Name: util.VaultUnsealerContainerName,
	}
	pt := core.PodTemplateSpec{
		Spec: core.PodSpec{
			Containers: []core.Container{cont},
		},
	}

	opts, err := NewOptions(api.AzureKeyVault{
		VaultBaseURL:       "vault.com",
		TenantID:           "1234",
		Cloud:              "TEST",
		UseManagedIdentity: true,
		ClientCertSecret:   "s1",
		AADClientSecret:    "s2",
	})
	assert.Nil(t, err)

	err = opts.Apply(&pt)
	if assert.Nil(t, err) {
		assert.Equal(t, expected, pt.Spec.Containers[0].Args)
	}
}
