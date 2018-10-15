package policybinding

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/appscode/pat"
	vaultapi "github.com/hashicorp/vault/api"
	"github.com/stretchr/testify/assert"
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

	switch val.(type) {
	case []string:
		y := val.([]string)
		switch v.(type) {
		case []string:
			x := v.([]string)
			for p := range x {
				if x[p] != y[p] {
					return false
				}
			}
			return true
		case []interface{}:
			x := v.([]interface{})
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
		y := val.(string)
		switch v.(type) {
		case string:
			return v.(string) == y
		default:
			return false
		}

	}
	return false
}

func NewFakeVaultServer() *httptest.Server {
	m := pat.New()
	m.Post("/v1/auth/kubernetes/role/ok", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		v := map[string]interface{}{}
		json.NewDecoder(r.Body).Decode(&v)
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
	}))
	m.Post("/v1/auth/test/role/try", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	m.Del("/v1/auth/kubernetes/role/ok", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	m.Del("/v1/auth/test/role/try", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	m.Del("/v1/auth/kubernetes/role/err", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	}))

	return httptest.NewServer(m)
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
	cfg.ConfigureTLS(&vaultapi.TLSConfig{
		Insecure: true,
	})
	c, err := vaultapi.NewClient(cfg)
	if err != nil {
		return nil, err
	}

	c.SetAddress(addr)
	c.SetToken(token)
	return c, nil
}
