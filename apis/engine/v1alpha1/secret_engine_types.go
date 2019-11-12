/*
Copyright The KubeVault Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
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
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
	Spec              SecretEngineSpec   `json:"spec,omitempty" protobuf:"bytes,2,opt,name=spec"`
	Status            SecretEngineStatus `json:"status,omitempty" protobuf:"bytes,3,opt,name=status"`
}

type SecretEngineSpec struct {
	VaultRef core.LocalObjectReference `json:"vaultRef" protobuf:"bytes,1,opt,name=vaultRef"`

	// Path defines the path used to enable this secret engine
	// +optional
	Path string `json:"path,omitempty" protobuf:"bytes,2,opt,name=path"`

	SecretEngineConfiguration `json:",inline" protobuf:"bytes,3,opt,name=secretEngineConfiguration"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:object:root=true

type SecretEngineList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	Items []SecretEngine `json:"items,omitempty" protobuf:"bytes,2,rep,name=items"`
}

type SecretEngineConfiguration struct {
	AWS      *AWSConfiguration      `json:"aws,omitempty" protobuf:"bytes,1,opt,name=aws"`
	Azure    *AzureConfiguration    `json:"azure,omitempty" protobuf:"bytes,2,opt,name=azure"`
	GCP      *GCPConfiguration      `json:"gcp,omitempty" protobuf:"bytes,3,opt,name=gcp"`
	Postgres *PostgresConfiguration `json:"postgres,omitempty" protobuf:"bytes,4,opt,name=postgres"`
	MongoDB  *MongoDBConfiguration  `json:"mongodb,omitempty" protobuf:"bytes,5,opt,name=mongodb"`
	MySQL    *MySQLConfiguration    `json:"mysql,omitempty" protobuf:"bytes,6,opt,name=mysql"`
}

// https://www.vaultproject.io/api/secret/aws/index.html#configure-root-iam-credentials
// AWSConfiguration contains information to communicate with AWS
type AWSConfiguration struct {
	// Specifies the secret containing AWS access key ID and secret access key
	// secret.Data:
	//	- access_key=<value>
	//  - secret_key=<value>
	CredentialSecret string `json:"credentialSecret" protobuf:"bytes,1,opt,name=credentialSecret"`

	// Specifies the AWS region
	Region string `json:"region" protobuf:"bytes,2,opt,name=region"`

	// Specifies a custom HTTP IAM enminidpoint to use
	IAMEndpoint string `json:"iamEndpoint,omitempty" protobuf:"bytes,3,opt,name=iamEndpoint"`

	//Specifies a custom HTTP STS endpoint to use
	STSEndpoint string `json:"stsEndpoint,omitempty" protobuf:"bytes,4,opt,name=stsEndpoint"`

	// Number of max retries the client should use for recoverable errors.
	// The default (-1) falls back to the AWS SDK's default behavior
	MaxRetries *int64 `json:"maxRetries,omitempty" protobuf:"varint,5,opt,name=maxRetries"`

	LeaseConfig *LeaseConfig `json:"leaseConfig,omitempty" protobuf:"bytes,6,opt,name=leaseConfig"`
}

// https://www.vaultproject.io/api/secret/aws/index.html#configure-lease
// LeaseConfig contains lease configuration
type LeaseConfig struct {
	// Specifies the lease value provided as a string duration with time suffix.
	// "h" (hour) is the largest suffix.
	Lease string `json:"lease" protobuf:"bytes,1,opt,name=lease"`

	// Specifies the maximum lease value provided as a string duration with time suffix.
	// "h" (hour) is the largest suffix
	LeaseMax string `json:"leaseMax" protobuf:"bytes,2,opt,name=leaseMax"`
}

// https://www.vaultproject.io/api/secret/gcp/index.html#write-config
// GCPConfiguration contains information to communicate with GCP
type GCPConfiguration struct {
	// Specifies the secret containing GCP credentials
	// secret.Data:
	//	- sa.json
	CredentialSecret string `json:"credentialSecret" protobuf:"bytes,1,opt,name=credentialSecret"`

	// Specifies default config TTL for long-lived credentials
	// (i.e. service account keys).
	// +optional
	TTL string `json:"ttl,omitempty" protobuf:"bytes,2,opt,name=ttl"`

	// Specifies the maximum config TTL for long-lived
	// credentials (i.e. service account keys).
	// +optional
	MaxTTL string `json:"maxTTL,omitempty" protobuf:"bytes,3,opt,name=maxTTL"`
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
	CredentialSecret string `json:"credentialSecret" protobuf:"bytes,1,opt,name=credentialSecret"`

	// The Azure environment.
	// If not specified, Vault will use Azure Public Cloud.
	// +optional
	Environment string `json:"environment,omitempty" protobuf:"bytes,2,opt,name=environment"`
}

// PostgresConfiguration defines a PostgreSQL app configuration.
// https://www.vaultproject.io/api/secret/databases/index.html
// https://www.vaultproject.io/api/secret/databases/postgresql.html#configure-connection
type PostgresConfiguration struct {
	// Specifies the Postgres database appbinding reference
	DatabaseRef appcat.AppReference `json:"databaseRef" protobuf:"bytes,1,opt,name=databaseRef"`

	// Specifies the name of the plugin to use for this connection.
	// Default plugin:
	//	- for postgres: postgresql-database-plugin
	PluginName string `json:"pluginName,omitempty" protobuf:"bytes,2,opt,name=pluginName"`

	// List of the roles allowed to use this connection.
	// Defaults to empty (no roles), if contains a "*" any role can use this connection.
	AllowedRoles []string `json:"allowedRoles,omitempty" protobuf:"bytes,3,rep,name=allowedRoles"`

	// Specifies the maximum number of open connections to the database.
	MaxOpenConnections int64 `json:"maxOpenConnections,omitempty" protobuf:"varint,4,opt,name=maxOpenConnections"`

	// Specifies the maximum number of idle connections to the database.
	// A zero uses the value of max_open_connections and a negative value disables idle connections.
	// If larger than max_open_connections it will be reduced to be equal.
	MaxIdleConnections int64 `json:"maxIdleConnections,omitempty" protobuf:"varint,5,opt,name=maxIdleConnections"`

	// Specifies the maximum amount of time a connection may be reused.
	// If <= 0s connections are reused forever.
	MaxConnectionLifetime string `json:"maxConnectionLifetime,omitempty" protobuf:"bytes,6,opt,name=maxConnectionLifetime"`
}

// MongoDBConfiguration defines a MongoDB app configuration.
// https://www.vaultproject.io/api/secret/databases/index.html
// https://www.vaultproject.io/api/secret/databases/mongodb.html#configure-connection
type MongoDBConfiguration struct {
	// Specifies the database appbinding reference
	DatabaseRef appcat.AppReference `json:"databaseRef" protobuf:"bytes,1,opt,name=databaseRef"`

	// Specifies the name of the plugin to use for this connection.
	// Default plugin:
	//  - for mongodb: mongodb-database-plugin
	PluginName string `json:"pluginName,omitempty" protobuf:"bytes,2,opt,name=pluginName"`

	// List of the roles allowed to use this connection.
	// Defaults to empty (no roles), if contains a "*" any role can use this connection.
	AllowedRoles []string `json:"allowedRoles,omitempty" protobuf:"bytes,3,rep,name=allowedRoles"`

	// Specifies the MongoDB write concern. This is set for the entirety
	// of the session, maintained for the lifecycle of the plugin process.
	WriteConcern string `json:"writeConcern,omitempty" protobuf:"bytes,4,opt,name=writeConcern"`
}

// MySQLConfiguration defines a MySQL app configuration.
// https://www.vaultproject.io/api/secret/databases/index.html
// https://www.vaultproject.io/api/secret/databases/mysql-maria.html#configure-connection
type MySQLConfiguration struct {
	// DatabaseRef refers to a MySQL/MariaDB database AppBinding in any namespace
	DatabaseRef appcat.AppReference `json:"databaseRef" protobuf:"bytes,1,opt,name=databaseRef"`

	// Specifies the name of the plugin to use for this connection.
	// Default plugin:
	//  - for mysql: mysql-database-plugin
	PluginName string `json:"pluginName,omitempty" protobuf:"bytes,2,opt,name=pluginName"`

	// List of the roles allowed to use this connection.
	// Defaults to empty (no roles), if contains a "*" any role can use this connection.
	AllowedRoles []string `json:"allowedRoles,omitempty" protobuf:"bytes,3,rep,name=allowedRoles"`

	// Specifies the maximum number of open connections to the database.
	MaxOpenConnections int64 `json:"maxOpenConnections,omitempty" protobuf:"varint,4,opt,name=maxOpenConnections"`

	// Specifies the maximum number of idle connections to the database.
	// A zero uses the value of max_open_connections and a negative value disables idle connections.
	// If larger than max_open_connections it will be reduced to be equal.
	MaxIdleConnections int64 `json:"maxIdleConnections,omitempty" protobuf:"varint,5,opt,name=maxIdleConnections"`

	// Specifies the maximum amount of time a connection may be reused.
	// If <= 0s connections are reused forever.
	MaxConnectionLifetime string `json:"maxConnectionLifetime,omitempty" protobuf:"bytes,6,opt,name=maxConnectionLifetime"`
}

type SecretEnginePhase string

type SecretEngineStatus struct {
	Phase SecretEnginePhase `json:"phase,omitempty" protobuf:"bytes,1,opt,name=phase,casttype=SecretEnginePhase"`

	ObservedGeneration int64 `json:"observedGeneration,omitempty" protobuf:"varint,2,opt,name=observedGeneration"`

	Conditions []SecretEngineCondition `json:"conditions,omitempty" protobuf:"bytes,3,rep,name=conditions"`
}

type SecretEngineCondition struct {
	Type string `json:"type,omitempty" protobuf:"bytes,1,opt,name=type"`

	Status core.ConditionStatus `json:"status,omitempty" protobuf:"bytes,2,opt,name=status,casttype=k8s.io/api/core/v1.ConditionStatus"`

	// The reason for the condition's.
	Reason string `json:"reason,omitempty" protobuf:"bytes,3,opt,name=reason"`

	// A human readable message indicating details about the transition.
	Message string `json:"message,omitempty" protobuf:"bytes,4,opt,name=message"`
}
