package controller

import (
	"fmt"
	"time"

	"github.com/golang/glog"
	"github.com/pkg/errors"
	core "k8s.io/api/core/v1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	core_util "kmodules.xyz/client-go/core/v1"
	"kmodules.xyz/client-go/tools/queue"
	api "kubevault.dev/operator/apis/engine/v1alpha1"
	patchutil "kubevault.dev/operator/client/clientset/versioned/typed/engine/v1alpha1/util"
	"kubevault.dev/operator/pkg/vault/role/gcp"
)

const (
	GCPRolePhaseSuccess    api.GCPRolePhase = "Success"
	GCPRoleConditionFailed                  = "Failed"
	GCPRoleFinalizer                        = "gcprole.engine.kubevault.com"
)

func (c *VaultController) initGCPRoleWatcher() {
	c.gcpRoleInformer = c.extInformerFactory.Engine().V1alpha1().GCPRoles().Informer()
	c.gcpRoleQueue = queue.New(api.ResourceKindGCPRole, c.MaxNumRequeues, c.NumThreads, c.runGCPRoleInjector)
	c.gcpRoleInformer.AddEventHandler(queue.NewReconcilableHandler(c.gcpRoleQueue.GetQueue()))
	c.gcpRoleLister = c.extInformerFactory.Engine().V1alpha1().GCPRoles().Lister()
}

func (c *VaultController) runGCPRoleInjector(key string) error {
	obj, exist, err := c.gcpRoleInformer.GetIndexer().GetByKey(key)
	if err != nil {
		glog.Errorf("Fetching object with key %s from store failed with %v", key, err)
		return err
	}

	if !exist {
		glog.Warningf("GCPRole %s does not exist anymore", key)

	} else {
		gcpRole := obj.(*api.GCPRole).DeepCopy()

		glog.Infof("Sync/Add/Update for GCPRole %s/%s", gcpRole.Namespace, gcpRole.Name)

		if gcpRole.DeletionTimestamp != nil {
			if core_util.HasFinalizer(gcpRole.ObjectMeta, GCPRoleFinalizer) {
				go c.runGCPRoleFinalizer(gcpRole, finalizerTimeout, finalizerInterval)
			}
		} else {
			if !core_util.HasFinalizer(gcpRole.ObjectMeta, GCPRoleFinalizer) {
				// Add finalizer
				_, _, err := patchutil.PatchGCPRole(c.extClient.EngineV1alpha1(), gcpRole, func(role *api.GCPRole) *api.GCPRole {
					role.ObjectMeta = core_util.AddFinalizer(role.ObjectMeta, GCPRoleFinalizer)
					return role
				})
				if err != nil {
					return errors.Wrapf(err, "failed to set GCPRole finalizer for %s/%s", gcpRole.Namespace, gcpRole.Name)
				}
			}

			gcpRClient, err := gcp.NewGCPRole(c.kubeClient, c.appCatalogClient, gcpRole)
			if err != nil {
				return err
			}

			err = c.reconcileGCPRole(gcpRClient, gcpRole)
			if err != nil {
				return errors.Wrapf(err, "for GCPRole %s/%s:", gcpRole.Namespace, gcpRole.Name)
			}
		}
	}
	return nil
}

// Will do:
//	For vault:
// 	  - configure a GCP role
//    - sync role
func (c *VaultController) reconcileGCPRole(gcpRClient gcp.GCPRoleInterface, gcpRole *api.GCPRole) error {
	status := gcpRole.Status

	// create role
	err := gcpRClient.CreateRole()
	if err != nil {
		status.Conditions = []api.GCPRoleCondition{
			{
				Type:    GCPRoleConditionFailed,
				Status:  core.ConditionTrue,
				Reason:  "FailedToCreateRole",
				Message: err.Error(),
			},
		}

		err2 := c.updatedGCPRoleStatus(&status, gcpRole)
		if err2 != nil {
			return errors.Wrap(err2, "failed to update status")
		}
		return errors.Wrap(err, "failed to create role")
	}

	status.Conditions = []api.GCPRoleCondition{}
	status.Phase = GCPRolePhaseSuccess
	status.ObservedGeneration = gcpRole.Generation

	err = c.updatedGCPRoleStatus(&status, gcpRole)
	if err != nil {
		return errors.Wrapf(err, "failed to update GCPRole status")
	}
	return nil
}

func (c *VaultController) updatedGCPRoleStatus(status *api.GCPRoleStatus, gcpRole *api.GCPRole) error {
	_, err := patchutil.UpdateGCPRoleStatus(c.extClient.EngineV1alpha1(), gcpRole, func(s *api.GCPRoleStatus) *api.GCPRoleStatus {
		s = status
		return s
	})
	return err
}

func (c *VaultController) runGCPRoleFinalizer(gcpRole *api.GCPRole, timeout time.Duration, interval time.Duration) {
	if gcpRole == nil {
		glog.Infoln("GCPRole is nil")
		return
	}

	id := getGCPRoleId(gcpRole)
	if c.finalizerInfo.IsAlreadyProcessing(id) {
		// already processing
		return
	}

	glog.Infof("Processing finalizer for GCPRole %s/%s", gcpRole.Namespace, gcpRole.Name)
	// Add key to finalizerInfo, it will prevent other go routine to processing for this GCPRole
	c.finalizerInfo.Add(id)

	stopCh := time.After(timeout)
	finalizationDone := false
	timeOutOccured := false
	attempt := 0

	for {
		glog.Infof("GCPRole %s/%s finalizer: attempt %d\n", gcpRole.Namespace, gcpRole.Name, attempt)

		select {
		case <-stopCh:
			timeOutOccured = true
		default:
		}

		if timeOutOccured {
			break
		}

		if !finalizationDone {
			d, err := gcp.NewGCPRole(c.kubeClient, c.appCatalogClient, gcpRole)
			if err != nil {
				glog.Errorf("GCPRole %s/%s finalizer: %v", gcpRole.Namespace, gcpRole.Name, err)
			} else {
				err = c.finalizeGCPRole(d, gcpRole)
				if err != nil {
					glog.Errorf("GCPRole %s/%s finalizer: %v", gcpRole.Namespace, gcpRole.Name, err)
				} else {
					finalizationDone = true
				}
			}
		}

		if finalizationDone {
			err := c.removeGCPRoleFinalizer(gcpRole)
			if err != nil {
				glog.Errorf("GCPRole %s/%s finalizer: removing finalizer %v", gcpRole.Namespace, gcpRole.Name, err)
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

	err := c.removeGCPRoleFinalizer(gcpRole)
	if err != nil {
		glog.Errorf("GCPRole %s/%s finalizer: removing finalizer %v", gcpRole.Namespace, gcpRole.Name, err)
	} else {
		glog.Infof("Removed finalizer for GCPRole %s/%s", gcpRole.Namespace, gcpRole.Name)
	}

	// Delete key from finalizer info as processing is done
	c.finalizerInfo.Delete(id)
}

// Do:
//	- delete role in vault
func (c *VaultController) finalizeGCPRole(gcpRClient gcp.GCPRoleInterface, gcpRole *api.GCPRole) error {
	err := gcpRClient.DeleteRole(gcpRole.RoleName())
	if err != nil {
		return errors.Wrap(err, "failed to delete gcp role")
	}
	return nil
}

func (c *VaultController) removeGCPRoleFinalizer(gcpRole *api.GCPRole) error {
	m, err := c.extClient.EngineV1alpha1().GCPRoles(gcpRole.Namespace).Get(gcpRole.Name, metav1.GetOptions{})
	if kerr.IsNotFound(err) {
		return nil
	} else if err != nil {
		return err
	}

	// remove finalizer
	_, _, err = patchutil.PatchGCPRole(c.extClient.EngineV1alpha1(), m, func(role *api.GCPRole) *api.GCPRole {
		role.ObjectMeta = core_util.RemoveFinalizer(role.ObjectMeta, GCPRoleFinalizer)
		return role
	})
	return err
}

func getGCPRoleId(gcpRole *api.GCPRole) string {
	return fmt.Sprintf("%s/%s/%s", api.ResourceGCPRole, gcpRole.Namespace, gcpRole.Name)
}
