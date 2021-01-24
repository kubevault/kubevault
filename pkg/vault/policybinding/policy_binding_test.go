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

package policybinding

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	policyapi "kubevault.dev/apimachinery/apis/policy/v1alpha1"

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
			path:         "kubernetes/role",
		},
		authAppRole: &pBindingAppRole{
			roleName:             "ok",
			path:                 "approle/role",
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
		authLdapGroup: &pBindingLdapGroup{
			name: "ok",
			path: "ldap/groups",
		},
		authLdapUser: &pBindingLdapUser{
			username: "ok",
			path:     "ldap/users",
			Groups:   []string{"group1", "group2"},
		},
		authJWT: &pBindingJWT{
			name:                 "ok",
			path:                 "jwt/role",
			RoleType:             "oidc",
			BoundAudiences:       []string{"FFXlsY2atr_aaaa_hMtsE-zTAeTZnu8"},
			UserClaim:            "sub",
			BoundSubject:         "bound_sub",
			BoundClaims:          map[string]string{"test": "test"},
			BoundClaimsType:      "glob",
			GroupClaim:           "group_claim",
			ClaimMappings:        map[string]string{"test": "test"},
			OIDCScopes:           []string{"openid"},
			AllowedRedirectUris:  []string{"http://localhost"},
			VerboseOIDCLogging:   true,
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
		authKubernetes: &pBindingKubernetes{name: "err", path: "kubernetes/role"},
		authAppRole:    &pBindingAppRole{roleName: "err", path: "approle/role"},
		authLdapGroup:  &pBindingLdapGroup{name: "err", path: "ldap/groups"},
		authLdapUser:   &pBindingLdapUser{username: "err", path: "ldap/groups"},
		authJWT:        &pBindingJWT{name: "err", path: "jwt/role"},
	}
)

func TestIsKeyValExist(t *testing.T) {
	cases := []struct {
		testName  string
		store     map[string]interface{}
		key       string
		val       interface{}
		expectErr bool
	}{
		{
			testName:  "string_good",
			store:     map[string]interface{}{"test": "string"},
			key:       "test",
			val:       "string",
			expectErr: false,
		},
		{
			testName:  "string_bad",
			store:     map[string]interface{}{"test": "string"},
			key:       "test",
			val:       "bad",
			expectErr: true,
		},
		{
			testName:  "int64_good",
			store:     map[string]interface{}{"test": float64(10)},
			key:       "test",
			val:       int64(10),
			expectErr: false,
		},
		{
			testName:  "int64_bad",
			store:     map[string]interface{}{"test": float64(5)},
			key:       "test",
			val:       int64(10),
			expectErr: true,
		},
		{
			testName:  "bool_good",
			store:     map[string]interface{}{"test": true},
			key:       "test",
			val:       true,
			expectErr: false,
		},
		{
			testName:  "bool_bad",
			store:     map[string]interface{}{"test": false},
			key:       "test",
			val:       true,
			expectErr: true,
		},
		{
			testName:  "slice_good",
			store:     map[string]interface{}{"test": []string{"test1", "test2"}},
			key:       "test",
			val:       []string{"test1", "test2"},
			expectErr: false,
		},
		{
			testName:  "slice_bad",
			store:     map[string]interface{}{"test": []string{"test1", "test2"}},
			key:       "test",
			val:       []string{"test1", "test3"},
			expectErr: true,
		},
		{
			testName:  "map_good",
			store:     map[string]interface{}{"test": map[string]string{"test1": "test2"}},
			key:       "test",
			val:       map[string]string{"test1": "test2"},
			expectErr: false,
		},
		{
			testName:  "map_bad",
			store:     map[string]interface{}{"test": map[string]string{"test1": "test2"}},
			key:       "test",
			val:       map[string]string{"test1": "test3"},
			expectErr: true,
		},
	}

	for _, c := range cases {
		t.Run(c.testName, func(t *testing.T) {
			res := isKeyValExist(c.store, c.key, c.val)
			if c.expectErr {
				assert.Equal(t, false, res)
			} else {
				assert.Equal(t, true, res)
			}
		})
	}
}

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
		case map[string]string:
			switch x := v.(type) {
			case map[string]string:
				if len(y) != len(x) {
					return false
				}
				for k, kV := range y {
					if cV, ok := x[k]; !ok || cV != kV {
						return false
					}
				}
			case map[string]interface{}:
				if len(y) != len(x) {
					return false
				}
				for k, kV := range y {
					if cV, ok := x[k]; !ok || cV != kV {
						return false
					}
				}
			default:
				return false
			}
			return true
		default:
			return false
		}
	}
	return false
}

func NewFakeVaultServer() *httptest.Server {
	router := mux.NewRouter()

	// kubernetes

	router.HandleFunc("/v1/auth/kubernetes/role/ok", func(w http.ResponseWriter, r *http.Request) {
		v := map[string]interface{}{}
		utilruntime.Must(json.NewDecoder(r.Body).Decode(&v))
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

	// approle
	router.HandleFunc("/v1/auth/approle/role/ok", func(w http.ResponseWriter, r *http.Request) {
		v := map[string]interface{}{}
		utilruntime.Must(json.NewDecoder(r.Body).Decode(&v))
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

	// ldap group
	router.HandleFunc("/v1/auth/ldap/groups/ok", func(w http.ResponseWriter, r *http.Request) {
		v := map[string]interface{}{}
		utilruntime.Must(json.NewDecoder(r.Body).Decode(&v))
		if ok := isKeyValExist(v, "policies", goodPBind.authLdapGroup.Policies); !ok {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	}).Methods(http.MethodPut)

	router.HandleFunc("/v1/auth/ldap/groups/ok", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}).Methods(http.MethodDelete)

	router.HandleFunc("/v1/auth/ldap/groups/err", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	}).Methods(http.MethodDelete)

	// ldap user
	router.HandleFunc("/v1/auth/ldap/users/ok", func(w http.ResponseWriter, r *http.Request) {
		v := map[string]interface{}{}
		utilruntime.Must(json.NewDecoder(r.Body).Decode(&v))
		if ok := isKeyValExist(v, "policies", goodPBind.authLdapUser.Policies); !ok {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if ok := isKeyValExist(v, "groups", goodPBind.authLdapUser.Groups); !ok {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	}).Methods(http.MethodPut)

	router.HandleFunc("/v1/auth/ldap/users/ok", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}).Methods(http.MethodDelete)

	router.HandleFunc("/v1/auth/ldap/users/err", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	}).Methods(http.MethodDelete)

	// jwt
	router.HandleFunc("/v1/auth/jwt/role/ok", func(w http.ResponseWriter, r *http.Request) {
		v := map[string]interface{}{}
		utilruntime.Must(json.NewDecoder(r.Body).Decode(&v))
		if ok := isKeyValExist(v, "role_type", goodPBind.authJWT.RoleType); !ok {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if ok := isKeyValExist(v, "bound_audiences", goodPBind.authJWT.BoundAudiences); !ok {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if ok := isKeyValExist(v, "user_claim", goodPBind.authJWT.UserClaim); !ok {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if ok := isKeyValExist(v, "bound_subject", goodPBind.authJWT.BoundSubject); !ok {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if ok := isKeyValExist(v, "bound_claims", goodPBind.authJWT.BoundClaims); !ok {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if ok := isKeyValExist(v, "bound_claims_type", goodPBind.authJWT.BoundClaimsType); !ok {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if ok := isKeyValExist(v, "group_claim", goodPBind.authJWT.GroupClaim); !ok {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if ok := isKeyValExist(v, "claim_mappings", goodPBind.authJWT.ClaimMappings); !ok {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if ok := isKeyValExist(v, "oidc_scopes", goodPBind.authJWT.OIDCScopes); !ok {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if ok := isKeyValExist(v, "allowed_redirect_uris", goodPBind.authJWT.AllowedRedirectUris); !ok {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if ok := isKeyValExist(v, "verbose_oidc_logging", goodPBind.authJWT.VerboseOIDCLogging); !ok {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if ok := isKeyValExist(v, "token_ttl", goodPBind.authJWT.TokenTTL); !ok {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if ok := isKeyValExist(v, "token_max_ttl", goodPBind.authJWT.TokenMaxTTL); !ok {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if ok := isKeyValExist(v, "token_policies", goodPBind.authJWT.TokenPolicies); !ok {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if ok := isKeyValExist(v, "token_bound_cidrs", goodPBind.authJWT.TokenBoundCidrs); !ok {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if ok := isKeyValExist(v, "token_explicit_max_ttl", goodPBind.authJWT.TokenExplicitMaxTTL); !ok {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if ok := isKeyValExist(v, "token_no_default_policy", goodPBind.authJWT.TokenNoDefaultPolicy); !ok {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if ok := isKeyValExist(v, "token_num_uses", goodPBind.authJWT.TokenNumUses); !ok {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if ok := isKeyValExist(v, "token_period", goodPBind.authJWT.TokenPeriod); !ok {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if ok := isKeyValExist(v, "token_type", goodPBind.authJWT.TokenType); !ok {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

	}).Methods(http.MethodPut)

	router.HandleFunc("/v1/auth/jwt/role/ok", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}).Methods(http.MethodDelete)

	router.HandleFunc("/v1/auth/jwt/role/err", func(w http.ResponseWriter, r *http.Request) {
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
		pb        *pBinding
		expectErr bool
	}{
		{
			testName:  "no error",
			pb:        goodPBind,
			expectErr: false,
		},
		{
			testName:  "error, some fields are missing",
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
			pb:        goodPBind,
			expectErr: false,
		},
		{
			testName:  "error",
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
