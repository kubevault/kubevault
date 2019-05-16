package controller

import (
	"fmt"
	"testing"

	"github.com/kubevault/operator/pkg/vault/role/azure"

	api "github.com/kubevault/operator/apis/engine/v1alpha1"
	opfake "github.com/kubevault/operator/client/clientset/versioned/fake"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kfake "k8s.io/client-go/kubernetes/fake"
	appcat "kmodules.xyz/custom-resources/apis/appcatalog/v1alpha1"
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

func (f *fakeAzureRole) DeleteRole(name string) error {
	return nil
}

func TestAzureRole_reconcileAzureRole(t *testing.T) {

	aRole := &api.AzureRole{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "azure-role-879",
			Namespace: "demo",
		},
		Spec: api.AzureRoleSpec{
			AuthManagerRef: &appcat.AppReference{
				Namespace: "demo",
				Name:      "test-212321",
			},
			Config: &api.AzureConfig{
				CredentialSecret: "azure-cred",
				Environment:      "AzurePublicCloud",
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
			testName: "Failed to enable Azure",
			azureRClient: &fakeAzureRole{
				errorOccurredInEnableAzure: true,
			},
			azureRole:          aRole,
			hasStatusCondition: true,
			expectedErr:        true,
		},
		{
			testName: "Failed to create config",
			azureRClient: &fakeAzureRole{
				errorOccurredInCreateConfig: true,
			},
			azureRole:          aRole,
			hasStatusCondition: true,
			expectedErr:        true,
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

	for _, test := range testData {
		t.Run(test.testName, func(t *testing.T) {

			c := &VaultController{
				kubeClient: kfake.NewSimpleClientset(),
				extClient:  opfake.NewSimpleClientset(),
			}

			_, err := c.extClient.EngineV1alpha1().AzureRoles(test.azureRole.Namespace).Create(test.azureRole)
			assert.Nil(t, err)
			err = c.reconcileAzureRole(test.azureRClient, test.azureRole)

			if test.expectedErr {

				if assert.NotNil(t, err) {
					if test.hasStatusCondition {
						p, err2 := c.extClient.EngineV1alpha1().AzureRoles(test.azureRole.Namespace).Get(test.azureRole.Name, metav1.GetOptions{})

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

					p, err2 := c.extClient.EngineV1alpha1().AzureRoles(test.azureRole.Namespace).Get(test.azureRole.Name, metav1.GetOptions{})

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

			err = c.extClient.EngineV1alpha1().AzureRoles(test.azureRole.Namespace).Delete(test.azureRole.Name, &metav1.DeleteOptions{})
			assert.Nil(t, err)
		})

	}
}
