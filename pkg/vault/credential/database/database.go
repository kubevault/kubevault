package database

import (
	vaultapi "github.com/hashicorp/vault/api"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	meta_util "kmodules.xyz/client-go/meta"
	appcat "kmodules.xyz/custom-resources/apis/appcatalog/v1alpha1"
	appcat_cs "kmodules.xyz/custom-resources/client/clientset/versioned/typed/appcatalog/v1alpha1"
	api "kubevault.dev/operator/apis/engine/v1alpha1"
	crd "kubevault.dev/operator/client/clientset/versioned"
	"kubevault.dev/operator/pkg/vault"
	databaserole "kubevault.dev/operator/pkg/vault/role/database"
	"kubevault.dev/operator/pkg/vault/secret"
	"kubevault.dev/operator/pkg/vault/secret/engines/database"
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
		r, err := cr.EngineV1alpha1().MongoDBRoles(ref.Namespace).Get(ref.Name, metav1.GetOptions{})
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
		r, err := cr.EngineV1alpha1().MySQLRoles(ref.Namespace).Get(ref.Name, metav1.GetOptions{})
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
		r, err := cr.EngineV1alpha1().PostgresRoles(ref.Namespace).Get(ref.Name, metav1.GetOptions{})
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

func (d *DBCredManager) GetOwnerReference() metav1.OwnerReference {
	trueVar := true
	return metav1.OwnerReference{
		APIVersion: api.SchemeGroupVersion.String(),
		Kind:       api.ResourceKindDatabaseAccessRequest,
		Name:       d.DBAccessReq.Name,
		UID:        d.DBAccessReq.UID,
		Controller: &trueVar,
	}
}
