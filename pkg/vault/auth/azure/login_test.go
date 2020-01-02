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

package azure

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/hashicorp/vault/api"
	"github.com/stretchr/testify/assert"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
)

const authResp = `
{
  "auth": {
    "client_token": "1234"
  }
}
`

func NewFakeVaultServer() *httptest.Server {
	router := mux.NewRouter()
	router.HandleFunc("/v1/auth/azure/login", func(w http.ResponseWriter, r *http.Request) {
		var v map[string]interface{}
		defer r.Body.Close()
		utilruntime.Must(json.NewDecoder(r.Body).Decode(&v))
		value, ok := v["role"]
		if !ok || value == "" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		value, ok = v["jwt"]
		if !ok || value == "" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte(authResp))
		utilruntime.Must(err)
	}).Methods(http.MethodPost)

	router.HandleFunc("/v1/auth/my-path/login", func(w http.ResponseWriter, r *http.Request) {
		var v map[string]interface{}
		defer r.Body.Close()
		utilruntime.Must(json.NewDecoder(r.Body).Decode(&v))

		value, ok := v["role"]
		if !ok || value == "" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		value, ok = v["jwt"]
		if !ok || value == "" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte(authResp))
		utilruntime.Must(err)
	}).Methods(http.MethodPost)

	return httptest.NewServer(router)
}

func TestAuth_Login(t *testing.T) {
	srv := NewFakeVaultServer()
	defer srv.Close()

	vc, err := api.NewClient(api.DefaultConfig())
	if !assert.Nil(t, err) {
		return
	}

	utilruntime.Must(vc.SetAddress(srv.URL))

	cases := []struct {
		testName    string
		au          *auth
		expectedErr bool
	}{
		{
			testName: "Successful login",
			au: &auth{
				vClient:           vc,
				role:              "role-one",
				path:              "azure",
				signedJWT:         "kf09eiurekijlkjflkdj9f.flkjdlf.fjdflkjd",
				subscriptionID:    "565-5443-5-4545-43",
				resourceGroupName: "vault-test",
				vmName:            "vault",
			},
			expectedErr: false,
		},
		{
			testName: "Empty role, unsuccessful login",
			au: &auth{
				vClient:           vc,
				path:              "azure",
				signedJWT:         "dfdfd.sfdsf.dfdsdfds",
				subscriptionID:    "342-0324-03234",
				resourceGroupName: "vault-test",
				vmName:            "test",
			},
			expectedErr: true,
		},
		{
			testName: "Empty jwt, unsuccessful login",
			au: &auth{
				vClient:        vc,
				role:           "role3",
				path:           "azure",
				subscriptionID: "342-234-324-3",
			},
			expectedErr: true,
		},
		{
			testName: "Successful login at user defined path",
			au: &auth{
				vClient:           vc,
				role:              "role",
				path:              "my-path",
				signedJWT:         "fdsf.fdfkjdsfro9r.dfdfd",
				subscriptionID:    "23432-34-324-2343",
				resourceGroupName: "test",
				vmName:            "vault",
			},
			expectedErr: false,
		},
	}

	for _, c := range cases {
		t.Run(c.testName, func(t *testing.T) {
			token, err := c.au.Login()
			if c.expectedErr {
				assert.NotNil(t, err)
			} else {
				if assert.Nil(t, err) {
					assert.Condition(t, func() (success bool) {
						return token == "1234"
					})
				}
			}
		})
	}
}
