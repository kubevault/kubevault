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

package storage

import (
	api "kubevault.dev/operator/apis/kubevault/v1alpha1"
	"kubevault.dev/operator/pkg/vault/storage/azure"
	"kubevault.dev/operator/pkg/vault/storage/consul"
	"kubevault.dev/operator/pkg/vault/storage/dynamodb"
	"kubevault.dev/operator/pkg/vault/storage/etcd"
	"kubevault.dev/operator/pkg/vault/storage/file"
	"kubevault.dev/operator/pkg/vault/storage/gcs"
	"kubevault.dev/operator/pkg/vault/storage/inmem"
	"kubevault.dev/operator/pkg/vault/storage/mysql"
	postgresql "kubevault.dev/operator/pkg/vault/storage/postgersql"
	"kubevault.dev/operator/pkg/vault/storage/raft"
	"kubevault.dev/operator/pkg/vault/storage/s3"
	"kubevault.dev/operator/pkg/vault/storage/swift"

	"github.com/pkg/errors"
	core "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
)

type Storage interface {
	Apply(pt *core.PodTemplateSpec) error
	GetStorageConfig() (string, error)
}

func NewStorage(kubeClient kubernetes.Interface, vs *api.VaultServer) (Storage, error) {
	s := vs.Spec.Backend

	if s.Inmem != nil {
		return inmem.NewOptions()
	} else if s.Etcd != nil {
		return etcd.NewOptions(*s.Etcd)
	} else if s.Gcs != nil {
		return gcs.NewOptions(*s.Gcs)
	} else if s.S3 != nil {
		return s3.NewOptions(*s.S3)
	} else if s.Azure != nil {
		return azure.NewOptions(*s.Azure)
	} else if s.PostgreSQL != nil {
		return postgresql.NewOptions(kubeClient, vs.Namespace, *s.PostgreSQL)
	} else if s.MySQL != nil {
		return mysql.NewOptions(kubeClient, vs.Namespace, *s.MySQL)
	} else if s.File != nil {
		return file.NewOptions(kubeClient, vs, s.File)
	} else if s.DynamoDB != nil {
		return dynamodb.NewOptions(*s.DynamoDB)
	} else if s.Swift != nil {
		return swift.NewOptions(*s.Swift)
	} else if s.Consul != nil {
		return consul.NewOptions(kubeClient, vs.Namespace, *s.Consul)
	} else if s.Raft != nil {
		return raft.NewOptions(kubeClient, vs.Namespace, *s.Raft)
	} else {
		return nil, errors.New("invalid storage backend")
	}
}
