package postgres

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	vaultapi "github.com/hashicorp/vault/api"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	appcat "kmodules.xyz/custom-resources/apis/appcatalog/v1alpha1"
	api "kubevault.dev/operator/apis/engine/v1alpha1"
	"kubevault.dev/operator/pkg/vault/util"
)

func setupVaultServer() *httptest.Server {
	router := mux.NewRouter()

	router.HandleFunc("/v1/database/roles/k8s.-.pg.pg-read", func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		var data interface{}
		err := json.NewDecoder(r.Body).Decode(&data)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			util.LogWriteErr(w.Write([]byte(err.Error())))
			return
		} else {
			m := data.(map[string]interface{})
			if v, ok := m["db_name"]; !ok || len(v.(string)) == 0 {
				w.WriteHeader(http.StatusBadRequest)
				util.LogWriteErr(w.Write([]byte("db_name doesn't provided")))
				return
			}
			w.WriteHeader(http.StatusOK)
		}
	}).Methods(http.MethodPost)

	router.HandleFunc("/v1/database/roles/k8s.-.pg.pg-read", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}).Methods(http.MethodDelete)

	router.HandleFunc("/v1/database/roles/error", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		util.LogWriteErr(w.Write([]byte("error")))
	}).Methods(http.MethodDelete)

	return httptest.NewServer(router)
}

func TestPostgresRole_CreateRole(t *testing.T) {
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
		pgClient    *PostgresRole
		expectedErr bool
	}{
		{
			testName: "Create Role successful",
			pgClient: &PostgresRole{
				pgRole: &api.PostgresRole{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "pg-read",
						Namespace: "pg",
					},
					Spec: api.PostgresRoleSpec{
						DatabaseRef: &appcat.AppReference{
							Name:      "postgres",
							Namespace: "demo",
						},
						CreationStatements: []string{"create table"},
					},
				},
				vaultClient:  cl,
				databasePath: "database",
			},
			expectedErr: false,
		},
		{
			testName: "Create Role failed",
			pgClient: &PostgresRole{
				pgRole: &api.PostgresRole{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "pg-read",
						Namespace: "pg",
					},
					Spec: api.PostgresRoleSpec{
						DatabaseRef: &appcat.AppReference{
							Namespace: "demo",
						},
						CreationStatements: []string{"create table"},
					},
				},
				vaultClient:  cl,
				databasePath: "database",
			},
			expectedErr: true,
		},
	}

	for _, test := range testData {
		t.Run(test.testName, func(t *testing.T) {
			p := test.pgClient

			err := p.CreateRole()
			if test.expectedErr {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}
