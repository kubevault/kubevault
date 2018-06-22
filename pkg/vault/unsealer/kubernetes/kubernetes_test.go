package kubernetes

import (
	"testing"

	api "github.com/kubevault/operator/apis/core/v1alpha1"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
)

func TestOptions_Apply(t *testing.T) {
	expected := []string{
		"--mode=kubernetes-secret",
		"--k8s.secret-name=test",
	}
	cont := corev1.Container{}

	opts, err := NewOptions(api.KubernetesSecretSpec{
		"test",
	})
	assert.Nil(t, err)

	err = opts.Apply(nil, &cont)
	assert.Nil(t, err)

	assert.Equal(t, expected, cont.Args)
}
