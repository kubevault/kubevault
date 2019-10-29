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
package aws

import (
	api "kubevault.dev/operator/apis/engine/v1alpha1"
	crd "kubevault.dev/operator/client/clientset/versioned"
	"kubevault.dev/operator/pkg/vault"
	awsrole "kubevault.dev/operator/pkg/vault/role/aws"
	"kubevault.dev/operator/pkg/vault/secret"
	"kubevault.dev/operator/pkg/vault/secret/engines/aws"

	vaultapi "github.com/hashicorp/vault/api"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	appcat "kmodules.xyz/custom-resources/apis/appcatalog/v1alpha1"
	appcat_cs "kmodules.xyz/custom-resources/client/clientset/versioned/typed/appcatalog/v1alpha1"
)

type AWSCredManager struct {
	secret.SecretGetter

	AWSAccessReq *api.AWSAccessKeyRequest
	KubeClient   kubernetes.Interface
	VaultClient  *vaultapi.Client
}

func NewAWSCredentialManager(kClient kubernetes.Interface, appClient appcat_cs.AppcatalogV1alpha1Interface, cr crd.Interface, awsAKReq *api.AWSAccessKeyRequest) (*AWSCredManager, error) {
	role, err := GetVaultRefAndRole(cr, awsAKReq.Spec.RoleRef)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get vault app reference and vault role")
	}
	vaultRef := &appcat.AppReference{
		Namespace: role.Namespace,
		Name:      role.Spec.VaultRef.Name,
	}

	v, err := vault.NewClient(kClient, appClient, vaultRef)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create vault client")
	}

	awsPath, err := awsrole.GetAWSPath(role)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get aws path")
	}
	return &AWSCredManager{
		AWSAccessReq: awsAKReq,
		KubeClient:   kClient,
		VaultClient:  v,
		SecretGetter: aws.NewSecretGetter(v, awsPath, role.RoleName(), awsAKReq.Spec.UseSTS),
	}, nil
}

func GetVaultRefAndRole(cr crd.Interface, ref api.RoleRef) (*api.AWSRole, error) {
	r, err := cr.EngineV1alpha1().AWSRoles(ref.Namespace).Get(ref.Name, metav1.GetOptions{})
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get AWSRole %s/%s", ref.Namespace, ref.Name)
	}
	return r, nil
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
