/*
Copyright AppsCode Inc. and Contributors

Licensed under the AppsCode Community License 1.0.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    https://github.com/appscode/licenses/raw/1.0.0/AppsCode-Community-1.0.0.md

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package gcp

import (
	api "kubevault.dev/apimachinery/apis/engine/v1alpha1"
	"kubevault.dev/operator/pkg/vault"
	"kubevault.dev/operator/pkg/vault/role"

	"github.com/pkg/errors"
	"k8s.io/client-go/kubernetes"
	appcat "kmodules.xyz/custom-resources/apis/appcatalog/v1alpha1"
	appcat_cs "kmodules.xyz/custom-resources/client/clientset/versioned/typed/appcatalog/v1alpha1"
)

const DefaultGCPPath = "gcp"

type GCPRoleInterface interface {
	role.RoleInterface

	// DeleteRole deletes role
	DeleteRole(name string) (int, error)
}

func NewGCPRole(kClient kubernetes.Interface, appClient appcat_cs.AppcatalogV1alpha1Interface, role *api.GCPRole) (GCPRoleInterface, error) {
	vAppRef := &appcat.AppReference{
		Namespace: role.Namespace,
		Name:      role.Spec.VaultRef.Name,
	}
	vClient, err := vault.NewClient(kClient, appClient, vAppRef)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create vault api client")
	}

	gcpPath, err := GetGCPPath(role)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get gcp path")
	}
	return &GCPRole{
		kubeClient:  kClient,
		vaultClient: vClient,
		gcpRole:     role,
		gcpPath:     gcpPath,
	}, nil
}

// If gcp path does not exist, then use default gcp path
func GetGCPPath(role *api.GCPRole) (string, error) {
	if role.Spec.Path != "" {
		return role.Spec.Path, nil
	}
	return DefaultGCPPath, nil
}
