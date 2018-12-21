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
