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

package database

import (
	"fmt"

	api "kubevault.dev/apimachinery/apis/engine/v1alpha1"
	vault "kubevault.dev/operator/pkg/vault"
	"kubevault.dev/operator/pkg/vault/role"
	"kubevault.dev/operator/pkg/vault/role/database/elasticsearch"
	"kubevault.dev/operator/pkg/vault/role/database/mongodb"
	"kubevault.dev/operator/pkg/vault/role/database/mysql"
	"kubevault.dev/operator/pkg/vault/role/database/postgres"

	vaultapi "github.com/hashicorp/vault/api"
	"github.com/pkg/errors"
	"k8s.io/client-go/kubernetes"
	appcat "kmodules.xyz/custom-resources/apis/appcatalog/v1alpha1"
	appcat_cs "kmodules.xyz/custom-resources/client/clientset/versioned/typed/appcatalog/v1alpha1"
)

const (
	DefaultDatabasePath = "database"
)

type DatabaseRole struct {
	role.RoleInterface
	vaultClient *vaultapi.Client
	path        string
}

func NewDatabaseRoleForPostgres(kClient kubernetes.Interface, appClient appcat_cs.AppcatalogV1alpha1Interface, role *api.PostgresRole) (DatabaseRoleInterface, error) {
	vAppRef := &appcat.AppReference{
		Namespace: role.Namespace,
		Name:      role.Spec.VaultRef.Name,
	}

	vClient, err := vault.NewClient(kClient, appClient, vAppRef)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	path, err := GetPostgresDatabasePath(role)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get database path")
	}

	pg, err := postgres.NewPostgresRole(kClient, appClient, vClient, role, path)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create postgres role client")
	}

	d := &DatabaseRole{
		RoleInterface: pg,
		path:          path,
		vaultClient:   vClient,
	}
	return d, nil
}

func NewDatabaseRoleForMysql(kClient kubernetes.Interface, appClient appcat_cs.AppcatalogV1alpha1Interface, role *api.MySQLRole) (DatabaseRoleInterface, error) {
	vAppRef := &appcat.AppReference{
		Namespace: role.Namespace,
		Name:      role.Spec.VaultRef.Name,
	}
	vClient, err := vault.NewClient(kClient, appClient, vAppRef)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	path, err := GetMySQLDatabasePath(role)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get database path")
	}

	m, err := mysql.NewMySQLRole(kClient, appClient, vClient, role, path)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create mysql role client")
	}

	d := &DatabaseRole{
		RoleInterface: m,
		path:          path,
		vaultClient:   vClient,
	}
	return d, nil
}

func NewDatabaseRoleForMongodb(kClient kubernetes.Interface, appClient appcat_cs.AppcatalogV1alpha1Interface, role *api.MongoDBRole) (DatabaseRoleInterface, error) {
	vAppRef := &appcat.AppReference{
		Namespace: role.Namespace,
		Name:      role.Spec.VaultRef.Name,
	}
	vClient, err := vault.NewClient(kClient, appClient, vAppRef)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	path, err := GetMongoDBDatabasePath(role)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get database path")
	}

	m, err := mongodb.NewMongoDBRole(kClient, appClient, vClient, role, path)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create mongodb role client")
	}
	d := &DatabaseRole{
		RoleInterface: m,
		path:          path,
		vaultClient:   vClient,
	}
	return d, nil
}

func NewDatabaseRoleForElasticsearch(kClient kubernetes.Interface, appClient appcat_cs.AppcatalogV1alpha1Interface, role *api.ElasticsearchRole) (DatabaseRoleInterface, error) {
	vAppRef := &appcat.AppReference{
		Namespace: role.Namespace,
		Name:      role.Spec.VaultRef.Name,
	}
	vClient, err := vault.NewClient(kClient, appClient, vAppRef)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	path, err := GetElasticsearchDatabasePath(role)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get database path")
	}

	es, err := elasticsearch.NewElasticsearchRole(kClient, appClient, vClient, role, path)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create elasticsearch role client")
	}
	d := &DatabaseRole{
		RoleInterface: es,
		path:          path,
		vaultClient:   vClient,
	}
	return d, nil
}

// EnableDatabase enables database secret engine
// It first checks whether database is enabled or not
func (d *DatabaseRole) EnableDatabase() error {
	enabled, err := d.IsDatabaseEnabled()
	if err != nil {
		return err
	}

	if enabled {
		return nil
	}

	err = d.vaultClient.Sys().Mount(d.path, &vaultapi.MountInput{
		Type: "database",
	})
	if err != nil {
		return err
	}
	return nil
}

// IsDatabaseEnabled checks whether database is enabled or not
func (d *DatabaseRole) IsDatabaseEnabled() (bool, error) {
	mnt, err := d.vaultClient.Sys().ListMounts()
	if err != nil {
		return false, errors.Wrap(err, "failed to list mounted secrets engines")
	}

	mntPath := d.path + "/"
	for k := range mnt {
		if k == mntPath {
			return true, nil
		}
	}
	return false, nil
}

// https://www.vaultproject.io/api/secret/databases/index.html#delete-role
//
// DeleteRole deletes role
// It doesn't give error even if respective role doesn't exist.
// But does give error (404) if the secret engine itself is missing in the given path.
func (d *DatabaseRole) DeleteRole(name string) (int, error) {
	path := fmt.Sprintf("/v1/%s/roles/%s", d.path, name)
	req := d.vaultClient.NewRequest("DELETE", path)

	resp, err := d.vaultClient.RawRequest(req)
	if err != nil {
		return resp.StatusCode, errors.Wrapf(err, "failed to delete database role %s", name)
	}
	return resp.StatusCode, nil
}

// If database path does not exist, then use default database path
func GetMySQLDatabasePath(role *api.MySQLRole) (string, error) {
	if role.Spec.Path != "" {
		return role.Spec.Path, nil
	}
	return DefaultDatabasePath, nil
}

// If database path does not exist, then use default database path
func GetMongoDBDatabasePath(role *api.MongoDBRole) (string, error) {
	if role.Spec.Path != "" {
		return role.Spec.Path, nil
	}
	return DefaultDatabasePath, nil
}

// If database path does not exist, then use default database path
func GetPostgresDatabasePath(role *api.PostgresRole) (string, error) {
	if role.Spec.Path != "" {
		return role.Spec.Path, nil
	}
	return DefaultDatabasePath, nil
}

// If database path does not exist, then use default database path
func GetElasticsearchDatabasePath(role *api.ElasticsearchRole) (string, error) {
	if role.Spec.Path != "" {
		return role.Spec.Path, nil
	}
	return DefaultDatabasePath, nil
}
