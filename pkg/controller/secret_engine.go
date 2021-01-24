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

	"kubevault.dev/operator/apis"
	api "kubevault.dev/operator/apis/engine/v1alpha1"
	patchutil "kubevault.dev/operator/client/clientset/versioned/typed/engine/v1alpha1/util"
	"kubevault.dev/operator/pkg/vault/engine"

	"github.com/golang/glog"
	"github.com/pkg/errors"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	kmapi "kmodules.xyz/client-go/api/v1"
	core_util "kmodules.xyz/client-go/core/v1"
	"kmodules.xyz/client-go/tools/queue"
)

const (
	SecretEnginePhaseSuccess    api.SecretEnginePhase = "Success"
	SecretEnginePhaseProcessing api.SecretEnginePhase = "Processing"
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
		se := obj.(*api.SecretEngine).DeepCopy()

		glog.Infof("Sync/Add/Update for SecretEngine %s/%s", se.Namespace, se.Name)

		if se.DeletionTimestamp != nil {
			if core_util.HasFinalizer(se.ObjectMeta, apis.Finalizer) {
				return c.runSecretEngineFinalizer(se)

			}
		} else {
			if !core_util.HasFinalizer(se.ObjectMeta, apis.Finalizer) {
				// Add finalizer
				_, _, err := patchutil.PatchSecretEngine(context.TODO(), c.extClient.EngineV1alpha1(), se, func(in *api.SecretEngine) *api.SecretEngine {
					in.ObjectMeta = core_util.AddFinalizer(in.ObjectMeta, apis.Finalizer)
					return in
				}, metav1.PatchOptions{})
				if err != nil {
					return errors.Wrapf(err, "failed to add finalizer for secretEngine: %s/%s", se.Namespace, se.Name)
				}
			}

			// Conditions are empty, when the secretEngine obj is enqueued for first time.
			// Set status.phase to "Processing".
			if se.Status.Conditions == nil {
				newSE, err := patchutil.UpdateSecretEngineStatus(
					context.TODO(),
					c.extClient.EngineV1alpha1(),
					se.ObjectMeta,
					func(status *api.SecretEngineStatus) *api.SecretEngineStatus {
						status.Phase = SecretEnginePhaseProcessing
						return status
					},
					metav1.UpdateOptions{},
				)
				if err != nil {
					return errors.Wrapf(err, "failed to update status for SecretEngine: %s/%s", se.Namespace, se.Name)
				}
				se = newSE
			}

			seClient, err := engine.NewSecretEngine(c.kubeClient, c.appCatalogClient, se)
			if err != nil {
				return err
			}
			err = c.reconcileSecretEngine(seClient, se)
			if err != nil {
				return errors.Wrapf(err, "for SecretEngine %s/%s:", se.Namespace, se.Name)
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
func (c *VaultController) reconcileSecretEngine(seClient engine.EngineInterface, se *api.SecretEngine) error {
	// Create required policies for secret engine
	err := seClient.CreatePolicy()
	if err != nil {
		_, err2 := patchutil.UpdateSecretEngineStatus(
			context.TODO(),
			c.extClient.EngineV1alpha1(),
			se.ObjectMeta,
			func(status *api.SecretEngineStatus) *api.SecretEngineStatus {
				status.Conditions = kmapi.SetCondition(status.Conditions, kmapi.Condition{
					Type:    kmapi.ConditionFailed,
					Status:  core.ConditionTrue,
					Reason:  "FailedToCreateSecretEnginePolicy",
					Message: err.Error(),
				})
				return status
			},
			metav1.UpdateOptions{},
		)
		return utilerrors.NewAggregate([]error{err2, errors.Wrap(err, "failed to create secret engine policy")})
	}

	// Update the policy field of the auth method
	err = seClient.UpdateAuthRole()
	if err != nil {
		_, err2 := patchutil.UpdateSecretEngineStatus(
			context.TODO(),
			c.extClient.EngineV1alpha1(),
			se.ObjectMeta,
			func(status *api.SecretEngineStatus) *api.SecretEngineStatus {
				status.Conditions = kmapi.SetCondition(status.Conditions, kmapi.Condition{
					Type:    kmapi.ConditionFailed,
					Status:  core.ConditionTrue,
					Reason:  "FailedToUpdateAuthRole",
					Message: err.Error(),
				})
				return status
			},
			metav1.UpdateOptions{},
		)
		return utilerrors.NewAggregate([]error{err2, errors.Wrap(err, "failed to update auth role")})
	}

	// enable the secret engine if it is not already enabled
	err = seClient.EnableSecretEngine()
	if err != nil {
		_, err2 := patchutil.UpdateSecretEngineStatus(
			context.TODO(),
			c.extClient.EngineV1alpha1(),
			se.ObjectMeta,
			func(status *api.SecretEngineStatus) *api.SecretEngineStatus {
				status.Conditions = kmapi.SetCondition(status.Conditions, kmapi.Condition{
					Type:    kmapi.ConditionFailed,
					Status:  core.ConditionTrue,
					Reason:  "FailedToEnableSecretEngine",
					Message: err.Error(),
				})
				return status
			},
			metav1.UpdateOptions{},
		)
		return utilerrors.NewAggregate([]error{err2, errors.Wrap(err, "failed to enable secret engine")})
	}

	// Create secret engine config
	err = seClient.CreateConfig()
	if err != nil {
		_, err2 := patchutil.UpdateSecretEngineStatus(
			context.TODO(),
			c.extClient.EngineV1alpha1(),
			se.ObjectMeta,
			func(status *api.SecretEngineStatus) *api.SecretEngineStatus {
				status.Conditions = kmapi.SetCondition(status.Conditions, kmapi.Condition{
					Type:    kmapi.ConditionFailed,
					Status:  core.ConditionTrue,
					Reason:  "FailedToCreateSecretEngineConfig",
					Message: err.Error(),
				})
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
		se.ObjectMeta,
		func(status *api.SecretEngineStatus) *api.SecretEngineStatus {
			status.ObservedGeneration = se.Generation
			status.Phase = SecretEnginePhaseSuccess
			status.Conditions = kmapi.RemoveCondition(status.Conditions, kmapi.ConditionFailed)
			status.Conditions = kmapi.SetCondition(status.Conditions, kmapi.Condition{
				Type:    kmapi.ConditionAvailable,
				Status:  core.ConditionTrue,
				Reason:  "Provisioned",
				Message: "secret engine is ready to use",
			})
			return status
		},
		metav1.UpdateOptions{},
	)
	if err != nil {
		return err
	}

	glog.Infof("Successfully processed SecretEngine: %s/%s", se.Namespace, se.Name)
	return nil
}

func (c *VaultController) runSecretEngineFinalizer(se *api.SecretEngine) error {
	glog.Infof("Processing finalizer for SecretEngine %s/%s", se.Namespace, se.Name)

	seClient, err := engine.NewSecretEngine(c.kubeClient, c.appCatalogClient, se)
	// The error could be generated for:
	//   - invalid vaultRef in the spec
	// In this case, the operator should be able to delete the SecretEngine(ie. remove finalizer).
	// If no error occurred:
	//	- Delete the policy created for this secret engine
	//	- remove the policy from policy controller role
	//	- disable secret engine
	if err == nil {
		err = seClient.DeletePolicyAndUpdateRole()
		if err != nil {
			return errors.Wrap(err, "failed to delete policy or update policy controller role")
		}

		err = seClient.DisableSecretEngine()
		if err != nil {
			return errors.Wrap(err, "failed to disable secret engine")
		}
	} else {
		glog.Warningf("Skipping cleanup for SecretEngine: %s/%s with error: %v", se.Namespace, se.Name, err)
	}

	// remove finalizer
	_, _, err = patchutil.PatchSecretEngine(context.TODO(), c.extClient.EngineV1alpha1(), se, func(in *api.SecretEngine) *api.SecretEngine {
		in.ObjectMeta = core_util.RemoveFinalizer(in.ObjectMeta, apis.Finalizer)
		return in
	}, metav1.PatchOptions{})
	if err != nil {
		return errors.Wrapf(err, "failed to remove finalizer for SecretEngine: %s/%s", se.Namespace, se.Name)
	}

	glog.Infof("Removed finalizer for SecretEngine: %s/%s", se.Namespace, se.Name)
	return nil
}
