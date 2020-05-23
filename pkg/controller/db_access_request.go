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
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	kmapi "kmodules.xyz/client-go/api/v1"
	core_util "kmodules.xyz/client-go/core/v1"
	"kmodules.xyz/client-go/tools/queue"
)

func (c *VaultController) initDatabaseAccessWatcher() {
	c.dbAccessInformer = c.extInformerFactory.Engine().V1alpha1().DatabaseAccessRequests().Informer()
	c.dbAccessQueue = queue.New(api.ResourceKindDatabaseAccessRequest, c.MaxNumRequeues, c.NumThreads, c.runDatabaseAccessRequestInjector)
	c.dbAccessInformer.AddEventHandler(queue.NewEventHandler(c.dbAccessQueue.GetQueue(), func(oldObj, newObj interface{}) bool {
		old := oldObj.(*api.DatabaseAccessRequest)
		nu := newObj.(*api.DatabaseAccessRequest)

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
	c.dbAccessLister = c.extInformerFactory.Engine().V1alpha1().DatabaseAccessRequests().Lister()
}

func (c *VaultController) runDatabaseAccessRequestInjector(key string) error {
	obj, exist, err := c.dbAccessInformer.GetIndexer().GetByKey(key)
	if err != nil {
		glog.Errorf("Fetching object with key %s from store failed with %v", key, err)
		return err
	}

	if !exist {
		glog.Warningf("DatabaseAccessRequest %s does not exist anymore", key)

	} else {
		dbAccessReq := obj.(*api.DatabaseAccessRequest).DeepCopy()

		glog.Infof("Sync/Add/Update for DatabaseAccessRequest %s/%s", dbAccessReq.Namespace, dbAccessReq.Name)

		if dbAccessReq.DeletionTimestamp != nil {
			if core_util.HasFinalizer(dbAccessReq.ObjectMeta, apis.Finalizer) {
				go c.runDatabaseAccessRequestFinalizer(dbAccessReq, finalizerTimeout, finalizerInterval)
			}
		} else {
			if !core_util.HasFinalizer(dbAccessReq.ObjectMeta, apis.Finalizer) {
				// Add finalizer
				_, _, err = patchutil.PatchDatabaseAccessRequest(context.TODO(), c.extClient.EngineV1alpha1(), dbAccessReq, func(binding *api.DatabaseAccessRequest) *api.DatabaseAccessRequest {
					binding.ObjectMeta = core_util.AddFinalizer(binding.ObjectMeta, apis.Finalizer)
					return binding
				}, metav1.PatchOptions{})
				if err != nil {
					return errors.Wrapf(err, "failed to set DatabaseAccessRequest finalizer for %s/%s", dbAccessReq.Namespace, dbAccessReq.Name)
				}
			}

			var condType string
			for _, c := range dbAccessReq.Status.Conditions {
				if c.Type == kmapi.ConditionRequestApproved || c.Type == kmapi.ConditionRequestDenied {
					condType = c.Type
				}
			}

			if condType == kmapi.ConditionRequestApproved {
				dbCredManager, err := credential.NewCredentialManagerForDatabase(c.kubeClient, c.appCatalogClient, c.extClient, dbAccessReq)
				if err != nil {
					return err
				}

				err = c.reconcileDatabaseAccessRequest(dbCredManager, dbAccessReq)
				if err != nil {
					return errors.Wrapf(err, "For DatabaseAccessRequest %s/%s", dbAccessReq.Namespace, dbAccessReq.Name)
				}
			} else if condType == kmapi.ConditionRequestDenied {
				glog.Infof("For DatabaseAccessRequest %s/%s: request is denied", dbAccessReq.Namespace, dbAccessReq.Name)
			} else {
				glog.Infof("For DatabaseAccessRequest %s/%s: request is not approved yet", dbAccessReq.Namespace, dbAccessReq.Name)
			}
		}
	}
	return nil
}

// Will do:
//	For vault:
//	  - get db credential
//	  - create secret containing credential
//	  - create rbac role and role binding
//    - sync role binding
func (c *VaultController) reconcileDatabaseAccessRequest(dbCM credential.CredentialManager, req *api.DatabaseAccessRequest) error {
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
		// get database credential secret
		credSecret, err := dbCM.GetCredential()
		if err != nil {
			_, err2 := patchutil.UpdateDatabaseAccessRequestStatus(
				context.TODO(),
				c.extClient.EngineV1alpha1(),
				req.ObjectMeta,
				func(status *api.DatabaseAccessRequestStatus) *api.DatabaseAccessRequestStatus {
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
		err = dbCM.CreateSecret(secretName, ns, credSecret)
		if err != nil {
			err2 := dbCM.RevokeLease(credSecret.LeaseID)
			if err2 != nil {
				return errors.Wrapf(err2, "failed to revoke lease")
			}

			_, err2 = patchutil.UpdateDatabaseAccessRequestStatus(
				context.TODO(),
				c.extClient.EngineV1alpha1(),
				req.ObjectMeta,
				func(status *api.DatabaseAccessRequestStatus) *api.DatabaseAccessRequestStatus {
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

		_, err = patchutil.UpdateDatabaseAccessRequestStatus(
			context.TODO(),
			c.extClient.EngineV1alpha1(),
			req.ObjectMeta,
			func(status *api.DatabaseAccessRequestStatus) *api.DatabaseAccessRequestStatus {
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

	roleName := getSecretAccessRoleName(api.ResourceKindDatabaseAccessRequest, ns, req.Name)

	err := dbCM.CreateRole(roleName, ns, secretName)
	if err != nil {
		_, err2 := patchutil.UpdateDatabaseAccessRequestStatus(
			context.TODO(),
			c.extClient.EngineV1alpha1(),
			req.ObjectMeta,
			func(status *api.DatabaseAccessRequestStatus) *api.DatabaseAccessRequestStatus {
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

	err = dbCM.CreateRoleBinding(roleName, ns, roleName, req.Spec.Subjects)
	if err != nil {
		_, err2 := patchutil.UpdateDatabaseAccessRequestStatus(
			context.TODO(),
			c.extClient.EngineV1alpha1(),
			req.ObjectMeta,
			func(status *api.DatabaseAccessRequestStatus) *api.DatabaseAccessRequestStatus {
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

	_, err = patchutil.UpdateDatabaseAccessRequestStatus(
		context.TODO(),
		c.extClient.EngineV1alpha1(),
		req.ObjectMeta,
		func(status *api.DatabaseAccessRequestStatus) *api.DatabaseAccessRequestStatus {
			status.Conditions = kmapi.RemoveCondition(status.Conditions, kmapi.ConditionFailure)
			return status
		},
		metav1.UpdateOptions{},
	)
	return err
}

func (c *VaultController) runDatabaseAccessRequestFinalizer(req *api.DatabaseAccessRequest, timeout time.Duration, interval time.Duration) {
	if req == nil {
		glog.Infoln("DatabaseAccessRequest is nil")
		return
	}

	id := getDatabaseAccessRequestId(req)
	if c.finalizerInfo.IsAlreadyProcessing(id) {
		// already processing
		return
	}

	glog.Infof("Processing finalizer for DatabaseAccessRequest %s/%s", req.Namespace, req.Name)
	// Add key to finalizerInfo, it will prevent other go routine to processing for this DatabaseAccessRequest
	c.finalizerInfo.Add(id)

	stopCh := time.After(timeout)
	finalizationDone := false
	timeOutOccured := false
	attempt := 0

	for {
		glog.Infof("DatabaseAccessRequest %s/%s finalizer: attempt %d\n", req.Namespace, req.Name, attempt)

		select {
		case <-stopCh:
			timeOutOccured = true
		default:
		}

		if timeOutOccured {
			break
		}

		if !finalizationDone {
			d, err := credential.NewCredentialManagerForDatabase(c.kubeClient, c.appCatalogClient, c.extClient, req)
			if err != nil {
				glog.Errorf("DatabaseAccessRequest %s/%s finalizer: %v", req.Namespace, req.Name, err)
			} else {
				err = c.finalizeDatabaseAccessRequest(d, req.Status.Lease)
				if err != nil {
					glog.Errorf("DatabaseAccessRequest %s/%s finalizer: %v", req.Namespace, req.Name, err)
				} else {
					finalizationDone = true
				}
			}
		}

		if finalizationDone {
			err := c.removeDatabaseAccessRequestFinalizer(req)
			if err != nil {
				glog.Errorf("DatabaseAccessRequest %s/%s finalizer: %v", req.Namespace, req.Name, err)
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

	err := c.removeDatabaseAccessRequestFinalizer(req)
	if err != nil {
		glog.Errorf("DatabaseAccessRequest %s/%s finalizer: %v", req.Namespace, req.Name, err)
	} else {
		glog.Infof("Removed finalizer for DatabaseAccessRequest %s/%s", req.Namespace, req.Name)
	}

	// Delete key from finalizer info as processing is done
	c.finalizerInfo.Delete(id)
}

func (c *VaultController) finalizeDatabaseAccessRequest(dbCM credential.CredentialManager, lease *api.Lease) error {
	if lease == nil {
		return nil
	}
	if lease.ID == "" {
		return nil
	}
	return dbCM.RevokeLease(lease.ID)
}

func (c *VaultController) removeDatabaseAccessRequestFinalizer(dbAReq *api.DatabaseAccessRequest) error {
	d, err := c.extClient.EngineV1alpha1().DatabaseAccessRequests(dbAReq.Namespace).Get(context.TODO(), dbAReq.Name, metav1.GetOptions{})
	if kerr.IsNotFound(err) {
		return nil
	} else if err != nil {
		return err
	}

	_, _, err = patchutil.PatchDatabaseAccessRequest(context.TODO(), c.extClient.EngineV1alpha1(), d, func(in *api.DatabaseAccessRequest) *api.DatabaseAccessRequest {
		in.ObjectMeta = core_util.RemoveFinalizer(in.ObjectMeta, apis.Finalizer)
		return in
	}, metav1.PatchOptions{})
	return err
}

func getDatabaseAccessRequestId(dbAReq *api.DatabaseAccessRequest) string {
	return fmt.Sprintf("%s/%s/%s", api.ResourceDatabaseAccessRequest, dbAReq.Namespace, dbAReq.Name)
}

func getSecretAccessRoleName(kind, namespace, name string) string {
	return fmt.Sprintf("%s-%s-%s-credential-reader", kind, namespace, name)
}
