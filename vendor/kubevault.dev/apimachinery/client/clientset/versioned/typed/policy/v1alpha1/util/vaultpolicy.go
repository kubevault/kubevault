/*
Copyright AppsCode Inc. and Contributors

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

package util

import (
	"context"
	"encoding/json"
	"fmt"

	api "kubevault.dev/apimachinery/apis/policy/v1alpha1"
	cs "kubevault.dev/apimachinery/client/clientset/versioned/typed/policy/v1alpha1"

	jsonpatch "github.com/evanphx/json-patch"
	"github.com/golang/glog"
	"github.com/pkg/errors"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	kutil "kmodules.xyz/client-go"
)

func CreateOrPatchVaultPolicy(
	ctx context.Context,
	c cs.PolicyV1alpha1Interface,
	meta metav1.ObjectMeta,
	transform func(alert *api.VaultPolicy) *api.VaultPolicy,
	opts metav1.PatchOptions,
) (*api.VaultPolicy, kutil.VerbType, error) {
	cur, err := c.VaultPolicies(meta.Namespace).Get(ctx, meta.Name, metav1.GetOptions{})
	if kerr.IsNotFound(err) {
		glog.V(3).Infof("Creating VaultPolicy %s/%s.", meta.Namespace, meta.Name)
		out, err := c.VaultPolicies(meta.Namespace).Create(ctx, transform(&api.VaultPolicy{
			TypeMeta: metav1.TypeMeta{
				Kind:       api.ResourceKindVaultPolicy,
				APIVersion: api.SchemeGroupVersion.String(),
			},
			ObjectMeta: meta,
		}), metav1.CreateOptions{
			DryRun:       opts.DryRun,
			FieldManager: opts.FieldManager,
		})
		return out, kutil.VerbCreated, err
	} else if err != nil {
		return nil, kutil.VerbUnchanged, err
	}
	return PatchVaultPolicy(ctx, c, cur, transform, opts)
}

func PatchVaultPolicy(
	ctx context.Context,
	c cs.PolicyV1alpha1Interface,
	cur *api.VaultPolicy,
	transform func(*api.VaultPolicy) *api.VaultPolicy,
	opts metav1.PatchOptions,
) (*api.VaultPolicy, kutil.VerbType, error) {
	return PatchVaultPolicyObject(ctx, c, cur, transform(cur.DeepCopy()), opts)
}

func PatchVaultPolicyObject(
	ctx context.Context,
	c cs.PolicyV1alpha1Interface,
	cur, mod *api.VaultPolicy,
	opts metav1.PatchOptions,
) (*api.VaultPolicy, kutil.VerbType, error) {
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
	out, err := c.VaultPolicies(cur.Namespace).Patch(ctx, cur.Name, types.MergePatchType, patch, opts)
	return out, kutil.VerbPatched, err
}

func TryPatchVaultPolicy(
	ctx context.Context,
	c cs.PolicyV1alpha1Interface,
	cur *api.VaultPolicy,
	transform func(*api.VaultPolicy) *api.VaultPolicy,
	opts metav1.PatchOptions,
) (*api.VaultPolicy, error) {
	var (
		out *api.VaultPolicy
		e2  error
	)
	attempt := 0
	err := wait.PollImmediate(kutil.RetryInterval, kutil.RetryTimeout, func() (bool, error) {
		attempt++
		cur, e2 = c.VaultPolicies(cur.Namespace).Get(ctx, cur.Name, metav1.GetOptions{})
		if kerr.IsNotFound(e2) {
			return false, e2
		} else if e2 == nil {
			out, _, e2 = PatchVaultPolicyObject(ctx, c, cur, transform(cur.DeepCopy()), opts)
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

func TryUpdateVaultPolicy(
	ctx context.Context,
	c cs.PolicyV1alpha1Interface,
	meta metav1.ObjectMeta,
	transform func(*api.VaultPolicy) *api.VaultPolicy,
	opts metav1.UpdateOptions,
) (result *api.VaultPolicy, err error) {
	attempt := 0
	err = wait.PollImmediate(kutil.RetryInterval, kutil.RetryTimeout, func() (bool, error) {
		attempt++
		cur, e2 := c.VaultPolicies(meta.Namespace).Get(ctx, meta.Name, metav1.GetOptions{})
		if kerr.IsNotFound(e2) {
			return false, e2
		} else if e2 == nil {
			result, e2 = c.VaultPolicies(cur.Namespace).Update(ctx, transform(cur.DeepCopy()), opts)
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
	ctx context.Context,
	c cs.PolicyV1alpha1Interface,
	meta metav1.ObjectMeta,
	transform func(*api.VaultPolicyStatus) *api.VaultPolicyStatus,
	opts metav1.UpdateOptions,
) (result *api.VaultPolicy, err error) {
	apply := func(x *api.VaultPolicy) *api.VaultPolicy {
		return &api.VaultPolicy{
			TypeMeta:   x.TypeMeta,
			ObjectMeta: x.ObjectMeta,
			Spec:       x.Spec,
			Status:     *transform(x.Status.DeepCopy()),
		}
	}

	attempt := 0
	cur, err := c.VaultPolicies(meta.Namespace).Get(ctx, meta.Name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	err = wait.PollImmediate(kutil.RetryInterval, kutil.RetryTimeout, func() (bool, error) {
		attempt++
		var e2 error
		result, e2 = c.VaultPolicies(meta.Namespace).UpdateStatus(ctx, apply(cur), opts)
		if kerr.IsConflict(e2) {
			latest, e3 := c.VaultPolicies(meta.Namespace).Get(ctx, meta.Name, metav1.GetOptions{})
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
		err = fmt.Errorf("failed to update status of VaultPolicy %s/%s after %d attempts due to %v", meta.Namespace, meta.Name, attempt, err)
	}
	return
}
