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
	basicauth "kubevault.dev/operator/pkg/vault/auth/userpass"

	"github.com/pkg/errors"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	appcat "kmodules.xyz/custom-resources/apis/appcatalog/v1alpha1"
)

type AuthInterface interface {
	// successful login will return token
	// unsuccessful login will return err
	Login() (string, error)
}

func NewAuth(kc kubernetes.Interface, vApp *appcat.AppBinding, saRef *core.ObjectReference) (AuthInterface, error) {
	if vApp == nil {
		return nil, errors.New("vault AppBinding is not provided")
	}

	// if ServiceAccountReference exists, use Kubernetes service account authentication
	// otherwise use secret
	if saRef != nil {
		return saauth.New(kc, vApp, saRef)
	}

	if vApp.Spec.Secret == nil {
		return nil, errors.New("secret is not provided")
	}

	secret, err := kc.CoreV1().Secrets(vApp.Namespace).Get(vApp.Spec.Secret.Name, metav1.GetOptions{})
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get secret %s/%s", vApp.Namespace, vApp.Spec.Secret.Name)

	}

	switch secret.Type {
	case core.SecretTypeBasicAuth:
		return basicauth.New(vApp, secret)
	case core.SecretTypeTLS:
		return certauth.New(vApp, secret)
	case core.SecretTypeServiceAccountToken:
		return k8sauth.New(vApp, secret)
	case apis.SecretTypeTokenAuth:
		return tokenauth.New(secret)
	case apis.SecretTypeAWSAuth:
		return awsauth.New(vApp, secret)
	case apis.SecretTypeGCPAuth:
		return gcpauth.New(vApp, secret)
	case apis.SecretTypeAzureAuth:
		return azureauth.New(vApp, secret)
	default:
		return nil, errors.New("Invalid/Unsupported secret type")
	}
}
