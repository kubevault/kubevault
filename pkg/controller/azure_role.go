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
	"time"

	api "kubevault.dev/operator/apis/engine/v1alpha1"
	patchutil "kubevault.dev/operator/client/clientset/versioned/typed/engine/v1alpha1/util"
	"kubevault.dev/operator/pkg/vault/role/azure"

	"github.com/golang/glog"
	"github.com/pkg/errors"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	kmapi "kmodules.xyz/client-go/api/v1"
	core_util "kmodules.xyz/client-go/core/v1"
	"kmodules.xyz/client-go/tools/queue"
)

const (
	AzureRolePhaseSuccess api.AzureRolePhase = "Success"
	AzureRoleFinalizer    string             = "azurerole.engine.kubevault.com"
)

func (c *VaultController) initAzureRoleWatcher() {
	c.azureRoleInformer = c.extInformerFactory.Engine().V1alpha1().AzureRoles().Informer()
	c.azureRoleQueue = queue.New(api.ResourceKindAzureRole, c.MaxNumRequeues, c.NumThreads, c.runAzureRoleInjector)
	c.azureRoleInformer.AddEventHandler(queue.NewReconcilableHandler(c.azureRoleQueue.GetQueue()))
	c.azureRoleLister = c.extInformerFactory.Engine().V1alpha1().AzureRoles().Lister()
}

func (c *VaultController) runAzureRoleInjector(key string) error {
	obj, exist, err := c.azureRoleInformer.GetIndexer().GetByKey(key)
	if err != nil {
		glog.Errorf("Fetching object with key %s from store failed with %v", key, err)
		return err
	}

	if !exist {
		glog.Warningf("AzureRole %s does not exist anymore", key)

	} else {
		role := obj.(*api.AzureRole).DeepCopy()

		glog.Infof("Sync/Add/Update for AzureRole %s/%s", role.Namespace, role.Name)

		if role.DeletionTimestamp != nil {
			if core_util.HasFinalizer(role.ObjectMeta, AzureRoleFinalizer) {
				go c.runAzureRoleFinalizer(role, finalizerTimeout, finalizerInterval)
			}
		} else {
			if !core_util.HasFinalizer(role.ObjectMeta, AzureRoleFinalizer) {
				// Add finalizer
				_, _, err := patchutil.PatchAzureRole(context.TODO(), c.extClient.EngineV1alpha1(), role, func(role *api.AzureRole) *api.AzureRole {
					role.ObjectMeta = core_util.AddFinalizer(role.ObjectMeta, AzureRoleFinalizer)
					return role
				}, metav1.PatchOptions{})
				if err != nil {
					return errors.Wrapf(err, "failed to set AzureRole finalizer for %s/%s", role.Namespace, role.Name)
				}
			}

			azureRClient, err := azure.NewAzureRole(c.kubeClient, c.appCatalogClient, role)
			if err != nil {
				return err
			}

			err = c.reconcileAzureRole(azureRClient, role)
			if err != nil {
				return errors.Wrapf(err, "for AzureRole %s/%s:", role.Namespace, role.Name)
			}
		}
	}
	return nil
}

// Will do:
//	For vault:
// 	  - configure a Azure role
//    - sync role
func (c *VaultController) reconcileAzureRole(azureRClient azure.AzureRoleInterface, role *api.AzureRole) error {
	// create role
	err := azureRClient.CreateRole()
	if err != nil {
		_, err2 := patchutil.UpdateAzureRoleStatus(
			context.TODO(),
			c.extClient.EngineV1alpha1(),
			role.ObjectMeta,
			func(status *api.AzureRoleStatus) *api.AzureRoleStatus {
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

	_, err = patchutil.UpdateAzureRoleStatus(
		context.TODO(),
		c.extClient.EngineV1alpha1(),
		role.ObjectMeta,
		func(status *api.AzureRoleStatus) *api.AzureRoleStatus {
			status.Phase = AzureRolePhaseSuccess
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
	return err
}

func (c *VaultController) runAzureRoleFinalizer(role *api.AzureRole, timeout time.Duration, interval time.Duration) {
	if role == nil {
		glog.Infoln("AzureRole is nil")
		return
	}

	id := getAzureRoleId(role)
	if c.finalizerInfo.IsAlreadyProcessing(id) {
		// already processing
		return
	}

	glog.Infof("Processing finalizer for AzureRole %s/%s", role.Namespace, role.Name)
	// Add key to finalizerInfo, it will prevent other go routine to processing for this AzureRole
	c.finalizerInfo.Add(id)

	stopCh := time.After(timeout)
	finalizationDone := false
	timeOutOccured := false
	attempt := 0

	for {
		glog.Infof("AzureRole %s/%s finalizer: attempt %d\n", role.Namespace, role.Name, attempt)

		select {
		case <-stopCh:
			timeOutOccured = true
		default:
		}

		if timeOutOccured {
			break
		}

		if !finalizationDone {
			d, err := azure.NewAzureRole(c.kubeClient, c.appCatalogClient, role)
			if err != nil {
				glog.Errorf("AzureRole %s/%s finalizer: %v", role.Namespace, role.Name, err)
			} else {
				err = c.finalizeAzureRole(d, role)
				if err != nil {
					glog.Errorf("AzureRole %s/%s finalizer: %v", role.Namespace, role.Name, err)
				} else {
					finalizationDone = true
				}
			}
		}

		if finalizationDone {
			err := c.removeAzureRoleFinalizer(role)
			if err != nil {
				glog.Errorf("AzureRole %s/%s finalizer: removing finalizer %v", role.Namespace, role.Name, err)
			} else {
				break
			}
		}

		select {
		case <-stopCh:
			timeOutOccured = true
		case <-time.After(interval):
		}
		attempt++
	}

	err := c.removeAzureRoleFinalizer(role)
	if err != nil {
		glog.Errorf("AzureRole %s/%s finalizer: removing finalizer %v", role.Namespace, role.Name, err)
	} else {
		glog.Infof("Removed finalizer for AzureRole %s/%s", role.Namespace, role.Name)
	}

	// Delete key from finalizer info as processing is done
	c.finalizerInfo.Delete(id)
}

// Do:
//	- delete role in vault
func (c *VaultController) finalizeAzureRole(azureRClient azure.AzureRoleInterface, role *api.AzureRole) error {
	err := azureRClient.DeleteRole(role.RoleName())
	if err != nil {
		return errors.Wrap(err, "failed to delete azure role")
	}
	return nil
}

func (c *VaultController) removeAzureRoleFinalizer(role *api.AzureRole) error {
	m, err := c.extClient.EngineV1alpha1().AzureRoles(role.Namespace).Get(context.TODO(), role.Name, metav1.GetOptions{})
	if kerr.IsNotFound(err) {
		return nil
	} else if err != nil {
		return err
	}

	// remove finalizer
	_, _, err = patchutil.PatchAzureRole(context.TODO(), c.extClient.EngineV1alpha1(), m, func(role *api.AzureRole) *api.AzureRole {
		role.ObjectMeta = core_util.RemoveFinalizer(role.ObjectMeta, AzureRoleFinalizer)
		return role
	}, metav1.PatchOptions{})
	return err
}

func getAzureRoleId(role *api.AzureRole) string {
	return fmt.Sprintf("%s/%s/%s", api.ResourceAzureRole, role.Namespace, role.Name)
}
