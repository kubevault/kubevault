package controller

import (
	"fmt"
	"time"

	"github.com/appscode/go/crypto/rand"
	"github.com/golang/glog"
	"github.com/pkg/errors"
	core "k8s.io/api/core/v1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	core_util "kmodules.xyz/client-go/core/v1"
	"kmodules.xyz/client-go/tools/queue"
	"kubevault.dev/operator/apis"
	api "kubevault.dev/operator/apis/engine/v1alpha1"
	patchutil "kubevault.dev/operator/client/clientset/versioned/typed/engine/v1alpha1/util"
	"kubevault.dev/operator/pkg/vault/credential"
)

const RequestFailed api.RequestConditionType = "Failed"

func (c *VaultController) initDatabaseAccessWatcher() {
	c.dbAccessInformer = c.extInformerFactory.Engine().V1alpha1().DatabaseAccessRequests().Informer()
	c.dbAccessQueue = queue.New(api.ResourceKindDatabaseAccessRequest, c.MaxNumRequeues, c.NumThreads, c.runDatabaseAccessRequestInjector)
	c.dbAccessInformer.AddEventHandler(queue.NewEventHandler(c.dbAccessQueue.GetQueue(), func(oldObj, newObj interface{}) bool {
		old := oldObj.(*api.DatabaseAccessRequest)
		nu := newObj.(*api.DatabaseAccessRequest)

		oldCondType := ""
		nuCondType := ""
		for _, c := range old.Status.Conditions {
			if c.Type == api.AccessApproved || c.Type == api.AccessDenied {
				oldCondType = string(c.Type)
			}
		}
		for _, c := range nu.Status.Conditions {
			if c.Type == api.AccessApproved || c.Type == api.AccessDenied {
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
				_, _, err = patchutil.PatchDatabaseAccessRequest(c.extClient.EngineV1alpha1(), dbAccessReq, func(binding *api.DatabaseAccessRequest) *api.DatabaseAccessRequest {
					binding.ObjectMeta = core_util.AddFinalizer(binding.ObjectMeta, apis.Finalizer)
					return binding
				})
				if err != nil {
					return errors.Wrapf(err, "failed to set DatabaseAccessRequest finalizer for %s/%s", dbAccessReq.Namespace, dbAccessReq.Name)
				}
			}

			var condType api.RequestConditionType
			for _, c := range dbAccessReq.Status.Conditions {
				if c.Type == api.AccessApproved || c.Type == api.AccessDenied {
					condType = c.Type
				}
			}

			if condType == api.AccessApproved {
				dbCredManager, err := credential.NewCredentialManagerForDatabase(c.kubeClient, c.appCatalogClient, c.extClient, dbAccessReq)
				if err != nil {
					return err
				}

				err = c.reconcileDatabaseAccessRequest(dbCredManager, dbAccessReq)
				if err != nil {
					return errors.Wrapf(err, "For DatabaseAccessRequest %s/%s", dbAccessReq.Namespace, dbAccessReq.Name)
				}
			} else if condType == api.AccessDenied {
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
func (c *VaultController) reconcileDatabaseAccessRequest(dbCM credential.CredentialManager, dbAccessReq *api.DatabaseAccessRequest) error {
	var (
		name   = dbAccessReq.Name
		ns     = dbAccessReq.Namespace
		status = dbAccessReq.Status
	)

	var secretName string
	if dbAccessReq.Status.Secret != nil {
		secretName = dbAccessReq.Status.Secret.Name
	}

	// check whether lease id exists in .status.lease or not
	// if does not exist in .status.lease, then get credential
	if dbAccessReq.Status.Lease == nil {
		// get database credential secret
		credSecret, err := dbCM.GetCredential()
		if err != nil {
			status.Conditions = UpsertDatabaseAccessCondition(status.Conditions, api.DatabaseAccessRequestCondition{
				Type:           RequestFailed,
				Reason:         "FailedToGetCredential",
				Message:        err.Error(),
				LastUpdateTime: metav1.Now(),
			})

			err2 := c.updateDatabaseAccessRequestStatus(&status, dbAccessReq)
			if err2 != nil {
				return errors.Wrapf(err2, "failed to update status")
			}
			return errors.WithStack(err)
		}

		secretName = rand.WithUniqSuffix(name)
		err = dbCM.CreateSecret(secretName, ns, credSecret)
		if err != nil {
			err2 := dbCM.RevokeLease(credSecret.LeaseID)
			if err2 != nil {
				return errors.Wrapf(err2, "failed to revoke lease")
			}

			status.Conditions = UpsertDatabaseAccessCondition(status.Conditions, api.DatabaseAccessRequestCondition{
				Type:           RequestFailed,
				Reason:         "FailedToCreateSecret",
				Message:        err.Error(),
				LastUpdateTime: metav1.Now(),
			})

			err2 = c.updateDatabaseAccessRequestStatus(&status, dbAccessReq)
			if err2 != nil {
				return errors.Wrapf(err2, "failed to update status")
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

	roleName := getSecretAccessRoleName(api.ResourceKindDatabaseAccessRequest, ns, dbAccessReq.Name)

	err := dbCM.CreateRole(roleName, ns, secretName)
	if err != nil {
		status.Conditions = UpsertDatabaseAccessCondition(status.Conditions, api.DatabaseAccessRequestCondition{
			Type:           RequestFailed,
			Reason:         "FailedToCreateRole",
			Message:        err.Error(),
			LastUpdateTime: metav1.Now(),
		})

		err2 := c.updateDatabaseAccessRequestStatus(&status, dbAccessReq)
		if err2 != nil {
			return errors.Wrapf(err2, "failed to update status")
		}
		return errors.WithStack(err)
	}

	err = dbCM.CreateRoleBinding(roleName, ns, roleName, dbAccessReq.Spec.Subjects)
	if err != nil {
		status.Conditions = UpsertDatabaseAccessCondition(status.Conditions, api.DatabaseAccessRequestCondition{
			Type:           RequestFailed,
			Reason:         "FailedToCreateRoleBinding",
			Message:        err.Error(),
			LastUpdateTime: metav1.Now(),
		})

		err2 := c.updateDatabaseAccessRequestStatus(&status, dbAccessReq)
		if err2 != nil {
			return errors.Wrapf(err2, "failed to update status")
		}
		return errors.WithStack(err)
	}

	status.Conditions = DeleteDatabaseAccessCondition(status.Conditions, api.RequestConditionType(RequestFailed))
	err = c.updateDatabaseAccessRequestStatus(&status, dbAccessReq)
	if err != nil {
		return errors.Wrap(err, "failed to update status")
	}
	return nil
}

func (c *VaultController) updateDatabaseAccessRequestStatus(status *api.DatabaseAccessRequestStatus, dbAReq *api.DatabaseAccessRequest) error {
	_, err := patchutil.UpdateDatabaseAccessRequestStatus(c.extClient.EngineV1alpha1(), dbAReq, func(s *api.DatabaseAccessRequestStatus) *api.DatabaseAccessRequestStatus {
		return status
	})
	return err
}

func (c *VaultController) runDatabaseAccessRequestFinalizer(dbAReq *api.DatabaseAccessRequest, timeout time.Duration, interval time.Duration) {
	if dbAReq == nil {
		glog.Infoln("DatabaseAccessRequest is nil")
		return
	}

	id := getDatabaseAccessRequestId(dbAReq)
	if c.finalizerInfo.IsAlreadyProcessing(id) {
		// already processing
		return
	}

	glog.Infof("Processing finalizer for DatabaseAccessRequest %s/%s", dbAReq.Namespace, dbAReq.Name)
	// Add key to finalizerInfo, it will prevent other go routine to processing for this DatabaseAccessRequest
	c.finalizerInfo.Add(id)

	stopCh := time.After(timeout)
	finalizationDone := false
	timeOutOccured := false
	attempt := 0

	for {
		glog.Infof("DatabaseAccessRequest %s/%s finalizer: attempt %d\n", dbAReq.Namespace, dbAReq.Name, attempt)

		select {
		case <-stopCh:
			timeOutOccured = true
		default:
		}

		if timeOutOccured {
			break
		}

		if !finalizationDone {
			d, err := credential.NewCredentialManagerForDatabase(c.kubeClient, c.appCatalogClient, c.extClient, dbAReq)
			if err != nil {
				glog.Errorf("DatabaseAccessRequest %s/%s finalizer: %v", dbAReq.Namespace, dbAReq.Name, err)
			} else {
				err = c.finalizeDatabaseAccessRequest(d, dbAReq.Status.Lease)
				if err != nil {
					glog.Errorf("DatabaseAccessRequest %s/%s finalizer: %v", dbAReq.Namespace, dbAReq.Name, err)
				} else {
					finalizationDone = true
				}
			}
		}

		if finalizationDone {
			err := c.removeDatabaseAccessRequestFinalizer(dbAReq)
			if err != nil {
				glog.Errorf("DatabaseAccessRequest %s/%s finalizer: %v", dbAReq.Namespace, dbAReq.Name, err)
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

	err := c.removeDatabaseAccessRequestFinalizer(dbAReq)
	if err != nil {
		glog.Errorf("DatabaseAccessRequest %s/%s finalizer: %v", dbAReq.Namespace, dbAReq.Name, err)
	} else {
		glog.Infof("Removed finalizer for DatabaseAccessRequest %s/%s", dbAReq.Namespace, dbAReq.Name)
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
	d, err := c.extClient.EngineV1alpha1().DatabaseAccessRequests(dbAReq.Namespace).Get(dbAReq.Name, metav1.GetOptions{})
	if kerr.IsNotFound(err) {
		return nil
	} else if err != nil {
		return err
	}

	_, _, err = patchutil.PatchDatabaseAccessRequest(c.extClient.EngineV1alpha1(), d, func(in *api.DatabaseAccessRequest) *api.DatabaseAccessRequest {
		in.ObjectMeta = core_util.RemoveFinalizer(in.ObjectMeta, apis.Finalizer)
		return in
	})
	return err
}

func getDatabaseAccessRequestId(dbAReq *api.DatabaseAccessRequest) string {
	return fmt.Sprintf("%s/%s/%s", api.ResourceDatabaseAccessRequest, dbAReq.Namespace, dbAReq.Name)
}

func getSecretAccessRoleName(kind, namespace, name string) string {
	return fmt.Sprintf("%s-%s-%s-credential-reader", kind, namespace, name)
}

func UpsertDatabaseAccessCondition(condList []api.DatabaseAccessRequestCondition, cond api.DatabaseAccessRequestCondition) []api.DatabaseAccessRequestCondition {
	res := []api.DatabaseAccessRequestCondition{}
	inserted := false
	for _, c := range condList {
		if c.Type == cond.Type {
			res = append(res, cond)
			inserted = true
		} else {
			res = append(res, c)
		}
	}
	if !inserted {
		res = append(res, cond)
	}
	return res
}

func DeleteDatabaseAccessCondition(condList []api.DatabaseAccessRequestCondition, condType api.RequestConditionType) []api.DatabaseAccessRequestCondition {
	res := []api.DatabaseAccessRequestCondition{}
	for _, c := range condList {
		if c.Type != condType {
			res = append(res, c)
		}
	}
	return res
}
