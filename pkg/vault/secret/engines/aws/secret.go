package aws

import (
	"fmt"
	"strconv"

	vaultapi "github.com/hashicorp/vault/api"
	"github.com/pkg/errors"
	"kubevault.dev/operator/pkg/vault/secret"
)

const (
	UID       = "AWS"
	UseSTSKey = "useSTS"
)

type SecretInfo struct {
	// Specifies the path where secret engine is enabled
	Path string

	// Specifies the role for credential
	Role string

	// If true, '/aws/sts' endpoint will be used to retrieve credential
	// Otherwise, '/aws/creds' endpoint will be used to retrieve credential
	UseSTS bool

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

func NewSecretGetter(vc *vaultapi.Client, path string, role string, useSTS bool) secret.SecretGetter {
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
	if val, ok := opts[UseSTSKey]; ok {
		if v, err := strconv.ParseBool(val); err == nil {
			s.UseSTS = v
		}
	}
	return nil
}

func (s *SecretInfo) GetSecret() (*vaultapi.Secret, error) {
	if s.Path == "" {
		return nil, errors.New("aws secret engine path is empty")
	}
	if s.Role == "" {
		return nil, errors.New("aws role is empty")
	}
	if s.Client == nil {
		return nil, errors.New("vault api client is nil")
	}

	var path string
	if s.UseSTS {
		path = fmt.Sprintf("/v1/%s/sts/%s", s.Path, s.Role)
	} else {
		path = fmt.Sprintf("/v1/%s/creds/%s", s.Path, s.Role)
	}
	req := s.Client.NewRequest("GET", path)

	resp, err := s.Client.RawRequest(req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get aws secret")
	}

	defer resp.Body.Close()
	sr, err := vaultapi.ParseSecret(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse secret from response body")
	}
	return sr, nil
}
