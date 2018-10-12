package aws

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
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
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
	m := pat.New()
	m.Post("/v1/auth/aws/login/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var v map[string]interface{}
		defer r.Body.Close()
		json.NewDecoder(r.Body).Decode(&v)
		if val, ok := v["role"]; ok {
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
				role:    "good",
			},
			expectErr: false,
		},
		{
			testName: "login failed, bad user/password",
			au: &auth{
				vClient: vc,
				role:    "bad",
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

	au, err := New(&appcat.AppBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "aws",
			Namespace: "default",
		},
		Spec: appcat.AppBindingSpec{
			ClientConfig: appcat.ClientConfig{
				URL: &addr,
				InsecureSkipTLSVerify: true,
			},
			Secret: &core.LocalObjectReference{
				Name: "aws",
			},
			Parameters: &runtime.RawExtension{
				Raw: []byte(fmt.Sprintf(`{ "role" : "%s" }`, role)),
			},
		},
	}, &core.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "aws",
			Namespace: "default",
		},
		Data: map[string][]byte{
			"access_key_id":     []byte(accessKey),
			"secret_access_key": []byte(secretKey),
		},
	})

	token, err := au.Login()
	if assert.Nil(t, err) {
		fmt.Println(token)
	}
}
