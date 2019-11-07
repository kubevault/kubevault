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

package mysql

import (
	"fmt"

	api "kubevault.dev/operator/apis/engine/v1alpha1"

	vaultapi "github.com/hashicorp/vault/api"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	appcat "kmodules.xyz/custom-resources/apis/appcatalog/v1alpha1"
	appcat_cs "kmodules.xyz/custom-resources/client/clientset/versioned/typed/appcatalog/v1alpha1"
)

type MySQLRole struct {
	mRole        *api.MySQLRole
	vaultClient  *vaultapi.Client
	kubeClient   kubernetes.Interface
	dbBinding    *appcat.AppBinding
	databasePath string
}

func NewMySQLRole(kClient kubernetes.Interface, appClient appcat_cs.AppcatalogV1alpha1Interface, v *vaultapi.Client, mRole *api.MySQLRole, databasePath string) (*MySQLRole, error) {
	ref := mRole.Spec.DatabaseRef
	dbBinding, err := appClient.AppBindings(ref.Namespace).Get(ref.Name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	return &MySQLRole{
		mRole:        mRole,
		vaultClient:  v,
		kubeClient:   kClient,
		dbBinding:    dbBinding,
		databasePath: databasePath,
	}, nil
}

// https://www.vaultproject.io/api/secret/databases/index.html#create-role
//
// CreateRole creates role
func (m *MySQLRole) CreateRole() error {
	name := m.mRole.RoleName()
	my := m.mRole.Spec

	path := fmt.Sprintf("/v1/%s/roles/%s", m.databasePath, name)
	req := m.vaultClient.NewRequest("POST", path)

	var dbName string
	if my.DatabaseRef != nil {
		if my.DatabaseRef.Name == "" {
			return errors.New("DatabaseRef.Name is empty")
		}
		if my.DatabaseRef.Namespace == "" {
			return errors.New("DatabaseRef.Namespace is empty")
		}
		dbName = api.GetDBNameFromAppBindingRef(my.DatabaseRef)
	} else if my.DatabaseName != "" {
		dbName = my.DatabaseName
	} else {
		return errors.New("both DatabaseRef and DatabaseName are empty")
	}
	payload := map[string]interface{}{
		"db_name":             dbName,
		"creation_statements": my.CreationStatements,
	}

	if len(my.RevocationStatements) > 0 {
		payload["revocation_statements"] = my.RevocationStatements
	}
	if my.DefaultTTL != "" {
		payload["default_ttl"] = my.DefaultTTL
	}
	if my.MaxTTL != "" {
		payload["max_ttl"] = my.MaxTTL
	}

	err := req.SetJSONBody(payload)
	if err != nil {
		return errors.WithStack(err)
	}

	_, err = m.vaultClient.RawRequest(req)
	if err != nil {
		return errors.Wrapf(err, "failed to create database role %s for config %s", name, my.DatabaseRef.Name)
	}
	return nil
}
