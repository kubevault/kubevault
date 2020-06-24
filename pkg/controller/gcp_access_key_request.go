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
			if core_util.HasFinalizer(gcpAccessReq.ObjectMeta, KubeVaultFinalizer) {
				return c.runGCPAccessKeyRequestFinalizer(gcpAccessReq)
			}
		} else {
			if !core_util.HasFinalizer(gcpAccessReq.ObjectMeta, KubeVaultFinalizer) {
				// Add finalizer
				_, _, err = patchutil.PatchGCPAccessKeyRequest(context.TODO(), c.extClient.EngineV1alpha1(), gcpAccessReq, func(binding *api.GCPAccessKeyRequest) *api.GCPAccessKeyRequest {
					binding.ObjectMeta = core_util.AddFinalizer(binding.ObjectMeta, KubeVaultFinalizer)
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

			// If condition type is not set yet, set the phase to "WaitingForApproval".
			if condType == "" {
				newAKR, err := patchutil.UpdateGCPAccessKeyRequestStatus(
					context.TODO(),
					c.extClient.EngineV1alpha1(),
					gcpAccessReq.ObjectMeta,
					func(status *api.GCPAccessKeyRequestStatus) *api.GCPAccessKeyRequestStatus {
						status.Phase = api.RequestStatusPhaseWaitingForApproval
						return status
					},
					metav1.UpdateOptions{},
				)
				if err != nil {
					return errors.Wrap(err, fmt.Sprintf("failed to update the status of gcpAccessKeyRequest: %s/%s", gcpAccessReq.Namespace, gcpAccessReq.Name))
				}
				gcpAccessReq = newAKR
			}

			if condType == kmapi.ConditionRequestApproved {

				// If accessKeyRequest is successfully processed,
				// skip processing.
				if gcpAccessKeyRequestSuccessfullyProcessed(gcpAccessReq) {
					return nil
				}

				// Create credential manager which handle communication to vault server
				gcpCredManager, err := credential.NewCredentialManagerForGCP(c.kubeClient, c.appCatalogClient, c.extClient, gcpAccessReq)
				if err != nil {
					_, err2 := patchutil.UpdateGCPAccessKeyRequestStatus(
						context.TODO(),
						c.extClient.EngineV1alpha1(),
						gcpAccessReq.ObjectMeta,
						func(status *api.GCPAccessKeyRequestStatus) *api.GCPAccessKeyRequestStatus {
							status.Conditions = kmapi.SetCondition(status.Conditions, kmapi.Condition{
								Type:               kmapi.ConditionFailure,
								Status:             kmapi.ConditionTrue,
								Reason:             "FailedToCreateCredentialManager",
								Message:            err.Error(),
								LastTransitionTime: metav1.Now(),
							})
							return status
						},
						metav1.UpdateOptions{},
					)

					return utilerrors.NewAggregate([]error{err2, err})
				}

				err = c.reconcileGCPAccessKeyRequest(gcpCredManager, gcpAccessReq)
				// If reconcileGCPAccessKeyRequest fails,
				//	- Revoke lease if any
				// 	- Delete k8s secret if any
				//	- Update lease & secret references with nil value
				if err != nil {
					err1 := revokeLease(gcpCredManager, gcpAccessReq.Status.Lease)
					err2 := c.deleteCredSecretForGCPAccessKeyRequest(gcpAccessReq)
					// If it fails to revoke lease or delete secret,
					// no need to update status.
					if err1 != nil || err2 != nil {
						return utilerrors.NewAggregate([]error{err2, err1})
					}
					// successfully revoked key and deleted the k8s secret,
					// update the status.
					_, err3 := patchutil.UpdateGCPAccessKeyRequestStatus(
						context.TODO(),
						c.extClient.EngineV1alpha1(),
						gcpAccessReq.ObjectMeta,
						func(status *api.GCPAccessKeyRequestStatus) *api.GCPAccessKeyRequestStatus {
							status.Secret = nil
							status.Lease = nil
							return status
						},
						metav1.UpdateOptions{},
					)
					return errors.Wrapf(utilerrors.NewAggregate([]error{err3, err}), "For GCPAccessKeyRequest %s/%s", gcpAccessReq.Namespace, gcpAccessReq.Name)
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

	// if lease or secret ref was set during the previous cycle which was failed.
	// return error.
	if req.Status.Lease != nil || req.Status.Secret != nil {
		return errors.New("lease or secret ref is not empty")
	}

	// Get new credentials ( gcp service account key )
	credSecret, err := gcpCM.GetCredential()
	if err != nil {
		_, err2 := patchutil.UpdateGCPAccessKeyRequestStatus(
			context.TODO(),
			c.extClient.EngineV1alpha1(),
			req.ObjectMeta,
			func(status *api.GCPAccessKeyRequestStatus) *api.GCPAccessKeyRequestStatus {
				status.Conditions = kmapi.SetCondition(status.Conditions, kmapi.Condition{
					Type:               kmapi.ConditionFailure,
					Status:             kmapi.ConditionTrue,
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

	// Create k8s secret with the issued credentials
	secretName := rand.WithUniqSuffix(name)
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
					Status:             kmapi.ConditionTrue,
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

	// Set lease info & k8s secret ref at AccessKeyRequest's status
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

	roleName := getSecretAccessRoleName(api.ResourceKindGCPAccessKeyRequest, ns, req.Name)

	err = gcpCM.CreateRole(roleName, ns, secretName)
	if err != nil {
		_, err2 := patchutil.UpdateGCPAccessKeyRequestStatus(
			context.TODO(),
			c.extClient.EngineV1alpha1(),
			req.ObjectMeta,
			func(status *api.GCPAccessKeyRequestStatus) *api.GCPAccessKeyRequestStatus {
				status.Conditions = kmapi.SetCondition(status.Conditions, kmapi.Condition{
					Type:               kmapi.ConditionFailure,
					Status:             kmapi.ConditionTrue,
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
					Status:             kmapi.ConditionTrue,
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
			status.Conditions = kmapi.SetCondition(kmapi.RemoveCondition(status.Conditions, kmapi.ConditionFailure), kmapi.Condition{
				Type:               kmapi.ConditionAvailable,
				Status:             kmapi.ConditionTrue,
				Message:            "The requested credentials successfully issued.",
				Reason:             "SuccessfullyIssuedCredential",
				LastTransitionTime: metav1.Now(),
			})
			return status
		},
		metav1.UpdateOptions{},
	)
	return err
}

func (c *VaultController) runGCPAccessKeyRequestFinalizer(req *api.GCPAccessKeyRequest) error {
	if req == nil {
		return errors.New("GCPAccessKeyRequest is nil")
	}

	glog.Infof("Processing finalizer for GCPAccessKeyRequest %s/%s", req.Namespace, req.Name)

	gcpCM, err := credential.NewCredentialManagerForGCP(c.kubeClient, c.appCatalogClient, c.extClient, req)

	// The error could be generated for:
	// 	- invalid roleRef
	// 		- invalid vaultRef in role object
	// In both cases, the operator should be able to delete the AccessKeyRequest(ie. remove finalizer).
	// Revoke the lease if no error occurred.
	if err == nil {
		err = c.finalizeGCPAccessKeyRequest(gcpCM, req.Status.Lease)
		if err != nil {
			return errors.Errorf("GCPAccessKeyRequest %s/%s finalizer: %v", req.Namespace, req.Name, err)
		}
	}

	err = c.removeGCPAccessKeyRequestFinalizer(req)
	if err != nil {
		return errors.Errorf("GCPAccessKeyRequest %s/%s finalizer: %v", req.Namespace, req.Name, err)
	} else {
		glog.Infof("Removed finalizer for GCPAccessKeyRequest %s/%s", req.Namespace, req.Name)
	}

	return nil
}

func (c *VaultController) finalizeGCPAccessKeyRequest(gcpCM credential.CredentialManager, lease *api.Lease) error {
	return revokeLease(gcpCM, lease)
}

func (c *VaultController) removeGCPAccessKeyRequestFinalizer(gcpAKReq *api.GCPAccessKeyRequest) error {
	accessReq, err := c.extClient.EngineV1alpha1().GCPAccessKeyRequests(gcpAKReq.Namespace).Get(context.TODO(), gcpAKReq.Name, metav1.GetOptions{})
	if kerr.IsNotFound(err) {
		return nil
	} else if err != nil {
		return err
	}

	_, _, err = patchutil.PatchGCPAccessKeyRequest(context.TODO(), c.extClient.EngineV1alpha1(), accessReq, func(in *api.GCPAccessKeyRequest) *api.GCPAccessKeyRequest {
		in.ObjectMeta = core_util.RemoveFinalizer(in.ObjectMeta, KubeVaultFinalizer)
		return in
	}, metav1.PatchOptions{})
	return err
}

func (c *VaultController) deleteCredSecretForGCPAccessKeyRequest(gcpAKReq *api.GCPAccessKeyRequest) error {
	if gcpAKReq == nil {
		return errors.New("gcpAccessKeyRequest object is empty")
	}

	// if secret reference is nil, there is nothing to delete.
	if gcpAKReq.Status.Secret == nil {
		return nil
	}

	// Delete the secret if exists.
	return c.kubeClient.CoreV1().Secrets(gcpAKReq.Namespace).Delete(context.TODO(), gcpAKReq.Status.Secret.Name, metav1.DeleteOptions{})
}

func gcpAccessKeyRequestSuccessfullyProcessed(gcpAKReq *api.GCPAccessKeyRequest) bool {
	if gcpAKReq == nil {
		return false
	}
	// If conditions is empty (ie. enqueued for the first time), return false
	// If secret reference is empty, return false
	if gcpAKReq.Status.Conditions == nil || gcpAKReq.Status.Secret == nil {
		return false
	}

	// lookup for failed condition
	for _, cond := range gcpAKReq.Status.Conditions {
		if cond.Type == kmapi.ConditionFailure {
			return false
		}
	}

	// successfully processed
	return true
}
