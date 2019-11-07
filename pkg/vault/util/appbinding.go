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
