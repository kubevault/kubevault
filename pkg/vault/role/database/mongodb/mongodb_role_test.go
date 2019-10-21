package mongodb

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gorilla/mux"
	vaultapi "github.com/hashicorp/vault/api"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	kfake "k8s.io/client-go/kubernetes/fake"
	appcat "kmodules.xyz/custom-resources/apis/appcatalog/v1alpha1"
	api "kubevault.dev/operator/apis/engine/v1alpha1"
)

func setupVaultServer() *httptest.Server {
	router := mux.NewRouter()

	router.HandleFunc("/v1/database/roles/k8s.-.m.m-read", func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		var data interface{}
		err := json.NewDecoder(r.Body).Decode(&data)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_, err := w.Write([]byte(err.Error()))
			utilruntime.Must(err)
			return
		} else {
			m := data.(map[string]interface{})
			if v, ok := m["db_name"]; !ok || len(v.(string)) == 0 {
				w.WriteHeader(http.StatusBadRequest)
				_, err := w.Write([]byte("db_name doesn't provided"))
				utilruntime.Must(err)
				return
			}
			w.WriteHeader(http.StatusOK)
		}
	}).Methods(http.MethodPost)

	router.HandleFunc("/v1/database/roles/k8s.-.m.m-read", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}).Methods(http.MethodDelete)

	router.HandleFunc("/v1/database/roles/error", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, err := w.Write([]byte("error"))
		utilruntime.Must(err)
	}).Methods(http.MethodDelete)

	return httptest.NewServer(router)
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
						DatabaseRef: &appcat.AppReference{
							Name:      "mongodb",
							Namespace: "demo",
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
						DatabaseRef: &appcat.AppReference{
							Namespace: "demo",
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

func TestNewMongoDBRoleBindingCreatRole(t *testing.T) {
	addr := os.Getenv("VAULT_ADDR")
	token := os.Getenv("VAULT_TOKEN")
	if addr == "" || token == "" {
		t.Skip()
	}

	cfg := vaultapi.DefaultConfig()
	utilruntime.Must(cfg.ConfigureTLS(&vaultapi.TLSConfig{
		Insecure: true,
	}))

	v, _ := vaultapi.NewClient(cfg)

	if !assert.NotNil(t, v) {
		return
	}

	k := kfake.NewSimpleClientset()
	_, err := k.CoreV1().Secrets("default").Create(&corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Name:      "cred",
		},
		Data: map[string][]byte{
			"username": []byte("root"),
			"password": []byte("root"),
		},
	})
	utilruntime.Must(err)

	utilruntime.Must(v.SetAddress(addr))
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

	err = m.CreateRole()
	assert.Nil(t, err)
}
