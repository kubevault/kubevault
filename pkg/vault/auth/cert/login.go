package cert

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"

	vaultapi "github.com/hashicorp/vault/api"
	"github.com/kubevault/operator/apis"
	config "github.com/kubevault/operator/apis/config/v1alpha1"
	vsapi "github.com/kubevault/operator/apis/kubevault/v1alpha1"
	"github.com/kubevault/operator/pkg/vault/auth/types"
	vaultuitl "github.com/kubevault/operator/pkg/vault/util"
	"github.com/pkg/errors"
	core "k8s.io/api/core/v1"
	appcat "kmodules.xyz/custom-resources/apis/appcatalog/v1alpha1"
)

type auth struct {
	vClient *vaultapi.Client
	name    string
	path    string
}

// links : https://www.vaultproject.io/docs/auth/aws.html

func New(vApp *appcat.AppBinding, secret *core.Secret) (*auth, error) {
	cfg, err := vaultuitl.VaultConfigFromAppBinding(vApp)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create vault config from AppBinding")
	}

	clientTLSConfig := cfg.HttpClient.Transport.(*http.Transport).TLSClientConfig
	clientTLSConfig.InsecureSkipVerify = true
	clientTLSConfig.GetClientCertificate = func(*tls.CertificateRequestInfo) (*tls.Certificate, error) {
		cert, err := tls.X509KeyPair(secret.Data[core.TLSCertKey], secret.Data[core.TLSPrivateKeyKey])
		if err != nil {
			return nil, errors.WithStack(err)
		}
		return &cert, nil
	}

	vc, err := vaultapi.NewClient(cfg)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create vault client")
	}

	var cf config.VaultServerConfiguration
	if vApp.Spec.Parameters != nil {
		err = json.Unmarshal([]byte(vApp.Spec.Parameters.Raw), &cf)
		if err != nil {
			return nil, errors.Wrap(err, "failed to unmarshal parameters")
		}
	}

	authPath := string(vsapi.AuthTypeCert)
	if val, ok := secret.Annotations[apis.AuthPathKey]; ok && len(val) > 0 {
		authPath = val
	}

	return &auth{
		vClient: vc,
		name:    cf.PolicyControllerRole,
		path:    authPath,
	}, nil
}

// Login will log into vault and return client token
func (a *auth) Login() (string, error) {
	path := fmt.Sprintf("/v1/auth/%s/login", a.path)
	req := a.vClient.NewRequest("POST", path)
	payload := map[string]interface{}{
		"name": a.name,
	}

	if err := req.SetJSONBody(payload); err != nil {
		return "", err
	}

	resp, err := a.vClient.RawRequest(req)
	if err != nil {
		return "", err
	}

	var loginResp types.AuthLoginResponse
	defer resp.Body.Close()
	err = json.NewDecoder(resp.Body).Decode(&loginResp)
	if err != nil {
		return "", err
	}
	return loginResp.Auth.ClientToken, nil
}
