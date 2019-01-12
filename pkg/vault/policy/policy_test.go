package policy

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/appscode/pat"
	vautlapi "github.com/hashicorp/vault/api"
	"github.com/stretchr/testify/assert"
)

func NewFakeVaultServer() *httptest.Server {
	m := pat.New()
	m.Put("/v1/sys/policies/acl/ok", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	m.Put("/v1/sys/policies/acl/err", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	m.Del("/v1/sys/policies/acl/ok", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	m.Del("/v1/sys/policies/acl/err", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))

	return httptest.NewServer(m)
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

func vaultClient(addr, token string) (*vautlapi.Client, error) {
	cfg := vautlapi.DefaultConfig()
	cfg.ConfigureTLS(&vautlapi.TLSConfig{
		Insecure: true,
	})
	c, err := vautlapi.NewClient(cfg)
	if err != nil {
		return nil, err
	}

	c.SetToken(token)
	c.SetAddress(addr)
	return c, nil
}
