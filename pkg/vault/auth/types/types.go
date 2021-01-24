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

package types

import (
	"context"
	"encoding/json"

	config "kubevault.dev/apimachinery/apis/config/v1alpha1"

	"github.com/pkg/errors"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	appcat "kmodules.xyz/custom-resources/apis/appcatalog/v1alpha1"
)

type AuthLoginResponse struct {
	Auth *Auth `json:"auth"`
}

type Auth struct {
	ClientToken string `json:"client_token"`
}

type AuthInfo struct {
	VaultApp          *appcat.AppBinding
	ServiceAccountRef *core.ObjectReference
	Secret            *core.Secret
	ExtraInfo         *AuthExtraInfo
	VaultRole         string
	Path              string
}

type AuthExtraInfo struct {
	Kubernetes *config.KubernetesAuthConfig
	Azure      *config.AzureAuthConfig
	AWS        *config.AWSAuthConfig
}

func GetAuthInfoFromAppBinding(kc kubernetes.Interface, vApp *appcat.AppBinding) (*AuthInfo, error) {
	if kc == nil {
		return nil, errors.New("Kubernetes client is empty")
	}
	if vApp == nil {
		return nil, errors.New("AppBinding is empty")
	}

	// If k8s service account name is provided as AppBinding parameters,
	// the operator will perform Kubernetes authentication to the Vault server.
	// Generate service account reference from AppBinding parameters
	var sa *core.ObjectReference
	cf := new(config.VaultServerConfiguration)
	var secret *core.Secret
	if vApp.Spec.Parameters != nil && vApp.Spec.Parameters.Raw != nil {
		err := json.Unmarshal(vApp.Spec.Parameters.Raw, &cf)
		if err != nil {
			return nil, errors.Wrap(err, "failed to unmarshal parameters")
		}

		if cf.Kubernetes != nil && cf.Kubernetes.ServiceAccountName != "" {
			sa = &core.ObjectReference{
				Namespace: vApp.Namespace,
				Name:      cf.Kubernetes.ServiceAccountName,
			}
		}
	}

	if vApp.Spec.Secret != nil {
		var err error
		secret, err = kc.CoreV1().Secrets(vApp.Namespace).Get(context.TODO(), vApp.Spec.Secret.Name, metav1.GetOptions{})
		if err != nil {
			return nil, errors.Wrapf(err, "failed to get secret %s/%s", vApp.Namespace, vApp.Spec.Secret.Name)

		}
	}

	return &AuthInfo{
		VaultApp:          vApp,
		ServiceAccountRef: sa,
		VaultRole:         cf.VaultRole,
		Secret:            secret,
		Path:              cf.Path,
		ExtraInfo: &AuthExtraInfo{
			Kubernetes: cf.Kubernetes,
			Azure:      cf.Azure,
			AWS:        cf.AWS,
		},
	}, nil
}
