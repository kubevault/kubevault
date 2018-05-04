package util

import (
	"github.com/appscode/kutil/meta"
	"github.com/pkg/errors"
	api "github.com/soter/vault-operator/apis/vault/v1alpha1"
	"k8s.io/apimachinery/pkg/conversion"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func GetGroupVersionKind(v interface{}) schema.GroupVersionKind {
	return api.SchemeGroupVersion.WithKind(meta.GetKind(v))
}

func AssignTypeKind(v interface{}) error {
	_, err := conversion.EnforcePtr(v)
	if err != nil {
		return err
	}

	switch u := v.(type) {
	case *api.VaultServer:
		u.APIVersion = api.SchemeGroupVersion.String()
		u.Kind = api.ResourceKindVaultServer
		return nil
	}
	return errors.New("unknown api object type")
}
