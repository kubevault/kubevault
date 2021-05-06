/*
Copyright AppsCode Inc. and Contributors

Licensed under the AppsCode Community License 1.0.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    https://github.com/appscode/licenses/raw/1.0.0/AppsCode-Community-1.0.0.md

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"
	"encoding/json"
	"fmt"

	"kubevault.dev/apimachinery/apis"
	policyapi "kubevault.dev/apimachinery/apis/policy/v1alpha1"
	patchutil "kubevault.dev/apimachinery/client/clientset/versioned/typed/policy/v1alpha1/util"
	"kubevault.dev/operator/pkg/vault/policy"

	"github.com/pkg/errors"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/klog/v2"
	kmapi "kmodules.xyz/client-go/api/v1"
	core_util "kmodules.xyz/client-go/core/v1"
	"kmodules.xyz/client-go/tools/queue"
)

func (c *VaultController) initVaultPolicyWatcher() {
	c.vplcyInformer = c.extInformerFactory.Policy().V1alpha1().VaultPolicies().Informer()
	c.vplcyQueue = queue.New(policyapi.ResourceKindVaultPolicy, c.MaxNumRequeues, c.NumThreads, c.runVaultPolicyInjector)
	c.vplcyInformer.AddEventHandler(queue.NewReconcilableHandler(c.vplcyQueue.GetQueue()))
	c.vplcyLister = c.extInformerFactory.Policy().V1alpha1().VaultPolicies().Lister()
}

// runVaultPolicyInjector gets the vault policy object indexed by the key from cache
// and initializes, reconciles or garbage collects the vault policy as needed.
func (c *VaultController) runVaultPolicyInjector(key string) error {
	obj, exists, err := c.vplcyInformer.GetIndexer().GetByKey(key)
	if err != nil {
		klog.Errorf("Fetching object with key %s from store failed with %v", key, err)
		return err
	}

	if !exists {
		klog.Warningf("VaultPolicy %s does not exist anymore\n", key)
	} else {
		vPolicy := obj.(*policyapi.VaultPolicy).DeepCopy()
		klog.Infof("Sync/Add/Update for VaultPolicy %s/%s\n", vPolicy.Namespace, vPolicy.Name)

		if vPolicy.DeletionTimestamp != nil {
			if core_util.HasFinalizer(vPolicy.ObjectMeta, apis.Finalizer) {
				return c.runPolicyFinalizer(vPolicy)
			} else {
				klog.Infof("Finalizer not found for VaultPolicy %s/%s", vPolicy.Namespace, vPolicy.Name)
			}
		} else {
			if !core_util.HasFinalizer(vPolicy.ObjectMeta, apis.Finalizer) {
				// Add finalizer
				_, _, err := patchutil.PatchVaultPolicy(context.TODO(), c.extClient.PolicyV1alpha1(), vPolicy, func(in *policyapi.VaultPolicy) *policyapi.VaultPolicy {
					in.ObjectMeta = core_util.AddFinalizer(vPolicy.ObjectMeta, apis.Finalizer)
					return in
				}, metav1.PatchOptions{})
				if err != nil {
					return errors.Wrapf(err, "failed to add VaultPolicy finalizer for %s/%s", vPolicy.Namespace, vPolicy.Name)
				}
			}

			pClient, err := policy.NewPolicyClientForVault(c.kubeClient, c.appCatalogClient, vPolicy)
			if err != nil {
				return errors.Wrapf(err, "for VaultPolicy %s/%s", vPolicy.Namespace, vPolicy.Name)
			}

			err = c.reconcilePolicy(vPolicy, pClient)
			if err != nil {
				return errors.Wrapf(err, "for VaultPolicy %s/%s", vPolicy.Namespace, vPolicy.Name)
			}
		}
	}
	return nil
}

// reconcileVault reconciles the vault's policy
// it will create or update policy in vault
func (c *VaultController) reconcilePolicy(vPolicy *policyapi.VaultPolicy, pClient policy.Policy) error {
	// create or update policy
	// its safe to call multiple times

	doc := vPolicy.Spec.PolicyDocument
	if vPolicy.Spec.PolicyDocument == "" && vPolicy.Spec.Policy != nil {
		data, err := json.Marshal(vPolicy.Spec.Policy)
		if err != nil {
			return fmt.Errorf("failed to serialize VaultPolicy %s/%s. Reason: %v", vPolicy.Namespace, vPolicy.Name, err)
		}
		doc = string(data)
	}

	err := pClient.EnsurePolicy(vPolicy.PolicyName(), doc)
	if err != nil {
		_, err2 := patchutil.UpdateVaultPolicyStatus(
			context.TODO(),
			c.extClient.PolicyV1alpha1(),
			vPolicy.ObjectMeta,
			func(status *policyapi.VaultPolicyStatus) *policyapi.VaultPolicyStatus {
				status.Phase = policyapi.PolicyFailed
				status.Conditions = kmapi.SetCondition(status.Conditions, kmapi.Condition{
					Type:    kmapi.ConditionFailed,
					Status:  core.ConditionTrue,
					Reason:  "FailedToPutPolicy",
					Message: err.Error(),
				})
				return status
			},
			metav1.UpdateOptions{},
		)
		return utilerrors.NewAggregate([]error{err2, err})
	}

	// update status
	_, err = patchutil.UpdateVaultPolicyStatus(
		context.TODO(),
		c.extClient.PolicyV1alpha1(),
		vPolicy.ObjectMeta,
		func(status *policyapi.VaultPolicyStatus) *policyapi.VaultPolicyStatus {
			status.ObservedGeneration = vPolicy.Generation
			status.Phase = policyapi.PolicySuccess
			status.Conditions = kmapi.RemoveCondition(status.Conditions, kmapi.ConditionFailed)
			status.Conditions = kmapi.SetCondition(status.Conditions, kmapi.Condition{
				Type:    kmapi.ConditionAvailable,
				Status:  core.ConditionTrue,
				Reason:  "Provisioned",
				Message: "policy is ready to use",
			})
			return status
		},
		metav1.UpdateOptions{},
	)
	if err != nil {
		return err
	}

	klog.Infof("Successfully processed VaultPolicy: %s/%s", vPolicy.Namespace, vPolicy.Name)
	return nil
}

// runPolicyFinalizer wil periodically run the finalizePolicy until finalizePolicy func produces no error
// After that it will remove the finalizer string from the objectMeta of VaultPolicy
func (c *VaultController) runPolicyFinalizer(vPolicy *policyapi.VaultPolicy) error {
	klog.Infof("Processing finalizer for VaultPolicy %s/%s", vPolicy.Namespace, vPolicy.Name)

	pClient, err := policy.NewPolicyClientForVault(c.kubeClient, c.appCatalogClient, vPolicy)
	// The error could be generated for:
	//   - invalid vaultRef in the spec
	// In this case, the operator should be able to delete the VaultPolicy(ie. remove finalizer).
	// If no error occurred:
	//	- Delete the policy
	if err == nil {
		err = pClient.DeletePolicy(vPolicy.PolicyName())
		if err != nil {
			return errors.Wrap(err, "failed to delete vault policy")
		}
	} else {
		klog.Warningf("Skipping cleanup for VaultPolicy: %s/%s with error: %v", vPolicy.Namespace, vPolicy.Name, err)
	}

	// Remove finalizer
	_, err = patchutil.TryPatchVaultPolicy(context.TODO(), c.extClient.PolicyV1alpha1(), vPolicy, func(in *policyapi.VaultPolicy) *policyapi.VaultPolicy {
		in.ObjectMeta = core_util.RemoveFinalizer(in.ObjectMeta, apis.Finalizer)
		return in
	}, metav1.PatchOptions{})
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("failed to remove finalizer for VaultPolicy: %s/%s", vPolicy.Namespace, vPolicy.Name))
	}

	klog.Infof("Removed finalizer for VaultPolicy: %s/%s", vPolicy.Namespace, vPolicy.Name)
	return nil
}
