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

package storage

import (
	api "kubevault.dev/apimachinery/apis/kubevault/v1alpha1"
	"kubevault.dev/operator/pkg/vault/storage/raft"

	"github.com/pkg/errors"
	core "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
)

// StatefulStorage represents a storage that requires the Vault server to be
// deployed in a StatefulSet
//
// XXX the interface will change!
type StatefulStorage interface {
	Apply(pt *core.PodTemplateSpec) error
	GetStorageConfig() (string, error)
}

// NewStatefulStorage is the factory creating the StatefulStorage based on the Backend given in
// the VaultServer spec.
func NewStatefulStorage(kubeClient kubernetes.Interface, vs *api.VaultServer) (StatefulStorage, error) {
	s := vs.Spec.Backend

	switch {
	case s.Raft != nil:
		return raft.NewOptions(kubeClient, vs, *s.Raft)
	default:
		return nil, errors.New("invalid stateful storage backend")
	}
}
