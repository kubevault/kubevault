package policybinding

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	vaultapi "github.com/hashicorp/vault/api"
	"github.com/stretchr/testify/assert"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
)

var (
	goodPBind = &pBinding{
		policies:     []string{"test,hi"},
		saNames:      []string{"test1,test2"},
		saNamespaces: []string{"test3,test4"},
		ttl:          "100",
		maxTTL:       "100",
		period:       "100",
		path:         "kubernetes",
	}
	badPBind = &pBinding{}
)

func isKeyValExist(store map[string]interface{}, key string, val interface{}) bool {
	v, ok := store[key]
	if !ok {
		return ok
	}

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
		if ok := isKeyValExist(v, "bound_service_account_names", goodPBind.saNames); !ok {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if ok := isKeyValExist(v, "bound_service_account_namespaces", goodPBind.saNamespaces); !ok {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if ok := isKeyValExist(v, "policies", goodPBind.policies); !ok {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if ok := isKeyValExist(v, "ttl", goodPBind.ttl); !ok {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if ok := isKeyValExist(v, "max_ttl", goodPBind.maxTTL); !ok {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if ok := isKeyValExist(v, "period", goodPBind.period); !ok {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusOK)
	}).Methods(http.MethodPost)

	router.HandleFunc("/v1/auth/test/role/try", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}).Methods(http.MethodPost)

	router.HandleFunc("/v1/auth/kubernetes/role/ok", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}).Methods(http.MethodDelete)

	router.HandleFunc("/v1/auth/test/role/try", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}).Methods(http.MethodDelete)

	router.HandleFunc("/v1/auth/kubernetes/role/err", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	}).Methods(http.MethodDelete)

	return httptest.NewServer(router)
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
			testName:  "no error, auth enabled in different path",
			name:      "try",
			pb:        func(p pBinding) *pBinding { p.path = "test"; return &p }(*goodPBind),
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
			err := c.pb.Ensure(c.name)
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
			testName:  "no error, auth enabled in different path",
			name:      "try",
			pb:        func(p pBinding) *pBinding { p.path = "test"; return &p }(*goodPBind),
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
			err := c.pb.Delete(c.name)
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
