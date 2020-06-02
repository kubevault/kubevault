/*
Copyright The KubeVault Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	approleapi "kubevault.dev/operator/apis/approle/v1alpha1"
	csfake "kubevault.dev/operator/client/clientset/versioned/fake"
	"kubevault.dev/operator/pkg/vault/approle"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kfake "k8s.io/client-go/kubernetes/fake"
	appcat "kmodules.xyz/custom-resources/apis/appcatalog/v1alpha1"
	appcatfake "kmodules.xyz/custom-resources/client/clientset/versioned/fake"
)

func NewFakeVaultAppRoleServer() *httptest.Server {
	router := mux.NewRouter()

	router.HandleFunc("/v1/auth/approle/role/{role}", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}).Methods(http.MethodDelete)

	return httptest.NewServer(router)
}

type fakeAppRole struct {
	errInPutAppRole bool
}

func (f *fakeAppRole) EnsureAppRole(n string, m map[string]interface{}) error {
	if f.errInPutAppRole {
		return errors.New("error")
	}
	return nil
}

func (f *fakeAppRole) DeleteAppRole(n string) error {
	return nil
}

func simpleVaultAppRole() *approleapi.VaultAppRole {
	return &approleapi.VaultAppRole{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "simple",
			Namespace:  "test",
			Finalizers: []string{VaultAppRoleFinalizer},
		},
		Spec: approleapi.VaultAppRoleSpec{},
	}
}

func validVaultAppRole(app *appcat.AppBinding) *approleapi.VaultAppRole {
	return &approleapi.VaultAppRole{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "ok",
			Namespace:  "test",
			Finalizers: []string{VaultAppRoleFinalizer},
		},
		Spec: approleapi.VaultAppRoleSpec{
			VaultRef: core.LocalObjectReference{
				Name: app.Name,
			},
		},
	}
}

func TestReconcileAppRole(t *testing.T) {
	cases := []struct {
		testName     string
		vAppRole     *approleapi.VaultAppRole
		aClient      approle.AppRole
		expectStatus string
		expectErr    bool
	}{
		{
			testName:     "reconcile successful",
			vAppRole:     simpleVaultAppRole(),
			aClient:      &fakeAppRole{},
			expectStatus: string(approleapi.AppRoleSuccess),
			expectErr:    false,
		},
		{
			testName:     "reconcile unsuccessful, error occure in EnsurePolicy",
			vAppRole:     simpleVaultAppRole(),
			aClient:      &fakeAppRole{errInPutAppRole: true},
			expectStatus: string(approleapi.AppRoleFailed),
			expectErr:    true,
		},
	}

	for _, c := range cases {
		t.Run(c.testName, func(t *testing.T) {
			ctrl := &VaultController{
				extClient: csfake.NewSimpleClientset(simpleVaultAppRole()),
			}

			err := ctrl.reconcileAppRole(c.vAppRole, c.aClient)
			if c.expectErr {
				assert.NotNil(t, err, "expected error")
			} else {
				assert.Nil(t, err)
			}
			if c.expectStatus != "" {
				p, err := ctrl.extClient.ApproleV1alpha1().VaultAppRoles(c.vAppRole.Namespace).Get(context.TODO(), c.vAppRole.Name, metav1.GetOptions{})
				if assert.Nil(t, err) {
					assert.Condition(t, func() (success bool) {
						return c.expectStatus == string(p.Status.Phase)
					}, ".status.status should match")
				}
			}
		})
	}
}

func TestFinalizeAppRole(t *testing.T) {
	srv := NewFakeVaultAppRoleServer()
	defer srv.Close()

	vApp := vaultAppBinding(srv.URL, vaultTokenSecret().Name)
	appc := appcatfake.NewSimpleClientset(vApp)
	kc := kfake.NewSimpleClientset(vaultTokenSecret())

	cases := []struct {
		testName  string
		vAppRole  *approleapi.VaultAppRole
		expectErr bool
	}{
		{
			testName:  "no error, valid VaultAppRole",
			vAppRole:  validVaultAppRole(vApp),
			expectErr: false,
		},
		{
			testName:  "no error, VaultAppRole doesn't exist",
			vAppRole:  nil,
			expectErr: false,
		},
		{
			testName: "error, invalid VaultAppRole",
			vAppRole: validVaultAppRole(&appcat.AppBinding{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "invalid",
					Namespace: "test",
				},
			}),
			expectErr: true,
		},
	}

	for _, c := range cases {
		t.Run(c.testName, func(t *testing.T) {
			cc := csfake.NewSimpleClientset()
			pc := cc.ApproleV1alpha1()
			if c.vAppRole != nil {
				_, err := pc.VaultAppRoles(c.vAppRole.Namespace).Create(context.TODO(), c.vAppRole, metav1.CreateOptions{})
				assert.Nil(t, err)
			} else {
				c.vAppRole = simpleVaultAppRole()
			}
			ctrl := &VaultController{
				extClient:        cc,
				kubeClient:       kc,
				appCatalogClient: appc.AppcatalogV1alpha1(),
			}

			err := ctrl.finalizeAppRole(c.vAppRole)
			if c.expectErr {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func TestRunAppRoleFinalizer(t *testing.T) {
	srv := NewFakeVaultAppRoleServer()
	defer srv.Close()
	vApp := vaultAppBinding(srv.URL, vaultTokenSecret().Name)
	appc := appcatfake.NewSimpleClientset(vApp)
	ctrl := &VaultController{
		extClient:        csfake.NewSimpleClientset(simpleVaultAppRole(), validVaultAppRole(vApp)),
		kubeClient:       kfake.NewSimpleClientset(vaultTokenSecret()),
		finalizerInfo:    NewMapFinalizer(),
		appCatalogClient: appc.AppcatalogV1alpha1(),
	}
	ctrl.finalizerInfo.Add(simpleVaultAppRole().GetKey())

	cases := []struct {
		testName  string
		vAppRole  *approleapi.VaultAppRole
		completed bool
	}{
		{
			testName:  "remove finalizer successfully, valid VaultAppRole",
			vAppRole:  validVaultAppRole(vApp),
			completed: true,
		},
		{
			testName: "remove finalizer successfully, invalid VaultAppRole",
			vAppRole: validVaultAppRole(&appcat.AppBinding{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "invalid",
					Namespace: "test",
				},
			}),
			completed: true,
		},
		{
			testName:  "already processing finalizer",
			vAppRole:  simpleVaultAppRole(),
			completed: false,
		},
	}

	for _, c := range cases {
		t.Run(c.testName, func(t *testing.T) {
			ctrl.runAppRoleFinalizer(c.vAppRole, 3*time.Second, 1*time.Second)
			if c.completed {
				assert.Condition(t, func() (success bool) {
					return !ctrl.finalizerInfo.IsAlreadyProcessing(c.vAppRole.GetKey())
				}, "IsAlreadyProcessing(key) should be false")

			} else {
				assert.Condition(t, func() (success bool) {
					return ctrl.finalizerInfo.IsAlreadyProcessing(c.vAppRole.GetKey())
				}, "IsAlreadyProcessing(key) should be true")
			}
		})
	}
}
