package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	api "kubevault.dev/operator/apis/kubevault/v1alpha1"
)

func TestEnsureOwnerRefToObject(t *testing.T) {
	owner := &api.VaultServer{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "vs",
			Namespace: "vs",
			UID:       "1234",
		},
	}

	sPointer := &core.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "hi",
			Namespace: "hi",
			UID:       "1234",
		},
	}
	s := core.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "hi",
			Namespace: "hi",
			UID:       "1234",
		},
	}

	EnsureOwnerRefToObject(sPointer, AsOwner(owner))
	assert.Condition(t, func() (success bool) {
		return IsOwnerRefAlreadyExists(sPointer, AsOwner(owner))
	})

	EnsureOwnerRefToObject(s.GetObjectMeta(), AsOwner(owner))
	assert.Condition(t, func() (success bool) {
		return IsOwnerRefAlreadyExists(s.GetObjectMeta(), AsOwner(owner))
	})
}
