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
	"time"

	api "kubevault.dev/operator/apis/engine/v1alpha1"
	cs "kubevault.dev/operator/client/clientset/versioned/fake"
	dbinformers "kubevault.dev/operator/client/informers/externalversions"
	"kubevault.dev/operator/pkg/vault/role/database"

	"github.com/stretchr/testify/assert"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kfake "k8s.io/client-go/kubernetes/fake"
	kmapi "kmodules.xyz/client-go/api/v1"
	appcat "kmodules.xyz/custom-resources/apis/appcatalog/v1alpha1"
)

type fakeDRole struct {
	errorOccurredInEnableDatabase bool
	errorOccurredInCreateConfig   bool
	errorOccurredInCreateRole     bool
}

func (f *fakeDRole) EnableDatabase() error {
	if f.errorOccurredInEnableDatabase {
		return fmt.Errorf("error")
	}
	return nil
}

func (f *fakeDRole) IsDatabaseEnabled() (bool, error) {
	return true, nil
}

func (f *fakeDRole) DeleteRole(name string) (int, error) {
	return 0, nil
}

func (f *fakeDRole) CreateConfig() error {
	if f.errorOccurredInCreateConfig {
		return fmt.Errorf("error")
	}
	return nil
}

func (f *fakeDRole) CreateRole() error {
	if f.errorOccurredInCreateRole {
		return fmt.Errorf("error")
	}
	return nil
}

func TestUserManagerController_reconcilePostgresRole(t *testing.T) {
	pRole := api.PostgresRole{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "pg-role",
			Namespace:  "pg",
			Generation: 0,
		},
		Spec: api.PostgresRoleSpec{
			VaultRef: core.LocalObjectReference{},
			DatabaseRef: &appcat.AppReference{
				Name: "test",
			},
		},
	}

	testData := []struct {
		testName           string
		pgRole             api.PostgresRole
		dbRClient          database.DatabaseRoleInterface
		hasStatusCondition bool
		expectedErr        bool
	}{
		{
			testName:           "initial stage, no error",
			pgRole:             pRole,
			dbRClient:          &fakeDRole{},
			expectedErr:        false,
			hasStatusCondition: false,
		},
		{
			testName: "initial stage, failed to create database role",
			pgRole:   pRole,
			dbRClient: &fakeDRole{
				errorOccurredInCreateRole: true,
			},
			expectedErr:        true,
			hasStatusCondition: true,
		},
		{
			testName: "update role, successfully updated database role",
			pgRole: func(p api.PostgresRole) api.PostgresRole {
				p.Generation = 2
				p.Status.ObservedGeneration = 1
				return p
			}(pRole),
			dbRClient:          &fakeDRole{},
			expectedErr:        false,
			hasStatusCondition: false,
		},
		{
			testName: "update role, failed to update database role",
			pgRole: func(p api.PostgresRole) api.PostgresRole {
				p.Generation = 2
				p.Status.ObservedGeneration = 1
				return p
			}(pRole),
			dbRClient: &fakeDRole{
				errorOccurredInCreateRole: true,
			},
			expectedErr:        true,
			hasStatusCondition: true,
		},
	}

	for idx := range testData {
		test := testData[idx]
		t.Run(test.testName, func(t *testing.T) {
			c := &VaultController{
				kubeClient: kfake.NewSimpleClientset(),
				extClient:  cs.NewSimpleClientset(),
			}
			c.extInformerFactory = dbinformers.NewSharedInformerFactory(c.extClient, time.Minute*10)

			_, err := c.extClient.EngineV1alpha1().PostgresRoles(test.pgRole.Namespace).Create(context.TODO(), &test.pgRole, metav1.CreateOptions{})
			if !assert.Nil(t, err) {
				return
			}

			err = c.reconcilePostgresRole(test.dbRClient, &test.pgRole)
			if test.expectedErr {
				if assert.NotNil(t, err) {
					if test.hasStatusCondition {
						p, err2 := c.extClient.EngineV1alpha1().PostgresRoles(test.pgRole.Namespace).Get(context.TODO(), test.pgRole.Name, metav1.GetOptions{})
						if assert.Nil(t, err2) {
							assert.Condition(t, func() (success bool) {
								return len(p.Status.Conditions) > 0 &&
									kmapi.IsConditionTrue(p.Status.Conditions, kmapi.ConditionFailure) &&
									!kmapi.HasCondition(p.Status.Conditions, kmapi.ConditionAvailable)
							}, "should have status.conditions")
						}
					}
				}
			} else {
				if assert.Nil(t, err) {
					p, err2 := c.extClient.EngineV1alpha1().PostgresRoles(test.pgRole.Namespace).Get(context.TODO(), test.pgRole.Name, metav1.GetOptions{})
					if assert.Nil(t, err2) {
						assert.Condition(t, func() (success bool) {
							return p.Status.Phase == PostgresRolePhaseSuccess &&
								len(p.Status.Conditions) > 0 &&
								!kmapi.HasCondition(p.Status.Conditions, kmapi.ConditionFailure) &&
								kmapi.IsConditionTrue(p.Status.Conditions, kmapi.ConditionAvailable)
						}, "should not have status.conditions")
					}
				}
			}
		})
	}

}
