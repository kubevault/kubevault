package consul

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	api "github.com/kubevault/operator/apis/kubevault/v1alpha1"
	"github.com/stretchr/testify/assert"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	masterURL  string
	kubeconfig string
)

func TestGetEtcdConfig(t *testing.T) {

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
		TlsMinVersion:       "tls12",
		TlsSkipVerify:       false,
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
tls_ca_file = "/etc/vault/storage/consul/tls/consul-ca.crt"
tls_cert_file = "/etc/vault/storage/consul/tls/consul-client.crt"
tls_key_file = "/etc/vault/storage/consul/tls/consul-client.key"
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

	kubeconfig = filepath.Join(os.Getenv("HOME"), ".kube/config")

	cfg, err := clientcmd.BuildConfigFromFlags(masterURL, kubeconfig)
	if err != nil {
		panic(err)

	}

	kubeClient, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		panic(err)

	}

	for _, test := range testCase {
		t.Run(test.testName, func(t *testing.T) {
			etcd, err := NewOptions(kubeClient, "default", *test.consulSpec)
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
