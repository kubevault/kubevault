package storage

import (
	api "github.com/kubevault/operator/apis/core/v1alpha1"
	"github.com/kubevault/operator/pkg/vault/storage/azure"
	"github.com/kubevault/operator/pkg/vault/storage/etcd"
	"github.com/kubevault/operator/pkg/vault/storage/file"
	"github.com/kubevault/operator/pkg/vault/storage/gcs"
	"github.com/kubevault/operator/pkg/vault/storage/inmem"
	"github.com/kubevault/operator/pkg/vault/storage/mysql"
	"github.com/kubevault/operator/pkg/vault/storage/postgersql"
	"github.com/kubevault/operator/pkg/vault/storage/s3"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
)

type Storage interface {
	Apply(pt *corev1.PodTemplateSpec) error
	GetStorageConfig() (string, error)
}

func NewStorage(kubeClient kubernetes.Interface, vs *api.VaultServer) (Storage, error) {
	s := vs.Spec.BackendStorage

	if s.Inmem {
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
		return postgresql.NewOptions(kubeClient, vs.GetNamespace(), *s.PostgreSQL)
	} else if s.MySQL != nil {
		return mysql.NewOptions(kubeClient, vs.GetNamespace(), *s.MySQL)
	} else if s.File != nil {
		return file.NewOptions(*s.File)
	} else {
		return nil, errors.New("invalid storage backend")
	}
}
