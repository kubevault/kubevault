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

	"kubevault.dev/apimachinery/apis"
	api "kubevault.dev/apimachinery/apis/engine/v1alpha1"
	patchutil "kubevault.dev/apimachinery/client/clientset/versioned/typed/engine/v1alpha1/util"
	"kubevault.dev/operator/pkg/vault/role/database"

	"github.com/golang/glog"
	"github.com/pkg/errors"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	kmapi "kmodules.xyz/client-go/api/v1"
	core_util "kmodules.xyz/client-go/core/v1"
	"kmodules.xyz/client-go/tools/queue"
)

const (
	ElasticsearchRolePhaseSuccess    api.ElasticsearchRolePhase = "Success"
	ElasticsearchRolePhaseProcessing api.ElasticsearchRolePhase = "Processing"
)

func (c *VaultController) initElasticsearchRoleWatcher() {
	c.esRoleInformer = c.extInformerFactory.Engine().V1alpha1().ElasticsearchRoles().Informer()
	c.esRoleQueue = queue.New(api.ResourceKindElasticsearchRole, c.MaxNumRequeues, c.NumThreads, c.runElasticsearchRoleInjector)
	c.esRoleInformer.AddEventHandler(queue.NewReconcilableHandler(c.esRoleQueue.GetQueue()))
	c.esRoleLister = c.extInformerFactory.Engine().V1alpha1().ElasticsearchRoles().Lister()
}

func (c *VaultController) runElasticsearchRoleInjector(key string) error {
	// key := name/namespace
	obj, exist, err := c.esRoleInformer.GetIndexer().GetByKey(key)
	if err != nil {
		glog.Errorf("Fetching object with key %s from store failed with %v", key, err)
		return err
	}

	if !exist {
		glog.Warningf("ElasticsearchRole %s does not exist anymore", key)

	} else {
		role := obj.(*api.ElasticsearchRole).DeepCopy()

		glog.Infof("Sync/Add/Update for ElasticsearchRole %s/%s", role.Namespace, role.Name)
		// DeletionTimestamp has been set, if HasFinalizer, then run ESRoleFinalizer.
		if role.DeletionTimestamp != nil {
			if core_util.HasFinalizer(role.ObjectMeta, apis.Finalizer) {
				return c.runElasticsearchRoleFinalizer(role)
			}
		} else {
			if !core_util.HasFinalizer(role.ObjectMeta, apis.Finalizer) {
				// Add finalizer
				_, _, err := patchutil.PatchElasticsearchRole(context.TODO(), c.extClient.EngineV1alpha1(), role, func(in *api.ElasticsearchRole) *api.ElasticsearchRole {
					in.ObjectMeta = core_util.AddFinalizer(in.ObjectMeta, apis.Finalizer)
					return in
				}, metav1.PatchOptions{})
				if err != nil {
					return errors.Wrapf(err, "failed to set ElasticsearchRole finalizer for %s/%s", role.Namespace, role.Name)
				}
			}

			// Conditions are empty, when the ElasticsearchRole obj is enqueued for the first time.
			// Set status.phase to "Processing".
			if role.Status.Conditions == nil {
				// using a "newRole" var is preferred here. If we directly set the "role" here, we may get nil pointer exception in case of error.
				newRole, err := patchutil.UpdateElasticsearchRoleStatus(
					context.TODO(),
					c.extClient.EngineV1alpha1(),
					role.ObjectMeta,
					func(status *api.ElasticsearchRoleStatus) *api.ElasticsearchRoleStatus {
						status.Phase = ElasticsearchRolePhaseProcessing
						return status
					},
					metav1.UpdateOptions{},
				)
				if err != nil {
					return errors.Wrapf(err, "failed to update status for ElasticsearchRole: %s/%s", role.Namespace, role.Name)
				}
				role = newRole
			}

			rClient, err := database.NewDatabaseRoleForElasticsearch(c.kubeClient, c.appCatalogClient, role)
			if err != nil {
				return err
			}

			err = c.reconcileElasticsearchRole(rClient, role)
			if err != nil {
				return errors.Wrapf(err, "failed to reconcile ElasticsearchRole: %s/%s", role.Namespace, role.Name)
			}
		}
	}
	return nil
}

// Will do:
//	For vault:
// 	  - configure a role that maps a name in Vault to an SQL statement to execute to create the database credential.
//    - sync role
//	  - revoke previous lease of all the respective elasticsearchRoleBinding and reissue a new lease
func (c *VaultController) reconcileElasticsearchRole(rClient database.DatabaseRoleInterface, role *api.ElasticsearchRole) error {
	// create role
	err := rClient.CreateRole()
	if err != nil {
		_, err2 := patchutil.UpdateElasticsearchRoleStatus(
			context.TODO(),
			c.extClient.EngineV1alpha1(),
			role.ObjectMeta,
			func(status *api.ElasticsearchRoleStatus) *api.ElasticsearchRoleStatus {
				status.Conditions = kmapi.SetCondition(status.Conditions, kmapi.Condition{
					Type:    kmapi.ConditionFailed,
					Status:  core.ConditionTrue,
					Reason:  "FailedToCreateRole",
					Message: err.Error(),
				})
				return status
			},
			metav1.UpdateOptions{},
		)
		return utilerrors.NewAggregate([]error{err2, errors.Wrap(err, "failed to create role")})
	}

	_, err = patchutil.UpdateElasticsearchRoleStatus(
		context.TODO(),
		c.extClient.EngineV1alpha1(),
		role.ObjectMeta,
		func(status *api.ElasticsearchRoleStatus) *api.ElasticsearchRoleStatus {
			status.Phase = ElasticsearchRolePhaseSuccess
			status.ObservedGeneration = role.Generation
			status.Conditions = kmapi.SetCondition(status.Conditions, kmapi.Condition{
				Type:    kmapi.ConditionAvailable,
				Status:  core.ConditionTrue,
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

	glog.Infof("Successfully processed ElasticsearchRole: %s/%s", role.Namespace, role.Name)
	return nil
}

func (c *VaultController) runElasticsearchRoleFinalizer(role *api.ElasticsearchRole) error {
	glog.Infof("Processing finalizer for ElasticsearchRole: %s/%s", role.Namespace, role.Name)

	rClient, err := database.NewDatabaseRoleForElasticsearch(c.kubeClient, c.appCatalogClient, role)
	// The error could be generated for:
	//   - invalid vaultRef in the spec
	// In this case, the operator should be able to delete the ElasticsearchRole (ie. remove finalizer).
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
		glog.Warningf("Skipping cleanup for ElasticsearchRole: %s/%s with error: %v", role.Namespace, role.Name, err)
	}

	// remove finalizer
	_, _, err = patchutil.PatchElasticsearchRole(context.TODO(), c.extClient.EngineV1alpha1(), role, func(in *api.ElasticsearchRole) *api.ElasticsearchRole {
		in.ObjectMeta = core_util.RemoveFinalizer(in.ObjectMeta, apis.Finalizer)
		return in
	}, metav1.PatchOptions{})
	if err != nil {
		return errors.Wrapf(err, "failed to remove finalizer for ElasticsearchRole: %s/%s", role.Namespace, role.Name)
	}

	glog.Infof("Removed finalizer for ElasticsearchRole: %s/%s", role.Namespace, role.Name)
	return nil
}
