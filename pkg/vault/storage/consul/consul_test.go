package consul

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	core "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kfake "k8s.io/client-go/kubernetes/fake"
	api "kubevault.dev/operator/apis/kubevault/v1alpha1"
)

const (
	aclCredName     = "aclcred"
	aclToken        = "aclToken"
	secretNamespace = "default"
)

func TestGetConsulConfig(t *testing.T) {

	spec1 := &api.ConsulSpec{
		Address: "127.0.0.1:8500",
		Path:    "vault",
	}
	spec2 := &api.ConsulSpec{
		Address:             "localhost:3333",
		CheckTimeout:        "30",
		ConsistencyMode:     "strong",
		DisableRegistration: "false",
		MaxParallel:         "130",
		Path:                "vault",
		Scheme:              "http",
		Service:             "vault",
		ServiceTags:         "dev,aud",
		ServiceAddress:      "",
		ACLTokenSecretName:  "aclcred",
		SessionTTL:          "20s",
		LockWaitTime:        "25s",
		TLSSecretName:       "TLSCred",
		TLSMinVersion:       "tls12",
		TLSSkipVerify:       false,
	}
	out1 := `
storage "consul" {
address = "127.0.0.1:8500"
path = "vault"
}
`
	out2 := `
storage "consul" {
address = "localhost:3333"
check_timeout = "30"
consistency_mode = "strong"
disable_registration = "false"
max_parallel = "130"
path = "vault"
scheme = "http"
service = "vault"
service_tags = "dev,aud"
token = "data"
session_ttl = "20s"
lock_wait_time = "25s"
tls_ca_file = "/etc/vault/storage/consul/tls/ca.crt"
tls_cert_file = "/etc/vault/storage/consul/tls/client.crt"
tls_key_file = "/etc/vault/storage/consul/tls/client.key"
tls_min_version = "tls12"
}
`

	testCase := []struct {
		testName       string
		consulSpec     *api.ConsulSpec
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

	kubeClient := kfake.NewSimpleClientset(&core.Secret{
		ObjectMeta: v1.ObjectMeta{
			Name:      aclCredName,
			Namespace: secretNamespace,
		},
		Data: map[string][]byte{
			aclToken: []byte("data"),
		},
	})

	for _, test := range testCase {
		t.Run(test.testName, func(t *testing.T) {
			consul, err := NewOptions(kubeClient, "default", *test.consulSpec)
			assert.Nil(t, err)

			config, err := consul.GetStorageConfig()
			assert.Nil(t, err)
			if !assert.Equal(t, test.expectedOutput, config) {
				fmt.Println("expected:", test.expectedOutput)
				fmt.Println("got:", config)
			}
		})
	}
}
