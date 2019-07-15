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
	api "kubevault.dev/operator/apis/engine/v1alpha1"
	cs "kubevault.dev/operator/client/clientset/versioned/typed/engine/v1alpha1"
)

func CreateOrPatchAzureRole(c cs.EngineV1alpha1Interface, meta metav1.ObjectMeta, transform func(alert *api.AzureRole) *api.AzureRole) (*api.AzureRole, kutil.VerbType, error) {
	cur, err := c.AzureRoles(meta.Namespace).Get(meta.Name, metav1.GetOptions{})
	if kerr.IsNotFound(err) {
		glog.V(3).Infof("Creating AzureRole %s/%s.", meta.Namespace, meta.Name)
		out, err := c.AzureRoles(meta.Namespace).Create(transform(&api.AzureRole{
			TypeMeta: metav1.TypeMeta{
				Kind:       api.ResourceKindAzureRole,
				APIVersion: api.SchemeGroupVersion.String(),
			},
			ObjectMeta: meta,
		}))
		return out, kutil.VerbCreated, err
	} else if err != nil {
		return nil, kutil.VerbUnchanged, err
	}
	return PatchAzureRole(c, cur, transform)
}

func PatchAzureRole(c cs.EngineV1alpha1Interface, cur *api.AzureRole, transform func(*api.AzureRole) *api.AzureRole) (*api.AzureRole, kutil.VerbType, error) {
	return PatchAzureRoleObject(c, cur, transform(cur.DeepCopy()))
}

func PatchAzureRoleObject(c cs.EngineV1alpha1Interface, cur, mod *api.AzureRole) (*api.AzureRole, kutil.VerbType, error) {
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
	glog.V(3).Infof("Patching AzureRole %s/%s with %s.", cur.Namespace, cur.Name, string(patch))
	out, err := c.AzureRoles(cur.Namespace).Patch(cur.Name, types.MergePatchType, patch)
	return out, kutil.VerbPatched, err
}

func TryUpdateAzureRole(c cs.EngineV1alpha1Interface, meta metav1.ObjectMeta, transform func(*api.AzureRole) *api.AzureRole) (result *api.AzureRole, err error) {
	attempt := 0
	err = wait.PollImmediate(kutil.RetryInterval, kutil.RetryTimeout, func() (bool, error) {
		attempt++
		cur, e2 := c.AzureRoles(meta.Namespace).Get(meta.Name, metav1.GetOptions{})
		if kerr.IsNotFound(e2) {
			return false, e2
		} else if e2 == nil {
			result, e2 = c.AzureRoles(cur.Namespace).Update(transform(cur.DeepCopy()))
			return e2 == nil, nil
		}
		glog.Errorf("Attempt %d failed to update AzureRole %s/%s due to %v.", attempt, cur.Namespace, cur.Name, e2)
		return false, nil
	})

	if err != nil {
		err = errors.Errorf("failed to update AzureRole %s/%s after %d attempts due to %v", meta.Namespace, meta.Name, attempt, err)
	}
	return
}

func UpdateAzureRoleStatus(
	c cs.EngineV1alpha1Interface,
	in *api.AzureRole,
	transform func(*api.AzureRoleStatus) *api.AzureRoleStatus,
	useSubresource ...bool,
) (result *api.AzureRole, err error) {
	if len(useSubresource) > 1 {
		return nil, errors.Errorf("invalid value passed for useSubresource: %v", useSubresource)
	}

	apply := func(x *api.AzureRole) *api.AzureRole {
		return &api.AzureRole{
			TypeMeta:   x.TypeMeta,
			ObjectMeta: x.ObjectMeta,
			Spec:       x.Spec,
			Status:     *transform(in.Status.DeepCopy()),
		}
	}

	if len(useSubresource) == 1 && useSubresource[0] {
		attempt := 0
		cur := in.DeepCopy()
		err = wait.PollImmediate(kutil.RetryInterval, kutil.RetryTimeout, func() (bool, error) {
			attempt++
			var e2 error
			result, e2 = c.AzureRoles(in.Namespace).UpdateStatus(apply(cur))
			if kerr.IsConflict(e2) {
				latest, e3 := c.AzureRoles(in.Namespace).Get(in.Name, metav1.GetOptions{})
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
			err = fmt.Errorf("failed to update status of AzureRole %s/%s after %d attempts due to %v", in.Namespace, in.Name, attempt, err)
		}
		return
	}

	result, _, err = PatchAzureRoleObject(c, in, apply(in))
	return
}
