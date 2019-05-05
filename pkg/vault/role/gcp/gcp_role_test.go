package gcp

import (
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/appscode/pat"
	vaultapi "github.com/hashicorp/vault/api"
	api "github.com/kubevault/operator/apis/engine/v1alpha1"
	"github.com/stretchr/testify/assert"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kfake "k8s.io/client-go/kubernetes/fake"
	appcat "kmodules.xyz/custom-resources/apis/appcatalog/v1alpha1"
)

func CreateDemoDB() ([]GCPRole, *httptest.Server) {
	srv := setupVaultServer()

	kubeClient := kfake.NewSimpleClientset(&core.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "gcp-cred",
			Namespace: "demo",
		},
		Data: map[string][]byte{
			api.GCPSACredentialJson: []byte(`{}`),
		},
	})

	cfg := vaultapi.DefaultConfig()
	cfg.Address = srv.URL
	cl, err := vaultapi.NewClient(cfg)
	if err != nil {
		log.Println("Failed to create vault client!")
		return nil, nil
	}

	DB := []GCPRole{
		{
			gcpRole: &api.GCPRole{
				TypeMeta: metav1.TypeMeta{},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "my-role",
					Namespace: "demo",
				},
				Spec: api.GCPRoleSpec{
					AuthManagerRef: &appcat.AppReference{
						Namespace:  "demo",
						Name:       "vault-app",
						Parameters: nil,
					},
					Config: &api.GCPConfig{
						CredentialSecret: "gcp-cred",
					},
					SecretType: "access_token",
					Project:    "ackube",
					Bindings: `resource "//cloudresourcemanager.googleapis.com/projects/ackube" {
        roles = ["roles/viewer"]
      }`,
					TokenScopes: []string{"https://www.googleapis.com/auth/cloud-platform"},
				},
			},
			vaultClient: cl,
			kubeClient:  kubeClient,
			gcpPath:     "gcp",
		},
		{
			gcpRole: &api.GCPRole{
				TypeMeta: metav1.TypeMeta{},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "my-role",
					Namespace: "demo",
				},
				Spec: api.GCPRoleSpec{
					AuthManagerRef: &appcat.AppReference{
						Namespace:  "demo",
						Name:       "vault-app",
						Parameters: nil,
					},
					SecretType: "access_token",
					Project:    "ackube",
					Bindings: `resource "//cloudresourcemanager.googleapis.com/projects/ackube" {
        roles = ["roles/viewer"]
      }`,
					TokenScopes: []string{"https://www.googleapis.com/auth/cloud-platform"},
				},
			},
			vaultClient: cl,
			kubeClient:  kubeClient,
			gcpPath:     "gcp",
		},
		{
			gcpRole: &api.GCPRole{
				TypeMeta: metav1.TypeMeta{},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "my-role",
					Namespace: "demo",
				},
				Spec: api.GCPRoleSpec{
					AuthManagerRef: &appcat.AppReference{
						Namespace:  "demo",
						Name:       "vault-app",
						Parameters: nil,
					},
					Config: &api.GCPConfig{
						CredentialSecret: "gcp-cred",
					},
					SecretType:  "access_token",
					Project:     "ackube",
					TokenScopes: []string{"https://www.googleapis.com/auth/cloud-platform"},
				},
			},
			vaultClient: cl,
			kubeClient:  kubeClient,
			gcpPath:     "gcp",
		},
		{
			gcpRole: &api.GCPRole{
				TypeMeta: metav1.TypeMeta{},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "my-role",
					Namespace: "demo",
				},
				Spec: api.GCPRoleSpec{
					AuthManagerRef: &appcat.AppReference{
						Namespace:  "demo",
						Name:       "vault-app",
						Parameters: nil,
					},
					Config: &api.GCPConfig{
						CredentialSecret: "gcp-cred",
					},
					SecretType: "access_token",
					Project:    "ackube",
					Bindings: `resource "//cloudresourcemanager.googleapis.com/projects/ackube" {
        roles = ["roles/viewer"]
      }`,
					TokenScopes: []string{"https://www.googleapis.com/auth/cloud-platform"},
				},
			},
			vaultClient: cl,
			kubeClient:  kubeClient,
			gcpPath:     "",
		},
	}

	return DB, srv
}

func setupVaultServer() *httptest.Server {
	m := pat.New()

	m.Post("/v1/gcp/config", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		var data interface{}
		err := json.NewDecoder(r.Body).Decode(&data)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		} else {
			m := data.(map[string]interface{})
			if v, ok := m["credentials"]; !ok || len(v.(string)) == 0 {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte("credentials aren't provided"))
				return
			}
			w.WriteHeader(http.StatusOK)
		}
	}))

	m.Post("/v1/gcp/roleset/k8s.-.demo.my-role", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		var data interface{}
		err := json.NewDecoder(r.Body).Decode(&data)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		} else {
			m := data.(map[string]interface{})
			if v, ok := m["secret_type"]; !ok || len(v.(string)) == 0 {
				w.WriteHeader(http.StatusNotFound)
				w.Write([]byte("secret_type isn't specified!"))
				return
			}

			if v, ok := m["project"]; !ok || len(v.(string)) == 0 {
				w.WriteHeader(http.StatusNotFound)
				w.Write([]byte("project name isn't specified!"))
				return
			}

			if v, ok := m["bindings"]; !ok || len(v.(string)) == 0 {
				w.WriteHeader(http.StatusNotFound)
				w.Write([]byte("bindings aren't specified!"))
				return
			}

			if v, ok := m["token_scopes"]; !ok || v == nil {
				w.WriteHeader(http.StatusNotFound)
				w.Write([]byte("token_scopes aren't specified!"))
				return
			}
			w.WriteHeader(http.StatusOK)
		}

	}))

	m.Del("/v1/gcp/roleset/k8s.-.demo.my-role", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	return httptest.NewServer(m)
}

func TestGCPRole_CreateConfig(t *testing.T) {
	demoRole, srv := CreateDemoDB()
	defer srv.Close()

	testData := []struct {
		testName    string
		gcpRole     *GCPRole
		expectedErr bool
	}{
		{
			testName:    "Create Config Successful",
			gcpRole:     &demoRole[0],
			expectedErr: false,
		},
		{
			testName:    "Create Config Failed",
			gcpRole:     &demoRole[1],
			expectedErr: true,
		},
		{
			testName:    "Create Config Successful",
			gcpRole:     &demoRole[2],
			expectedErr: false,
		},
		{
			testName:    "Create Config Failed",
			gcpRole:     &demoRole[3],
			expectedErr: true,
		},
	}

	for _, test := range testData {
		t.Run(test.testName, func(t *testing.T) {
			m := test.gcpRole
			err := m.CreateConfig()
			if test.expectedErr {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func TestGCPRole_CreateRole(t *testing.T) {
	demoRole, srv := CreateDemoDB()
	defer srv.Close()

	testData := []struct {
		testName    string
		gcpRole     *GCPRole
		expectedErr bool
	}{
		{
			testName:    "Create Role Successful",
			gcpRole:     &demoRole[0],
			expectedErr: false,
		},
		{
			testName:    "Create Role Successful",
			gcpRole:     &demoRole[1],
			expectedErr: false,
		},
		{
			testName:    "Create Role Failed",
			gcpRole:     &demoRole[2],
			expectedErr: true,
		},
		{
			testName:    "Create Role Failed",
			gcpRole:     &demoRole[3],
			expectedErr: true,
		},
	}

	for _, test := range testData {
		t.Run(test.testName, func(t *testing.T) {
			m := test.gcpRole
			err := m.CreateRole()
			if test.expectedErr {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func TestGCPRole_DeleteRole(t *testing.T) {
	demoRole, srv := CreateDemoDB()
	defer srv.Close()

	testData := []struct {
		testName    string
		gcpRole     *GCPRole
		expectedErr bool
	}{
		{
			testName:    "Delete Role Successful",
			gcpRole:     &demoRole[0],
			expectedErr: false,
		},
	}

	for _, test := range testData {
		t.Run(test.testName, func(t *testing.T) {
			m := test.gcpRole
			err := m.DeleteRole("k8s.-.demo.my-role")
			if test.expectedErr {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}
		})
	}

}
