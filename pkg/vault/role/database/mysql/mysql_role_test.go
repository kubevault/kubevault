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
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	api "kubevault.dev/operator/apis/engine/v1alpha1"

	"github.com/gorilla/mux"
	vaultapi "github.com/hashicorp/vault/api"
	"github.com/stretchr/testify/assert"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	appcat "kmodules.xyz/custom-resources/apis/appcatalog/v1alpha1"
)

func setupVaultServer() *httptest.Server {
	router := mux.NewRouter()

	router.HandleFunc("/v1/database/roles/k8s.-.m.m-read", func(w http.ResponseWriter, r *http.Request) {
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
			if v, ok := m["db_name"]; !ok || len(v.(string)) == 0 {
				w.WriteHeader(http.StatusBadRequest)
				_, err := w.Write([]byte("db_name doesn't provided"))
				utilruntime.Must(err)
				return
			}
			w.WriteHeader(http.StatusOK)
		}
	}).Methods(http.MethodPost)

	router.HandleFunc("/v1/database/roles/k8s.-.m.m-read", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}).Methods(http.MethodDelete)

	router.HandleFunc("/v1/database/roles/error", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, err := w.Write([]byte("error"))
		utilruntime.Must(err)
	}).Methods(http.MethodDelete)

	return httptest.NewServer(router)
}

func TestMySQLRole_CreateRole(t *testing.T) {
	srv := setupVaultServer()
	defer srv.Close()

	cfg := vaultapi.DefaultConfig()
	cfg.Address = srv.URL

	cl, err := vaultapi.NewClient(cfg)
	if !assert.Nil(t, err, "failed to create vault client") {
		return
	}

	testData := []struct {
		testName    string
		mClient     *MySQLRole
		expectedErr bool
	}{
		{
			testName: "Create Role successful",
			mClient: &MySQLRole{
				mRole: &api.MySQLRole{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "m-read",
						Namespace: "m",
					},
					Spec: api.MySQLRoleSpec{
						CreationStatements: []string{"create table"},
						VaultRef: core.LocalObjectReference{
							Name: "vault-app",
						},
						DatabaseRef: &appcat.AppReference{
							Namespace: "demo",
							Name:      "mysql",
						},
					},
				},
				vaultClient:  cl,
				databasePath: "database",
			},
			expectedErr: false,
		},
		{
			testName: "Create Role failed",
			mClient: &MySQLRole{
				mRole: &api.MySQLRole{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "m-read",
						Namespace: "m",
					},
					Spec: api.MySQLRoleSpec{
						CreationStatements: []string{"create table"},
						VaultRef: core.LocalObjectReference{
							Name: "vault-app",
						},
					},
				},
				vaultClient:  cl,
				databasePath: "database",
			},
			expectedErr: true,
		},
	}

	for _, test := range testData {
		t.Run(test.testName, func(t *testing.T) {
			m := test.mClient

			err := m.CreateRole()
			if test.expectedErr {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}
