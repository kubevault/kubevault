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
	ofst "kmodules.xyz/offshoot-api/api/v1"
)

const (
	VaultServerContainerName   = "vault"
	VaultUnsealerContainerName = "unsealer"
	VaultExporterContainerName = "exporter"
)

// +kubebuilder:validation:Enum=Initializing;Unsealing;Sealed;Ready;NotReady;Critical
type VaultServerPhase string

const (
	// used for VaultServer that are Initializing
	VaultServerPhaseInitializing VaultServerPhase = "Initializing"
	// used for VaultServer that are Unsealing
	VaultServerPhaseUnsealing VaultServerPhase = "Unsealing"
	// used for VaultServer that are sealed
	VaultServerPhaseSealed VaultServerPhase = "Sealed"
	// used for VaultServer that are Ready
	VaultServerPhaseReady VaultServerPhase = "Ready"
	// used for VaultServer that are NotReady
	VaultServerPhaseNotReady VaultServerPhase = "NotReady"
	// used for VaultServer that are Critical
	VaultServerPhaseCritical VaultServerPhase = "Critical"
)

// +kubebuilder:validation:Enum=Halt;Delete;WipeOut;DoNotTerminate
type TerminationPolicy string

const (
	// Deletes VaultServer pods, service but leave the PVCs and stash backup data intact.
	TerminationPolicyHalt TerminationPolicy = "Halt"
	// Deletes VaultServer pods, service, pvcs but leave the stash backup data intact.
	TerminationPolicyDelete TerminationPolicy = "Delete"
	// Deletes VaultServer pods, service, pvcs and stash backup data.
	TerminationPolicyWipeOut TerminationPolicy = "WipeOut"
	// Rejects attempt to delete VaultServer using ValidationWebhook.
	TerminationPolicyDoNotTerminate TerminationPolicy = "DoNotTerminate"
)

// +kubebuilder:validation:Enum=internal;vault;stats
type ServiceAlias string

const (
	VaultServerServiceInternal ServiceAlias = "internal"
	VaultServerServiceVault    ServiceAlias = "vault"
	VaultServerServiceStats    ServiceAlias = "stats"
)

type NamedServiceTemplateSpec struct {
	// Alias represents the identifier of the service.
	Alias ServiceAlias `json:"alias" protobuf:"bytes,1,opt,name=alias"`

	// ServiceTemplate is an optional configuration for a service used to expose VaultServer
	// +optional
	ofst.ServiceTemplateSpec `json:",inline,omitempty" protobuf:"bytes,2,opt,name=serviceTemplateSpec"`
}

// +kubebuilder:validation:Enum=ca;server;client;storage
type VaultCertificateAlias string

const (
	VaultCACert      VaultCertificateAlias = "ca"
	VaultServerCert  VaultCertificateAlias = "server"
	VaultClientCert  VaultCertificateAlias = "client"
	VaultStorageCert VaultCertificateAlias = "storage"
)

// +kubebuilder:validation:Enum=inmem;etcd;gcs;s3;azure;postgresql;mysql;file;dynamodb;swift;consul;raft
type VaultServerBackend string

const (
	VaultServerInmem      VaultServerBackend = "inmem"
	VaultServerEtcd       VaultServerBackend = "etcd"
	VaultServerGcs        VaultServerBackend = "gcs"
	VaultServerS3         VaultServerBackend = "s3"
	VaultServerAzure      VaultServerBackend = "azure"
	VaultServerPostgreSQL VaultServerBackend = "postgresql"
	VaultServerMySQL      VaultServerBackend = "mysql"
	VaultServerFile       VaultServerBackend = "file"
	VaultServerDynamoDB   VaultServerBackend = "dynamodb"
	VaultServerSwift      VaultServerBackend = "swift"
	VaultServerConsul     VaultServerBackend = "consul"
	VaultServerRaft       VaultServerBackend = "raft"
)
