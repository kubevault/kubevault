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

	api "kubevault.dev/operator/apis/engine/v1alpha1"
	patchutil "kubevault.dev/operator/client/clientset/versioned/typed/engine/v1alpha1/util"
	"kubevault.dev/operator/pkg/vault/role/aws"

	"github.com/golang/glog"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	core_util "kmodules.xyz/client-go/core/v1"
	"kmodules.xyz/client-go/tools/queue"
)

const (
	AWSRolePhaseSuccess    api.AWSRolePhase = "Success"
	AWSRoleConditionFailed string           = "Failed"
	AWSRoleFinalizer       string           = "awsrole.engine.kubevault.com"
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
		awsRole := obj.(*api.AWSRole).DeepCopy()

		glog.Infof("Sync/Add/Update for AWSRole %s/%s", awsRole.Namespace, awsRole.Name)

		if awsRole.DeletionTimestamp != nil {
			if core_util.HasFinalizer(awsRole.ObjectMeta, AWSRoleFinalizer) {
				go c.runAWSRoleFinalizer(awsRole, finalizerTimeout, finalizerInterval)
			}
		} else {
			if !core_util.HasFinalizer(awsRole.ObjectMeta, AWSRoleFinalizer) {
				// Add finalizer
				_, _, err := patchutil.PatchAWSRole(c.extClient.EngineV1alpha1(), awsRole, func(role *api.AWSRole) *api.AWSRole {
					role.ObjectMeta = core_util.AddFinalizer(role.ObjectMeta, AWSRoleFinalizer)
					return role
				})
				if err != nil {
					return errors.Wrapf(err, "failed to set AWSRole finalizer for %s/%s", awsRole.Namespace, awsRole.Name)
				}
			}

			awsRClient, err := aws.NewAWSRole(c.kubeClient, c.appCatalogClient, awsRole)
			if err != nil {
				return err
			}

			err = c.reconcileAWSRole(awsRClient, awsRole)
			if err != nil {
				return errors.Wrapf(err, "for AWSRole %s/%s:", awsRole.Namespace, awsRole.Name)
			}
		}
	}
	return nil
}

// Will do:
//	For vault:
// 	  - configure a AWS role
//    - sync role
func (c *VaultController) reconcileAWSRole(awsRClient aws.AWSRoleInterface, awsRole *api.AWSRole) error {
	status := awsRole.Status

	// create role
	err := awsRClient.CreateRole()
	if err != nil {
		status.Conditions = []api.AWSRoleCondition{
			{
				Type:    AWSRoleConditionFailed,
				Status:  corev1.ConditionTrue,
				Reason:  "FailedToCreateRole",
				Message: err.Error(),
			},
		}

		err2 := c.updatedAWSRoleStatus(&status, awsRole)
		if err2 != nil {
			return errors.Wrap(err2, "failed to update status")
		}
		return errors.Wrap(err, "failed to create role")
	}

	status.Conditions = []api.AWSRoleCondition{}
	status.Phase = AWSRolePhaseSuccess
	status.ObservedGeneration = awsRole.Generation

	err = c.updatedAWSRoleStatus(&status, awsRole)
	if err != nil {
		return errors.Wrapf(err, "failed to update AWSRole status")
	}
	return nil
}

func (c *VaultController) updatedAWSRoleStatus(status *api.AWSRoleStatus, awsRole *api.AWSRole) error {
	_, err := patchutil.UpdateAWSRoleStatus(c.extClient.EngineV1alpha1(), awsRole, func(s *api.AWSRoleStatus) *api.AWSRoleStatus {
		return status
	})
	return err
}

func (c *VaultController) runAWSRoleFinalizer(awsRole *api.AWSRole, timeout time.Duration, interval time.Duration) {
	if awsRole == nil {
		glog.Infoln("AWSRole is nil")
		return
	}

	id := getAWSRoleId(awsRole)
	if c.finalizerInfo.IsAlreadyProcessing(id) {
		// already processing
		return
	}

	glog.Infof("Processing finalizer for AWSRole %s/%s", awsRole.Namespace, awsRole.Name)
	// Add key to finalizerInfo, it will prevent other go routine to processing for this AWSRole
	c.finalizerInfo.Add(id)

	stopCh := time.After(timeout)
	finalizationDone := false
	timeOutOccured := false
	attempt := 0

	for {
		glog.Infof("AWSRole %s/%s finalizer: attempt %d\n", awsRole.Namespace, awsRole.Name, attempt)

		select {
		case <-stopCh:
			timeOutOccured = true
		default:
		}

		if timeOutOccured {
			break
		}

		if !finalizationDone {
			d, err := aws.NewAWSRole(c.kubeClient, c.appCatalogClient, awsRole)
			if err != nil {
				glog.Errorf("AWSRole %s/%s finalizer: %v", awsRole.Namespace, awsRole.Name, err)
			} else {
				err = c.finalizeAWSRole(d, awsRole)
				if err != nil {
					glog.Errorf("AWSRole %s/%s finalizer: %v", awsRole.Namespace, awsRole.Name, err)
				} else {
					finalizationDone = true
				}
			}
		}

		if finalizationDone {
			err := c.removeAWSRoleFinalizer(awsRole)
			if err != nil {
				glog.Errorf("AWSRole %s/%s finalizer: removing finalizer %v", awsRole.Namespace, awsRole.Name, err)
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

	err := c.removeAWSRoleFinalizer(awsRole)
	if err != nil {
		glog.Errorf("AWSRole %s/%s finalizer: removing finalizer %v", awsRole.Namespace, awsRole.Name, err)
	} else {
		glog.Infof("Removed finalizer for AWSRole %s/%s", awsRole.Namespace, awsRole.Name)
	}

	// Delete key from finalizer info as processing is done
	c.finalizerInfo.Delete(id)
}

// Do:
//	- delete role in vault
func (c *VaultController) finalizeAWSRole(awsRClient aws.AWSRoleInterface, awsRole *api.AWSRole) error {
	err := awsRClient.DeleteRole(awsRole.RoleName())
	if err != nil {
		return errors.Wrap(err, "failed to delete aws role")
	}
	return nil
}

func (c *VaultController) removeAWSRoleFinalizer(awsRole *api.AWSRole) error {
	m, err := c.extClient.EngineV1alpha1().AWSRoles(awsRole.Namespace).Get(awsRole.Name, metav1.GetOptions{})
	if kerr.IsNotFound(err) {
		return nil
	} else if err != nil {
		return err
	}

	// remove finalizer
	_, _, err = patchutil.PatchAWSRole(c.extClient.EngineV1alpha1(), m, func(role *api.AWSRole) *api.AWSRole {
		role.ObjectMeta = core_util.RemoveFinalizer(role.ObjectMeta, AWSRoleFinalizer)
		return role
	})
	return err
}

func getAWSRoleId(awsRole *api.AWSRole) string {
	return fmt.Sprintf("%s/%s/%s", api.ResourceAWSRole, awsRole.Namespace, awsRole.Name)
}
