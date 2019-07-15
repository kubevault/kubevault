package controller

import (
	"encoding/json"

	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	appcat "kmodules.xyz/custom-resources/apis/appcatalog/v1alpha1"
	appcat_util "kmodules.xyz/custom-resources/client/clientset/versioned/typed/appcatalog/v1alpha1/util"
	vaultconfig "kubevault.dev/operator/apis/config/v1alpha1"
	api "kubevault.dev/operator/apis/kubevault/v1alpha1"
	"kubevault.dev/operator/pkg/vault/util"
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
		UsePodServiceAccountForCSIDriver: true,
		AuthPath:                         string(api.AuthTypeKubernetes),
		ServiceAccountName:               vs.ServiceAccountName(),
		PolicyControllerRole:             vs.PolicyNameForPolicyController(),
		AuthMethodControllerRole:         vaultPolicyBindingForAuthMethod(vs).PolicyBindingName(),
		TokenReviewerServiceAccountName:  vs.ServiceAccountForTokenReviewer(),
	}
	dataConf, err := json.Marshal(vsConf)
	if err != nil {
		return err
	}
	_, _, err = appcat_util.CreateOrPatchAppBinding(c.appCatalogClient, meta, func(in *appcat.AppBinding) *appcat.AppBinding {
		in.Labels = vs.OffshootLabels()
		in.Spec.ClientConfig = vClientConf
		in.Spec.Parameters = &runtime.RawExtension{
			Raw: dataConf,
		}
		util.EnsureOwnerRefToObject(in.GetObjectMeta(), util.AsOwner(vs))
		return in
	})
	return err
}
