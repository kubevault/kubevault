/*
Copyright AppsCode Inc. and Contributors

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
	"time"

	"gomodules.xyz/x/encoding/json/types"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kmapi "kmodules.xyz/client-go/api/v1"
	mona "kmodules.xyz/monitoring-agent-api/api/v1"
	ofst "kmodules.xyz/offshoot-api/api/v1"
)

const (
	ResourceKindVaultServer = "VaultServer"
	ResourceVaultServer     = "vaultserver"
	ResourceVaultServers    = "vaultservers"
)

// +genclient
// +k8s:openapi-gen=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// +kubebuilder:object:root=true
// +kubebuilder:resource:path=vaultservers,singular=vaultserver,shortName=vs,categories={vault,appscode,all}
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Replicas",type="string",JSONPath=".spec.replicas"
// +kubebuilder:printcolumn:name="Version",type="string",JSONPath=".spec.version"
// +kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.phase"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
type VaultServer struct {
	metav1.TypeMeta   `json:",inline,omitempty"`
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
	Spec              VaultServerSpec   `json:"spec,omitempty" protobuf:"bytes,2,opt,name=spec"`
	Status            VaultServerStatus `json:"status,omitempty" protobuf:"bytes,3,opt,name=status"`
}

type VaultServerSpec struct {
	// Number of replicas to deploy for a Vault deployment.
	// If unspecified, defaults to 1.
	// +optional
	Replicas *int32 `json:"replicas,omitempty" protobuf:"varint,1,opt,name=replicas"`

	// Version of Vault server to be deployed.
	Version types.StrYo `json:"version" protobuf:"bytes,2,opt,name=version,casttype=gomodules.xyz/x/encoding/json/types.StrYo"`

	// Name of the ConfigMap for Vault's configuration
	// In this configMap contain extra config for vault
	// ConfigSource is an optional field to provide extra configuration for vault.
	// File name should be 'vault.hcl'.
	// If specified, this file will be appended to the controller configuration file.
	// +optional
	ConfigSource *core.VolumeSource `json:"configSource,omitempty" protobuf:"bytes,3,opt,name=configSource"`

	// DataSources is a list of Configmaps/Secrets in the same namespace as the VaultServer
	// object, which shall be mounted into the VaultServer Pods.
	// The data are mounted into /etc/vault/data/<name>.
	// The first data will be named as "data-0", second one will be named as "data-1" and so on.
	// +optional
	DataSources []core.VolumeSource `json:"dataSources,omitempty" protobuf:"bytes,4,rep,name=dataSources"`

	// TLS policy of vault nodes
	// +optional
	TLS *TLSPolicy `json:"tls,omitempty" protobuf:"bytes,5,opt,name=tls"`

	// backend storage configuration for vault
	Backend BackendStorageSpec `json:"backend" protobuf:"bytes,6,opt,name=backend"`

	// Unsealer configuration for vault
	// +optional
	Unsealer *UnsealerSpec `json:"unsealer,omitempty" protobuf:"bytes,7,opt,name=unsealer"`

	// Specifies the list of auth methods to enable
	// +optional
	AuthMethods []AuthMethod `json:"authMethods,omitempty" protobuf:"bytes,8,rep,name=authMethods"`

	// Monitor is used monitor database instance
	// +optional
	Monitor *mona.AgentSpec `json:"monitor,omitempty" protobuf:"bytes,9,opt,name=monitor"`

	// PodTemplate is an optional configuration for pods used to run vault
	// +optional
	PodTemplate ofst.PodTemplateSpec `json:"podTemplate,omitempty" protobuf:"bytes,10,opt,name=podTemplate"`

	// ServiceTemplate is an optional configuration for service used to expose vault
	// +optional
	ServiceTemplate ofst.ServiceTemplateSpec `json:"serviceTemplate,omitempty" protobuf:"bytes,11,opt,name=serviceTemplate"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:object:root=true

type VaultServerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
	Items           []VaultServer `json:"items,omitempty" protobuf:"bytes,2,rep,name=items"`
}

// +kubebuilder:validation:Enum=Processing;Uninitialized;Running;Sealed
type ClusterPhase string

const (
	ClusterPhaseProcessing    ClusterPhase = "Processing"
	ClusterPhaseUnInitialized ClusterPhase = "Uninitialized"
	ClusterPhaseRunning       ClusterPhase = "Running"
	ClusterPhaseSealed        ClusterPhase = "Sealed"
)

type VaultServerStatus struct {
	// ObservedGeneration is the most recent generation observed for this resource. It corresponds to the
	// resource's generation, which is updated on mutation by the API Server.
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty" protobuf:"varint,1,opt,name=observedGeneration"`

	// Phase indicates the state this Vault cluster jumps in.
	// +optional
	Phase ClusterPhase `json:"phase,omitempty" protobuf:"bytes,2,opt,name=phase,casttype=ClusterPhase"`

	// Initialized indicates if the Vault service is initialized.
	// +optional
	Initialized bool `json:"initialized,omitempty" protobuf:"varint,3,opt,name=initialized"`

	// ServiceName is the LB service for accessing vault nodes.
	// +optional
	ServiceName string `json:"serviceName,omitempty" protobuf:"bytes,4,opt,name=serviceName"`

	// ClientPort is the port for vault client to access.
	// It's the same on client LB service and vault nodes.
	// +optional
	ClientPort int64 `json:"clientPort,omitempty" protobuf:"varint,5,opt,name=clientPort"`

	// VaultStatus is the set of Vault node specific statuses: Active, Standby, and Sealed
	// +optional
	VaultStatus VaultStatus `json:"vaultStatus,omitempty" protobuf:"bytes,6,opt,name=vaultStatus"`

	// PodNames of updated Vault nodes. Updated means the Vault container image version
	// matches the spec's version.
	// +optional
	UpdatedNodes []string `json:"updatedNodes,omitempty" protobuf:"bytes,7,rep,name=updatedNodes"`

	// Represents the latest available observations of a VaultServer current state.
	// +optional
	Conditions []kmapi.Condition `json:"conditions,omitempty" protobuf:"bytes,8,rep,name=conditions"`

	// Status of the vault auth methods
	// +optional
	AuthMethodStatus []AuthMethodStatus `json:"authMethodStatus,omitempty" protobuf:"bytes,9,rep,name=authMethodStatus"`
}

type VaultStatus struct {
	// PodName of the active Vault node. Active node is unsealed.
	// Only active node can serve requests.
	// Vault service only points to the active node.
	// +optional
	Active string `json:"active,omitempty" protobuf:"bytes,1,opt,name=active"`

	// PodNames of the standby Vault nodes. Standby nodes are unsealed.
	// Standby nodes do not process requests, and instead redirect to the active Vault.
	// +optional
	Standby []string `json:"standby,omitempty" protobuf:"bytes,2,rep,name=standby"`

	// PodNames of Sealed Vault nodes. Sealed nodes MUST be unsealed to
	// become standby or leader.
	// +optional
	Sealed []string `json:"sealed,omitempty" protobuf:"bytes,3,rep,name=sealed"`

	// PodNames of Unsealed Vault nodes.
	// +optional
	Unsealed []string `json:"unsealed,omitempty" protobuf:"bytes,4,rep,name=unsealed"`
}

// TLSPolicy defines the TLS policy of the vault nodes
// If this is not set, operator will auto-gen TLS assets and secrets.
type TLSPolicy struct {
	// TLSSecret is the secret containing TLS certs used by each vault node
	// for the communication between the vault server and its clients.
	// The secret should contain three files:
	// 	- tls.crt
	// 	- tls.key
	//
	// The server certificate must allow the following wildcard domains:
	// 	- localhost
	// 	- *.<namespace>.pod
	// 	- <vaultServer-name>.<namespace>.svc
	TLSSecret string `json:"tlsSecret" protobuf:"bytes,1,opt,name=tlsSecret"`

	// CABundle is a PEM encoded CA bundle which will be used to validate the serving certificate.
	// +optional
	CABundle []byte `json:"caBundle,omitempty" protobuf:"bytes,2,opt,name=caBundle"`
}

// TODO : set defaults and validation
// BackendStorageSpec defines storage backend configuration of vault
type BackendStorageSpec struct {
	// ref: https://www.vaultproject.io/docs/configuration/storage/in-memory.html
	// +optional
	Inmem *InmemSpec `json:"inmem,omitempty" protobuf:"bytes,1,opt,name=inmem"`

	// +optional
	Etcd *EtcdSpec `json:"etcd,omitempty" protobuf:"bytes,2,opt,name=etcd"`

	// +optional
	Gcs *GcsSpec `json:"gcs,omitempty" protobuf:"bytes,3,opt,name=gcs"`

	// +optional
	S3 *S3Spec `json:"s3,omitempty" protobuf:"bytes,4,opt,name=s3"`

	// +optional
	Azure *AzureSpec `json:"azure,omitempty" protobuf:"bytes,5,opt,name=azure"`

	// +optional
	PostgreSQL *PostgreSQLSpec `json:"postgresql,omitempty" protobuf:"bytes,6,opt,name=postgresql"`

	// +optional
	MySQL *MySQLSpec `json:"mysql,omitempty" protobuf:"bytes,7,opt,name=mysql"`

	// +optional
	File *FileSpec `json:"file,omitempty" protobuf:"bytes,8,opt,name=file"`

	// +optional
	DynamoDB *DynamoDBSpec `json:"dynamodb,omitempty" protobuf:"bytes,9,opt,name=dynamodb"`

	// +optional
	Swift *SwiftSpec `json:"swift,omitempty" protobuf:"bytes,10,opt,name=swift"`

	// +optional
	Consul *ConsulSpec `json:"consul,omitempty" protobuf:"bytes,11,opt,name=consul"`

	// +optional
	Raft *RaftSpec `json:"raft,omitempty" protobuf:"bytes,12,opt,name=raft"`
}

// ref: https://www.vaultproject.io/docs/configuration/storage/consul.html
//
// ConsulSpec defines the configuration to set up consul as backend storage in vault
type ConsulSpec struct {
	// Specifies the address of the Consul agent to communicate with.
	// This can be an IP address, DNS record, or unix socket.
	// +optional
	Address string `json:"address,omitempty" protobuf:"bytes,1,opt,name=address"`

	// Specifies the check interval used to send health check information
	// back to Consul.
	// This is specified using a label suffix like "30s" or "1h".
	// +optional
	CheckTimeout string `json:"checkTimeout,omitempty" protobuf:"bytes,2,opt,name=checkTimeout"`

	// Specifies the Consul consistency mode.
	// Possible values are "default" or "strong".
	// +optional
	ConsistencyMode string `json:"consistencyMode,omitempty" protobuf:"bytes,3,opt,name=consistencyMode"`

	// Specifies whether Vault should register itself with Consul.
	// Possible values are "true" or "false"
	// +optional
	DisableRegistration string `json:"disableRegistration,omitempty" protobuf:"bytes,4,opt,name=disableRegistration"`

	// Specifies the maximum number of concurrent requests to Consul.
	// +optional
	MaxParallel string `json:"maxParallel,omitempty" protobuf:"bytes,5,opt,name=maxParallel"`

	// Specifies the path in Consul's key-value store
	// where Vault data will be stored.
	// +optional
	Path string `json:"path,omitempty" protobuf:"bytes,6,opt,name=path"`

	// Specifies the scheme to use when communicating with Consul.
	// This can be set to "http" or "https".
	// +optional
	Scheme string `json:"scheme,omitempty" protobuf:"bytes,7,opt,name=scheme"`

	// Specifies the name of the service to register in Consul.
	// +optional
	Service string `json:"service,omitempty" protobuf:"bytes,8,opt,name=service"`

	// Specifies a comma-separated list of tags
	// to attach to the service registration in Consul.
	// +optional
	ServiceTags string `json:"serviceTags,omitempty" protobuf:"bytes,9,opt,name=serviceTags"`

	// Specifies a service-specific address to set on the service registration
	// in Consul.
	// If unset, Vault will use what it knows to be the HA redirect address
	// - which is usually desirable.
	// Setting this parameter to "" will tell Consul to leverage the configuration
	// of the node the service is registered on dynamically.
	// +optional
	ServiceAddress string `json:"serviceAddress,omitempty" protobuf:"bytes,10,opt,name=serviceAddress"`

	// Specifies the secret name that contains ACL token with permission
	// to read and write from the path in Consul's key-value store.
	// secret data:
	//  - aclToken:<value>
	// +optional
	ACLTokenSecretName string `json:"aclTokenSecretName,omitempty" protobuf:"bytes,11,opt,name=aclTokenSecretName"`

	// Specifies the minimum allowed session TTL.
	// Consul server has a lower limit of 10s on the session TTL by default.
	// +optional
	SessionTTL string `json:"sessionTTL,omitempty" protobuf:"bytes,12,opt,name=sessionTTL"`

	// Specifies the wait time before a lock lock acquisition is made.
	// This affects the minimum time it takes to cancel a lock acquisition.
	// +optional
	LockWaitTime string `json:"lockWaitTime,omitempty" protobuf:"bytes,13,opt,name=lockWaitTime"`

	// Specifies the secret name that contains tls_ca_file, tls_cert_file and tls_key_file
	// for consul communication
	// Secret data:
	//  - ca.crt
	//  - client.crt
	//  - client.key
	// +optional
	TLSSecretName string `json:"tlsSecretName,omitempty" protobuf:"bytes,14,opt,name=tlsSecretName"`

	// Specifies the minimum TLS version to use.
	// Accepted values are "tls10", "tls11" or "tls12".
	// +optional
	TLSMinVersion string `json:"tlsMinVersion,omitempty" protobuf:"bytes,15,opt,name=tlsMinVersion"`

	// Specifies if the TLS host verification should be disabled.
	// It is highly discouraged that you disable this option.
	// +optional
	TLSSkipVerify bool `json:"tlsSkipVerify,omitempty" protobuf:"varint,16,opt,name=tlsSkipVerify"`
}

// ref: https://www.vaultproject.io/docs/configuration/storage/in-memory.html
type InmemSpec struct {
}

// TODO : set defaults and validation
// vault doc: https://www.vaultproject.io/docs/configuration/storage/etcd.html
//
// EtcdSpec defines configuration to set up etcd as backend storage in vault
type EtcdSpec struct {
	// Specifies the addresses of the etcd instances
	Address string `json:"address" protobuf:"bytes,1,opt,name=address"`

	// Specifies the version of the API to communicate with etcd
	// +optional
	EtcdApi string `json:"etcdApi,omitempty" protobuf:"bytes,2,opt,name=etcdApi"`

	// Specifies if high availability should be enabled
	// +optional
	HAEnable bool `json:"haEnable,omitempty" protobuf:"varint,3,opt,name=haEnable"`

	// Specifies the path in etcd where vault data will be stored
	// +optional
	Path string `json:"path,omitempty" protobuf:"bytes,4,opt,name=path"`

	// Specifies whether to sync list of available etcd services on startup
	// +optional
	Sync bool `json:"sync,omitempty" protobuf:"varint,5,opt,name=sync"`

	// Specifies the domain name to query for SRV records describing cluster endpoints
	// +optional
	DiscoverySrv string `json:"discoverySrv,omitempty" protobuf:"bytes,6,opt,name=discoverySrv"`

	// Specifies the secret name that contain username and password to use when authenticating with the etcd server
	// secret data:
	//  - username:<value>
	//  - password:<value>
	// +optional
	CredentialSecretName string `json:"credentialSecretName,omitempty" protobuf:"bytes,7,opt,name=credentialSecretName"`

	// Specifies the secret name that contains tls_ca_file, tls_cert_file and tls_key_file for etcd communication
	// secret data:
	//  - ca.crt
	//  - client.crt
	//  - client.key
	// +optional
	TLSSecretName string `json:"tlsSecretName,omitempty" protobuf:"bytes,8,opt,name=tlsSecretName"`
}

// vault doc: https://www.vaultproject.io/docs/configuration/storage/google-cloud-storage.html
//
// GcsSpec defines configuration to set up Google Cloud Storage as backend storage in vault
type GcsSpec struct {
	// Specifies the name of the bucket to use for storage.
	Bucket string `json:"bucket" protobuf:"bytes,1,opt,name=bucket"`

	// Specifies the maximum size (in kilobytes) to send in a single request. If set to 0,
	// it will attempt to send the whole object at once, but will not retry any failures.
	// +optional
	ChunkSize string `json:"chunkSize,omitempty" protobuf:"bytes,2,opt,name=chunkSize"`

	//  Specifies the maximum number of parallel operations to take place.
	// +optional
	MaxParallel int64 `json:"maxParallel,omitempty" protobuf:"varint,3,opt,name=maxParallel"`

	// Specifies if high availability mode is enabled.
	// +optional
	HAEnabled bool `json:"haEnabled,omitempty" protobuf:"varint,4,opt,name=haEnabled"`

	// Secret containing Google application credential
	// secret data:
	//  - sa.json:<value>
	// +optional
	CredentialSecret string `json:"credentialSecret,omitempty" protobuf:"bytes,5,opt,name=credentialSecret"`
}

// vault doc: https://www.vaultproject.io/docs/configuration/storage/s3.html
//
// S3Spec defines configuration to set up Amazon S3 Storage as backend storage in vault
type S3Spec struct {
	// Specifies the name of the bucket to use for storage.
	Bucket string `json:"bucket" protobuf:"bytes,1,opt,name=bucket"`

	// Specifies an alternative, AWS compatible, S3 endpoint.
	// +optional
	Endpoint string `json:"endpoint,omitempty" protobuf:"bytes,2,opt,name=endpoint"`

	// Specifies the AWS region
	// +optional
	Region string `json:"region,omitempty" protobuf:"bytes,3,opt,name=region"`

	// Specifies the secret name containing AWS access key and AWS secret key
	// secret data:
	//  - access_key=<value>
	//  - secret_key=<value>
	// +optional
	CredentialSecret string `json:"credentialSecret,omitempty" protobuf:"bytes,4,opt,name=credentialSecret"`

	// Specifies the secret name containing AWS session token
	// secret data:
	//  - session_token:<value>
	// +optional
	SessionTokenSecret string `json:"sessionTokenSecret,omitempty" protobuf:"bytes,5,opt,name=sessionTokenSecret"`

	// Specifies the maximum number of parallel operations to take place.
	// +optional
	MaxParallel int64 `json:"maxParallel,omitempty" protobuf:"varint,6,opt,name=maxParallel"`

	// Specifies whether to use host bucket style domains with the configured endpoint.
	// +optional
	ForcePathStyle bool `json:"forcePathStyle,omitempty" protobuf:"varint,7,opt,name=forcePathStyle"`

	// Specifies if SSL should be used for the endpoint connection
	// +optional
	DisableSSL bool `json:"disableSSL,omitempty" protobuf:"varint,8,opt,name=disableSSL"`
}

// vault doc: https://www.vaultproject.io/docs/configuration/storage/azure.html
//
// AzureSpec defines configuration to set up Google Cloud Storage as backend storage in vault
type AzureSpec struct {
	// Specifies the Azure Storage account name.
	AccountName string `json:"accountName" protobuf:"bytes,1,opt,name=accountName"`

	// Specifies the secret containing Azure Storage account key.
	// secret data:
	//  - account_key:<value>
	AccountKeySecret string `json:"accountKeySecret" protobuf:"bytes,2,opt,name=accountKeySecret"`

	// Specifies the Azure Storage Blob container name.
	Container string `json:"container" protobuf:"bytes,3,opt,name=container"`

	//  Specifies the maximum number of concurrent operations to take place.
	// +optional
	MaxParallel int64 `json:"maxParallel,omitempty" protobuf:"varint,4,opt,name=maxParallel"`
}

// vault doc: https://www.vaultproject.io/docs/configuration/storage/postgresql.html
//
// PostgreSQLSpec defines configuration to set up PostgreSQL storage as backend storage in vault
type PostgreSQLSpec struct {
	//Specifies the name of the secret containing the connection string to use to authenticate and connect to PostgreSQL.
	// A full list of supported parameters can be found in the pq library documentation(https://godoc.org/github.com/lib/pq#hdr-Connection_String_Parameters).
	// secret data:
	//  - connection_url:<data>
	ConnectionURLSecret string `json:"connectionURLSecret" protobuf:"bytes,1,opt,name=connectionURLSecret"`

	// Specifies the name of the table in which to write Vault data.
	// This table must already exist (Vault will not attempt to create it).
	// +optional
	Table string `json:"table,omitempty" protobuf:"bytes,2,opt,name=table"`

	//  Specifies the maximum number of concurrent requests to take place.
	// +optional
	MaxParallel int64 `json:"maxParallel,omitempty" protobuf:"varint,3,opt,name=maxParallel"`
}

// vault doc: https://www.vaultproject.io/docs/configuration/storage/mysql.html
//
// MySQLSpec defines configuration to set up MySQL Storage as backend storage in vault
type MySQLSpec struct {
	// Specifies the address of the MySQL host.
	// +optional
	Address string `json:"address" protobuf:"bytes,1,opt,name=address"`

	// Specifies the name of the database. If the database does not exist, Vault will attempt to create it.
	// +optional
	Database string `json:"database,omitempty" protobuf:"bytes,2,opt,name=database"`

	// Specifies the name of the table. If the table does not exist, Vault will attempt to create it.
	// +optional
	Table string `json:"table,omitempty" protobuf:"bytes,3,opt,name=table"`

	// Specifies the MySQL username and password to connect to the database
	// secret data:
	//  - username=<value>
	//  - password=<value>
	UserCredentialSecret string `json:"userCredentialSecret" protobuf:"bytes,4,opt,name=userCredentialSecret"`

	// Specifies the name of the secret containing the CA certificate to connect using TLS.
	// secret data:
	//  - tls_ca_file=<ca_cert>
	// +optional
	TLSCASecret string `json:"tlsCASecret,omitempty" protobuf:"bytes,5,opt,name=tlsCASecret"`

	//  Specifies the maximum number of concurrent requests to take place.
	// +optional
	MaxParallel int64 `json:"maxParallel,omitempty" protobuf:"varint,6,opt,name=maxParallel"`
}

// vault doc: https://www.vaultproject.io/docs/configuration/storage/filesystem.html
//
// FileSpec defines configuration to set up File system Storage as backend storage in vault
type FileSpec struct {
	// The absolute path on disk to the directory where the data will be stored.
	// If the directory does not exist, Vault will create it.
	Path string `json:"path" protobuf:"bytes,1,opt,name=path"`

	// volumeClaimTemplate is a claim that pods are allowed to reference.
	// The VaultServer controller is responsible for deploying the claim
	// and update the volumeMounts in the Vault server container in the template.
	VolumeClaimTemplate ofst.PersistentVolumeClaim `json:"volumeClaimTemplate" protobuf:"bytes,2,opt,name=volumeClaimTemplate"`
}

// vault doc: https://www.vaultproject.io/docs/configuration/storage/dynamodb.html
//
// DynamoDBSpec defines configuration to set up DynamoDB Storage as backend storage in vault
type DynamoDBSpec struct {
	// Specifies an alternative, AWS compatible, DynamoDB endpoint.
	// +optional
	Endpoint string `json:"endpoint,omitempty" protobuf:"bytes,1,opt,name=endpoint"`

	// Specifies the AWS region
	// +optional
	Region string `json:"region,omitempty" protobuf:"bytes,2,opt,name=region"`

	// Specifies whether this backend should be used to run Vault in high availability mode.
	// +optional
	HaEnabled bool `json:"haEnabled,omitempty" protobuf:"varint,3,opt,name=haEnabled"`

	// Specifies the maximum number of reads consumed per second on the table
	// +optional
	ReadCapacity int64 `json:"readCapacity,omitempty" protobuf:"varint,4,opt,name=readCapacity"`

	// Specifies the maximum number of writes performed per second on the table.
	// +optional
	WriteCapacity int64 `json:"writeCapacity,omitempty" protobuf:"varint,5,opt,name=writeCapacity"`

	// Specifies the name of the DynamoDB table in which to store Vault data.
	// If the specified table does not yet exist, it will be created during initialization.
	// default: vault-dynamodb-backend
	// +optional
	Table string `json:"table,omitempty" protobuf:"bytes,6,opt,name=table"`

	// Specifies the secret name containing AWS access key and AWS secret key
	// secret data:
	//  - access_key=<value>
	//  - secret_key=<value>
	// +optional
	CredentialSecret string `json:"credentialSecret,omitempty" protobuf:"bytes,7,opt,name=credentialSecret"`

	// Specifies the secret name containing AWS session token
	// secret data:
	//  - session_token:<value>
	// +optional
	SessionTokenSecret string `json:"sessionTokenSecret,omitempty" protobuf:"bytes,8,opt,name=sessionTokenSecret"`

	// Specifies the maximum number of parallel operations to take place.
	// +optional
	MaxParallel int64 `json:"maxParallel,omitempty" protobuf:"varint,9,opt,name=maxParallel"`
}

// vault doc: https://www.vaultproject.io/docs/configuration/storage/swift.html
//
// SwiftSpec defines configuration to set up Swift Storage as backend storage in vault
type SwiftSpec struct {
	// Specifies the OpenStack authentication endpoint.
	AuthURL string `json:"authURL" protobuf:"bytes,1,opt,name=authURL"`

	// Specifies the name of the Swift container.
	Container string `json:"container" protobuf:"bytes,2,opt,name=container"`

	// Specifies the name of the secret containing the OpenStack account/username and password
	// secret data:
	//  - username=<value>
	//  - password=<value>
	CredentialSecret string `json:"credentialSecret" protobuf:"bytes,3,opt,name=credentialSecret"`

	// Specifies the name of the tenant. If left blank, this will default to the default tenant of the username.
	// +optional
	Tenant string `json:"tenant,omitempty" protobuf:"bytes,4,opt,name=tenant"`

	// Specifies the name of the region.
	// +optional
	Region string `json:"region,omitempty" protobuf:"bytes,5,opt,name=region"`

	// Specifies the id of the tenant.
	// +optional
	TenantID string `json:"tenantID,omitempty" protobuf:"bytes,6,opt,name=tenantID"`

	// Specifies the name of the user domain.
	// +optional
	Domain string `json:"domain,omitempty" protobuf:"bytes,7,opt,name=domain"`

	// Specifies the name of the project's domain.
	// +optional
	ProjectDomain string `json:"projectDomain,omitempty" protobuf:"bytes,8,opt,name=projectDomain"`

	// Specifies the id of the trust.
	// +optional
	TrustID string `json:"trustID,omitempty" protobuf:"bytes,9,opt,name=trustID"`

	// Specifies storage URL from alternate authentication.
	// +optional
	StorageURL string `json:"storageURL,omitempty" protobuf:"bytes,10,opt,name=storageURL"`

	// Specifies secret containing auth token from alternate authentication.
	// secret data:
	//  - auth_token=<value>
	// +optional
	AuthTokenSecret string `json:"authTokenSecret,omitempty" protobuf:"bytes,11,opt,name=authTokenSecret"`

	//  Specifies the maximum number of concurrent requests to take place.
	// +optional
	MaxParallel int64 `json:"maxParallel,omitempty" protobuf:"varint,12,opt,name=maxParallel"`
}

// UnsealerSpec contain the configuration for auto vault initialize/unseal
type UnsealerSpec struct {
	// Total count of secret shares that exist
	// +optional
	SecretShares int64 `json:"secretShares,omitempty" protobuf:"varint,1,opt,name=secretShares"`

	// Minimum required secret shares to unseal
	// +optional
	SecretThreshold int64 `json:"secretThreshold,omitempty" protobuf:"varint,2,opt,name=secretThreshold"`

	// How often to attempt to unseal the vault instance
	// +optional
	RetryPeriodSeconds time.Duration `json:"retryPeriodSeconds,omitempty" protobuf:"varint,3,opt,name=retryPeriodSeconds,casttype=time.Duration"`

	// overwrite existing unseal keys and root tokens, possibly dangerous!
	// +optional
	OverwriteExisting bool `json:"overwriteExisting,omitempty" protobuf:"varint,4,opt,name=overwriteExisting"`

	// should the root token be stored in the key store (default true)
	// +optional
	StoreRootToken bool `json:"storeRootToken,omitempty" protobuf:"varint,5,opt,name=storeRootToken"`

	// mode contains unseal mechanism
	// +optional
	Mode ModeSpec `json:"mode,omitempty" protobuf:"bytes,6,opt,name=mode"`
}

// ModeSpec contain unseal mechanism
type ModeSpec struct {
	// +optional
	KubernetesSecret *KubernetesSecretSpec `json:"kubernetesSecret,omitempty" protobuf:"bytes,1,opt,name=kubernetesSecret"`

	// +optional
	GoogleKmsGcs *GoogleKmsGcsSpec `json:"googleKmsGcs,omitempty" protobuf:"bytes,2,opt,name=googleKmsGcs"`

	// +optional
	AwsKmsSsm *AwsKmsSsmSpec `json:"awsKmsSsm,omitempty" protobuf:"bytes,3,opt,name=awsKmsSsm"`

	// +optional
	AzureKeyVault *AzureKeyVault `json:"azureKeyVault,omitempty" protobuf:"bytes,4,opt,name=azureKeyVault"`
}

// KubernetesSecretSpec contain the fields that required to unseal using kubernetes secret
type KubernetesSecretSpec struct {
	SecretName string `json:"secretName" protobuf:"bytes,1,opt,name=secretName"`
}

// GoogleKmsGcsSpec contain the fields that required to unseal vault using google kms
type GoogleKmsGcsSpec struct {
	// The name of the Google Cloud KMS crypto key to use
	KmsCryptoKey string `json:"kmsCryptoKey" protobuf:"bytes,1,opt,name=kmsCryptoKey"`

	// The name of the Google Cloud KMS key ring to use
	KmsKeyRing string `json:"kmsKeyRing" protobuf:"bytes,2,opt,name=kmsKeyRing"`

	// The Google Cloud KMS location to use (eg. 'global', 'europe-west1')
	KmsLocation string `json:"kmsLocation" protobuf:"bytes,3,opt,name=kmsLocation"`

	// The Google Cloud KMS project to use
	KmsProject string `json:"kmsProject" protobuf:"bytes,4,opt,name=kmsProject"`

	// The name of the Google Cloud Storage bucket to store values in
	Bucket string `json:"bucket" protobuf:"bytes,5,opt,name=bucket"`

	// Secret containing Google application credential
	// secret data:
	//  - sa.json:<value>
	// +optional
	CredentialSecret string `json:"credentialSecret,omitempty" protobuf:"bytes,6,opt,name=credentialSecret"`
}

// AwsKmsSsmSpec contain the fields that required to unseal vault using aws kms ssm
type AwsKmsSsmSpec struct {
	// The ID or ARN of the AWS KMS key to encrypt values
	KmsKeyID string `json:"kmsKeyID" protobuf:"bytes,1,opt,name=kmsKeyID"`

	// +optional
	// An optional Key prefix for SSM Parameter store
	SsmKeyPrefix string `json:"ssmKeyPrefix,omitempty" protobuf:"bytes,2,opt,name=ssmKeyPrefix"`

	Region string `json:"region,omitempty" protobuf:"bytes,3,opt,name=region"`

	// Specifies the secret name containing AWS access key and AWS secret key
	// secret data:
	//  - access_key:<value>
	//  - secret_key:<value>
	// +optional
	CredentialSecret string `json:"credentialSecret,omitempty" protobuf:"bytes,4,opt,name=credentialSecret"`

	// Used to make AWS KMS requests. This is useful,
	// for example, when connecting to KMS over a VPC Endpoint.
	// If not set, Vault will use the default API endpoint for your region.
	Endpoint string `json:"endpoint,omitempty" protobuf:"bytes,5,opt,name=endpoint"`
}

// RaftSpec defines the configuration for the Raft integrated storage.
//
// https://www.vaultproject.io/docs/configuration/storage/raft
type RaftSpec struct {
	// Path specifies the filesystem path where the vault data gets stored.
	//
	// This value can be overriden by setting the VAULT_RAFT_PATH environment variable.
	Path string `json:"path" protobuf:"bytes,1,opt,name=path"`
}

// AzureKeyVault contain the fields that required to unseal vault using azure key vault
type AzureKeyVault struct {
	// Azure key vault url, for example https://myvault.vault.azure.net
	VaultBaseURL string `json:"vaultBaseURL" protobuf:"bytes,1,opt,name=vaultBaseURL"`

	// The cloud environment identifier
	// default: "AZUREPUBLICCLOUD"
	// +optional
	Cloud string `json:"cloud,omitempty" protobuf:"bytes,2,opt,name=cloud"`

	// The AAD Tenant ID
	TenantID string `json:"tenantID" protobuf:"bytes,3,opt,name=tenantID"`

	// Specifies the name of secret containing client cert and client cert password
	// secret data:
	//  - client-cert:<value>
	// 	- client-cert-password: <value>
	// +optional
	ClientCertSecret string `json:"clientCertSecret,omitempty" protobuf:"bytes,4,opt,name=clientCertSecret"`

	// Specifies the name of secret containing client id and client secret of AAD application
	// secret data:
	//  - client-id:<value>
	//  - client-secret:<value>
	// +optional
	AADClientSecret string `json:"aadClientSecret,omitempty" protobuf:"bytes,5,opt,name=aadClientSecret"`

	// Use managed service identity for the virtual machine
	// +optional
	UseManagedIdentity bool `json:"useManagedIdentity,omitempty" protobuf:"varint,6,opt,name=useManagedIdentity"`
}

// +kubebuilder:validation:Enum=kubernetes;aws;gcp;userpass;cert;azure
type AuthMethodType string

const (
	AuthTypeKubernetes AuthMethodType = "kubernetes"
	AuthTypeAws        AuthMethodType = "aws"
	AuthTypeGcp        AuthMethodType = "gcp"
	AuthTypeUserPass   AuthMethodType = "userpass"
	AuthTypeCert       AuthMethodType = "cert"
	AuthTypeAzure      AuthMethodType = "azure"
)

// AuthMethod contains the information to enable vault auth method
// links: https://www.vaultproject.io/api/system/auth.html
type AuthMethod struct {
	//  Specifies the name of the authentication method type, such as "github" or "token".
	Type string `json:"type" protobuf:"bytes,1,opt,name=type"`

	// Specifies the path in which to enable the auth method.
	// Default value is the same as the 'type'
	Path string `json:"path" protobuf:"bytes,2,opt,name=path"`

	// Specifies a human-friendly description of the auth method.
	// +optional
	Description string `json:"description,omitempty" protobuf:"bytes,3,opt,name=description"`

	// Specifies configuration options for this auth method.
	// +optional
	Config *AuthConfig `json:"config,omitempty" protobuf:"bytes,4,opt,name=config"`

	// Specifies the name of the auth plugin to use based from the name in the plugin catalog.
	// Applies only to plugin methods.
	// +optional
	PluginName string `json:"pluginName,omitempty" protobuf:"bytes,5,opt,name=pluginName"`

	// Specifies if the auth method is a local only. Local auth methods are not replicated nor (if a secondary) removed by replication.
	// +optional
	Local bool `json:"local,omitempty" protobuf:"varint,6,opt,name=local"`
}

// +kubebuilder:validation:Enum=EnableSucceeded;EnableFailed;DisableSucceeded;DisableFailed
type AuthMethodEnableDisableStatus string

const (
	AuthMethodEnableSucceeded  AuthMethodEnableDisableStatus = "EnableSucceeded"
	AuthMethodEnableFailed     AuthMethodEnableDisableStatus = "EnableFailed"
	AuthMethodDisableSucceeded AuthMethodEnableDisableStatus = "DisableSucceeded"
	AuthMethodDisableFailed    AuthMethodEnableDisableStatus = "DisableFailed"
)

// AuthMethodStatus specifies the status of the auth method maintained by the auth method controller
type AuthMethodStatus struct {
	//  Specifies the name of the authentication method type, such as "github" or "token".
	Type string `json:"type" protobuf:"bytes,1,opt,name=type"`

	// Specifies the path in which to enable the auth method.
	Path string `json:"path" protobuf:"bytes,2,opt,name=path"`

	// Specifies whether auth method is enabled or not
	Status AuthMethodEnableDisableStatus `json:"status" protobuf:"bytes,3,opt,name=status,casttype=AuthMethodEnableDisableStatus"`

	// Specifies the reason why failed to enable auth method
	// +optional
	Reason string `json:"reason,omitempty" protobuf:"bytes,4,opt,name=reason"`
}

type AuthConfig struct {
	// The default lease duration, specified as a string duration like "5s" or "30m".
	// +optional
	DefaultLeaseTTL string `json:"defaultLeaseTTL,omitempty" protobuf:"bytes,1,opt,name=defaultLeaseTTL"`

	// The maximum lease duration, specified as a string duration like "5s" or "30m".
	// +optional
	MaxLeaseTTL string `json:"maxLeaseTTL,omitempty" protobuf:"bytes,2,opt,name=maxLeaseTTL"`

	// The name of the plugin in the plugin catalog to use.
	// +optional
	PluginName string `json:"pluginName,omitempty" protobuf:"bytes,3,opt,name=pluginName"`

	// List of keys that will not be HMAC'd by audit devices in the request data object.
	// +optional
	AuditNonHMACRequestKeys []string `json:"auditNonHMACRequestKeys,omitempty" protobuf:"bytes,4,rep,name=auditNonHMACRequestKeys"`

	// List of keys that will not be HMAC'd by audit devices in the response data object.
	// +optional
	AuditNonHMACResponseKeys []string `json:"auditNonHMACResponseKeys,omitempty" protobuf:"bytes,5,rep,name=auditNonHMACResponseKeys"`

	// Speficies whether to show this mount in the UI-specific listing endpoint.
	// +optional
	ListingVisibility string `json:"listingVisibility,omitempty" protobuf:"bytes,6,opt,name=listingVisibility"`

	// List of headers to whitelist and pass from the request to the backend.
	// +optional
	PassthroughRequestHeaders []string `json:"passthroughRequestHeaders,omitempty" protobuf:"bytes,7,rep,name=passthroughRequestHeaders"`
}
