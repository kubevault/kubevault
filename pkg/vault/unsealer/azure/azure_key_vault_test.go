package azure

import (
	"testing"

	api "github.com/kubevault/operator/apis/core/v1alpha1"
	"github.com/kubevault/operator/pkg/vault/util"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
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

	cont := corev1.Container{
		Name: util.VaultUnsealerImageName(),
	}
	pt := corev1.PodTemplateSpec{
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{cont},
		},
	}

	opts, err := NewOptions(api.AzureKeyVault{
		VaultBaseUrl:       "vault.com",
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
