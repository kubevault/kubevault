package database

import (
	vaultapi "github.com/hashicorp/vault/api"
	api "github.com/kubedb/apimachinery/apis/authorization/v1alpha1"
	crd "github.com/kubedb/apimachinery/client/clientset/versioned"
	"github.com/kubevault/operator/pkg/vault"
	databaserole "github.com/kubevault/operator/pkg/vault/role/database"
	"github.com/kubevault/operator/pkg/vault/secret"
	"github.com/kubevault/operator/pkg/vault/secret/engines/database"
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
	vaultRef, roleName, err := GetVaultRefAndRole(cr, dbAR.Spec.RoleRef)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get vault app reference and vault role")
	}

	v, err := vault.NewClient(kClient, appClient, vaultRef)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create vault client")
	}

	dbPath, err := databaserole.GetDatabasePath(appClient, *vaultRef)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get database path")
	}
	return &DBCredManager{
		DBAccessReq:  dbAR,
		KubeClient:   kClient,
		VaultClient:  v,
		SecretGetter: database.NewSecretGetter(v, dbPath, roleName),
	}, nil
}

func GetVaultRefAndRole(cr crd.Interface, ref api.RoleReference) (*appcat.AppReference, string, error) {
	switch ref.Kind {
	case api.ResourceKindMongoDBRole:
		r, err := cr.AuthorizationV1alpha1().MongoDBRoles(ref.Namespace).Get(ref.Name, metav1.GetOptions{})
		if err != nil {
			return nil, "", errors.Wrapf(err, "MongoDBRole %s/%s", ref.Namespace, ref.Name)
		}
		return r.Spec.AuthManagerRef, r.RoleName(), nil

	case api.ResourceKindMySQLRole:
		r, err := cr.AuthorizationV1alpha1().MySQLRoles(ref.Namespace).Get(ref.Name, metav1.GetOptions{})
		if err != nil {
			return nil, "", errors.Wrapf(err, "MySQLRole %s/%s", ref.Namespace, ref.Name)
		}
		return r.Spec.AuthManagerRef, r.RoleName(), nil

	case api.ResourceKindPostgresRole:
		r, err := cr.AuthorizationV1alpha1().PostgresRoles(ref.Namespace).Get(ref.Name, metav1.GetOptions{})
		if err != nil {
			return nil, "", errors.Wrapf(err, "PostgresRole %s/%s", ref.Namespace, ref.Name)
		}
		return r.Spec.AuthManagerRef, r.RoleName(), nil

	default:
		return nil, "", errors.Errorf("unknown or unsupported role kind '%s'", ref.Kind)
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
