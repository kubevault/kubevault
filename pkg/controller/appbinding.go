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

package controller

import (
	"context"
	"encoding/json"

	vaultconfig "kubevault.dev/apimachinery/apis/config/v1alpha1"
	api "kubevault.dev/apimachinery/apis/kubevault/v1alpha1"

	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	core_util "kmodules.xyz/client-go/core/v1"
	appcat "kmodules.xyz/custom-resources/apis/appcatalog/v1alpha1"
	appcat_util "kmodules.xyz/custom-resources/client/clientset/versioned/typed/appcatalog/v1alpha1/util"
)

func (c *VaultController) ensureAppBindings(vs *api.VaultServer, v Vault) error {
	meta := metav1.ObjectMeta{
		Name:      vs.AppBindingName(),
		Namespace: vs.Namespace,
	}
	_, caBundle, err := v.GetServerTLS()
	if err != nil {
		return err
	}

	vClientConf := appcat.ClientConfig{
		Service: &appcat.ServiceReference{
			Name:   v.GetService().Name,
			Scheme: string(core.URISchemeHTTPS),
			Port:   VaultClientPort,
		},
		CABundle: caBundle,
	}

	vsConf := vaultconfig.VaultServerConfiguration{
		TypeMeta: metav1.TypeMeta{
			APIVersion: vaultconfig.SchemeGroupVersion.String(),
			Kind:       vaultconfig.ResourceKindVaultServerConfiguration,
		},
		Kubernetes: &vaultconfig.KubernetesAuthConfig{
			ServiceAccountName:               vs.ServiceAccountName(),
			TokenReviewerServiceAccountName:  vs.ServiceAccountForTokenReviewer(),
			UsePodServiceAccountForCSIDriver: true,
		},
		Path:      string(api.AuthTypeKubernetes),
		VaultRole: vs.PolicyNameForPolicyController(),
	}
	dataConf, err := json.Marshal(vsConf)
	if err != nil {
		return err
	}
	_, _, err = appcat_util.CreateOrPatchAppBinding(context.TODO(), c.appCatalogClient, meta, func(in *appcat.AppBinding) *appcat.AppBinding {
		in.Labels = vs.OffshootLabels()
		in.Spec.ClientConfig = vClientConf
		in.Spec.Parameters = &runtime.RawExtension{
			Raw: dataConf,
		}
		core_util.EnsureOwnerReference(in, metav1.NewControllerRef(vs, api.SchemeGroupVersion.WithKind(api.ResourceKindVaultServer)))
		return in
	}, metav1.PatchOptions{})
	return err
}
