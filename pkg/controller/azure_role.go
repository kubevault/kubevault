package controller

import (
	"fmt"
	"time"

	"github.com/appscode/go/encoding/json/types"
	"github.com/golang/glog"
	"github.com/pkg/errors"
	core "k8s.io/api/core/v1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	core_util "kmodules.xyz/client-go/core/v1"
	meta_util "kmodules.xyz/client-go/meta"
	"kmodules.xyz/client-go/tools/queue"
	"kubevault.dev/operator/apis"
	vsapis "kubevault.dev/operator/apis"
	api "kubevault.dev/operator/apis/engine/v1alpha1"
	patchutil "kubevault.dev/operator/client/clientset/versioned/typed/engine/v1alpha1/util"
	"kubevault.dev/operator/pkg/vault/role/azure"
)

const (
	AzureRolePhaseSuccess    api.AzureRolePhase = "Success"
	AzureRoleConditionFailed                    = "Failed"
	AzureRoleFinalizer                          = "azurerole.engine.kubevault.com"
)

func (c *VaultController) initAzureRoleWatcher() {
	c.azureRoleInformer = c.extInformerFactory.Engine().V1alpha1().AzureRoles().Informer()
	c.azureRoleQueue = queue.New(api.ResourceKindAzureRole, c.MaxNumRequeues, c.NumThreads, c.runAzureRoleInjector)
	c.azureRoleInformer.AddEventHandler(queue.NewObservableHandler(c.azureRoleQueue.GetQueue(), apis.EnableStatusSubresource))
	c.azureRoleLister = c.extInformerFactory.Engine().V1alpha1().AzureRoles().Lister()
}

func (c *VaultController) runAzureRoleInjector(key string) error {
	obj, exist, err := c.azureRoleInformer.GetIndexer().GetByKey(key)
	if err != nil {
		glog.Errorf("Fetching object with key %s from store failed with %v", key, err)
		return err
	}

	if !exist {
		glog.Warningf("AzureRole %s does not exist anymore", key)

	} else {
		azureRole := obj.(*api.AzureRole).DeepCopy()

		glog.Infof("Sync/Add/Update for AzureRole %s/%s", azureRole.Namespace, azureRole.Name)

		if azureRole.DeletionTimestamp != nil {
			if core_util.HasFinalizer(azureRole.ObjectMeta, AzureRoleFinalizer) {
				go c.runAzureRoleFinalizer(azureRole, finalizerTimeout, finalizerInterval)
			}
		} else {
			if !core_util.HasFinalizer(azureRole.ObjectMeta, AzureRoleFinalizer) {
				// Add finalizer
				_, _, err := patchutil.PatchAzureRole(c.extClient.EngineV1alpha1(), azureRole, func(role *api.AzureRole) *api.AzureRole {
					role.ObjectMeta = core_util.AddFinalizer(role.ObjectMeta, AzureRoleFinalizer)
					return role
				})
				if err != nil {
					return errors.Wrapf(err, "failed to set AzureRole finalizer for %s/%s", azureRole.Namespace, azureRole.Name)
				}
			}

			azureRClient, err := azure.NewAzureRole(c.kubeClient, c.appCatalogClient, azureRole)
			if err != nil {
				return err
			}

			err = c.reconcileAzureRole(azureRClient, azureRole)
			if err != nil {
				return errors.Wrapf(err, "for AzureRole %s/%s:", azureRole.Namespace, azureRole.Name)
			}
		}
	}
	return nil
}

// Will do:
//	For vault:
// 	  - configure a Azure role
//    - sync role
func (c *VaultController) reconcileAzureRole(azureRClient azure.AzureRoleInterface, azureRole *api.AzureRole) error {
	status := azureRole.Status

	// create role
	err := azureRClient.CreateRole()
	if err != nil {
		status.Conditions = []api.AzureRoleCondition{
			{
				Type:    AzureRoleConditionFailed,
				Status:  core.ConditionTrue,
				Reason:  "FailedToCreateRole",
				Message: err.Error(),
			},
		}

		err2 := c.updatedAzureRoleStatus(&status, azureRole)
		if err2 != nil {
			return errors.Wrap(err2, "failed to update status")
		}
		return errors.Wrap(err, "failed to create role")
	}

	status.Conditions = []api.AzureRoleCondition{}

	status.Phase = AzureRolePhaseSuccess
	status.ObservedGeneration = types.NewIntHash(azureRole.Generation, meta_util.GenerationHash(azureRole))

	err = c.updatedAzureRoleStatus(&status, azureRole)
	if err != nil {
		return errors.Wrapf(err, "failed to update AzureRole status")
	}
	return nil
}

func (c *VaultController) updatedAzureRoleStatus(status *api.AzureRoleStatus, azureRole *api.AzureRole) error {
	_, err := patchutil.UpdateAzureRoleStatus(c.extClient.EngineV1alpha1(), azureRole, func(s *api.AzureRoleStatus) *api.AzureRoleStatus {
		s = status
		return s
	}, vsapis.EnableStatusSubresource)
	return err
}

func (c *VaultController) runAzureRoleFinalizer(azureRole *api.AzureRole, timeout time.Duration, interval time.Duration) {
	if azureRole == nil {
		glog.Infoln("AzureRole is nil")
		return
	}

	id := getAzureRoleId(azureRole)
	if c.finalizerInfo.IsAlreadyProcessing(id) {
		// already processing
		return
	}

	glog.Infof("Processing finalizer for AzureRole %s/%s", azureRole.Namespace, azureRole.Name)
	// Add key to finalizerInfo, it will prevent other go routine to processing for this AzureRole
	c.finalizerInfo.Add(id)

	stopCh := time.After(timeout)
	finalizationDone := false
	timeOutOccured := false
	attempt := 0

	for {
		glog.Infof("AzureRole %s/%s finalizer: attempt %d\n", azureRole.Namespace, azureRole.Name, attempt)

		select {
		case <-stopCh:
			timeOutOccured = true
		default:
		}

		if timeOutOccured {
			break
		}

		if !finalizationDone {
			d, err := azure.NewAzureRole(c.kubeClient, c.appCatalogClient, azureRole)
			if err != nil {
				glog.Errorf("AzureRole %s/%s finalizer: %v", azureRole.Namespace, azureRole.Name, err)
			} else {
				err = c.finalizeAzureRole(d, azureRole)
				if err != nil {
					glog.Errorf("AzureRole %s/%s finalizer: %v", azureRole.Namespace, azureRole.Name, err)
				} else {
					finalizationDone = true
				}
			}
		}

		if finalizationDone {
			err := c.removeAzureRoleFinalizer(azureRole)
			if err != nil {
				glog.Errorf("AzureRole %s/%s finalizer: removing finalizer %v", azureRole.Namespace, azureRole.Name, err)
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

	err := c.removeAzureRoleFinalizer(azureRole)
	if err != nil {
		glog.Errorf("AzureRole %s/%s finalizer: removing finalizer %v", azureRole.Namespace, azureRole.Name, err)
	} else {
		glog.Infof("Removed finalizer for AzureRole %s/%s", azureRole.Namespace, azureRole.Name)
	}

	// Delete key from finalizer info as processing is done
	c.finalizerInfo.Delete(id)
}

// Do:
//	- delete role in vault
func (c *VaultController) finalizeAzureRole(azureRClient azure.AzureRoleInterface, azureRole *api.AzureRole) error {
	err := azureRClient.DeleteRole(azureRole.RoleName())
	if err != nil {
		return errors.Wrap(err, "failed to delete azure role")
	}
	return nil
}

func (c *VaultController) removeAzureRoleFinalizer(azureRole *api.AzureRole) error {
	m, err := c.extClient.EngineV1alpha1().AzureRoles(azureRole.Namespace).Get(azureRole.Name, metav1.GetOptions{})
	if kerr.IsNotFound(err) {
		return nil
	} else if err != nil {
		return err
	}

	// remove finalizer
	_, _, err = patchutil.PatchAzureRole(c.extClient.EngineV1alpha1(), m, func(role *api.AzureRole) *api.AzureRole {
		role.ObjectMeta = core_util.RemoveFinalizer(role.ObjectMeta, AzureRoleFinalizer)
		return role
	})
	return err
}

func getAzureRoleId(azureRole *api.AzureRole) string {
	return fmt.Sprintf("%s/%s/%s", api.ResourceAzureRole, azureRole.Namespace, azureRole.Name)
}
