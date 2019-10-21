package userpass

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gorilla/mux"
	vaultapi "github.com/hashicorp/vault/api"
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

	router.HandleFunc("/v1/auth/userpass/login/{username}", func(w http.ResponseWriter, r *http.Request) {
		var v map[string]interface{}
		defer r.Body.Close()
		utilruntime.Must(json.NewDecoder(r.Body).Decode(&v))
		if val, ok := v["password"]; ok {
			if val.(string) == "good" {
				w.WriteHeader(http.StatusOK)
				_, err := w.Write([]byte(authResp))
				utilruntime.Must(err)
				return
			}
		}
		w.WriteHeader(http.StatusBadRequest)
	}).Methods(http.MethodPost)

	router.HandleFunc("/v1/auth/test/login/{username}", func(w http.ResponseWriter, r *http.Request) {
		var v map[string]interface{}
		defer r.Body.Close()
		utilruntime.Must(json.NewDecoder(r.Body).Decode(&v))
		if val, ok := v["password"]; ok {
			if val.(string) == "try" {
				w.WriteHeader(http.StatusOK)
				_, err := w.Write([]byte(authResp))
				utilruntime.Must(err)
				return
			}
		}
		w.WriteHeader(http.StatusBadRequest)
	}).Methods(http.MethodPost)

	return httptest.NewServer(router)
}

func TestAuth_Login(t *testing.T) {
	srv := NewFakeVaultServer()
	defer srv.Close()

	vc, err := vaultapi.NewClient(vaultapi.DefaultConfig())
	if !assert.Nil(t, err) {
		return
	}
	utilruntime.Must(vc.SetAddress(srv.URL))

	cases := []struct {
		testName  string
		au        *auth
		expectErr bool
	}{
		{
			testName: "login success",
			au: &auth{
				vClient: vc,
				user:    "test",
				pass:    "good",
				path:    "userpass",
			},
			expectErr: false,
		},
		{
			testName: "login success, auth is enabled in another path",
			au: &auth{
				vClient: vc,
				user:    "test",
				pass:    "try",
				path:    "test",
			},
			expectErr: false,
		},
		{
			testName: "login failed, bad user/password",
			au: &auth{
				vClient: vc,
				user:    "test",
				pass:    "bad",
				path:    "userpass",
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
	user := os.Getenv("VAULT_USERNAME")
	pass := os.Getenv("VAULT_PASSWORD")
	if addr == "" || user == "" || pass == "" {
		t.Skip()
	}
	vc, err := vaultapi.NewClient(vaultapi.DefaultConfig())
	if !assert.Nil(t, err) {
		return
	}
	err = vc.SetAddress(addr)
	if !assert.Nil(t, err) {
		return
	}

	au := &auth{
		vClient: vc,
		user:    user,
		pass:    pass,
	}

	token, err := au.Login()
	if assert.Nil(t, err) {
		fmt.Println(token)
	}
}
