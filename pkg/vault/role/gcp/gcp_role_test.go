/*
Copyright AppsCode Inc. and Contributors

Licensed under the AppsCode Community License 1.0.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    https://github.com/appscode/licenses/raw/1.0.0/AppsCode-Community-1.0.0.md

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package gcp

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	api "kubevault.dev/apimachinery/apis/engine/v1alpha1"

	"github.com/gorilla/mux"
	vaultapi "github.com/hashicorp/vault/api"
	"github.com/stretchr/testify/assert"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	kfake "k8s.io/client-go/kubernetes/fake"
	"k8s.io/klog/v2"
)

func CreateDemoDB() ([]GCPRole, *httptest.Server) {
	srv := setupVaultServer()

	kubeClient := kfake.NewSimpleClientset(&core.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "gcp-cred",
			Namespace: "demo",
		},
		Data: map[string][]byte{
			api.GCPSACredentialJson: []byte(`{}`),
		},
	})

	cfg := vaultapi.DefaultConfig()
	cfg.Address = srv.URL
	cl, err := vaultapi.NewClient(cfg)
	if err != nil {
		klog.Infoln("Failed to create vault client!")
		return nil, nil
	}

	DB := []GCPRole{
		{
			gcpRole: &api.GCPRole{
				TypeMeta: metav1.TypeMeta{},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "my-role",
					Namespace: "demo",
				},
				Spec: api.GCPRoleSpec{
					VaultRef: core.LocalObjectReference{
						Name: "vault-app",
					},
					SecretType: "access_token",
					Project:    "ackube",
					Bindings: `resource "//cloudresourcemanager.googleapis.com/projects/ackube" {
        roles = ["roles/viewer"]
      }`,
					TokenScopes: []string{"https://www.googleapis.com/auth/cloud-platform"},
				},
			},
			vaultClient: cl,
			kubeClient:  kubeClient,
			gcpPath:     "gcp",
		},
		{
			gcpRole: &api.GCPRole{
				TypeMeta: metav1.TypeMeta{},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "my-role",
					Namespace: "demo",
				},
				Spec: api.GCPRoleSpec{
					VaultRef: core.LocalObjectReference{
						Name: "vault-app",
					},
					SecretType: "access_token",
					Project:    "ackube",
					Bindings: `resource "//cloudresourcemanager.googleapis.com/projects/ackube" {
        roles = ["roles/viewer"]
      }`,
					TokenScopes: []string{"https://www.googleapis.com/auth/cloud-platform"},
				},
			},
			vaultClient: cl,
			kubeClient:  kubeClient,
			gcpPath:     "gcp",
		},
		{
			gcpRole: &api.GCPRole{
				TypeMeta: metav1.TypeMeta{},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "my-role",
					Namespace: "demo",
				},
				Spec: api.GCPRoleSpec{
					VaultRef: core.LocalObjectReference{
						Name: "vault-app",
					},
					SecretType:  "access_token",
					Project:     "ackube",
					TokenScopes: []string{"https://www.googleapis.com/auth/cloud-platform"},
				},
			},
			vaultClient: cl,
			kubeClient:  kubeClient,
			gcpPath:     "gcp",
		},
		{
			gcpRole: &api.GCPRole{
				TypeMeta: metav1.TypeMeta{},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "my-role",
					Namespace: "demo",
				},
				Spec: api.GCPRoleSpec{
					VaultRef: core.LocalObjectReference{
						Name: "vault-app",
					},
					SecretType: "access_token",
					Project:    "ackube",
					Bindings: `resource "//cloudresourcemanager.googleapis.com/projects/ackube" {
        roles = ["roles/viewer"]
      }`,
					TokenScopes: []string{"https://www.googleapis.com/auth/cloud-platform"},
				},
			},
			vaultClient: cl,
			kubeClient:  kubeClient,
			gcpPath:     "",
		},
	}

	return DB, srv
}

func setupVaultServer() *httptest.Server {
	router := mux.NewRouter()

	router.HandleFunc("/v1/gcp/roleset/k8s.-.demo.my-role", func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		var data interface{}
		err := json.NewDecoder(r.Body).Decode(&data)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_, err := w.Write([]byte(err.Error()))
			utilruntime.Must(err)
			return
		} else {
			m := data.(map[string]interface{})
			if v, ok := m["secret_type"]; !ok || len(v.(string)) == 0 {
				w.WriteHeader(http.StatusNotFound)
				_, err := w.Write([]byte("secret_type isn't specified!"))
				utilruntime.Must(err)
				return
			}

			if v, ok := m["project"]; !ok || len(v.(string)) == 0 {
				w.WriteHeader(http.StatusNotFound)
				_, err := w.Write([]byte("project name isn't specified!"))
				utilruntime.Must(err)
				return
			}

			if v, ok := m["bindings"]; !ok || len(v.(string)) == 0 {
				w.WriteHeader(http.StatusNotFound)
				_, err := w.Write([]byte("bindings aren't specified!"))
				utilruntime.Must(err)
				return
			}

			if v, ok := m["token_scopes"]; !ok || v == nil {
				w.WriteHeader(http.StatusNotFound)
				_, err := w.Write([]byte("token_scopes aren't specified!"))
				utilruntime.Must(err)
				return
			}
			w.WriteHeader(http.StatusOK)
		}

	}).Methods(http.MethodPost)

	router.HandleFunc("/v1/gcp/roleset/k8s.-.demo.my-role", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}).Methods(http.MethodDelete)

	return httptest.NewServer(router)
}

func TestGCPRole_CreateRole(t *testing.T) {
	demoRole, srv := CreateDemoDB()
	defer srv.Close()

	testData := []struct {
		testName    string
		gcpRole     *GCPRole
		expectedErr bool
	}{
		{
			testName:    "Create Role Successful",
			gcpRole:     &demoRole[0],
			expectedErr: false,
		},
		{
			testName:    "Create Role Successful",
			gcpRole:     &demoRole[1],
			expectedErr: false,
		},
		{
			testName:    "Create Role Failed",
			gcpRole:     &demoRole[2],
			expectedErr: true,
		},
		{
			testName:    "Create Role Failed",
			gcpRole:     &demoRole[3],
			expectedErr: true,
		},
	}

	for _, test := range testData {
		t.Run(test.testName, func(t *testing.T) {
			m := test.gcpRole
			err := m.CreateRole()
			if test.expectedErr {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func TestGCPRole_DeleteRole(t *testing.T) {
	demoRole, srv := CreateDemoDB()
	defer srv.Close()

	testData := []struct {
		testName    string
		gcpRole     *GCPRole
		expectedErr bool
	}{
		{
			testName:    "Delete Role Successful",
			gcpRole:     &demoRole[0],
			expectedErr: false,
		},
	}

	for _, test := range testData {
		t.Run(test.testName, func(t *testing.T) {
			m := test.gcpRole
			_, err := m.DeleteRole("k8s.-.demo.my-role")
			if test.expectedErr {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}
		})
	}

}
