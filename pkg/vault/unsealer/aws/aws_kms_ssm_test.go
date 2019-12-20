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

package aws

import (
	"testing"

	api "kubevault.dev/operator/apis/kubevault/v1alpha1"
	"kubevault.dev/operator/pkg/vault/util"

	"github.com/stretchr/testify/assert"
	core "k8s.io/api/core/v1"
)

func TestOptions_Apply(t *testing.T) {
	expected := []string{
		"--mode=aws-kms-ssm",
		"--aws.kms-key-id=test-key",
		"--aws.ssm-key-prefix=/cluster/demo",
	}
	cont := core.Container{
		Name: util.VaultUnsealerContainerName,
	}
	pt := &core.PodTemplateSpec{
		Spec: core.PodSpec{
			Containers: []core.Container{cont},
		},
	}

	opts, err := NewOptions(api.AwsKmsSsmSpec{
		KmsKeyID:     "test-key",
		SsmKeyPrefix: "/cluster/demo",
	})
	assert.Nil(t, err)

	err = opts.Apply(pt)
	if assert.Nil(t, err) {
		assert.Equal(t, expected, pt.Spec.Containers[0].Args)
	}
}
