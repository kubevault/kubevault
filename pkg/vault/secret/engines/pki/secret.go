package pki

import (
	"fmt"
	"io/ioutil"

	vaultapi "github.com/hashicorp/vault/api"
	"github.com/kubevault/operator/pkg/vault/secret"
	"github.com/pkg/errors"
)

const UID = "PKI"

type SecretInfo struct {
	// Specifies the path where secret engine is enabled
	Path string

	// Specifies the role
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

func NewSecretGetter(vc *vaultapi.Client, path string, role string) secret.SecretGetter {
	return &SecretInfo{
		Client: vc,
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

func (s *SecretInfo) GetSecret() ([]byte, error) {
	if s.Path == "" {
		return nil, errors.New("pki secret engine path is empty")
	}
	if s.Role == "" {
		return nil, errors.New("role is empty")
	}
	if s.Client == nil {
		return nil, errors.New("vault api client is nil")
	}

	path := fmt.Sprintf("/v1/%s/issue/%s", s.Path, s.Role)
	req := s.Client.NewRequest("GET", path)

	resp, err := s.Client.RawRequest(req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get secret")
	}

	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read response body")
	}
	return data, nil
}
