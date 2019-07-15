package google

import (
	"testing"

	"github.com/stretchr/testify/assert"
	core "k8s.io/api/core/v1"
	api "kubevault.dev/operator/apis/kubevault/v1alpha1"
	"kubevault.dev/operator/pkg/vault/util"
)

func TestOptions_Apply(t *testing.T) {
	expected := []string{
		"--mode=google-cloud-kms-gcs",
		"--google.storage-bucket=test",
		"--google.kms-project=pro",
		"--google.kms-location=loc",
		"--google.kms-key-ring=r-key",
		"--google.kms-crypto-key=c-key",
	}
	cont := core.Container{
		Name: util.VaultUnsealerContainerName,
	}
	pt := &core.PodTemplateSpec{
		Spec: core.PodSpec{
			Containers: []core.Container{cont},
		},
	}

	opts, err := NewOptions(api.GoogleKmsGcsSpec{
		KmsCryptoKey: "c-key",
		KmsKeyRing:   "r-key",
		KmsLocation:  "loc",
		KmsProject:   "pro",
		Bucket:       "test",
	})
	assert.Nil(t, err)

	err = opts.Apply(pt)
	assert.Nil(t, err)

	assert.Equal(t, expected, pt.Spec.Containers[0].Args)
}
