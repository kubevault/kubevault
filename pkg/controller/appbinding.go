package controller

import (
	"encoding/json"
	"time"

	vaultconfig "github.com/kubevault/operator/apis/config/v1alpha1"
	api "github.com/kubevault/operator/apis/kubevault/v1alpha1"
	sa_util "github.com/kubevault/operator/pkg/util"
	"github.com/kubevault/operator/pkg/vault/util"
	"github.com/pkg/errors"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	appcat "kmodules.xyz/custom-resources/apis/appcatalog/v1alpha1"
	appcat_util "kmodules.xyz/custom-resources/client/clientset/versioned/typed/appcatalog/v1alpha1/util"
)

func (c *VaultController) ensureAppBindings(vs *api.VaultServer, v Vault) error {
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

	vClientConf := appcat.ClientConfig{
		Service: &appcat.ServiceReference{
			Name:   v.GetService().Name,
			Scheme: "https",
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
		AuthPath: "kubernetes",
		PolicyControllerServiceAccountName: vs.ServiceAccountForPolicyController(),
		TokenReviewerServiceAccountName:    vs.ServiceAccountForTokenReviewer(),
	}
	dataConf, err := json.Marshal(vsConf)
	if err != nil {
		return err
	}
	dataConf, _ = json.Marshal(string(dataConf))
	_, _, err = appcat_util.CreateOrPatchAppBinding(c.appCatalogClient, meta, func(in *appcat.AppBinding) *appcat.AppBinding {
		in.Labels = vs.OffshootLabels()
		in.Spec.ClientConfig = vClientConf
		in.Spec.Parameters = &runtime.RawExtension{
			Raw: dataConf,
		}
		util.EnsureOwnerRefToObject(in.GetObjectMeta(), util.AsOwner(vs))
		return in
	})
	if err != nil {
		return err
	}

	// AppBinding for policy controller
	pcMeta := metav1.ObjectMeta{
		Name:      vs.AppBindingNameForPolicyController(),
		Namespace: vs.Namespace,
	}

	// use policy controller service account for vault authentication
	secretName, err := sa_util.TryGetJwtTokenSecretNameFromServiceAccount(c.kubeClient, vs.ServiceAccountForPolicyController(), vs.Namespace, 2*time.Second, 30*time.Second)
	if err != nil {
		return errors.Wrapf(err, "failed to get jwt token secret name of service account(%s/%s)", vs.Namespace, vs.ServiceAccountForPolicyController())
	}

	k8sConf, err := json.Marshal(vaultconfig.KubernetesAuthConfiguration{
		Role: vs.PolicyNameForPolicyController(),
	})
	if err != nil {
		return err
	}

	// without this it got validation error
	// containing double guote in json causes this problem
	params, err := json.Marshal(string(k8sConf))
	if err != nil {
		return err
	}

	_, _, err = appcat_util.CreateOrPatchAppBinding(c.appCatalogClient, pcMeta, func(in *appcat.AppBinding) *appcat.AppBinding {
		in.Labels = vs.OffshootLabels()
		in.Spec.ClientConfig = vClientConf
		in.Spec.Secret = &core.LocalObjectReference{
			Name: secretName,
		}
		in.Spec.Parameters = &runtime.RawExtension{
			Raw: json.RawMessage(params),
		}
		util.EnsureOwnerRefToObject(in.GetObjectMeta(), util.AsOwner(vs))
		return in
	})
	if err != nil {
		return err
	}

	// AppBinding for auth method controller
	pcMeta = metav1.ObjectMeta{
		Name:      vs.AppBindingNameForAuthMethodController(),
		Namespace: vs.Namespace,
	}

	// use policy controller service account for vault authentication
	secretName, err = sa_util.TryGetJwtTokenSecretNameFromServiceAccount(c.kubeClient, vs.ServiceAccountForAuthMethodController(), vs.Namespace, 2*time.Second, 30*time.Second)
	if err != nil {
		return errors.Wrapf(err, "failed to get jwt token secret name of service account(%s/%s)", vs.Namespace, vs.ServiceAccountForAuthMethodController())
	}

	k8sConf, err = json.Marshal(vaultconfig.KubernetesAuthConfiguration{
		Role: vaultPolicyBindingForAuthMethod(vs).PolicyBindingName(),
	})
	if err != nil {
		return err
	}

	// without this it got validation error
	// containing double guote in json causes this problem
	params, err = json.Marshal(string(k8sConf))
	if err != nil {
		return err
	}

	_, _, err = appcat_util.CreateOrPatchAppBinding(c.appCatalogClient, pcMeta, func(in *appcat.AppBinding) *appcat.AppBinding {
		in.Labels = vs.OffshootLabels()
		in.Spec.ClientConfig = vClientConf
		in.Spec.Secret = &core.LocalObjectReference{
			Name: secretName,
		}
		in.Spec.Parameters = &runtime.RawExtension{
			Raw: params,
		}
		util.EnsureOwnerRefToObject(in.GetObjectMeta(), util.AsOwner(vs))
		return in
	})
	return err
}
