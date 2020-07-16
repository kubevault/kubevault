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

package policybinding

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	policyapi "kubevault.dev/operator/apis/policy/v1alpha1"

	"github.com/gorilla/mux"
	vaultapi "github.com/hashicorp/vault/api"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
)

var (
	goodPBind = &pBinding{
		policies: []string{"test,hi"},
		authKubernetes: &pBindingKubernetes{
			name:         "ok",
			SaNames:      []string{"test1,test2"},
			SaNamespaces: []string{"test3,test4"},
			TokenTTL:     "100",
			TokenMaxTTL:  "100",
			TokenPeriod:  "100",
			path:         "kubernetes",
		},
		authAppRole: &pBindingAppRole{
			roleName:             "ok",
			path:                 "approle",
			BindSecretID:         true,
			SecretIDBoundCidrs:   []string{"192.168.0.200/32"},
			SecretIDNumUses:      200,
			SecretIDTTL:          "60",
			EnableLocalSecretIDs: true,
			TokenTTL:             60,
			TokenMaxTTL:          60,
			TokenBoundCidrs:      []string{"192.168.0.200/32"},
			TokenExplicitMaxTTL:  60,
			TokenNoDefaultPolicy: true,
			TokenNumUses:         60,
			TokenPeriod:          60,
			TokenType:            "default",
		},
	}
	badPBind = &pBinding{
		authKubernetes: &pBindingKubernetes{},
		authAppRole:    &pBindingAppRole{},
	}
)

func isKeyValExist(store map[string]interface{}, key string, val interface{}) bool {
	if v, ok := store[key]; ok {
		switch y := val.(type) {
		case []string:
			switch x := v.(type) {
			case []string:
				for p := range x {
					if x[p] != y[p] {
						return false
					}
				}
				return true
			case []interface{}:
				for p := range x {
					if x[p].(string) != y[p] {
						return false
					}
				}
				return true
			default:
				return false
			}
		case string:
			switch z := v.(type) {
			case string:
				return z == y
			default:
				return false
			}
		case int64:
			return v == float64(val.(int64))
		case bool:
			return v == val
		default:
			return false
		}
	}
	return false
}

func NewFakeVaultServer() *httptest.Server {
	router := mux.NewRouter()

	router.HandleFunc("/v1/auth/kubernetes/role/ok", func(w http.ResponseWriter, r *http.Request) {
		v := map[string]interface{}{}
		utilruntime.Must(json.NewDecoder(r.Body).Decode(&v))
		fmt.Println("***")
		fmt.Println(v)
		fmt.Println("***")
		if ok := isKeyValExist(v, "bound_service_account_names", goodPBind.authKubernetes.SaNames); !ok {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if ok := isKeyValExist(v, "bound_service_account_namespaces", goodPBind.authKubernetes.SaNamespaces); !ok {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if ok := isKeyValExist(v, "token_policies", goodPBind.authKubernetes.TokenPolicies); !ok {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if ok := isKeyValExist(v, "token_ttl", goodPBind.authKubernetes.TokenTTL); !ok {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if ok := isKeyValExist(v, "token_max_ttl", goodPBind.authKubernetes.TokenMaxTTL); !ok {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if ok := isKeyValExist(v, "token_period", goodPBind.authKubernetes.TokenPeriod); !ok {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusOK)
	}).Methods(http.MethodPut)

	router.HandleFunc("/v1/auth/kubernetes/role/ok", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}).Methods(http.MethodDelete)

	router.HandleFunc("/v1/auth/kubernetes/role/err", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	}).Methods(http.MethodDelete)

	router.HandleFunc("/v1/auth/approle/role/ok", func(w http.ResponseWriter, r *http.Request) {
		v := map[string]interface{}{}
		utilruntime.Must(json.NewDecoder(r.Body).Decode(&v))
		fmt.Println("***")
		fmt.Println(v)
		fmt.Println("***")
		if ok := isKeyValExist(v, "bind_secret_id", goodPBind.authAppRole.BindSecretID); !ok {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if ok := isKeyValExist(v, "secret_id_bound_cidrs", goodPBind.authAppRole.SecretIDBoundCidrs); !ok {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if ok := isKeyValExist(v, "secret_id_num_uses", goodPBind.authAppRole.SecretIDNumUses); !ok {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if ok := isKeyValExist(v, "secret_id_ttl", goodPBind.authAppRole.SecretIDTTL); !ok {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if ok := isKeyValExist(v, "enable_local_secret_ids", goodPBind.authAppRole.EnableLocalSecretIDs); !ok {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if ok := isKeyValExist(v, "token_ttl", goodPBind.authAppRole.TokenTTL); !ok {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if ok := isKeyValExist(v, "token_ttl", goodPBind.authAppRole.TokenTTL); !ok {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if ok := isKeyValExist(v, "token_max_ttl", goodPBind.authAppRole.TokenMaxTTL); !ok {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if ok := isKeyValExist(v, "token_policies", goodPBind.authAppRole.TokenPolicies); !ok {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if ok := isKeyValExist(v, "token_bound_cidrs", goodPBind.authAppRole.TokenBoundCidrs); !ok {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if ok := isKeyValExist(v, "token_explicit_max_ttl", goodPBind.authAppRole.TokenExplicitMaxTTL); !ok {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if ok := isKeyValExist(v, "token_no_default_policy", goodPBind.authAppRole.TokenNoDefaultPolicy); !ok {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if ok := isKeyValExist(v, "token_num_uses", goodPBind.authAppRole.TokenNumUses); !ok {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if ok := isKeyValExist(v, "token_period", goodPBind.authAppRole.TokenPeriod); !ok {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if ok := isKeyValExist(v, "token_type", goodPBind.authAppRole.TokenType); !ok {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusOK)
	}).Methods(http.MethodPut)

	router.HandleFunc("/v1/auth/approle/role/ok", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}).Methods(http.MethodDelete)

	router.HandleFunc("/v1/auth/approle/role/err", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	}).Methods(http.MethodDelete)

	return httptest.NewServer(router)
}

func simpleVaultPolicyBinding() *policyapi.VaultPolicyBinding {
	return &policyapi.VaultPolicyBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "simple",
			Namespace: "test",
		},
		Spec: policyapi.VaultPolicyBindingSpec{
			SubjectRef: policyapi.SubjectRef{
				Kubernetes: &policyapi.KubernetesSubjectRef{},
				AppRole:    &policyapi.AppRoleSubjectRef{},
			},
		},
	}
}

func TestEnsure(t *testing.T) {
	srv := NewFakeVaultServer()
	defer srv.Close()

	vc, err := vaultClient(srv.URL, "root")
	if !assert.Nil(t, err, "failed to create vault client") {
		return
	}
	goodPBind.vClient = vc
	badPBind.vClient = vc

	cases := []struct {
		testName  string
		name      string
		pb        *pBinding
		expectErr bool
	}{
		{
			testName:  "no error",
			name:      "ok",
			pb:        goodPBind,
			expectErr: false,
		},
		{
			testName:  "error, some fields are missing",
			name:      "ok",
			pb:        badPBind,
			expectErr: true,
		},
	}

	for _, c := range cases {
		t.Run(c.testName, func(t *testing.T) {
			err := c.pb.Ensure(simpleVaultPolicyBinding())
			if c.expectErr {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func TestDelete(t *testing.T) {
	srv := NewFakeVaultServer()
	defer srv.Close()

	vc, err := vaultClient(srv.URL, "root")
	if !assert.Nil(t, err, "failed to create vault client") {
		return
	}
	goodPBind.vClient = vc
	badPBind.vClient = vc

	cases := []struct {
		testName  string
		name      string
		pb        *pBinding
		expectErr bool
	}{
		{
			testName:  "no error",
			name:      "ok",
			pb:        goodPBind,
			expectErr: false,
		},
		{
			testName:  "error",
			name:      "err",
			pb:        badPBind,
			expectErr: true,
		},
	}

	for _, c := range cases {
		t.Run(c.testName, func(t *testing.T) {
			err := c.pb.Delete(simpleVaultPolicyBinding())
			if c.expectErr {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
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
