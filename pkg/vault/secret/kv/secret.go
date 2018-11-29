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

	// Specifies the name of secret
	name string

	vc *vaultapi.Client
}

func NewSecretGetter(vc *vaultapi.Client, path string, secretName string) secret.SecretGetter {
	return &secretManager{
		vc:   vc,
		path: path,
		name: secretName,
	}
}

func (s *secretManager) GetSecret() ([]byte, error) {
	if s.path == "" {
		return nil, errors.New("kv secret engine path is empty")
	}
	if s.name == "" {
		return nil, errors.New("secret name is empty")
	}
	if s.vc == nil {
		return nil, errors.New("vault api client is nil")
	}

	path := fmt.Sprintf("/v1/%s/%s", s.path, s.name)
	req := s.vc.NewRequest("GET", path)

	resp, err := s.vc.RawRequest(req)
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
