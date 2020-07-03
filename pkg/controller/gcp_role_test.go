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
	"kubevault.dev/operator/pkg/vault/role/gcp"

	"github.com/stretchr/testify/assert"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kfake "k8s.io/client-go/kubernetes/fake"
	kmapi "kmodules.xyz/client-go/api/v1"
	meta_util "kmodules.xyz/client-go/meta"
)

type fakeGCPRole struct {
	errorOccurredInEnableGCP    bool
	errorOccurredInCreateConfig bool
	errorOccurredInCreateRole   bool
}

func (f *fakeGCPRole) EnableGCP() error {
	if f.errorOccurredInEnableGCP {
		return fmt.Errorf("error enabling GCP")
	}
	return nil

}

func (f *fakeGCPRole) CreateConfig() error {
	if f.errorOccurredInCreateConfig {
		return fmt.Errorf("error creating config")
	}
	return nil
}

func (f *fakeGCPRole) CreateRole() error {
	if f.errorOccurredInCreateRole {
		return fmt.Errorf("error creating role")
	}
	return nil
}

func (f *fakeGCPRole) IsGCPEnabled() (bool, error) {
	return true, nil
}

func (f *fakeGCPRole) DeleteRole(name string) (int, error) {
	return 0, nil
}

func TestGCPRole_reconcileGCPRole(t *testing.T) {

	gRole := &api.GCPRole{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "gcp-role-879",
			Namespace: "demo",
		},
		Spec: api.GCPRoleSpec{
			VaultRef: core.LocalObjectReference{
				Name: "test-212321",
			},
			SecretType: "access_token",
			Project:    "ackube",
			Bindings: `resource "//cloudresourcemanager.googleapis.com/projects/ackube" {
	        roles = ["roles/viewer"]
	      }`,
			TokenScopes: []string{"https://www.googleapis.com/auth/cloud-platform"},
		},
	}

	testData := []struct {
		testName           string
		gcpRClient         gcp.GCPRoleInterface
		gcpRole            *api.GCPRole
		hasStatusCondition bool
		expectedErr        bool
	}{
		{
			testName:           "Initial stage with no error",
			gcpRole:            gRole,
			gcpRClient:         &fakeGCPRole{},
			hasStatusCondition: false,
			expectedErr:        false,
		},
		{
			testName: "Failed to create role",
			gcpRClient: &fakeGCPRole{
				errorOccurredInCreateRole: true,
			},
			gcpRole:            gRole,
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

			_, err := c.extClient.EngineV1alpha1().GCPRoles(test.gcpRole.Namespace).Create(context.TODO(), test.gcpRole, metav1.CreateOptions{})
			assert.Nil(t, err)

			err = c.reconcileGCPRole(test.gcpRClient, test.gcpRole)

			if test.expectedErr {

				if assert.NotNil(t, err) {
					if test.hasStatusCondition {
						p, err2 := c.extClient.EngineV1alpha1().GCPRoles(test.gcpRole.Namespace).Get(context.TODO(), test.gcpRole.Name, metav1.GetOptions{})
						if assert.Nil(t, err2) {
							assert.Condition(t, func() (success bool) {
								return len(p.Status.Conditions) > 0 &&
									kmapi.IsConditionTrue(p.Status.Conditions, kmapi.ConditionFailure) &&
									!kmapi.HasCondition(p.Status.Conditions, kmapi.ConditionAvailable)
							}, "Should have status.conditions")
						}
					}
				}
			} else {
				if assert.Nil(t, err) {

					p, err2 := c.extClient.EngineV1alpha1().GCPRoles(test.gcpRole.Namespace).Get(context.TODO(), test.gcpRole.Name, metav1.GetOptions{})
					if assert.Nil(t, err2) {
						assert.Condition(t, func() (success bool) {
							return p.Status.Phase == GCPRolePhaseSuccess &&
								len(p.Status.Conditions) > 0 &&
								!kmapi.HasCondition(p.Status.Conditions, kmapi.ConditionFailure) &&
								kmapi.IsConditionTrue(p.Status.Conditions, kmapi.ConditionAvailable)
						}, "Should not have status.conditions")
					}

				}

			}

			err = c.extClient.EngineV1alpha1().GCPRoles(test.gcpRole.Namespace).Delete(context.TODO(), test.gcpRole.Name, meta_util.DeleteInForeground())
			assert.Nil(t, err)
		})

	}
}
