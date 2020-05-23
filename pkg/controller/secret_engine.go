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
	"kubevault.dev/operator/pkg/vault/engine"

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
	SecretEnginePhaseSuccess api.SecretEnginePhase = "Success"
	SecretEngineFinalizer    string                = "secretengine.engine.kubevault.com"
)

func (c *VaultController) initSecretEngineWatcher() {
	c.secretEngineInformer = c.extInformerFactory.Engine().V1alpha1().SecretEngines().Informer()
	c.secretEngineQueue = queue.New(api.ResourceKindSecretEngine, c.MaxNumRequeues, c.NumThreads, c.runSecretEngineInjector)
	c.secretEngineInformer.AddEventHandler(queue.NewReconcilableHandler(c.secretEngineQueue.GetQueue()))
	c.secretEngineLister = c.extInformerFactory.Engine().V1alpha1().SecretEngines().Lister()
}

func (c *VaultController) runSecretEngineInjector(key string) error {
	obj, exist, err := c.secretEngineInformer.GetIndexer().GetByKey(key)
	if err != nil {
		glog.Errorf("Fetching object with key %s from store failed with %v", key, err)
		return err
	}

	if !exist {
		glog.Warningf("SecretEngine %s does not exist anymore", key)

	} else {
		secretEngine := obj.(*api.SecretEngine).DeepCopy()

		glog.Infof("Sync/Add/Update for SecretEngine %s/%s", secretEngine.Namespace, secretEngine.Name)

		if secretEngine.DeletionTimestamp != nil {
			if core_util.HasFinalizer(secretEngine.ObjectMeta, SecretEngineFinalizer) {
				go c.runSecretEngineFinalizer(secretEngine, finalizerTimeout, finalizerInterval)
			}
		} else {
			if !core_util.HasFinalizer(secretEngine.ObjectMeta, SecretEngineFinalizer) {
				// Add finalizer
				_, _, err := patchutil.PatchSecretEngine(context.TODO(), c.extClient.EngineV1alpha1(), secretEngine, func(engine *api.SecretEngine) *api.SecretEngine {
					engine.ObjectMeta = core_util.AddFinalizer(engine.ObjectMeta, SecretEngineFinalizer)
					return engine
				}, metav1.PatchOptions{})
				if err != nil {
					return errors.Wrapf(err, "failed to set SecretEngine finalizer for %s/%s", secretEngine.Namespace, secretEngine.Name)
				}
			}

			seClient, err := engine.NewSecretEngine(c.kubeClient, c.appCatalogClient, secretEngine)
			if err != nil {
				return err
			}
			err = c.reconcileSecretEngine(seClient, secretEngine)
			if err != nil {
				return errors.Wrapf(err, "for SecretEngine %s/%s:", secretEngine.Namespace, secretEngine.Name)
			}
		}
	}
	return nil
}

//	For vault:
//	  - create policy and update auth role for s/a of VaultAppRef
//	  - enable the secrets engine if it is not already enabled
//	  - configure Vault secret engine
//    - create policy and policybinding for s/a of VaultAppRef
func (c *VaultController) reconcileSecretEngine(secretEngineClient engine.EngineInterface, secretEngine *api.SecretEngine) error {
	// Create required policies for secret engine
	err := secretEngineClient.CreatePolicy()
	if err != nil {
		_, err2 := patchutil.UpdateSecretEngineStatus(
			context.TODO(),
			c.extClient.EngineV1alpha1(),
			secretEngine.ObjectMeta,
			func(status *api.SecretEngineStatus) *api.SecretEngineStatus {
				status.Conditions = []kmapi.Condition{
					{
						Type:    kmapi.ConditionFailure,
						Status:  kmapi.ConditionTrue,
						Reason:  "FailedToCreateSecretEnginePolicy",
						Message: err.Error(),
					},
				}
				return status
			},
			metav1.UpdateOptions{},
		)
		return utilerrors.NewAggregate([]error{err2, errors.Wrap(err, "failed to create secret engine policy")})
	}

	// Update the policy field of the auth method
	err = secretEngineClient.UpdateAuthRole()
	if err != nil {
		_, err2 := patchutil.UpdateSecretEngineStatus(
			context.TODO(),
			c.extClient.EngineV1alpha1(),
			secretEngine.ObjectMeta,
			func(status *api.SecretEngineStatus) *api.SecretEngineStatus {
				status.Conditions = []kmapi.Condition{
					{
						Type:    kmapi.ConditionFailure,
						Status:  kmapi.ConditionTrue,
						Reason:  "FailedToUpdateAuthRole",
						Message: err.Error(),
					},
				}
				return status
			},
			metav1.UpdateOptions{},
		)
		return utilerrors.NewAggregate([]error{err2, errors.Wrap(err, "failed to update auth role")})
	}

	// enable the secret engine if it is not already enabled
	err = secretEngineClient.EnableSecretEngine()
	if err != nil {
		_, err2 := patchutil.UpdateSecretEngineStatus(
			context.TODO(),
			c.extClient.EngineV1alpha1(),
			secretEngine.ObjectMeta,
			func(status *api.SecretEngineStatus) *api.SecretEngineStatus {
				status.Conditions = []kmapi.Condition{
					{
						Type:    kmapi.ConditionFailure,
						Status:  kmapi.ConditionTrue,
						Reason:  "FailedToEnableSecretEngine",
						Message: err.Error(),
					},
				}
				return status
			},
			metav1.UpdateOptions{},
		)
		return utilerrors.NewAggregate([]error{err2, errors.Wrap(err, "failed to enable secret engine")})
	}

	// Create secret engine config
	err = secretEngineClient.CreateConfig()
	if err != nil {
		_, err2 := patchutil.UpdateSecretEngineStatus(
			context.TODO(),
			c.extClient.EngineV1alpha1(),
			secretEngine.ObjectMeta,
			func(status *api.SecretEngineStatus) *api.SecretEngineStatus {
				status.Conditions = []kmapi.Condition{
					{
						Type:    kmapi.ConditionFailure,
						Status:  kmapi.ConditionTrue,
						Reason:  "FailedToCreateSecretEngineConfig",
						Message: err.Error(),
					},
				}
				return status
			},
			metav1.UpdateOptions{},
		)
		return utilerrors.NewAggregate([]error{err2, errors.Wrap(err, "failed to create secret engine config")})
	}

	// update status
	_, err = patchutil.UpdateSecretEngineStatus(
		context.TODO(),
		c.extClient.EngineV1alpha1(),
		secretEngine.ObjectMeta,
		func(status *api.SecretEngineStatus) *api.SecretEngineStatus {
			status.ObservedGeneration = secretEngine.Generation
			status.Conditions = []kmapi.Condition{}
			status.Phase = SecretEnginePhaseSuccess
			return status
		},
		metav1.UpdateOptions{},
	)
	return err
}

func (c *VaultController) runSecretEngineFinalizer(secretEngine *api.SecretEngine, timeout time.Duration, interval time.Duration) {
	if secretEngine == nil {
		glog.Infoln("SecretEngine is nil")
		return
	}

	id := getSecretEngineId(secretEngine)
	if c.finalizerInfo.IsAlreadyProcessing(id) {
		// already processing
		return
	}

	glog.Infof("Processing finalizer for SecretEngine %s/%s", secretEngine.Namespace, secretEngine.Name)

	// Add key to finalizerInfo, it will prevent other go routine to processing for this SecretEngine
	c.finalizerInfo.Add(id)

	stopCh := time.After(timeout)
	finalizationDone := false
	timeOutOccured := false
	attempt := 0

	for {
		glog.Infof("SecretEngine %s/%s finalizer: attempt %d\n", secretEngine.Namespace, secretEngine.Name, attempt)

		select {
		case <-stopCh:
			timeOutOccured = true
		default:
		}

		if timeOutOccured {
			break
		}

		if !finalizationDone {
			secretEngineClient, err := engine.NewSecretEngine(c.kubeClient, c.appCatalogClient, secretEngine)
			if err != nil {
				glog.Errorf("SecretEngine %s/%s finalizer: %v", secretEngine.Namespace, secretEngine.Name, err)
			} else {
				err = c.finalizeSecretEngine(secretEngineClient)
				if err != nil {
					glog.Errorf("SecretEngine %s/%s finalizer: %v", secretEngine.Namespace, secretEngine.Name, err)
				} else {
					finalizationDone = true
				}
			}
		}

		if finalizationDone {
			err := c.removeSecretEngineFinalizer(secretEngine)
			if err != nil {
				glog.Errorf("SecretEngine %s/%s finalizer: removing finalizer %v", secretEngine.Namespace, secretEngine.Name, err)
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

	err := c.removeSecretEngineFinalizer(secretEngine)
	if err != nil {
		glog.Errorf("SecretEngine %s/%s finalizer: removing finalizer %v", secretEngine.Namespace, secretEngine.Name, err)
	} else {
		glog.Infof("Removed finalizer for SecretEngine %s/%s", secretEngine.Namespace, secretEngine.Name)
	}

	// Delete key from finalizer info as processing is done
	c.finalizerInfo.Delete(id)
}

// will do:
//	- Delete the policy created for this secret engine
//	- remove the policy from policy controller role
//	- disable secret engine
func (c *VaultController) finalizeSecretEngine(secretEngineClient *engine.SecretEngine) error {
	err := secretEngineClient.DeletePolicyAndUpdateRole()
	if err != nil {
		return errors.Wrap(err, "failed to delete policy or update policy controller role")
	}

	err = secretEngineClient.DisableSecretEngine()
	if err != nil {
		return errors.Wrap(err, "failed to disable secret engine")
	}
	return nil
}

func (c *VaultController) removeSecretEngineFinalizer(secretEngine *api.SecretEngine) error {
	m, err := c.extClient.EngineV1alpha1().SecretEngines(secretEngine.Namespace).Get(context.TODO(), secretEngine.Name, metav1.GetOptions{})
	if kerr.IsNotFound(err) {
		return nil
	} else if err != nil {
		return err
	}

	// remove finalizer
	_, _, err = patchutil.PatchSecretEngine(context.TODO(), c.extClient.EngineV1alpha1(), m, func(secretEngine *api.SecretEngine) *api.SecretEngine {
		secretEngine.ObjectMeta = core_util.RemoveFinalizer(secretEngine.ObjectMeta, SecretEngineFinalizer)
		return secretEngine
	}, metav1.PatchOptions{})
	return err
}

func getSecretEngineId(secretEngine *api.SecretEngine) string {
	return fmt.Sprintf("%s/%s/%s", api.ResourceSecretEngine, secretEngine.Namespace, secretEngine.Name)
}
