/*
Copyright AppsCode Inc. and Contributors

Licensed under the AppsCode Community License 1.0.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    https://github.com/appscode/licenses/raw/1.0.0/AppsCode-Community-1.0.0.md

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package raft

import (
	"fmt"
	"path/filepath"
	"strings"

	api "kubevault.dev/apimachinery/apis/kubevault/v1alpha1"

	"github.com/pkg/errors"
	"gomodules.xyz/pointer"
	core "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
	core_util "kmodules.xyz/client-go/core/v1"
)

// Options represents the instance of the Raft storage.
type Options struct {
	api.RaftSpec
	vs      *api.VaultServer
	kClient kubernetes.Interface
}

// VaultRaftVolumeName is the name of the backend volume created.
const VaultRaftVolumeName = "vault-raft-backend"

// TLS related file name for Raft
const (
	RaftTLSAssetDir    = "/etc/vault/tls/storage/raft"
	RaftClientCaName   = "ca.crt"
	RaftClientCertName = "tls.crt"
	RaftClientKeyName  = "tls.key"
)

var raftStorageFmt = `
storage "raft" {
%s
}
`
var retryJoinFmt = `
retry_join {
%s
}
`

// NewOptions instantiate the Raft storage.
func NewOptions(kubeClient kubernetes.Interface, vs *api.VaultServer) (*Options, error) {
	o := &Options{
		RaftSpec: *vs.Spec.Backend.Raft,
		vs:       vs,
		kClient:  kubeClient,
	}
	return o, nil
}

// Apply ...
func (o *Options) Apply(pt *core.PodTemplateSpec) error {
	raftTLSAssetVolume := "vault-raft-tls"
	if o.Path == "" {
		return errors.New("path is empty")
	}

	// Change the environments variables
	// changed to a generic form, from the hardcoded 0'th index
	for idx := range pt.Spec.Containers {
		if pt.Spec.Containers[idx].Name == string(api.VaultServerServiceVault) {
			pt.Spec.Containers[idx].Env = core_util.UpsertEnvVars(
				pt.Spec.Containers[idx].Env,
				core.EnvVar{
					Name:  "VAULT_API_ADDR",
					Value: "https://$(POD_IP):8200",
				},
				core.EnvVar{
					Name:  "VAULT_CLUSTER_ADDR",
					Value: "https://$(HOSTNAME).vault-internal:8201",
				},
				core.EnvVar{
					Name: "VAULT_RAFT_NODE_ID",
					ValueFrom: &core.EnvVarSource{
						FieldRef: &core.ObjectFieldSelector{
							APIVersion: "v1",
							FieldPath:  "metadata.name",
						},
					},
				},
			)
		}
	}

	for idx := range pt.Spec.Containers {
		if pt.Spec.Containers[idx].Name == string(api.VaultServerServiceVault) {
			pt.Spec.Containers[idx].VolumeMounts = append(pt.Spec.Containers[idx].VolumeMounts, core.VolumeMount{
				Name:      "data",
				MountPath: o.Path,
			})
		}
	}

	if o.vs.Spec.TLS != nil {
		pt.Spec.Volumes = append(pt.Spec.Volumes, core.Volume{
			Name: raftTLSAssetVolume,
			VolumeSource: core.VolumeSource{
				Secret: &core.SecretVolumeSource{
					SecretName: o.vs.GetCertSecretName(string(api.VaultStorageCert)),
				},
			},
		})

		for idx := range pt.Spec.Containers {
			if pt.Spec.Containers[idx].Name == string(api.VaultServerServiceVault) {
				pt.Spec.Containers[idx].VolumeMounts = append(pt.Spec.Containers[idx].VolumeMounts, core.VolumeMount{
					Name:      raftTLSAssetVolume,
					MountPath: RaftTLSAssetDir,
				})
			}
		}
	}

	return nil
}

// GetStorageConfig creates raft storage config from RaftSpec
// https://www.vaultproject.io/docs/configuration/storage/raft
func (o *Options) GetStorageConfig() (string, error) {
	var params []string

	klog.Infoln("Generating storage config for raft backend")

	if o.Path != "" {
		params = append(params, fmt.Sprintf(`path = "%s"`, o.Path))
	}
	if o.NodeID != "" {
		params = append(params, fmt.Sprintf(`node_id = "%s"`, o.NodeID))
	}
	if o.PerformanceMultiplier != 0 {
		params = append(params, fmt.Sprintf(`performance_multiplier = %d`, o.PerformanceMultiplier))
	}
	if o.TrailingLogs != nil {
		params = append(params, fmt.Sprintf(`trailing_logs = %d`, o.TrailingLogs))
	}
	if o.SnapshotThreshold != nil {
		params = append(params, fmt.Sprintf(`snapshot_threshold = %d`, o.SnapshotThreshold))
	}

	// Generate & Insert the retryJoin stanza(s)
	replicas := pointer.Int32(o.vs.Spec.Replicas)
	if replicas == 0 {
		replicas = 1
	}
	for id := 0; id < int(replicas); id++ {
		var retryJoin []string
		// e.g: retryJoin = append(retryJoin, fmt.Sprint(` leader_api_addr = "https://vault-0.vault-internal.demo.svc:8200"`))
		retryJoin = append(retryJoin, fmt.Sprintf(` leader_api_addr = "%s://%s-%d.%s.%s.svc:8200"`, o.vs.Scheme(), o.vs.Name, id, o.vs.ServiceName(api.VaultServerServiceInternal), o.vs.Namespace))
		if o.vs.Spec.TLS != nil {
			retryJoin = append(retryJoin, fmt.Sprintf(` leader_ca_cert_file = "%s"`, filepath.Join(RaftTLSAssetDir, RaftClientCaName)))
			retryJoin = append(retryJoin, fmt.Sprintf(` leader_client_cert_file = "%s"`, filepath.Join(RaftTLSAssetDir, RaftClientCertName)))
			retryJoin = append(retryJoin, fmt.Sprintf(` leader_client_key_file = "%s"`, filepath.Join(RaftTLSAssetDir, RaftClientKeyName)))
		}
		params = append(params, fmt.Sprintf(retryJoinFmt, strings.Join(retryJoin, "\n")))
	}

	if o.MaxEntrySize != nil {
		params = append(params, fmt.Sprintf(`max_entry_size = %d`, o.MaxEntrySize))
	}
	if o.AutopilotReconcileInterval != "" {
		params = append(params, fmt.Sprintf(`autopilot_reconcile_interval = "%s"`, o.AutopilotReconcileInterval))
	}

	storageCfg := fmt.Sprintf(raftStorageFmt, strings.Join(params, "\n"))

	return storageCfg, nil
}
