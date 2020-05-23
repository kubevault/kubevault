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
	"kubevault.dev/operator/pkg/vault/role/aws"

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
	AWSRolePhaseSuccess api.AWSRolePhase = "Success"
	AWSRoleFinalizer    string           = "awsrole.engine.kubevault.com"
)

func (c *VaultController) initAWSRoleWatcher() {
	c.awsRoleInformer = c.extInformerFactory.Engine().V1alpha1().AWSRoles().Informer()
	c.awsRoleQueue = queue.New(api.ResourceKindAWSRole, c.MaxNumRequeues, c.NumThreads, c.runAWSRoleInjector)
	c.awsRoleInformer.AddEventHandler(queue.NewReconcilableHandler(c.awsRoleQueue.GetQueue()))
	c.awsRoleLister = c.extInformerFactory.Engine().V1alpha1().AWSRoles().Lister()
}

func (c *VaultController) runAWSRoleInjector(key string) error {
	obj, exist, err := c.awsRoleInformer.GetIndexer().GetByKey(key)
	if err != nil {
		glog.Errorf("Fetching object with key %s from store failed with %v", key, err)
		return err
	}

	if !exist {
		glog.Warningf("AWSRole %s does not exist anymore", key)

	} else {
		role := obj.(*api.AWSRole).DeepCopy()

		glog.Infof("Sync/Add/Update for AWSRole %s/%s", role.Namespace, role.Name)

		if role.DeletionTimestamp != nil {
			if core_util.HasFinalizer(role.ObjectMeta, AWSRoleFinalizer) {
				go c.runAWSRoleFinalizer(role, finalizerTimeout, finalizerInterval)
			}
		} else {
			if !core_util.HasFinalizer(role.ObjectMeta, AWSRoleFinalizer) {
				// Add finalizer
				_, _, err := patchutil.PatchAWSRole(context.TODO(), c.extClient.EngineV1alpha1(), role, func(role *api.AWSRole) *api.AWSRole {
					role.ObjectMeta = core_util.AddFinalizer(role.ObjectMeta, AWSRoleFinalizer)
					return role
				}, metav1.PatchOptions{})
				if err != nil {
					return errors.Wrapf(err, "failed to set AWSRole finalizer for %s/%s", role.Namespace, role.Name)
				}
			}

			awsRClient, err := aws.NewAWSRole(c.kubeClient, c.appCatalogClient, role)
			if err != nil {
				return err
			}

			err = c.reconcileAWSRole(awsRClient, role)
			if err != nil {
				return errors.Wrapf(err, "for AWSRole %s/%s:", role.Namespace, role.Name)
			}
		}
	}
	return nil
}

// Will do:
//	For vault:
// 	  - configure a AWS role
//    - sync role
func (c *VaultController) reconcileAWSRole(awsRClient aws.AWSRoleInterface, role *api.AWSRole) error {
	// create role
	err := awsRClient.CreateRole()
	if err != nil {
		_, err2 := patchutil.UpdateAWSRoleStatus(
			context.TODO(),
			c.extClient.EngineV1alpha1(),
			role.ObjectMeta, func(status *api.AWSRoleStatus) *api.AWSRoleStatus {
				status.Conditions = []kmapi.Condition{
					{
						Type:    kmapi.ConditionFailure,
						Status:  kmapi.ConditionTrue,
						Reason:  "FailedToCreateRole",
						Message: err.Error(),
					},
				}
				return status
			},
			metav1.UpdateOptions{},
		)
		return utilerrors.NewAggregate([]error{err2, errors.Wrap(err, "failed to create role")})
	}

	_, err = patchutil.UpdateAWSRoleStatus(
		context.TODO(),
		c.extClient.EngineV1alpha1(),
		role.ObjectMeta, func(status *api.AWSRoleStatus) *api.AWSRoleStatus {
			status.Conditions = []kmapi.Condition{}
			status.Phase = AWSRolePhaseSuccess
			status.ObservedGeneration = role.Generation
			return status
		},
		metav1.UpdateOptions{},
	)
	return err
}

func (c *VaultController) runAWSRoleFinalizer(role *api.AWSRole, timeout time.Duration, interval time.Duration) {
	if role == nil {
		glog.Infoln("AWSRole is nil")
		return
	}

	id := getAWSRoleId(role)
	if c.finalizerInfo.IsAlreadyProcessing(id) {
		// already processing
		return
	}

	glog.Infof("Processing finalizer for AWSRole %s/%s", role.Namespace, role.Name)
	// Add key to finalizerInfo, it will prevent other go routine to processing for this AWSRole
	c.finalizerInfo.Add(id)

	stopCh := time.After(timeout)
	finalizationDone := false
	timeOutOccured := false
	attempt := 0

	for {
		glog.Infof("AWSRole %s/%s finalizer: attempt %d\n", role.Namespace, role.Name, attempt)

		select {
		case <-stopCh:
			timeOutOccured = true
		default:
		}

		if timeOutOccured {
			break
		}

		if !finalizationDone {
			d, err := aws.NewAWSRole(c.kubeClient, c.appCatalogClient, role)
			if err != nil {
				glog.Errorf("AWSRole %s/%s finalizer: %v", role.Namespace, role.Name, err)
			} else {
				err = c.finalizeAWSRole(d, role)
				if err != nil {
					glog.Errorf("AWSRole %s/%s finalizer: %v", role.Namespace, role.Name, err)
				} else {
					finalizationDone = true
				}
			}
		}

		if finalizationDone {
			err := c.removeAWSRoleFinalizer(role)
			if err != nil {
				glog.Errorf("AWSRole %s/%s finalizer: removing finalizer %v", role.Namespace, role.Name, err)
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

	err := c.removeAWSRoleFinalizer(role)
	if err != nil {
		glog.Errorf("AWSRole %s/%s finalizer: removing finalizer %v", role.Namespace, role.Name, err)
	} else {
		glog.Infof("Removed finalizer for AWSRole %s/%s", role.Namespace, role.Name)
	}

	// Delete key from finalizer info as processing is done
	c.finalizerInfo.Delete(id)
}

// Do:
//	- delete role in vault
func (c *VaultController) finalizeAWSRole(awsRClient aws.AWSRoleInterface, role *api.AWSRole) error {
	err := awsRClient.DeleteRole(role.RoleName())
	if err != nil {
		return errors.Wrap(err, "failed to delete aws role")
	}
	return nil
}

func (c *VaultController) removeAWSRoleFinalizer(role *api.AWSRole) error {
	m, err := c.extClient.EngineV1alpha1().AWSRoles(role.Namespace).Get(context.TODO(), role.Name, metav1.GetOptions{})
	if kerr.IsNotFound(err) {
		return nil
	} else if err != nil {
		return err
	}

	// remove finalizer
	_, _, err = patchutil.PatchAWSRole(context.TODO(), c.extClient.EngineV1alpha1(), m, func(role *api.AWSRole) *api.AWSRole {
		role.ObjectMeta = core_util.RemoveFinalizer(role.ObjectMeta, AWSRoleFinalizer)
		return role
	}, metav1.PatchOptions{})
	return err
}

func getAWSRoleId(role *api.AWSRole) string {
	return fmt.Sprintf("%s/%s/%s", api.ResourceAWSRole, role.Namespace, role.Name)
}
