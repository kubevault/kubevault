package etcd

import (
	"testing"

	api "github.com/kube-vault/operator/apis/core/v1alpha1"
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
ha_enable = "true"
sync = "true"
}
`
	out2 := `
storage "etcd" {
address = "localhost:2379"
etcd_api = "v3"
path = "path/"
discovery_srv = "etcd.com"
ha_enable = "false"
sync = "false"
tls-ca-file = "/etc/vault/storage/etcd/tls/etcd-client-ca.crt"
tls-cert-file = "/etc/vault/storage/etcd/tls/etcd-client.crt"
tls-key-file = "/etc/vault/storage/etcd/tls/etcd-client.key"
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
			assert.Equal(t, test.expectedOutput, config)
		})
	}
}
