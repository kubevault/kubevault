package inmem

import (
	api "github.com/kube-vault/operator/apis/core/v1alpha1"
	corev1 "k8s.io/api/core/v1"
)

var inmenStorage = `
storage "inmem" {
}
`

type Options struct {
	api.InmemSpec
}

func NewOptions(s api.InmemSpec) (*Options,error) {
	return &Options{
		s,
	}, nil
}

func (o *Options) Apply(pt *corev1.PodTemplateSpec) error {
	return nil
}

// GetStorageConfig will create storage config for inmem backend
func (o *Options) GetStorageConfig() (string, error) {
		return inmenStorage,nil
}
