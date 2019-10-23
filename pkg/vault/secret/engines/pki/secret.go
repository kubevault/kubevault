package pki

import (
	"encoding/json"
	"fmt"

	"kubevault.dev/operator/pkg/vault/secret"

	vaultapi "github.com/hashicorp/vault/api"
	"github.com/pkg/errors"
)

const UID = "PKI"

type CertOptions struct {
	CommonName        string `json:"common_name,omitempty"`
	AltName           string `json:"alt_name,omitempty"`
	IpSans            string `json:"ip_sans,omitempty"`
	UriSans           string `json:"uri_sans,omitempty"`
	OtherSans         string `json:"other_sans,omitempty"`
	Ttl               string `json:"ttl,omitempty"`
	Format            string `json:"format,omitempty"`
	PrivateKeyFormat  string `json:"private_key_format,omitempty"`
	ExcludeCnFromSans bool   `json:"exclude_cn_from_sans,omitempty"`
}

type SecretInfo struct {
	// Specifies the path where secret engine is enabled
	Path string

	// Specifies the role
	Role string

	CertOpts *CertOptions

	Client *vaultapi.Client
}

func New() *SecretInfo {
	return &SecretInfo{}
}

func NewSecretManager() secret.SecretManager {
	return &SecretInfo{
		CertOpts: &CertOptions{},
	}
}

func NewSecretManagerWithOptions(c *vaultapi.Client, opts map[string]string) (secret.SecretManager, error) {
	s := &SecretInfo{}
	s.Client = c
	if val, ok := opts[secret.RoleKey]; ok {
		s.Role = val
	}
	if val, ok := opts[secret.PathKey]; ok {
		s.Path = val
	}

	if s.CertOpts == nil {
		s.CertOpts = &CertOptions{}
	}

	data, err := json.Marshal(opts)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal cert options")
	}

	err = json.Unmarshal(data, s.CertOpts)
	if err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal cert options")
	}
	return s, nil
}

func NewSecretGetter(vc *vaultapi.Client, path string, role string, certOpts *CertOptions) secret.SecretGetter {
	return &SecretInfo{
		Client:   vc,
		Path:     path,
		Role:     role,
		CertOpts: certOpts,
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
	if s.CertOpts == nil {
		s.CertOpts = &CertOptions{}
	}

	data, err := json.Marshal(opts)
	if err != nil {
		return errors.Wrap(err, "failed to marshal cert options")
	}

	err = json.Unmarshal(data, s.CertOpts)
	if err != nil {
		return errors.Wrap(err, "failed to unmarshal cert options")
	}
	return nil
}

func (s *SecretInfo) GetSecret() (*vaultapi.Secret, error) {
	if s.Path == "" {
		return nil, errors.New("pki secret engine path is empty")
	}
	if s.Role == "" {
		return nil, errors.New("role is empty")
	}
	if s.Client == nil {
		return nil, errors.New("vault api client is nil")
	}
	if s.CertOpts == nil {
		s.CertOpts = &CertOptions{}
	}

	path := fmt.Sprintf("/v1/%s/issue/%s", s.Path, s.Role)
	req := s.Client.NewRequest("POST", path)
	if err := req.SetJSONBody(s.CertOpts); err != nil {
		return nil, errors.Wrap(err, "filed to set json body in request")
	}

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
