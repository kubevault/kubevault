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
	"context"

	api "kubevault.dev/apimachinery/apis/engine/v1alpha1"
	crd "kubevault.dev/apimachinery/client/clientset/versioned"
	"kubevault.dev/operator/pkg/vault"
	databaserole "kubevault.dev/operator/pkg/vault/role/database"
	"kubevault.dev/operator/pkg/vault/secret"
	"kubevault.dev/operator/pkg/vault/secret/engines/database"

	vaultapi "github.com/hashicorp/vault/api"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	meta_util "kmodules.xyz/client-go/meta"
	appcat "kmodules.xyz/custom-resources/apis/appcatalog/v1alpha1"
	appcat_cs "kmodules.xyz/custom-resources/client/clientset/versioned/typed/appcatalog/v1alpha1"
)

type DBCredManager struct {
	secret.SecretGetter

	DBAccessReq *api.DatabaseAccessRequest
	KubeClient  kubernetes.Interface
	VaultClient *vaultapi.Client
}

func NewDatabaseCredentialManager(kClient kubernetes.Interface, appClient appcat_cs.AppcatalogV1alpha1Interface, cr crd.Interface, dbAR *api.DatabaseAccessRequest) (*DBCredManager, error) {
	vaultRef, roleName, dbPath, err := GetVaultRefAndRole(cr, dbAR.Spec.RoleRef)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get vault app reference and vault role")
	}

	v, err := vault.NewClient(kClient, appClient, vaultRef)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create vault client")
	}

	return &DBCredManager{
		DBAccessReq:  dbAR,
		KubeClient:   kClient,
		VaultClient:  v,
		SecretGetter: database.NewSecretGetter(v, dbPath, roleName),
	}, nil
}

func GetVaultRefAndRole(cr crd.Interface, ref api.RoleRef) (*appcat.AppReference, string, string, error) {
	switch ref.Kind {
	case api.ResourceKindMongoDBRole:
		r, err := cr.EngineV1alpha1().MongoDBRoles(ref.Namespace).Get(context.TODO(), ref.Name, metav1.GetOptions{})
		if err != nil {
			return nil, "", "", errors.Wrapf(err, "MongoDBRole %s/%s", ref.Namespace, ref.Name)
		}
		vAppRef := &appcat.AppReference{
			Namespace: r.Namespace,
			Name:      r.Spec.VaultRef.Name,
		}
		dbPath, err := databaserole.GetMongoDBDatabasePath(r)
		if err != nil {
			return nil, "", "", errors.Wrapf(err, "failed to get database path for MongoDBRole %s/%s", ref.Namespace, ref.Name)
		}
		return vAppRef, r.RoleName(), dbPath, nil

	case api.ResourceKindMySQLRole:
		r, err := cr.EngineV1alpha1().MySQLRoles(ref.Namespace).Get(context.TODO(), ref.Name, metav1.GetOptions{})
		if err != nil {
			return nil, "", "", errors.Wrapf(err, "MySQLRole %s/%s", ref.Namespace, ref.Name)
		}
		vAppRef := &appcat.AppReference{
			Namespace: r.Namespace,
			Name:      r.Spec.VaultRef.Name,
		}
		dbPath, err := databaserole.GetMySQLDatabasePath(r)
		if err != nil {
			return nil, "", "", errors.Wrapf(err, "failed to get database path for MySQLRole %s/%s", ref.Namespace, ref.Name)
		}
		return vAppRef, r.RoleName(), dbPath, nil

	case api.ResourceKindPostgresRole:
		r, err := cr.EngineV1alpha1().PostgresRoles(ref.Namespace).Get(context.TODO(), ref.Name, metav1.GetOptions{})
		if err != nil {
			return nil, "", "", errors.Wrapf(err, "PostgresRole %s/%s", ref.Namespace, ref.Name)
		}
		vAppRef := &appcat.AppReference{
			Namespace: r.Namespace,
			Name:      r.Spec.VaultRef.Name,
		}
		dbPath, err := databaserole.GetPostgresDatabasePath(r)
		if err != nil {
			return nil, "", "", errors.Wrapf(err, "failed to get database path for PostgresRole %s/%s", ref.Namespace, ref.Name)
		}
		return vAppRef, r.RoleName(), dbPath, nil

	case api.ResourceKindElasticsearchRole:
		r, err := cr.EngineV1alpha1().ElasticsearchRoles(ref.Namespace).Get(context.TODO(), ref.Name, metav1.GetOptions{})
		if err != nil {
			return nil, "", "", errors.Wrapf(err, "ElasticsearchRole %s/%s", ref.Namespace, ref.Name)
		}
		vAppRef := &appcat.AppReference{
			Namespace: r.Namespace,
			Name:      r.Spec.VaultRef.Name,
		}
		dbPath, err := databaserole.GetElasticsearchDatabasePath(r)
		if err != nil {
			return nil, "", "", errors.Wrapf(err, "failed to get database path for ElasticsearchRole %s/%s", ref.Namespace, ref.Name)
		}
		return vAppRef, r.RoleName(), dbPath, nil

	default:
		return nil, "", "", errors.Errorf("unknown or unsupported role kind '%s'", ref.Kind)
	}
}

func (d *DBCredManager) ParseCredential(credSecret *vaultapi.Secret) (map[string][]byte, error) {
	var cred struct {
		Password string `json:"password"`
		Username string `json:"username"`
	}

	err := meta_util.Decode(credSecret.Data, &cred)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse credential from vault secret")
	}
	return map[string][]byte{
		"username": []byte(cred.Username),
		"password": []byte(cred.Password),
	}, nil
}

func (d *DBCredManager) GetOwnerReference() *metav1.OwnerReference {
	return metav1.NewControllerRef(d.DBAccessReq, api.SchemeGroupVersion.WithKind(api.ResourceKindDatabaseAccessRequest))
}
