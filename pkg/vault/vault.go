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

package vault

import (
	"encoding/json"

	config "kubevault.dev/operator/apis/config/v1alpha1"
	vaultauth "kubevault.dev/operator/pkg/vault/auth"
	vaultutil "kubevault.dev/operator/pkg/vault/util"

	vaultapi "github.com/hashicorp/vault/api"
	"github.com/pkg/errors"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	appcat "kmodules.xyz/custom-resources/apis/appcatalog/v1alpha1"
	appcat_cs "kmodules.xyz/custom-resources/client/clientset/versioned/typed/appcatalog/v1alpha1"
)

func NewClient(kc kubernetes.Interface, appc appcat_cs.AppcatalogV1alpha1Interface, vAppRef *appcat.AppReference) (*vaultapi.Client, error) {

	vApp, err := appc.AppBindings(vAppRef.Namespace).Get(vAppRef.Name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	return NewClientWithAppBinding(kc, vApp)
}

func NewClientWithAppBinding(kc kubernetes.Interface, vApp *appcat.AppBinding) (*vaultapi.Client, error) {
	// If k8s service account name is provided as AppBinding parameters,
	// the operator will perform Kubernetes authentication to the Vault server.
	// Generate service account reference from AppBinding parameters
	var saRef *core.ObjectReference
	if vApp.Spec.Parameters != nil && vApp.Spec.Parameters.Raw != nil {
		var cf config.VaultServerConfiguration
		err := json.Unmarshal(vApp.Spec.Parameters.Raw, &cf)
		if err != nil {
			return nil, errors.Wrap(err, "failed to unmarshal parameters")
		}

		if cf.ServiceAccountName != "" {
			saRef = &core.ObjectReference{
				Namespace: vApp.Namespace,
				Name:      cf.ServiceAccountName,
			}
		}
	}

	return NewClientWithAppBindingAndSaRef(kc, vApp, saRef)
}

func NewClientWithAppBindingAndSaRef(kc kubernetes.Interface, vApp *appcat.AppBinding, saRef *core.ObjectReference) (*vaultapi.Client, error) {
	if vApp == nil {
		return nil, errors.New("AppBinding is nil")
	}

	auth, err := vaultauth.NewAuth(kc, vApp, saRef)
	if err != nil {
		return nil, err
	}

	token, err := auth.Login()
	if err != nil {
		return nil, errors.Wrap(err, "failed to login")
	}

	cfg, err := vaultutil.VaultConfigFromAppBinding(vApp)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create vault client config")
	}

	vc, err := vaultapi.NewClient(cfg)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	vc.SetToken(token)
	return vc, nil
}
