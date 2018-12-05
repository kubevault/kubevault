package mongodb

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
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

	m.Post("/v1/database/config/mongodb", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
	m.Post("/v1/database/roles/k8s.-.m.m-read", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
	m.Del("/v1/database/roles/k8s.-.m.m-read", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	m.Del("/v1/database/roles/error", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("error"))
	}))

	return httptest.NewServer(m)
}

func TestMongoDBRole_CreateConfig(t *testing.T) {
	srv := setupVaultServer()
	defer srv.Close()

	cfg := vaultapi.DefaultConfig()
	cfg.Address = srv.URL

	cl, err := vaultapi.NewClient(cfg)
	if !assert.Nil(t, err, "failed to create vault client") {
		return
	}

	mySql := &MongoDBRole{
		mdbRole: &api.MongoDBRole{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "m-role",
				Namespace: "m",
			},
			Spec: api.MongoDBRoleSpec{
				DatabaseRef: &corev1.LocalObjectReference{
					Name: "mongodb",
				},
			},
		},
		vaultClient:  cl,
		databasePath: "database",
		dbConnUrl:    "hi.com",
		config: &configapi.MongoDBConfiguration{
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
		mClient                *MongoDBRole
		createCredentialSecret bool
		expectedErr            bool
	}{
		{
			testName:               "Create Config successful",
			mClient:                mySql,
			createCredentialSecret: true,
			expectedErr:            false,
		},
		{
			testName: "Create Config failed, connection_url not provided",
			mClient: func() *MongoDBRole {
				p := &MongoDBRole{
					mdbRole: &api.MongoDBRole{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "m-role",
							Namespace: "m",
						},
						Spec: api.MongoDBRoleSpec{
							DatabaseRef: &corev1.LocalObjectReference{
								Name: "mongodb",
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
			m := test.mClient
			m.kubeClient = kfake.NewSimpleClientset()

			err := m.CreateConfig()
			if test.expectedErr {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func TestMongoDBRole_CreateRole(t *testing.T) {
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
		mClient     *MongoDBRole
		expectedErr bool
	}{
		{
			testName: "Create Role successful",
			mClient: &MongoDBRole{
				mdbRole: &api.MongoDBRole{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "m-read",
						Namespace: "m",
					},
					Spec: api.MongoDBRoleSpec{
						CreationStatements: []string{"create table"},
						DatabaseRef: &corev1.LocalObjectReference{
							Name: "mongodb",
						},
					},
				},
				vaultClient:  cl,
				databasePath: "database",
			},
			expectedErr: false,
		},
		{
			testName: "Create Role failed",
			mClient: &MongoDBRole{
				mdbRole: &api.MongoDBRole{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "m-read",
						Namespace: "m",
					},
					Spec: api.MongoDBRoleSpec{
						CreationStatements: []string{"create table"},
						DatabaseRef: &corev1.LocalObjectReference{
							Name: "",
						},
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
			m := test.mClient

			err := m.CreateRole()
			if test.expectedErr {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func TestNewMongoDBRoleBindingCreatConfig(t *testing.T) {
	addr := os.Getenv("VAULT_ADDR")
	token := os.Getenv("VAULT_TOKEN")
	if addr == "" || token == "" {
		t.Skip()
	}

	cfg := vaultapi.DefaultConfig()
	cfg.ConfigureTLS(&vaultapi.TLSConfig{
		Insecure: true,
	})

	v, _ := vaultapi.NewClient(cfg)

	if !assert.NotNil(t, v) {
		return
	}

	k := kfake.NewSimpleClientset()
	k.CoreV1().Secrets("default").Create(&corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Name:      "cred",
		},
		Data: map[string][]byte{
			"username": []byte("root"),
			"password": []byte("root"),
		},
	})

	v.SetAddress(addr)
	v.SetToken(token)

	m := &MongoDBRole{
		vaultClient:  v,
		kubeClient:   k,
		databasePath: "database",
		mdbRole: &api.MongoDBRole{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "default",
				Name:      "mg",
			},
			Spec: api.MongoDBRoleSpec{
				DatabaseRef: &corev1.LocalObjectReference{
					Name: "mongodb",
				},
			},
		},
	}

	err := m.CreateConfig()
	assert.Nil(t, err)
}

func TestNewMongoDBRoleBindingCreatRole(t *testing.T) {
	addr := os.Getenv("VAULT_ADDR")
	token := os.Getenv("VAULT_TOKEN")
	if addr == "" || token == "" {
		t.Skip()
	}

	cfg := vaultapi.DefaultConfig()
	cfg.ConfigureTLS(&vaultapi.TLSConfig{
		Insecure: true,
	})

	v, _ := vaultapi.NewClient(cfg)

	if !assert.NotNil(t, v) {
		return
	}

	k := kfake.NewSimpleClientset()
	k.CoreV1().Secrets("default").Create(&corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Name:      "cred",
		},
		Data: map[string][]byte{
			"username": []byte("root"),
			"password": []byte("root"),
		},
	})

	v.SetAddress(addr)
	v.SetToken(token)

	m := &MongoDBRole{
		vaultClient:  v,
		kubeClient:   k,
		databasePath: "database",
		mdbRole: &api.MongoDBRole{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "default",
				Name:      "mg",
			},
			Spec: api.MongoDBRoleSpec{
				MaxTTL:     "1h",
				DefaultTTL: "300",
				CreationStatements: []string{
					"{ \"db\": \"admin\", \"roles\": [{ \"role\": \"readWrite\" }, {\"role\": \"read\", \"db\": \"foo\"}] }",
				},
			},
		},
	}

	err := m.CreateRole()
	assert.Nil(t, err)
}
