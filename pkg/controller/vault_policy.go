package controller

import (
	"time"

	"github.com/appscode/go/encoding/json/types"
	"github.com/golang/glog"
	"github.com/kubevault/operator/apis"
	policyapi "github.com/kubevault/operator/apis/policy/v1alpha1"
	patchutil "github.com/kubevault/operator/client/clientset/versioned/typed/policy/v1alpha1/util"
	"github.com/kubevault/operator/pkg/vault/policy"
	"github.com/pkg/errors"
	core "k8s.io/api/core/v1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	core_util "kmodules.xyz/client-go/core/v1"
	meta_util "kmodules.xyz/client-go/meta"
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
	c.vplcyInformer.AddEventHandler(queue.NewObservableHandler(c.vplcyQueue.GetQueue(), apis.EnableStatusSubresource))
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
				_, _, err := patchutil.PatchVaultPolicy(c.extClient.PolicyV1alpha1(), vPolicy, func(vp *policyapi.VaultPolicy) *policyapi.VaultPolicy {
					vp.ObjectMeta = core_util.AddFinalizer(vPolicy.ObjectMeta, VaultPolicyFinalizer)
					return vp
				})
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
	status := vPolicy.Status

	// create or update policy
	// its safe to call multiple times
	err := pClient.EnsurePolicy(vPolicy.PolicyName(), vPolicy.Spec.Policy)
	if err != nil {
		status.Status = policyapi.PolicyFailed
		status.Conditions = []policyapi.PolicyCondition{
			{
				Type:    policyapi.PolicyConditionFailure,
				Status:  core.ConditionTrue,
				Reason:  "FailedToPutPolicy",
				Message: err.Error(),
			},
		}

		err2 := c.updatePolicyStatus(&status, vPolicy)
		if err2 != nil {
			return errors.Wrap(err2, "failed to update VaultPolicy status")
		}
		return err
	}

	// update status
	status.ObservedGeneration = types.NewIntHash(vPolicy.Generation, meta_util.GenerationHash(vPolicy))
	status.Conditions = []policyapi.PolicyCondition{}
	status.Status = policyapi.PolicySuccess
	err2 := c.updatePolicyStatus(&status, vPolicy)
	if err2 != nil {
		return errors.Wrap(err2, "failed to update VaultPolicy status")
	}
	return nil
}

// updatePolicyStatus updates policy status
func (c *VaultController) updatePolicyStatus(status *policyapi.VaultPolicyStatus, vPolicy *policyapi.VaultPolicy) error {
	_, err := patchutil.UpdateVaultPolicyStatus(c.extClient.PolicyV1alpha1(), vPolicy, func(s *policyapi.VaultPolicyStatus) *policyapi.VaultPolicyStatus {
		s = status
		return s
	}, apis.EnableStatusSubresource)
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
	_, err := patchutil.TryPatchVaultPolicy(c.extClient.PolicyV1alpha1(), vPolicy, func(in *policyapi.VaultPolicy) *policyapi.VaultPolicy {
		in.ObjectMeta = core_util.RemoveFinalizer(in.ObjectMeta, VaultPolicyFinalizer)
		return in
	})
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
	out, err := c.extClient.PolicyV1alpha1().VaultPolicies(vPolicy.Namespace).Get(vPolicy.Name, metav1.GetOptions{})
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
