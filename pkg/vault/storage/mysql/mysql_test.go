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

package mysql

import (
	"fmt"
	"testing"

	api "kubevault.dev/operator/apis/kubevault/v1alpha1"

	"github.com/stretchr/testify/assert"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kfake "k8s.io/client-go/kubernetes/fake"
)

func TestOptions_GetStorageConfig(t *testing.T) {
	kClient := kfake.NewSimpleClientset()
	ns := "test"
	sr := core.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "cred",
			Namespace: ns,
		},
		Data: map[string][]byte{
			"username": []byte("root"),
			"password": []byte("root"),
		},
	}
	_, err := kClient.CoreV1().Secrets(ns).Create(&sr)
	if !assert.Nil(t, err) {
		return
	}

	opts, err := NewOptions(kClient, ns, api.MySQLSpec{
		Address:              "hi.com",
		Database:             "try",
		Table:                "vault",
		TLSCASecret:          "test",
		UserCredentialSecret: "cred",
		MaxParallel:          128,
	})
	if !assert.Nil(t, err) {
		return
	}

	out := `
storage "mysql" {
address = "hi.com"
database = "try"
table = "vault"
tls_ca_file = "/etc/vault/mysql/certs/ca.crt"
username = "root"
password = "root"
max_parallel = "128"
}
`
	t.Run("MySQL storage config", func(t *testing.T) {
		got, err := opts.GetStorageConfig()
		assert.Nil(t, err)
		if !assert.Equal(t, out, got) {
			fmt.Println("expected:", out)
			fmt.Println("got:", got)
		}
	})
}
