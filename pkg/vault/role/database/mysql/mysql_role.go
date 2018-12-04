package mysql

import (
	"encoding/json"
	"fmt"
	"net/url"
	"path/filepath"

	vaultapi "github.com/hashicorp/vault/api"
	api "github.com/kubedb/apimachinery/apis/authorization/v1alpha1"
	configapi "github.com/kubedb/apimachinery/apis/config/v1alpha1"
	"github.com/pkg/errors"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	appcat "kmodules.xyz/custom-resources/apis/appcatalog/v1alpha1"
	appcat_cs "kmodules.xyz/custom-resources/client/clientset/versioned/typed/appcatalog/v1alpha1"
)

type MySQLRole struct {
	config       *configapi.MySQLConfiguration
	secret       *core.Secret
	mRole        *api.MySQLRole
	vaultClient  *vaultapi.Client
	kubeClient   kubernetes.Interface
	databasePath string
	dbConnUrl    string
}

func NewMySQLRole(kClient kubernetes.Interface, appClient appcat_cs.AppcatalogV1alpha1Interface, v *vaultapi.Client, mRole *api.MySQLRole, databasePath string) (*MySQLRole, error) {
	ref := mRole.Spec.DatabaseRef
	dApp, err := appClient.AppBindings(mRole.Namespace).Get(ref.Name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	secretRef := dApp.Spec.Secret
	if secretRef == nil {
		return nil, errors.New("database secret is not provided")
	}

	sr, err := kClient.CoreV1().Secrets(mRole.Namespace).Get(secretRef.Name, metav1.GetOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "failed to get database secret")
	}

	cf := &configapi.MySQLConfiguration{}
	if dApp.Spec.Parameters != nil {
		err := json.Unmarshal(dApp.Spec.Parameters.Raw, cf)
		if err != nil {
			return nil, errors.Wrap(err, "failed to unmarshal database parameter")
		}
	}
	cf.SetDefaults()

	connUrl, err := getConnectionUrl(dApp)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get database connection url")
	}

	return &MySQLRole{
		config:       cf,
		secret:       sr,
		mRole:        mRole,
		vaultClient:  v,
		kubeClient:   kClient,
		databasePath: databasePath,
		dbConnUrl:    connUrl,
	}, nil
}

// https://www.vaultproject.io/api/secret/databases/index.html#configure-connection
// https:https://www.vaultproject.io/api/secret/databases/mysql-maria.html#configure-connection
//
// CreateConfig creates database configuration
func (m *MySQLRole) CreateConfig() error {
	if m.config == nil {
		return errors.New("database config is nil")
	}
	if m.secret == nil {
		return errors.New("database config is nil")
	}

	dRef := m.mRole.Spec.DatabaseRef
	path := fmt.Sprintf("/v1/%s/config/%s", m.databasePath, dRef.Name)
	req := m.vaultClient.NewRequest("POST", path)
	payload := map[string]interface{}{
		"plugin_name":    m.config.PluginName,
		"allowed_roles":  m.config.AllowedRoles,
		"connection_url": m.dbConnUrl,
	}

	data := m.secret.Data
	if val, ok := data["username"]; ok {
		payload["username"] = string(val)
	}
	if val, ok := data["password"]; ok {
		payload["password"] = string(val)
	}

	if m.config.MaxOpenConnections > 0 {
		payload["max_open_connections"] = m.config.MaxOpenConnections
	}
	if m.config.MaxIdleConnections > 0 {
		payload["max_idle_connections"] = m.config.MaxIdleConnections
	}
	if m.config.MaxConnectionLifetime != "" {
		payload["max_connection_lifetime"] = m.config.MaxConnectionLifetime
	}

	err := req.SetJSONBody(payload)
	if err != nil {
		return errors.WithStack(err)
	}
	_, err = m.vaultClient.RawRequest(req)
	return err
}

// https://www.vaultproject.io/api/secret/databases/index.html#create-role
//
// CreateRole creates role
func (m *MySQLRole) CreateRole() error {
	name := m.mRole.RoleName()
	my := m.mRole.Spec

	path := fmt.Sprintf("/v1/%s/roles/%s", m.databasePath, name)
	req := m.vaultClient.NewRequest("POST", path)

	payload := map[string]interface{}{
		"db_name":             my.DatabaseRef.Name,
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

func getConnectionUrl(app *appcat.AppBinding) (string, error) {
	c := app.Spec.ClientConfig
	if c.URL != nil {
		u, err := url.Parse(*c.URL)
		if err == nil {
			if u.User != nil {
				return "", errors.New("username/password must not be included in url, use {{field_name}} template instead and provide username and password in secret")
			}
		}
		return *c.URL, nil

	} else if c.Service != nil {
		srv := c.Service
		rawUrl := fmt.Sprintf("{{username}}:{{password}}@tcp(%s.%s.svc:%d)/", srv.Name, app.Namespace, srv.Port)
		if srv.Path != nil {
			rawUrl = filepath.Join(rawUrl, *srv.Path)
		}
		return rawUrl, nil

	} else {
		return "", errors.New("connection url is not provided")
	}
}
