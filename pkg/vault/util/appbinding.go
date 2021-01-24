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
	"crypto/x509"
	"net/http"

	vaultapi "github.com/hashicorp/vault/api"
	"github.com/pkg/errors"
	appcat "kmodules.xyz/custom-resources/apis/appcatalog/v1alpha1"
)

func VaultConfigFromAppBinding(app *appcat.AppBinding) (*vaultapi.Config, error) {
	clientCfg := app.Spec.ClientConfig
	cfg := vaultapi.DefaultConfig()
	var err error
	cfg.Address, err = app.URL()
	if err != nil {
		return nil, err
	}

	// set max retry to 4 ( 5 times )
	cfg.MaxRetries = 4

	clientTLSConfig := cfg.HttpClient.Transport.(*http.Transport).TLSClientConfig
	if clientCfg.InsecureSkipTLSVerify {
		clientTLSConfig.InsecureSkipVerify = true
	} else {
		if len(clientCfg.CABundle) != 0 {
			pool := x509.NewCertPool()
			ok := pool.AppendCertsFromPEM(clientCfg.CABundle)
			if !ok {
				return nil, errors.New("error loading CA File: couldn't parse PEM data in CA bundle")
			}
			clientTLSConfig.RootCAs = pool
		}
	}
	clientTLSConfig.ServerName, err = app.Hostname()
	if err != nil {
		return nil, err
	}

	return cfg, nil
}
