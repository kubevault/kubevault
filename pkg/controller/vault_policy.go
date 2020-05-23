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
	"encoding/json"
	"fmt"
	"time"

	policyapi "kubevault.dev/operator/apis/policy/v1alpha1"
	patchutil "kubevault.dev/operator/client/clientset/versioned/typed/policy/v1alpha1/util"
	"kubevault.dev/operator/pkg/vault/policy"

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
	VaultPolicyFinalizer     = "policy.kubevault.com"
	timeoutForFinalizer      = 1 * time.Minute
	timeIntervalForFinalizer = 5 * time.Second
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
		glog.Errorf("Fetching object with key %s from store failed with %v", key, err)
		return err
	}

	if !exists {
		glog.Warningf("VaultPolicy %s does not exist anymore\n", key)
	} else {
		vPolicy := obj.(*policyapi.VaultPolicy).DeepCopy()
		glog.Infof("Sync/Add/Update for VaultPolicy %s/%s\n", vPolicy.Namespace, vPolicy.Name)

		if vPolicy.DeletionTimestamp != nil {
			if core_util.HasFinalizer(vPolicy.ObjectMeta, VaultPolicyFinalizer) {
				// Finalize VaultPolicy
				go c.runPolicyFinalizer(vPolicy, timeoutForFinalizer, timeIntervalForFinalizer)
			} else {
				glog.Infof("Finalizer not found for VaultPolicy %s/%s", vPolicy.Namespace, vPolicy.Name)
			}
		} else {
			if !core_util.HasFinalizer(vPolicy.ObjectMeta, VaultPolicyFinalizer) {
				// Add finalizer
				_, _, err := patchutil.PatchVaultPolicy(context.TODO(), c.extClient.PolicyV1alpha1(), vPolicy, func(vp *policyapi.VaultPolicy) *policyapi.VaultPolicy {
					vp.ObjectMeta = core_util.AddFinalizer(vPolicy.ObjectMeta, VaultPolicyFinalizer)
					return vp
				}, metav1.PatchOptions{})
				if err != nil {
					return errors.Wrapf(err, "failed to set VaultPolicy finalizer for %s/%s", vPolicy.Namespace, vPolicy.Name)
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
					Type:    kmapi.ConditionFailure,
					Status:  kmapi.ConditionTrue,
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
			status.Conditions = kmapi.SetCondition(status.Conditions, kmapi.Condition{
				Type:    kmapi.ConditionAvailable,
				Status:  kmapi.ConditionTrue,
				Reason:  "Provisioned",
				Message: "policy is ready to use",
			})
			return status
		},
		metav1.UpdateOptions{},
	)
	return err
}

// runPolicyFinalizer wil periodically run the finalizePolicy until finalizePolicy func produces no error or timeout occurs.
// After that it will remove the finalizer string from the objectMeta of VaultPolicy
func (c *VaultController) runPolicyFinalizer(vPolicy *policyapi.VaultPolicy, timeout time.Duration, interval time.Duration) {
	if vPolicy == nil {
		glog.Infoln("VaultPolicy in nil")
		return
	}

	key := vPolicy.GetKey()
	if c.finalizerInfo.IsAlreadyProcessing(key) {
		// already processing it
		return
	}

	glog.Infof("Processing finalizer for VaultPolicy %s/%s", vPolicy.Namespace, vPolicy.Name)
	// Add key to finalizerInfo, it will prevent other go routine to processing for this VaultPolicy
	c.finalizerInfo.Add(key)
	stopCh := time.After(timeout)
	timeOutOccured := false
	for {
		select {
		case <-stopCh:
			timeOutOccured = true
		default:
		}

		if timeOutOccured {
			break
		}

		// finalize policy
		if err := c.finalizePolicy(vPolicy); err == nil {
			glog.Infof("For VaultPolicy %s/%s: successfully removed policy from vault", vPolicy.Namespace, vPolicy.Name)
			break
		} else {
			glog.Infof("For VaultPolicy %s/%s: %v", vPolicy.Namespace, vPolicy.Name, err)
		}

		select {
		case <-stopCh:
			timeOutOccured = true
		case <-time.After(interval):
		}
	}

	// Remove finalizer
	_, err := patchutil.TryPatchVaultPolicy(context.TODO(), c.extClient.PolicyV1alpha1(), vPolicy, func(in *policyapi.VaultPolicy) *policyapi.VaultPolicy {
		in.ObjectMeta = core_util.RemoveFinalizer(in.ObjectMeta, VaultPolicyFinalizer)
		return in
	}, metav1.PatchOptions{})
	if err != nil {
		glog.Errorf("For VaultPolicy %s/%s: %v", vPolicy.Namespace, vPolicy.Name, err)
	} else {
		glog.Infof("For VaultPolicy %s/%s: removed finalizer '%s'", vPolicy.Namespace, vPolicy.Name, VaultPolicyFinalizer)
	}
	// Delete key from finalizer info as processing is done
	c.finalizerInfo.Delete(key)
	glog.Infof("Removed finalizer for VaultPolicy %s/%s", vPolicy.Namespace, vPolicy.Name)
}

// finalizePolicy will delete the policy in vault
func (c *VaultController) finalizePolicy(vPolicy *policyapi.VaultPolicy) error {
	out, err := c.extClient.PolicyV1alpha1().VaultPolicies(vPolicy.Namespace).Get(context.TODO(), vPolicy.Name, metav1.GetOptions{})
	if kerr.IsNotFound(err) {
		return nil
	} else if err != nil {
		return err
	}

	pClient, err := policy.NewPolicyClientForVault(c.kubeClient, c.appCatalogClient, out)
	if err != nil {
		return err
	}
	return pClient.DeletePolicy(vPolicy.PolicyName())
}
