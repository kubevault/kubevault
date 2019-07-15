package controller

import (
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kfake "k8s.io/client-go/kubernetes/fake"
	appcatfake "kmodules.xyz/custom-resources/client/clientset/versioned/fake"
	policyapi "kubevault.dev/operator/apis/policy/v1alpha1"
	csfake "kubevault.dev/operator/client/clientset/versioned/fake"
	pbinding "kubevault.dev/operator/pkg/vault/policybinding"
)

type fakePBind struct {
	errInEnsure bool
}

func (f *fakePBind) Ensure(n string) error {
	if f.errInEnsure {
		return errors.New("error")
	}
	return nil
}

func (f *fakePBind) Delete(n string) error {
	return nil
}

func simpleVaultPolicyBinding() *policyapi.VaultPolicyBinding {
	return &policyapi.VaultPolicyBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "simple",
			Namespace:  "test",
			Finalizers: []string{VaultPolicyBindingFinalizer},
		},
		Spec: policyapi.VaultPolicyBindingSpec{},
	}
}

func validVaultPolicyBinding(policyName string) *policyapi.VaultPolicyBinding {
	p := simpleVaultPolicyBinding()
	p.Name = "valid"
	p.Spec.Policies = []string{policyName}
	return p
}

func TestReconcilePolicyBinding(t *testing.T) {
	cases := []struct {
		testName     string
		vPBind       *policyapi.VaultPolicyBinding
		pBClient     pbinding.PolicyBinding
		expectStatus string
		expectErr    bool
	}{
		{
			testName:     "reconcile successful",
			vPBind:       simpleVaultPolicyBinding(),
			pBClient:     &fakePBind{},
			expectStatus: string(policyapi.PolicyBindingSuccess),
			expectErr:    false,
		},
		{
			testName:     "reconcile unsuccessful, error occur in Ensure",
			vPBind:       simpleVaultPolicyBinding(),
			pBClient:     &fakePBind{errInEnsure: true},
			expectStatus: string(policyapi.PolicyBindingFailed),
			expectErr:    true,
		},
	}

	for _, c := range cases {
		t.Run(c.testName, func(t *testing.T) {
			ctrl := &VaultController{
				extClient: csfake.NewSimpleClientset(simpleVaultPolicyBinding()),
			}

			err := ctrl.reconcilePolicyBinding(c.vPBind, c.pBClient)
			if c.expectErr {
				assert.NotNil(t, err, "expected error")
			} else {
				assert.Nil(t, err)
			}
			if c.expectStatus != "" {
				p, err := ctrl.extClient.PolicyV1alpha1().VaultPolicyBindings(c.vPBind.Namespace).Get(c.vPBind.Name, metav1.GetOptions{})
				if assert.Nil(t, err) {
					assert.Condition(t, func() (success bool) {
						return c.expectStatus == string(p.Status.Status)
					}, ".status.status should match")
				}
			}
		})
	}
}

func TestFinalizePolicyBinding(t *testing.T) {
	srv := NewFakeVaultServer()
	defer srv.Close()
	vApp := vaultAppBinding(srv.URL, vaultTokenSecret().Name)
	appc := appcatfake.NewSimpleClientset(vApp)
	kc := kfake.NewSimpleClientset(vaultTokenSecret())

	cases := []struct {
		testName  string
		vPBind    *policyapi.VaultPolicyBinding
		vPolicy   *policyapi.VaultPolicy
		expectErr bool
	}{
		{
			testName:  "no error, valid VaultPolicyBinding",
			vPolicy:   validVaultPolicy(vApp),
			vPBind:    validVaultPolicyBinding(validVaultPolicy(vApp).Name),
			expectErr: false,
		},
		{
			testName:  "no error, VaultPolicyBinding doesn't exist",
			vPolicy:   nil,
			vPBind:    nil,
			expectErr: false,
		},
		{
			testName:  "error, invalid VaultPolicyBinding",
			vPolicy:   nil,
			vPBind:    simpleVaultPolicyBinding(),
			expectErr: true,
		},
	}

	for _, c := range cases {
		t.Run(c.testName, func(t *testing.T) {
			cs := csfake.NewSimpleClientset()
			pc := cs.PolicyV1alpha1()
			if c.vPolicy != nil {
				_, err := pc.VaultPolicies(c.vPolicy.Namespace).Create(c.vPolicy)
				assert.Nil(t, err)
			} else {
				c.vPolicy = simpleVaultPolicy()
			}
			if c.vPBind != nil {
				_, err := pc.VaultPolicyBindings(c.vPBind.Namespace).Create(c.vPBind)
				assert.Nil(t, err)
			} else {
				c.vPBind = simpleVaultPolicyBinding()
			}

			ctrl := &VaultController{
				kubeClient:       kc,
				extClient:        cs,
				appCatalogClient: appc.AppcatalogV1alpha1(),
			}

			err := ctrl.finalizePolicyBinding(c.vPBind)
			if c.expectErr {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func TestRunPolicyBindingFinalizer(t *testing.T) {
	srv := NewFakeVaultServer()
	defer srv.Close()
	vApp := vaultAppBinding(srv.URL, vaultTokenSecret().Name)
	appc := appcatfake.NewSimpleClientset(vApp)
	ctrl := &VaultController{
		extClient: csfake.NewSimpleClientset(simpleVaultPolicyBinding(),
			validVaultPolicy(vApp),
			validVaultPolicyBinding(validVaultPolicy(vApp).Name)),
		kubeClient:       kfake.NewSimpleClientset(vaultTokenSecret()),
		appCatalogClient: appc.AppcatalogV1alpha1(),
		finalizerInfo:    NewMapFinalizer(),
	}
	ctrl.finalizerInfo.Add(simpleVaultPolicyBinding().GetKey())

	cases := []struct {
		testName  string
		vPBind    *policyapi.VaultPolicyBinding
		completed bool
	}{
		{
			testName:  "remove finalizer successfully, valid VaultPolicyBinding",
			vPBind:    validVaultPolicyBinding(validVaultPolicy(vApp).Name),
			completed: true,
		},
		{
			testName:  "remove finalizer successfully, invalid VaultPolicyBinding",
			vPBind:    validVaultPolicyBinding("test"),
			completed: true,
		},
		{
			testName:  "already processing finalizer",
			vPBind:    simpleVaultPolicyBinding(),
			completed: false,
		},
	}

	for _, c := range cases {
		t.Run(c.testName, func(t *testing.T) {
			ctrl.runPolicyBindingFinalizer(c.vPBind, 3*time.Second, 1*time.Second)
			if c.completed {
				assert.Condition(t, func() (success bool) {
					return !ctrl.finalizerInfo.IsAlreadyProcessing(c.vPBind.GetKey())
				}, "IsAlreadyProcessing(key) should be false")

			} else {
				assert.Condition(t, func() (success bool) {
					return ctrl.finalizerInfo.IsAlreadyProcessing(c.vPBind.GetKey())
				}, "IsAlreadyProcessing(key) should be true")
			}
		})
	}
}
