package azure

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/appscode/pat"
	"github.com/hashicorp/vault/api"
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
	m.Post("/v1/auth/azure/login", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var v map[string]interface{}
		defer r.Body.Close()
		json.NewDecoder(r.Body).Decode(&v)

		value, ok := v["role"]
		if !ok || value == "" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		value, ok = v["jwt"]
		if !ok || value == "" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(authResp))
	}))

	m.Post("/v1/auth/my-path/login/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var v map[string]interface{}
		defer r.Body.Close()
		json.NewDecoder(r.Body).Decode(&v)

		value, ok := v["role"]
		if !ok || value == "" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		value, ok = v["jwt"]
		if !ok || value == "" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(authResp))
	}))

	return httptest.NewServer(m)
}

func TestAuth_Login(t *testing.T) {
	srv := NewFakeVaultServer()
	defer srv.Close()

	vc, err := api.NewClient(api.DefaultConfig())
	if !assert.Nil(t, err) {
		return
	}

	vc.SetAddress(srv.URL)

	cases := []struct {
		testName    string
		au          *auth
		expectedErr bool
	}{
		{
			testName: "Successful login",
			au: &auth{
				vClient:           vc,
				role:              "role-one",
				path:              "azure",
				signedJwt:         "kf09eiurekijlkjflkdj9f.flkjdlf.fjdflkjd",
				subscriptionId:    "565-5443-5-4545-43",
				resourceGroupName: "vault-test",
				vmName:            "vault",
			},
			expectedErr: false,
		},
		{
			testName: "Empty role, unsuccessful login",
			au: &auth{
				vClient:           vc,
				path:              "azure",
				signedJwt:         "dfdfd.sfdsf.dfdsdfds",
				subscriptionId:    "342-0324-03234",
				resourceGroupName: "vault-test",
				vmName:            "test",
			},
			expectedErr: true,
		},
		{
			testName: "Empty jwt, unsuccessful login",
			au: &auth{
				vClient:        vc,
				role:           "role3",
				path:           "azure",
				subscriptionId: "342-234-324-3",
			},
			expectedErr: true,
		},
		{
			testName: "Successful login at user defined path",
			au: &auth{
				vClient:           vc,
				role:              "role",
				path:              "my-path",
				signedJwt:         "fdsf.fdfkjdsfro9r.dfdfd",
				subscriptionId:    "23432-34-324-2343",
				resourceGroupName: "test",
				vmName:            "vault",
			},
			expectedErr: false,
		},
	}

	for _, c := range cases {
		t.Run(c.testName, func(t *testing.T) {
			token, err := c.au.Login()
			if c.expectedErr {
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
