package database

import (
	"fmt"

	api "kubevault.dev/operator/apis/engine/v1alpha1"
	vault "kubevault.dev/operator/pkg/vault"
	"kubevault.dev/operator/pkg/vault/role"
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
// It's safe to call multiple time. It doesn't give
// error even if respective role doesn't exist
func (d *DatabaseRole) DeleteRole(name string) error {
	path := fmt.Sprintf("/v1/%s/roles/%s", d.path, name)
	req := d.vaultClient.NewRequest("DELETE", path)

	_, err := d.vaultClient.RawRequest(req)
	if err != nil {
		return errors.Wrapf(err, "failed to delete database role %s", name)
	}
	return nil
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
