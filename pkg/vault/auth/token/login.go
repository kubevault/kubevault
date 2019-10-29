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
package token

import (
	"kubevault.dev/operator/apis"

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
