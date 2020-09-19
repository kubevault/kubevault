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

	"kubevault.dev/operator/apis"
	api "kubevault.dev/operator/apis/engine/v1alpha1"
	patchutil "kubevault.dev/operator/client/clientset/versioned/typed/engine/v1alpha1/util"
	"kubevault.dev/operator/pkg/vault/role/database"

	"github.com/golang/glog"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	kmapi "kmodules.xyz/client-go/api/v1"
	core_util "kmodules.xyz/client-go/core/v1"
	"kmodules.xyz/client-go/tools/queue"
)

const (
	MongoDBRolePhaseSuccess    api.MongoDBRolePhase = "Success"
	MongoDBRolePhaseProcessing api.MongoDBRolePhase = "Processing"
)

func (c *VaultController) initMongoDBRoleWatcher() {
	c.mgRoleInformer = c.extInformerFactory.Engine().V1alpha1().MongoDBRoles().Informer()
	c.mgRoleQueue = queue.New(api.ResourceKindMongoDBRole, c.MaxNumRequeues, c.NumThreads, c.runMongoDBRoleInjector)
	c.mgRoleInformer.AddEventHandler(queue.NewReconcilableHandler(c.mgRoleQueue.GetQueue()))
	c.mgRoleLister = c.extInformerFactory.Engine().V1alpha1().MongoDBRoles().Lister()
}

func (c *VaultController) runMongoDBRoleInjector(key string) error {
	obj, exist, err := c.mgRoleInformer.GetIndexer().GetByKey(key)
	if err != nil {
		glog.Errorf("Fetching object with key %s from store failed with %v", key, err)
		return err
	}

	if !exist {
		glog.Warningf("MongoDBRole %s does not exist anymore", key)

	} else {
		role := obj.(*api.MongoDBRole).DeepCopy()

		glog.Infof("Sync/Add/Update for MongoDBRole %s/%s", role.Namespace, role.Name)

		if role.DeletionTimestamp != nil {
			if core_util.HasFinalizer(role.ObjectMeta, apis.Finalizer) {
				return c.runMongoDBRoleFinalizer(role)
			}
		} else {
			if !core_util.HasFinalizer(role.ObjectMeta, apis.Finalizer) {
				// Add finalizer
				_, _, err := patchutil.PatchMongoDBRole(context.TODO(), c.extClient.EngineV1alpha1(), role, func(in *api.MongoDBRole) *api.MongoDBRole {
					in.ObjectMeta = core_util.AddFinalizer(in.ObjectMeta, apis.Finalizer)
					return in
				}, metav1.PatchOptions{})
				if err != nil {
					return errors.Wrapf(err, "failed to set MongoDBRole finalizer for %s/%s", role.Namespace, role.Name)
				}
			}

			// Conditions are empty, when the MongoDBRole obj is enqueued for the first time.
			// Set status.phase to "Processing".
			if role.Status.Conditions == nil {
				newRole, err := patchutil.UpdateMongoDBRoleStatus(
					context.TODO(),
					c.extClient.EngineV1alpha1(),
					role.ObjectMeta,
					func(status *api.MongoDBRoleStatus) *api.MongoDBRoleStatus {
						status.Phase = MongoDBRolePhaseProcessing
						return status
					},
					metav1.UpdateOptions{},
				)
				if err != nil {
					return errors.Wrapf(err, "failed to update status for MongoDBRole: %s/%s", role.Namespace, role.Name)
				}
				role = newRole
			}

			rClient, err := database.NewDatabaseRoleForMongodb(c.kubeClient, c.appCatalogClient, role)
			if err != nil {
				return err
			}

			err = c.reconcileMongoDBRole(rClient, role)
			if err != nil {
				return errors.Wrapf(err, "failed to reconcile MongoDBRole: %s/%s", role.Namespace, role.Name)
			}
		}
	}
	return nil
}

// Will do:
//	For vault:
// 	  - configure a role that maps a name in Vault to an SQL statement to execute to create the database credential.
//    - sync role
//	  - revoke previous lease of all the respective mongodbRoleBinding and reissue a new lease
func (c *VaultController) reconcileMongoDBRole(rClient database.DatabaseRoleInterface, role *api.MongoDBRole) error {
	// create role
	err := rClient.CreateRole()
	if err != nil {
		_, err2 := patchutil.UpdateMongoDBRoleStatus(
			context.TODO(),
			c.extClient.EngineV1alpha1(),
			role.ObjectMeta,
			func(status *api.MongoDBRoleStatus) *api.MongoDBRoleStatus {
				status.Conditions = kmapi.SetCondition(status.Conditions, kmapi.Condition{
					Type:    kmapi.ConditionFailed,
					Status:  kmapi.ConditionTrue,
					Reason:  "FailedToCreateRole",
					Message: err.Error(),
				})
				return status
			},
			metav1.UpdateOptions{},
		)
		return utilerrors.NewAggregate([]error{err2, errors.Wrap(err, "failed to create role")})
	}

	_, err = patchutil.UpdateMongoDBRoleStatus(
		context.TODO(),
		c.extClient.EngineV1alpha1(),
		role.ObjectMeta,
		func(status *api.MongoDBRoleStatus) *api.MongoDBRoleStatus {
			status.Phase = MongoDBRolePhaseSuccess
			status.ObservedGeneration = role.Generation
			status.Conditions = kmapi.SetCondition(status.Conditions, kmapi.Condition{
				Type:    kmapi.ConditionAvailable,
				Status:  kmapi.ConditionTrue,
				Reason:  "Provisioned",
				Message: "role is ready to use",
			})
			return status
		},
		metav1.UpdateOptions{},
	)
	if err != nil {
		return err
	}

	glog.Infof("Successfully processed MongoDBRole: %s/%s", role.Namespace, role.Name)
	return nil
}

func (c *VaultController) runMongoDBRoleFinalizer(role *api.MongoDBRole) error {
	glog.Infof("Processing finalizer for MongoDBRole: %s/%s", role.Namespace, role.Name)

	rClient, err := database.NewDatabaseRoleForMongodb(c.kubeClient, c.appCatalogClient, role)
	// The error could be generated for:
	//   - invalid vaultRef in the spec
	// In this case, the operator should be able to delete the MongoDBRole (ie. remove finalizer).
	// If no error occurred:
	//	- Delete the db role created in vault
	if err == nil {
		statusCode, err := rClient.DeleteRole(role.RoleName())
		// For the following errors, the operator should be
		// able to delete the role obj.
		// 	- 400 - Invalid request, missing or invalid data.
		// 	- 403 - Forbidden, your authentication details are either incorrect, you don't have access to this feature, or - if CORS is enabled - you made a cross-origin request from an origin that is not allowed to make such requests.
		//  - 404 - Invalid path. This can both mean that the path truly doesn't exist or that you don't have permission to view a specific path. We use 404 in some cases to avoid state leakage.
		// return error if it is network error.
		if err != nil && (statusCode/100) != 4 {
			return errors.Wrap(err, "failed to delete database role")
		}
	} else {
		glog.Warningf("Skipping cleanup for MongoDBRole: %s/%s with error: %v", role.Namespace, role.Name, err)
	}

	// remove finalizer
	_, _, err = patchutil.PatchMongoDBRole(context.TODO(), c.extClient.EngineV1alpha1(), role, func(in *api.MongoDBRole) *api.MongoDBRole {
		in.ObjectMeta = core_util.RemoveFinalizer(in.ObjectMeta, apis.Finalizer)
		return in
	}, metav1.PatchOptions{})
	if err != nil {
		return errors.Wrapf(err, "failed to remove finalizer for MongoDBRole: %s/%s", role.Namespace, role.Name)
	}

	glog.Infof("Removed finalizer for MongoDBRole: %s/%s", role.Namespace, role.Name)
	return nil
}
