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
	"fmt"
	"testing"

	api "kubevault.dev/operator/apis/engine/v1alpha1"
	opfake "kubevault.dev/operator/client/clientset/versioned/fake"
	"kubevault.dev/operator/pkg/vault/role/azure"

	"github.com/stretchr/testify/assert"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kfake "k8s.io/client-go/kubernetes/fake"
	kmapi "kmodules.xyz/client-go/api/v1"
	meta_util "kmodules.xyz/client-go/meta"
)

type fakeAzureRole struct {
	errorOccurredInEnableAzure  bool
	errorOccurredInCreateConfig bool
	errorOccurredInCreateRole   bool
}

func (f *fakeAzureRole) EnableAzure() error {
	if f.errorOccurredInEnableAzure {
		return fmt.Errorf("error enabling Azure")
	}
	return nil

}

func (f *fakeAzureRole) CreateConfig() error {
	if f.errorOccurredInCreateConfig {
		return fmt.Errorf("error creating config")
	}
	return nil
}

func (f *fakeAzureRole) CreateRole() error {
	if f.errorOccurredInCreateRole {
		return fmt.Errorf("error creating role")
	}
	return nil
}

func (f *fakeAzureRole) IsAzureEnabled() (bool, error) {
	return true, nil
}

func (f *fakeAzureRole) DeleteRole(name string) (int, error) {
	return 0, nil
}

func TestAzureRole_reconcileAzureRole(t *testing.T) {

	aRole := &api.AzureRole{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "azure-role-879",
			Namespace: "demo",
		},
		Spec: api.AzureRoleSpec{
			VaultRef: core.LocalObjectReference{
				Name: "test-212321",
			},
			AzureRoles:          "[{}]",
			ApplicationObjectID: "443-324-3432-44",
			TTL:                 "0h",
			MaxTTL:              "1h",
		},
	}

	testData := []struct {
		testName           string
		azureRClient       azure.AzureRoleInterface
		azureRole          *api.AzureRole
		hasStatusCondition bool
		expectedErr        bool
	}{
		{
			testName:           "Initial stage with no error",
			azureRole:          aRole,
			azureRClient:       &fakeAzureRole{},
			hasStatusCondition: false,
			expectedErr:        false,
		},
		{
			testName: "Failed to create role",
			azureRClient: &fakeAzureRole{
				errorOccurredInCreateRole: true,
			},
			azureRole:          aRole,
			hasStatusCondition: true,
			expectedErr:        true,
		},
	}

	for idx := range testData {
		test := testData[idx]
		t.Run(test.testName, func(t *testing.T) {

			c := &VaultController{
				kubeClient: kfake.NewSimpleClientset(),
				extClient:  opfake.NewSimpleClientset(),
			}

			_, err := c.extClient.EngineV1alpha1().AzureRoles(test.azureRole.Namespace).Create(context.TODO(), test.azureRole, metav1.CreateOptions{})
			assert.Nil(t, err)
			err = c.reconcileAzureRole(test.azureRClient, test.azureRole)

			if test.expectedErr {

				if assert.NotNil(t, err) {
					if test.hasStatusCondition {
						p, err2 := c.extClient.EngineV1alpha1().AzureRoles(test.azureRole.Namespace).Get(context.TODO(), test.azureRole.Name, metav1.GetOptions{})

						if assert.Nil(t, err2) {
							assert.Condition(t, func() (success bool) {
								return len(p.Status.Conditions) > 0 &&
									kmapi.IsConditionTrue(p.Status.Conditions, kmapi.ConditionFailed) &&
									!kmapi.HasCondition(p.Status.Conditions, kmapi.ConditionAvailable)
							}, "Should have status.conditions")
						}
					}
				}
			} else {
				if assert.Nil(t, err) {

					p, err2 := c.extClient.EngineV1alpha1().AzureRoles(test.azureRole.Namespace).Get(context.TODO(), test.azureRole.Name, metav1.GetOptions{})

					if assert.Nil(t, err2) {
						assert.Condition(t, func() (success bool) {
							return p.Status.Phase == AzureRolePhaseSuccess &&
								len(p.Status.Conditions) > 0 &&
								!kmapi.HasCondition(p.Status.Conditions, kmapi.ConditionFailed) &&
								kmapi.IsConditionTrue(p.Status.Conditions, kmapi.ConditionAvailable)
						}, "Should not have status.conditions")
					}

				}

			}

			err = c.extClient.EngineV1alpha1().AzureRoles(test.azureRole.Namespace).Delete(context.TODO(), test.azureRole.Name, meta_util.DeleteInForeground())
			assert.Nil(t, err)
		})

	}
}
