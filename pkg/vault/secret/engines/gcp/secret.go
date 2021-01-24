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
	"fmt"

	engine "kubevault.dev/apimachinery/apis/engine/v1alpha1"
	"kubevault.dev/operator/pkg/vault/role/gcp"
	"kubevault.dev/operator/pkg/vault/secret"

	vaultapi "github.com/hashicorp/vault/api"
	"github.com/pkg/errors"
)

const (
	UID = "GCP"
)

type SecretInfo struct {
	// Specifies the path where secret engine is enabled
	Path string

	// Specifies the role for credential
	Role string

	// Contains the information about secret type, i.e. access_token or service_account_key
	SecretType string

	// Key algorithm used to generate key, required when SecretType is set to service_account_key
	KeyAlgorithm string

	// Private key type to generate, required when SecretType is set to service_account_key
	KeyType string

	Client *vaultapi.Client
}

func New() *SecretInfo {
	return &SecretInfo{}
}

func NewSecretManager() secret.SecretManager {
	return &SecretInfo{}
}

func NewSecretManagerWithOptions(c *vaultapi.Client, opts map[string]string) (secret.SecretManager, error) {
	s := &SecretInfo{}
	err := s.SetOptions(c, opts)
	return s, err
}

func NewSecretGetter(vc *vaultapi.Client, path string, roleName string, secretType engine.GCPSecretType, reqSpec engine.GCPAccessKeyRequestSpec) secret.SecretGetter {
	return &SecretInfo{
		Path:         path,
		Role:         roleName,
		SecretType:   string(secretType),
		KeyAlgorithm: reqSpec.KeyAlgorithm,
		KeyType:      reqSpec.KeyType,
		Client:       vc,
	}
}

func (s *SecretInfo) SetOptions(c *vaultapi.Client, opts map[string]string) error {
	s.Client = c
	if val, ok := opts[secret.RoleKey]; ok {
		s.Role = val
	}
	if val, ok := opts[secret.PathKey]; ok {
		s.Path = val
	}
	if val, ok := opts[gcp.GCPSecretType]; ok {
		s.SecretType = val
	}
	if val, ok := opts[secret.KeyAlgorithm]; ok {
		s.KeyAlgorithm = val
	}
	if val, ok := opts[secret.KeyType]; ok {
		s.KeyType = val
	}
	return nil
}

func (s *SecretInfo) GetSecret() (*vaultapi.Secret, error) {
	if s.Path == "" {
		return nil, errors.New("gcp secret engine path is empty")
	}
	if s.Role == "" {
		return nil, errors.New("gcp role is empty")
	}
	if s.Client == nil {
		return nil, errors.New("vault api client is nil")
	}

	var path string
	if s.SecretType == string(engine.GCPSecretAccessToken) {
		path = fmt.Sprintf("/v1/%s/token/%s", s.Path, s.Role)
	} else if s.SecretType == string(engine.GCPSecretServiceAccountKey) {
		path = fmt.Sprintf("/v1/%s/key/%s", s.Path, s.Role)
	} else {
		return nil, errors.New("secret_type is not specified")
	}

	req := s.Client.NewRequest("GET", path)
	if s.SecretType == string(engine.GCPSecretServiceAccountKey) {
		payload := map[string]interface{}{}
		if len(s.KeyAlgorithm) != 0 {
			payload[secret.KeyAlgorithm] = s.KeyAlgorithm
		}
		if len(s.KeyType) != 0 {
			payload[secret.KeyType] = s.KeyType
		}
		if err := req.SetJSONBody(payload); err != nil {
			return nil, errors.Wrap(err, "failed to load payload in gcp access key request")
		}
	}
	resp, err := s.Client.RawRequest(req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get gcp access secret")
	}

	defer resp.Body.Close()
	sr, err := vaultapi.ParseSecret(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse secret from response body")
	}
	return sr, nil
}
