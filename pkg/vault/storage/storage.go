package storage

import (
	api "github.com/kube-vault/operator/apis/core/v1alpha1"
	"github.com/kube-vault/operator/pkg/vault/storage/etcd"
	"github.com/kube-vault/operator/pkg/vault/storage/inmem"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
)

type Storage interface {
	Apply(pt *corev1.PodTemplateSpec) error
	GetStorageConfig() (string, error)
}

func NewStorage(s *api.BackendStorageSpec) (Storage, error) {
	if s.Inmem {
		return inmem.NewOptions()
	} else if s.Etcd != nil {
		return etcd.NewOptions(*s.Etcd)
	} else {
		return nil, errors.New("invalid storage backend")
	}
}
