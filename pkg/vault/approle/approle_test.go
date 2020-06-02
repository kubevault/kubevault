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

package approle

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/gorilla/mux"

	vaultapi "github.com/hashicorp/vault/api"
	approleapi "kubevault.dev/operator/apis/approle/v1alpha1"
)

func NewFakeVaultServer() *httptest.Server {
	router := mux.NewRouter()

	router.HandleFunc("/v1/auth/approle/role/ok", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}).Methods(http.MethodDelete)

	router.HandleFunc("/v1/auth/approle/role/err", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}).Methods(http.MethodDelete)

	router.HandleFunc("/v1/auth/approle/role/ok", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}).Methods(http.MethodPut)

	router.HandleFunc("/v1/auth/approle/role/err", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}).Methods(http.MethodPut)

	return httptest.NewServer(router)
}

func TestVAppRole_EnsureAppRole(t *testing.T) {

	fakeAppRole := approleapi.VaultAppRole{}
	fakeAppRole.Spec = approleapi.VaultAppRoleSpec{RoleName: "test-approle"}
	payload, err := fakeAppRole.GeneratePayLoad()

	assert.Nil(t, err, "failed to generate payload")

	srv := NewFakeVaultServer()
	defer srv.Close()

	cases := []struct {
		testName  string
		name      string
		appRole   map[string]interface{}
		expectErr bool
	}{
		{
			testName:  "put approle success",
			name:      "ok",
			appRole:   payload,
			expectErr: false,
		},
		{
			testName:  "put approle failed",
			name:      "err",
			appRole:   payload,
			expectErr: true,
		},
	}

	for _, c := range cases {
		t.Run(c.testName, func(t *testing.T) {
			vc, err := vaultClient(srv.URL, "root")
			if assert.Nil(t, err, "failed to create vault client") {
				vp := &vAppRole{
					client: vc,
				}

				err = vp.EnsureAppRole(c.name, c.appRole)
				if !c.expectErr {
					assert.Nil(t, err, "failed to put approle")
				} else {
					assert.NotNil(t, err, "expected error")
				}
			}
		})
	}
}

func TestVAppRole_DeleteAppRole(t *testing.T) {
	srv := NewFakeVaultServer()
	defer srv.Close()

	cases := []struct {
		testName  string
		name      string
		expectErr bool
	}{
		{
			testName:  "delete approle success",
			name:      "ok",
			expectErr: false,
		},
		{
			testName:  "delete approle failed",
			name:      "err",
			expectErr: true,
		},
	}

	for _, c := range cases {
		t.Run(c.testName, func(t *testing.T) {
			vc, err := vaultClient(srv.URL, "root")
			if assert.Nil(t, err, "failed to create vault client") {
				vp := &vAppRole{
					client: vc,
				}

				err = vp.DeleteAppRole(c.name)
				if !c.expectErr {
					assert.Nil(t, err, "failed to delete approle")
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
