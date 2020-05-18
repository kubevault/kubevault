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

package util

import (
	"encoding/json"
	"fmt"

	api "kubevault.dev/operator/apis/engine/v1alpha1"
	cs "kubevault.dev/operator/client/clientset/versioned/typed/engine/v1alpha1"

	jsonpatch "github.com/evanphx/json-patch"
	"github.com/golang/glog"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	kutil "kmodules.xyz/client-go"
)

func PatchSecretEngine(c cs.EngineV1alpha1Interface, cur *api.SecretEngine, transform func(engine *api.SecretEngine) *api.SecretEngine) (*api.SecretEngine, kutil.VerbType, error) {
	return PatchSecretEngineObject(c, cur, transform(cur.DeepCopy()))
}

func PatchSecretEngineObject(c cs.EngineV1alpha1Interface, cur, mod *api.SecretEngine) (*api.SecretEngine, kutil.VerbType, error) {
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
	glog.V(3).Infof("Patching SecretEngine %s/%s with %s.", cur.Namespace, cur.Name, string(patch))
	out, err := c.SecretEngines(cur.Namespace).Patch(cur.Name, types.MergePatchType, patch)
	return out, kutil.VerbPatched, err
}

func UpdateSecretEngineStatus(
	c cs.EngineV1alpha1Interface,
	meta metav1.ObjectMeta,
	transform func(*api.SecretEngineStatus) *api.SecretEngineStatus,
) (result *api.SecretEngine, err error) {
	apply := func(x *api.SecretEngine) *api.SecretEngine {
		return &api.SecretEngine{
			TypeMeta:   x.TypeMeta,
			ObjectMeta: x.ObjectMeta,
			Spec:       x.Spec,
			Status:     *transform(x.Status.DeepCopy()),
		}
	}

	attempt := 0
	cur, err := c.SecretEngines(meta.Namespace).Get(meta.Name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	err = wait.PollImmediate(kutil.RetryInterval, kutil.RetryTimeout, func() (bool, error) {
		attempt++
		var e2 error
		result, e2 = c.SecretEngines(meta.Namespace).UpdateStatus(apply(cur))
		if kerr.IsConflict(e2) {
			latest, e3 := c.SecretEngines(meta.Namespace).Get(meta.Name, metav1.GetOptions{})
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
		err = fmt.Errorf("failed to update status of SecretEngine %s/%s after %d attempts due to %v", meta.Namespace, meta.Name, attempt, err)
	}
	return
}
