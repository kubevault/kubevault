package util

import (
	"testing"

	api "github.com/kubevault/operator/apis/core/v1alpha1"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestEnsureOwnerRefToObject(t *testing.T) {
	owner := &api.VaultServer{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "vs",
			Namespace: "vs",
			UID:       "1234",
		},
	}

	sPointer := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "hi",
			Namespace: "hi",
			UID:       "1234",
		},
	}
	s := corev1.Secret{
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
