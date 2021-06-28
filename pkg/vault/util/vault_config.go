/*
Copyright AppsCode Inc. and Contributors

Licensed under the AppsCode Community License 1.0.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    https://github.com/appscode/licenses/raw/1.0.0/AppsCode-Community-1.0.0.md

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
	"strings"

	conapi "kubevault.dev/apimachinery/apis"

	vaultapi "github.com/hashicorp/vault/api"
	core "k8s.io/api/core/v1"
	"k8s.io/klog/v2"
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
)

var listenerFmt = `
listener "tcp" {
  address = "0.0.0.0:8200"
  cluster_address = "0.0.0.0:8201"
  %s
}
`

// ListenerConfig creates tcp listener config
func GetListenerConfig(mountPath string, isTLSEnabled bool) string {
	var params []string
	if isTLSEnabled {
		params = append(params, fmt.Sprintf(`tls_cert_file = "%s"`, filepath.Join(mountPath, core.TLSCertKey)))
		params = append(params, fmt.Sprintf(`tls_key_file = "%s"`, filepath.Join(mountPath, core.TLSPrivateKeyKey)))
		params = append(params, fmt.Sprintf(`tls_client_ca_file = "%s"`, filepath.Join(mountPath, conapi.TLSCACertKey)))
	} else {
		params = append(params, "tls_disable = true")
	}

	listenerCfg := fmt.Sprintf(listenerFmt, strings.Join(params, "\n"))

	return listenerCfg
}

func NewVaultClient(hostname string, port string, tlsConfig *vaultapi.TLSConfig) (*vaultapi.Client, error) {
	cfg := vaultapi.DefaultConfig()
	klog.Info("TLSConfig() => value of TLS Insecure: ", tlsConfig.Insecure)
	podURL := fmt.Sprintf("%s://%s:%s", Scheme(tlsConfig.Insecure), hostname, port)
	cfg.Address = podURL
	err := cfg.ConfigureTLS(tlsConfig)
	if err != nil {
		return nil, err
	}
	return vaultapi.NewClient(cfg)
}

func Scheme(tlsInsecure bool) string {
	if tlsInsecure {
		return "http"
	}
	return "https"
}
