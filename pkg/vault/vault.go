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
	vaultauth "kubevault.dev/operator/pkg/vault/auth"
	vaultutil "kubevault.dev/operator/pkg/vault/util"

	vaultapi "github.com/hashicorp/vault/api"
	"github.com/pkg/errors"
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
	if vApp == nil {
		return nil, errors.New("AppBinding is nil")
	}

	auth, err := vaultauth.NewAuth(kc, vApp)
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
