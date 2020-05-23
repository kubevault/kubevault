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
	"kubevault.dev/operator/pkg/vault/role/database"

	"github.com/golang/glog"
	"github.com/pkg/errors"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	kmapi "kmodules.xyz/client-go/api/v1"
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
		role := obj.(*api.PostgresRole).DeepCopy()

		glog.Infof("Sync/Add/Update for PostgresRole %s/%s", role.Namespace, role.Name)

		if role.DeletionTimestamp != nil {
			if core_util.HasFinalizer(role.ObjectMeta, apis.Finalizer) {
				go c.runPostgresRoleFinalizer(role, finalizerTimeout, finalizerInterval)
			}

		} else {
			if !core_util.HasFinalizer(role.ObjectMeta, apis.Finalizer) {
				// Add finalizer
				_, _, err := patchutil.PatchPostgresRole(context.TODO(), c.extClient.EngineV1alpha1(), role, func(role *api.PostgresRole) *api.PostgresRole {
					role.ObjectMeta = core_util.AddFinalizer(role.ObjectMeta, apis.Finalizer)
					return role
				}, metav1.PatchOptions{})
				if err != nil {
					return errors.Wrapf(err, "failed to set postgresRole finalizer for %s/%s", role.Namespace, role.Name)
				}
			}

			dbRClient, err := database.NewDatabaseRoleForPostgres(c.kubeClient, c.appCatalogClient, role)
			if err != nil {
				return err
			}

			err = c.reconcilePostgresRole(dbRClient, role)
			if err != nil {
				return errors.Wrapf(err, "for PostgresRole %s/%s:", role.Namespace, role.Name)
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
func (c *VaultController) reconcilePostgresRole(dbRClient database.DatabaseRoleInterface, role *api.PostgresRole) error {
	// create role
	err := dbRClient.CreateRole()
	if err != nil {
		_, err2 := patchutil.UpdatePostgresRoleStatus(
			context.TODO(),
			c.extClient.EngineV1alpha1(),
			role.ObjectMeta,
			func(status *api.PostgresRoleStatus) *api.PostgresRoleStatus {
				status.Conditions = []kmapi.Condition{
					{
						Type:    kmapi.ConditionAvailable,
						Status:  kmapi.ConditionFalse,
						Reason:  "FailedToCreateDatabaseRole",
						Message: err.Error(),
					},
				}
				return status
			},
			metav1.UpdateOptions{},
		)
		return utilerrors.NewAggregate([]error{err2, errors.Wrapf(err, "failed to create role for postgresRole %s/%s", role.Namespace, role.Name)})
	}

	_, err = patchutil.UpdatePostgresRoleStatus(
		context.TODO(),
		c.extClient.EngineV1alpha1(),
		role.ObjectMeta,
		func(status *api.PostgresRoleStatus) *api.PostgresRoleStatus {
			status.ObservedGeneration = role.Generation
			status.Conditions = []kmapi.Condition{}
			status.Phase = PostgresRolePhaseSuccess

			return status
		},
		metav1.UpdateOptions{},
	)
	return err
}

func (c *VaultController) runPostgresRoleFinalizer(role *api.PostgresRole, timeout time.Duration, interval time.Duration) {
	if role == nil {
		glog.Infoln("PostgresRole is nil")
		return
	}

	id := getPostgresRoleId(role)
	if c.finalizerInfo.IsAlreadyProcessing(id) {
		// already processing
		return
	}

	glog.Infof("Processing finalizer for PostgresRole %s/%s", role.Namespace, role.Name)
	// Add key to finalizerInfo, it will prevent other go routine to processing for this PostgresRole
	c.finalizerInfo.Add(id)

	stopCh := time.After(timeout)
	finalizationDone := false
	timeOutOccured := false
	attempt := 0

	for {
		glog.Infof("PostgresRole %s/%s finalizer: attempt %d\n", role.Namespace, role.Name, attempt)

		select {
		case <-stopCh:
			timeOutOccured = true
		default:
		}

		if timeOutOccured {
			break
		}

		if !finalizationDone {
			d, err := database.NewDatabaseRoleForPostgres(c.kubeClient, c.appCatalogClient, role)
			if err != nil {
				glog.Errorf("PostgresRole %s/%s finalizer: %v", role.Namespace, role.Name, err)
			} else {
				err = c.finalizePostgresRole(d, role)
				if err != nil {
					glog.Errorf("PostgresRole %s/%s finalizer: %v", role.Namespace, role.Name, err)
				} else {
					finalizationDone = true
				}
			}
		}

		if finalizationDone {
			err := c.removePostgresRoleFinalizer(role)
			if err != nil {
				glog.Errorf("PostgresRole %s/%s finalizer: removing finalizer %v", role.Namespace, role.Name, err)
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

	err := c.removePostgresRoleFinalizer(role)
	if err != nil {
		glog.Errorf("PostgresRole %s/%s finalizer: removing finalizer %v", role.Namespace, role.Name, err)
	} else {
		glog.Infof("Removed finalizer for PostgresRole %s/%s", role.Namespace, role.Name)
	}

	// Delete key from finalizer info as processing is done
	c.finalizerInfo.Delete(id)
}

// Do:
//	- delete role in vault
//	- revoke lease of all the corresponding postgresRoleBinding
func (c *VaultController) finalizePostgresRole(dbRClient database.DatabaseRoleInterface, role *api.PostgresRole) error {
	err := dbRClient.DeleteRole(role.RoleName())
	if err != nil {
		return errors.Wrap(err, "failed to database role")
	}
	return nil
}

func (c *VaultController) removePostgresRoleFinalizer(role *api.PostgresRole) error {
	p, err := c.extClient.EngineV1alpha1().PostgresRoles(role.Namespace).Get(context.TODO(), role.Name, metav1.GetOptions{})
	if kerr.IsNotFound(err) {
		return nil
	} else if err != nil {
		return err
	}

	// remove finalizer
	_, _, err = patchutil.PatchPostgresRole(context.TODO(), c.extClient.EngineV1alpha1(), p, func(role *api.PostgresRole) *api.PostgresRole {
		role.ObjectMeta = core_util.RemoveFinalizer(role.ObjectMeta, apis.Finalizer)
		return role
	}, metav1.PatchOptions{})
	return err
}

func getPostgresRoleId(role *api.PostgresRole) string {
	return fmt.Sprintf("%s/%s/%s", api.ResourcePostgresRole, role.Namespace, role.Name)
}
