package vault

import (
	vaultapi "github.com/hashicorp/vault/api"
	vaultauth "github.com/kubevault/operator/pkg/vault/auth"
	vaultutil "github.com/kubevault/operator/pkg/vault/util"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	appcat "kmodules.xyz/custom-resources/apis/appcatalog/v1alpha1"
	appcat_cs "kmodules.xyz/custom-resources/client/clientset/versioned/typed/appcatalog/v1alpha1"
)

func NewClient(kc kubernetes.Interface, appc appcat_cs.AppcatalogV1alpha1Interface, vAppRef *appcat.AppReference) (*vaultapi.Client, error) {
	if vAppRef == nil {
		return nil, errors.New(".spec.vaultAppRef is nil")
	}

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
