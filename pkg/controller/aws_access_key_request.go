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
	AWSAccessKeyRequestFinalizer = "awsaccesskeyrequest.engine.kubevault.com"
)

func (c *VaultController) initAWSAccessKeyWatcher() {
	c.awsAccessInformer = c.extInformerFactory.Engine().V1alpha1().AWSAccessKeyRequests().Informer()
	c.awsAccessQueue = queue.New(api.ResourceKindAWSAccessKeyRequest, c.MaxNumRequeues, c.NumThreads, c.runAWSAccessKeyRequestInjector)
	c.awsAccessInformer.AddEventHandler(queue.NewEventHandler(c.awsAccessQueue.GetQueue(), func(oldObj, newObj interface{}) bool {
		old := oldObj.(*api.AWSAccessKeyRequest)
		nu := newObj.(*api.AWSAccessKeyRequest)

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
	c.awsAccessLister = c.extInformerFactory.Engine().V1alpha1().AWSAccessKeyRequests().Lister()
}

func (c *VaultController) runAWSAccessKeyRequestInjector(key string) error {
	obj, exist, err := c.awsAccessInformer.GetIndexer().GetByKey(key)
	if err != nil {
		glog.Errorf("Fetching object with key %s from store failed with %v", key, err)
		return err
	}

	if !exist {
		glog.Warningf("AWSAccessKeyRequest %s does not exist anymore", key)

	} else {
		req := obj.(*api.AWSAccessKeyRequest).DeepCopy()

		glog.Infof("Sync/Add/Update for AWSAccessKeyRequest %s/%s", req.Namespace, req.Name)

		if req.DeletionTimestamp != nil {
			if core_util.HasFinalizer(req.ObjectMeta, AWSAccessKeyRequestFinalizer) {
				go c.runAWSAccessKeyRequestFinalizer(req, finalizerTimeout, finalizerInterval)
			}
		} else {
			if !core_util.HasFinalizer(req.ObjectMeta, AWSAccessKeyRequestFinalizer) {
				// Add finalizer
				_, _, err = patchutil.PatchAWSAccessKeyRequest(context.TODO(), c.extClient.EngineV1alpha1(), req, func(binding *api.AWSAccessKeyRequest) *api.AWSAccessKeyRequest {
					binding.ObjectMeta = core_util.AddFinalizer(binding.ObjectMeta, AWSAccessKeyRequestFinalizer)
					return binding
				}, metav1.PatchOptions{})
				if err != nil {
					return errors.Wrapf(err, "failed to set AWSAccessKeyRequest finalizer for %s/%s", req.Namespace, req.Name)
				}
			}

			var condType string
			for _, c := range req.Status.Conditions {
				if c.Type == kmapi.ConditionRequestApproved || c.Type == kmapi.ConditionRequestDenied {
					condType = c.Type
				}
			}

			if condType == kmapi.ConditionRequestApproved {
				awsCredManager, err := credential.NewCredentialManagerForAWS(c.kubeClient, c.appCatalogClient, c.extClient, req)
				if err != nil {
					return err
				}

				err = c.reconcileAWSAccessKeyRequest(awsCredManager, req)
				if err != nil {
					return errors.Wrapf(err, "For AWSAccessKeyRequest %s/%s", req.Namespace, req.Name)
				}
			} else if condType == kmapi.ConditionRequestDenied {
				glog.Infof("For AWSAccessKeyRequest %s/%s: request is denied", req.Namespace, req.Name)
			} else {
				glog.Infof("For AWSAccessKeyRequest %s/%s: request is not approved yet", req.Namespace, req.Name)
			}
		}
	}
	return nil
}

// Will do:
//	For vault:
//	  - get aws credential
//	  - create secret containing credential
//	  - create rbac role and role binding
//    - sync role binding
func (c *VaultController) reconcileAWSAccessKeyRequest(awsCM credential.CredentialManager, req *api.AWSAccessKeyRequest) error {
	var (
		name = req.Name
		ns   = req.Namespace
		//status = awsAccessReq.Status
	)

	var secretName string
	if req.Status.Secret != nil {
		secretName = req.Status.Secret.Name
	}

	// check whether lease id exists in .status.lease or not
	// if does not exist in .status.lease, then get credential
	if req.Status.Lease == nil {
		// get aws credential secret
		credSecret, err := awsCM.GetCredential()
		if err != nil {
			_, err2 := patchutil.UpdateAWSAccessKeyRequestStatus(
				context.TODO(),
				c.extClient.EngineV1alpha1(),
				req.ObjectMeta,
				func(status *api.AWSAccessKeyRequestStatus) *api.AWSAccessKeyRequestStatus {
					status.Conditions = kmapi.SetCondition(status.Conditions, kmapi.Condition{
						Type:    kmapi.ConditionFailure,
						Reason:  "FailedToGetCredential",
						Message: err.Error(),
					})
					return status
				},
				metav1.UpdateOptions{},
			)
			return utilerrors.NewAggregate([]error{err2, err})
		}

		secretName = rand.WithUniqSuffix(name)
		err = awsCM.CreateSecret(secretName, ns, credSecret)
		if err != nil {
			err2 := awsCM.RevokeLease(credSecret.LeaseID)
			if err2 != nil {
				return errors.Wrapf(err2, "failed to revoke lease")
			}

			_, err2 = patchutil.UpdateAWSAccessKeyRequestStatus(
				context.TODO(),
				c.extClient.EngineV1alpha1(),
				req.ObjectMeta,
				func(status *api.AWSAccessKeyRequestStatus) *api.AWSAccessKeyRequestStatus {
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

		_, err = patchutil.UpdateAWSAccessKeyRequestStatus(
			context.TODO(),
			c.extClient.EngineV1alpha1(),
			req.ObjectMeta,
			func(status *api.AWSAccessKeyRequestStatus) *api.AWSAccessKeyRequestStatus {
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

	roleName := getSecretAccessRoleName(api.ResourceKindAWSAccessKeyRequest, ns, req.Name)

	err := awsCM.CreateRole(roleName, ns, secretName)
	if err != nil {
		_, err2 := patchutil.UpdateAWSAccessKeyRequestStatus(
			context.TODO(),
			c.extClient.EngineV1alpha1(),
			req.ObjectMeta,
			func(status *api.AWSAccessKeyRequestStatus) *api.AWSAccessKeyRequestStatus {
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

	err = awsCM.CreateRoleBinding(roleName, ns, roleName, req.Spec.Subjects)
	if err != nil {
		_, err2 := patchutil.UpdateAWSAccessKeyRequestStatus(
			context.TODO(),
			c.extClient.EngineV1alpha1(),
			req.ObjectMeta,
			func(status *api.AWSAccessKeyRequestStatus) *api.AWSAccessKeyRequestStatus {
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

	_, err = patchutil.UpdateAWSAccessKeyRequestStatus(
		context.TODO(),
		c.extClient.EngineV1alpha1(),
		req.ObjectMeta,
		func(status *api.AWSAccessKeyRequestStatus) *api.AWSAccessKeyRequestStatus {
			status.Conditions = kmapi.RemoveCondition(status.Conditions, kmapi.ConditionFailure)
			return status
		},
		metav1.UpdateOptions{},
	)
	return err
}

func (c *VaultController) runAWSAccessKeyRequestFinalizer(req *api.AWSAccessKeyRequest, timeout time.Duration, interval time.Duration) {
	if req == nil {
		glog.Infoln("AWSAccessKeyRequest is nil")
		return
	}

	id := getAWSAccessKeyRequestId(req)
	if c.finalizerInfo.IsAlreadyProcessing(id) {
		// already processing
		return
	}

	glog.Infof("Processing finalizer for AWSAccessKeyRequest %s/%s", req.Namespace, req.Name)
	// Add key to finalizerInfo, it will prevent other go routine to processing for this AWSAccessKeyRequest
	c.finalizerInfo.Add(id)

	stopCh := time.After(timeout)
	finalizationDone := false
	timeOutOccured := false
	attempt := 0

	for {
		glog.Infof("AWSAccessKeyRequest %s/%s finalizer: attempt %d\n", req.Namespace, req.Name, attempt)

		select {
		case <-stopCh:
			timeOutOccured = true
		default:
		}

		if timeOutOccured {
			break
		}

		if !finalizationDone {
			awsCM, err := credential.NewCredentialManagerForAWS(c.kubeClient, c.appCatalogClient, c.extClient, req)
			if err != nil {
				glog.Errorf("AWSAccessKeyRequest %s/%s finalizer: %v", req.Namespace, req.Name, err)
			} else {
				err = c.finalizeAWSAccessKeyRequest(awsCM, req.Status.Lease)
				if err != nil {
					glog.Errorf("AWSAccessKeyRequest %s/%s finalizer: %v", req.Namespace, req.Name, err)
				} else {
					finalizationDone = true
				}
			}
		}

		if finalizationDone {
			err := c.removeAWSAccessKeyRequestFinalizer(req)
			if err != nil {
				glog.Errorf("AWSAccessKeyRequest %s/%s finalizer: %v", req.Namespace, req.Name, err)
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

	err := c.removeAWSAccessKeyRequestFinalizer(req)
	if err != nil {
		glog.Errorf("AWSAccessKeyRequest %s/%s finalizer: %v", req.Namespace, req.Name, err)
	} else {
		glog.Infof("Removed finalizer for AWSAccessKeyRequest %s/%s", req.Namespace, req.Name)
	}

	// Delete key from finalizer info as processing is done
	c.finalizerInfo.Delete(id)
}

func (c *VaultController) finalizeAWSAccessKeyRequest(awsCM credential.CredentialManager, lease *api.Lease) error {
	if lease == nil {
		return nil
	}
	if lease.ID == "" {
		return nil
	}
	return awsCM.RevokeLease(lease.ID)
}

func (c *VaultController) removeAWSAccessKeyRequestFinalizer(awsAKReq *api.AWSAccessKeyRequest) error {
	accessReq, err := c.extClient.EngineV1alpha1().AWSAccessKeyRequests(awsAKReq.Namespace).Get(context.TODO(), awsAKReq.Name, metav1.GetOptions{})
	if kerr.IsNotFound(err) {
		return nil
	} else if err != nil {
		return err
	}

	_, _, err = patchutil.PatchAWSAccessKeyRequest(context.TODO(), c.extClient.EngineV1alpha1(), accessReq, func(in *api.AWSAccessKeyRequest) *api.AWSAccessKeyRequest {
		in.ObjectMeta = core_util.RemoveFinalizer(in.ObjectMeta, AWSAccessKeyRequestFinalizer)
		return in
	}, metav1.PatchOptions{})
	return err
}

func getAWSAccessKeyRequestId(awsAKReq *api.AWSAccessKeyRequest) string {
	return fmt.Sprintf("%s/%s/%s", api.ResourceAWSAccessKeyRequest, awsAKReq.Namespace, awsAKReq.Name)
}
