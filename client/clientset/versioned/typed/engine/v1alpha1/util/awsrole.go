package util

import (
	"encoding/json"
	"fmt"

	jsonpatch "github.com/evanphx/json-patch"
	"github.com/golang/glog"
	api "github.com/kubevault/operator/apis/engine/v1alpha1"
	cs "github.com/kubevault/operator/client/clientset/versioned/typed/engine/v1alpha1"
	"github.com/pkg/errors"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	kutil "kmodules.xyz/client-go"
)

func CreateOrPatchAWSRole(c cs.EngineV1alpha1Interface, meta metav1.ObjectMeta, transform func(alert *api.AWSRole) *api.AWSRole) (*api.AWSRole, kutil.VerbType, error) {
	cur, err := c.AWSRoles(meta.Namespace).Get(meta.Name, metav1.GetOptions{})
	if kerr.IsNotFound(err) {
		glog.V(3).Infof("Creating AWSRole %s/%s.", meta.Namespace, meta.Name)
		out, err := c.AWSRoles(meta.Namespace).Create(transform(&api.AWSRole{
			TypeMeta: metav1.TypeMeta{
				Kind:       api.ResourceKindAWSRole,
				APIVersion: api.SchemeGroupVersion.String(),
			},
			ObjectMeta: meta,
		}))
		return out, kutil.VerbCreated, err
	} else if err != nil {
		return nil, kutil.VerbUnchanged, err
	}
	return PatchAWSRole(c, cur, transform)
}

func PatchAWSRole(c cs.EngineV1alpha1Interface, cur *api.AWSRole, transform func(*api.AWSRole) *api.AWSRole) (*api.AWSRole, kutil.VerbType, error) {
	return PatchAWSRoleObject(c, cur, transform(cur.DeepCopy()))
}

func PatchAWSRoleObject(c cs.EngineV1alpha1Interface, cur, mod *api.AWSRole) (*api.AWSRole, kutil.VerbType, error) {
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
	glog.V(3).Infof("Patching AWSRole %s/%s with %s.", cur.Namespace, cur.Name, string(patch))
	out, err := c.AWSRoles(cur.Namespace).Patch(cur.Name, types.MergePatchType, patch)
	return out, kutil.VerbPatched, err
}

func TryUpdateAWSRole(c cs.EngineV1alpha1Interface, meta metav1.ObjectMeta, transform func(*api.AWSRole) *api.AWSRole) (result *api.AWSRole, err error) {
	attempt := 0
	err = wait.PollImmediate(kutil.RetryInterval, kutil.RetryTimeout, func() (bool, error) {
		attempt++
		cur, e2 := c.AWSRoles(meta.Namespace).Get(meta.Name, metav1.GetOptions{})
		if kerr.IsNotFound(e2) {
			return false, e2
		} else if e2 == nil {
			result, e2 = c.AWSRoles(cur.Namespace).Update(transform(cur.DeepCopy()))
			return e2 == nil, nil
		}
		glog.Errorf("Attempt %d failed to update AWSRole %s/%s due to %v.", attempt, cur.Namespace, cur.Name, e2)
		return false, nil
	})

	if err != nil {
		err = errors.Errorf("failed to update AWSRole %s/%s after %d attempts due to %v", meta.Namespace, meta.Name, attempt, err)
	}
	return
}

func UpdateAWSRoleStatus(
	c cs.EngineV1alpha1Interface,
	in *api.AWSRole,
	transform func(*api.AWSRoleStatus) *api.AWSRoleStatus,
	useSubresource ...bool,
) (result *api.AWSRole, err error) {
	if len(useSubresource) > 1 {
		return nil, errors.Errorf("invalid value passed for useSubresource: %v", useSubresource)
	}

	apply := func(x *api.AWSRole) *api.AWSRole {
		return &api.AWSRole{
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
			result, e2 = c.AWSRoles(in.Namespace).UpdateStatus(apply(cur))
			if kerr.IsConflict(e2) {
				latest, e3 := c.AWSRoles(in.Namespace).Get(in.Name, metav1.GetOptions{})
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
			err = fmt.Errorf("failed to update status of AWSRole %s/%s after %d attempts due to %v", in.Namespace, in.Name, attempt, err)
		}
		return
	}

	result, _, err = PatchAWSRoleObject(c, in, apply(in))
	return
}
