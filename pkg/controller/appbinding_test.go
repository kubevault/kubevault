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
