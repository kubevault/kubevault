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
	PostgresRolePhaseSuccess api.PostgresRolePhase = "Success"
)

func (c *VaultController) initPostgresRoleWatcher() {
	c.pgRoleInformer = c.extInformerFactory.Engine().V1alpha1().PostgresRoles().Informer()
	c.pgRoleQueue = queue.New(api.ResourceKindPostgresRole, c.MaxNumRequeues, c.NumThreads, c.runPostgresRoleInjector)
	c.pgRoleInformer.AddEventHandler(queue.NewReconcilableHandler(c.pgRoleQueue.GetQueue()))
	c.pgRoleLister = c.extInformerFactory.Engine().V1alpha1().PostgresRoles().Lister()
}

func (c *VaultController) runPostgresRoleInjector(key string) error {
	obj, exist, err := c.pgRoleInformer.GetIndexer().GetByKey(key)
	if err != nil {
		glog.Errorf("Fetching object with key %s from store failed with %v", key, err)
		return err
	}

	if !exist {
		glog.Warningf("PostgresRole %s does not exist anymore", key)

	} else {
		pgRole := obj.(*api.PostgresRole).DeepCopy()

		glog.Infof("Sync/Add/Update for PostgresRole %s/%s", pgRole.Namespace, pgRole.Name)

		if pgRole.DeletionTimestamp != nil {
			if core_util.HasFinalizer(pgRole.ObjectMeta, apis.Finalizer) {
				go c.runPostgresRoleFinalizer(pgRole, finalizerTimeout, finalizerInterval)
			}

		} else {
			if !core_util.HasFinalizer(pgRole.ObjectMeta, apis.Finalizer) {
				// Add finalizer
				_, _, err := patchutil.PatchPostgresRole(c.extClient.EngineV1alpha1(), pgRole, func(role *api.PostgresRole) *api.PostgresRole {
					role.ObjectMeta = core_util.AddFinalizer(role.ObjectMeta, apis.Finalizer)
					return role
				})
				if err != nil {
					return errors.Wrapf(err, "failed to set postgresRole finalizer for %s/%s", pgRole.Namespace, pgRole.Name)
				}
			}

			dbRClient, err := database.NewDatabaseRoleForPostgres(c.kubeClient, c.appCatalogClient, pgRole)
			if err != nil {
				return err
			}

			err = c.reconcilePostgresRole(dbRClient, pgRole)
			if err != nil {
				return errors.Wrapf(err, "for PostgresRole %s/%s:", pgRole.Namespace, pgRole.Name)
			}
		}
	}
	return nil
}

// Will do:
//	For vault:
// 	  - configure a role that maps a name in Vault to an SQL statement to execute to create the database credential.
//    - sync role
//	  - revoke previous lease of all the respective postgresRoleBinding and reissue a new lease
func (c *VaultController) reconcilePostgresRole(dbRClient database.DatabaseRoleInterface, pgRole *api.PostgresRole) error {
	status := pgRole.Status

	// create role
	err := dbRClient.CreateRole()
	if err != nil {
		status.Conditions = []api.PostgresRoleCondition{
			{
				Type:    "Available",
				Status:  corev1.ConditionFalse,
				Reason:  "FailedToCreateDatabaseRole",
				Message: err.Error(),
			},
		}

		err2 := c.updatePostgresRoleStatus(&status, pgRole)
		if err2 != nil {
			return errors.Wrap(err2, "for postgresRole %s/%s: failed to update status")
		}
		return errors.Wrap(err, "for postgresRole %s/%s: failed to create role")
	}

	status.ObservedGeneration = pgRole.Generation
	status.Conditions = []api.PostgresRoleCondition{}
	status.Phase = PostgresRolePhaseSuccess

	err = c.updatePostgresRoleStatus(&status, pgRole)
	if err != nil {
		return errors.Wrap(err, "failed to update postgresRole status")
	}
	return nil
}

func (c *VaultController) updatePostgresRoleStatus(status *api.PostgresRoleStatus, pgRole *api.PostgresRole) error {
	_, err := patchutil.UpdatePostgresRoleStatus(c.extClient.EngineV1alpha1(), pgRole, func(s *api.PostgresRoleStatus) *api.PostgresRoleStatus {
		return status
	})
	return err
}

func (c *VaultController) runPostgresRoleFinalizer(pgRole *api.PostgresRole, timeout time.Duration, interval time.Duration) {
	if pgRole == nil {
		glog.Infoln("PostgresRole is nil")
		return
	}

	id := getPostgresRoleId(pgRole)
	if c.finalizerInfo.IsAlreadyProcessing(id) {
		// already processing
		return
	}

	glog.Infof("Processing finalizer for PostgresRole %s/%s", pgRole.Namespace, pgRole.Name)
	// Add key to finalizerInfo, it will prevent other go routine to processing for this PostgresRole
	c.finalizerInfo.Add(id)

	stopCh := time.After(timeout)
	finalizationDone := false
	timeOutOccured := false
	attempt := 0

	for {
		glog.Infof("PostgresRole %s/%s finalizer: attempt %d\n", pgRole.Namespace, pgRole.Name, attempt)

		select {
		case <-stopCh:
			timeOutOccured = true
		default:
		}

		if timeOutOccured {
			break
		}

		if !finalizationDone {
			d, err := database.NewDatabaseRoleForPostgres(c.kubeClient, c.appCatalogClient, pgRole)
			if err != nil {
				glog.Errorf("PostgresRole %s/%s finalizer: %v", pgRole.Namespace, pgRole.Name, err)
			} else {
				err = c.finalizePostgresRole(d, pgRole)
				if err != nil {
					glog.Errorf("PostgresRole %s/%s finalizer: %v", pgRole.Namespace, pgRole.Name, err)
				} else {
					finalizationDone = true
				}
			}
		}

		if finalizationDone {
			err := c.removePostgresRoleFinalizer(pgRole)
			if err != nil {
				glog.Errorf("PostgresRole %s/%s finalizer: removing finalizer %v", pgRole.Namespace, pgRole.Name, err)
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

	err := c.removePostgresRoleFinalizer(pgRole)
	if err != nil {
		glog.Errorf("PostgresRole %s/%s finalizer: removing finalizer %v", pgRole.Namespace, pgRole.Name, err)
	} else {
		glog.Infof("Removed finalizer for PostgresRole %s/%s", pgRole.Namespace, pgRole.Name)
	}

	// Delete key from finalizer info as processing is done
	c.finalizerInfo.Delete(id)
}

// Do:
//	- delete role in vault
//	- revoke lease of all the corresponding postgresRoleBinding
func (c *VaultController) finalizePostgresRole(dbRClient database.DatabaseRoleInterface, pgRole *api.PostgresRole) error {
	err := dbRClient.DeleteRole(pgRole.RoleName())
	if err != nil {
		return errors.Wrap(err, "failed to database role")
	}
	return nil
}

func (c *VaultController) removePostgresRoleFinalizer(pgRole *api.PostgresRole) error {
	p, err := c.extClient.EngineV1alpha1().PostgresRoles(pgRole.Namespace).Get(pgRole.Name, metav1.GetOptions{})
	if kerr.IsNotFound(err) {
		return nil
	} else if err != nil {
		return err
	}

	// remove finalizer
	_, _, err = patchutil.PatchPostgresRole(c.extClient.EngineV1alpha1(), p, func(role *api.PostgresRole) *api.PostgresRole {
		role.ObjectMeta = core_util.RemoveFinalizer(role.ObjectMeta, apis.Finalizer)
		return role
	})
	return err
}

func getPostgresRoleId(pgRole *api.PostgresRole) string {
	return fmt.Sprintf("%s/%s/%s", api.ResourcePostgresRole, pgRole.Namespace, pgRole.Name)
}
