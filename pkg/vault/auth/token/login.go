package token

import (
	"github.com/kubevault/operator/apis"
	"github.com/pkg/errors"
	core "k8s.io/api/core/v1"
)

type auth struct {
	token string
}

func New(secret *core.Secret) (*auth, error) {
	token, ok := secret.Data[apis.TokenAuthTokenKey]
	if !ok {
		return nil, errors.New("token is missing")
	}
	return &auth{
		token: string(token),
	}, nil
}

// Login will log into vault and return client token
func (a *auth) Login() (string, error) {
	return a.token, nil
}
