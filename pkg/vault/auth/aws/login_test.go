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

package aws

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	authtype "kubevault.dev/operator/pkg/vault/auth/types"

	"github.com/gorilla/mux"
	vaultapi "github.com/hashicorp/vault/api"
	"github.com/stretchr/testify/assert"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	appcat "kmodules.xyz/custom-resources/apis/appcatalog/v1alpha1"
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
	router.HandleFunc("/v1/auth/aws/login", func(w http.ResponseWriter, r *http.Request) {
		var v map[string]interface{}
		defer r.Body.Close()
		utilruntime.Must(json.NewDecoder(r.Body).Decode(&v))
		if val, ok := v["role"]; ok {
			if val.(string) == "good" {
				w.WriteHeader(http.StatusOK)
				_, err := w.Write([]byte(authResp))
				utilruntime.Must(err)
				return
			}
		}
		w.WriteHeader(http.StatusBadRequest)
	}).Methods(http.MethodPost)

	router.HandleFunc("/v1/auth/test/login", func(w http.ResponseWriter, r *http.Request) {
		var v map[string]interface{}
		defer r.Body.Close()
		utilruntime.Must(json.NewDecoder(r.Body).Decode(&v))
		if val, ok := v["role"]; ok {
			if val.(string) == "try" {
				w.WriteHeader(http.StatusOK)
				_, err := w.Write([]byte(authResp))
				utilruntime.Must(err)
				return
			}
		}
		w.WriteHeader(http.StatusBadRequest)
	})

	return httptest.NewServer(router)
}

func TestAuth_Login(t *testing.T) {
	accessKey := os.Getenv("AWS_ACCESS_KEY_ID")
	secretKey := os.Getenv("AWS_SECRET_ACCESS_KEY")
	if accessKey == "" || secretKey == "" {
		t.Skip()
	}

	srv := NewFakeVaultServer()
	defer srv.Close()

	vc, err := vaultapi.NewClient(vaultapi.DefaultConfig())
	if !assert.Nil(t, err) {
		t.Skip()
	}
	utilruntime.Must(vc.SetAddress(srv.URL))

	awsCred, err := retrieveCreds(accessKey, secretKey, "")
	if !assert.Nil(t, err) {
		t.Skip()
	}

	cases := []struct {
		testName  string
		au        *auth
		expectErr bool
	}{
		{
			testName: "login success",
			au: &auth{
				vClient: vc,
				creds:   awsCred,
				role:    "good",
				path:    "aws",
			},
			expectErr: false,
		},
		{
			testName: "login success, auth enabled in another path",
			au: &auth{
				vClient: vc,
				role:    "try",
				path:    "test",
			},
			expectErr: false,
		},
		{
			testName: "login failed, bad user/password",
			au: &auth{
				vClient: vc,
				role:    "bad",
				path:    "aws",
			},
			expectErr: true,
		},
	}

	for _, c := range cases {
		t.Run(c.testName, func(t *testing.T) {
			token, err := c.au.Login()
			if c.expectErr {
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

func TestLogin(t *testing.T) {
	addr := os.Getenv("VAULT_ADDR")
	accessKey := os.Getenv("AWS_ACCESS_KEY")
	secretKey := os.Getenv("AWS_SECRET_KEY")
	role := os.Getenv("VAULT_ROLE")
	if addr == "" || accessKey == "" || secretKey == "" || role == "" {
		t.Skip()
	}

	au, err := New(&authtype.AuthInfo{
		VaultApp: &appcat.AppBinding{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "aws",
				Namespace: "default",
			},
			Spec: appcat.AppBindingSpec{
				ClientConfig: appcat.ClientConfig{
					URL:                   &addr,
					InsecureSkipTLSVerify: true,
				},
				Secret: &core.LocalObjectReference{
					Name: "aws",
				},
				Parameters: &runtime.RawExtension{
					Raw: []byte(fmt.Sprintf(`{ "role" : "%s" }`, role)),
				},
			},
		},
		ServiceAccountRef: nil,
		Secret: &core.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "aws",
				Namespace: "default",
			},
			Data: map[string][]byte{
				"access_key_id":     []byte(accessKey),
				"secret_access_key": []byte(secretKey),
			},
		},
		VaultRole: "",
		Path:      "",
	})

	assert.Nil(t, err)

	token, err := au.Login()
	if assert.Nil(t, err) {
		fmt.Println(token)
	}
}
