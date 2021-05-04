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

package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	api "kubevault.dev/apimachinery/apis/engine/v1alpha1"

	"github.com/gorilla/mux"
	vaultapi "github.com/hashicorp/vault/api"
	"github.com/stretchr/testify/assert"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	kfake "k8s.io/client-go/kubernetes/fake"
	appcatfake "kmodules.xyz/custom-resources/client/clientset/versioned/typed/appcatalog/v1alpha1/fake"
)

const Token = "root"
const ExpectBlank = "{{{BLANK}}}"

func vaultClient(addr string) (*vaultapi.Client, error) {
	cfg := vaultapi.DefaultConfig()
	err := cfg.ConfigureTLS(&vaultapi.TLSConfig{
		Insecure: true,
	})
	if err != nil {
		return nil, err
	}
	c, err := vaultapi.NewClient(cfg)
	if err != nil {
		return nil, err
	}

	c.SetToken(Token)
	err = c.SetAddress(addr)
	if err != nil {
		return nil, err
	}
	return c, nil
}

func NewFakeVaultServer() *httptest.Server {
	router := mux.NewRouter()

	router.HandleFunc("/v1/gcp/config", func(w http.ResponseWriter, r *http.Request) {
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
			if v, ok := m["credentials"]; !ok || len(v.(string)) == 0 {
				w.WriteHeader(http.StatusBadRequest)
				_, err := w.Write([]byte("credentials aren't provided"))
				utilruntime.Must(err)
				return
			}
			w.WriteHeader(http.StatusOK)
		}
	}).Methods(http.MethodPost)

	router.HandleFunc("/v1/azure/config", func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		var data interface{}
		err := json.NewDecoder(r.Body).Decode(&data)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_, err := w.Write([]byte(err.Error()))
			klog.Println(err)
			return
		} else {
			m := data.(map[string]interface{})
			if v, ok := m["subscription_id"]; !ok || len(v.(string)) == 0 {
				w.WriteHeader(http.StatusBadRequest)
				_, err := w.Write([]byte("subscription id isn't provided"))
				klog.Println(err)
				return
			}

			if v, ok := m["tenant_id"]; !ok || len(v.(string)) == 0 {
				w.WriteHeader(http.StatusBadRequest)
				_, err := w.Write([]byte("tenant id isn't provided"))
				klog.Println(err)
				return
			}
			w.WriteHeader(http.StatusOK)
		}
	}).Methods(http.MethodPost)

	router.HandleFunc("/v1/aws/config/root", func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		var data interface{}
		err := json.NewDecoder(r.Body).Decode(&data)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_, err := w.Write([]byte(err.Error()))
			klog.Println(err)
			return
		}
		m := data.(map[string]interface{})
		if v, ok := m["access_key"]; !ok || len(v.(string)) == 0 {
			w.WriteHeader(http.StatusBadRequest)
			_, err := w.Write([]byte("access_key id isn't provided"))
			klog.Println(err)
			return
		}

		if v, ok := m["secret_key"]; !ok || len(v.(string)) == 0 {
			w.WriteHeader(http.StatusBadRequest)
			_, err := w.Write([]byte("secret_key isn't provided"))
			klog.Println(err)
			return
		}
		w.WriteHeader(http.StatusOK)

	}).Methods(http.MethodPost)

	router.HandleFunc("/v1/my-azure-path/config", func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		var data interface{}
		err := json.NewDecoder(r.Body).Decode(&data)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_, err := w.Write([]byte(err.Error()))
			klog.Println(err)
			return
		} else {
			m := data.(map[string]interface{})
			if v, ok := m["subscription_id"]; !ok || len(v.(string)) == 0 {
				w.WriteHeader(http.StatusBadRequest)
				_, err := w.Write([]byte("subscription id isn't provided"))
				klog.Println(err)
				return
			}

			if v, ok := m["tenant_id"]; !ok || len(v.(string)) == 0 {
				w.WriteHeader(http.StatusBadRequest)
				_, err := w.Write([]byte("tenant id isn't provided"))
				klog.Println(err)
				return
			}
			w.WriteHeader(http.StatusOK)
		}
	}).Methods(http.MethodPost)

	router.HandleFunc("/v1/database/config/mongodb", func(w http.ResponseWriter, r *http.Request) {
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
			if v, ok := m["plugin_name"]; !ok || len(v.(string)) == 0 {
				w.WriteHeader(http.StatusBadRequest)
				_, err := w.Write([]byte("plugin_name doesn't provided"))
				utilruntime.Must(err)
				return
			}
			if v, ok := m["allowed_roles"]; !ok || len(v.(string)) == 0 {
				w.WriteHeader(http.StatusBadRequest)
				_, err := w.Write([]byte("allowed_roles doesn't provided"))
				utilruntime.Must(err)
				return
			}
			if v, ok := m["connection_url"]; !ok || len(v.(string)) == 0 {
				w.WriteHeader(http.StatusBadRequest)
				_, err := w.Write([]byte("connection_url doesn't provided"))
				utilruntime.Must(err)
				return
			}
			if v, ok := m["username"]; !ok || len(v.(string)) == 0 {
				w.WriteHeader(http.StatusBadRequest)
				_, err := w.Write([]byte("username doesn't provided"))
				utilruntime.Must(err)
				return
			}
			if v, ok := m["password"]; !ok || len(v.(string)) == 0 {
				w.WriteHeader(http.StatusBadRequest)
				_, err := w.Write([]byte("username doesn't provided"))
				utilruntime.Must(err)
				return
			}

			w.WriteHeader(http.StatusOK)
		}
	}).Methods(http.MethodPost)

	router.HandleFunc("/v1/database/config/mysql", func(w http.ResponseWriter, r *http.Request) {
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
			if v, ok := m["plugin_name"]; !ok || len(v.(string)) == 0 {
				w.WriteHeader(http.StatusBadRequest)
				_, err := w.Write([]byte("plugin_name doesn't provided"))
				utilruntime.Must(err)
				return
			}
			if v, ok := m["allowed_roles"]; !ok || len(v.(string)) == 0 {
				w.WriteHeader(http.StatusBadRequest)
				_, err := w.Write([]byte("allowed_roles doesn't provided"))
				utilruntime.Must(err)
				return
			}
			if v, ok := m["connection_url"]; !ok || len(v.(string)) == 0 {
				w.WriteHeader(http.StatusBadRequest)
				_, err := w.Write([]byte("connection_url doesn't provided"))
				utilruntime.Must(err)
				return
			}
			if v, ok := m["username"]; !ok || len(v.(string)) == 0 {
				w.WriteHeader(http.StatusBadRequest)
				_, err := w.Write([]byte("username doesn't provided"))
				utilruntime.Must(err)
				return
			}
			if v, ok := m["password"]; !ok || len(v.(string)) == 0 {
				w.WriteHeader(http.StatusBadRequest)
				_, err := w.Write([]byte("username doesn't provided"))
				utilruntime.Must(err)
				return
			}

			w.WriteHeader(http.StatusOK)
		}
	}).Methods(http.MethodPost)

	router.HandleFunc("/v1/database/config/postgres", func(w http.ResponseWriter, r *http.Request) {
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
			if v, ok := m["plugin_name"]; !ok || len(v.(string)) == 0 {
				w.WriteHeader(http.StatusBadRequest)
				_, err := w.Write([]byte("plugin_name doesn't provided"))
				utilruntime.Must(err)
				return
			}
			if v, ok := m["allowed_roles"]; !ok || len(v.(string)) == 0 {
				w.WriteHeader(http.StatusBadRequest)
				_, err := w.Write([]byte("allowed_roles doesn't provided"))
				utilruntime.Must(err)
				return
			}
			if v, ok := m["connection_url"]; !ok || len(v.(string)) == 0 {
				w.WriteHeader(http.StatusBadRequest)
				_, err := w.Write([]byte("connection_url doesn't provided"))
				utilruntime.Must(err)
				return
			}
			if v, ok := m["username"]; !ok || len(v.(string)) == 0 {
				w.WriteHeader(http.StatusBadRequest)
				_, err := w.Write([]byte("username doesn't provided"))
				utilruntime.Must(err)
				return
			}
			if v, ok := m["password"]; !ok || len(v.(string)) == 0 {
				w.WriteHeader(http.StatusBadRequest)
				_, err := w.Write([]byte("username doesn't provided"))
				utilruntime.Must(err)
				return
			}

			w.WriteHeader(http.StatusOK)
		}
	}).Methods(http.MethodPost)

	router.HandleFunc(fmt.Sprintf("/v1/%s/config", DefaultKVPath), func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		var data interface{}

		fail := func(message string) {
			fail(message, w)
		}

		expectedVersion := r.Header.Get(KVTestHeaderExpectedVersion)
		if expectedVersion == "1" {
			fail("KV version 1 does not support the `config` endpoint")
			return
		}

		if expectedVersion != "2" {
			fail(fmt.Sprintf("Unknown expected KV version: %v", expectedVersion))
			return
		}

		err := json.NewDecoder(r.Body).Decode(&data)

		if err != nil {
			fail("Unable to decode request payload:")
			mustWriteString(err.Error(), w)
			return
		} else {
			m := data.(map[string]interface{})

			check := func(header, param string) bool {
				if e := r.Header.Get(header); len(e) != 0 {
					if e == ExpectBlank {
						e = ""
					}

					v, ok := m[param]

					if !ok || (e == ExpectBlank && len(v.(string)) == 0) {
						fail(fmt.Sprintf("`%s` not supplied, but expected '%v'", param, e))
						return false
					}

					if e != v {
						fail(fmt.Sprintf("incorrect or invalid `%s`: expected: '%v', got: '%v'", param, e, v))
						return false
					}
				}

				return true
			}

			if !check(KVTestHeaderExpectedMaxVersions, KVConfigMaxVersions) ||
				!check(KVTestHeaderExpectedCasRequired, KVConfigCasRequired) ||
				!check(KVTestHeaderExpectedDeleteVersionsAfter, KVConfigDeleteVersionsAfter) {
				return
			}
		}

		w.WriteHeader(http.StatusOK)
	}).Methods(http.MethodPost)

	return httptest.NewServer(router)
}

func TestSecretEngine_CreateGCPConfig(t *testing.T) {
	srv := NewFakeVaultServer()
	defer srv.Close()

	tests := []struct {
		name         string
		path         string
		secretEngine *api.SecretEngine
		wantErr      bool
	}{
		{
			name: "GCPConfig: Successful operation",
			path: "gcp",
			secretEngine: &api.SecretEngine{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test12321",
					Namespace: "demo",
				},
				Spec: api.SecretEngineSpec{
					SecretEngineConfiguration: api.SecretEngineConfiguration{
						GCP: &api.GCPConfiguration{
							CredentialSecret: "gcp-cred",
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "GCPConfig: Unsuccessful operation: Missing credentials",
			path: "gcp",
			secretEngine: &api.SecretEngine{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test12321",
					Namespace: "demo",
				},
				Spec: api.SecretEngineSpec{
					SecretEngineConfiguration: api.SecretEngineConfiguration{
						GCP: &api.GCPConfiguration{
							CredentialSecret: "gcp-cred2",
						},
					},
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			vc, err := vaultClient(srv.URL)
			assert.Nil(t, err, "failed to create vault client")

			secretEngineClient := &SecretEngine{
				appClient:    &appcatfake.FakeAppcatalogV1alpha1{},
				secretEngine: tt.secretEngine,
				vaultClient:  vc,
				kubeClient:   kfake.NewSimpleClientset(),
				path:         tt.path,
			}
			// Create fake secret for gcp config
			_, err = secretEngineClient.kubeClient.CoreV1().Secrets("demo").Create(
				context.TODO(),
				&core.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name: "gcp-cred",
					},
					Data: map[string][]byte{
						"sa.json": []byte("fakeKey"),
					},
				},
				metav1.CreateOptions{},
			)
			assert.Nil(t, err)

			if err := secretEngineClient.CreateGCPConfig(); (err != nil) != tt.wantErr {
				t.Errorf("CreateConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSecretEngine_CreateAzureConfig(t *testing.T) {
	srv := NewFakeVaultServer()
	defer srv.Close()

	tests := []struct {
		name         string
		path         string
		secretEngine *api.SecretEngine
		secret       *core.Secret
		wantErr      bool
	}{
		{
			name: "AzureConfig: Successful operation",
			path: "azure",
			secretEngine: &api.SecretEngine{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test12321",
					Namespace: "demo",
				},
				Spec: api.SecretEngineSpec{
					SecretEngineConfiguration: api.SecretEngineConfiguration{
						Azure: &api.AzureConfiguration{
							CredentialSecret: "azure-cred",
						},
					},
				},
			},
			secret: &core.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "azure-cred",
					Namespace: "demo",
				},
				Data: map[string][]byte{
					"subscription-id": []byte("1232-2132-123-132"),
					"tenant-id":       []byte("acdenfi-fkjdsk-dsfjds-fsdjf"),
				},
			},
			wantErr: false,
		},
		{
			name: "AzureConfig: Successful operation: User define path",
			path: "my-azure-path",
			secretEngine: &api.SecretEngine{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test12321",
					Namespace: "demo",
				},
				Spec: api.SecretEngineSpec{
					SecretEngineConfiguration: api.SecretEngineConfiguration{
						Azure: &api.AzureConfiguration{
							CredentialSecret: "azure-cred",
						},
					},
				},
			},
			secret: &core.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "azure-cred",
					Namespace: "demo",
				},
				Data: map[string][]byte{
					"subscription-id": []byte("1232-2132-123-132"),
					"tenant-id":       []byte("acdenfi-fkjdsk-dsfjds-fsdjf"),
				},
			},
			wantErr: false,
		},
		{
			name: "AzureConfig: Unsuccessful operation: Missing tenant-id",
			path: "azure",
			secretEngine: &api.SecretEngine{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test12321",
					Namespace: "demo",
				},
				Spec: api.SecretEngineSpec{
					SecretEngineConfiguration: api.SecretEngineConfiguration{
						Azure: &api.AzureConfiguration{
							CredentialSecret: "azure-cred",
						},
					},
				},
			},
			secret: &core.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "azure-cred",
					Namespace: "demo",
				},
				Data: map[string][]byte{
					"subscription-id": []byte("1232-2132-123-132"),
				},
			},
			wantErr: true,
		},
		{
			name: "AzureConfig: Unsuccessful operation: Missing subscription-id",
			path: "azure",
			secretEngine: &api.SecretEngine{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test12321",
					Namespace: "demo",
				},
				Spec: api.SecretEngineSpec{
					SecretEngineConfiguration: api.SecretEngineConfiguration{
						Azure: &api.AzureConfiguration{
							CredentialSecret: "azure-cred",
						},
					},
				},
			},
			secret: &core.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "azure-cred",
					Namespace: "demo",
				},
				Data: map[string][]byte{
					"tenant-id": []byte("acdenfi-fkjdsk-dsfjds-fsdjf"),
				},
			},
			wantErr: true,
		},
		{
			name: "AzureConfig: Unsuccessful operation: Missing secret",
			path: "azure",
			secretEngine: &api.SecretEngine{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test12321",
					Namespace: "demo",
				},
				Spec: api.SecretEngineSpec{
					SecretEngineConfiguration: api.SecretEngineConfiguration{
						Azure: &api.AzureConfiguration{
							CredentialSecret: "azure-cred",
						},
					},
				},
			},
			secret: &core.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "azure-cred23",
					Namespace: "demo",
				},
				Data: map[string][]byte{
					"subscription-id": []byte("1232-2132-123-132"),
					"tenant-id":       []byte("acdenfi-fkjdsk-dsfjds-fsdjf"),
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			vc, err := vaultClient(srv.URL)
			assert.Nil(t, err, "failed to create vault client")

			seClient := &SecretEngine{
				appClient:    &appcatfake.FakeAppcatalogV1alpha1{},
				secretEngine: tt.secretEngine,
				vaultClient:  vc,
				kubeClient:   kfake.NewSimpleClientset(),
				path:         tt.path,
			}
			// Create fake secret for azure config
			_, err = seClient.kubeClient.CoreV1().Secrets("demo").Create(context.TODO(), tt.secret, metav1.CreateOptions{})
			assert.Nil(t, err)
			if err := seClient.CreateAzureConfig(); (err != nil) != tt.wantErr {
				t.Errorf("CreateAzureConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSecretEngine_CreateAWSConfig(t *testing.T) {
	srv := NewFakeVaultServer()
	defer srv.Close()

	tests := []struct {
		name         string
		path         string
		secretEngine *api.SecretEngine
		secret       *core.Secret
		wantErr      bool
	}{
		{
			name: "AWSConfig: Successful operation",
			path: "aws",
			secretEngine: &api.SecretEngine{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test12321",
					Namespace: "demo",
				},
				Spec: api.SecretEngineSpec{
					SecretEngineConfiguration: api.SecretEngineConfiguration{
						AWS: &api.AWSConfiguration{
							CredentialSecret: "aws-cred",
							Region:           "asia",
						},
					},
				},
			},
			secret: &core.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "aws-cred",
					Namespace: "demo",
				},
				Data: map[string][]byte{
					"access_key": []byte("1232-2132-123-132"),
					"secret_key": []byte("acdenfi-fkjdsk-dsfjds-fsdjf"),
				},
			},
			wantErr: false,
		},
		{
			name: "AWSConfig: Unsuccessful operation: Missing access_key",
			path: "aws",
			secretEngine: &api.SecretEngine{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test12321",
					Namespace: "demo",
				},
				Spec: api.SecretEngineSpec{
					SecretEngineConfiguration: api.SecretEngineConfiguration{
						AWS: &api.AWSConfiguration{
							CredentialSecret: "aws-cred",
							Region:           "asia",
						},
					},
				},
			},
			secret: &core.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "aws-cred",
					Namespace: "demo",
				},
				Data: map[string][]byte{
					"secret_key": []byte("acdenfi-fkjdsk-dsfjds-fsdjf"),
				},
			},
			wantErr: true,
		},
		{
			name: "awsConfig: Unsuccessful operation: Missing secret_key",
			path: "aws",
			secretEngine: &api.SecretEngine{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test12321",
					Namespace: "demo",
				},
				Spec: api.SecretEngineSpec{
					SecretEngineConfiguration: api.SecretEngineConfiguration{
						AWS: &api.AWSConfiguration{
							CredentialSecret: "aws-cred",
							Region:           "asia",
						},
					},
				},
			},
			secret: &core.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "aws-cred",
					Namespace: "demo",
				},
				Data: map[string][]byte{
					"access_key": []byte("1232-2132-123-132"),
				},
			},
			wantErr: true,
		},
		{
			name: "GCPConfig: Unsuccessful operation: Missing secret",
			path: "aws",
			secretEngine: &api.SecretEngine{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test12321",
					Namespace: "demo",
				},
				Spec: api.SecretEngineSpec{
					SecretEngineConfiguration: api.SecretEngineConfiguration{
						AWS: &api.AWSConfiguration{
							CredentialSecret: "aws-cred",
							Region:           "asia",
						},
					},
				},
			},
			secret: &core.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "aws-cred-2343324",
					Namespace: "demo",
				},
				Data: map[string][]byte{
					"access_key": []byte("1232-2132-123-132"),
					"secret_key": []byte("acdenfi-fkjdsk-dsfjds-fsdjf"),
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			vc, err := vaultClient(srv.URL)
			assert.Nil(t, err, "failed to create vault client")

			seClient := &SecretEngine{
				appClient:    &appcatfake.FakeAppcatalogV1alpha1{},
				secretEngine: tt.secretEngine,
				vaultClient:  vc,
				kubeClient:   kfake.NewSimpleClientset(),
				path:         tt.path,
			}
			// Create fake secret for aws config
			_, err = seClient.kubeClient.CoreV1().Secrets("demo").Create(context.TODO(), tt.secret, metav1.CreateOptions{})
			assert.Nil(t, err)

			if err := seClient.CreateAWSConfig(); (err != nil) != tt.wantErr {
				t.Errorf("CreateAWSConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSecretEngine_CreateKVConfig(t *testing.T) {
	srv := NewFakeVaultServer()
	defer srv.Close()

	tests := []struct {
		name         string
		secretEngine *api.SecretEngine
		wantErr      bool
		extraHeaders map[string]string
	}{
		{
			name: "KV - Missing config",
			secretEngine: &api.SecretEngine{
				Spec: api.SecretEngineSpec{
					SecretEngineConfiguration: api.SecretEngineConfiguration{},
				},
			},
			wantErr: true,
		},
		{
			name: "KV V1 - Successful Operation",
			secretEngine: &api.SecretEngine{
				Spec: api.SecretEngineSpec{
					SecretEngineConfiguration: api.SecretEngineConfiguration{
						KV: &api.KVConfiguration{
							Version: 1,
						},
					},
				},
			},
			wantErr: false,
			extraHeaders: map[string]string{
				KVTestHeaderExpectedVersion: "1",
			},
		},
		{
			name: "KV V2 - Successful Operation",
			secretEngine: &api.SecretEngine{
				Spec: api.SecretEngineSpec{
					SecretEngineConfiguration: api.SecretEngineConfiguration{
						KV: &api.KVConfiguration{
							Version: 2,
						},
					},
				},
			},
			wantErr: false,
			extraHeaders: map[string]string{
				KVTestHeaderExpectedVersion: "2",
			},
		},
		{
			name: "KV V2 - Default MaxVersions",
			secretEngine: &api.SecretEngine{
				Spec: api.SecretEngineSpec{
					SecretEngineConfiguration: api.SecretEngineConfiguration{
						KV: &api.KVConfiguration{
							Version: 2,
						},
					},
				},
			},
			wantErr: false,
			extraHeaders: map[string]string{
				KVTestHeaderExpectedVersion:     "2",
				KVTestHeaderExpectedMaxVersions: "0",
			},
		},
		{
			name: "KV V2 - Explicit MaxVersions",
			secretEngine: &api.SecretEngine{
				Spec: api.SecretEngineSpec{
					SecretEngineConfiguration: api.SecretEngineConfiguration{
						KV: &api.KVConfiguration{
							Version:     2,
							MaxVersions: 5,
						},
					},
				},
			},
			wantErr: false,
			extraHeaders: map[string]string{
				KVTestHeaderExpectedVersion:     "2",
				KVTestHeaderExpectedMaxVersions: "5",
			},
		},
		{
			name: "KV V2 - Default CasRequired",
			secretEngine: &api.SecretEngine{
				Spec: api.SecretEngineSpec{
					SecretEngineConfiguration: api.SecretEngineConfiguration{
						KV: &api.KVConfiguration{
							Version: 2,
						},
					},
				},
			},
			wantErr: false,
			extraHeaders: map[string]string{
				KVTestHeaderExpectedVersion:     "2",
				KVTestHeaderExpectedCasRequired: "false",
			},
		},
		{
			name: "KV V2 - Explicit CasRequired",
			secretEngine: &api.SecretEngine{
				Spec: api.SecretEngineSpec{
					SecretEngineConfiguration: api.SecretEngineConfiguration{
						KV: &api.KVConfiguration{
							Version:     2,
							CasRequired: true,
						},
					},
				},
			},
			wantErr: false,
			extraHeaders: map[string]string{
				KVTestHeaderExpectedVersion:     "2",
				KVTestHeaderExpectedCasRequired: "true",
			},
		},
		{
			name: "KV V2 - Default DeleteVersionsAfter",
			secretEngine: &api.SecretEngine{
				Spec: api.SecretEngineSpec{
					SecretEngineConfiguration: api.SecretEngineConfiguration{
						KV: &api.KVConfiguration{
							Version: 2,
						},
					},
				},
			},
			wantErr: false,
			extraHeaders: map[string]string{
				KVTestHeaderExpectedVersion:             "2",
				KVTestHeaderExpectedDeleteVersionsAfter: ExpectBlank,
			},
		},
		{
			name: "KV V2 - Explicit DefaultVersionsAfter",
			secretEngine: &api.SecretEngine{
				Spec: api.SecretEngineSpec{
					SecretEngineConfiguration: api.SecretEngineConfiguration{
						KV: &api.KVConfiguration{
							Version:             2,
							DeleteVersionsAfter: "3h25m19s",
						},
					},
				},
			},
			wantErr: false,
			extraHeaders: map[string]string{
				KVTestHeaderExpectedVersion:             "2",
				KVTestHeaderExpectedDeleteVersionsAfter: "3h25m19s",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			vc, err := vaultClient(srv.URL)
			assert.Nil(t, err, "failed to create vault client")

			if tt.extraHeaders != nil {
				headers := vc.Headers()
				for k, v := range tt.extraHeaders {
					headers.Add(k, v)
				}
				vc.SetHeaders(headers)
			}

			seClient := &SecretEngine{
				appClient:    &appcatfake.FakeAppcatalogV1alpha1{},
				secretEngine: tt.secretEngine,
				vaultClient:  vc,
				kubeClient:   kfake.NewSimpleClientset(),
				path:         DefaultKVPath,
			}

			if err := seClient.CreateKVConfig(); (err != nil) != tt.wantErr {
				t.Errorf("CreateKVConfig() error = %v, wantErr: %v", err, tt.wantErr)
			}
		})
	}
}
