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

package engine

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	api "kubevault.dev/operator/apis/engine/v1alpha1"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const gcpPolicyTest1 = `
path "gcp/config" {
	capabilities = ["create", "update", "read", "delete"]
}

path "gcp/roleset/*" {
	capabilities = ["create", "update", "read", "delete"]
}

path "gcp/token/*" {
	capabilities = ["create", "update", "read"]
}

path "gcp/key/*" {
	capabilities = ["create", "update", "read"]
}
`
const gcpPolicyTest2 = `
path "my-gcp-path/config" {
	capabilities = ["create", "update", "read", "delete"]
}

path "my-gcp-path/roleset/*" {
	capabilities = ["create", "update", "read", "delete"]
}

path "my-gcp-path/token/*" {
	capabilities = ["create", "update", "read"]
}

path "my-gcp-path/key/*" {
	capabilities = ["create", "update", "read"]
}
`

func NewFakeVaultPolicyServer() *httptest.Server {
	router := mux.NewRouter()

	router.HandleFunc("/v1/sys/policies/acl/{path}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		path := vars["path"]
		body := r.Body
		data, _ := ioutil.ReadAll(body)
		var newdata map[string]interface{}
		_ = json.Unmarshal(data, &newdata)

		if path == "k8s.-.demo.gcpse" {
			if newdata["policy"] == gcpPolicyTest1 || newdata["policy"] == gcpPolicyTest2 {
				w.WriteHeader(http.StatusOK)
				return
			}
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusBadRequest)
	}).Methods(http.MethodPut)

	return httptest.NewServer(router)
}

func TestSecretEngine_CreatePolicy(t *testing.T) {

	srv := NewFakeVaultPolicyServer()
	defer srv.Close()

	tests := []struct {
		name         string
		secretEngine *api.SecretEngine
		path         string
		wantErr      bool
	}{
		{
			name: "Create policy for gcp secret engine",
			path: "gcp",
			secretEngine: &api.SecretEngine{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "gcpse",
					Namespace: "demo",
				},
				Spec: api.SecretEngineSpec{
					VaultRef: v1.LocalObjectReference{},
					Path:     "",
					SecretEngineConfiguration: api.SecretEngineConfiguration{
						GCP: &api.GCPConfiguration{},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "Create policy for my-gcp-path secret engine",
			path: "my-gcp-path",
			secretEngine: &api.SecretEngine{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "gcpse",
					Namespace: "demo",
				},
				Spec: api.SecretEngineSpec{
					VaultRef: v1.LocalObjectReference{},
					Path:     "",
					SecretEngineConfiguration: api.SecretEngineConfiguration{
						GCP: &api.GCPConfiguration{},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "Create policy for gcp secret engine failed",
			path: "my-gcp-path",
			secretEngine: &api.SecretEngine{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "gcpse",
					Namespace: "demo",
				},
				Spec: api.SecretEngineSpec{
					VaultRef: v1.LocalObjectReference{},
					Path:     "",
					SecretEngineConfiguration: api.SecretEngineConfiguration{
						AWS: &api.AWSConfiguration{},
					},
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			vc, err := vaultClient(srv.URL)
			assert.Nil(t, err, "failed to create vault client")

			seClient := &SecretEngine{
				secretEngine: tt.secretEngine,
				path:         tt.path,
				vaultClient:  vc,
			}
			if err := seClient.CreatePolicy(); (err != nil) != tt.wantErr {
				t.Errorf("CreatePolicy() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
