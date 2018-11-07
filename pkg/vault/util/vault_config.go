package util

import (
	"fmt"
	"path/filepath"

	vaultapi "github.com/hashicorp/vault/api"
)

const (
	VaultContainerName         = "vault"
	VaultUnsealerContainerName = "vault-unsealer"
	VaultInitContainerName     = "vault-config"
	VaultExporterContainerName = "vault-exporter"
)

const (
	// VaultConfigFile is the file that vault pod uses to read config from
	VaultConfigFile = "/etc/vault/config/vault.hcl"

	// VaultTLSAssetDir is the dir where vault's server TLS sits
	VaultTLSAssetDir = "/etc/vault/tls/"

	// ServerTLSCertName is the filename of the vault server cert
	ServerTLSCertName = "server.crt"

	// ServerTLSKeyName is the filename of the vault server key
	ServerTLSKeyName = "server.key"
)

var listenerFmt = `
listener "tcp" {
  address = "0.0.0.0:8200"
  cluster_address = "0.0.0.0:8201"
  tls_cert_file = "%s"
  tls_key_file  = "%s"
}
`

// NewConfigWithDefaultParams appends to given config data some default params:
// - tcp listener
func NewConfigWithDefaultParams() string {
	return fmt.Sprintf(listenerFmt, filepath.Join(VaultTLSAssetDir, ServerTLSCertName), filepath.Join(VaultTLSAssetDir, ServerTLSKeyName))
}

// ListenerConfig creates tcp listener config
func GetListenerConfig() string {
	listenerCfg := fmt.Sprintf(listenerFmt,
		filepath.Join(VaultTLSAssetDir, ServerTLSCertName),
		filepath.Join(VaultTLSAssetDir, ServerTLSKeyName))

	return listenerCfg
}

func NewVaultClient(hostname string, port string, tlsConfig *vaultapi.TLSConfig) (*vaultapi.Client, error) {
	cfg := vaultapi.DefaultConfig()
	podURL := fmt.Sprintf("https://%s:%s", hostname, port)
	cfg.Address = podURL
	cfg.ConfigureTLS(tlsConfig)
	return vaultapi.NewClient(cfg)
}
