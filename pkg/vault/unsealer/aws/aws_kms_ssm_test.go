package aws

import (
	"testing"

	api "github.com/kubevault/operator/apis/core/v1alpha1"
	"github.com/kubevault/operator/pkg/vault/util"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
)

func TestOptions_Apply(t *testing.T) {
	expected := []string{
		"--mode=aws-kms-ssm",
		"--aws.kms-key-id=test-key",
	}
	cont := corev1.Container{
		Name: util.VaultUnsealerContainerName(),
	}
	pt := &corev1.PodTemplateSpec{
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{cont},
		},
	}

	opts, err := NewOptions(api.AwsKmsSsmSpec{
		KmsKeyID: "test-key",
	})
	assert.Nil(t, err)

	err = opts.Apply(pt)
	if assert.Nil(t, err) {
		assert.Equal(t, expected, pt.Spec.Containers[0].Args)
	}
}
