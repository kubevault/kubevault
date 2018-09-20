package kubernetes

import (
	"testing"

	api "github.com/kubevault/operator/apis/kubevault/v1alpha1"
	"github.com/kubevault/operator/pkg/vault/util"
	"github.com/stretchr/testify/assert"
	core "k8s.io/api/core/v1"
)

func TestOptions_Apply(t *testing.T) {
	expected := []string{
		"--mode=kubernetes-secret",
		"--k8s.secret-name=test",
	}
	cont := core.Container{
		Name: util.VaultUnsealerContainerName,
	}
	pt := core.PodTemplateSpec{
		Spec: core.PodSpec{
			Containers: []core.Container{cont},
		},
	}

	opts, err := NewOptions(api.KubernetesSecretSpec{
		"test",
	})
	assert.Nil(t, err)

	err = opts.Apply(&pt)
	assert.Nil(t, err)

	assert.Equal(t, expected, pt.Spec.Containers[0].Args)
}
