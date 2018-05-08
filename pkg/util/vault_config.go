package util

import (
	"fmt"
	"path/filepath"
	"strings"

	vaultapi "github.com/hashicorp/vault/api"
	api "github.com/soter/vault-operator/apis/vault/v1alpha1"
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

	// TLS related file name for etcd
	EtcdTLSAssetDir    = "/etc/vault/storage/etcd/tls/"
	EtcdClientCaName   = "etcd-client-ca.crt"
	EtcdClientCertName = "etcd-client.crt"
	EtcdClientKeyName  = "etcd-client.key"
)

var listenerFmt = `
listener "tcp" {
  address = "0.0.0.0:8200"
  cluster_address = "0.0.0.0:8201"
  tls_cert_file = "%s"
  tls_key_file  = "%s"
}
`
var inmenStorage = `
storage "inmem" {
}
`

var etcdStorageFmt = `
storage "etcd" {
%s
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

// GetStorageConfig creates storage config from BackendStorage Spec
func GetStorageConfig(s *api.BackendStorageSpec) (string, error) {
	if s.Inmem != nil {
		return inmenStorage, nil
	} else if s.Etcd != nil {
		return GetEtcdConfig(s.Etcd)
	}
	return "", nil
}

// vault doc: https://www.vaultproject.io/docs/configuration/storage/etcd.html
//
// Note:
// - Secret `TLSSecretName` mounted in `EtcdTLSAssetDir`
// - Secret `CredentialSecret` will be used as environment variable
//
// GetEtcdConfig creates etcd storage config from EtcdSpec
func GetEtcdConfig(s *api.EtcdSpec) (string, error) {
	params := []string{}
	if s.Address != "" {
		params = append(params, fmt.Sprintf(`address = "%s"`, s.Address))
	}
	if s.EtcdApi != "" {
		params = append(params, fmt.Sprintf(`etcd_api = "%s"`, s.EtcdApi))
	}
	if s.Path != "" {
		params = append(params, fmt.Sprintf(`path = "%s"`, s.Path))
	}
	if s.DiscoverySrv != "" {
		params = append(params, fmt.Sprintf(`discovery_srv = "%s"`, s.DiscoverySrv))
	}
	if s.HAEnable {
		params = append(params, fmt.Sprintf(`ha_enable = "true"`))
	} else {
		params = append(params, fmt.Sprintf(`ha_enable = "false"`))
	}
	if s.Sync {
		params = append(params, fmt.Sprintf(`sync = "true"`))
	} else {
		params = append(params, fmt.Sprintf(`sync = "false"`))
	}
	if s.TLSSecretName != "" {
		params = append(params, fmt.Sprintf(`tls-ca-file = "%s"`, filepath.Join(EtcdTLSAssetDir, EtcdClientCaName)),
			fmt.Sprintf(`tls-cert-file = "%s"`, filepath.Join(EtcdTLSAssetDir, EtcdClientCertName)),
			fmt.Sprintf(`tls-key-file = "%s"`, filepath.Join(EtcdTLSAssetDir, EtcdClientKeyName)))
	}

	storageCfg := fmt.Sprintf(etcdStorageFmt, strings.Join(params, "\n"))
	return storageCfg, nil
}

func NewVaultClient(hostname string, port string, tlsConfig *vaultapi.TLSConfig) (*vaultapi.Client, error) {
	cfg := vaultapi.DefaultConfig()
	podURL := fmt.Sprintf("https://%s:%s", hostname, port)
	cfg.Address = podURL
	cfg.ConfigureTLS(tlsConfig)
	return vaultapi.NewClient(cfg)
}
