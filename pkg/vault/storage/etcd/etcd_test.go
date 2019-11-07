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

package etcd

import (
	"fmt"
	"testing"

	api "kubevault.dev/operator/apis/kubevault/v1alpha1"

	"github.com/stretchr/testify/assert"
)

func TestGetEtcdConfig(t *testing.T) {

	spec1 := &api.EtcdSpec{
		Address:  "123",
		HAEnable: true,
		Sync:     true,
	}
	spec2 := &api.EtcdSpec{
		Address:              "localhost:2379",
		EtcdApi:              "v3",
		HAEnable:             false,
		Sync:                 false,
		Path:                 "path/",
		DiscoverySrv:         "etcd.com",
		TLSSecretName:        "tls",
		CredentialSecretName: "cred",
	}
	out1 := `
storage "etcd" {
address = "123"
ha_enabled = "true"
sync = "true"
}
`
	out2 := `
storage "etcd" {
address = "localhost:2379"
etcd_api = "v3"
path = "path/"
discovery_srv = "etcd.com"
ha_enabled = "false"
sync = "false"
tls_ca_file = "/etc/vault/storage/etcd/tls/ca.crt"
tls_cert_file = "/etc/vault/storage/etcd/tls/client.crt"
tls_key_file = "/etc/vault/storage/etcd/tls/client.key"
}
`

	testaData := []struct {
		testName       string
		etcdSpec       *api.EtcdSpec
		expectedOutput string
	}{
		{
			"Some fields are not defined",
			spec1,
			out1,
		},
		{
			"All fields are defined",
			spec2,
			out2,
		},
	}

	for _, test := range testaData {
		t.Run(test.testName, func(t *testing.T) {
			etcd, err := NewOptions(*test.etcdSpec)
			assert.Nil(t, err)

			config, err := etcd.GetStorageConfig()
			assert.Nil(t, err)
			if !assert.Equal(t, test.expectedOutput, config) {
				fmt.Println("expected:", test.expectedOutput)
				fmt.Println("got:", config)
			}
		})
	}
}
