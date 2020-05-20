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
	"fmt"
	"strings"

	api "kubevault.dev/operator/apis/kubevault/v1alpha1"

	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
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
}

func NewOptions(kc kubernetes.Interface, namespace string, s api.RaftSpec) (*Options, error) {
	return &Options{
		kc,
		namespace,
		s,
	}, nil
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
		configMap, err := o.kc.CoreV1().ConfigMaps(o.namespace).Get(o.RetryJoinConfig, metav1.GetOptions{})
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
