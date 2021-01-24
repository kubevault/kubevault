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

package gcp

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"kubevault.dev/apimachinery/apis"
	vsapi "kubevault.dev/apimachinery/apis/kubevault/v1alpha1"
	"kubevault.dev/operator/pkg/vault/auth/types"
	authtype "kubevault.dev/operator/pkg/vault/auth/types"
	vaultuitl "kubevault.dev/operator/pkg/vault/util"

	vaultapi "github.com/hashicorp/vault/api"
	"github.com/pkg/errors"
	"golang.org/x/oauth2/google"
	"golang.org/x/oauth2/jwt"
	"google.golang.org/api/iam/v1"
	"google.golang.org/api/option"
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
func New(authInfo *authtype.AuthInfo) (*auth, error) {
	if authInfo == nil {
		return nil, errors.New("authentication information is empty")
	}
	if authInfo.VaultApp == nil {
		return nil, errors.New("AppBinding is empty")
	}

	vApp := authInfo.VaultApp
	cfg, err := vaultuitl.VaultConfigFromAppBinding(vApp)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create vault config from AppBinding")
	}

	vc, err := vaultapi.NewClient(cfg)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create vault client")
	}

	if authInfo.Secret == nil {
		return nil, errors.New("authentication secret is missing")
	}

	secret := authInfo.Secret
	saJson, ok := secret.Data[apis.GCPAuthSACredentialJson]
	if !ok {
		return nil, errors.New("google service account credential (i.e. sa.json) is missing")
	}

	authPath := string(vsapi.AuthTypeGcp)
	if authInfo.Path != "" {
		authPath = authInfo.Path
	}

	if authInfo.VaultRole == "" {
		return nil, errors.New("Vault role is empty")
	}

	var cred credentialsFile
	if err := json.Unmarshal(saJson, &cred); err != nil {
		return nil, errors.Wrap(err, "credential Unmarshal failed!")
	}

	resp, err := getJWT(cred, authInfo.VaultRole)
	if err != nil {
		return nil, errors.Wrap(err, "JWT generation failed!")
	}

	return &auth{
		vClient:   vc,
		signedJwt: resp.SignedJwt,
		role:      authInfo.VaultRole,
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
	ctx := context.Background()
	iamClient, err := iam.NewService(ctx, option.WithHTTPClient(config.Client(ctx)))
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
