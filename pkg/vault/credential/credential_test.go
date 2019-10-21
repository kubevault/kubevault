package credential

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	vaultapi "github.com/hashicorp/vault/api"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kfake "k8s.io/client-go/kubernetes/fake"
	"kubevault.dev/operator/pkg/vault/util"
)

func vaultServer() *httptest.Server {
	router := mux.NewRouter()

	router.HandleFunc("/v1/sys/leases/revoke/success", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		util.LogWriteErr(w.Write([]byte("{}")))
	}).Methods(http.MethodPut)

	router.HandleFunc("/v1/sys/leases/revoke/error", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		util.LogWriteErr(w.Write([]byte("error")))
	}).Methods(http.MethodPut)

	router.HandleFunc("/v1/sys/leases/lookup", func(w http.ResponseWriter, r *http.Request) {
		data := struct {
			LeaseID string `json:"lease_id"`
		}{}
		err := json.NewDecoder(r.Body).Decode(&data)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			util.LogWriteErr(w.Write([]byte(err.Error())))
		}

		if data.LeaseID == "1234" {
			w.WriteHeader(http.StatusOK)
			util.LogWriteErr(w.Write([]byte(`{}`)))
		} else {
			w.WriteHeader(http.StatusBadRequest)
			util.LogWriteErr(w.Write([]byte(`{"errors":["invalid lease"]}`)))
		}

	}).Methods(http.MethodPut)

	return httptest.NewServer(router)
}

type fakeDBCredM struct {
	getSecretErr bool
	cred         *vaultapi.Secret
}

func (f *fakeDBCredM) GetSecret() (*vaultapi.Secret, error) {
	if f.getSecretErr {
		return nil, errors.New("error")
	}
	return f.cred, nil
}

func (f *fakeDBCredM) ParseCredential(secret *vaultapi.Secret) (map[string][]byte, error) {
	return nil, nil
}

func (f *fakeDBCredM) GetOwnerReference() metav1.OwnerReference {
	return metav1.OwnerReference{}
}

func TestCreateSecret(t *testing.T) {
	cred := &vaultapi.Secret{
		LeaseID:       "1204",
		LeaseDuration: 300,
		Data: map[string]interface{}{
			"password": "1234",
			"username": "nahid",
		},
	}

	testData := []struct {
		testName     string
		credManager  *CredManager
		cred         *vaultapi.Secret
		secretName   string
		namespace    string
		createSecret bool
	}{
		{
			testName: "Successfully secret created",
			credManager: &CredManager{
				secretEngine: &fakeDBCredM{},
			},
			cred:         cred,
			secretName:   "pg-cred",
			namespace:    "pg",
			createSecret: false,
		},
		{
			testName: "create secret, secret already exists, no error",
			credManager: &CredManager{
				secretEngine: &fakeDBCredM{},
			},
			cred:         cred,
			secretName:   "pg-cred",
			namespace:    "pg",
			createSecret: true,
		},
	}

	for _, test := range testData {
		t.Run(test.testName, func(t *testing.T) {
			d := test.credManager
			d.kubeClient = kfake.NewSimpleClientset()

			if test.createSecret {
				_, err := d.kubeClient.CoreV1().Secrets(test.namespace).Create(&corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: test.namespace,
						Name:      test.secretName,
					},
					Data: map[string][]byte{
						"test": []byte("hi"),
					},
				})

				assert.Nil(t, err)
			}

			err := d.CreateSecret(test.secretName, test.namespace, test.cred)
			if assert.Nil(t, err) {
				_, err := d.kubeClient.CoreV1().Secrets(test.namespace).Get(test.secretName, metav1.GetOptions{})
				assert.Nil(t, err)
			}
		})
	}
}

func TestCreateRole(t *testing.T) {
	testData := []struct {
		testName    string
		credManager *CredManager
		createRole  bool
		roleName    string
		secretName  string
		namespace   string
	}{
		{
			testName: "Successfully role created",
			credManager: &CredManager{
				secretEngine: &fakeDBCredM{},
			},
			createRole: false,
			roleName:   "pg-role",
			secretName: "pg-cred",
			namespace:  "pg",
		},
		{
			testName: "create role, role already exists, no error",
			credManager: &CredManager{
				secretEngine: &fakeDBCredM{},
			},
			createRole: true,
			roleName:   "pg-role",
			secretName: "pg-cred",
			namespace:  "pg",
		},
	}

	for _, test := range testData {
		t.Run(test.testName, func(t *testing.T) {
			d := test.credManager
			d.kubeClient = kfake.NewSimpleClientset()

			if test.createRole {
				_, err := d.kubeClient.RbacV1().Roles(test.namespace).Create(&rbacv1.Role{
					ObjectMeta: metav1.ObjectMeta{
						Name:      test.roleName,
						Namespace: test.namespace,
					},
				})

				assert.Nil(t, err)
			}

			err := d.CreateRole(test.roleName, test.namespace, test.secretName)
			if assert.Nil(t, err) {
				r, err := d.kubeClient.RbacV1().Roles(test.namespace).Get(test.roleName, metav1.GetOptions{})
				if assert.Nil(t, err) {
					assert.Equal(t, "", r.Rules[0].APIGroups[0], "api group")
					assert.Equal(t, "secrets", r.Rules[0].Resources[0], "resources")
					assert.Equal(t, test.secretName, r.Rules[0].ResourceNames[0], "resource name")
					assert.Equal(t, "get", r.Rules[0].Verbs[0], "verbs")
				}
			}
		})
	}
}

func TestCreateRoleBinding(t *testing.T) {
	subs := []rbacv1.Subject{
		{
			Kind:      rbacv1.ServiceAccountKind,
			Namespace: "pg",
			Name:      "pg-sa",
		},
	}

	testData := []struct {
		testName        string
		credManager     *CredManager
		createRB        bool
		roleName        string
		roleBindingName string
		namespace       string
		subjects        []rbacv1.Subject
	}{
		{
			testName: "Successfully role binding created",
			credManager: &CredManager{
				secretEngine: &fakeDBCredM{},
			},
			createRB:        false,
			roleName:        "pg-role",
			roleBindingName: "pg-role-binding",
			subjects:        subs,
		},
		{
			testName: "Successfully role binding patched",
			credManager: &CredManager{
				secretEngine: &fakeDBCredM{},
			},
			createRB:        true,
			roleName:        "pg-role",
			roleBindingName: "pg-role-binding",
			subjects:        subs,
			namespace:       "pg",
		},
	}

	for _, test := range testData {
		t.Run(test.testName, func(t *testing.T) {
			d := test.credManager
			d.kubeClient = kfake.NewSimpleClientset()

			if test.createRB {
				_, err := d.kubeClient.RbacV1().RoleBindings(test.namespace).Create(&rbacv1.RoleBinding{
					ObjectMeta: metav1.ObjectMeta{
						Name:      test.roleBindingName,
						Namespace: test.namespace,
					},
				})

				assert.Nil(t, err)
			}

			err := d.CreateRoleBinding(test.roleBindingName, test.namespace, test.roleName, test.subjects)
			if assert.Nil(t, err) {
				r, err := d.kubeClient.RbacV1().RoleBindings(test.namespace).Get(test.roleBindingName, metav1.GetOptions{})
				if assert.Nil(t, err) {
					assert.Equal(t, test.roleName, r.RoleRef.Name, "role ref role name")
					assert.Equal(t, "Role", r.RoleRef.Kind, "role ref role kind")
					assert.Equal(t, rbacv1.GroupName, r.RoleRef.APIGroup, "role ref role api group")
				}
			}
		})
	}
}

func TestRevokeLease(t *testing.T) {
	srv := vaultServer()
	defer srv.Close()

	cfg := vaultapi.DefaultConfig()
	cfg.Address = srv.URL

	cl, err := vaultapi.NewClient(cfg)
	if !assert.Nil(t, err, "failed to create vault client") {
		return
	}

	testData := []struct {
		testName    string
		credManager *CredManager
		expectedErr bool
		leaseID     string
	}{
		{
			testName: "Lease revoke successful",
			credManager: &CredManager{
				vaultClient: cl,
			},
			leaseID:     "success",
			expectedErr: false,
		},
		{
			testName: "Lease revoke failed",
			credManager: &CredManager{
				vaultClient: cl,
			},
			leaseID:     "error",
			expectedErr: true,
		},
	}

	for _, test := range testData {
		t.Run(test.testName, func(t *testing.T) {
			err := test.credManager.RevokeLease(test.leaseID)
			if test.expectedErr {
				assert.NotNil(t, err, "expected error")
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func TestDatabaseRoleBinding_IsLeaseExpired(t *testing.T) {
	srv := vaultServer()
	defer srv.Close()

	cfg := vaultapi.DefaultConfig()
	cfg.Address = srv.URL

	cl, err := vaultapi.NewClient(cfg)
	if !assert.Nil(t, err, "failed to create vault client") {
		return
	}

	testData := []struct {
		testName    string
		credManager *CredManager
		isExpired   bool
		leaseID     string
	}{
		{
			testName: "lease is expired",
			credManager: &CredManager{
				vaultClient: cl,
			},
			isExpired: true,
			leaseID:   "1222",
		},
		{
			testName: "lease is valid",
			credManager: &CredManager{
				vaultClient: cl,
			},
			isExpired: false,
			leaseID:   "1234",
		},
	}

	for _, test := range testData {
		t.Run(test.testName, func(t *testing.T) {
			ok, err := test.credManager.IsLeaseExpired(test.leaseID)
			if assert.Nil(t, err) {
				assert.Condition(t, func() (success bool) {
					return ok == test.isExpired
				})
			}
		})
	}
}

func TestCredManager_GetCredential(t *testing.T) {
	cred := &vaultapi.Secret{
		LeaseID:       "1204",
		LeaseDuration: 300,
		Data: map[string]interface{}{
			"password": "1234",
			"username": "nahid",
		},
	}

	testData := []struct {
		testName    string
		credManager *CredManager
		cred        *vaultapi.Secret
		expectedErr bool
	}{
		{
			testName: "Successfully get credential",
			credManager: &CredManager{
				secretEngine: &fakeDBCredM{
					cred: cred,
				},
			},
			cred:        cred,
			expectedErr: false,
		},
		{
			testName: "failed to get credential, json error",
			credManager: &CredManager{
				secretEngine: &fakeDBCredM{
					getSecretErr: true,
				},
			},
			cred:        nil,
			expectedErr: true,
		},
	}

	for _, test := range testData {
		t.Run(test.testName, func(t *testing.T) {
			d := test.credManager
			cred, err := d.GetCredential()
			if test.expectedErr {
				assert.NotNil(t, err, "error should occur")
			} else {
				if assert.Nil(t, err) {
					assert.Equal(t, *test.cred, *cred, "credential should match")
				}
			}
		})
	}
}
