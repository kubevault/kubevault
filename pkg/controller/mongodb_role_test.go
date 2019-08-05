package controller

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/appscode/go/encoding/json/types"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kfake "k8s.io/client-go/kubernetes/fake"
	appcat "kmodules.xyz/custom-resources/apis/appcatalog/v1alpha1"
	api "kubevault.dev/operator/apis/engine/v1alpha1"
	cs "kubevault.dev/operator/client/clientset/versioned/fake"
	dbinformers "kubevault.dev/operator/client/informers/externalversions"
	"kubevault.dev/operator/pkg/vault/role/database"
)

func setupVaultServerForMongodb() *httptest.Server {
	router := mux.NewRouter()

	router.HandleFunc("/v1/database/roles/m-read", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}).Methods(http.MethodDelete)

	router.HandleFunc("/v1/database/roles/error", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("error"))
	}).Methods(http.MethodDelete)

	return httptest.NewServer(router)
}

func TestUserManagerController_reconcileMongoDBRole(t *testing.T) {
	mRole := api.MongoDBRole{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "m-role",
			Namespace:  "m",
			Generation: 0,
		},
		Spec: api.MongoDBRoleSpec{
			VaultRef:    corev1.LocalObjectReference{Name: "test121"},
			DatabaseRef: appcat.AppReference{Name: "test"},
		},
	}

	testData := []struct {
		testName           string
		mRole              api.MongoDBRole
		dbRClient          database.DatabaseRoleInterface
		hasStatusCondition bool
		expectedErr        bool
	}{
		{
			testName:           "initial stage, no error",
			mRole:              mRole,
			dbRClient:          &fakeDRole{},
			expectedErr:        false,
			hasStatusCondition: false,
		},
		{
			testName: "initial stage, failed to create database role",
			mRole:    mRole,
			dbRClient: &fakeDRole{
				errorOccurredInCreateRole: true,
			},
			expectedErr:        true,
			hasStatusCondition: true,
		},
		{
			testName: "update role, successfully updated database role",
			mRole: func(p api.MongoDBRole) api.MongoDBRole {
				p.Generation = 2
				p.Status.ObservedGeneration = types.IntHashForGeneration(1)
				return p
			}(mRole),
			dbRClient:          &fakeDRole{},
			expectedErr:        false,
			hasStatusCondition: false,
		},
		{
			testName: "update role, failed to update database role",
			mRole: func(p api.MongoDBRole) api.MongoDBRole {
				p.Generation = 2
				p.Status.ObservedGeneration = types.IntHashForGeneration(1)
				return p
			}(mRole),
			dbRClient: &fakeDRole{
				errorOccurredInCreateRole: true,
			},
			expectedErr:        true,
			hasStatusCondition: true,
		},
	}

	for _, test := range testData {
		t.Run(test.testName, func(t *testing.T) {
			c := &VaultController{
				kubeClient: kfake.NewSimpleClientset(),
				extClient:  cs.NewSimpleClientset(),
			}
			c.extInformerFactory = dbinformers.NewSharedInformerFactory(c.extClient, time.Minute*10)

			_, err := c.extClient.EngineV1alpha1().MongoDBRoles(test.mRole.Namespace).Create(&test.mRole)
			if !assert.Nil(t, err) {
				return
			}

			err = c.reconcileMongoDBRole(test.dbRClient, &test.mRole)
			if test.expectedErr {
				if assert.NotNil(t, err) {
					if test.hasStatusCondition {
						p, err2 := c.extClient.EngineV1alpha1().MongoDBRoles(test.mRole.Namespace).Get(test.mRole.Name, metav1.GetOptions{})
						if assert.Nil(t, err2) {
							assert.Condition(t, func() (success bool) {
								if len(p.Status.Conditions) == 0 {
									return false
								}
								return true
							}, "should have status.conditions")
						}
					}
				}
			} else {
				if assert.Nil(t, err) {
					p, err2 := c.extClient.EngineV1alpha1().MongoDBRoles(test.mRole.Namespace).Get(test.mRole.Name, metav1.GetOptions{})
					if assert.Nil(t, err2) {
						assert.Condition(t, func() (success bool) {
							if len(p.Status.Conditions) != 0 {
								return false
							}
							return true
						}, "should not have status.conditions")
					}
				}
			}
		})
	}

}
