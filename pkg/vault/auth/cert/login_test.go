package cert

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
	m.Post("/v1/auth/cert/login/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var v map[string]interface{}
		defer r.Body.Close()
		json.NewDecoder(r.Body).Decode(&v)
		if val, ok := v["name"]; ok {
			if val.(string) == "good" {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(authResp))
				return
			}
		}
		w.WriteHeader(http.StatusBadRequest)
	}))

	m.Post("/v1/auth/test/login/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var v map[string]interface{}
		defer r.Body.Close()
		json.NewDecoder(r.Body).Decode(&v)
		if val, ok := v["name"]; ok {
			if val.(string) == "try" {
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
				name:    "good",
				path:    "cert",
			},
			expectErr: false,
		},
		{
			testName: "login success, auth enabled in another path",
			au: &auth{
				vClient: vc,
				name:    "try",
				path:    "test",
			},
			expectErr: false,
		},
		{
			testName: "login failed, bad user/password",
			au: &auth{
				vClient: vc,
				name:    "bad",
				path:    "cert",
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
	if addr == "" {
		t.Skip()
	}

	app := &appcat.AppBinding{
		Spec: appcat.AppBindingSpec{
			ClientConfig: appcat.ClientConfig{
				URL: &addr,
			},
			Parameters: &runtime.RawExtension{
				Raw: []byte(fmt.Sprintf(`{ "name" : "demo" }`)),
			},
		},
	}
	au, err := New(app, &core.Secret{
		Data: map[string][]byte{
			"tls.key": []byte(`-----BEGIN RSA PRIVATE KEY-----
MIIEpQIBAAKCAQEAvzavl3xpk8A08Mq2zjdLX7wbl6rny8dPJ2gPTwbMi+rto+q8
73+6vKfDsTN/p+Wr7r7oaXv2xd5Dd0WVaQzkJ3yAg5VhrhT9YCVOeQ8ND4+FQjO7
hcvXIzVyhphX2RFs0ASvT8bxAttyHZyaPyKhZ15XAnOyZNw3918cQbZT1TwtKXo7
dhJudyFADv05kIGND1obmrlh+VlbOIuUBSzeS6jCaFjEjK/l4pnqNvR2GK2jmETj
yJZaV7WkA/zx1kbrUhegREntCuA1cfl8biepadfCsw5RCa0DK4D9zYMWy9Uau0SA
ExefCuwv2h8UtOCW4bvLtruGiR1tyJu2CyNodwIDAQABAoIBAE6CCX5NGpwLYrEq
yfvJQC1CcqHHDfzhDGLFmuN7iyg3gPK4QnKjIuyqhPKQjm1FI16XC52jxCJhq5mg
/ENxg4ui1rEv+Dcdbxq2to2F3HrzFmekDe5VTzOpkigpDIiXWvHduo0qxXHC4AvA
bKRXd6WVWmwrTKeUs3Xhmxxv2+PEZx8QMJQXNu/uGqy01IPNp2KFKm6X0lWU/eH0
ZTjbWkpL4N1JwQQfJxEGStdNf7fjMc8WFHCqeFCJ29B8Z1V06E5Xgu0kDCmt/e1a
U7N7Rr5s74Z1VLU8uJHDDwsnFjiCl8A1MamtlumV0mOgRYUQf+CDGtokR+MtLlOa
Gj0BeCECgYEA7uTjwSUuDBojDxdx55cfjGm/ieuRWe6h+VfnmDQw0E6oS0wr8DwQ
ci1Zg7Gc0uolxe+q37ws/YvC7uwk4Ss/l2+YhTLUP1AAFTsmggFGIn6o+X7+k/Nm
tVrrdNg+zcU1cahcitTmtmZsWxRiVBmBqYcRf6k69y9KhSoLpzydFh0CgYEAzOfG
2ZEEjOrkcN+mZ3gzlRmhS72FG0jeFwg8GYaeVPI/37a7ZtlkoDF9eyG/YuSF3u5I
UKDWuZXtL5wfiQUuFu1Emirtu0ylXyoSthkrGvoQx5KfGhvHco2IoGEYlZj6Tm/U
MHfaNmXkEeA8m7fkrwJsifUUKIeh9HJ9gdt3ZKMCgYEAkEL4pnJlVDmUYlCuIERK
cOiLGiZ/J+fLOF+1I1yg/aoCRzZAclpTNB/epoBjS5rKJLWOYn2oTZRQqyc/Php3
1GM3n3gKZBFTe360yl0qlToXoFLoOUALDglRlsXfZzNoCrK4772RdSR02qt8lXyx
qEZAcu8nBI4yWigB0YPw+KECgYEAtAfMianFkr5qmdWW0gAlahILyo0oXvGl2Byv
GUpS4JW7kyZs/w9wPuNcuYvMKOpZyKYZOWYnYwWcUKFef7fiZ9ht1vpyx4avIa7I
o9/3JIujpIVpbroLgdVivm6w9/dhrPrKNw+G1RauzRn0hmiK70006f0/ieCpZioV
pbua6fsCgYEAkXYt+RRf1fMKJP/QuSUlH7xyxJCnNbTLC7G4R1E0zbNxcuheCVd7
UsZgpyOlbAaENmheid3lBke/z+lljcyJE5EUW3estslRAQWvaEQUtq/2NDVAMmFf
Gqe3rYqAdJ3bWBBk+8gi8zyY2pbQpTIg6MSSMazHFPoThdBZ1pWIkss=
-----END RSA PRIVATE KEY-----
`),
			"tls.crt": []byte(`-----BEGIN CERTIFICATE-----
MIIC3jCCAcagAwIBAgIIL7N6jjiZ5p0wDQYJKoZIhvcNAQELBQAwDTELMAkGA1UE
AxMCY2EwHhcNMTgwOTI3MDQ1NDI5WhcNMTkwOTI4MDQ1MTM2WjAoMQ8wDQYDVQQK
EwZnb29nbGUxFTATBgNVBAMTDGFwcHNjb2RlLmNvbTCCASIwDQYJKoZIhvcNAQEB
BQADggEPADCCAQoCggEBAL82r5d8aZPANPDKts43S1+8G5eq58vHTydoD08GzIvq
7aPqvO9/urynw7Ezf6flq+6+6Gl79sXeQ3dFlWkM5Cd8gIOVYa4U/WAlTnkPDQ+P
hUIzu4XL1yM1coaYV9kRbNAEr0/G8QLbch2cmj8ioWdeVwJzsmTcN/dfHEG2U9U8
LSl6O3YSbnchQA79OZCBjQ9aG5q5YflZWziLlAUs3kuowmhYxIyv5eKZ6jb0dhit
o5hE48iWWle1pAP88dZG61IXoERJ7QrgNXH5fG4nqWnXwrMOUQmtAyuA/c2DFsvV
GrtEgBMXnwrsL9ofFLTgluG7y7a7hokdbcibtgsjaHcCAwEAAaMnMCUwDgYDVR0P
AQH/BAQDAgWgMBMGA1UdJQQMMAoGCCsGAQUFBwMCMA0GCSqGSIb3DQEBCwUAA4IB
AQBNCahbHcO7Pdu8s/gIgn5cB4nWc3813jzVMDo0ujjVB1jl16pOb3vtzeTxoMJ4
ewB6C0EArTdjVK9d8PJuDL2cJwrdIuYaFzjwTpFOIWX89/p3XE2yRRMETLMccYBJ
PYskPkDz6TidYflX/H7KA9qsv+4N1KoB7PUIG4sHeVNFIN0xXZvzEXH5fUjPdpv5
W195cVunLFIlEVfJvYmMuKgGfLTj96t7GUTJUOjJtW2GWW8QI43L6BQZcCfSIdSI
YatctDlrGk9IQeKwea8u4LlRrX9eHBNDKTpxmxsiBuBWxwSkK3eyVC7PKUzerBj6
vZvzz7lCsjRshgwyDcgM5O+m
-----END CERTIFICATE-----
`),
		},
	})

	token, err := au.Login()
	if assert.Nil(t, err) {
		fmt.Println(token)
	}
}
