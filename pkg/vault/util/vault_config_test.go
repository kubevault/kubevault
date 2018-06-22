package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetListenerConfig(t *testing.T) {
	expectedOutput := `
listener "tcp" {
  address = "0.0.0.0:8200"
  cluster_address = "0.0.0.0:8201"
  tls_cert_file = "/etc/vault/tls/server.crt"
  tls_key_file  = "/etc/vault/tls/server.key"
}
`
	assert.Equal(t, expectedOutput, GetListenerConfig())
}
