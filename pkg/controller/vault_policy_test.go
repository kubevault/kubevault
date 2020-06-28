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

	"kubevault.dev/operator/apis"
	policyapi "kubevault.dev/operator/apis/policy/v1alpha1"
	csfake "kubevault.dev/operator/client/clientset/versioned/fake"
	"kubevault.dev/operator/pkg/vault/policy"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kfake "k8s.io/client-go/kubernetes/fake"
	appcat "kmodules.xyz/custom-resources/apis/appcatalog/v1alpha1"
	appcatfake "kmodules.xyz/custom-resources/client/clientset/versioned/fake"
)

type fakePolicy struct {
	errInPutPolicy bool
}

func (f *fakePolicy) EnsurePolicy(n, p string) error {
	if f.errInPutPolicy {
		return errors.New("error")
	}
	return nil
}

func (f *fakePolicy) DeletePolicy(n string) error {
	return nil
}

func simpleVaultPolicy() *policyapi.VaultPolicy {
	return &policyapi.VaultPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "simple",
			Namespace:  "test",
			Finalizers: []string{VaultPolicyFinalizer},
		},
		Spec: policyapi.VaultPolicySpec{
			PolicyDocument: "simple {}",
		},
	}
}

func vaultAppBinding(vAddr, tokenSecret string) *appcat.AppBinding {
	return &appcat.AppBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "vault",
			Namespace: "test",
		},
		Spec: appcat.AppBindingSpec{
			Secret: &core.LocalObjectReference{
				Name: tokenSecret,
			},
			ClientConfig: appcat.ClientConfig{
				URL:                   &vAddr,
				InsecureSkipTLSVerify: true,
			},
		},
	}
}

func validVaultPolicy(app *appcat.AppBinding) *policyapi.VaultPolicy {
	return &policyapi.VaultPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "ok",
			Namespace:  "test",
			Finalizers: []string{VaultPolicyFinalizer},
		},
		Spec: policyapi.VaultPolicySpec{
			PolicyDocument: "simple {}",
			VaultRef: core.LocalObjectReference{
				Name: app.Name,
			},
		},
	}
}

func vaultTokenSecret() *core.Secret {
	return &core.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "vault",
			Namespace: "test",
		},
		Type: apis.SecretTypeTokenAuth,
		Data: map[string][]byte{
			"token": []byte("root"),
		},
	}
}

func NewFakeVaultServer() *httptest.Server {
	router := mux.NewRouter()

	router.HandleFunc("/v1/sys/policies/acl/ok", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}).Methods(http.MethodDelete)

	router.HandleFunc("/v1/sys/policies/acl/{policy}", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}).Methods(http.MethodDelete)

	router.HandleFunc("/v1/sys/policies/acl/simple", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}).Methods(http.MethodDelete)

	router.HandleFunc("/v1/auth/kubernetes/role/{role}", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}).Methods(http.MethodDelete)

	return httptest.NewServer(router)
}

func TestReconcilePolicy(t *testing.T) {
	cases := []struct {
		testName     string
		vPolicy      *policyapi.VaultPolicy
		pClient      policy.Policy
		expectStatus string
		expectErr    bool
	}{
		{
			testName:     "reconcile successful",
			vPolicy:      simpleVaultPolicy(),
			pClient:      &fakePolicy{},
			expectStatus: string(policyapi.PolicySuccess),
			expectErr:    false,
		},
		{
			testName:     "reconcile unsuccessful, error occure in EnsurePolicy",
			vPolicy:      simpleVaultPolicy(),
			pClient:      &fakePolicy{errInPutPolicy: true},
			expectStatus: string(policyapi.PolicyFailed),
			expectErr:    true,
		},
	}

	for _, c := range cases {
		t.Run(c.testName, func(t *testing.T) {
			ctrl := &VaultController{
				extClient: csfake.NewSimpleClientset(simpleVaultPolicy()),
			}

			err := ctrl.reconcilePolicy(c.vPolicy, c.pClient)
			if c.expectErr {
				assert.NotNil(t, err, "expected error")
			} else {
				assert.Nil(t, err)
			}
			if c.expectStatus != "" {
				p, err := ctrl.extClient.PolicyV1alpha1().VaultPolicies(c.vPolicy.Namespace).Get(context.TODO(), c.vPolicy.Name, metav1.GetOptions{})
				if assert.Nil(t, err) {
					assert.Condition(t, func() (success bool) {
						return c.expectStatus == string(p.Status.Phase)
					}, ".status.status should match")
				}
			}
		})
	}
}

func TestFinalizePolicy(t *testing.T) {
	srv := NewFakeVaultServer()
	defer srv.Close()

	vApp := vaultAppBinding(srv.URL, vaultTokenSecret().Name)
	appc := appcatfake.NewSimpleClientset(vApp)
	kc := kfake.NewSimpleClientset(vaultTokenSecret())

	cases := []struct {
		testName  string
		vPolicy   *policyapi.VaultPolicy
		expectErr bool
	}{
		{
			testName:  "no error, valid VaultPolicy",
			vPolicy:   validVaultPolicy(vApp),
			expectErr: false,
		},
		{
			testName:  "no error, VaultPolicy doesn't exist",
			vPolicy:   nil,
			expectErr: false,
		},
		{
			testName: "no error, invalid AppBinding",
			vPolicy: validVaultPolicy(&appcat.AppBinding{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "invalid",
					Namespace: "test",
				},
			}),
			expectErr: false,
		},
	}

	for _, c := range cases {
		t.Run(c.testName, func(t *testing.T) {
			cc := csfake.NewSimpleClientset()
			pc := cc.PolicyV1alpha1()
			if c.vPolicy != nil {
				_, err := pc.VaultPolicies(c.vPolicy.Namespace).Create(context.TODO(), c.vPolicy, metav1.CreateOptions{})
				assert.Nil(t, err)
			} else {
				c.vPolicy = simpleVaultPolicy()
			}
			ctrl := &VaultController{
				extClient:        cc,
				kubeClient:       kc,
				appCatalogClient: appc.AppcatalogV1alpha1(),
			}

			err := ctrl.finalizePolicy(c.vPolicy)
			if c.expectErr {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func TestVaultController_runPolicyFinalizer(t *testing.T) {
	srv := NewFakeVaultServer()
	defer srv.Close()
	vApp := vaultAppBinding(srv.URL, vaultTokenSecret().Name)
	appc := appcatfake.NewSimpleClientset(vApp)
	ctrl := &VaultController{
		extClient:        csfake.NewSimpleClientset(simpleVaultPolicy(), validVaultPolicy(vApp)),
		kubeClient:       kfake.NewSimpleClientset(vaultTokenSecret()),
		finalizerInfo:    NewMapFinalizer(),
		appCatalogClient: appc.AppcatalogV1alpha1(),
	}
	tests := []struct {
		name    string
		vPolicy *policyapi.VaultPolicy
		wantErr bool
	}{
		{
			name:    "remove finalizer successfully, valid VaultPolicy",
			vPolicy: validVaultPolicy(vApp),
			wantErr: false,
		},
		{
			name: "remove finalizer successfully, missing vault server",
			vPolicy: validVaultPolicy(&appcat.AppBinding{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "invalid",
					Namespace: "test",
				},
			}),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ctrl.runPolicyFinalizer(tt.vPolicy); (err != nil) != tt.wantErr {
				t.Errorf("runPolicyFinalizer() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
