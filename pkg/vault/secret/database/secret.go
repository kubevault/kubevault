package database

import (
	"fmt"
	"io/ioutil"

	vaultapi "github.com/hashicorp/vault/api"
	"github.com/kubevault/operator/pkg/vault/secret"
	"github.com/pkg/errors"
)

type secretManager struct {
	// Specifies the path where secret engine is enabled
	path string

	// Specifies the role for credential
	role string

	vc *vaultapi.Client
}

func NewSecretGetter(vc *vaultapi.Client, path string, role string) secret.SecretGetter {
	return &secretManager{
		vc:   vc,
		path: path,
		role: role,
	}
}

func (s *secretManager) GetSecret() ([]byte, error) {
	if s.path == "" {
		return nil, errors.New("database secret engine path is empty")
	}
	if s.role == "" {
		return nil, errors.New("database role is empty")
	}
	if s.vc == nil {
		return nil, errors.New("vault api client is nil")
	}

	path := fmt.Sprintf("/v1/%s/creds/%s", s.path, s.role)
	req := s.vc.NewRequest("GET", path)

	resp, err := s.vc.RawRequest(req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get database credential")
	}

	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read response body")
	}
	return data, nil
}
