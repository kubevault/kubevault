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
	"time"

	"kubevault.dev/operator/apis"
	policyapi "kubevault.dev/operator/apis/policy/v1alpha1"
	patchutil "kubevault.dev/operator/client/clientset/versioned/typed/policy/v1alpha1/util"
	pbinding "kubevault.dev/operator/pkg/vault/policybinding"

	"github.com/golang/glog"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	kmapi "kmodules.xyz/client-go/api/v1"
	core_util "kmodules.xyz/client-go/core/v1"
	"kmodules.xyz/client-go/tools/queue"
)

func (c *VaultController) initVaultPolicyBindingWatcher() {
	c.vplcyBindingInformer = c.extInformerFactory.Policy().V1alpha1().VaultPolicyBindings().Informer()
	c.vplcyBindingQueue = queue.New(policyapi.ResourceKindVaultPolicyBinding, c.MaxNumRequeues, c.NumThreads, c.runVaultPolicyBindingInjector)
	c.vplcyBindingInformer.AddEventHandler(queue.NewReconcilableHandler(c.vplcyBindingQueue.GetQueue()))
	c.vplcyBindingLister = c.extInformerFactory.Policy().V1alpha1().VaultPolicyBindings().Lister()
}

// runVaultPolicyBindingInjector gets the vault policy binding object indexed by the key from cache
// and initializes, reconciles or garbage collects the vault policy binding as needed.
func (c *VaultController) runVaultPolicyBindingInjector(key string) error {
	obj, exists, err := c.vplcyBindingInformer.GetIndexer().GetByKey(key)
	if err != nil {
		glog.Errorf("Fetching object with key %s from store failed with %v", key, err)
		return err
	}

	if !exists {
		glog.Warningf("VaultPolicyBinding %s does not exist anymore\n", key)
	} else {
		pb := obj.(*policyapi.VaultPolicyBinding).DeepCopy()
		glog.Infof("Sync/Add/Update for VaultPolicyBinding %s/%s\n", pb.Namespace, pb.Name)

		if pb.DeletionTimestamp != nil {
			if core_util.HasFinalizer(pb.ObjectMeta, apis.Finalizer) {
				// Finalize VaultPolicyBinding
				return c.runPolicyBindingFinalizer(pb)
			} else {
				glog.Infof("Finalizer not found for VaultPolicyBinding %s/%s", pb.Namespace, pb.Name)
			}
		} else {
			if !core_util.HasFinalizer(pb.ObjectMeta, apis.Finalizer) {
				// Add finalizer
				_, _, err := patchutil.PatchVaultPolicyBinding(context.TODO(), c.extClient.PolicyV1alpha1(), pb, func(in *policyapi.VaultPolicyBinding) *policyapi.VaultPolicyBinding {
					in.ObjectMeta = core_util.AddFinalizer(pb.ObjectMeta, apis.Finalizer)
					return in
				}, metav1.PatchOptions{})
				if err != nil {
					return errors.Wrapf(err, "failed to add VaultPolicyBinding finalizer for %s/%s", pb.Namespace, pb.Name)
				}
			}

			pbClient, err := pbinding.NewPolicyBindingClient(c.extClient, c.appCatalogClient, c.kubeClient, pb)
			if err != nil {
				return errors.Wrapf(err, "for VaultPolicyBinding %s/%s", pb.Namespace, pb.Name)
			}

			err = c.reconcilePolicyBinding(pb, pbClient)
			if err != nil {
				return errors.Wrapf(err, "for VaultPolicyBinding %s/%s", pb.Namespace, pb.Name)
			}
		}
	}
	return nil
}

// reconcilePolicyBinding reconciles the vault's policy binding
func (c *VaultController) reconcilePolicyBinding(pb *policyapi.VaultPolicyBinding, pbClient pbinding.PolicyBinding) error {
	// create or update policy
	err := pbClient.Ensure(pb)
	if err != nil {
		_, err2 := patchutil.UpdateVaultPolicyBindingStatus(
			context.TODO(),
			c.extClient.PolicyV1alpha1(),
			pb.ObjectMeta,
			func(status *policyapi.VaultPolicyBindingStatus) *policyapi.VaultPolicyBindingStatus {
				status.Phase = policyapi.PolicyBindingFailed
				status.Conditions = kmapi.SetCondition(status.Conditions, kmapi.Condition{
					Type:               kmapi.ConditionFailure,
					Status:             kmapi.ConditionTrue,
					Reason:             "FailedToEnsurePolicyBinding",
					Message:            err.Error(),
					LastTransitionTime: metav1.NewTime(time.Now()),
				})
				return status
			},
			metav1.UpdateOptions{},
		)
		return utilerrors.NewAggregate([]error{err2, err})
	}

	// update status
	_, err = patchutil.UpdateVaultPolicyBindingStatus(
		context.TODO(),
		c.extClient.PolicyV1alpha1(),
		pb.ObjectMeta,
		func(status *policyapi.VaultPolicyBindingStatus) *policyapi.VaultPolicyBindingStatus {
			status.ObservedGeneration = pb.Generation
			status.Phase = policyapi.PolicyBindingSuccess
			status.Conditions = kmapi.RemoveCondition(status.Conditions, kmapi.ConditionFailure)
			status.Conditions = kmapi.SetCondition(status.Conditions, kmapi.Condition{
				Type:    kmapi.ConditionAvailable,
				Status:  kmapi.ConditionTrue,
				Reason:  "Provisioned",
				Message: "policy binding is ready to use",
			})
			return status
		},
		metav1.UpdateOptions{},
	)
	if err != nil {
		return err
	}

	glog.Infof("Successfully processed VaultPolicyBinding: %s/%s", pb.Namespace, pb.Name)
	return nil
}

func (c *VaultController) runPolicyBindingFinalizer(pb *policyapi.VaultPolicyBinding) error {
	glog.Infof("Processing finalizer for VaultPolicyBinding: %s/%s", pb.Namespace, pb.Name)

	pbClient, err := pbinding.NewPolicyBindingClient(c.extClient, c.appCatalogClient, c.kubeClient, pb)
	// The error could be generated for:
	//   - invalid vaultRef in the spec
	// In this case, the operator should be able to delete the VaultPolicyBinding(ie. remove finalizer).
	// If no error occurred:
	//	- Delete the policy
	if err == nil {
		err = pbClient.Delete(pb)
		if err != nil {
			return errors.Wrap(err, "failed to delete the auth role created for policy binding")
		}
	} else {
		glog.Warningf("Skipping cleanup for VaultPolicyBinding: %s/%s with error: %v", pb.Namespace, pb.Name, err)
	}

	// Remove finalizer
	_, err = patchutil.TryPatchVaultPolicyBinding(context.TODO(), c.extClient.PolicyV1alpha1(), pb, func(in *policyapi.VaultPolicyBinding) *policyapi.VaultPolicyBinding {
		in.ObjectMeta = core_util.RemoveFinalizer(in.ObjectMeta, apis.Finalizer)
		return in
	}, metav1.PatchOptions{})
	if err != nil {
		return errors.Wrapf(err, "failed to remove finalizer for VaultPolicyBinding: %s/%s", pb.Namespace, pb.Name)
	}

	glog.Infof("Removed finalizer for VaultPolicyBinding %s/%s", pb.Namespace, pb.Name)
	return nil
}
