package engine

import (
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	vaultapi "github.com/hashicorp/vault/api"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kfake "k8s.io/client-go/kubernetes/fake"
	appcatfake "kmodules.xyz/custom-resources/client/clientset/versioned/typed/appcatalog/v1alpha1/fake"
	api "kubevault.dev/operator/apis/engine/v1alpha1"
)

const Token = "root"

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
	}).Methods(http.MethodPost)

	router.HandleFunc("/v1/azure/config", func(w http.ResponseWriter, r *http.Request) {
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
	}).Methods(http.MethodPost)

	router.HandleFunc("/v1/aws/config/root", func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		var data interface{}
		err := json.NewDecoder(r.Body).Decode(&data)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_, err := w.Write([]byte(err.Error()))
			log.Println(err)
			return
		}
		m := data.(map[string]interface{})
		if v, ok := m["access_key"]; !ok || len(v.(string)) == 0 {
			w.WriteHeader(http.StatusBadRequest)
			_, err := w.Write([]byte("access_key id isn't provided"))
			log.Println(err)
			return
		}

		if v, ok := m["secret_key"]; !ok || len(v.(string)) == 0 {
			w.WriteHeader(http.StatusBadRequest)
			_, err := w.Write([]byte("secret_key isn't provided"))
			log.Println(err)
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
	}).Methods(http.MethodPost)

	router.HandleFunc("/v1/database/config/mongodb", func(w http.ResponseWriter, r *http.Request) {
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
	}).Methods(http.MethodPost)

	router.HandleFunc("/v1/database/config/mysql", func(w http.ResponseWriter, r *http.Request) {
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
	}).Methods(http.MethodPost)

	router.HandleFunc("/v1/database/config/postgres", func(w http.ResponseWriter, r *http.Request) {
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
			_, err = secretEngineClient.kubeClient.CoreV1().Secrets("demo").Create(&corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name: "gcp-cred",
				},
				Data: map[string][]byte{
					"sa.json": []byte("fakeKey"),
				},
			})
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
		secret       *corev1.Secret
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
			secret: &corev1.Secret{
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
			secret: &corev1.Secret{
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
			secret: &corev1.Secret{
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
			secret: &corev1.Secret{
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
			secret: &corev1.Secret{
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
			_, err = seClient.kubeClient.CoreV1().Secrets("demo").Create(tt.secret)
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
		secret       *corev1.Secret
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
			secret: &corev1.Secret{
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
			secret: &corev1.Secret{
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
			secret: &corev1.Secret{
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
			secret: &corev1.Secret{
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
			_, err = seClient.kubeClient.CoreV1().Secrets("demo").Create(tt.secret)
			assert.Nil(t, err)

			if err := seClient.CreateAWSConfig(); (err != nil) != tt.wantErr {
				t.Errorf("CreateAWSConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
