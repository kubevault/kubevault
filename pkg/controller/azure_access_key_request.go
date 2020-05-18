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
	"kubevault.dev/operator/pkg/vault/credential"

	"github.com/appscode/go/crypto/rand"
	"github.com/golang/glog"
	"github.com/pkg/errors"
	core "k8s.io/api/core/v1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kmapi "kmodules.xyz/client-go/api/v1"
	core_util "kmodules.xyz/client-go/core/v1"
	"kmodules.xyz/client-go/tools/queue"
)

const (
	AzureAccessKeyRequestFinalizer = "azureaccesskeyrequest.engine.kubevault.com"
)

func (c *VaultController) initAzureAccessKeyWatcher() {
	c.azureAccessInformer = c.extInformerFactory.Engine().V1alpha1().AzureAccessKeyRequests().Informer()
	c.azureAccessQueue = queue.New(api.ResourceKindAzureAccessKeyRequest, c.MaxNumRequeues, c.NumThreads, c.runAzureAccessKeyRequestInjector)
	c.azureAccessInformer.AddEventHandler(queue.NewEventHandler(c.azureAccessQueue.GetQueue(), func(oldObj, newObj interface{}) bool {
		old := oldObj.(*api.AzureAccessKeyRequest)
		nu := newObj.(*api.AzureAccessKeyRequest)

		oldCondType := ""
		nuCondType := ""
		for _, c := range old.Status.Conditions {
			if c.Type == kmapi.ConditionRequestApproved || c.Type == kmapi.ConditionRequestDenied {
				oldCondType = string(c.Type)
			}
		}
		for _, c := range nu.Status.Conditions {
			if c.Type == kmapi.ConditionRequestApproved || c.Type == kmapi.ConditionRequestDenied {
				nuCondType = string(c.Type)
			}
		}
		if oldCondType != nuCondType {
			return true
		}
		return nu.GetDeletionTimestamp() != nil
	}))
	c.azureAccessLister = c.extInformerFactory.Engine().V1alpha1().AzureAccessKeyRequests().Lister()
}

func (c *VaultController) runAzureAccessKeyRequestInjector(key string) error {
	obj, exist, err := c.azureAccessInformer.GetIndexer().GetByKey(key)
	if err != nil {
		glog.Errorf("Fetching object with key %s from store failed with %v", key, err)
		return err
	}

	if !exist {
		glog.Warningf("AzureAccessKeyRequest %s does not exist anymore", key)

	} else {
		azureAccessReq := obj.(*api.AzureAccessKeyRequest).DeepCopy()

		glog.Infof("Sync/Add/Update for AzureAccessKeyRequest %s/%s", azureAccessReq.Namespace, azureAccessReq.Name)

		if azureAccessReq.DeletionTimestamp != nil {
			if core_util.HasFinalizer(azureAccessReq.ObjectMeta, AzureAccessKeyRequestFinalizer) {
				go c.runAzureAccessKeyRequestFinalizer(azureAccessReq, finalizerTimeout, finalizerInterval)
			}
		} else {
			if !core_util.HasFinalizer(azureAccessReq.ObjectMeta, AzureAccessKeyRequestFinalizer) {
				// Add finalizer
				_, _, err = patchutil.PatchAzureAccessKeyRequest(c.extClient.EngineV1alpha1(), azureAccessReq, func(aAKR *api.AzureAccessKeyRequest) *api.AzureAccessKeyRequest {
					aAKR.ObjectMeta = core_util.AddFinalizer(aAKR.ObjectMeta, AzureAccessKeyRequestFinalizer)
					return aAKR
				})
				if err != nil {
					return errors.Wrapf(err, "failed to set AzureAccessKeyRequest finalizer for %s/%s", azureAccessReq.Namespace, azureAccessReq.Name)
				}
			}

			var condType string
			for _, c := range azureAccessReq.Status.Conditions {
				if c.Type == kmapi.ConditionRequestApproved || c.Type == kmapi.ConditionRequestDenied {
					condType = c.Type
				}
			}

			if condType == kmapi.ConditionRequestApproved {
				azureCredManager, err := credential.NewCredentialManagerForAzure(c.kubeClient, c.appCatalogClient, c.extClient, azureAccessReq)
				if err != nil {
					return err
				}

				err = c.reconcileAzureAccessKeyRequest(azureCredManager, azureAccessReq)
				if err != nil {
					return errors.Wrapf(err, "For AzureAccessKeyRequest %s/%s", azureAccessReq.Namespace, azureAccessReq.Name)
				}
			} else if condType == kmapi.ConditionRequestDenied {
				glog.Infof("For AzureAccessKeyRequest %s/%s: request is denied", azureAccessReq.Namespace, azureAccessReq.Name)
			} else {
				glog.Infof("For AzureAccessKeyRequest %s/%s: request is not approved yet", azureAccessReq.Namespace, azureAccessReq.Name)
			}
		}
	}
	return nil
}

// Will do:
//	For vault:
//	  - get azure credential
//	  - create secret containing credential
//	  - create rbac role and role binding
//    - sync role binding
func (c *VaultController) reconcileAzureAccessKeyRequest(azureCM credential.CredentialManager, azureAccessKeyReq *api.AzureAccessKeyRequest) error {
	var (
		name   = azureAccessKeyReq.Name
		ns     = azureAccessKeyReq.Namespace
		status = azureAccessKeyReq.Status
	)

	var secretName string
	if azureAccessKeyReq.Status.Secret != nil {
		secretName = azureAccessKeyReq.Status.Secret.Name
	}

	// check whether lease id exists in .status.lease or not
	// if does not exist in .status.lease, then get credential
	if azureAccessKeyReq.Status.Lease == nil {
		// get azure credential secret
		credSecret, err := azureCM.GetCredential()
		if err != nil {
			status.Conditions = kmapi.SetCondition(status.Conditions, kmapi.Condition{
				Type:               kmapi.ConditionFailure,
				Reason:             "FailedToGetCredential",
				Message:            err.Error(),
				LastTransitionTime: metav1.Now(),
			})

			err2 := c.updateAzureAccessKeyRequestStatus(&status, azureAccessKeyReq)
			if err2 != nil {
				return errors.Wrapf(err2, "failed to update status")
			}
			return errors.WithStack(err)
		}

		secretName = rand.WithUniqSuffix(name)
		err = azureCM.CreateSecret(secretName, ns, credSecret)
		if err != nil {
			if len(credSecret.LeaseID) != 0 {
				err2 := azureCM.RevokeLease(credSecret.LeaseID)
				if err2 != nil {
					return errors.Wrapf(err, "failed to revoke lease with %v", err2)
				}
			}
			status.Conditions = kmapi.SetCondition(status.Conditions, kmapi.Condition{
				Type:               kmapi.ConditionFailure,
				Reason:             "FailedToCreateSecret",
				Message:            err.Error(),
				LastTransitionTime: metav1.Now(),
			})

			err2 := c.updateAzureAccessKeyRequestStatus(&status, azureAccessKeyReq)
			if err2 != nil {
				return errors.Wrapf(err, "failed to update status with %v", err2)
			}

			return errors.WithStack(err)
		}

		// add lease info in status
		status.Lease = &api.Lease{
			ID: credSecret.LeaseID,
			Duration: metav1.Duration{
				Duration: time.Second * time.Duration(credSecret.LeaseDuration),
			},
			Renewable: credSecret.Renewable,
		}

		// assign secret name
		status.Secret = &core.LocalObjectReference{
			Name: secretName,
		}
	}

	roleName := getSecretAccessRoleName(api.ResourceKindAzureAccessKeyRequest, ns, azureAccessKeyReq.Name)

	err := azureCM.CreateRole(roleName, ns, secretName)
	if err != nil {
		status.Conditions = kmapi.SetCondition(status.Conditions, kmapi.Condition{
			Type:               kmapi.ConditionFailure,
			Reason:             "FailedToCreateRole",
			Message:            err.Error(),
			LastTransitionTime: metav1.Now(),
		})

		err2 := c.updateAzureAccessKeyRequestStatus(&status, azureAccessKeyReq)
		if err2 != nil {
			return errors.Wrapf(err2, "failed to update status")
		}
		return errors.WithStack(err)
	}

	err = azureCM.CreateRoleBinding(roleName, ns, roleName, azureAccessKeyReq.Spec.Subjects)
	if err != nil {
		status.Conditions = kmapi.SetCondition(status.Conditions, kmapi.Condition{
			Type:               kmapi.ConditionFailure,
			Reason:             "FailedToCreateRoleBinding",
			Message:            err.Error(),
			LastTransitionTime: metav1.Now(),
		})

		err2 := c.updateAzureAccessKeyRequestStatus(&status, azureAccessKeyReq)
		if err2 != nil {
			return errors.Wrapf(err2, "failed to update status")
		}
		return errors.WithStack(err)
	}

	status.Conditions = kmapi.RemoveCondition(status.Conditions, kmapi.ConditionFailure)
	err = c.updateAzureAccessKeyRequestStatus(&status, azureAccessKeyReq)
	if err != nil {
		return errors.Wrap(err, "failed to update status")
	}
	return nil
}

func (c *VaultController) updateAzureAccessKeyRequestStatus(status *api.AzureAccessKeyRequestStatus, azureAKReq *api.AzureAccessKeyRequest) error {
	_, err := patchutil.UpdateAzureAccessKeyRequestStatus(c.extClient.EngineV1alpha1(), azureAKReq.ObjectMeta, func(s *api.AzureAccessKeyRequestStatus) *api.AzureAccessKeyRequestStatus {
		return status
	})
	return err
}

func (c *VaultController) runAzureAccessKeyRequestFinalizer(azureAKReq *api.AzureAccessKeyRequest, timeout time.Duration, interval time.Duration) {
	if azureAKReq == nil {
		glog.Infoln("AzureAccessKeyRequest is nil")
		return
	}

	id := getAzureAccessKeyRequestId(azureAKReq)
	if c.finalizerInfo.IsAlreadyProcessing(id) {
		// already processing
		return
	}

	glog.Infof("Processing finalizer for AzureAccessKeyRequest %s/%s", azureAKReq.Namespace, azureAKReq.Name)
	// Add key to finalizerInfo, it will prevent other go routine to processing for this AzureAccessKeyRequest
	c.finalizerInfo.Add(id)

	stopCh := time.After(timeout)
	finalizationDone := false
	timeOutOccured := false
	attempt := 0

	for {
		glog.Infof("AzureAccessKeyRequest %s/%s finalizer: attempt %d\n", azureAKReq.Namespace, azureAKReq.Name, attempt)

		select {
		case <-stopCh:
			timeOutOccured = true
		default:
		}

		if timeOutOccured {
			break
		}

		if !finalizationDone {
			azureCM, err := credential.NewCredentialManagerForAzure(c.kubeClient, c.appCatalogClient, c.extClient, azureAKReq)
			if err != nil {
				glog.Errorf("AzureAccessKeyRequest %s/%s finalizer: %v", azureAKReq.Namespace, azureAKReq.Name, err)
			} else {
				err = c.finalizeAzureAccessKeyRequest(azureCM, azureAKReq.Status.Lease)
				if err != nil {
					glog.Errorf("AzureAccessKeyRequest %s/%s finalizer: %v", azureAKReq.Namespace, azureAKReq.Name, err)
				} else {
					finalizationDone = true
				}
			}
		}

		if finalizationDone {
			err := c.removeAzureAccessKeyRequestFinalizer(azureAKReq)
			if err != nil {
				glog.Errorf("AzureAccessKeyRequest %s/%s finalizer: %v", azureAKReq.Namespace, azureAKReq.Name, err)
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

	err := c.removeAzureAccessKeyRequestFinalizer(azureAKReq)
	if err != nil {
		glog.Errorf("AzureAccessKeyRequest %s/%s finalizer: %v", azureAKReq.Namespace, azureAKReq.Name, err)
	} else {
		glog.Infof("Removed finalizer for AzureAccessKeyRequest %s/%s", azureAKReq.Namespace, azureAKReq.Name)
	}

	// Delete key from finalizer info as processing is done
	c.finalizerInfo.Delete(id)
}

func (c *VaultController) finalizeAzureAccessKeyRequest(azureCM credential.CredentialManager, lease *api.Lease) error {
	if lease == nil {
		return nil
	}
	if lease.ID == "" {
		return nil
	}
	return azureCM.RevokeLease(lease.ID)
}

func (c *VaultController) removeAzureAccessKeyRequestFinalizer(azureAKReq *api.AzureAccessKeyRequest) error {
	accessReq, err := c.extClient.EngineV1alpha1().AzureAccessKeyRequests(azureAKReq.Namespace).Get(azureAKReq.Name, metav1.GetOptions{})
	if kerr.IsNotFound(err) {
		return nil
	} else if err != nil {
		return err
	}

	_, _, err = patchutil.PatchAzureAccessKeyRequest(c.extClient.EngineV1alpha1(), accessReq, func(in *api.AzureAccessKeyRequest) *api.AzureAccessKeyRequest {
		in.ObjectMeta = core_util.RemoveFinalizer(in.ObjectMeta, AzureAccessKeyRequestFinalizer)
		return in
	})
	return err
}

func getAzureAccessKeyRequestId(azureAKReq *api.AzureAccessKeyRequest) string {
	return fmt.Sprintf("%s/%s/%s", api.ResourceAzureAccessKeyRequest, azureAKReq.Namespace, azureAKReq.Name)
}
