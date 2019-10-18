package kubernetes

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
	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	appcat "kmodules.xyz/custom-resources/apis/appcatalog/v1alpha1"
	"kubevault.dev/operator/pkg/vault/util"
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
	router.HandleFunc("/v1/auth/kubernetes/login", func(w http.ResponseWriter, r *http.Request) {
		var v map[string]interface{}
		defer r.Body.Close()
		util.LogErr(json.NewDecoder(r.Body).Decode(&v))
		if val, ok := v["jwt"]; ok {
			if val.(string) == "good" {
				w.WriteHeader(http.StatusOK)
				util.LogWriteErr(w.Write([]byte(authResp)))
				return
			}
		}
		w.WriteHeader(http.StatusBadRequest)
	}).Methods(http.MethodPost)

	router.HandleFunc("/v1/auth/test/login", func(w http.ResponseWriter, r *http.Request) {
		var v map[string]interface{}
		defer r.Body.Close()
		util.LogErr(json.NewDecoder(r.Body).Decode(&v))
		if val, ok := v["jwt"]; ok {
			if val.(string) == "try" {
				w.WriteHeader(http.StatusOK)
				util.LogWriteErr(w.Write([]byte(authResp)))
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
	util.LogErr(vc.SetAddress(srv.URL))

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
				path:    "kubernetes",
			},
			expectErr: false,
		},
		{
			testName: "login success, auth enabled in another path",
			au: &auth{
				vClient: vc,
				jwt:     "try",
				role:    "try",
				path:    "test",
			},
			expectErr: false,
		},
		{
			testName: "login failed, bad jwt",
			au: &auth{
				vClient: vc,
				jwt:     "bad",
				role:    "demo",
				path:    "kubernetes",
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
	jwt := os.Getenv("K8S_JWT")
	role := os.Getenv("VAULT_ROLE")

	if addr == "" || jwt == "" || role == "" {
		t.Skip()
	}

	au, err := New(&appcat.AppBinding{
		Spec: appcat.AppBindingSpec{
			ClientConfig: appcat.ClientConfig{
				URL: &addr,
			},
			Parameters: &runtime.RawExtension{
				Raw: []byte(fmt.Sprintf(`{"role":"%s"}`, role)),
			},
		},
	}, &core.Secret{
		Data: map[string][]byte{
			"token": []byte(jwt),
		},
	})
	if !assert.Nil(t, err) {
		return
	}

	token, err := au.Login()
	if assert.Nil(t, err) {
		fmt.Println(token)
	}
}
