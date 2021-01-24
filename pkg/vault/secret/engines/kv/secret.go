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

package kv

import (
	"fmt"

	"kubevault.dev/operator/pkg/vault/secret"

	vaultapi "github.com/hashicorp/vault/api"
	"github.com/pkg/errors"
)

const UID = "KV"

type SecretInfo struct {
	// Specifies the path where secret engine is enabled
	Path string

	// Specifies the name of secret
	SecretName string

	Client *vaultapi.Client
}

func New() *SecretInfo {
	return &SecretInfo{}
}

func NewSecretManager() secret.SecretManager {
	return &SecretInfo{}
}

func NewSecretManagerWithOptions(c *vaultapi.Client, opts map[string]string) secret.SecretManager {
	s := &SecretInfo{}
	s.Client = c
	if val, ok := opts[secret.SecretKey]; ok {
		s.SecretName = val
	}
	if val, ok := opts[secret.PathKey]; ok {
		s.Path = val
	}
	return s
}

func NewSecretGetter(c *vaultapi.Client, path string, secretName string) secret.SecretGetter {
	return &SecretInfo{
		Client:     c,
		Path:       path,
		SecretName: secretName,
	}
}

func (s *SecretInfo) SetOptions(c *vaultapi.Client, opts map[string]string) error {
	s.Client = c
	if val, ok := opts[secret.SecretKey]; ok {
		s.SecretName = val
	}
	if val, ok := opts[secret.PathKey]; ok {
		s.Path = val
	}
	return nil
}

func (s *SecretInfo) GetSecret() (*vaultapi.Secret, error) {
	if s.Path == "" {
		return nil, errors.New("kv secret engine path is empty")
	}
	if s.SecretName == "" {
		return nil, errors.New("secret name is empty")
	}
	if s.Client == nil {
		return nil, errors.New("vault api client is nil")
	}

	path := fmt.Sprintf("/v1/%s/%s", s.Path, s.SecretName)
	req := s.Client.NewRequest("GET", path)

	resp, err := s.Client.RawRequest(req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get secret")
	}

	defer resp.Body.Close()
	sr, err := vaultapi.ParseSecret(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse secret from response body")
	}
	return sr, nil
}
