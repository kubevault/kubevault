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

package auth

import (
	"kubevault.dev/operator/apis"
	awsauth "kubevault.dev/operator/pkg/vault/auth/aws"
	azureauth "kubevault.dev/operator/pkg/vault/auth/azure"
	certauth "kubevault.dev/operator/pkg/vault/auth/cert"
	gcpauth "kubevault.dev/operator/pkg/vault/auth/gcp"
	k8sauth "kubevault.dev/operator/pkg/vault/auth/kubernetes"
	saauth "kubevault.dev/operator/pkg/vault/auth/serviceaccount"
	tokenauth "kubevault.dev/operator/pkg/vault/auth/token"
	authtype "kubevault.dev/operator/pkg/vault/auth/types"
	basicauth "kubevault.dev/operator/pkg/vault/auth/userpass"

	"github.com/pkg/errors"
	core "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
)

type AuthInterface interface {
	// successful login will return token
	// unsuccessful login will return err
	Login() (string, error)
}

func NewAuth(kc kubernetes.Interface, authInfo *authtype.AuthInfo) (AuthInterface, error) {
	if authInfo == nil {
		return nil, errors.New("authentication information is empty")
	}

	// if ServiceAccountReference exists, use Kubernetes service account authentication
	// otherwise use secret
	if authInfo.ServiceAccountRef != nil {
		return saauth.New(kc, authInfo)
	}

	if authInfo.Secret == nil {
		return nil, errors.New("secret is not provided")
	}

	switch authInfo.Secret.Type {
	case core.SecretTypeBasicAuth:
		return basicauth.New(authInfo)
	case core.SecretTypeTLS:
		return certauth.New(authInfo)
	case core.SecretTypeServiceAccountToken:
		return k8sauth.New(authInfo)
	case apis.SecretTypeTokenAuth:
		return tokenauth.New(authInfo)
	case apis.SecretTypeAWSAuth:
		return awsauth.New(authInfo)
	case apis.SecretTypeGCPAuth:
		return gcpauth.New(authInfo)
	case apis.SecretTypeAzureAuth:
		return azureauth.New(authInfo)
	default:
		return nil, errors.New("Invalid/Unsupported secret type")
	}
}
