package util

import (
	"crypto/x509"
	"fmt"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"

	vaultapi "github.com/hashicorp/vault/api"
	"github.com/pkg/errors"
	appcat "kmodules.xyz/custom-resources/apis/appcatalog/v1alpha1"
)

func VaultConfigFromAppBinding(vApp *appcat.AppBinding) (*vaultapi.Config, error) {
	var err error
	clientCfg := vApp.Spec.ClientConfig
	cfg := vaultapi.DefaultConfig()
	cfg.Address, err = getAddress(vApp)
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

	clientTLSConfig.ServerName, err = getHostName(cfg.Address)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get hostname from url %s", cfg.Address)
	}

	return cfg, nil
}

func getHostName(addr string) (string, error) {
	u, err := url.Parse(addr)
	if err != nil {
		return "", errors.WithStack(err)
	}
	return u.Hostname(), nil
}

func getAddress(app *appcat.AppBinding) (string, error) {
	cfg := app.Spec.ClientConfig
	if cfg.URL == nil {
		if cfg.Service != nil {
			svc := cfg.Service
			host := fmt.Sprintf("%s.%s.svc", svc.Name, app.Namespace)
			if svc.Port == 0 {
				return "", errors.New("client port for vault doesn't provided")
			}
			if svc.Scheme == "" {
				return "", errors.New("url scheme is not specified")
			}

			addr := fmt.Sprintf("%s://%s:%d", strings.ToLower(svc.Scheme), host, svc.Port)
			if cfg.Service.Path != nil {
				addr = filepath.Join(addr, *cfg.Service.Path)
			}
			return addr, nil
		} else {
			return "", errors.New("vault address is not found")
		}
	} else {
		return *cfg.URL, nil
	}
}
