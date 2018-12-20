package mongodb

import (
	"encoding/json"
	"fmt"

	vaultapi "github.com/hashicorp/vault/api"
	api "github.com/kubedb/apimachinery/apis/authorization/v1alpha1"
	configapi "github.com/kubedb/apimachinery/apis/config/v1alpha1"
	"github.com/pkg/errors"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	appcat "kmodules.xyz/custom-resources/apis/appcatalog/v1alpha1"
	appcat_cs "kmodules.xyz/custom-resources/client/clientset/versioned/typed/appcatalog/v1alpha1"
	appcat_util "kmodules.xyz/custom-resources/client/clientset/versioned/typed/appcatalog/v1alpha1/util"
)

type MongoDBRole struct {
	config       *configapi.MongoDBConfiguration
	secret       *core.Secret
	mdbRole      *api.MongoDBRole
	vaultClient  *vaultapi.Client
	kubeClient   kubernetes.Interface
	dbBinding    *appcat.AppBinding
	databasePath string
	dbConnURL    string
}

func NewMongoDBRole(kClient kubernetes.Interface, appClient appcat_cs.AppcatalogV1alpha1Interface, v *vaultapi.Client, mdbRole *api.MongoDBRole, databasePath string) (*MongoDBRole, error) {
	ref := mdbRole.Spec.DatabaseRef
	dbBinding, err := appClient.AppBindings(mdbRole.Namespace).Get(ref.Name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	secretRef := dbBinding.Spec.Secret
	if secretRef == nil {
		return nil, errors.New("database secret is not provided")
	}

	sr, err := kClient.CoreV1().Secrets(mdbRole.Namespace).Get(secretRef.Name, metav1.GetOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "failed to get database secret")
	}

	cf := &configapi.MongoDBConfiguration{}
	if dbBinding.Spec.Parameters != nil {
		err := json.Unmarshal(dbBinding.Spec.Parameters.Raw, cf)
		if err != nil {
			return nil, errors.Wrap(err, "failed to unmarshal database parameter")
		}
	}
	cf.SetDefaults()

	connurl, err := dbBinding.URLTemplate()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get database connection url")
	}

	return &MongoDBRole{
		mdbRole:      mdbRole,
		config:       cf,
		secret:       sr,
		vaultClient:  v,
		kubeClient:   kClient,
		dbBinding:    dbBinding,
		databasePath: databasePath,
		dbConnURL:    connurl,
	}, nil
}

// https://www.vaultproject.io/api/secret/databases/index.html#configure-connection
// https://www.vaultproject.io/api/secret/databases/mongodb.html#configure-connection
//
// CreateConfig creates database configuration
func (m *MongoDBRole) CreateConfig() error {
	if m.config == nil {
		return errors.New("database config is nil")
	}
	if m.secret == nil {
		return errors.New("database config is nil")
	}

	dRef := m.mdbRole.Spec.DatabaseRef
	path := fmt.Sprintf("/v1/%s/config/%s", m.databasePath, dRef.Name)
	req := m.vaultClient.NewRequest("POST", path)

	payload := map[string]interface{}{
		"plugin_name":    m.config.PluginName,
		"allowed_roles":  m.config.AllowedRoles,
		"connection_url": m.dbConnURL,
	}

	if m.config.WriteConcern != "" {
		payload["write_concern"] = m.config.WriteConcern
	}

	data := make(map[string]interface{}, len(m.secret.Data))
	for k, v := range m.secret.Data {
		data[k] = v
	}
	err := appcat_util.TransformCredentials(m.kubeClient, m.dbBinding.Spec.SecretTransforms, data)
	if err != nil {
		return err
	}

	if val, ok := data["username"]; ok {
		payload["username"] = val
	}
	if val, ok := data["password"]; ok {
		payload["password"] = val
	}

	err = req.SetJSONBody(payload)
	if err != nil {
		return errors.WithStack(err)
	}
	_, err = m.vaultClient.RawRequest(req)
	if err != nil {
		return errors.Wrap(err, "failed to create database config")
	}
	return nil
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
