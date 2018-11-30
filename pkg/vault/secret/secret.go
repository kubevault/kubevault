package secret

import (
	vaultapi "github.com/hashicorp/vault/api"
)

const (
	RoleKey   = "role"
	PathKey   = "path"
	SecretKey = "secret"
)

type SecretManager interface {
	SecretGetter
	SetOptions(*vaultapi.Client, map[string]string) error
}

type SecretGetter interface {
	GetSecret() ([]byte, error)
}
