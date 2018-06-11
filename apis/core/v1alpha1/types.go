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

	// PodPolicy defines the policy for pods owned by vault operator.
	// This field cannot be updated once the CR is created.
	PodPolicy *PodPolicy `json:"podPolicy,omitempty"`

	// Name of the ConfigMap for Vault's configuration
	// In this configMap contain extra config for vault
	ConfigMapName string `json:"configMapName,omitempty"`

	// TLS policy of vault nodes
	TLS *TLSPolicy `json:"TLS,omitempty"`

	// backend storage configuration for vault
	BackendStorage BackendStorageSpec `json:"backendStorage"`

	// unseal configuration for vault
	Unsealer *UnsealerSpec `json:"unsealer,omitempty"`
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

// PodPolicy defines the policy for pods owned by vault operator.
type PodPolicy struct {
	// Resources is the resource requirements for the containers.
	Resources core.ResourceRequirements `json:"resources,omitempty"`
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
	Inmem bool      `json:"inmem,omitempty"`
	Etcd  *EtcdSpec `json:"etcd,omitempty"`
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
	CredentialSecretName string `json:"credentialSecretName,omitempty"`

	// Specifies the secret name that contains tls_ca_file, tls_cert_file and tls_key_file for etcd communication
	TLSSecretName string `json:"tlsSecretName,omitempty"`
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

// ModeSPpec contain unseal mechanism
type ModeSpec struct {
	KubernetesSecret *KubernetesSecretSpec `json:"kubernetesSecret,omitempty"`
}

// KubernetesSecretSpec contain the fields that required to unseal using kubernetes secret
type KubernetesSecretSpec struct {
	SecretName string `json:"secretName"`
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
