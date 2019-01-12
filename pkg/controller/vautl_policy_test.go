package controller

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/appscode/pat"
	"github.com/kubevault/operator/apis"
	policyapi "github.com/kubevault/operator/apis/policy/v1alpha1"
	csfake "github.com/kubevault/operator/client/clientset/versioned/fake"
	"github.com/kubevault/operator/pkg/vault/policy"
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
			Policy: "simple {}",
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
			Policy: "simple {}",
			VaultAppRef: &appcat.AppReference{
				Name:      app.Name,
				Namespace: app.Namespace,
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
	m := pat.New()
	m.Del("/v1/sys/policies/acl/ok", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	m.Del("/v1/sys/policies/acl/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	m.Del("/v1/sys/policies/acl/simple", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	m.Del("/v1/auth/kubernetes/role/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	return httptest.NewServer(m)
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
				p, err := ctrl.extClient.PolicyV1alpha1().VaultPolicies(c.vPolicy.Namespace).Get(c.vPolicy.Name, metav1.GetOptions{})
				if assert.Nil(t, err) {
					assert.Condition(t, func() (success bool) {
						return c.expectStatus == string(p.Status.Status)
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
			testName: "error, invalid VaultPolicy",
			vPolicy: validVaultPolicy(&appcat.AppBinding{
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
			pc := cc.PolicyV1alpha1()
			if c.vPolicy != nil {
				_, err := pc.VaultPolicies(c.vPolicy.Namespace).Create(c.vPolicy)
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

func TestRunPolicyFinalizer(t *testing.T) {
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
	ctrl.finalizerInfo.Add(simpleVaultPolicy().GetKey())

	cases := []struct {
		testName  string
		vPolicy   *policyapi.VaultPolicy
		completed bool
	}{
		{
			testName:  "remove finalizer successfully, valid VaultPolicy",
			vPolicy:   validVaultPolicy(vApp),
			completed: true,
		},
		{
			testName: "remove finalizer successfully, invalid VaultPolicy",
			vPolicy: validVaultPolicy(&appcat.AppBinding{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "invalid",
					Namespace: "test",
				},
			}),
			completed: true,
		},
		{
			testName:  "already processing finalizer",
			vPolicy:   simpleVaultPolicy(),
			completed: false,
		},
	}

	for _, c := range cases {
		t.Run(c.testName, func(t *testing.T) {
			ctrl.runPolicyFinalizer(c.vPolicy, 3*time.Second, 1*time.Second)
			if c.completed {
				assert.Condition(t, func() (success bool) {
					return !ctrl.finalizerInfo.IsAlreadyProcessing(c.vPolicy.GetKey())
				}, "IsAlreadyProcessing(key) should be false")

			} else {
				assert.Condition(t, func() (success bool) {
					return ctrl.finalizerInfo.IsAlreadyProcessing(c.vPolicy.GetKey())
				}, "IsAlreadyProcessing(key) should be true")
			}
		})
	}
}
