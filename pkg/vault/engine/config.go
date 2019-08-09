package engine

import (
	"fmt"

	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	appcat "kmodules.xyz/custom-resources/apis/appcatalog/v1alpha1"
	appcat_util "kmodules.xyz/custom-resources/client/clientset/versioned/typed/appcatalog/v1alpha1/util"
)

// https://www.vaultproject.io/api/secret/databases/index.html#configure-connection
// https:https://www.vaultproject.io/api/secret/databases/mysql-maria.html#configure-connection
//
// CreateMySQLConfig creates MySQL database configuration
func (secretEngineClient *SecretEngine) CreateMySQLConfig() error {
	config := secretEngineClient.secretEngine.Spec.MySQL
	if config == nil {
		return errors.New("MySQL database config is nil")
	}

	// Set Default plugin name, if config.PluginName is empty
	config.SetDefaults()

	dbAppRef := config.DatabaseRef
	dbApp, err := secretEngineClient.appClient.AppBindings(dbAppRef.Namespace).Get(dbAppRef.Name, metav1.GetOptions{})
	if err != nil {
		return errors.Wrap(err, "failed to get DatabaseAppBindng for MySQL database config")
	}

	connURL, err := dbApp.URLTemplate()
	if err != nil {
		return errors.Wrap(err, "failed to get MySQL database connection url")
	}

	path := fmt.Sprintf("/v1/%s/config/%s", secretEngineClient.path, dbApp.Name)
	req := secretEngineClient.vaultClient.NewRequest("POST", path)
	payload := map[string]interface{}{
		"plugin_name":    config.PluginName,
		"allowed_roles":  config.AllowedRoles,
		"connection_url": connURL,
	}

	if dbApp.Spec.Secret != nil {
		secret, err := secretEngineClient.kubeClient.CoreV1().Secrets(dbAppRef.Namespace).Get(dbApp.Spec.Secret.Name, metav1.GetOptions{})
		if err != nil {
			return errors.Wrap(err, "failed to get secret for MySQL database config")
		}

		data := make(map[string]interface{}, len(secret.Data))
		for k, v := range secret.Data {
			data[k] = string(v)
		}

		err = appcat_util.TransformCredentials(secretEngineClient.kubeClient, dbApp.Spec.SecretTransforms, data)
		if err != nil {
			return err
		}
		if v, ok := data[appcat.KeyUsername]; ok {
			payload[appcat.KeyUsername] = v
		}
		if v, ok := data[appcat.KeyPassword]; ok {
			payload[appcat.KeyPassword] = v
		}
	}

	if config.MaxOpenConnections > 0 {
		payload["max_open_connections"] = config.MaxOpenConnections
	}
	if config.MaxIdleConnections > 0 {
		payload["max_idle_connections"] = config.MaxIdleConnections
	}
	if config.MaxConnectionLifetime != "" {
		payload["max_connection_lifetime"] = config.MaxConnectionLifetime
	}

	err = req.SetJSONBody(payload)
	if err != nil {
		return errors.WithStack(err)
	}
	_, err = secretEngineClient.vaultClient.RawRequest(req)
	return err
}

// https://www.vaultproject.io/api/secret/databases/index.html#configure-connection
// https://www.vaultproject.io/api/secret/databases/mongodb.html#configure-connection
//
// CreateMongoDBConfig creates MongoDB database configuration
func (secretEngineClient *SecretEngine) CreateMongoDBConfig() error {
	config := secretEngineClient.secretEngine.Spec.MongoDB
	if config == nil {
		return errors.New("MongoDB database config is nil")
	}

	// Set Default plugin name, if config.PluginName is empty
	config.SetDefaults()

	dbAppRef := config.DatabaseRef
	dbApp, err := secretEngineClient.appClient.AppBindings(dbAppRef.Namespace).Get(dbAppRef.Name, metav1.GetOptions{})
	if err != nil {
		return errors.Wrap(err, "failed to get DatabaseAppBindng for MongoDB database config")
	}

	connURL, err := dbApp.URLTemplate()
	if err != nil {
		return errors.Wrap(err, "failed to get MongoDB database connection url")
	}

	path := fmt.Sprintf("/v1/%s/config/%s", secretEngineClient.path, dbApp.Name)
	req := secretEngineClient.vaultClient.NewRequest("POST", path)

	payload := map[string]interface{}{
		"plugin_name":    config.PluginName,
		"allowed_roles":  config.AllowedRoles,
		"connection_url": connURL,
	}

	if dbApp.Spec.Secret != nil {
		secret, err := secretEngineClient.kubeClient.CoreV1().Secrets(dbAppRef.Namespace).Get(dbApp.Spec.Secret.Name, metav1.GetOptions{})
		if err != nil {
			return errors.Wrap(err, "Failed to get secret for MongoDB database config")
		}

		data := make(map[string]interface{}, len(secret.Data))
		for k, v := range secret.Data {
			data[k] = string(v)
		}

		err = appcat_util.TransformCredentials(secretEngineClient.kubeClient, dbApp.Spec.SecretTransforms, data)
		if err != nil {
			return err
		}
		if v, ok := data[appcat.KeyUsername]; ok {
			payload[appcat.KeyUsername] = v
		}
		if v, ok := data[appcat.KeyPassword]; ok {
			payload[appcat.KeyPassword] = v
		}
	}

	if config.WriteConcern != "" {
		payload["write_concern"] = config.WriteConcern
	}

	err = req.SetJSONBody(payload)
	if err != nil {
		return errors.WithStack(err)
	}
	_, err = secretEngineClient.vaultClient.RawRequest(req)
	if err != nil {
		return errors.Wrap(err, "failed to create database config")
	}
	return nil
}
