package controller

import (
	"fmt"
	"time"

	"github.com/appscode/go/crypto/rand"
	"github.com/golang/glog"
	"github.com/kubevault/operator/apis"
	api "github.com/kubevault/operator/apis/engine/v1alpha1"
	patchutil "github.com/kubevault/operator/client/clientset/versioned/typed/engine/v1alpha1/util"
	"github.com/kubevault/operator/pkg/vault/credential"
	"github.com/pkg/errors"
	core "k8s.io/api/core/v1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	core_util "kmodules.xyz/client-go/core/v1"
	"kmodules.xyz/client-go/tools/queue"
)

const (
	AWSAccessKeyRequestFailed    api.RequestConditionType = "Failed"
	AWSAccessKeyRequestFinalizer                          = "awsaccesskeyrequest.engine.kubevault.com"
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
		awsAccessReq := obj.(*api.AWSAccessKeyRequest).DeepCopy()

		glog.Infof("Sync/Add/Update for AWSAccessKeyRequest %s/%s", awsAccessReq.Namespace, awsAccessReq.Name)

		if awsAccessReq.DeletionTimestamp != nil {
			if core_util.HasFinalizer(awsAccessReq.ObjectMeta, AWSAccessKeyRequestFinalizer) {
				go c.runAWSAccessKeyRequestFinalizer(awsAccessReq, finalizerTimeout, finalizerInterval)
			}
		} else {
			if !core_util.HasFinalizer(awsAccessReq.ObjectMeta, AWSAccessKeyRequestFinalizer) {
				// Add finalizer
				_, _, err = patchutil.PatchAWSAccessKeyRequest(c.extClient.EngineV1alpha1(), awsAccessReq, func(binding *api.AWSAccessKeyRequest) *api.AWSAccessKeyRequest {
					binding.ObjectMeta = core_util.AddFinalizer(binding.ObjectMeta, AWSAccessKeyRequestFinalizer)
					return binding
				})
				if err != nil {
					return errors.Wrapf(err, "failed to set AWSAccessKeyRequest finalizer for %s/%s", awsAccessReq.Namespace, awsAccessReq.Name)
				}
			}

			var condType api.RequestConditionType
			for _, c := range awsAccessReq.Status.Conditions {
				if c.Type == api.AccessApproved || c.Type == api.AccessDenied {
					condType = c.Type
				}
			}

			if condType == api.AccessApproved {
				awsCredManager, err := credential.NewCredentialManagerForAWS(c.kubeClient, c.appCatalogClient, c.extClient, awsAccessReq)
				if err != nil {
					return err
				}

				err = c.reconcileAWSAccessKeyRequest(awsCredManager, awsAccessReq)
				if err != nil {
					return errors.Wrapf(err, "For AWSAccessKeyRequest %s/%s", awsAccessReq.Namespace, awsAccessReq.Name)
				}
			} else if condType == api.AccessDenied {
				glog.Infof("For AWSAccessKeyRequest %s/%s: request is denied", awsAccessReq.Namespace, awsAccessReq.Name)
			} else {
				glog.Infof("For AWSAccessKeyRequest %s/%s: request is not approved yet", awsAccessReq.Namespace, awsAccessReq.Name)
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
func (c *VaultController) reconcileAWSAccessKeyRequest(awsCM credential.CredentialManager, awsAccessReq *api.AWSAccessKeyRequest) error {
	var (
		name   = awsAccessReq.Name
		ns     = awsAccessReq.Namespace
		status = awsAccessReq.Status
	)

	var secretName string
	if awsAccessReq.Status.Secret != nil {
		secretName = awsAccessReq.Status.Secret.Name
	}

	// check whether lease id exists in .status.lease or not
	// if does not exist in .status.lease, then get credential
	if awsAccessReq.Status.Lease == nil {
		// get aws credential secret
		credSecret, err := awsCM.GetCredential()
		if err != nil {
			status.Conditions = UpsertAWSAccessKeyCondition(status.Conditions, api.AWSAccessKeyRequestCondition{
				Type:           AWSAccessKeyRequestFailed,
				Reason:         "FailedToGetCredential",
				Message:        err.Error(),
				LastUpdateTime: metav1.Now(),
			})

			err2 := c.updateAWSAccessKeyRequestStatus(&status, awsAccessReq)
			if err2 != nil {
				return errors.Wrapf(err2, "failed to update status")
			}
			return errors.WithStack(err)
		}

		secretName = rand.WithUniqSuffix(name)
		err = awsCM.CreateSecret(secretName, ns, credSecret)
		if err != nil {
			err2 := awsCM.RevokeLease(credSecret.LeaseID)
			if err2 != nil {
				return errors.Wrapf(err2, "failed to revoke lease")
			}

			status.Conditions = UpsertAWSAccessKeyCondition(status.Conditions, api.AWSAccessKeyRequestCondition{
				Type:           AWSAccessKeyRequestFailed,
				Reason:         "FailedToCreateSecret",
				Message:        err.Error(),
				LastUpdateTime: metav1.Now(),
			})

			err2 = c.updateAWSAccessKeyRequestStatus(&status, awsAccessReq)
			if err2 != nil {
				return errors.Wrapf(err2, "failed to update status")
			}
			return errors.WithStack(err)
		}

		// add lease info in status
		status.Lease = &api.Lease{
			ID: credSecret.LeaseID,
			Duration: metav1.Duration{
				time.Second * time.Duration(credSecret.LeaseDuration),
			},
			Renewable: credSecret.Renewable,
		}

		// assign secret name
		status.Secret = &core.LocalObjectReference{
			Name: secretName,
		}
	}

	roleName := getSecretAccessRoleName(api.ResourceKindAWSAccessKeyRequest, ns, awsAccessReq.Name)

	err := awsCM.CreateRole(roleName, ns, secretName)
	if err != nil {
		status.Conditions = UpsertAWSAccessKeyCondition(status.Conditions, api.AWSAccessKeyRequestCondition{
			Type:           AWSAccessKeyRequestFailed,
			Reason:         "FailedToCreateRole",
			Message:        err.Error(),
			LastUpdateTime: metav1.Now(),
		})

		err2 := c.updateAWSAccessKeyRequestStatus(&status, awsAccessReq)
		if err2 != nil {
			return errors.Wrapf(err2, "failed to update status")
		}
		return errors.WithStack(err)
	}

	err = awsCM.CreateRoleBinding(roleName, ns, roleName, awsAccessReq.Spec.Subjects)
	if err != nil {
		status.Conditions = UpsertAWSAccessKeyCondition(status.Conditions, api.AWSAccessKeyRequestCondition{
			Type:           AWSAccessKeyRequestFailed,
			Reason:         "FailedToCreateRoleBinding",
			Message:        err.Error(),
			LastUpdateTime: metav1.Now(),
		})

		err2 := c.updateAWSAccessKeyRequestStatus(&status, awsAccessReq)
		if err2 != nil {
			return errors.Wrapf(err2, "failed to update status")
		}
		return errors.WithStack(err)
	}

	status.Conditions = DeleteAWSAccessKeyCondition(status.Conditions, api.RequestConditionType(AWSAccessKeyRequestFailed))
	err = c.updateAWSAccessKeyRequestStatus(&status, awsAccessReq)
	if err != nil {
		return errors.Wrap(err, "failed to update status")
	}
	return nil
}

func (c *VaultController) updateAWSAccessKeyRequestStatus(status *api.AWSAccessKeyRequestStatus, awsAKReq *api.AWSAccessKeyRequest) error {
	_, err := patchutil.UpdateAWSAccessKeyRequestStatus(c.extClient.EngineV1alpha1(), awsAKReq, func(s *api.AWSAccessKeyRequestStatus) *api.AWSAccessKeyRequestStatus {
		s = status
		return s
	}, apis.EnableStatusSubresource)
	return err
}

func (c *VaultController) runAWSAccessKeyRequestFinalizer(awsAKReq *api.AWSAccessKeyRequest, timeout time.Duration, interval time.Duration) {
	if awsAKReq == nil {
		glog.Infoln("AWSAccessKeyRequest is nil")
		return
	}

	id := getAWSAccessKeyRequestId(awsAKReq)
	if c.finalizerInfo.IsAlreadyProcessing(id) {
		// already processing
		return
	}

	glog.Infof("Processing finalizer for AWSAccessKeyRequest %s/%s", awsAKReq.Namespace, awsAKReq.Name)
	// Add key to finalizerInfo, it will prevent other go routine to processing for this AWSAccessKeyRequest
	c.finalizerInfo.Add(id)

	stopCh := time.After(timeout)
	finalizationDone := false
	timeOutOccured := false
	attempt := 0

	for {
		glog.Infof("AWSAccessKeyRequest %s/%s finalizer: attempt %d\n", awsAKReq.Namespace, awsAKReq.Name, attempt)

		select {
		case <-stopCh:
			timeOutOccured = true
		default:
		}

		if timeOutOccured {
			break
		}

		if !finalizationDone {
			awsCM, err := credential.NewCredentialManagerForAWS(c.kubeClient, c.appCatalogClient, c.extClient, awsAKReq)
			if err != nil {
				glog.Errorf("AWSAccessKeyRequest %s/%s finalizer: %v", awsAKReq.Namespace, awsAKReq.Name, err)
			} else {
				err = c.finalizeAWSAccessKeyRequest(awsCM, awsAKReq.Status.Lease)
				if err != nil {
					glog.Errorf("AWSAccessKeyRequest %s/%s finalizer: %v", awsAKReq.Namespace, awsAKReq.Name, err)
				} else {
					finalizationDone = true
				}
			}
		}

		if finalizationDone {
			err := c.removeAWSAccessKeyRequestFinalizer(awsAKReq)
			if err != nil {
				glog.Errorf("AWSAccessKeyRequest %s/%s finalizer: %v", awsAKReq.Namespace, awsAKReq.Name, err)
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

	err := c.removeAWSAccessKeyRequestFinalizer(awsAKReq)
	if err != nil {
		glog.Errorf("AWSAccessKeyRequest %s/%s finalizer: %v", awsAKReq.Namespace, awsAKReq.Name, err)
	} else {
		glog.Infof("Removed finalizer for AWSAccessKeyRequest %s/%s", awsAKReq.Namespace, awsAKReq.Name)
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
	accessReq, err := c.extClient.EngineV1alpha1().AWSAccessKeyRequests(awsAKReq.Namespace).Get(awsAKReq.Name, metav1.GetOptions{})
	if kerr.IsNotFound(err) {
		return nil
	} else if err != nil {
		return err
	}

	_, _, err = patchutil.PatchAWSAccessKeyRequest(c.extClient.EngineV1alpha1(), accessReq, func(in *api.AWSAccessKeyRequest) *api.AWSAccessKeyRequest {
		in.ObjectMeta = core_util.RemoveFinalizer(in.ObjectMeta, AWSAccessKeyRequestFinalizer)
		return in
	})
	return err
}

func getAWSAccessKeyRequestId(awsAKReq *api.AWSAccessKeyRequest) string {
	return fmt.Sprintf("%s/%s/%s", api.ResourceAWSAccessKeyRequest, awsAKReq.Namespace, awsAKReq.Name)
}

func UpsertAWSAccessKeyCondition(condList []api.AWSAccessKeyRequestCondition, cond api.AWSAccessKeyRequestCondition) []api.AWSAccessKeyRequestCondition {
	res := []api.AWSAccessKeyRequestCondition{}
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

func DeleteAWSAccessKeyCondition(condList []api.AWSAccessKeyRequestCondition, condType api.RequestConditionType) []api.AWSAccessKeyRequestCondition {
	res := []api.AWSAccessKeyRequestCondition{}
	for _, c := range condList {
		if c.Type != condType {
			res = append(res, c)
		}
	}
	return res
}
