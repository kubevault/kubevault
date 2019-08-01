package vault

import (
	vaultapi "github.com/hashicorp/vault/api"
	"github.com/pkg/errors"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	appcat "kmodules.xyz/custom-resources/apis/appcatalog/v1alpha1"
	appcat_cs "kmodules.xyz/custom-resources/client/clientset/versioned/typed/appcatalog/v1alpha1"
	vaultauth "kubevault.dev/operator/pkg/vault/auth"
	vaultutil "kubevault.dev/operator/pkg/vault/util"
)

func NewClient(kc kubernetes.Interface, appc appcat_cs.AppcatalogV1alpha1Interface, namespace string, vAppRef core.LocalObjectReference) (*vaultapi.Client, error) {

	vApp, err := appc.AppBindings(namespace).Get(vAppRef.Name, metav1.GetOptions{})
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
