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

package util

import (
	"fmt"
	"path/filepath"

	vaultapi "github.com/hashicorp/vault/api"
	core "k8s.io/api/core/v1"
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
	return fmt.Sprintf(listenerFmt, filepath.Join(VaultTLSAssetDir, core.TLSCertKey), filepath.Join(VaultTLSAssetDir, core.TLSPrivateKeyKey))
}

// ListenerConfig creates tcp listener config
func GetListenerConfig() string {
	listenerCfg := fmt.Sprintf(listenerFmt,
		filepath.Join(VaultTLSAssetDir, core.TLSCertKey),
		filepath.Join(VaultTLSAssetDir, core.TLSPrivateKeyKey))

	return listenerCfg
}

func NewVaultClient(hostname string, port string, tlsConfig *vaultapi.TLSConfig) (*vaultapi.Client, error) {
	cfg := vaultapi.DefaultConfig()
	podURL := fmt.Sprintf("https://%s:%s", hostname, port)
	cfg.Address = podURL
	err := cfg.ConfigureTLS(tlsConfig)
	if err != nil {
		return nil, err
	}
	return vaultapi.NewClient(cfg)
}
