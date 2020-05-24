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

func TestUserManagerController_reconcileMongoDBRole(t *testing.T) {
	mRole := api.MongoDBRole{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "m-role",
			Namespace:  "m",
			Generation: 0,
		},
		Spec: api.MongoDBRoleSpec{
			VaultRef:    core.LocalObjectReference{Name: "test121"},
			DatabaseRef: &appcat.AppReference{Name: "test"},
		},
	}

	testData := []struct {
		testName           string
		mgRole             api.MongoDBRole
		dbRClient          database.DatabaseRoleInterface
		hasStatusCondition bool
		expectedErr        bool
	}{
		{
			testName:           "initial stage, no error",
			mgRole:             mRole,
			dbRClient:          &fakeDRole{},
			expectedErr:        false,
			hasStatusCondition: false,
		},
		{
			testName: "initial stage, failed to create database role",
			mgRole:   mRole,
			dbRClient: &fakeDRole{
				errorOccurredInCreateRole: true,
			},
			expectedErr:        true,
			hasStatusCondition: true,
		},
		{
			testName: "update role, successfully updated database role",
			mgRole: func(p api.MongoDBRole) api.MongoDBRole {
				p.Generation = 2
				p.Status.ObservedGeneration = 1
				return p
			}(mRole),
			dbRClient:          &fakeDRole{},
			expectedErr:        false,
			hasStatusCondition: false,
		},
		{
			testName: "update role, failed to update database role",
			mgRole: func(p api.MongoDBRole) api.MongoDBRole {
				p.Generation = 2
				p.Status.ObservedGeneration = 1
				return p
			}(mRole),
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

			_, err := c.extClient.EngineV1alpha1().MongoDBRoles(test.mgRole.Namespace).Create(context.TODO(), &test.mgRole, metav1.CreateOptions{})
			if !assert.Nil(t, err) {
				return
			}

			err = c.reconcileMongoDBRole(test.dbRClient, &test.mgRole)
			if test.expectedErr {
				if assert.NotNil(t, err) {
					if test.hasStatusCondition {
						p, err2 := c.extClient.EngineV1alpha1().MongoDBRoles(test.mgRole.Namespace).Get(context.TODO(), test.mgRole.Name, metav1.GetOptions{})
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
					p, err2 := c.extClient.EngineV1alpha1().MongoDBRoles(test.mgRole.Namespace).Get(context.TODO(), test.mgRole.Name, metav1.GetOptions{})
					if assert.Nil(t, err2) {
						assert.Condition(t, func() (success bool) {
							return p.Status.Phase == MongoDBRolePhaseSuccess &&
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
