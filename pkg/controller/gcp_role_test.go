package controller

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kfake "k8s.io/client-go/kubernetes/fake"
	api "kubevault.dev/operator/apis/engine/v1alpha1"
	opfake "kubevault.dev/operator/client/clientset/versioned/fake"
	"kubevault.dev/operator/pkg/vault/role/gcp"
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

func (f *fakeGCPRole) DeleteRole(name string) error {
	return nil
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
			testName: "Failed to enable GCP",
			gcpRClient: &fakeGCPRole{
				errorOccurredInEnableGCP: true,
			},
			gcpRole:            gRole,
			hasStatusCondition: true,
			expectedErr:        true,
		},
		{
			testName: "Failed to create config",
			gcpRClient: &fakeGCPRole{
				errorOccurredInCreateConfig: true,
			},
			gcpRole:            gRole,
			hasStatusCondition: true,
			expectedErr:        true,
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

	for _, test := range testData {
		t.Run(test.testName, func(t *testing.T) {

			c := &VaultController{
				kubeClient: kfake.NewSimpleClientset(),
				extClient:  opfake.NewSimpleClientset(),
			}

			_, err := c.extClient.EngineV1alpha1().GCPRoles(test.gcpRole.Namespace).Create(test.gcpRole)
			assert.Nil(t, err)

			err = c.reconcileGCPRole(test.gcpRClient, test.gcpRole)

			if test.expectedErr {

				if assert.NotNil(t, err) {
					if test.hasStatusCondition {
						p, err2 := c.extClient.EngineV1alpha1().GCPRoles(test.gcpRole.Namespace).Get(test.gcpRole.Name, metav1.GetOptions{})
						if assert.Nil(t, err2) {
							assert.Condition(t, func() (success bool) {
								if len(p.Status.Conditions) != 0 {
									return true
								} else {
									return false
								}
							}, "Should have status.conditions")
						}
					}
				}
			} else {
				if assert.Nil(t, err) {

					p, err2 := c.extClient.EngineV1alpha1().GCPRoles(test.gcpRole.Namespace).Get(test.gcpRole.Name, metav1.GetOptions{})
					if assert.Nil(t, err2) {
						assert.Condition(t, func() (success bool) {
							if len(p.Status.Conditions) == 0 {
								return true
							} else {
								return false
							}
						}, "Should not have status.conditions")
					}

				}

			}

			err = c.extClient.EngineV1alpha1().GCPRoles(test.gcpRole.Namespace).Delete(test.gcpRole.Name, &metav1.DeleteOptions{})
			assert.Nil(t, err)
		})

	}
}
