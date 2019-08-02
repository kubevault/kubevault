package aws

import (
	vaultapi "github.com/hashicorp/vault/api"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	appcat "kmodules.xyz/custom-resources/apis/appcatalog/v1alpha1"
	appcat_cs "kmodules.xyz/custom-resources/client/clientset/versioned/typed/appcatalog/v1alpha1"
	api "kubevault.dev/operator/apis/engine/v1alpha1"
	crd "kubevault.dev/operator/client/clientset/versioned"
	"kubevault.dev/operator/pkg/vault"
	awsrole "kubevault.dev/operator/pkg/vault/role/aws"
	"kubevault.dev/operator/pkg/vault/secret"
	"kubevault.dev/operator/pkg/vault/secret/engines/aws"
)

type AWSCredManager struct {
	secret.SecretGetter

	AWSAccessReq *api.AWSAccessKeyRequest
	KubeClient   kubernetes.Interface
	VaultClient  *vaultapi.Client
}

func NewAWSCredentialManager(kClient kubernetes.Interface, appClient appcat_cs.AppcatalogV1alpha1Interface, cr crd.Interface, awsAKReq *api.AWSAccessKeyRequest) (*AWSCredManager, error) {
	vaultRef, roleName, err := GetVaultRefAndRole(cr, awsAKReq.Spec.RoleRef)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get vault app reference and vault role")
	}

	v, err := vault.NewClient(kClient, appClient, vaultRef)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create vault client")
	}

	awsPath, err := awsrole.GetAWSPath(appClient, vaultRef)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get aws path")
	}
	return &AWSCredManager{
		AWSAccessReq: awsAKReq,
		KubeClient:   kClient,
		VaultClient:  v,
		SecretGetter: aws.NewSecretGetter(v, awsPath, roleName, awsAKReq.Spec.UseSTS),
	}, nil
}

func GetVaultRefAndRole(cr crd.Interface, ref api.RoleRef) (*appcat.AppReference, string, error) {
	r, err := cr.EngineV1alpha1().AWSRoles(ref.Namespace).Get(ref.Name, metav1.GetOptions{})
	if err != nil {
		return nil, "", errors.Wrapf(err, "AWSRole %s/%s", ref.Namespace, ref.Name)
	}
	vAppRef := &appcat.AppReference{
		Namespace: r.Namespace,
		Name:      r.Spec.VaultRef.Name,
	}
	return vAppRef, r.RoleName(), nil
}

func (d *AWSCredManager) ParseCredential(credSecret *vaultapi.Secret) (map[string][]byte, error) {
	data := map[string][]byte{}
	for key, val := range credSecret.Data {
		if val == nil {
			data[key] = nil
		} else if v, ok := val.(string); ok {
			data[key] = []byte(v)
		} else {
			return nil, errors.Errorf("failed to convert interface{} to string for key %s", key)
		}
	}
	return data, nil
}

func (d *AWSCredManager) GetOwnerReference() metav1.OwnerReference {
	trueVar := true
	return metav1.OwnerReference{
		APIVersion: api.SchemeGroupVersion.String(),
		Kind:       api.ResourceKindAWSAccessKeyRequest,
		Name:       d.AWSAccessReq.Name,
		UID:        d.AWSAccessReq.UID,
		Controller: &trueVar,
	}
}
