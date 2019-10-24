package controller

import (
	"fmt"
	"testing"

	api "kubevault.dev/operator/apis/engine/v1alpha1"
	vfake "kubevault.dev/operator/client/clientset/versioned/fake"
	"kubevault.dev/operator/pkg/vault/engine"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kfake "k8s.io/client-go/kubernetes/fake"
)

type fakeSecretEngine struct {
	errorOccurredInCreatePolicy   bool
	errorOccurredInUpdateAuthRole bool
	errorOccurredInEnableSE       bool
	errorOccurredInCreateConfig   bool
}

func (f *fakeSecretEngine) IsSecretEngineEnabled() (bool, error) {
	return true, nil
}

func (f *fakeSecretEngine) EnableSecretEngine() error {
	if f.errorOccurredInEnableSE {
		return fmt.Errorf("error enabling secret engine")
	}
	return nil
}

func (f *fakeSecretEngine) CreatePolicy() error {
	if f.errorOccurredInCreatePolicy {
		return fmt.Errorf("error creating policy")
	}
	return nil
}

func (f *fakeSecretEngine) UpdateAuthRole() error {
	if f.errorOccurredInUpdateAuthRole {
		return fmt.Errorf("error updating auth role")
	}
	return nil
}

func (f *fakeSecretEngine) CreateConfig() error {
	if f.errorOccurredInCreateConfig {
		return fmt.Errorf("error creating config")
	}
	return nil
}

func TestVaultController_reconcileSecretEngine(t *testing.T) {

	secretEng := &api.SecretEngine{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "demo-secret-engine",
			Namespace: "demo",
		},
		Spec: api.SecretEngineSpec{
			VaultRef: corev1.LocalObjectReference{},
			Path:     "",
			SecretEngineConfiguration: api.SecretEngineConfiguration{
				GCP: &api.GCPConfiguration{
					CredentialSecret: "secret-1232123",
				},
			},
		},
	}

	tests := []struct {
		name               string
		secretEngineClient engine.EngineInterface
		secretEngine       *api.SecretEngine
		wantErr            bool
	}{
		{
			name:               "Successful operation",
			secretEngineClient: &fakeSecretEngine{},
			secretEngine:       secretEng,
			wantErr:            false,
		},
		{
			name: "CreatePolicy failed",
			secretEngineClient: &fakeSecretEngine{
				errorOccurredInCreatePolicy: true,
			},
			secretEngine: secretEng,
			wantErr:      true,
		},
		{
			name: "UpdateAuthRole failed",
			secretEngineClient: &fakeSecretEngine{
				errorOccurredInUpdateAuthRole: true,
			},
			secretEngine: secretEng,
			wantErr:      true,
		},
		{
			name:               "EnableSecretEngine failed",
			secretEngineClient: &fakeSecretEngine{errorOccurredInEnableSE: true},
			secretEngine:       secretEng,
			wantErr:            true,
		},
		{
			name:               "CreateConfig failed",
			secretEngineClient: &fakeSecretEngine{errorOccurredInCreateConfig: true},
			secretEngine:       secretEng,
			wantErr:            true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			c := &VaultController{
				kubeClient: kfake.NewSimpleClientset(),
				extClient:  vfake.NewSimpleClientset(),
			}
			_, err := c.extClient.EngineV1alpha1().SecretEngines(tt.secretEngine.Namespace).Create(tt.secretEngine)
			assert.Nil(t, err)

			if err := c.reconcileSecretEngine(tt.secretEngineClient, tt.secretEngine); (err != nil) != tt.wantErr {
				t.Errorf("reconcileSecretEngine() error = %v, wantErr %v", err, tt.wantErr)
			} else {
				se, err2 := c.extClient.EngineV1alpha1().SecretEngines(tt.secretEngine.Namespace).Get(tt.secretEngine.Name, metav1.GetOptions{})
				assert.Nil(t, err2)
				if tt.wantErr {
					assert.Condition(t, func() (success bool) {
						return len(se.Status.Conditions) != 0
					}, "Should have status.conditions")
				} else {
					assert.Condition(t, func() (success bool) {
						return len(se.Status.Conditions) == 0 && se.Status.Phase == SecretEnginePhaseSuccess
					}, "Shouldn't have status.conditions")
				}
			}

			err = c.extClient.EngineV1alpha1().SecretEngines(tt.secretEngine.Namespace).Delete(tt.secretEngine.Name, &metav1.DeleteOptions{})
			assert.Nil(t, err)
		})
	}
}
