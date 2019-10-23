package controller

import (
	"fmt"
	"time"

	"kubevault.dev/operator/apis"
	api "kubevault.dev/operator/apis/engine/v1alpha1"
	patchutil "kubevault.dev/operator/client/clientset/versioned/typed/engine/v1alpha1/util"
	"kubevault.dev/operator/pkg/vault/role/database"

	"github.com/golang/glog"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	core_util "kmodules.xyz/client-go/core/v1"
	"kmodules.xyz/client-go/tools/queue"
)

const (
	MySQLRolePhaseSuccess api.MySQLRolePhase = "Success"
)

func (c *VaultController) initMySQLRoleWatcher() {
	c.myRoleInformer = c.extInformerFactory.Engine().V1alpha1().MySQLRoles().Informer()
	c.myRoleQueue = queue.New(api.ResourceKindMySQLRole, c.MaxNumRequeues, c.NumThreads, c.runMySQLRoleInjector)
	c.myRoleInformer.AddEventHandler(queue.NewReconcilableHandler(c.myRoleQueue.GetQueue()))
	c.myRoleLister = c.extInformerFactory.Engine().V1alpha1().MySQLRoles().Lister()
}

func (c *VaultController) runMySQLRoleInjector(key string) error {
	obj, exist, err := c.myRoleInformer.GetIndexer().GetByKey(key)
	if err != nil {
		glog.Errorf("Fetching object with key %s from store failed with %v", key, err)
		return err
	}

	if !exist {
		glog.Warningf("MySQLRole %s does not exist anymore", key)

	} else {
		mRole := obj.(*api.MySQLRole).DeepCopy()

		glog.Infof("Sync/Add/Update for MySQLRole %s/%s", mRole.Namespace, mRole.Name)

		if mRole.DeletionTimestamp != nil {
			if core_util.HasFinalizer(mRole.ObjectMeta, apis.Finalizer) {
				go c.runMySQLRoleFinalizer(mRole, finalizerTimeout, finalizerInterval)
			}

		} else {
			if !core_util.HasFinalizer(mRole.ObjectMeta, apis.Finalizer) {
				// Add finalizer
				_, _, err := patchutil.PatchMySQLRole(c.extClient.EngineV1alpha1(), mRole, func(role *api.MySQLRole) *api.MySQLRole {
					role.ObjectMeta = core_util.AddFinalizer(role.ObjectMeta, apis.Finalizer)
					return role
				})
				if err != nil {
					return errors.Wrapf(err, "failed to set MySQLRole finalizer for %s/%s", mRole.Namespace, mRole.Name)
				}
			}

			dbRClient, err := database.NewDatabaseRoleForMysql(c.kubeClient, c.appCatalogClient, mRole)
			if err != nil {
				return err
			}

			err = c.reconcileMySQLRole(dbRClient, mRole)
			if err != nil {
				return errors.Wrapf(err, "for MySQLRole %s/%s:", mRole.Namespace, mRole.Name)
			}
		}
	}
	return nil
}

// Will do:
//	For vault:
// 	  - configure a role that maps a name in Vault to an SQL statement to execute to create the database credential.
//    - sync role
//	  - revoke previous lease of all the respective mysqlRoleBinding and reissue a new lease
func (c *VaultController) reconcileMySQLRole(dbRClient database.DatabaseRoleInterface, myRole *api.MySQLRole) error {
	status := myRole.Status

	// create role
	err := dbRClient.CreateRole()
	if err != nil {
		status.Conditions = []api.MySQLRoleCondition{
			{
				Type:    "Available",
				Status:  corev1.ConditionFalse,
				Reason:  "FailedToCreateRole",
				Message: err.Error(),
			},
		}

		err2 := c.updatedMySQLRoleStatus(&status, myRole)
		if err2 != nil {
			return errors.Wrap(err2, "failed to update status")
		}
		return errors.Wrap(err, "failed to create role")
	}

	status.Conditions = []api.MySQLRoleCondition{}
	status.Phase = MySQLRolePhaseSuccess
	status.ObservedGeneration = myRole.Generation

	err = c.updatedMySQLRoleStatus(&status, myRole)
	if err != nil {
		return errors.Wrap(err, "failed to update MySQLRole status")
	}
	return nil
}

func (c *VaultController) updatedMySQLRoleStatus(status *api.MySQLRoleStatus, mRole *api.MySQLRole) error {
	_, err := patchutil.UpdateMySQLRoleStatus(c.extClient.EngineV1alpha1(), mRole, func(s *api.MySQLRoleStatus) *api.MySQLRoleStatus {
		return status
	})
	return err
}

func (c *VaultController) runMySQLRoleFinalizer(mRole *api.MySQLRole, timeout time.Duration, interval time.Duration) {
	if mRole == nil {
		glog.Infoln("MySQLRole is nil")
		return
	}

	id := getMySQLRoleId(mRole)
	if c.finalizerInfo.IsAlreadyProcessing(id) {
		// already processing
		return
	}

	glog.Infof("Processing finalizer for MySQLRole %s/%s", mRole.Namespace, mRole.Name)
	// Add key to finalizerInfo, it will prevent other go routine to processing for this MySQLRole
	c.finalizerInfo.Add(id)

	stopCh := time.After(timeout)
	finalizationDone := false
	timeOutOccured := false
	attempt := 0

	for {
		glog.Infof("MySQLRole %s/%s finalizer: attempt %d\n", mRole.Namespace, mRole.Name, attempt)

		select {
		case <-stopCh:
			timeOutOccured = true
		default:
		}

		if timeOutOccured {
			break
		}

		if !finalizationDone {
			d, err := database.NewDatabaseRoleForMysql(c.kubeClient, c.appCatalogClient, mRole)
			if err != nil {
				glog.Errorf("MySQLRole %s/%s finalizer: %v", mRole.Namespace, mRole.Name, err)
			} else {
				err = c.finalizeMySQLRole(d, mRole)
				if err != nil {
					glog.Errorf("MySQLRole %s/%s finalizer: %v", mRole.Namespace, mRole.Name, err)
				} else {
					finalizationDone = true
				}
			}
		}

		if finalizationDone {
			err := c.removeMySQLRoleFinalizer(mRole)
			if err != nil {
				glog.Errorf("MySQLRole %s/%s finalizer: removing finalizer %v", mRole.Namespace, mRole.Name, err)
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

	err := c.removeMySQLRoleFinalizer(mRole)
	if err != nil {
		glog.Errorf("MySQLRole %s/%s finalizer: removing finalizer %v", mRole.Namespace, mRole.Name, err)
	} else {
		glog.Infof("Removed finalizer for MySQLRole %s/%s", mRole.Namespace, mRole.Name)
	}

	// Delete key from finalizer info as processing is done
	c.finalizerInfo.Delete(id)
}

// Do:
//	- delete role in vault
//	- revoke lease of all the corresponding mysqlRoleBinding
func (c *VaultController) finalizeMySQLRole(dbRClient database.DatabaseRoleInterface, mRole *api.MySQLRole) error {
	err := dbRClient.DeleteRole(mRole.RoleName())
	if err != nil {
		return errors.Wrap(err, "failed to database role")
	}
	return nil
}

func (c *VaultController) removeMySQLRoleFinalizer(mRole *api.MySQLRole) error {
	m, err := c.extClient.EngineV1alpha1().MySQLRoles(mRole.Namespace).Get(mRole.Name, metav1.GetOptions{})
	if kerr.IsNotFound(err) {
		return nil
	} else if err != nil {
		return err
	}
	// remove finalizer
	_, _, err = patchutil.PatchMySQLRole(c.extClient.EngineV1alpha1(), m, func(role *api.MySQLRole) *api.MySQLRole {
		role.ObjectMeta = core_util.RemoveFinalizer(role.ObjectMeta, apis.Finalizer)
		return role
	})
	return err
}

func getMySQLRoleId(mRole *api.MySQLRole) string {
	return fmt.Sprintf("%s/%s/%s", api.ResourceMySQLRole, mRole.Namespace, mRole.Name)
}
