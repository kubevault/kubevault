package vault

import (
	"crypto/x509"
	"fmt"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"

	vaultapi "github.com/hashicorp/vault/api"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	appcat "kmodules.xyz/custom-resources/apis/appcatalog/v1alpha1"
	appcat_cs "kmodules.xyz/custom-resources/client/clientset/versioned/typed/appcatalog/v1alpha1"
)

func NewClient(kc kubernetes.Interface, appc appcat_cs.AppcatalogV1alpha1Interface, vAppRef *appcat.AppReference) (*vaultapi.Client, error) {
	if vAppRef == nil {
		return nil, errors.New(".spec.vaultAppRef is nil")
	}

	vApp, err := appc.AppBindings(vAppRef.Namespace).Get(vAppRef.Name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	cfg, err := newVaultConfig(kc, vApp)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create vault client config")
	}

	cl, err := vaultapi.NewClient(cfg)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	if vApp.Spec.Secret == nil {
		return nil, errors.New("secret for vault token is not provided")
	}

	tokenSecret := vApp.Spec.Secret.Name
	sr, err := kc.CoreV1().Secrets(vApp.Namespace).Get(tokenSecret, metav1.GetOptions{})
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get vault token secret %s/%s", vApp.Namespace, tokenSecret)
	}

	if sr.Data == nil {
		return nil, errors.Errorf("vault token is not found in secret %s/%s", vApp.Namespace, tokenSecret)
	}
	if _, ok := sr.Data["token"]; !ok {
		return nil, errors.Errorf("vault token is not found in secret %s/%s", vApp.Namespace, tokenSecret)
	}
	cl.SetToken(string(sr.Data["token"]))

	return cl, nil
}

func newVaultConfig(kc kubernetes.Interface, vApp *appcat.AppBinding) (*vaultapi.Config, error) {
	clientCfg := vApp.Spec.ClientConfig
	cfg := vaultapi.DefaultConfig()
	if clientCfg.URL == nil {
		if clientCfg.Service != nil {
			host := fmt.Sprintf("%s.%s.svc", clientCfg.Service.Name, vApp.Namespace)
			var port int32
			for _, p := range clientCfg.Ports {
				if strings.ToLower(p.Name) == "client" {
					port = p.Port
				}
			}
			if port == 0 {
				return nil, errors.New("client port for vault doesn't provided")
			}
			addr := fmt.Sprintf("%s:%d", host, port)
			if clientCfg.Service.Path != nil {
				addr = filepath.Join(addr, *clientCfg.Service.Path)
			}
			cfg.Address = addr
		} else {
			return nil, errors.New("vault address is not found")
		}
	} else {
		cfg.Address = *clientCfg.URL
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

	var err error
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
