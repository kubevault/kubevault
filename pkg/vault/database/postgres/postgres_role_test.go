package postgres

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/appscode/pat"
	vaultapi "github.com/hashicorp/vault/api"
	api "github.com/kubedb/apimachinery/apis/authorization/v1alpha1"
	configapi "github.com/kubedb/apimachinery/apis/config/v1alpha1"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kfake "k8s.io/client-go/kubernetes/fake"
)

func setupVaultServer() *httptest.Server {
	m := pat.New()

	m.Post("/v1/database/config/postgres", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		var data interface{}
		err := json.NewDecoder(r.Body).Decode(&data)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		} else {
			m := data.(map[string]interface{})
			if v, ok := m["plugin_name"]; !ok || len(v.(string)) == 0 {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte("plugin_name doesn't provided"))
				return
			}
			if v, ok := m["allowed_roles"]; !ok || len(v.(string)) == 0 {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte("allowed_roles doesn't provided"))
				return
			}
			if v, ok := m["connection_url"]; !ok || len(v.(string)) == 0 {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte("connection_url doesn't provided"))
				return
			}
			if v, ok := m["username"]; !ok || len(v.(string)) == 0 {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte("username doesn't provided"))
				return
			}
			if v, ok := m["password"]; !ok || len(v.(string)) == 0 {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte("username doesn't provided"))
				return
			}

			w.WriteHeader(http.StatusOK)
		}
	}))
	m.Post("/v1/database/roles/k8s.-.pg.pg-read", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		var data interface{}
		err := json.NewDecoder(r.Body).Decode(&data)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		} else {
			m := data.(map[string]interface{})
			if v, ok := m["db_name"]; !ok || len(v.(string)) == 0 {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte("db_name doesn't provided"))
				return
			}
			w.WriteHeader(http.StatusOK)
		}
	}))
	m.Del("/v1/database/roles/k8s.-.pg.pg-read", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	m.Del("/v1/database/roles/error", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("error"))
	}))

	return httptest.NewServer(m)
}

func TestPostgresRole_CreateConfig(t *testing.T) {
	srv := setupVaultServer()
	defer srv.Close()

	cfg := vaultapi.DefaultConfig()
	cfg.Address = srv.URL

	cl, err := vaultapi.NewClient(cfg)
	if !assert.Nil(t, err, "failed to create vault client") {
		return
	}

	pg := &PostgresRole{
		pgRole: &api.PostgresRole{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "pg-role",
				Namespace: "pg",
			},
			Spec: api.PostgresRoleSpec{
				DatabaseRef: &corev1.LocalObjectReference{
					Name: "postgres",
				},
			},
		},
		vaultClient:  cl,
		databasePath: "database",
		dbConnUrl:    "hi.com",
		config: &configapi.PostgresConfiguration{
			AllowedRoles: "*",
			PluginName:   "mongo",
		},
		secret: &corev1.Secret{
			Data: map[string][]byte{
				"username": []byte("foo"),
				"password": []byte("bar"),
			},
		},
	}

	testData := []struct {
		testName               string
		pgClient               *PostgresRole
		createCredentialSecret bool
		expectedErr            bool
	}{
		{
			testName:               "Create Config successful",
			pgClient:               pg,
			createCredentialSecret: true,
			expectedErr:            false,
		},
		{
			testName: "Create Config failed, connection_url not provided",
			pgClient: func() *PostgresRole {
				p := &PostgresRole{
					pgRole: &api.PostgresRole{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "pg-role",
							Namespace: "pg",
						},
						Spec: api.PostgresRoleSpec{
							DatabaseRef: &corev1.LocalObjectReference{
								Name: "postgres",
							},
						},
					},
					vaultClient: cl,
				}
				return p
			}(),
			createCredentialSecret: true,
			expectedErr:            true,
		},
	}

	for _, test := range testData {
		t.Run(test.testName, func(t *testing.T) {
			p := test.pgClient
			p.kubeClient = kfake.NewSimpleClientset()

			err := p.CreateConfig()
			if test.expectedErr {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}
		})
	}
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
						DatabaseRef: &corev1.LocalObjectReference{
							Name: "postgres",
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
						DatabaseRef: &corev1.LocalObjectReference{
							Name: "",
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
