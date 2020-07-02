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
	"kubevault.dev/operator/pkg/vault/role/gcp"

	"github.com/golang/glog"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	kmapi "kmodules.xyz/client-go/api/v1"
	core_util "kmodules.xyz/client-go/core/v1"
	"kmodules.xyz/client-go/tools/queue"
)

const (
	GCPRolePhaseSuccess    api.GCPRolePhase = "Success"
	GCPRolePhaseProcessing api.GCPRolePhase = "Processing"
)

func (c *VaultController) initGCPRoleWatcher() {
	c.gcpRoleInformer = c.extInformerFactory.Engine().V1alpha1().GCPRoles().Informer()
	c.gcpRoleQueue = queue.New(api.ResourceKindGCPRole, c.MaxNumRequeues, c.NumThreads, c.runGCPRoleInjector)
	c.gcpRoleInformer.AddEventHandler(queue.NewReconcilableHandler(c.gcpRoleQueue.GetQueue()))
	c.gcpRoleLister = c.extInformerFactory.Engine().V1alpha1().GCPRoles().Lister()
}

func (c *VaultController) runGCPRoleInjector(key string) error {
	obj, exist, err := c.gcpRoleInformer.GetIndexer().GetByKey(key)
	if err != nil {
		glog.Errorf("Fetching object with key %s from store failed with %v", key, err)
		return err
	}

	if !exist {
		glog.Warningf("GCPRole %s does not exist anymore", key)

	} else {
		role := obj.(*api.GCPRole).DeepCopy()

		glog.Infof("Sync/Add/Update for GCPRole %s/%s", role.Namespace, role.Name)

		if role.DeletionTimestamp != nil {
			if core_util.HasFinalizer(role.ObjectMeta, apis.Finalizer) {
				return c.runGCPRoleFinalizer(role)
			}
		} else {
			if !core_util.HasFinalizer(role.ObjectMeta, apis.Finalizer) {
				// Add finalizer
				_, _, err := patchutil.PatchGCPRole(context.TODO(), c.extClient.EngineV1alpha1(), role, func(in *api.GCPRole) *api.GCPRole {
					in.ObjectMeta = core_util.AddFinalizer(in.ObjectMeta, apis.Finalizer)
					return in
				}, metav1.PatchOptions{})
				if err != nil {
					return errors.Wrapf(err, "failed to add finalizer for GCPRole: %s/%s", role.Namespace, role.Name)
				}
			}

			// Conditions are empty, when the GCPRole obj is enqueued for the first time.
			// Set status.phase to "Processing".
			if role.Status.Conditions == nil {
				newRole, err := patchutil.UpdateGCPRoleStatus(
					context.TODO(),
					c.extClient.EngineV1alpha1(),
					role.ObjectMeta,
					func(status *api.GCPRoleStatus) *api.GCPRoleStatus {
						status.Phase = GCPRolePhaseProcessing
						return status
					},
					metav1.UpdateOptions{},
				)
				if err != nil {
					return errors.Wrapf(err, "failed to update status for GCPRole: %s/%s", role.Namespace, role.Name)
				}
				role = newRole
			}

			rClient, err := gcp.NewGCPRole(c.kubeClient, c.appCatalogClient, role)
			if err != nil {
				return err
			}

			err = c.reconcileGCPRole(rClient, role)
			if err != nil {
				return errors.Wrapf(err, "failed to reconcile GCPRole: %s/%s", role.Namespace, role.Name)
			}
		}
	}
	return nil
}

// Will do:
//	For vault:
// 	  - configure a GCP role
//    - sync role
func (c *VaultController) reconcileGCPRole(rClient gcp.GCPRoleInterface, role *api.GCPRole) error {
	// create role
	err := rClient.CreateRole()
	if err != nil {
		_, err2 := patchutil.UpdateGCPRoleStatus(
			context.TODO(),
			c.extClient.EngineV1alpha1(),
			role.ObjectMeta, func(status *api.GCPRoleStatus) *api.GCPRoleStatus {
				status.Conditions = kmapi.SetCondition(status.Conditions, kmapi.Condition{
					Type:    kmapi.ConditionFailure,
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

	_, err = patchutil.UpdateGCPRoleStatus(
		context.TODO(),
		c.extClient.EngineV1alpha1(),
		role.ObjectMeta, func(status *api.GCPRoleStatus) *api.GCPRoleStatus {
			status.Phase = GCPRolePhaseSuccess
			status.ObservedGeneration = role.Generation
			status.Conditions = kmapi.RemoveCondition(status.Conditions, kmapi.ConditionFailure)
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

	glog.Infof("successfully processed GCPRole: %s/%s", role.Namespace, role.Name)
	return nil
}

func (c *VaultController) runGCPRoleFinalizer(role *api.GCPRole) error {
	glog.Infof("Processing finalizer for GCPRole %s/%s", role.Namespace, role.Name)

	rClient, err := gcp.NewGCPRole(c.kubeClient, c.appCatalogClient, role)
	// The error could be generated for:
	//   - invalid vaultRef in the spec
	// In this case, the operator should be able to delete the GCPRole (ie. remove finalizer).
	// If no error occurred:
	//	- Delete the gcp role created in vault
	if err == nil {
		err = rClient.DeleteRole(role.RoleName())
		if err != nil {
			return errors.Wrap(err, "failed to delete gcp role")
		}
	} else {
		glog.Warningf("skipping cleanup for GCPRole: %s/%s with error: %v", role.Namespace, role.Name, err)
	}

	// remove finalizer
	_, _, err = patchutil.PatchGCPRole(context.TODO(), c.extClient.EngineV1alpha1(), role, func(in *api.GCPRole) *api.GCPRole {
		in.ObjectMeta = core_util.RemoveFinalizer(in.ObjectMeta, apis.Finalizer)
		return in
	}, metav1.PatchOptions{})
	if err != nil {
		return errors.Wrapf(err, "failed to remove finalizer for GCPRole: %s/%s", role.Namespace, role.Name)
	}

	glog.Infof("removed finalizer for GCPRole: %s/%s", role.Namespace, role.Name)
	return nil
}
