package azure

import (
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/appscode/pat"
	vaultapi "github.com/hashicorp/vault/api"
	api "github.com/kubevault/operator/apis/engine/v1alpha1"
	core "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	kfake "k8s.io/client-go/kubernetes/fake"
)

func setupVaultServer() *httptest.Server {
	m := pat.New()

	m.Post("/v1/azure/config", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		var data interface{}
		err := json.NewDecoder(r.Body).Decode(&data)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_, err := w.Write([]byte(err.Error()))
			log.Println(err)
			return
		} else {
			m := data.(map[string]interface{})
			if v, ok := m["subscription_id"]; !ok || len(v.(string)) == 0 {
				w.WriteHeader(http.StatusBadRequest)
				_, err := w.Write([]byte("subscription id isn't provided"))
				log.Println(err)
				return
			}

			if v, ok := m["tenant_id"]; !ok || len(v.(string)) == 0 {
				w.WriteHeader(http.StatusBadRequest)
				_, err := w.Write([]byte("tenant id isn't provided"))
				log.Println(err)
				return
			}
			w.WriteHeader(http.StatusOK)
		}
	}))

	m.Post("/v1/my-azure-path/config", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		var data interface{}
		err := json.NewDecoder(r.Body).Decode(&data)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_, err := w.Write([]byte(err.Error()))
			log.Println(err)
			return
		} else {
			m := data.(map[string]interface{})
			if v, ok := m["subscription_id"]; !ok || len(v.(string)) == 0 {
				w.WriteHeader(http.StatusBadRequest)
				_, err := w.Write([]byte("subscription id isn't provided"))
				log.Println(err)
				return
			}

			if v, ok := m["tenant_id"]; !ok || len(v.(string)) == 0 {
				w.WriteHeader(http.StatusBadRequest)
				_, err := w.Write([]byte("tenant id isn't provided"))
				log.Println(err)
				return
			}
			w.WriteHeader(http.StatusOK)
		}
	}))

	m.Post("/v1/azure/roles/k8s.-.demo.demo-role", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		var data interface{}
		err := json.NewDecoder(r.Body).Decode(&data)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_, err := w.Write([]byte(err.Error()))
			log.Println(err)
			return
		} else {
			m := data.(map[string]interface{})
			value1, ok1 := m["azure_roles"]
			value2, ok2 := m["application_object_id"]
			if (!ok1 || len(value1.(string)) == 0) && (!ok2 || len(value2.(string)) == 0) {
				w.WriteHeader(http.StatusBadRequest)
				_, err := w.Write([]byte("both azure_roles and application_object_id are missing"))
				log.Println(err)
				return
			}
			w.WriteHeader(http.StatusOK)
		}

	}))

	m.Post("/v1/my-azure-path/roles/k8s.-.demo.demo-role", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		var data interface{}
		err := json.NewDecoder(r.Body).Decode(&data)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_, err := w.Write([]byte(err.Error()))
			log.Println(err)
			return
		} else {
			m := data.(map[string]interface{})
			value1, ok1 := m["azure_roles"]
			value2, ok2 := m["application_object_id"]
			if (!ok1 || len(value1.(string)) == 0) && (!ok2 || len(value2.(string)) == 0) {
				w.WriteHeader(http.StatusBadRequest)
				_, err := w.Write([]byte("both azure_roles and application_object_id are missing"))
				log.Println(err)
				return
			}
			w.WriteHeader(http.StatusOK)
		}

	}))

	m.Del("/v1/azure/roles/k8s.-.demo.demo-role", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	return httptest.NewServer(m)
}

func TestAzureRole_CreateConfig(t *testing.T) {
	srv := setupVaultServer()
	defer srv.Close()

	fkube := kfake.NewSimpleClientset(
		&core.Secret{
			ObjectMeta: v1.ObjectMeta{
				Name:      "azure-cred1",
				Namespace: "demo",
			},
			Data: map[string][]byte{
				api.AzureClientSecret:   []byte("******"),
				api.AzureSubscriptionID: []byte("******"),
				api.AzureTenantID:       []byte("*****"),
				api.AzureClientID:       []byte("******"),
			},
			Type: "",
		},
		&core.Secret{
			ObjectMeta: v1.ObjectMeta{
				Name:      "azure-cred2",
				Namespace: "demo",
			},
			Data: map[string][]byte{
				api.AzureClientSecret: []byte("******"),
				api.AzureTenantID:     []byte("*****"),
				api.AzureClientID:     []byte("******"),
			},
			Type: "",
		},
		&core.Secret{
			ObjectMeta: v1.ObjectMeta{
				Name:      "azure-cred3",
				Namespace: "demo",
			},
			Data: map[string][]byte{
				api.AzureClientSecret:   []byte("******"),
				api.AzureSubscriptionID: []byte("******"),
				api.AzureClientID:       []byte("******"),
			},
			Type: "",
		})

	cfg := vaultapi.DefaultConfig()
	cfg.Address = srv.URL
	cl, err := vaultapi.NewClient(cfg)
	if err != nil {
		log.Println("Failed to create vault client!")
		t.Skip()
	}

	type fields struct {
		azureRole   *api.AzureRole
		vaultClient *vaultapi.Client
		kubeClient  kubernetes.Interface
		azurePath   string
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "Successful operation!",
			fields: fields{
				azureRole: &api.AzureRole{
					ObjectMeta: v1.ObjectMeta{
						Name:      "demo-role",
						Namespace: "demo",
					},
					Spec: api.AzureRoleSpec{
						Config: &api.AzureConfig{
							CredentialSecret: "azure-cred1",
							Environment:      "",
						},
					},
				},
				vaultClient: cl,
				kubeClient:  fkube,
				azurePath:   "azure",
			},
			wantErr: false,
		},
		{
			name: "Successful operation! User defined path!",
			fields: fields{
				azureRole: &api.AzureRole{
					ObjectMeta: v1.ObjectMeta{
						Name:      "demo-role",
						Namespace: "demo",
					},
					Spec: api.AzureRoleSpec{
						Config: &api.AzureConfig{
							CredentialSecret: "azure-cred1",
							Environment:      "",
						},
					},
				},
				vaultClient: cl,
				kubeClient:  fkube,
				azurePath:   "my-azure-path",
			},
			wantErr: false,
		},
		{
			name: "Unsuccessful operation! Missing SubscriptionID!",
			fields: fields{
				azureRole: &api.AzureRole{
					ObjectMeta: v1.ObjectMeta{
						Name:      "demo-role",
						Namespace: "demo",
					},
					Spec: api.AzureRoleSpec{
						Config: &api.AzureConfig{
							CredentialSecret: "azure-cred2",
							Environment:      "",
						},
					},
				},
				vaultClient: cl,
				kubeClient:  fkube,
				azurePath:   "azure",
			},
			wantErr: true,
		},
		{
			name: "Unsuccessful operation! Missing Tenant ID!",
			fields: fields{
				azureRole: &api.AzureRole{
					ObjectMeta: v1.ObjectMeta{
						Name:      "demo-role",
						Namespace: "demo",
					},
					Spec: api.AzureRoleSpec{
						Config: &api.AzureConfig{
							CredentialSecret: "azure-cred3",
							Environment:      "",
						},
					},
				},
				vaultClient: cl,
				kubeClient:  fkube,
				azurePath:   "azure",
			},
			wantErr: true,
		},
		{
			name: "Unsuccessful operation! Missing Credential secret!",
			fields: fields{
				azureRole: &api.AzureRole{
					ObjectMeta: v1.ObjectMeta{
						Name:      "demo-role",
						Namespace: "demo",
					},
					Spec: api.AzureRoleSpec{
						Config: &api.AzureConfig{
							CredentialSecret: "unknown-cred",
							Environment:      "",
						},
					},
				},
				vaultClient: cl,
				kubeClient:  fkube,
				azurePath:   "azure",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &AzureRole{
				azureRole:   tt.fields.azureRole,
				vaultClient: tt.fields.vaultClient,
				kubeClient:  tt.fields.kubeClient,
				azurePath:   tt.fields.azurePath,
			}
			if err := a.CreateConfig(); (err != nil) != tt.wantErr {
				t.Errorf("AzureRole.CreateConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAzureRole_CreateRole(t *testing.T) {
	srv := setupVaultServer()
	defer srv.Close()

	fkube := kfake.NewSimpleClientset(&core.Secret{
		ObjectMeta: v1.ObjectMeta{
			Name:      "azure-cred",
			Namespace: "demo",
		},
		Data: map[string][]byte{
			api.AzureClientSecret: []byte("******"),
		},
		Type: "",
	})

	cfg := vaultapi.DefaultConfig()
	cfg.Address = srv.URL
	cl, err := vaultapi.NewClient(cfg)
	if err != nil {
		log.Println("Failed to create vault client!")
		t.Skip()
	}

	type fields struct {
		azureRole   *api.AzureRole
		vaultClient *vaultapi.Client
		kubeClient  kubernetes.Interface
		azurePath   string
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "Successful Operation!",
			fields: fields{
				azureRole: &api.AzureRole{
					ObjectMeta: v1.ObjectMeta{
						Name:      "demo-role",
						Namespace: "demo",
					},
					Spec: api.AzureRoleSpec{
						AzureRoles:          "[{}]",
						ApplicationObjectID: "3454-435-435-34",
						TTL:                 "0h",
						MaxTTL:              "0h",
					},
				},
				vaultClient: cl,
				kubeClient:  fkube,
				azurePath:   "azure",
			},
			wantErr: false,
		},
		{
			name: "Successful Operation! User defined path!",
			fields: fields{
				azureRole: &api.AzureRole{
					ObjectMeta: v1.ObjectMeta{
						Name:      "demo-role",
						Namespace: "demo",
					},
					Spec: api.AzureRoleSpec{
						AzureRoles:          "[{}]",
						ApplicationObjectID: "3454-435-435-34",
						TTL:                 "0h",
						MaxTTL:              "0h",
					},
				},
				vaultClient: cl,
				kubeClient:  fkube,
				azurePath:   "my-azure-path",
			},
			wantErr: false,
		},
		{
			name: "Successful Operation! ApplicationObjectID missing!",
			fields: fields{
				azureRole: &api.AzureRole{
					ObjectMeta: v1.ObjectMeta{
						Name:      "demo-role",
						Namespace: "demo",
					},
					Spec: api.AzureRoleSpec{
						AzureRoles:          "[{}]",
						ApplicationObjectID: "",
						TTL:                 "0h",
						MaxTTL:              "0h",
					},
				},
				vaultClient: cl,
				kubeClient:  fkube,
				azurePath:   "azure",
			},
			wantErr: false,
		},
		{
			name: "Successful Operation! AzureRoles missing!",
			fields: fields{
				azureRole: &api.AzureRole{
					ObjectMeta: v1.ObjectMeta{
						Name:      "demo-role",
						Namespace: "demo",
					},
					Spec: api.AzureRoleSpec{
						AzureRoles:          "",
						ApplicationObjectID: "3454-435-435-34",
						TTL:                 "0h",
						MaxTTL:              "0h",
					},
				},
				vaultClient: cl,
				kubeClient:  fkube,
				azurePath:   "azure",
			},
			wantErr: false,
		},
		{
			name: "Unsuccessful Operation! Both AzureRoles and ApplicationObjectID missing!",
			fields: fields{
				azureRole: &api.AzureRole{
					ObjectMeta: v1.ObjectMeta{
						Name:      "demo-role",
						Namespace: "demo",
					},
					Spec: api.AzureRoleSpec{
						AzureRoles:          "",
						ApplicationObjectID: "",
						TTL:                 "0h",
						MaxTTL:              "0h",
					},
				},
				vaultClient: cl,
				kubeClient:  fkube,
				azurePath:   "azure",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &AzureRole{
				azureRole:   tt.fields.azureRole,
				vaultClient: tt.fields.vaultClient,
				kubeClient:  tt.fields.kubeClient,
				azurePath:   tt.fields.azurePath,
			}
			if err := a.CreateRole(); (err != nil) != tt.wantErr {
				t.Errorf("AzureRole.CreateRole() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAzureRole_DeleteRole(t *testing.T) {
	srv := setupVaultServer()
	defer srv.Close()

	cfg := vaultapi.DefaultConfig()
	cfg.Address = srv.URL
	cl, err := vaultapi.NewClient(cfg)
	if err != nil {
		log.Println("Failed to create vault client!")
		t.Skip()
	}

	type fields struct {
		azureRole   *api.AzureRole
		vaultClient *vaultapi.Client
		kubeClient  kubernetes.Interface
		azurePath   string
	}
	type args struct {
		name string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "Successful Operation!",
			fields: fields{
				azurePath:   "azure",
				vaultClient: cl,
			},
			args: args{
				name: "k8s.-.demo.demo-role",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &AzureRole{
				azureRole:   tt.fields.azureRole,
				vaultClient: tt.fields.vaultClient,
				kubeClient:  tt.fields.kubeClient,
				azurePath:   tt.fields.azurePath,
			}
			if err := a.DeleteRole(tt.args.name); (err != nil) != tt.wantErr {
				t.Errorf("AzureRole.DeleteRole() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
