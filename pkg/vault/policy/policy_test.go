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

package policy

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	vaultapi "github.com/hashicorp/vault/api"
	"github.com/stretchr/testify/assert"
)

func NewFakeVaultServer() *httptest.Server {
	router := mux.NewRouter()

	router.HandleFunc("/v1/sys/policies/acl/ok", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}).Methods(http.MethodPut)

	router.HandleFunc("/v1/sys/policies/acl/err", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}).Methods(http.MethodPut)

	router.HandleFunc("/v1/sys/policies/acl/ok", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}).Methods(http.MethodDelete)

	router.HandleFunc("/v1/sys/policies/acl/err", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}).Methods(http.MethodDelete)

	return httptest.NewServer(router)
}

func TestVPolicy_EnsurePolicy(t *testing.T) {
	plcy := `
	 path "secret/*" {
	   capabilities = ["create", "read", "update", "delete", "list"]
	 }
`
	srv := NewFakeVaultServer()
	defer srv.Close()

	cases := []struct {
		testName  string
		name      string
		policy    string
		expectErr bool
	}{
		{
			testName:  "put policy success",
			name:      "ok",
			policy:    plcy,
			expectErr: false,
		},
		{
			testName:  "put policy failed",
			name:      "err",
			policy:    plcy,
			expectErr: true,
		},
	}

	for _, c := range cases {
		t.Run(c.testName, func(t *testing.T) {
			vc, err := vaultClient(srv.URL, "root")
			if assert.Nil(t, err, "failed to create vault client") {
				vp := &vPolicy{
					client: vc,
				}

				err = vp.EnsurePolicy(c.name, c.policy)
				if !c.expectErr {
					assert.Nil(t, err, "failed to put policy")
				} else {
					assert.NotNil(t, err, "expected error")
				}
			}
		})
	}
}

func TestVPolicy_DeletePolicy(t *testing.T) {
	srv := NewFakeVaultServer()
	defer srv.Close()

	cases := []struct {
		testName  string
		name      string
		expectErr bool
	}{
		{
			testName:  "delete policy success",
			name:      "ok",
			expectErr: false,
		},
		{
			testName:  "delete policy failed",
			name:      "err",
			expectErr: true,
		},
	}

	for _, c := range cases {
		t.Run(c.testName, func(t *testing.T) {
			vc, err := vaultClient(srv.URL, "root")
			if assert.Nil(t, err, "failed to create vault client") {
				vp := &vPolicy{
					client: vc,
				}

				err = vp.DeletePolicy(c.name)
				if !c.expectErr {
					assert.Nil(t, err, "failed to delete policy")
				} else {
					assert.NotNil(t, err, "expected error")
				}
			}
		})
	}
}

func vaultClient(addr, token string) (*vaultapi.Client, error) {
	cfg := vaultapi.DefaultConfig()
	err := cfg.ConfigureTLS(&vaultapi.TLSConfig{
		Insecure: true,
	})
	if err != nil {
		return nil, err
	}
	c, err := vaultapi.NewClient(cfg)
	if err != nil {
		return nil, err
	}

	c.SetToken(token)
	err = c.SetAddress(addr)
	if err != nil {
		return nil, err
	}
	return c, nil
}
