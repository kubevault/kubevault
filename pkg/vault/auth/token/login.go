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

package token

import (
	"kubevault.dev/apimachinery/apis"
	authtype "kubevault.dev/operator/pkg/vault/auth/types"

	"github.com/pkg/errors"
)

type auth struct {
	token string
}

func New(authInfo *authtype.AuthInfo) (*auth, error) {
	if authInfo == nil {
		return nil, errors.New("authentication information is empty")
	}
	if authInfo.Secret == nil {
		return nil, errors.New("authentication secret is missing")
	}

	token, ok := authInfo.Secret.Data[apis.TokenAuthTokenKey]
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
