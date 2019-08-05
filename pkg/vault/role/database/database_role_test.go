package database

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	vaultapi "github.com/hashicorp/vault/api"
	"github.com/stretchr/testify/assert"
)

func setupVaultServer() *httptest.Server {
	router := mux.NewRouter()

	router.HandleFunc("/v1/database/roles/read", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}).Methods(http.MethodDelete)

	router.HandleFunc("/v1/database/roles/error", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("error"))
	}).Methods(http.MethodDelete)

	return httptest.NewServer(router)
}

func TestDeleteRole(t *testing.T) {
	srv := setupVaultServer()
	defer srv.Close()

	cfg := vaultapi.DefaultConfig()
	cfg.Address = srv.URL

	cl, err := vaultapi.NewClient(cfg)
	if !assert.Nil(t, err, "failed to create vault client") {
		return
	}

	testData := []struct {
		testName    string
		dbRole      *DatabaseRole
		roleName    string
		expectedErr bool
	}{
		{
			testName: "Delete Role successful",
			dbRole: &DatabaseRole{
				path:        "database",
				vaultClient: cl,
			},
			roleName:    "read",
			expectedErr: false,
		},
		{
			testName: "Delete Role failed",
			dbRole: &DatabaseRole{
				path:        "database",
				vaultClient: cl,
			},
			roleName:    "error",
			expectedErr: true,
		},
	}

	for _, test := range testData {
		t.Run(test.testName, func(t *testing.T) {
			p := test.dbRole

			err := p.DeleteRole(test.roleName)
			if test.expectedErr {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}
