package v1alpha1

import (
	"time"

	"github.com/appscode/go/encoding/json/types"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

type VaultServer struct {
	metav1.TypeMeta   `json:",inline,omitempty"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              VaultServerSpec   `json:"spec,omitempty"`
	Status            VaultServerStatus `json:"status,omitempty"`
}

type VaultServerSpec struct {
	// Number of nodes to deploy for a Vault deployment.
	// Default: 1.
	// +optional
	Nodes int32 `json:"nodes,omitempty"`

	// Version of Vault server to be deployed.
	Version types.StrYo `json:"version"`

	// Name of the ConfigMap for Vault's configuration
	// In this configMap contain extra config for vault
	// ConfigSource is an optional field to provide extra configuration for vault.
	// File name should be 'vault.hcl'.
	// If specified, this file will be appended to the controller configuration file.
	// +optional
	ConfigSource *core.VolumeSource `json:"configSource,omitempty"`

	// TLS policy of vault nodes
	// +optional
	TLS *TLSPolicy `json:"tls,omitempty"`

	// backend storage configuration for vault
	Backend BackendStorageSpec `json:"backend"`

	// Unsealer configuration for vault
	// +optional
	Unsealer *UnsealerSpec `json:"unsealer,omitempty"`

	// Specifies the list of auth methods to enable
	// +optional
	AuthMethods []AuthMethod `json:"authMethods,omitempty"`

	// Monitor is used monitor database instance
	// +optional
	Monitor *mona.AgentSpec `json:"monitor,omitempty"`

	// PodTemplate is an optional configuration for pods used to run vault
	// +optional
	PodTemplate ofst.PodTemplateSpec `json:"podTemplate,omitempty"`

	// ServiceTemplate is an optional configuration for service used to expose vault
	// +optional
	ServiceTemplate ofst.ServiceTemplateSpec `json:"serviceTemplate,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type VaultServerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []VaultServer `json:"items,omitempty"`
}

type ClusterPhase string

const (
	ClusterPhaseProcessing    ClusterPhase = "Processing"
	ClusterPhaseUnInitialized ClusterPhase = "Uninitialized"
	ClusterPhaseRunning       ClusterPhase = "Running"
	ClusterPhaseSealed        ClusterPhase = "Sealed"
)

type VaultServerStatus struct {
	// observedGeneration is the most recent generation observed for this resource. It corresponds to the
	// resource's generation, which is updated on mutation by the API Server.
	// +optional
	ObservedGeneration *types.IntHash `json:"observedGeneration,omitempty"`

	// Phase indicates the state this Vault cluster jumps in.
	// +optional
	Phase ClusterPhase `json:"phase,omitempty"`

	// Initialized indicates if the Vault service is initialized.
	// +optional
	Initialized bool `json:"initialized,omitempty"`

	// ServiceName is the LB service for accessing vault nodes.
	// +optional
	ServiceName string `json:"serviceName,omitempty"`

	// ClientPort is the port for vault client to access.
	// It's the same on client LB service and vault nodes.
	// +optional
	ClientPort int `json:"clientPort,omitempty"`

	// VaultStatus is the set of Vault node specific statuses: Active, Standby, and Sealed
	// +optional
	VaultStatus VaultStatus `json:"vaultStatus,omitempty"`

	// PodNames of updated Vault nodes. Updated means the Vault container image version
	// matches the spec's version.
	// +optional
	UpdatedNodes []string `json:"updatedNodes,omitempty"`

	// Represents the latest available observations of a VaultServer current state.
	// +optional
	Conditions []VaultServerCondition `json:"conditions,omitempty"`

	// Status of the vault auth methods
	// +optional
	AuthMethodStatus []AuthMethodStatus `json:"authMethodStatus,omitempty"`
}

type VaultServerConditionType string

// These are valid conditions of a VaultServer.
const (
	VaultServerConditionFailure VaultServerConditionType = "Failure"
)

// VaultServerCondition describes the state of a VaultServer at a certain point.
type VaultServerCondition struct {
	// Type of VaultServerCondition condition.
	// +optional
	Type VaultServerConditionType `json:"type,omitempty"`

	// Status of the condition, one of True, False, Unknown.
	// +optional
	Status core.ConditionStatus `json:"status,omitempty"`

	// The reason for the condition's.
	// +optional
	Reason string `json:"reason,omitempty"`

	// A human readable message indicating details about the transition.
	// +optional
	Message string `json:"message,omitempty"`
}

type VaultStatus struct {
	// PodName of the active Vault node. Active node is unsealed.
	// Only active node can serve requests.
	// Vault service only points to the active node.
	// +optional
	Active string `json:"active,omitempty"`

	// PodNames of the standby Vault nodes. Standby nodes are unsealed.
	// Standby nodes do not process requests, and instead redirect to the active Vault.
	// +optional
	Standby []string `json:"standby,omitempty"`

	// PodNames of Sealed Vault nodes. Sealed nodes MUST be unsealed to
	// become standby or leader.
	// +optional
	Sealed []string `json:"sealed,omitempty"`

	// PodNames of Unsealed Vault nodes.
	// +optional
	Unsealed []string `json:"unsealed,omitempty"`
}

// TLSPolicy defines the TLS policy of the vault nodes
// If this is not set, operator will auto-gen TLS assets and secrets.
type TLSPolicy struct {
	// TLSSecret is the secret containing TLS certs used by each vault node
	// for the communication between the vault server and its clients.
	// The secret should contain three files:
	//	- ca.crt
	// 	- server.crt
	// 	- server.key
	//
	// The server certificate must allow the following wildcard domains:
	// 	- localhost
	// 	- *.<namespace>.pod
	// 	- <vaultServer-name>.<namespace>.svc
	TLSSecret string `json:"tlsSecret"`
}

// TODO : set defaults and validation
// BackendStorageSpec defines storage backend configuration of vault
type BackendStorageSpec struct {
	// ref: https://www.vaultproject.io/docs/configuration/storage/in-memory.html
	// +optional
	Inmem bool `json:"inmem,omitempty"`

	// +optional
	Etcd *EtcdSpec `json:"etcd,omitempty"`

	// +optional
	Gcs *GcsSpec `json:"gcs,omitempty"`

	// +optional
	S3 *S3Spec `json:"s3,omitempty"`

	// +optional
	Azure *AzureSpec `json:"azure,omitempty"`

	// +optional
	PostgreSQL *PostgreSQLSpec `json:"postgreSQL,omitempty"`

	// +optional
	MySQL *MySQLSpec `json:"mySQL,omitempty"`

	// +optional
	File *FileSpec `json:"file,omitempty"`

	// +optional
	DynamoDB *DynamoDBSpec `json:"dynamoDB,omitempty"`

	// +optional
	Swift *SwiftSpec `json:"swift,omitempty"`
}

// TODO : set defaults and validation
// vault doc: https://www.vaultproject.io/docs/configuration/storage/etcd.html
//
// EtcdSpec defines configuration to set up etcd as backend storage in vault
type EtcdSpec struct {
	// Specifies the addresses of the etcd instances
	Address string `json:"address"`

	// Specifies the version of the API to communicate with etcd
	// +optional
	EtcdApi string `json:"etcdApi,omitempty"`

	// Specifies if high availability should be enabled
	// +optional
	HAEnable bool `json:"haEnable,omitempty"`

	// Specifies the path in etcd where vault data will be stored
	// +optional
	Path string `json:"path,omitempty"`

	// Specifies whether to sync list of available etcd services on startup
	// +optional
	Sync bool `json:"sync,omitempty"`

	// Specifies the domain name to query for SRV records describing cluster endpoints
	// +optional
	DiscoverySrv string `json:"discoverySrv,omitempty"`

	// Specifies the secret name that contain username and password to use when authenticating with the etcd server
	// secret data:
	//	- username:<value>
	//	- password:<value>
	// +optional
	CredentialSecretName string `json:"credentialSecretName,omitempty"`

	// Specifies the secret name that contains tls_ca_file, tls_cert_file and tls_key_file for etcd communication
	// +optional
	TLSSecretName string `json:"tlsSecretName,omitempty"`
}

// vault doc: https://www.vaultproject.io/docs/configuration/storage/google-cloud-storage.html
//
// GcsSpec defines configuration to set up Google Cloud Storage as backend storage in vault
type GcsSpec struct {
	// Specifies the name of the bucket to use for storage.
	Bucket string `json:"bucket"`

	// Specifies the maximum size (in kilobytes) to send in a single request. If set to 0,
	// it will attempt to send the whole object at once, but will not retry any failures.
	// +optional
	ChunkSize string `json:"chunkSize,omitempty"`

	//  Specifies the maximum number of parallel operations to take place.
	// +optional
	MaxParallel int `json:"maxParallel,omitempty"`

	// Specifies if high availability mode is enabled.
	// +optional
	HAEnabled bool `json:"haEnabled,omitempty"`

	// Secret containing Google application credential
	// secret data:
	//	- sa.json:<value>
	// +optional
	CredentialSecret string `json:"credentialSecret,omitempty"`
}

// vault doc: https://www.vaultproject.io/docs/configuration/storage/s3.html
//
// S3Spec defines configuration to set up Amazon S3 Storage as backend storage in vault
type S3Spec struct {
	// Specifies the name of the bucket to use for storage.
	Bucket string `json:"bucket"`

	// Specifies an alternative, AWS compatible, S3 endpoint.
	// +optional
	EndPoint string `json:"endPoint,omitempty"`

	// Specifies the AWS region
	// +optional
	Region string `json:"region,omitempty"`

	// Specifies the secret name containing AWS access key and AWS secret key
	// secret data:
	//	- access_key=<value>
	//  - secret_key=<value>
	// +optional
	CredentialSecret string `json:"credentialSecret,omitempty"`

	// Specifies the secret name containing AWS session token
	// secret data:
	//	- session_token:<value>
	// +optional
	SessionTokenSecret string `json:"sessionTokenSecret,omitempty"`

	// Specifies the maximum number of parallel operations to take place.
	// +optional
	MaxParallel int `json:"maxParallel,omitempty"`

	// Specifies whether to use host bucket style domains with the configured endpoint.
	// +optional
	S3ForcePathStyle bool `json:"s3ForcePathStyle,omitempty"`

	// Specifies if SSL should be used for the endpoint connection
	// +optional
	DisableSSL bool `json:"disableSSL,omitempty"`
}

// vault doc: https://www.vaultproject.io/docs/configuration/storage/azure.html
//
// AzureSpec defines configuration to set up Google Cloud Storage as backend storage in vault
type AzureSpec struct {
	// Specifies the Azure Storage account name.
	AccountName string `json:"accountName"`

	// Specifies the secret containing Azure Storage account key.
	// secret data:
	//	- account_key:<value>
	AccountKeySecret string `json:"accountKeySecret"`

	// Specifies the Azure Storage Blob container name.
	Container string `json:"container"`

	//  Specifies the maximum number of concurrent operations to take place.
	// +optional
	MaxParallel int `json:"maxParallel,omitempty"`
}

// vault doc: https://www.vaultproject.io/docs/configuration/storage/postgresql.html
//
// PostgreSQLSpec defines configuration to set up PostgreSQL storage as backend storage in vault
type PostgreSQLSpec struct {
	//Specifies the name of the secret containing the connection string to use to authenticate and connect to PostgreSQL.
	// A full list of supported parameters can be found in the pq library documentation(https://godoc.org/github.com/lib/pq#hdr-Connection_String_Parameters).
	// secret data:
	//	- connection_url:<data>
	ConnectionUrlSecret string `json:"connectionUrlSecret"`

	// Specifies the name of the table in which to write Vault data.
	// This table must already exist (Vault will not attempt to create it).
	// +optional
	Table string `json:"table,omitempty"`

	//  Specifies the maximum number of concurrent requests to take place.
	// +optional
	MaxParallel int `json:"maxParallel,omitempty"`
}

// vault doc: https://www.vaultproject.io/docs/configuration/storage/mysql.html
//
// MySQLSpec defines configuration to set up MySQL Storage as backend storage in vault
type MySQLSpec struct {
	// Specifies the address of the MySQL host.
	// +optional
	Address string `json:"address"`

	// Specifies the name of the database. If the database does not exist, Vault will attempt to create it.
	// +optional
	Database string `json:"database,omitempty"`

	// Specifies the name of the table. If the table does not exist, Vault will attempt to create it.
	// +optional
	Table string `json:"table,omitempty"`

	// Specifies the MySQL username and password to connect to the database
	// secret data:
	//	- username=<value>
	//	- password=<value>
	UserCredentialSecret string `json:"userCredentialSecret"`

	// Specifies the name of the secret containing the CA certificate to connect using TLS.
	// secret data:
	//	- tls_ca_file=<ca_cert>
	// +optional
	TLSCASecret string `json:"tlsCASecret,omitempty"`

	//  Specifies the maximum number of concurrent requests to take place.
	// +optional
	MaxParallel int `json:"maxParallel,omitempty"`
}

// vault doc: https://www.vaultproject.io/docs/configuration/storage/filesystem.html
//
// FileSpec defines configuration to set up File system Storage as backend storage in vault
type FileSpec struct {
	// The absolute path on disk to the directory where the data will be stored.
	// If the directory does not exist, Vault will create it.
	Path string `json:"path"`
}

// vault doc: https://www.vaultproject.io/docs/configuration/storage/dynamodb.html
//
// DynamoDBSpec defines configuration to set up DynamoDB Storage as backend storage in vault
type DynamoDBSpec struct {
	// Specifies an alternative, AWS compatible, DynamoDB endpoint.
	// +optional
	EndPoint string `json:"endPoint,omitempty"`

	// Specifies the AWS region
	// +optional
	Region string `json:"region,omitempty"`

	// Specifies whether this backend should be used to run Vault in high availability mode.
	// +optional
	HaEnabled bool `json:"haEnabled,omitempty"`

	// Specifies the maximum number of reads consumed per second on the table
	// +optional
	ReadCapacity int `json:"readCapacity,omiempty"`

	// Specifies the maximum number of writes performed per second on the table.
	// +optional
	WriteCapacity int `json:"writeCapacity,omitempty"`

	// Specifies the name of the DynamoDB table in which to store Vault data.
	// If the specified table does not yet exist, it will be created during initialization.
	// default: vault-dynamodb-backend
	// +optional
	Table string `json:"table,omitempty"`

	// Specifies the secret name containing AWS access key and AWS secret key
	// secret data:
	//	- access_key=<value>
	//  - secret_key=<value>
	// +optional
	CredentialSecret string `json:"credentialSecret,omitempty"`

	// Specifies the secret name containing AWS session token
	// secret data:
	//	- session_token:<value>
	// +optional
	SessionTokenSecret string `json:"sessionTokenSecret,omitempty"`

	// Specifies the maximum number of parallel operations to take place.
	// +optional
	MaxParallel int `json:"maxParallel,omitempty"`
}

// vault doc: https://www.vaultproject.io/docs/configuration/storage/swift.html
//
// SwiftSpec defines configuration to set up Swift Storage as backend storage in vault
type SwiftSpec struct {
	// Specifies the OpenStack authentication endpoint.
	AuthUrl string `json:"authUrl"`

	// Specifies the name of the Swift container.
	Container string `json:"container"`

	// Specifies the name of the secret containing the OpenStack account/username and password
	// secret data:
	//	- username=<value>
	//	- password=<value>
	CredentialSecret string `json:"credentialSecret"`

	// Specifies the name of the tenant. If left blank, this will default to the default tenant of the username.
	// +optional
	Tenant string `json:"tenant,omitempty"`

	// Specifies the name of the region.
	// +optional
	Region string `json:"region,omitempty"`

	// Specifies the id of the tenant.
	// +optional
	TenantID string `json:"tenantID,omitempty"`

	// Specifies the name of the user domain.
	// +optional
	Domain string `json:"domain,omitempty"`

	// Specifies the name of the project's domain.
	// +optional
	ProjectDomain string `json:"projectDomain,omitempty"`

	// Specifies the id of the trust.
	// +optional
	TrustID string `json:"trustID,omitempty"`

	// Specifies storage URL from alternate authentication.
	// +optional
	StorageUrl string `json:"storageUrl,omitempty"`

	// Specifies secret containing auth token from alternate authentication.
	// secret data:
	//	- auth_token=<value>
	// +optional
	AuthTokenSecret string `json:"authTokenSecret,omitempty"`

	//  Specifies the maximum number of concurrent requests to take place.
	// +optional
	MaxParallel int `json:"maxParallel,omitempty"`
}

// UnsealerSpec contain the configuration for auto vault initialize/unseal
type UnsealerSpec struct {
	// Total count of secret shares that exist
	// +optional
	SecretShares int `json:"secretShares,omitempty"`

	// Minimum required secret shares to unseal
	// +optional
	SecretThreshold int `json:"secretThreshold,omitempty"`

	// How often to attempt to unseal the vault instance
	// +optional
	RetryPeriodSeconds time.Duration `json:"retryPeriodSeconds,omitempty"`

	// overwrite existing unseal keys and root tokens, possibly dangerous!
	// +optional
	OverwriteExisting bool `json:"overwriteExisting,omitempty"`

	// InsecureSkipTLSVerify disables TLS certificate verification
	// +optional
	InsecureSkipTLSVerify bool `json:"insecureSkipTLSVerify,omitempty"`

	// CABundle is a PEM encoded CA bundle which will be used to validate the serving certificate.
	// +optional
	CABundle []byte `json:"caBundle,omitempty"`

	// should the root token be stored in the key store (default true)
	// +optional
	StoreRootToken bool `json:"storeRootToken,omitempty"`

	// mode contains unseal mechanism
	// +optional
	Mode ModeSpec `json:"mode,omitempty"`
}

// ModeSpec contain unseal mechanism
type ModeSpec struct {
	// +optional
	KubernetesSecret *KubernetesSecretSpec `json:"kubernetesSecret,omitempty"`

	// +optional
	GoogleKmsGcs *GoogleKmsGcsSpec `json:"googleKmsGcs,omitempty"`

	// +optional
	AwsKmsSsm *AwsKmsSsmSpec `json:"awsKmsSsm,omitempty"`

	// +optional
	AzureKeyVault *AzureKeyVault `json:"azureKeyVault,omitempty"`
}

// KubernetesSecretSpec contain the fields that required to unseal using kubernetes secret
type KubernetesSecretSpec struct {
	SecretName string `json:"secretName"`
}

// GoogleKmsGcsSpec contain the fields that required to unseal vault using google kms
type GoogleKmsGcsSpec struct {
	// The name of the Google Cloud KMS crypto key to use
	KmsCryptoKey string `json:"kmsCryptoKey"`

	// The name of the Google Cloud KMS key ring to use
	KmsKeyRing string `json:"kmsKeyRing"`

	// The Google Cloud KMS location to use (eg. 'global', 'europe-west1')
	KmsLocation string `json:"kmsLocation"`

	// The Google Cloud KMS project to use
	KmsProject string `json:"kmsProject"`

	// The name of the Google Cloud Storage bucket to store values in
	Bucket string `json:"bucket"`

	// Secret containing Google application credential
	// secret data:
	//	- sa.json:<value>
	// +optional
	CredentialSecret string `json:"credentialSecret,omitempty"`
}

// AwsKmsSsmSpec contain the fields that required to unseal vault using aws kms ssm
type AwsKmsSsmSpec struct {
	// The ID or ARN of the AWS KMS key to encrypt values
	KmsKeyID string `json:"kmsKeyID"`

	// +optional
	Region string `json:"region,omitempty"`

	// Specifies the secret name containing AWS access key and AWS secret key
	// secret data:
	//	- access_key:<value>
	//  - secret_key:<value>
	// +optional
	CredentialSecret string `json:"credentialSecret,omitempty"`
}

// AzureKeyVault contain the fields that required to unseal vault using azure key vault
type AzureKeyVault struct {
	// Azure key vault url, for example https://myvault.vault.azure.net
	VaultBaseUrl string `json:"vaultBaseUrl"`

	// The cloud environment identifier
	// default: "AZUREPUBLICCLOUD"
	// +optional
	Cloud string `json:"cloud,omitempty"`

	// The AAD Tenant ID
	TenantID string `json:"tenantID"`

	// Specifies the name of secret containing client cert and client cert password
	// secret data:
	//	- client-cert:<value>
	// 	- client-cert-password: <value>
	// +optional
	ClientCertSecret string `json:"clientCertSecret,omitempty"`

	// Specifies the name of secret containing client id and client secret of AAD application
	// secret data:
	//	- client-id:<value>
	//	- client-secret:<value>
	// +optional
	AADClientSecret string `json:"aadClientSecret,omitempty"`

	// Use managed service identity for the virtual machine
	// +optional
	UseManagedIdentity bool `json:"useManagedIdentity,omitempty"`
}

type AuthMethodType string

const (
	AuthTypeKubernetes AuthMethodType = "kubernetes"
	AuthTypeAws        AuthMethodType = "aws"
	AuthTypeUserPass   AuthMethodType = "userpass"
	AuthTypeCert       AuthMethodType = "cert"
)

// AuthMethod contains the information to enable vault auth method
// links: https://www.vaultproject.io/api/system/auth.html
type AuthMethod struct {
	//  Specifies the name of the authentication method type, such as "github" or "token".
	Type string `json:"type"`

	// Specifies the path in which to enable the auth method.
	// Default value is the same as the 'type'
	Path string `json:"path"`

	// Specifies a human-friendly description of the auth method.
	// +optional
	Description string `json:"description,omitempty"`

	// Specifies configuration options for this auth method.
	// +optional
	Config *AuthConfig `json:"config,omitempty"`

	// Specifies the name of the auth plugin to use based from the name in the plugin catalog.
	// Applies only to plugin methods.
	// +optional
	PluginName string `json:"pluginName,omitempty"`

	// Specifies if the auth method is a local only. Local auth methods are not replicated nor (if a secondary) removed by replication.
	// +optional
	Local bool `json:"local,omitempty"`
}

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
	Type string `json:"type"`

	// Specifies the path in which to enable the auth method.
	Path string `json:"path"`

	// Specifies whether auth method is enabled or not
	Status AuthMethodEnableDisableStatus `json:"status"`

	// Specifies the reason why failed to enable auth method
	// +optional
	Reason string `json:"reason,omitempty"`
}

type AuthConfig struct {
	// The default lease duration, specified as a string duration like "5s" or "30m".
	// +optional
	DefaultLeaseTTL string `json:"defaultLeaseTTL,omitempty"`

	// The maximum lease duration, specified as a string duration like "5s" or "30m".
	// +optional
	MaxLeaseTTL string `json:"maxLeaseTTL,omitempty"`

	// The name of the plugin in the plugin catalog to use.
	// +optional
	PluginName string `json:"pluginName,omitempty"`

	// List of keys that will not be HMAC'd by audit devices in the request data object.
	// +optional
	AuditNonHMACRequestKeys []string `json:"auditNonHMACRequestKeys,omitempty"`

	// List of keys that will not be HMAC'd by audit devices in the response data object.
	// +optional
	AuditNonHMACResponseKeys []string `json:"auditNonHMACResponseKeys,omitempty"`

	// Speficies whether to show this mount in the UI-specific listing endpoint.
	// +optional
	ListingVisibility string `json:"listingVisibility,omitempty"`

	// List of headers to whitelist and pass from the request to the backend.
	// +optional
	PassthroughRequestHeaders []string `json:"passthroughRequestHeaders,omitempty"`
}
