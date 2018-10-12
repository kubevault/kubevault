package controller

import (
	api "github.com/kubevault/operator/apis/kubevault/v1alpha1"
	"github.com/kubevault/operator/pkg/vault/util"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	appcat "kmodules.xyz/custom-resources/apis/appcatalog/v1alpha1"
	appcat_util "kmodules.xyz/custom-resources/client/clientset/versioned/typed/appcatalog/v1alpha1/util"
)

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

	appcat_util.CreateOrPatchAppBinding(c.appCatalogClient, meta, func(in *appcat.AppBinding) *appcat.AppBinding {
		in.Labels = vs.OffshootLabels()
		in.Spec.ClientConfig.Service = &appcat.ServiceReference{
			Name: v.GetService().Name,
		}
		in.Spec.ClientConfig.Service.Scheme = "https"
		in.Spec.ClientConfig.CABundle = caBundle
		in.Spec.ClientConfig.Service.Port = VaultClientPort
		util.EnsureOwnerRefToObject(in.GetObjectMeta(), util.AsOwner(vs))
		return in
	})
	return nil
}
