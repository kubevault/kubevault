/*
Copyright The KubeVault Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package database

import (
	"fmt"

	"kubevault.dev/operator/pkg/vault/secret"

	vaultapi "github.com/hashicorp/vault/api"
	"github.com/pkg/errors"
)

const UID = "DATABASE"

type SecretInfo struct {
	// Specifies the path where secret engine is enabled
	Path string

	// Specifies the role for credential
	Role string

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
	if val, ok := opts[secret.RoleKey]; ok {
		s.Role = val
	}
	if val, ok := opts[secret.PathKey]; ok {
		s.Path = val
	}
	return s
}

func NewSecretGetter(c *vaultapi.Client, path string, role string) secret.SecretGetter {
	return &SecretInfo{
		Client: c,
		Path:   path,
		Role:   role,
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
	return nil
}

func (s *SecretInfo) GetSecret() (*vaultapi.Secret, error) {
	if s.Path == "" {
		return nil, errors.New("database secret engine path is empty")
	}
	if s.Role == "" {
		return nil, errors.New("database role is empty")
	}
	if s.Client == nil {
		return nil, errors.New("vault api client is nil")
	}

	path := fmt.Sprintf("/v1/%s/creds/%s", s.Path, s.Role)
	req := s.Client.NewRequest("GET", path)

	resp, err := s.Client.RawRequest(req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get database credential")
	}

	defer resp.Body.Close()
	sr, err := vaultapi.ParseSecret(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse secret from response body")
	}
	return sr, nil
}
