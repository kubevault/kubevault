package mongodb

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

type MongoDBRole struct {
	mdbRole      *api.MongoDBRole
	vaultClient  *vaultapi.Client
	dbBinding    *appcat.AppBinding
	kubeClient   kubernetes.Interface
	databasePath string
}

func NewMongoDBRole(kClient kubernetes.Interface, appClient appcat_cs.AppcatalogV1alpha1Interface, v *vaultapi.Client, mdbRole *api.MongoDBRole, databasePath string) (*MongoDBRole, error) {
	ref := mdbRole.Spec.DatabaseRef
	dbBinding, err := appClient.AppBindings(ref.Namespace).Get(ref.Name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	return &MongoDBRole{
		mdbRole:      mdbRole,
		vaultClient:  v,
		kubeClient:   kClient,
		dbBinding:    dbBinding,
		databasePath: databasePath,
	}, nil
}

// https://www.vaultproject.io/api/secret/databases/index.html#create-role
//
// CreateRole creates role
func (m *MongoDBRole) CreateRole() error {
	name := m.mdbRole.RoleName()
	mdb := m.mdbRole.Spec

	path := fmt.Sprintf("/v1/%s/roles/%s", m.databasePath, name)
	req := m.vaultClient.NewRequest("POST", path)

	var dbName string
	if mdb.DatabaseRef != nil {
		if mdb.DatabaseRef.Name == "" {
			return errors.New("DatabaseRef.Name is empty")
		}
		if mdb.DatabaseRef.Namespace == "" {
			return errors.New("DatabaseRef.Namespace is empty")
		}
		dbName = api.GetDBNameFromAppBindingRef(mdb.DatabaseRef)
	} else if mdb.DatabaseName != "" {
		dbName = mdb.DatabaseName
	} else {
		return errors.New("both DatabaseRef and DatabaseName are empty")
	}
	payload := map[string]interface{}{
		"db_name":             dbName,
		"creation_statements": mdb.CreationStatements,
	}

	if len(mdb.RevocationStatements) > 0 {
		payload["revocation_statements"] = mdb.RevocationStatements
	}
	if mdb.DefaultTTL != "" {
		payload["default_ttl"] = mdb.DefaultTTL
	}
	if mdb.MaxTTL != "" {
		payload["max_ttl"] = mdb.MaxTTL
	}

	err := req.SetJSONBody(payload)
	if err != nil {
		return errors.WithStack(err)
	}

	_, err = m.vaultClient.RawRequest(req)
	if err != nil {
		return errors.Wrapf(err, "failed to create database role %s for config %s", name, mdb.DatabaseRef.Name)
	}

	return nil
}
