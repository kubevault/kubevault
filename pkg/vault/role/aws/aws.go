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
	"kubevault.dev/operator/pkg/vault"
	"kubevault.dev/operator/pkg/vault/role"

	"github.com/pkg/errors"
	"k8s.io/client-go/kubernetes"
	appcat "kmodules.xyz/custom-resources/apis/appcatalog/v1alpha1"
	appcat_cs "kmodules.xyz/custom-resources/client/clientset/versioned/typed/appcatalog/v1alpha1"
)

const DefaultAWSPath = "aws"

type AWSRoleInterface interface {
	role.RoleInterface

	// DeleteRole deletes role
	DeleteRole(name string) error
}

func NewAWSRole(kClient kubernetes.Interface, appClient appcat_cs.AppcatalogV1alpha1Interface, role *api.AWSRole) (AWSRoleInterface, error) {
	vAppRef := &appcat.AppReference{
		Namespace: role.Namespace,
		Name:      role.Spec.VaultRef.Name,
	}

	vClient, err := vault.NewClient(kClient, appClient, vAppRef)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create vault api client")
	}

	awsPath, err := GetAWSPath(role)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get aws path")
	}
	return &AWSRole{
		kubeClient:  kClient,
		vaultClient: vClient,
		awsRole:     role,
		awsPath:     awsPath,
	}, nil

}

// If aws path does not exist, then use default aws path
func GetAWSPath(role *api.AWSRole) (string, error) {

	if role.Spec.Path != "" {
		return role.Spec.Path, nil
	}
	return DefaultAWSPath, nil
}
