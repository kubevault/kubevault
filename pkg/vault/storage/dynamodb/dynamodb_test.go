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
package dynamodb

import (
	"fmt"
	"testing"

	api "kubevault.dev/operator/apis/kubevault/v1alpha1"

	"github.com/stretchr/testify/assert"
)

func TestOptions_GetStorageConfig(t *testing.T) {
	opts, err := NewOptions(api.DynamoDBSpec{
		EndPoint:      "endpoint",
		HaEnabled:     true,
		MaxParallel:   128,
		Region:        "us",
		ReadCapacity:  1,
		WriteCapacity: 1,
		Table:         "vault",
	})
	assert.Nil(t, err)

	out := `
storage "dynamodb" {
endpoint = "endpoint"
ha_enabled = "true"
region = "us"
read_capacity = 1
write_capacity = 1
table = "vault"
max_parallel = 128
}
`
	t.Run("DynamoDB storage config", func(t *testing.T) {
		got, err := opts.GetStorageConfig()
		assert.Nil(t, err)
		if !assert.Equal(t, out, got) {
			fmt.Println("expected:", out)
			fmt.Println("got:", got)
		}
	})
}
