package kubernetes

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/appscode/pat"
	vaultapi "github.com/hashicorp/vault/api"
	"github.com/stretchr/testify/assert"
)

const authResp = `
{
  "auth": {
    "client_token": "1234"
  }
}
`

func NewFakeVaultServer() *httptest.Server {
	m := pat.New()
	m.Post("/v1/auth/kubernetes/login", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var v map[string]interface{}
		defer r.Body.Close()
		json.NewDecoder(r.Body).Decode(&v)
		if val, ok := v["jwt"]; ok {
			if val.(string) == "good" {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(authResp))
				return
			}
		}
		w.WriteHeader(http.StatusBadRequest)
	}))

	return httptest.NewServer(m)
}

func TestAuth_Login(t *testing.T) {
	srv := NewFakeVaultServer()
	defer srv.Close()

	vc, err := vaultapi.NewClient(vaultapi.DefaultConfig())
	if !assert.Nil(t, err) {
		return
	}
	vc.SetAddress(srv.URL)

	cases := []struct {
		testName  string
		au        *auth
		expectErr bool
	}{
		{
			testName: "login success",
			au: &auth{
				vClient: vc,
				jwt:     "good",
				role:    "demo",
			},
			expectErr: false,
		},
		{
			testName: "login failed, bad jwt",
			au: &auth{
				vClient: vc,
				jwt:     "bad",
				role:    "demo",
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
	jwt := os.Getenv("JWT")
	role := os.Getenv("ROLE")
	if addr == "" || jwt == "" || role == "" {
		t.Skip()
	}
	vc, err := vaultapi.NewClient(vaultapi.DefaultConfig())
	vc.SetAddress(addr)
	if !assert.Nil(t, err) {
		return
	}

	au := &auth{
		vClient: vc,
		jwt:     jwt,
		role:    role,
	}

	token, err := au.Login()
	if assert.Nil(t, err) {
		fmt.Println(token)
	}
}
