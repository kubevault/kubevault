package controller

import (
	"time"

	policyapi "kubevault.dev/operator/apis/policy/v1alpha1"
	patchutil "kubevault.dev/operator/client/clientset/versioned/typed/policy/v1alpha1/util"
	pbinding "kubevault.dev/operator/pkg/vault/policybinding"

	"github.com/golang/glog"
	"github.com/pkg/errors"
	core "k8s.io/api/core/v1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	core_util "kmodules.xyz/client-go/core/v1"
	"kmodules.xyz/client-go/tools/queue"
)

const (
	VaultPolicyBindingFinalizer = "policybinding.kubevault.com"
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
		vPBind := obj.(*policyapi.VaultPolicyBinding).DeepCopy()
		glog.Infof("Sync/Add/Update for VaultPolicyBinding %s/%s\n", vPBind.Namespace, vPBind.Name)

		if vPBind.DeletionTimestamp != nil {
			if core_util.HasFinalizer(vPBind.ObjectMeta, VaultPolicyBindingFinalizer) {
				// Finalize VaultPolicyBinding
				go c.runPolicyBindingFinalizer(vPBind, timeoutForFinalizer, timeIntervalForFinalizer)
			} else {
				glog.Infof("Finalizer not found for VaultPolicyBinding %s/%s", vPBind.Namespace, vPBind.Name)
			}
		} else {
			if !core_util.HasFinalizer(vPBind.ObjectMeta, VaultPolicyBindingFinalizer) {
				// Add finalizer
				_, _, err := patchutil.PatchVaultPolicyBinding(c.extClient.PolicyV1alpha1(), vPBind, func(vp *policyapi.VaultPolicyBinding) *policyapi.VaultPolicyBinding {
					vp.ObjectMeta = core_util.AddFinalizer(vPBind.ObjectMeta, VaultPolicyBindingFinalizer)
					return vp
				})
				if err != nil {
					return errors.Wrapf(err, "failed to set VaultPolicyBinding finalizer for %s/%s", vPBind.Namespace, vPBind.Name)
				}
			}

			pBClient, err := pbinding.NewPolicyBindingClient(c.extClient, c.appCatalogClient, c.kubeClient, vPBind)
			if err != nil {
				return errors.Wrapf(err, "for VaultPolicyBinding %s/%s", vPBind.Namespace, vPBind.Name)
			}

			err = c.reconcilePolicyBinding(vPBind, pBClient)
			if err != nil {
				return errors.Wrapf(err, "for VaultPolicyBinding %s/%s", vPBind.Namespace, vPBind.Name)
			}
		}
	}
	return nil
}

// reconcilePolicyBinding reconciles the vault's policy binding
func (c *VaultController) reconcilePolicyBinding(vPBind *policyapi.VaultPolicyBinding, pBClient pbinding.PolicyBinding) error {
	status := vPBind.Status

	// create or update policy
	// it's safe to call multiple times
	err := pBClient.Ensure(vPBind.PolicyBindingName())
	if err != nil {
		status.Phase = policyapi.PolicyBindingFailed
		status.Conditions = []policyapi.PolicyBindingCondition{
			{
				Type:    policyapi.PolicyBindingConditionFailure,
				Status:  core.ConditionTrue,
				Reason:  "FailedToEnsurePolicyBinding",
				Message: err.Error(),
			},
		}

		err2 := c.updatePolicyBindingStatus(&status, vPBind)
		if err2 != nil {
			return errors.Wrap(err2, "failed to update VaultPolicyBinding status")
		}
		return err
	}

	// update status
	status.ObservedGeneration = vPBind.Generation
	status.Conditions = []policyapi.PolicyBindingCondition{}
	status.Phase = policyapi.PolicyBindingSuccess
	err2 := c.updatePolicyBindingStatus(&status, vPBind)
	if err2 != nil {
		return errors.Wrap(err2, "failed to update VaultPolicyBinding status")
	}
	return nil
}

// updatePolicyBindingStatus updates policy binding status
func (c *VaultController) updatePolicyBindingStatus(status *policyapi.VaultPolicyBindingStatus, vPBind *policyapi.VaultPolicyBinding) error {
	_, err := patchutil.UpdateVaultPolicyBindingStatus(c.extClient.PolicyV1alpha1(), vPBind, func(s *policyapi.VaultPolicyBindingStatus) *policyapi.VaultPolicyBindingStatus {
		return status
	})
	return err
}

// runPolicyBindingFinalizer wil periodically run the finalizePolicyBinding until finalizePolicyBinding func produces no error or timeout occurs.
// After that it will remove the finalizer string from the objectMeta of VaultPolicyBinding
func (c *VaultController) runPolicyBindingFinalizer(vPBind *policyapi.VaultPolicyBinding, timeout time.Duration, interval time.Duration) {
	if vPBind == nil {
		glog.Infoln("VaultPolicyBinding is nil")
		return
	}

	key := vPBind.GetKey()
	if c.finalizerInfo.IsAlreadyProcessing(key) {
		// already processing it
		return
	}

	glog.Infof("Processing finalizer for VaultPolicyBinding %s/%s", vPBind.Namespace, vPBind.Name)
	// Add key to finalizerInfo, it will prevent other go routine to processing for this VaultPolicyBinding
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

		// finalize policy binding
		if err := c.finalizePolicyBinding(vPBind); err == nil {
			glog.Infof("For VaultPolicyBinding %s/%s: successfully removed policy from vault", vPBind.Namespace, vPBind.Name)
			break
		} else {
			glog.Infof("For VaultPolicyBinding %s/%s: %v", vPBind.Namespace, vPBind.Name, err)
		}

		select {
		case <-stopCh:
			timeOutOccured = true
		case <-time.After(interval):
		}
	}

	// Remove finalizer
	_, err := patchutil.TryPatchVaultPolicyBinding(c.extClient.PolicyV1alpha1(), vPBind, func(in *policyapi.VaultPolicyBinding) *policyapi.VaultPolicyBinding {
		in.ObjectMeta = core_util.RemoveFinalizer(in.ObjectMeta, VaultPolicyBindingFinalizer)
		return in
	})
	if err != nil {
		glog.Errorf("For VaultPolicyBinding %s/%s: %v", vPBind.Namespace, vPBind.Name, err)
	} else {
		glog.Infof("For VaultPolicyBinding %s/%s: removed finalizer '%s'", vPBind.Namespace, vPBind.Name, VaultPolicyBindingFinalizer)
	}
	// Delete key from finalizer info as processing is done
	c.finalizerInfo.Delete(key)
	glog.Infof("Removed finalizer for VaultPolicyBinding %s/%s", vPBind.Namespace, vPBind.Name)
}

// finalizePolicyBinding will delete the policy in vault
func (c *VaultController) finalizePolicyBinding(vPBind *policyapi.VaultPolicyBinding) error {
	out, err := c.extClient.PolicyV1alpha1().VaultPolicyBindings(vPBind.Namespace).Get(vPBind.Name, metav1.GetOptions{})
	if kerr.IsNotFound(err) {
		return nil
	} else if err != nil {
		return err
	}

	pBClient, err := pbinding.NewPolicyBindingClient(c.extClient, c.appCatalogClient, c.kubeClient, out)
	if err != nil {
		return err
	}
	return pBClient.Delete(vPBind.PolicyBindingName())
}
