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
	"fmt"
	"time"

	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	patchutil "kubevault.dev/operator/client/clientset/versioned/typed/approle/v1alpha1/util"

	"github.com/golang/glog"
	"github.com/pkg/errors"
	kmapi "kmodules.xyz/client-go/api/v1"
	core_util "kmodules.xyz/client-go/core/v1"
	"kmodules.xyz/client-go/tools/queue"
	approleapi "kubevault.dev/operator/apis/approle/v1alpha1"
	"kubevault.dev/operator/pkg/vault/approle"
)

const (
	VaultAppRoleFinalizer = "approle.kubevault.com"
)

func (c *VaultController) initVaultAppRoleWatcher() {
	c.vAppRoleInformer = c.extInformerFactory.Approle().V1alpha1().VaultAppRoles().Informer()
	c.vAppRoleQueue = queue.New(approleapi.ResourceKindVaultAppRole, c.MaxNumRequeues, c.NumThreads, c.runVaultAppRoleInjector)
	c.vAppRoleInformer.AddEventHandler(queue.NewReconcilableHandler(c.vAppRoleQueue.GetQueue()))
	c.vAppRoleLister = c.extInformerFactory.Approle().V1alpha1().VaultAppRoles().Lister()
}

// runVaultAppRoleInjector gets the vault approle object indexed by the key from cache
// and initializes, reconciles or garbage collects the vault approle as needed.
func (c *VaultController) runVaultAppRoleInjector(key string) error {
	obj, exists, err := c.vAppRoleInformer.GetIndexer().GetByKey(key)
	if err != nil {
		glog.Errorf("Fetching object with key %s from store failed with %v", key, err)
		return err
	}

	if !exists {
		glog.Warningf("VaultAppRole %s does not exist anymore\n", key)
	} else {
		vAppRole := obj.(*approleapi.VaultAppRole).DeepCopy()
		glog.Infof("Sync/Add/Update for VaultAppRole %s/%s\n", vAppRole.Namespace, vAppRole.Name)

		if vAppRole.DeletionTimestamp != nil {
			if core_util.HasFinalizer(vAppRole.ObjectMeta, VaultAppRoleFinalizer) {
				// Finalize VaultAppRole
				go c.runAppRoleFinalizer(vAppRole, timeoutForFinalizer, timeIntervalForFinalizer)
			} else {
				glog.Infof("Finalizer not found for VaultAppRole %s/%s", vAppRole.Namespace, vAppRole.Name)
			}
		} else {
			if !core_util.HasFinalizer(vAppRole.ObjectMeta, VaultAppRoleFinalizer) {
				// Add finalizer
				_, _, err := patchutil.PatchVaultAppRole(c.extClient.ApproleV1alpha1(), vAppRole, func(vp *approleapi.VaultAppRole) *approleapi.VaultAppRole {
					vp.ObjectMeta = core_util.AddFinalizer(vAppRole.ObjectMeta, VaultAppRoleFinalizer)
					return vp
				})
				if err != nil {
					return errors.Wrapf(err, "failed to set VaultAppRole finalizer for %s/%s", vAppRole.Namespace, vAppRole.Name)
				}
			}

			aClient, err := approle.NewAppRoleClientForVault(c.kubeClient, c.appCatalogClient, vAppRole)
			if err != nil {
				return errors.Wrapf(err, "for VaultAppRole %s/%s", vAppRole.Namespace, vAppRole.Name)
			}

			err = c.reconcileAppRole(vAppRole, aClient)
			if err != nil {
				return errors.Wrapf(err, "for VaultAppRole %s/%s", vAppRole.Namespace, vAppRole.Name)
			}
		}
	}
	return nil
}

// reconcileAppRole reconciles the vault's approle
// it will create or update approle in vault
func (c *VaultController) reconcileAppRole(vAppRole *approleapi.VaultAppRole, aClient approle.AppRole) error {
	status := vAppRole.Status

	// generate paypload map
	payload, err := vAppRole.GeneratePayLoad()

	if err != nil {
		return errors.Wrap(err, "failed to generate payload for VaultAppRole")
	}

	err2 := aClient.EnsureAppRole(vAppRole.AppRoleName(), payload)
	if err2 != nil {
		fmt.Println("Ok, error!")
		status.Phase = approleapi.AppRoleFailed
		status.Conditions = []kmapi.Condition{
			{
				Type:    kmapi.ConditionFailure,
				Status:  kmapi.ConditionTrue,
				Reason:  "FailedToPutAppRole",
				Message: err2.Error(),
			},
		}

		err3 := c.updateAppRoleStatus(&status, vAppRole)
		if err3 != nil {
			return errors.Wrap(err3, "failed to update VaultAppRole status")
		}
		return err2
	}

	// update status
	status.ObservedGeneration = vAppRole.Generation
	status.Conditions = []kmapi.Condition{}
	status.Phase = approleapi.AppRoleSuccess
	err4 := c.updateAppRoleStatus(&status, vAppRole)
	if err4 != nil {
		return errors.Wrap(err4, "failed to update VaultAppRole status")
	}
	return nil
}

// updateAppRoleStatus updates approle status
func (c *VaultController) updateAppRoleStatus(status *approleapi.VaultAppRoleStatus, vAppRole *approleapi.VaultAppRole) error {
	_, err := patchutil.UpdateVaultAppRoleStatus(c.extClient.ApproleV1alpha1(), vAppRole.ObjectMeta, func(s *approleapi.VaultAppRoleStatus) *approleapi.VaultAppRoleStatus {
		return status
	})
	return err
}

// runAppRoleFinalizer wil periodically run the finalizeAppRole until finalizeAppRole func produces no error or timeout occurs.
// After that it will remove the finalizer string from the objectMeta of VaultAppRole
func (c *VaultController) runAppRoleFinalizer(vAppRole *approleapi.VaultAppRole, timeout time.Duration, interval time.Duration) {
	if vAppRole == nil {
		glog.Infoln("VaultAppRole in nil")
		return
	}

	key := vAppRole.GetKey()
	if c.finalizerInfo.IsAlreadyProcessing(key) {
		// already processing it
		return
	}

	glog.Infof("Processing finalizer for VaultAppRole %s/%s", vAppRole.Namespace, vAppRole.Name)
	// Add key to finalizerInfo, it will prevent other go routine to processing for this VaultAppRole
	c.finalizerInfo.Add(key)
	stopCh := time.After(timeout)
	timeOutOccured := false
	for {
		select {
		case <-stopCh:
			timeOutOccured = true
		default:
		}

		if timeOutOccured {
			break
		}

		// finalize approle
		if err := c.finalizeAppRole(vAppRole); err == nil {
			glog.Infof("For VaultAppRole %s/%s: successfully removed approle from vault", vAppRole.Namespace, vAppRole.Name)
			break
		} else {
			glog.Infof("For VaultAppRole %s/%s: %v", vAppRole.Namespace, vAppRole.Name, err)
		}

		select {
		case <-stopCh:
			timeOutOccured = true
		case <-time.After(interval):
		}
	}

	// Remove finalizer
	_, err := patchutil.TryPatchVaultAppRole(c.extClient.ApproleV1alpha1(), vAppRole, func(in *approleapi.VaultAppRole) *approleapi.VaultAppRole {
		in.ObjectMeta = core_util.RemoveFinalizer(in.ObjectMeta, VaultAppRoleFinalizer)
		return in
	})
	if err != nil {
		glog.Errorf("For VaultAppRole %s/%s: %v", vAppRole.Namespace, vAppRole.Name, err)
	} else {
		glog.Infof("For VaultAppRole %s/%s: removed finalizer '%s'", vAppRole.Namespace, vAppRole.Name, VaultAppRoleFinalizer)
	}
	// Delete key from finalizer info as processing is done
	c.finalizerInfo.Delete(key)
	glog.Infof("Removed finalizer for VaultAppRole %s/%s", vAppRole.Namespace, vAppRole.Name)
}

// finalizeAppRole will delete the approle in vault
func (c *VaultController) finalizeAppRole(vAppRole *approleapi.VaultAppRole) error {
	out, err := c.extClient.ApproleV1alpha1().VaultAppRoles(vAppRole.Namespace).Get(vAppRole.Name, metav1.GetOptions{})
	if kerr.IsNotFound(err) {
		return nil
	} else if err != nil {
		return err
	}

	aClient, err := approle.NewAppRoleClientForVault(c.kubeClient, c.appCatalogClient, out)
	if err != nil {
		return err
	}
	return aClient.DeleteAppRole(vAppRole.AppRoleName())
}
