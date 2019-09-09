package v1alpha1

import (
	"github.com/appscode/go/encoding/json/types"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	appcat "kmodules.xyz/custom-resources/apis/appcatalog/v1alpha1"
)

const (
	ResourceKindSecretEngine = "SecretEngine"
	ResourceSecretEngine     = "secretengine"
	ResourceSecretEngines    = "secretengines"
	EngineTypeAWS            = "aws"
	EngineTypeGCP            = "gcp"
	EngineTypeAzure          = "azure"
	EngineTypeDatabase       = "database"
)

// +genclient
// +k8s:openapi-gen=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// +kubebuilder:object:root=true
// +kubebuilder:resource:path=secretengines,singular=secretengine,categories={vault,appscode,all}
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.phase"
type SecretEngine struct {
	metav1.TypeMeta   `json:",inline,omitempty"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              SecretEngineSpec   `json:"spec,omitempty"`
	Status            SecretEngineStatus `json:"status,omitempty"`
}

type SecretEngineSpec struct {
	VaultRef core.LocalObjectReference `json:"vaultRef"`

	// Path defines the path used to enable this secret engine
	// +optional
	Path string `json:"path,omitempty"`

	SecretEngineConfiguration `json:",inline"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:object:root=true

type SecretEngineList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata, omitempty"`

	Items []SecretEngine `json:"items, omitempty"`
}

type SecretEngineConfiguration struct {
	AWS      *AWSConfiguration      `json:"aws,omitempty"`
	Azure    *AzureConfiguration    `json:"azure,omitempty"`
	GCP      *GCPConfiguration      `json:"gcp,omitempty"`
	Postgres *PostgresConfiguration `json:"postgres,omitempty"`
	MongoDB  *MongoDBConfiguration  `json:"mongodb,omitempty"`
	MySQL    *MySQLConfiguration    `json:"mysql,omitempty"`
}

// https://www.vaultproject.io/api/secret/aws/index.html#configure-root-iam-credentials
// AWSConfiguration contains information to communicate with AWS
type AWSConfiguration struct {
	// Specifies the secret containing AWS access key ID and secret access key
	// secret.Data:
	//	- access_key=<value>
	//  - secret_key=<value>
	CredentialSecret string `json:"credentialSecret"`

	// Specifies the AWS region
	Region string `json:"region"`

	// Specifies a custom HTTP IAM enminidpoint to use
	IAMEndpoint string `json:"iamEndpoint,omitempty"`

	//Specifies a custom HTTP STS endpoint to use
	STSEndpoint string `json:"stsEndpoint,omitempty"`

	// Number of max retries the client should use for recoverable errors.
	// The default (-1) falls back to the AWS SDK's default behavior
	MaxRetries *int `json:"maxRetries,omitempty"`

	LeaseConfig *LeaseConfig `json:"leaseConfig,omitempty"`
}

// https://www.vaultproject.io/api/secret/aws/index.html#configure-lease
// LeaseConfig contains lease configuration
type LeaseConfig struct {
	// Specifies the lease value provided as a string duration with time suffix.
	// "h" (hour) is the largest suffix.
	Lease string `json:"lease"`

	// Specifies the maximum lease value provided as a string duration with time suffix.
	// "h" (hour) is the largest suffix
	LeaseMax string `json:"leaseMax"`
}

// https://www.vaultproject.io/api/secret/gcp/index.html#write-config
// GCPConfiguration contains information to communicate with GCP
type GCPConfiguration struct {
	// Specifies the secret containing GCP credentials
	// secret.Data:
	//	- sa.json
	CredentialSecret string `json:"credentialSecret"`

	// Specifies default config TTL for long-lived credentials
	// (i.e. service account keys).
	// +optional
	TTL string `json:"ttl,omitempty"`

	// Specifies the maximum config TTL for long-lived
	// credentials (i.e. service account keys).
	// +optional
	MaxTTL string `json:"maxTTL,omitempty"`
}

// ref:
//	- https://www.vaultproject.io/api/secret/azure/index.html#configure-access

// AzureConfiguration contains information to communicate with Azure
type AzureConfiguration struct {

	// Specifies the secret name containing Azure credentials
	// secret.Data:
	// 	- subscription-id: <value>, The subscription id for the Azure Active Directory.
	//	- tenant-id: <value>, The tenant id for the Azure Active Directory.
	//	- client-id: <value>, The OAuth2 client id to connect to Azure.
	//	- client-secret: <value>, The OAuth2 client secret to connect to Azure.
	// +required
	CredentialSecret string `json:"credentialSecret"`

	// The Azure environment.
	// If not specified, Vault will use Azure Public Cloud.
	// +optional
	Environment string `json:"environment, omitempty"`
}

// PostgresConfiguration defines a PostgreSQL app configuration.
// https://www.vaultproject.io/api/secret/databases/index.html
// https://www.vaultproject.io/api/secret/databases/postgresql.html#configure-connection
type PostgresConfiguration struct {
	// Specifies the Postgres database appbinding reference
	DatabaseRef appcat.AppReference `json:"databaseRef"`

	// Specifies the name of the plugin to use for this connection.
	// Default plugin:
	//	- for postgres: postgresql-database-plugin
	PluginName string `json:"pluginName,omitempty"`

	// List of the roles allowed to use this connection.
	// Defaults to empty (no roles), if contains a "*" any role can use this connection.
	AllowedRoles []string `json:"allowedRoles,omitempty"`

	// Specifies the maximum number of open connections to the database.
	MaxOpenConnections int `json:"maxOpenConnections,omitempty"`

	// Specifies the maximum number of idle connections to the database.
	// A zero uses the value of max_open_connections and a negative value disables idle connections.
	// If larger than max_open_connections it will be reduced to be equal.
	MaxIdleConnections int `json:"maxIdleConnections,omitempty"`

	// Specifies the maximum amount of time a connection may be reused.
	// If <= 0s connections are reused forever.
	MaxConnectionLifetime string `json:"maxConnectionLifetime,omitempty"`
}

// MongoDBConfiguration defines a MongoDB app configuration.
// https://www.vaultproject.io/api/secret/databases/index.html
// https://www.vaultproject.io/api/secret/databases/mongodb.html#configure-connection
type MongoDBConfiguration struct {
	// Specifies the database appbinding reference
	DatabaseRef appcat.AppReference `json:"databaseRef"`

	// Specifies the name of the plugin to use for this connection.
	// Default plugin:
	//  - for mongodb: mongodb-database-plugin
	PluginName string `json:"pluginName,omitempty"`

	// List of the roles allowed to use this connection.
	// Defaults to empty (no roles), if contains a "*" any role can use this connection.
	AllowedRoles []string `json:"allowedRoles,omitempty"`

	// Specifies the MongoDB write concern. This is set for the entirety
	// of the session, maintained for the lifecycle of the plugin process.
	WriteConcern string `json:"writeConcern,omitempty"`
}

// MySQLConfiguration defines a MySQL app configuration.
// https://www.vaultproject.io/api/secret/databases/index.html
// https://www.vaultproject.io/api/secret/databases/mysql-maria.html#configure-connection
type MySQLConfiguration struct {
	// DatabaseRef refers to a MySQL/MariaDB database AppBinding in any namespace
	DatabaseRef appcat.AppReference `json:"databaseRef"`

	// Specifies the name of the plugin to use for this connection.
	// Default plugin:
	//  - for mysql: mysql-database-plugin
	PluginName string `json:"pluginName,omitempty"`

	// List of the roles allowed to use this connection.
	// Defaults to empty (no roles), if contains a "*" any role can use this connection.
	AllowedRoles []string `json:"allowedRoles,omitempty"`

	// Specifies the maximum number of open connections to the database.
	MaxOpenConnections int `json:"maxOpenConnections,omitempty"`

	// Specifies the maximum number of idle connections to the database.
	// A zero uses the value of max_open_connections and a negative value disables idle connections.
	// If larger than max_open_connections it will be reduced to be equal.
	MaxIdleConnections int `json:"maxIdleConnections,omitempty"`

	// Specifies the maximum amount of time a connection may be reused.
	// If <= 0s connections are reused forever.
	MaxConnectionLifetime string `json:"maxConnectionLifetime,omitempty"`
}

type SecretEnginePhase string

type SecretEngineStatus struct {
	Phase SecretEnginePhase `json:"phase,omitempty"`

	ObservedGeneration *types.IntHash `json:"observedGeneration,omitempty"`

	Conditions []SecretEngineCondition `json:"conditions,omitempty"`
}

type SecretEngineCondition struct {
	Type string `json:"type,omitempty"`

	Status core.ConditionStatus `json:"status,omitempty"`

	// The reason for the condition's.
	Reason string `json:"reason,omitempty"`

	// A human readable message indicating details about the transition.
	Message string `json:"message,omitempty"`
}
