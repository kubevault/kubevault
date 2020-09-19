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

	"kubevault.dev/operator/apis"
	api "kubevault.dev/operator/apis/engine/v1alpha1"
	patchutil "kubevault.dev/operator/client/clientset/versioned/typed/engine/v1alpha1/util"
	"kubevault.dev/operator/pkg/vault/credential"

	"github.com/appscode/go/crypto/rand"
	"github.com/golang/glog"
	"github.com/pkg/errors"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	kmapi "kmodules.xyz/client-go/api/v1"
	core_util "kmodules.xyz/client-go/core/v1"
	"kmodules.xyz/client-go/tools/queue"
)

func (c *VaultController) initGCPAccessKeyWatcher() {
	c.gcpAccessInformer = c.extInformerFactory.Engine().V1alpha1().GCPAccessKeyRequests().Informer()
	c.gcpAccessQueue = queue.New(api.ResourceKindGCPAccessKeyRequest, c.MaxNumRequeues, c.NumThreads, c.runGCPAccessKeyRequestInjector)
	c.gcpAccessInformer.AddEventHandler(queue.NewReconcilableHandler(c.gcpAccessQueue.GetQueue()))
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
		req := obj.(*api.GCPAccessKeyRequest).DeepCopy()

		glog.Infof("Sync/Add/Update for GCPAccessKeyRequest %s/%s", req.Namespace, req.Name)

		if req.DeletionTimestamp != nil {
			if core_util.HasFinalizer(req.ObjectMeta, apis.Finalizer) {
				return c.runGCPAccessKeyRequestFinalizer(req)
			}
		} else {
			if !core_util.HasFinalizer(req.ObjectMeta, apis.Finalizer) {
				// Add finalizer
				_, _, err = patchutil.PatchGCPAccessKeyRequest(context.TODO(), c.extClient.EngineV1alpha1(), req, func(in *api.GCPAccessKeyRequest) *api.GCPAccessKeyRequest {
					in.ObjectMeta = core_util.AddFinalizer(in.ObjectMeta, apis.Finalizer)
					return in
				}, metav1.PatchOptions{})
				if err != nil {
					return errors.Wrapf(err, "failed to add GCPAccessKeyRequest finalizer for %s/%s", req.Namespace, req.Name)
				}
			}

			var condType string
			for _, c := range req.Status.Conditions {
				if (c.Type == kmapi.ConditionRequestApproved && c.Status == kmapi.ConditionTrue) || (c.Type == kmapi.ConditionRequestDenied && c.Status == kmapi.ConditionTrue) {
					condType = c.Type
				}
			}

			// If condition type is not set yet, set the phase to "WaitingForApproval".
			if condType == "" {
				glog.Infof("For GCPAccessKeyRequest %s/%s: request is not approved/denied yet", req.Namespace, req.Name)

				_, err := patchutil.UpdateGCPAccessKeyRequestStatus(
					context.TODO(),
					c.extClient.EngineV1alpha1(),
					req.ObjectMeta,
					func(status *api.GCPAccessKeyRequestStatus) *api.GCPAccessKeyRequestStatus {
						if status.Phase == "" {
							status.Phase = api.RequestStatusPhaseWaitingForApproval
						}
						return status
					},
					metav1.UpdateOptions{},
				)
				if err != nil {
					return errors.Wrap(err, fmt.Sprintf("failed to update the status of gcpAccessKeyRequest: %s/%s", req.Namespace, req.Name))
				}

				return nil
			}

			if condType == kmapi.ConditionRequestApproved {

				// If accessKeyRequest is successfully processed,
				// skip processing.
				if gcpAccessKeyRequestSuccessfullyProcessed(req) {
					return nil
				}

				// Create credential manager which handle communication to vault server
				cm, err := credential.NewCredentialManagerForGCP(c.kubeClient, c.appCatalogClient, c.extClient, req)
				if err != nil {
					_, err2 := patchutil.UpdateGCPAccessKeyRequestStatus(
						context.TODO(),
						c.extClient.EngineV1alpha1(),
						req.ObjectMeta,
						func(status *api.GCPAccessKeyRequestStatus) *api.GCPAccessKeyRequestStatus {
							status.Conditions = kmapi.SetCondition(status.Conditions, kmapi.Condition{
								Type:               kmapi.ConditionFailed,
								Status:             kmapi.ConditionTrue,
								Reason:             "FailedToCreateCredentialManager",
								Message:            err.Error(),
								LastTransitionTime: metav1.Now(),
							})
							status.ObservedGeneration = req.Generation
							return status
						},
						metav1.UpdateOptions{},
					)

					return utilerrors.NewAggregate([]error{err2, err})
				}

				err = c.reconcileGCPAccessKeyRequest(cm, req)
				// If reconcileGCPAccessKeyRequest fails,
				//	- Revoke lease if any
				// 	- Delete k8s secret if any
				//	- Update lease & secret references with nil value
				if err != nil {
					err1 := revokeLease(cm, req.Status.Lease)
					err2 := c.deleteCredSecretForGCPAccessKeyRequest(req)
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
						req.ObjectMeta,
						func(status *api.GCPAccessKeyRequestStatus) *api.GCPAccessKeyRequestStatus {
							status.Secret = nil
							status.Lease = nil
							status.ObservedGeneration = req.Generation
							return status
						},
						metav1.UpdateOptions{},
					)
					return errors.Wrapf(utilerrors.NewAggregate([]error{err3, err}), "For GCPAccessKeyRequest %s/%s", req.Namespace, req.Name)
				}
			} else if condType == kmapi.ConditionRequestDenied {
				glog.Infof("For GCPAccessKeyRequest %s/%s: request is denied", req.Namespace, req.Name)
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
func (c *VaultController) reconcileGCPAccessKeyRequest(cm credential.CredentialManager, req *api.GCPAccessKeyRequest) error {
	// if lease or secret ref was set during the previous cycle which was failed.
	// return error.
	if req.Status.Lease != nil || req.Status.Secret != nil {
		return errors.New("lease or secret ref is not empty")
	}

	// Get new credentials ( gcp service account key )
	credSecret, err := cm.GetCredential()
	if err != nil {
		_, err2 := patchutil.UpdateGCPAccessKeyRequestStatus(
			context.TODO(),
			c.extClient.EngineV1alpha1(),
			req.ObjectMeta,
			func(status *api.GCPAccessKeyRequestStatus) *api.GCPAccessKeyRequestStatus {
				status.Conditions = kmapi.SetCondition(status.Conditions, kmapi.Condition{
					Type:               kmapi.ConditionFailed,
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
	secretName := rand.WithUniqSuffix(req.Name)
	err = cm.CreateSecret(secretName, req.Namespace, credSecret)
	if err != nil {
		if len(credSecret.LeaseID) != 0 {
			err2 := cm.RevokeLease(credSecret.LeaseID)
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
					Type:               kmapi.ConditionFailed,
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

	roleName := getSecretAccessRoleName(api.ResourceKindGCPAccessKeyRequest, req.Namespace, req.Name)

	err = cm.CreateRole(roleName, req.Namespace, secretName)
	if err != nil {
		_, err2 := patchutil.UpdateGCPAccessKeyRequestStatus(
			context.TODO(),
			c.extClient.EngineV1alpha1(),
			req.ObjectMeta,
			func(status *api.GCPAccessKeyRequestStatus) *api.GCPAccessKeyRequestStatus {
				status.Conditions = kmapi.SetCondition(status.Conditions, kmapi.Condition{
					Type:               kmapi.ConditionFailed,
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

	err = cm.CreateRoleBinding(roleName, req.Namespace, roleName, req.Spec.Subjects)
	if err != nil {
		_, err2 := patchutil.UpdateGCPAccessKeyRequestStatus(
			context.TODO(),
			c.extClient.EngineV1alpha1(),
			req.ObjectMeta,
			func(status *api.GCPAccessKeyRequestStatus) *api.GCPAccessKeyRequestStatus {
				status.Conditions = kmapi.SetCondition(status.Conditions, kmapi.Condition{
					Type:               kmapi.ConditionFailed,
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
			status.Conditions = kmapi.RemoveCondition(status.Conditions, kmapi.ConditionFailed)
			status.Conditions = kmapi.SetCondition(status.Conditions, kmapi.Condition{
				Type:               kmapi.ConditionAvailable,
				Status:             kmapi.ConditionTrue,
				Message:            "The requested credentials successfully issued.",
				Reason:             "SuccessfullyIssuedCredential",
				LastTransitionTime: metav1.Now(),
			})
			status.ObservedGeneration = req.Generation
			return status
		},
		metav1.UpdateOptions{},
	)
	return err
}

func (c *VaultController) runGCPAccessKeyRequestFinalizer(req *api.GCPAccessKeyRequest) error {
	glog.Infof("Processing finalizer for GCPAccessKeyRequest %s/%s", req.Namespace, req.Name)

	cm, err := credential.NewCredentialManagerForGCP(c.kubeClient, c.appCatalogClient, c.extClient, req)
	// The error could be generated for:
	// 	- invalid roleRef
	// 		- invalid vaultRef in role object
	// In both cases, the operator should be able to delete the AccessKeyRequest(ie. remove finalizer).
	// Revoke the lease if no error occurred.
	if err == nil {
		err = revokeLease(cm, req.Status.Lease)
		if err != nil {
			return errors.Errorf("GCPAccessKeyRequest %s/%s finalizer: %v", req.Namespace, req.Name, err)
		}
	} else {
		glog.Warningf("Skipping cleanup for GCPAccessKeyRequest: %s/%s with error: %v", req.Namespace, req.Name, err)
	}

	_, _, err = patchutil.PatchGCPAccessKeyRequest(context.TODO(), c.extClient.EngineV1alpha1(), req, func(in *api.GCPAccessKeyRequest) *api.GCPAccessKeyRequest {
		in.ObjectMeta = core_util.RemoveFinalizer(in.ObjectMeta, apis.Finalizer)
		return in
	}, metav1.PatchOptions{})
	if err != nil {
		return errors.Errorf("GCPAccessKeyRequest %s/%s finalizer: %v", req.Namespace, req.Name, err)
	} else {
		glog.Infof("Removed finalizer for GCPAccessKeyRequest %s/%s", req.Namespace, req.Name)
	}

	return nil
}

func (c *VaultController) deleteCredSecretForGCPAccessKeyRequest(req *api.GCPAccessKeyRequest) error {
	// if secret reference is nil, there is nothing to delete.
	if req.Status.Secret == nil {
		return nil
	}

	// Delete the secret if exists.
	return c.kubeClient.CoreV1().Secrets(req.Namespace).Delete(context.TODO(), req.Status.Secret.Name, metav1.DeleteOptions{})
}

func gcpAccessKeyRequestSuccessfullyProcessed(req *api.GCPAccessKeyRequest) bool {
	// If conditions is empty (ie. enqueued for the first time), return false
	// If secret reference is empty, return false
	if req.Status.Conditions == nil || req.Status.Secret == nil {
		return false
	}

	// lookup for failed condition
	for _, cond := range req.Status.Conditions {
		if cond.Type == kmapi.ConditionFailed {
			return false
		}
	}

	// successfully processed
	return true
}
