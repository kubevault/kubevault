/*
Copyright AppsCode Inc. and Contributors

Licensed under the AppsCode Community License 1.0.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    https://github.com/appscode/licenses/raw/1.0.0/AppsCode-Community-1.0.0.md

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package azure

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	api "kubevault.dev/apimachinery/apis/engine/v1alpha1"

	"github.com/gorilla/mux"
	vaultapi "github.com/hashicorp/vault/api"
	core "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	kfake "k8s.io/client-go/kubernetes/fake"
	"k8s.io/klog/v2"
)

func setupVaultServer() *httptest.Server {
	router := mux.NewRouter()

	router.HandleFunc("/v1/azure/roles/k8s.-.demo.demo-role", func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		var data interface{}
		err := json.NewDecoder(r.Body).Decode(&data)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_, err := w.Write([]byte(err.Error()))
			klog.Infoln(err)
			return
		} else {
			m := data.(map[string]interface{})
			value1, ok1 := m["azure_roles"]
			value2, ok2 := m["application_object_id"]
			if (!ok1 || len(value1.(string)) == 0) && (!ok2 || len(value2.(string)) == 0) {
				w.WriteHeader(http.StatusBadRequest)
				_, err := w.Write([]byte("both azure_roles and application_object_id are missing"))
				klog.Infoln(err)
				return
			}
			w.WriteHeader(http.StatusOK)
		}

	}).Methods(http.MethodPost)

	router.HandleFunc("/v1/my-azure-path/roles/k8s.-.demo.demo-role", func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		var data interface{}
		err := json.NewDecoder(r.Body).Decode(&data)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_, err := w.Write([]byte(err.Error()))
			klog.Infoln(err)
			return
		} else {
			m := data.(map[string]interface{})
			value1, ok1 := m["azure_roles"]
			value2, ok2 := m["application_object_id"]
			if (!ok1 || len(value1.(string)) == 0) && (!ok2 || len(value2.(string)) == 0) {
				w.WriteHeader(http.StatusBadRequest)
				_, err := w.Write([]byte("both azure_roles and application_object_id are missing"))
				klog.Infoln(err)
				return
			}
			w.WriteHeader(http.StatusOK)
		}

	}).Methods(http.MethodPost)

	router.HandleFunc("/v1/azure/roles/k8s.-.demo.demo-role", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}).Methods(http.MethodDelete)

	return httptest.NewServer(router)
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
		klog.Infoln("Failed to create vault client!")
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
		klog.Infoln("Failed to create vault client!")
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
			if _, err := a.DeleteRole(tt.args.name); (err != nil) != tt.wantErr {
				t.Errorf("AzureRole.DeleteRole() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
