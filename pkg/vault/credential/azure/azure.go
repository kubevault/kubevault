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

package azure

import (
	"encoding/json"

	api "kubevault.dev/operator/apis/engine/v1alpha1"
	crd "kubevault.dev/operator/client/clientset/versioned"
	"kubevault.dev/operator/pkg/vault"
	azurerole "kubevault.dev/operator/pkg/vault/role/azure"
	"kubevault.dev/operator/pkg/vault/secret"
	azureengines "kubevault.dev/operator/pkg/vault/secret/engines/azure"

	vaultapi "github.com/hashicorp/vault/api"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	appcat "kmodules.xyz/custom-resources/apis/appcatalog/v1alpha1"
	appcat_cs "kmodules.xyz/custom-resources/client/clientset/versioned/typed/appcatalog/v1alpha1"
)

type AzureCredManager struct {
	secret.SecretGetter

	AzureAccessKeyReq *api.AzureAccessKeyRequest
	KubeClient        kubernetes.Interface
	VaultClient       *vaultapi.Client
}

func NewAzureCredentialManager(kClient kubernetes.Interface, appClient appcat_cs.AppcatalogV1alpha1Interface, cr crd.Interface, azureAKReq *api.AzureAccessKeyRequest) (*AzureCredManager, error) {
	role, err := GetVaultRefAndRole(cr, azureAKReq.Spec.RoleRef)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get vault app reference and AzureRole name")
	}
	vaultRef := &appcat.AppReference{
		Namespace: role.Namespace,
		Name:      role.Spec.VaultRef.Name,
	}

	vClient, err := vault.NewClient(kClient, appClient, vaultRef)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create vault client")
	}

	azurePath, err := azurerole.GetAzurePath(role)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get azure path")
	}

	return &AzureCredManager{
		SecretGetter:      azureengines.NewSecretGetter(vClient, azurePath, role.RoleName()),
		AzureAccessKeyReq: azureAKReq,
		KubeClient:        kClient,
		VaultClient:       vClient,
	}, nil
}

func GetVaultRefAndRole(cr crd.Interface, ref api.RoleRef) (*api.AzureRole, error) {
	r, err := cr.EngineV1alpha1().AzureRoles(ref.Namespace).Get(ref.Name, metav1.GetOptions{})
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get AzureRole %s/%s", ref.Namespace, ref.Name)
	}
	return r, nil
}

func (d *AzureCredManager) ParseCredential(credSecret *vaultapi.Secret) (map[string][]byte, error) {
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

func (d *AzureCredManager) GetOwnerReference() metav1.OwnerReference {
	trueVar := true
	return metav1.OwnerReference{
		APIVersion: api.SchemeGroupVersion.String(),
		Kind:       api.ResourceKindAzureAccessKeyRequest,
		Name:       d.AzureAccessKeyReq.Name,
		UID:        d.AzureAccessKeyReq.UID,
		Controller: &trueVar,
	}
}
