package util

import (
	"bytes"
	"fmt"
	"path/filepath"

	vaultapi "github.com/hashicorp/vault/api"
	corev1 "k8s.io/api/core/v1"
)

const (
	// VaultConfigPath is the path that vault pod uses to read config from
	VaultConfigPath = "/run/vault/config/vault.hcl"

	// VaultTLSAssetDir is the dir where vault's server TLS and etcd TLS assets sits
	VaultTLSAssetDir = "/run/vault/tls/"

	// ServerTLSCertName is the filename of the vault server cert
	ServerTLSCertName = "server.crt"

	// ServerTLSKeyName is the filename of the vault server key
	ServerTLSKeyName = "server.key"

	// TLS related file name for etcd
	EtcdClientCaName   = "etcd-client-ca.crt"
	EtcdClientCertName = "etcd-client.crt"
	EtcdClientKeyName  = "etcd-client.key"
)

var listenerFmt = `
listener "tcp" {
  address     = "0.0.0.0:8200"
  cluster_address = "0.0.0.0:8201"
  tls_cert_file = "%s"
  tls_key_file  = "%s"
}
`

var etcdStorageFmt = `
storage "etcd" {
  address = "%s"
  etcd_api = "v3"
  ha_enabled = "true"
  tls_ca_file = "%s"
  tls_cert_file = "%s"
  tls_key_file = "%s"
  sync = "false"
}
`

// NewConfigWithDefaultParams appends to given config data some default params:
// - tcp listener
func NewConfigWithDefaultParams() string {
	return fmt.Sprintf(listenerFmt, filepath.Join(VaultTLSAssetDir, ServerTLSCertName), filepath.Join(VaultTLSAssetDir, ServerTLSKeyName))
}

// AppenListenerInConfig will append tcp listener in given config
func AppendListenerInConfig(data string) string {
	buf := bytes.NewBufferString(data)

	// TODO : telemetry
	/*buf.WriteString(`
	telemetry {
		statsd_address = "localhost:9125"
	}
	`)*/

	listenerSection := fmt.Sprintf(listenerFmt,
		filepath.Join(VaultTLSAssetDir, ServerTLSCertName),
		filepath.Join(VaultTLSAssetDir, ServerTLSKeyName))
	buf.WriteString(listenerSection)

	return buf.String()
}

// NewConfigWithEtcd returns the new config data combining
// original config and new etcd storage section.
func NewConfigWithEtcd(data, etcdURL string) string {
	storageSection := fmt.Sprintf(etcdStorageFmt, etcdURL, filepath.Join(VaultTLSAssetDir, EtcdClientCaName),
		filepath.Join(VaultTLSAssetDir, EtcdClientCertName), filepath.Join(VaultTLSAssetDir, EtcdClientKeyName))
	data = fmt.Sprintf("%s%s", data, storageSection)
	return data
}

func NewConfigFormConfigMap(data string, s *corev1.ConfigMap) string {
	storageSection := fmt.Sprintf(`
storage "%s" {`, s.Data["name"])

	for k, v := range s.Data {
		if k != "name" {
			storageSection = storageSection + fmt.Sprintf(`
%s = "%s"`, k, v)
		}
	}

	storageSection += "\n}"
	data = fmt.Sprintf("%s%s", data, storageSection)
	return data
}

func NewVaultClient(hostname string, port string, tlsConfig *vaultapi.TLSConfig) (*vaultapi.Client, error) {
	cfg := vaultapi.DefaultConfig()
	podURL := fmt.Sprintf("https://%s:%s", hostname, port)
	cfg.Address = podURL
	cfg.ConfigureTLS(tlsConfig)
	return vaultapi.NewClient(cfg)
}
