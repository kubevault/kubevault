package util

import (
	"encoding/json"
	"fmt"

	jsonpatch "github.com/evanphx/json-patch"
	"github.com/golang/glog"
	"github.com/pkg/errors"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	kutil "kmodules.xyz/client-go"
	api "kubevault.dev/operator/apis/policy/v1alpha1"
	cs "kubevault.dev/operator/client/clientset/versioned/typed/policy/v1alpha1"
)

func CreateOrPatchVaultPolicy(c cs.PolicyV1alpha1Interface, meta metav1.ObjectMeta, transform func(alert *api.VaultPolicy) *api.VaultPolicy) (*api.VaultPolicy, kutil.VerbType, error) {
	cur, err := c.VaultPolicies(meta.Namespace).Get(meta.Name, metav1.GetOptions{})
	if kerr.IsNotFound(err) {
		glog.V(3).Infof("Creating VaultPolicy %s/%s.", meta.Namespace, meta.Name)
		out, err := c.VaultPolicies(meta.Namespace).Create(transform(&api.VaultPolicy{
			TypeMeta: metav1.TypeMeta{
				Kind:       api.ResourceKindVaultPolicy,
				APIVersion: api.SchemeGroupVersion.String(),
			},
			ObjectMeta: meta,
		}))
		return out, kutil.VerbCreated, err
	} else if err != nil {
		return nil, kutil.VerbUnchanged, err
	}
	return PatchVaultPolicy(c, cur, transform)
}

func PatchVaultPolicy(c cs.PolicyV1alpha1Interface, cur *api.VaultPolicy, transform func(*api.VaultPolicy) *api.VaultPolicy) (*api.VaultPolicy, kutil.VerbType, error) {
	return PatchVaultPolicyObject(c, cur, transform(cur.DeepCopy()))
}

func PatchVaultPolicyObject(c cs.PolicyV1alpha1Interface, cur, mod *api.VaultPolicy) (*api.VaultPolicy, kutil.VerbType, error) {
	curJson, err := json.Marshal(cur)
	if err != nil {
		return nil, kutil.VerbUnchanged, err
	}

	modJson, err := json.Marshal(mod)
	if err != nil {
		return nil, kutil.VerbUnchanged, err
	}

	patch, err := jsonpatch.CreateMergePatch(curJson, modJson)
	if err != nil {
		return nil, kutil.VerbUnchanged, err
	}
	if len(patch) == 0 || string(patch) == "{}" {
		return cur, kutil.VerbUnchanged, nil
	}
	glog.V(3).Infof("Patching VaultPolicy %s/%s with %s.", cur.Namespace, cur.Name, string(patch))
	out, err := c.VaultPolicies(cur.Namespace).Patch(cur.Name, types.MergePatchType, patch)
	return out, kutil.VerbPatched, err
}

func TryPatchVaultPolicy(c cs.PolicyV1alpha1Interface, cur *api.VaultPolicy, transform func(*api.VaultPolicy) *api.VaultPolicy) (*api.VaultPolicy, error) {
	var (
		out *api.VaultPolicy
		e2  error
	)
	attempt := 0
	err := wait.PollImmediate(kutil.RetryInterval, kutil.RetryTimeout, func() (bool, error) {
		attempt++
		cur, e2 = c.VaultPolicies(cur.Namespace).Get(cur.Name, metav1.GetOptions{})
		if kerr.IsNotFound(e2) {
			return false, e2
		} else if e2 == nil {
			out, _, e2 = PatchVaultPolicyObject(c, cur, transform(cur.DeepCopy()))
			return e2 == nil, nil
		}
		glog.Errorf("Attempt %d failed to patch VaultPolicy %s/%s due to %v.", attempt, cur.Namespace, cur.Name, e2)
		return false, nil
	})

	if err != nil {
		return nil, errors.Errorf("failed to patch VaultPolicy %s/%s after %d attempts due to %v", cur.Namespace, cur.Name, attempt, err)
	}
	return out, nil
}

func TryUpdateVaultPolicy(c cs.PolicyV1alpha1Interface, meta metav1.ObjectMeta, transform func(*api.VaultPolicy) *api.VaultPolicy) (result *api.VaultPolicy, err error) {
	attempt := 0
	err = wait.PollImmediate(kutil.RetryInterval, kutil.RetryTimeout, func() (bool, error) {
		attempt++
		cur, e2 := c.VaultPolicies(meta.Namespace).Get(meta.Name, metav1.GetOptions{})
		if kerr.IsNotFound(e2) {
			return false, e2
		} else if e2 == nil {
			result, e2 = c.VaultPolicies(cur.Namespace).Update(transform(cur.DeepCopy()))
			return e2 == nil, nil
		}
		glog.Errorf("Attempt %d failed to update VaultPolicy %s/%s due to %v.", attempt, cur.Namespace, cur.Name, e2)
		return false, nil
	})

	if err != nil {
		err = errors.Errorf("failed to update VaultPolicy %s/%s after %d attempts due to %v", meta.Namespace, meta.Name, attempt, err)
	}
	return
}

func UpdateVaultPolicyStatus(
	c cs.PolicyV1alpha1Interface,
	in *api.VaultPolicy,
	transform func(*api.VaultPolicyStatus) *api.VaultPolicyStatus,
) (result *api.VaultPolicy, err error) {
	apply := func(x *api.VaultPolicy, copy bool) *api.VaultPolicy {
		out := &api.VaultPolicy{
			TypeMeta:   x.TypeMeta,
			ObjectMeta: x.ObjectMeta,
			Spec:       x.Spec,
		}
		if copy {
			out.Status = *transform(in.Status.DeepCopy())
		} else {
			out.Status = *transform(&in.Status)
		}
		return out
	}

	attempt := 0
	cur := in.DeepCopy()
	err = wait.PollImmediate(kutil.RetryInterval, kutil.RetryTimeout, func() (bool, error) {
		attempt++
		var e2 error
		result, e2 = c.VaultPolicies(in.Namespace).UpdateStatus(apply(cur, false))
		if kerr.IsConflict(e2) {
			latest, e3 := c.VaultPolicies(in.Namespace).Get(in.Name, metav1.GetOptions{})
			switch {
			case e3 == nil:
				cur = latest
				return false, nil
			case kutil.IsRequestRetryable(e3):
				return false, nil
			default:
				return false, e3
			}
		} else if err != nil && !kutil.IsRequestRetryable(e2) {
			return false, e2
		}
		return e2 == nil, nil
	})

	if err != nil {
		err = fmt.Errorf("failed to update status of VaultPolicy %s/%s after %d attempts due to %v", in.Namespace, in.Name, attempt, err)
	}
	return
}
