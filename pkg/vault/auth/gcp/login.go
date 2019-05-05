package gcp

import (
	"encoding/json"
	"fmt"
	"time"

	vaultapi "github.com/hashicorp/vault/api"
	"github.com/kubevault/operator/apis"
	config "github.com/kubevault/operator/apis/config/v1alpha1"
	vsapi "github.com/kubevault/operator/apis/kubevault/v1alpha1"
	"github.com/kubevault/operator/pkg/vault/auth/types"
	vaultuitl "github.com/kubevault/operator/pkg/vault/util"
	"github.com/pkg/errors"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"golang.org/x/oauth2/jwt"
	"google.golang.org/api/iam/v1"
	core "k8s.io/api/core/v1"
	appcat "kmodules.xyz/custom-resources/apis/appcatalog/v1alpha1"
)

type auth struct {
	vClient   *vaultapi.Client
	signedJwt string
	role      string
	path      string
}

// credentialsFile is the unmarshalled representation of a credentials file.
type credentialsFile struct {
	Type string `json:"type"` // serviceAccountKey or userCredentialsKey

	// Service Account fields
	ClientEmail  string `json:"client_email"`
	PrivateKeyID string `json:"private_key_id"`
	PrivateKey   string `json:"private_key"`
	TokenURL     string `json:"token_uri"`
	ProjectID    string `json:"project_id"`

	// User Credential fields
	// (These typically come from gcloud auth.)
	ClientSecret string `json:"client_secret"`
	ClientID     string `json:"client_id"`
	RefreshToken string `json:"refresh_token"`
}

// https://www.vaultproject.io/api/auth/gcp/index.html
func New(vApp *appcat.AppBinding, secret *core.Secret) (*auth, error) {

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

	saJson, ok := secret.Data[apis.GCPAuthSACredentialJson]
	if !ok {
		return nil, errors.New("sa.json is missing")
	}

	var cf config.VaultServerConfiguration
	err = json.Unmarshal([]byte(vApp.Spec.Parameters.Raw), &cf)
	if err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal parameters")
	}

	if cf.PolicyControllerRole == "" {
		return nil, errors.Wrap(err, "policyControllerRole is empty")
	}

	var cred credentialsFile
	if err := json.Unmarshal(saJson, &cred); err != nil {
		return nil, errors.Wrap(err, "credential Unmarshal failed!")
	}

	resp, err := getJWT(cred, cf.PolicyControllerRole)
	if err != nil {
		return nil, errors.Wrap(err, "JWT generation failed!")
	}

	authPath := string(vsapi.AuthTypeGcp)
	if val, ok := secret.Annotations[apis.AuthPathKey]; ok && len(val) > 0 {
		authPath = val
	}

	return &auth{
		vClient:   vc,
		signedJwt: resp.SignedJwt,
		role:      cf.PolicyControllerRole,
		path:      authPath,
	}, nil
}

func getJWT(cred credentialsFile, role string) (*iam.SignJwtResponse, error) {

	config := &jwt.Config{
		Email:        cred.ClientEmail,
		PrivateKey:   []byte(cred.PrivateKey),
		PrivateKeyID: cred.PrivateKeyID,
		Scopes:       []string{iam.CloudPlatformScope},
		TokenURL:     cred.TokenURL,
	}
	if config.TokenURL == "" {
		config.TokenURL = google.JWTTokenURL
	}
	httpClient := config.Client(oauth2.NoContext)
	iamClient, err := iam.New(httpClient)
	if err != nil {
		return nil, err
	}

	// 1. Generate signed JWT using IAM.
	// https://opensource.googleblog.com/2017/08/hashicorp-vault-and-google-cloud-iam.html
	resourceName := fmt.Sprintf("projects/%s/serviceAccounts/%s", cred.ProjectID, config.Email)
	jwtPayload := map[string]interface{}{
		"aud": "vault/" + role,
		"sub": config.Email,
		"exp": time.Now().Add(time.Minute * 15).Unix(),
	}

	payloadBytes, err := json.Marshal(jwtPayload)
	if err != nil {
		return nil, err
	}
	signJwtReq := &iam.SignJwtRequest{
		Payload: string(payloadBytes),
	}

	resp, err := iamClient.Projects.ServiceAccounts.SignJwt(
		resourceName, signJwtReq).Do()
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// Login will log into vault and return client token
func (a *auth) Login() (string, error) {
	path := fmt.Sprintf("/v1/auth/%s/login", a.path)
	req := a.vClient.NewRequest("POST", path)

	payload := make(map[string]interface{})
	payload["role"] = a.role
	payload["jwt"] = a.signedJwt
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
