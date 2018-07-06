package v1alpha1

import (
	"time"

	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	ResourceKindVaultServer = "VaultServer"
	ResourceVaultServer     = "vaultserver"
	ResourceVaultServers    = "vaultservers"

	// vault base image
	defaultBaseImage = "vault"
	// version format is "<upstream-version>-<our-version>"
	defaultVersion = "0.10.0"
)

type ClusterPhase string

const (
	ClusterPhaseInitial ClusterPhase = ""
	ClusterPhaseRunning              = "Running"
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
	Nodes int32 `json:"nodes,omitempty"`

	// Base image to use for a Vault deployment.
	BaseImage string `json:"baseImage"`

	// Version of Vault to be deployed.
	Version string `json:"version"`

	// Name of the ConfigMap for Vault's configuration
	// In this configMap contain extra config for vault
	ConfigMapName string `json:"configMapName,omitempty"`

	// TLS policy of vault nodes
	TLS *TLSPolicy `json:"TLS,omitempty"`

	// backend storage configuration for vault
	BackendStorage BackendStorageSpec `json:"backendStorage"`

	// unseal configuration for vault
	Unsealer *UnsealerSpec `json:"unsealer,omitempty"`

	// Compute Resources for vault container.
	Resources core.ResourceRequirements `json:"resources,omitempty"`

	// NodeSelector is a selector which must be true for the pod to fit on a node.
	// Selector which must match a node's labels for the pod to be scheduled on that node.
	// More info: https://kubernetes.io/docs/concepts/configuration/assign-pod-node/
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`

	// If specified, the pod's scheduling constraints
	// +optional
	Affinity *core.Affinity `json:"affinity,omitempty"`

	// If specified, the pod will be dispatched by specified scheduler.
	// If not specified, the pod will be dispatched by default scheduler.
	// +optional
	SchedulerName string `json:"schedulerName,omitempty"`

	// If specified, the pod's tolerations.
	// +optional
	Tolerations []core.Toleration `json:"tolerations,omitempty"`

	// ImagePullSecrets is an optional list of references to secrets in the same namespace to use for pulling any of the images used by this PodSpec.
	// If specified, these secrets will be passed to individual puller implementations for them to use. For example,
	// in the case of docker, only DockerConfig type secrets are honored.
	// More info: https://kubernetes.io/docs/concepts/containers/images#specifying-imagepullsecrets-on-a-pod
	// +optional
	ImagePullSecrets []core.LocalObjectReference `json:"imagePullSecrets,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type VaultServerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []VaultServer `json:"items,omitempty"`
}

type VaultServerStatus struct {
	// Phase indicates the state this Vault cluster jumps in.
	// Phase goes as one way as below:
	//   Initial -> Running
	Phase ClusterPhase `json:"phase,omitempty"`

	// Initialized indicates if the Vault service is initialized.
	Initialized bool `json:"initialized,omitempty"`

	// ServiceName is the LB service for accessing vault nodes.
	ServiceName string `json:"serviceName,omitempty"`

	// ClientPort is the port for vault client to access.
	// It's the same on client LB service and vault nodes.
	ClientPort int `json:"clientPort,omitempty"`

	// VaultStatus is the set of Vault node specific statuses: Active, Standby, and Sealed
	VaultStatus VaultStatus `json:"vaultStatus,omitempty"`

	// PodNames of updated Vault nodes. Updated means the Vault container image version
	// matches the spec's version.
	UpdatedNodes []string `json:"updatedNodes,omitempty"`
}

type VaultStatus struct {
	// PodName of the active Vault node. Active node is unsealed.
	// Only active node can serve requests.
	// Vault service only points to the active node.
	Active string `json:"active,omitempty"`

	// PodNames of the standby Vault nodes. Standby nodes are unsealed.
	// Standby nodes do not process requests, and instead redirect to the active Vault.
	Standby []string `json:"standby,omitempty"`

	// PodNames of Sealed Vault nodes. Sealed nodes MUST be unsealed to
	// become standby or leader.
	Sealed []string `json:"sealed,omitempty"`

	// PodNames of Unsealed Vault nodes.
	Unsealed []string `json:"unsealed,omitempty"`
}

// TLSPolicy defines the TLS policy of the vault nodes
type TLSPolicy struct {
	// StaticTLS enables user to use static x509 certificates and keys,
	// by putting them into Kubernetes secrets, and specifying them here.
	// If this is not set, operator will auto-gen TLS assets and secrets.
	Static *StaticTLS `json:"static,omitempty"`
}

type StaticTLS struct {
	// ServerSecret is the secret containing TLS certs used by each vault node
	// for the communication between the vault server and its clients.
	// The server secret should contain two files: server.crt and server.key
	// The server.crt file should only contain the server certificate.
	// It should not be concatenated with the optional ca certificate as allowed by https://www.vaultproject.io/docs/configuration/listener/tcp.html#tls_cert_file
	// The server certificate must allow the following wildcard domains:
	// localhost
	// *.<namespace>.pod
	// <vault-cluster-name>.<namespace>.svc
	ServerSecret string `json:"serverSecret,omitempty"`

	// ClientSecret is the secret containing the CA certificate
	// that will be used to verify the above server certificate
	// The ca secret should contain one file: vault-client-ca.crt
	ClientSecret string `json:"clientSecret,omitempty"`
}

// TODO : set defaults and validation
// BackendStorageSpec defines storage backend configuration of vault
type BackendStorageSpec struct {
	// ref: https://www.vaultproject.io/docs/configuration/storage/in-memory.html
	Inmem      bool            `json:"inmem,omitempty"`
	Etcd       *EtcdSpec       `json:"etcd,omitempty"`
	Gcs        *GcsSpec        `json:"gcs,omitempty"`
	S3         *S3Spec         `json:"s3,omitempty"`
	Azure      *AzureSpec      `json:"azure,omitempty"`
	PostgreSQL *PostgreSQLSpec `json:"postgreSQL,omitempty"`
	MySQL      *MySQLSpec      `json:"mySQL,omitempty"`
}

// TODO : set defaults and validation
// vault doc: https://www.vaultproject.io/docs/configuration/storage/etcd.html
//
// EtcdSpec defines configuration to set up etcd as backend storage in vault
type EtcdSpec struct {
	// Specifies the addresses of the etcd instances
	Address string `json:"address,omitempty"`

	// Specifies the version of the API to communicate with etcd
	EtcdApi string `json:"etcdApi,omitempty"`

	// Specifies if high availability should be enabled
	HAEnable bool `json:"haEnable,omitempty"`

	// Specifies the path in etcd where vault data will be stored
	Path string `json:"path,omitempty"`

	// Specifies whether to sync list of available etcd services on startup
	Sync bool `json:"sync,omitempty"`

	// Specifies the domain name to query for SRV records describing cluster endpoints
	DiscoverySrv string `json:"discoverySrv,omitempty"`

	// Specifies the secret name that contain username and password to use when authenticating with the etcd server
	// secret data:
	//	- username:<value>
	//	- password:<value>
	CredentialSecretName string `json:"credentialSecretName,omitempty"`

	// Specifies the secret name that contains tls_ca_file, tls_cert_file and tls_key_file for etcd communication
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
	ChunkSize string `json:"chunkSize,omitempty"`

	//  Specifies the maximum number of parallel operations to take place.
	MaxParallel int `json:"maxParallel,omitempty"`

	// Specifies if high availability mode is enabled.
	HAEnabled bool `json:"haEnabled,omitempty"`

	// Secret containing Google application credential
	// secret data:
	//	- sa.json:<value>
	CredentialSecret string `json:"credentialSecret,omitempty"`
}

// vault doc: https://www.vaultproject.io/docs/configuration/storage/s3.html
//
// S3Spec defines configuration to set up Amazon S3 Storage as backend storage in vault
type S3Spec struct {
	// Specifies the name of the bucket to use for storage.
	Bucket string `json:"bucket"`

	// Specifies an alternative, AWS compatible, S3 endpoint.
	EndPoint string `json:"endPoint,omitempty"`

	// Specifies the AWS region
	Region string `json:"region,omitempty"`

	// Specifies the secret name containing AWS access key and AWS secret key
	// secret data:
	//	- access_key=<value>
	//  - secret_key=<value>
	CredentialSecret string `json:"credentialSecret,omitempty"`

	// Specifies the secret name containing AWS session token
	// secret data:
	//	- session_token:<value>
	SessionTokenSecret string `json:"sessionTokenSecret,omitempty"`

	// Specifies the maximum number of parallel operations to take place.
	MaxParallel int `json:"maxParallel,omitempty"`

	// Specifies whether to use host bucket style domains with the configured endpoint.
	S3ForcePathStyle bool `json:"s3ForcePathStyle,omitempty"`

	// Specifies if SSL should be used for the endpoint connection
	DisableSSL bool `json:"disableSSL,omitempty"`
}

// vault doc: https://www.vaultproject.io/docs/configuration/storage/azure.html
//
// AzureSpec defines configuration to set up Google Cloud Storage as backend storage in vault
type AzureSpec struct {
	// Specifies the Azure Storage account name.
	AccountName string `json:"accountName"`

	// Specifies the Azure Storage account key.
	AccountKey string `json:"accountKey"`

	// Specifies the Azure Storage Blob container name.
	Container string `json:"container"`

	//  Specifies the maximum number of concurrent operations to take place.
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
	Table string `json:"table,omitempty"`

	//  Specifies the maximum number of concurrent requests to take place.
	MaxParallel int `json:"maxParallel,omitempty"`
}

// vault doc: https://www.vaultproject.io/docs/configuration/storage/mysql.html
//
// MySQLSpec defines configuration to set up MySQL Storage as backend storage in vault
type MySQLSpec struct {
	// Specifies the address of the MySQL host.
	Address string `json:"address,omitempty"`

	// Specifies the name of the database. If the database does not exist, Vault will attempt to create it.
	Database string `json:"database,omitempty"`

	// Specifies the name of the table. If the table does not exist, Vault will attempt to create it.
	Table string `json:"table,omitempty"`

	// Specifies the MySQL username and password to connect to the database
	// secret data:
	//	- username=<value>
	//	- password=<value>
	UserCredentialSecret string `json:"userCredentialSecret"`

	// Specifies the name of the secret containing the CA certificate to connect using TLS.
	// secret data:
	//	- ca=<ca_cert>
	TLSCASecret string `json:"tlsCASecret,omitempty"`

	//  Specifies the maximum number of concurrent requests to take place.
	MaxParallel int `json:"maxParallel,omitempty"`
}

// UnsealerSpec contain the configuration for auto vault initialize/unseal
type UnsealerSpec struct {
	// Total count of secret shares that exist
	SecretShares int `json:"secretShares,omitempty"`

	// Minimum required secret shares to unseal
	SecretThreshold int `json:"secretThreshold,omitempty"`

	// How often to attempt to unseal the vault instance
	RetryPeriodSeconds time.Duration `json:"retryPeriodSeconds,omitempty"`

	// overwrite existing unseal keys and root tokens, possibly dangerous!
	OverwriteExisting bool `json:"overwriteExisting,omitempty"`

	// To skip tls verification when communicating with vault server
	InsecureTLS bool `json:"insecureTLS,omitempty"`

	// Secret name containing self signed ca cert of vault
	VaultCASecret string `json:"vaultCASecret,omitempty"`

	// should the root token be stored in the key store (default true)
	StoreRootToken bool `json:"storeRootToken,omitempty"`

	// mode contains unseal mechanism
	Mode ModeSpec `json:"mode,omitempty"`
}

// ModeSpec contain unseal mechanism
type ModeSpec struct {
	KubernetesSecret *KubernetesSecretSpec `json:"kubernetesSecret,omitempty"`
	GoogleKmsGcs     *GoogleKmsGcsSpec     `json:"googleKmsGcs,omitempty"`
	AwsKmsSsm        *AwsKmsSsmSpec        `json:"awsKmsSsm,omitempty"`
	AzureKeyVault    *AzureKeyVault        `json:"azureKeyVault,omitempty"`
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
	CredentialSecret string `json:"credentialSecret,omitempty"`
}

// AwsKmsSsmSpec contain the fields that required to unseal vault using aws kms ssm
type AwsKmsSsmSpec struct {
	// The ID or ARN of the AWS KMS key to encrypt values
	KmsKeyID string `json:"kmsKeyID"`

	Region string `json:"region,omitempty"`

	// Specifies the secret name containing AWS access key and AWS secret key
	// secret data:
	//	- access_key:<value>
	//  - secret_key:<value>
	CredentialSecret string `json:"credentialSecret,omitempty"`
}

// AzureKeyVault contain the fields that required to unseal vault using azure key vault
type AzureKeyVault struct {
	// Azure key vault url, for example https://myvault.vault.azure.net
	VaultBaseUrl string `json:"vaultBaseUrl"`

	// The cloud environment identifier
	// default: "AZUREPUBLICCLOUD"
	Cloud string `json:"cloud,omitempty"`

	// The AAD Tenant ID
	TenantID string `json:"tenantID"`

	// Specifies the name of secret containing client cert and client cert password
	// secret data:
	//	- client-cert:<value>
	// 	- client-cert-password: <value>
	ClientCertSecret string `json:"clientCertSecret,omitempty"`

	// Specifies the name of secret containing client id and client secret of AAD application
	// secret data:
	//	- client-id:<value>
	//	- client-secret:<value>
	AADClientSecret string `json:"aadClientSecret,omitempty"`

	// Use managed service identity for the virtual machine
	UseManagedIdentity bool `json:"useManagedIdentity,omitempty"`
}

// TODO : use webhook?
// SetDefaults sets the default values for the vault spec and returns true if the spec was changed
func (v *VaultServer) SetDefaults() bool {
	changed := false
	vs := &v.Spec
	if vs.Nodes == 0 {
		vs.Nodes = 1
		changed = true
	}
	if len(vs.BaseImage) == 0 {
		vs.BaseImage = defaultBaseImage
		changed = true
	}
	if len(vs.Version) == 0 {
		vs.Version = defaultVersion
		changed = true
	}
	return changed
}
