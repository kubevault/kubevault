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
	"testing"
	"time"

	policyapi "kubevault.dev/operator/apis/policy/v1alpha1"
	csfake "kubevault.dev/operator/client/clientset/versioned/fake"
	pbinding "kubevault.dev/operator/pkg/vault/policybinding"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kfake "k8s.io/client-go/kubernetes/fake"
	appcatfake "kmodules.xyz/custom-resources/client/clientset/versioned/fake"
)

type fakePBind struct {
	errInEnsure bool
}

func (f *fakePBind) Ensure(n *policyapi.VaultPolicyBinding) error {
	if f.errInEnsure {
		return errors.New("error")
	}
	return nil
}

func (f *fakePBind) Delete(n *policyapi.VaultPolicyBinding) error {
	return nil
}

func simpleVaultPolicyBinding() *policyapi.VaultPolicyBinding {
	return &policyapi.VaultPolicyBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "simple",
			Namespace:  "test",
			Finalizers: []string{VaultPolicyBindingFinalizer},
		},
		Spec: policyapi.VaultPolicyBindingSpec{
			SubjectRef: policyapi.SubjectRef{
				Kubernetes: &policyapi.KubernetesSubjectRef{},
				AppRole:    &policyapi.AppRoleSubjectRef{},
			},
		},
	}
}

func validVaultPolicyBinding(policyName string) *policyapi.VaultPolicyBinding {
	p := simpleVaultPolicyBinding()
	p.Name = "valid"
	p.Spec.VaultRef.Name = "vault"
	p.Spec.Policies = []policyapi.PolicyIdentifier{
		{
			Name: policyName,
		},
	}
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
				p, err := ctrl.extClient.PolicyV1alpha1().VaultPolicyBindings(c.vPBind.Namespace).Get(context.TODO(), c.vPBind.Name, metav1.GetOptions{})
				if assert.Nil(t, err) {
					assert.Condition(t, func() (success bool) {
						return c.expectStatus == string(p.Status.Phase)
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
				_, err := pc.VaultPolicies(c.vPolicy.Namespace).Create(context.TODO(), c.vPolicy, metav1.CreateOptions{})
				assert.Nil(t, err)
			} else {
				c.vPolicy = simpleVaultPolicy()
			}

			if c.vPBind != nil {
				_, err := pc.VaultPolicyBindings(c.vPBind.Namespace).Create(context.TODO(), c.vPBind, metav1.CreateOptions{})
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
