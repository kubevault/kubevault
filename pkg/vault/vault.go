package vault

import (
	"crypto/tls"
	"crypto/x509"
	"net/http"
	"net/url"

	vaultapi "github.com/hashicorp/vault/api"
	api "github.com/kubevault/operator/apis/policy/v1alpha1"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func NewClient(kc kubernetes.Interface, namespace string, v *api.Vault) (*vaultapi.Client, error) {
	if v == nil {
		return nil, errors.New(".spec.vault is nil")
	}

	cfg, err := newVaultConfig(kc, namespace, v)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create vault client config")
	}

	cl, err := vaultapi.NewClient(cfg)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	sr, err := kc.CoreV1().Secrets(namespace).Get(v.TokenSecret, metav1.GetOptions{})
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get vault token secret %s/%s", namespace, v.TokenSecret)
	}

	if sr.Data == nil {
		return nil, errors.Errorf("vault token is not found in secret %s/%s", namespace, v.TokenSecret)
	}
	if _, ok := sr.Data["token"]; !ok {
		return nil, errors.Errorf("vault token is not found in secret %s/%s", namespace, v.TokenSecret)
	}
	cl.SetToken(string(sr.Data["token"]))

	return cl, nil
}

func newVaultConfig(kc kubernetes.Interface, namespace string, v *api.Vault) (*vaultapi.Config, error) {

	cfg := vaultapi.DefaultConfig()
	cfg.Address = v.Address

	clientTLSConfig := cfg.HttpClient.Transport.(*http.Transport).TLSClientConfig

	if v.ClientTLSSecret != "" {
		sr, err := kc.CoreV1().Secrets(namespace).Get(v.ClientTLSSecret, metav1.GetOptions{})
		if err != nil {
			return nil, errors.Wrapf(err, "failed to get vault client tls secret %s/%s", namespace, v.ClientTLSSecret)
		}

		clientTLSConfig.GetClientCertificate = func(*tls.CertificateRequestInfo) (*tls.Certificate, error) {
			cert, err := tls.X509KeyPair(sr.Data["client.crt"], sr.Data["client.key"])
			if err != nil {
				return nil, errors.WithStack(err)
			}
			return &cert, nil
		}
	}

	if v.SkipTLSVerification {
		clientTLSConfig.InsecureSkipVerify = true
	} else {
		if v.ServerCASecret != "" {
			sr, err := kc.CoreV1().Secrets(namespace).Get(v.ClientTLSSecret, metav1.GetOptions{})
			if err != nil {
				return nil, errors.Wrapf(err, "failed to get vault server ca secret %s/%s", namespace, v.ServerCASecret)
			}

			pool := x509.NewCertPool()
			ok := pool.AppendCertsFromPEM(sr.Data["ca.crt"])
			if !ok {
				return nil, errors.Errorf("error loading CA File: couldn't parse PEM data in secret %s/%s", namespace, v.ServerCASecret)
			}

			clientTLSConfig.RootCAs = pool
		}
	}

	var err error

	clientTLSConfig.ServerName, err = getHostName(v.Address)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get hostname from url %s", v.Address)
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
