package gcp

import (
	"encoding/json"

	vaultapi "github.com/hashicorp/vault/api"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	appcat "kmodules.xyz/custom-resources/apis/appcatalog/v1alpha1"
	appcat_cs "kmodules.xyz/custom-resources/client/clientset/versioned/typed/appcatalog/v1alpha1"
	api "kubevault.dev/operator/apis/engine/v1alpha1"
	crd "kubevault.dev/operator/client/clientset/versioned"
	"kubevault.dev/operator/pkg/vault"
	gcprole "kubevault.dev/operator/pkg/vault/role/gcp"
	"kubevault.dev/operator/pkg/vault/secret"
	gcpengines "kubevault.dev/operator/pkg/vault/secret/engines/gcp"
)

type GCPCredManager struct {
	secret.SecretGetter

	GCPAccessKeyReq *api.GCPAccessKeyRequest
	KubeClient      kubernetes.Interface
	VaultClient     *vaultapi.Client
}

func NewGCPCredentialManager(kClient kubernetes.Interface, appClient appcat_cs.AppcatalogV1alpha1Interface, cr crd.Interface, gcpAKReq *api.GCPAccessKeyRequest) (*GCPCredManager, error) {
	vaultRef, roleName, secretType, err := GetVaultRefAndRole(cr, gcpAKReq.Spec.RoleRef)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get vault app reference and GCPRole name")
	}

	vClient, err := vault.NewClient(kClient, appClient, vaultRef)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create vault client")
	}

	gcpPath, err := gcprole.GetGCPPath(appClient, vaultRef)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get gcp path")
	}

	return &GCPCredManager{
		SecretGetter:    gcpengines.NewSecretGetter(vClient, gcpPath, roleName, secretType, gcpAKReq.Spec),
		GCPAccessKeyReq: gcpAKReq,
		KubeClient:      kClient,
		VaultClient:     vClient,
	}, nil
}

func GetVaultRefAndRole(cr crd.Interface, ref api.RoleReference) (*appcat.AppReference, string, api.GCPSecretType, error) {
	r, err := cr.EngineV1alpha1().GCPRoles(ref.Namespace).Get(ref.Name, metav1.GetOptions{})
	if err != nil {
		return nil, "", "", errors.Wrapf(err, "GCPRole %s/%s", ref.Namespace, ref.Name)
	}
	return r.Spec.Ref, r.RoleName(), r.Spec.SecretType, nil
}

func (d *GCPCredManager) ParseCredential(credSecret *vaultapi.Secret) (map[string][]byte, error) {
	data := map[string][]byte{}
	for key, value := range credSecret.Data {
		if value == nil {
			data[key] = nil
		} else if v, ok := value.(string); ok {
			data[key] = []byte(v)
		} else if v, ok := value.(json.Number); ok {
			data[key] = []byte(v)
		} else {
			return nil, errors.Errorf("Failed to convert interface{} to string for key %s", key)
		}
	}
	return data, nil
}

func (d *GCPCredManager) GetOwnerReference() metav1.OwnerReference {
	trueVar := true
	return metav1.OwnerReference{
		APIVersion: api.SchemeGroupVersion.String(),
		Kind:       api.ResourceKindGCPAccessKeyRequest,
		Name:       d.GCPAccessKeyReq.Name,
		UID:        d.GCPAccessKeyReq.UID,
		Controller: &trueVar,
	}
}
