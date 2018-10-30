package serviceaccount

import (
	"encoding/json"
	"fmt"
	"time"

	vaultapi "github.com/hashicorp/vault/api"
	config "github.com/kubevault/operator/apis/config/v1alpha1"
	sa_util "github.com/kubevault/operator/pkg/util"
	"github.com/kubevault/operator/pkg/vault/auth/types"
	vaultuitl "github.com/kubevault/operator/pkg/vault/util"
	"github.com/pkg/errors"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	appcat "kmodules.xyz/custom-resources/apis/appcatalog/v1alpha1"
)

const KubernetesDefaultAuthPath = "kubernetes"

type auth struct {
	vClient *vaultapi.Client
	jwt     string
	role    string
	path    string
}

func New(kc kubernetes.Interface, vApp *appcat.AppBinding) (*auth, error) {
	if vApp.Spec.Parameters == nil {
		return nil, errors.New("parameters are not provided")
	}

	cfg, err := vaultuitl.VaultConfigFromAppBinding(vApp)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create vault config from AppBinding")
	}

	vc, err := vaultapi.NewClient(cfg)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create vault client")
	}

	var cf config.VaultServerConfiguration
	raw, err := vaultuitl.UnQuoteJson(string(vApp.Spec.Parameters.Raw))
	if err != nil {
		return nil, errors.Wrap(err, "failed to unquote json")
	}
	err = json.Unmarshal([]byte(raw), &cf)
	if err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal parameters")
	}

	if cf.ServiceAccountName == "" {
		return nil, errors.Wrap(err, "service account is not found")
	}

	secretName, err := sa_util.TryGetJwtTokenSecretNameFromServiceAccount(kc, cf.ServiceAccountName, vApp.Namespace, 2*time.Second, 30*time.Second)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get jwt token secret name of service account %s/%s", vApp.Namespace, cf.ServiceAccountName)
	}

	secret, err := kc.CoreV1().Secrets(vApp.Namespace).Get(secretName, metav1.GetOptions{})
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get secret %s/%s", vApp.Namespace, secretName)

	}

	jwt, ok := secret.Data[core.ServiceAccountTokenKey]
	if !ok {
		return nil, errors.New("jwt is missing")
	}

	if cf.AuthPath == "" {
		cf.AuthPath = KubernetesDefaultAuthPath
	}
	if cf.PolicyControllerRole == "" {
		return nil, errors.Wrap(err, "policyControllerRole is empty")
	}

	return &auth{
		vClient: vc,
		jwt:     string(jwt),
		role:    cf.PolicyControllerRole,
		path:    cf.AuthPath,
	}, nil
}

// Login will log into vault and return client token
func (a *auth) Login() (string, error) {
	path := fmt.Sprintf("/v1/auth/%s/login", a.path)
	req := a.vClient.NewRequest("POST", path)
	payload := map[string]interface{}{
		"jwt":  a.jwt,
		"role": a.role,
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
