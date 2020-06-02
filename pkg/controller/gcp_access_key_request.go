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
	"kubevault.dev/operator/pkg/vault/credential"

	"github.com/appscode/go/crypto/rand"
	"github.com/golang/glog"
	"github.com/pkg/errors"
	core "k8s.io/api/core/v1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	kmapi "kmodules.xyz/client-go/api/v1"
	core_util "kmodules.xyz/client-go/core/v1"
	"kmodules.xyz/client-go/tools/queue"
)

const (
	GCPAccessKeyRequestFinalizer = "gcpaccesskeyrequest.engine.kubevault.com"
)

func (c *VaultController) initGCPAccessKeyWatcher() {
	c.gcpAccessInformer = c.extInformerFactory.Engine().V1alpha1().GCPAccessKeyRequests().Informer()
	c.gcpAccessQueue = queue.New(api.ResourceKindGCPAccessKeyRequest, c.MaxNumRequeues, c.NumThreads, c.runGCPAccessKeyRequestInjector)
	c.gcpAccessInformer.AddEventHandler(queue.NewEventHandler(c.gcpAccessQueue.GetQueue(), func(oldObj, newObj interface{}) bool {
		old := oldObj.(*api.GCPAccessKeyRequest)
		nu := newObj.(*api.GCPAccessKeyRequest)

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
	c.gcpAccessLister = c.extInformerFactory.Engine().V1alpha1().GCPAccessKeyRequests().Lister()
}

func (c *VaultController) runGCPAccessKeyRequestInjector(key string) error {
	obj, exist, err := c.gcpAccessInformer.GetIndexer().GetByKey(key)
	if err != nil {
		glog.Errorf("Fetching object with key %s from store failed with %v", key, err)
		return err
	}

	if !exist {
		glog.Warningf("GCPAccessKeyRequest %s does not exist anymore", key)

	} else {
		gcpAccessReq := obj.(*api.GCPAccessKeyRequest).DeepCopy()

		glog.Infof("Sync/Add/Update for GCPAccessKeyRequest %s/%s", gcpAccessReq.Namespace, gcpAccessReq.Name)

		if gcpAccessReq.DeletionTimestamp != nil {
			if core_util.HasFinalizer(gcpAccessReq.ObjectMeta, GCPAccessKeyRequestFinalizer) {
				go c.runGCPAccessKeyRequestFinalizer(gcpAccessReq, finalizerTimeout, finalizerInterval)
			}
		} else {
			if !core_util.HasFinalizer(gcpAccessReq.ObjectMeta, GCPAccessKeyRequestFinalizer) {
				// Add finalizer
				_, _, err = patchutil.PatchGCPAccessKeyRequest(context.TODO(), c.extClient.EngineV1alpha1(), gcpAccessReq, func(binding *api.GCPAccessKeyRequest) *api.GCPAccessKeyRequest {
					binding.ObjectMeta = core_util.AddFinalizer(binding.ObjectMeta, GCPAccessKeyRequestFinalizer)
					return binding
				}, metav1.PatchOptions{})
				if err != nil {
					return errors.Wrapf(err, "failed to set GCPAccessKeyRequest finalizer for %s/%s", gcpAccessReq.Namespace, gcpAccessReq.Name)
				}
			}

			var condType string
			for _, c := range gcpAccessReq.Status.Conditions {
				if c.Type == kmapi.ConditionRequestApproved || c.Type == kmapi.ConditionRequestDenied {
					condType = c.Type
				}
			}

			if condType == kmapi.ConditionRequestApproved {
				gcpCredManager, err := credential.NewCredentialManagerForGCP(c.kubeClient, c.appCatalogClient, c.extClient, gcpAccessReq)
				if err != nil {
					return err
				}

				err = c.reconcileGCPAccessKeyRequest(gcpCredManager, gcpAccessReq)
				if err != nil {
					return errors.Wrapf(err, "For GCPAccessKeyRequest %s/%s", gcpAccessReq.Namespace, gcpAccessReq.Name)
				}
			} else if condType == kmapi.ConditionRequestDenied {
				glog.Infof("For GCPAccessKeyRequest %s/%s: request is denied", gcpAccessReq.Namespace, gcpAccessReq.Name)
			} else {
				glog.Infof("For GCPAccessKeyRequest %s/%s: request is not approved yet", gcpAccessReq.Namespace, gcpAccessReq.Name)
			}
		}
	}
	return nil
}

// Will do:
//	For vault:
//	  - get gcp credential
//	  - create secret containing credential
//	  - create rbac role and role binding
//    - sync role binding
func (c *VaultController) reconcileGCPAccessKeyRequest(gcpCM credential.CredentialManager, req *api.GCPAccessKeyRequest) error {
	var (
		name = req.Name
		ns   = req.Namespace
	)

	var secretName string
	if req.Status.Secret != nil {
		secretName = req.Status.Secret.Name
	}

	// check whether lease id exists in .status.lease or not
	// if does not exist in .status.lease, then get credential
	if req.Status.Lease == nil {
		// get gcp credential secret
		credSecret, err := gcpCM.GetCredential()
		if err != nil {
			_, err2 := patchutil.UpdateGCPAccessKeyRequestStatus(
				context.TODO(),
				c.extClient.EngineV1alpha1(),
				req.ObjectMeta,
				func(status *api.GCPAccessKeyRequestStatus) *api.GCPAccessKeyRequestStatus {
					status.Conditions = kmapi.SetCondition(status.Conditions, kmapi.Condition{
						Type:               kmapi.ConditionFailure,
						Reason:             "FailedToGetCredential",
						Message:            err.Error(),
						LastTransitionTime: metav1.Now(),
					})
					return status
				},
				metav1.UpdateOptions{},
			)
			return utilerrors.NewAggregate([]error{err2, err})
		}

		secretName = rand.WithUniqSuffix(name)
		err = gcpCM.CreateSecret(secretName, ns, credSecret)
		if err != nil {
			if len(credSecret.LeaseID) != 0 {
				err2 := gcpCM.RevokeLease(credSecret.LeaseID)
				if err2 != nil {
					return errors.Wrapf(err, "failed to revoke lease with %v", err2)
				}
			}

			_, err2 := patchutil.UpdateGCPAccessKeyRequestStatus(
				context.TODO(),
				c.extClient.EngineV1alpha1(),
				req.ObjectMeta,
				func(status *api.GCPAccessKeyRequestStatus) *api.GCPAccessKeyRequestStatus {
					status.Conditions = kmapi.SetCondition(status.Conditions, kmapi.Condition{
						Type:               kmapi.ConditionFailure,
						Reason:             "FailedToCreateSecret",
						Message:            err.Error(),
						LastTransitionTime: metav1.Now(),
					})
					return status
				},
				metav1.UpdateOptions{},
			)
			return utilerrors.NewAggregate([]error{err2, err})
		}

		_, err = patchutil.UpdateGCPAccessKeyRequestStatus(
			context.TODO(),
			c.extClient.EngineV1alpha1(),
			req.ObjectMeta,
			func(status *api.GCPAccessKeyRequestStatus) *api.GCPAccessKeyRequestStatus {
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

				return status
			},
			metav1.UpdateOptions{},
		)
		if err != nil {
			return err
		}
	}

	roleName := getSecretAccessRoleName(api.ResourceKindGCPAccessKeyRequest, ns, req.Name)

	err := gcpCM.CreateRole(roleName, ns, secretName)
	if err != nil {
		_, err2 := patchutil.UpdateGCPAccessKeyRequestStatus(
			context.TODO(),
			c.extClient.EngineV1alpha1(),
			req.ObjectMeta,
			func(status *api.GCPAccessKeyRequestStatus) *api.GCPAccessKeyRequestStatus {
				status.Conditions = kmapi.SetCondition(status.Conditions, kmapi.Condition{
					Type:               kmapi.ConditionFailure,
					Reason:             "FailedToCreateRole",
					Message:            err.Error(),
					LastTransitionTime: metav1.Now(),
				})
				return status
			},
			metav1.UpdateOptions{},
		)
		return utilerrors.NewAggregate([]error{err2, err})
	}

	err = gcpCM.CreateRoleBinding(roleName, ns, roleName, req.Spec.Subjects)
	if err != nil {
		_, err2 := patchutil.UpdateGCPAccessKeyRequestStatus(
			context.TODO(),
			c.extClient.EngineV1alpha1(),
			req.ObjectMeta,
			func(status *api.GCPAccessKeyRequestStatus) *api.GCPAccessKeyRequestStatus {
				status.Conditions = kmapi.SetCondition(status.Conditions, kmapi.Condition{
					Type:               kmapi.ConditionFailure,
					Reason:             "FailedToCreateRoleBinding",
					Message:            err.Error(),
					LastTransitionTime: metav1.Now(),
				})
				return status
			},
			metav1.UpdateOptions{},
		)
		return utilerrors.NewAggregate([]error{err2, err})
	}

	_, err = patchutil.UpdateGCPAccessKeyRequestStatus(
		context.TODO(),
		c.extClient.EngineV1alpha1(),
		req.ObjectMeta,
		func(status *api.GCPAccessKeyRequestStatus) *api.GCPAccessKeyRequestStatus {
			status.Conditions = kmapi.RemoveCondition(status.Conditions, kmapi.ConditionFailure)
			return status
		},
		metav1.UpdateOptions{},
	)
	return err
}

func (c *VaultController) runGCPAccessKeyRequestFinalizer(req *api.GCPAccessKeyRequest, timeout time.Duration, interval time.Duration) {
	if req == nil {
		glog.Infoln("GCPAccessKeyRequest is nil")
		return
	}

	id := getGCPAccessKeyRequestId(req)
	if c.finalizerInfo.IsAlreadyProcessing(id) {
		// already processing
		return
	}

	glog.Infof("Processing finalizer for GCPAccessKeyRequest %s/%s", req.Namespace, req.Name)
	// Add key to finalizerInfo, it will prevent other go routine to processing for this GCPAccessKeyRequest
	c.finalizerInfo.Add(id)

	stopCh := time.After(timeout)
	finalizationDone := false
	timeOutOccured := false
	attempt := 0

	for {
		glog.Infof("GCPAccessKeyRequest %s/%s finalizer: attempt %d\n", req.Namespace, req.Name, attempt)

		select {
		case <-stopCh:
			timeOutOccured = true
		default:
		}

		if timeOutOccured {
			break
		}

		if !finalizationDone {
			gcpCM, err := credential.NewCredentialManagerForGCP(c.kubeClient, c.appCatalogClient, c.extClient, req)
			if err != nil {
				glog.Errorf("GCPAccessKeyRequest %s/%s finalizer: %v", req.Namespace, req.Name, err)
			} else {
				err = c.finalizeGCPAccessKeyRequest(gcpCM, req.Status.Lease)
				if err != nil {
					glog.Errorf("GCPAccessKeyRequest %s/%s finalizer: %v", req.Namespace, req.Name, err)
				} else {
					finalizationDone = true
				}
			}
		}

		if finalizationDone {
			err := c.removeGCPAccessKeyRequestFinalizer(req)
			if err != nil {
				glog.Errorf("GCPAccessKeyRequest %s/%s finalizer: %v", req.Namespace, req.Name, err)
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

	err := c.removeGCPAccessKeyRequestFinalizer(req)
	if err != nil {
		glog.Errorf("GCPAccessKeyRequest %s/%s finalizer: %v", req.Namespace, req.Name, err)
	} else {
		glog.Infof("Removed finalizer for GCPAccessKeyRequest %s/%s", req.Namespace, req.Name)
	}

	// Delete key from finalizer info as processing is done
	c.finalizerInfo.Delete(id)
}

func (c *VaultController) finalizeGCPAccessKeyRequest(gcpCM credential.CredentialManager, lease *api.Lease) error {
	if lease == nil {
		return nil
	}
	if lease.ID == "" {
		return nil
	}
	return gcpCM.RevokeLease(lease.ID)
}

func (c *VaultController) removeGCPAccessKeyRequestFinalizer(gcpAKReq *api.GCPAccessKeyRequest) error {
	accessReq, err := c.extClient.EngineV1alpha1().GCPAccessKeyRequests(gcpAKReq.Namespace).Get(context.TODO(), gcpAKReq.Name, metav1.GetOptions{})
	if kerr.IsNotFound(err) {
		return nil
	} else if err != nil {
		return err
	}

	_, _, err = patchutil.PatchGCPAccessKeyRequest(context.TODO(), c.extClient.EngineV1alpha1(), accessReq, func(in *api.GCPAccessKeyRequest) *api.GCPAccessKeyRequest {
		in.ObjectMeta = core_util.RemoveFinalizer(in.ObjectMeta, GCPAccessKeyRequestFinalizer)
		return in
	}, metav1.PatchOptions{})
	return err
}

func getGCPAccessKeyRequestId(gcpAKReq *api.GCPAccessKeyRequest) string {
	return fmt.Sprintf("%s/%s/%s", api.ResourceGCPAccessKeyRequest, gcpAKReq.Namespace, gcpAKReq.Name)
}
