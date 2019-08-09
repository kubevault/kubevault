package mongodb

import (
	"fmt"

	vaultapi "github.com/hashicorp/vault/api"
	"github.com/pkg/errors"
	"k8s.io/client-go/kubernetes"
	appcat_cs "kmodules.xyz/custom-resources/client/clientset/versioned/typed/appcatalog/v1alpha1"
	api "kubevault.dev/operator/apis/engine/v1alpha1"
)

type MongoDBRole struct {
	mdbRole      *api.MongoDBRole
	vaultClient  *vaultapi.Client
	kubeClient   kubernetes.Interface
	databasePath string
}

func NewMongoDBRole(kClient kubernetes.Interface, appClient appcat_cs.AppcatalogV1alpha1Interface, v *vaultapi.Client, mdbRole *api.MongoDBRole, databasePath string) (*MongoDBRole, error) {
	return &MongoDBRole{
		mdbRole:      mdbRole,
		vaultClient:  v,
		kubeClient:   kClient,
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

	payload := map[string]interface{}{
		"db_name":             mdb.DatabaseRef.Name,
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
