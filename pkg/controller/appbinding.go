package controller

import (
	"encoding/json"
	"fmt"

	api "github.com/kubevault/operator/apis/kubevault/v1alpha1"
	"github.com/kubevault/operator/pkg/vault/util"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	appcat "kmodules.xyz/custom-resources/apis/appcatalog/v1alpha1"
	appcat_util "kmodules.xyz/custom-resources/client/clientset/versioned/typed/appcatalog/v1alpha1/util"
	vaultconfig "github.com/kubevault/operator/apis/config/v1alpha1"
)

const vaultConfigFmt = `
{
  "apiVersion": "kubevault.com/v1alpha1",
  "kind": "VaultServerConfiguration",
  "usePodServiceAccountForCSIDriver": true,
  "policyControllerServiceAccountName": "%s",
  "tokenReviewerServiceAccountName": "%s",
  "authPath": "kubernetes"
}
`

func (c *VaultController) ensureAppBinding(vs *api.VaultServer, v Vault) error {
	meta := metav1.ObjectMeta{
		Name:      vs.OffshootName(),
		Namespace: vs.Namespace,
	}
	sr, err := v.GetServerTLS()
	if err != nil {
		return err
	}
	caBundle, ok := sr.Data["ca.crt"]
	if !ok {
		return errors.New("ca bundle not found in server tls secret")
	}

	vsConf := vaultconfig.VaultServerConfiguration{
		TypeMeta: metav1.TypeMeta{
			APIVersion: vaultconfig.SchemeGroupVersion.String(),
			Kind: vaultconfig.ResourceKindVaultServerConfiguration,
		},
		UsePodServiceAccountForCSIDriver: true,
		AuthPath: "kubernetes",
		PolicyControllerServiceAccountName: vs.ServiceAccountForPolicyController(),
		TokenReviewerServiceAccountName: vs.ServiceAccountForTokenReviewer(),
	}
	dataConf, err := json.Marshal(vsConf)
	if err != nil {
		return err
	}
	_, _, err = appcat_util.CreateOrPatchAppBinding(c.appCatalogClient, meta, func(in *appcat.AppBinding) *appcat.AppBinding {
		in.Labels = vs.OffshootLabels()
		in.Spec.ClientConfig.Service = &appcat.ServiceReference{
			Name: v.GetService().Name,
		}
		in.Spec.ClientConfig.Service.Scheme = "https"
		in.Spec.ClientConfig.CABundle = caBundle
		in.Spec.ClientConfig.Service.Port = VaultClientPort
		in.Spec.Parameters = &runtime.RawExtension{
			Raw: []byte(dataConf),
		}
		util.EnsureOwnerRefToObject(in.GetObjectMeta(), util.AsOwner(vs))
		return in
	})
	return err
}
