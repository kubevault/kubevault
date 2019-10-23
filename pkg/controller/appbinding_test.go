package controller

import (
	"testing"

	api "kubevault.dev/operator/apis/kubevault/v1alpha1"

	"github.com/stretchr/testify/assert"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kfake "k8s.io/client-go/kubernetes/fake"
	appcat_csfake "kmodules.xyz/custom-resources/client/clientset/versioned/fake"
)

func TestRun(t *testing.T) {
	ctrl := &VaultController{
		kubeClient:       kfake.NewSimpleClientset(),
		appCatalogClient: appcat_csfake.NewSimpleClientset().AppcatalogV1alpha1(),
	}

	err := ctrl.ensureAppBindings(&api.VaultServer{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "default",
			UID:       "123",
		},
	}, &vaultFake{
		sr: &core.Secret{
			Data: map[string][]byte{
				"ca.crt": []byte("ca"),
			},
		},
		svc: &core.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test",
				Namespace: "default",
			},
		},
	})

	assert.Nil(t, err)
}
