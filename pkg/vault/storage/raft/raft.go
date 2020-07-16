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

package raft

import (
	"context"
	"fmt"
	"strings"

	api "kubevault.dev/operator/apis/kubevault/v1alpha1"

	"github.com/pkg/errors"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	core_util "kmodules.xyz/client-go/core/v1"
)

const (
	VaultRaftVolumeName = "vault-raft-backend"
)

var raftStorageFmt = `
storage "raft" {
%s
}
`

type Options struct {
	kc        kubernetes.Interface
	namespace string
	api.RaftSpec
	claimName string
}

func NewOptions(kc kubernetes.Interface, vaultServer *api.VaultServer, s api.RaftSpec) (*Options, error) {
	if vaultServer == nil {
		return nil, errors.New("vaultServer object is empty")
	}

	// Set the pvc name and labels if given.
	// Otherwise default to VaultServer's name, namespace and labels.
	objMeta := metav1.ObjectMeta{
		Name:      vaultServer.Name,
		Namespace: vaultServer.Namespace,
		Labels:    vaultServer.OffshootLabels(),
	}

	if s.VolumeClaimTemplate.Name != "" {
		objMeta.Name = s.VolumeClaimTemplate.Name
	}

	if s.VolumeClaimTemplate.Labels != nil {
		objMeta.Labels = s.VolumeClaimTemplate.Labels
	}

	// Create or Patch the requested PVC
	_, _, err := core_util.CreateOrPatchPVC(context.TODO(), kc, objMeta, func(claim *core.PersistentVolumeClaim) *core.PersistentVolumeClaim {
		// pvc.spec is immutable except spec.resources.request field.
		// But values need to be set while creating the pvc for the first time.
		// Here, "Spec.AccessModes" will be "nil" in two cases; invalid pvc template
		// & creating pvc for the first time.
		if claim.Spec.AccessModes == nil {
			claim.Spec = s.VolumeClaimTemplate.Spec
		}

		// Update labels
		claim.Labels = objMeta.Labels

		// Update the only mutable field.
		claim.Spec.Resources.Requests = s.VolumeClaimTemplate.Spec.Resources.Requests
		return claim
	}, metav1.PatchOptions{})
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create pvc %s/%s", objMeta.Namespace, objMeta.Name)
	}

	return &Options{
		kc,
		vaultServer.Namespace,
		s,
		objMeta.Name,
	}, nil
}

func (o *Options) Apply(pt *core.PodTemplateSpec) error {
	if o.Path == "" || o.claimName == "" {
		return errors.New("path or pvc name is empty")
	}

	pt.Spec.Volumes = append(pt.Spec.Volumes, core.Volume{
		Name: VaultRaftVolumeName,
		VolumeSource: core.VolumeSource{
			PersistentVolumeClaim: &core.PersistentVolumeClaimVolumeSource{
				ClaimName: o.claimName,
			},
		},
	})

	pt.Spec.Containers[0].VolumeMounts = append(pt.Spec.Containers[0].VolumeMounts, core.VolumeMount{
		Name:      VaultRaftVolumeName,
		MountPath: o.Path,
	})

	return nil
}

// vault doc: https://www.vaultproject.io/docs/configuration/storage/raft.html
//
// Note:
// - Secret `TLSSecretName` mounted in `ConsulTLSAssetDir`
// - Secret `ACLTokenSecretName` will be used to aclToken from secret
//
// GetStorageConfig creates raft storage config from RaftSpec
func (o *Options) GetStorageConfig() (string, error) {
	params := []string{}
	if o.Path != "" {
		params = append(params, fmt.Sprintf(`path = "%s"`, o.Path))
	}
	if o.NodeID != "" {
		params = append(params, fmt.Sprintf(`node_id = "%s"`, o.NodeID))
	}
	if o.PerformanceMultiplier != 0 {
		params = append(params, fmt.Sprintf(`performance_multiplier = "%d"`, o.PerformanceMultiplier))
	}
	if o.TrailingLogs != 10000 {
		params = append(params, fmt.Sprintf(`trailing_logs = "%d"`, o.TrailingLogs))
	}
	if o.SnapshotThreshold != 8192 {
		params = append(params, fmt.Sprintf(`snapshot_threshold = "%d"`, o.SnapshotThreshold))
	}
	// Get RetryJoin stanza from configMap
	if o.RetryJoinConfig != "" {
		configMap, err := o.kc.CoreV1().ConfigMaps(o.namespace).Get(context.TODO(), o.RetryJoinConfig, metav1.GetOptions{})
		if err != nil {
			return "", errors.Wrapf(err, "failed to get configMap %s/%s", o.namespace, o.RetryJoinConfig)
		}

		if value, exist := configMap.Data["retry_join.hcl"]; !exist {
			return "", errors.Wrapf(err, "Data field is empty in %s/%s", o.namespace, o.RetryJoinConfig)
		} else {
			params = append(params, fmt.Sprintf(`token = "%s"`, string(value)))
		}

	}

	storageCfg := fmt.Sprintf(raftStorageFmt, strings.Join(params, "\n"))
	return storageCfg, nil
}
