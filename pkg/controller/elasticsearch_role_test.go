/*
Copyright AppsCode Inc. and Contributors

Licensed under the AppsCode Community License 1.0.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    https://github.com/appscode/licenses/raw/1.0.0/AppsCode-Community-1.0.0.md

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

	api "kubevault.dev/apimachinery/apis/engine/v1alpha1"
	cs "kubevault.dev/apimachinery/client/clientset/versioned/fake"
	dbinformers "kubevault.dev/apimachinery/client/informers/externalversions"
	"kubevault.dev/operator/pkg/vault/role/database"

	"github.com/stretchr/testify/assert"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kfake "k8s.io/client-go/kubernetes/fake"
	kmapi "kmodules.xyz/client-go/api/v1"
	appcat "kmodules.xyz/custom-resources/apis/appcatalog/v1alpha1"
)

func TestUserManagerController_reconcileElasticsearchRole(t *testing.T) {
	eRole := api.ElasticsearchRole{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "e-role",
			Namespace:  "e",
			Generation: 0,
		},
		Spec: api.ElasticsearchRoleSpec{
			VaultRef:    core.LocalObjectReference{Name: "test121"},
			DatabaseRef: &appcat.AppReference{Name: "test"},
		},
	}

	testData := []struct {
		testName           string
		esRole             api.ElasticsearchRole
		dbRClient          database.DatabaseRoleInterface
		hasStatusCondition bool
		expectedErr        bool
	}{
		{
			testName:           "initial stage, no error",
			esRole:             eRole,
			dbRClient:          &fakeDRole{},
			expectedErr:        false,
			hasStatusCondition: false,
		},
		{
			testName: "initial stage, failed to create database role",
			esRole:   eRole,
			dbRClient: &fakeDRole{
				errorOccurredInCreateRole: true,
			},
			expectedErr:        true,
			hasStatusCondition: true,
		},
		{
			testName: "update role, successfully updated database role",
			esRole: func(p api.ElasticsearchRole) api.ElasticsearchRole {
				p.Generation = 2
				p.Status.ObservedGeneration = 1
				return p
			}(eRole),
			dbRClient:          &fakeDRole{},
			expectedErr:        false,
			hasStatusCondition: false,
		},
		{
			testName: "update role, failed to update database role",
			esRole: func(p api.ElasticsearchRole) api.ElasticsearchRole {
				p.Generation = 2
				p.Status.ObservedGeneration = 1
				return p
			}(eRole),
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

			_, err := c.extClient.EngineV1alpha1().ElasticsearchRoles(test.esRole.Namespace).Create(context.TODO(), &test.esRole, metav1.CreateOptions{})
			if !assert.Nil(t, err) {
				return
			}

			err = c.reconcileElasticsearchRole(test.dbRClient, &test.esRole)
			if test.expectedErr {
				if assert.NotNil(t, err) {
					if test.hasStatusCondition {
						p, err2 := c.extClient.EngineV1alpha1().ElasticsearchRoles(test.esRole.Namespace).Get(context.TODO(), test.esRole.Name, metav1.GetOptions{})
						if assert.Nil(t, err2) {
							assert.Condition(t, func() (success bool) {
								return len(p.Status.Conditions) > 0 &&
									kmapi.IsConditionTrue(p.Status.Conditions, kmapi.ConditionFailed) &&
									!kmapi.HasCondition(p.Status.Conditions, kmapi.ConditionAvailable)
							}, "should have status.conditions")
						}
					}
				}
			} else {
				if assert.Nil(t, err) {
					p, err2 := c.extClient.EngineV1alpha1().ElasticsearchRoles(test.esRole.Namespace).Get(context.TODO(), test.esRole.Name, metav1.GetOptions{})
					if assert.Nil(t, err2) {
						assert.Condition(t, func() (success bool) {
							return p.Status.Phase == ElasticsearchRolePhaseSuccess &&
								len(p.Status.Conditions) > 0 &&
								!kmapi.HasCondition(p.Status.Conditions, kmapi.ConditionFailed) &&
								kmapi.IsConditionTrue(p.Status.Conditions, kmapi.ConditionAvailable)
						}, "should not have status.conditions")
					}
				}
			}
		})
	}

}
