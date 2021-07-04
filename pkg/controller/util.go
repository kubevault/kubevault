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
	"fmt"
	"sync"

	"kubevault.dev/apimachinery/apis"

	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/klog/v2"
	kmapi "kmodules.xyz/client-go/api/v1"
	dmcond "kmodules.xyz/client-go/dynamic/conditions"
)

// contains the key of the currently processing finalizer
// it's concurrency safe
type mapFinalizer struct {
	keys map[string]bool
	lock *sync.Mutex
}

func NewMapFinalizer() *mapFinalizer {
	return &mapFinalizer{
		keys: make(map[string]bool),
		lock: &sync.Mutex{},
	}
}

func (f *mapFinalizer) IsAlreadyProcessing(key string) bool {
	f.lock.Lock()
	defer f.lock.Unlock()
	_, ok := f.keys[key]
	return ok
}

func (f *mapFinalizer) Add(key string) {
	f.lock.Lock()
	defer f.lock.Unlock()
	f.keys[key] = true
}

func (f *mapFinalizer) Delete(key string) {
	f.lock.Lock()
	defer f.lock.Unlock()
	delete(f.keys, key)
}

type CtxWithCancel struct {
	Ctx    context.Context
	Cancel context.CancelFunc
}

type vaultserverInfo struct {
	opts          dmcond.DynamicOptions
	replicasReady bool
	msg           string
}

func (c *VaultController) extractVaultserverInfo(sts *apps.StatefulSet) (*vaultserverInfo, error) {
	klog.Infoln("=============================== extraceVaultserverInfo ==========================")
	// read the controlling owner
	owner := metav1.GetControllerOf(sts)
	if owner == nil {
		return nil, fmt.Errorf("StatefulSet %s/%s has no controlling owner", sts.Namespace, sts.Name)
	}
	klog.Infoln("=============================== Owner ========================== ", owner)

	gv, err := schema.ParseGroupVersion(owner.APIVersion)
	if err != nil {
		return nil, err
	}
	vsInfo := &vaultserverInfo{
		opts: dmcond.DynamicOptions{
			Client:    c.DynamicClient,
			Kind:      owner.Kind,
			Name:      owner.Name,
			Namespace: sts.Namespace,
		},
	}
	vsInfo.opts.GVR = schema.GroupVersionResource{
		Group:   gv.Group,
		Version: gv.Version,
	}
	switch owner.Kind {
	case apis.ResourceKindStatefulSet:
		klog.Infoln("=============================== switch case in ==========================")
		vsInfo.opts.GVR.Resource = apis.ResourceKindStatefulSet
		vs, err := c.extClient.KubevaultV1alpha1().VaultServers(vsInfo.opts.Namespace).Get(context.TODO(), vsInfo.opts.Name, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}
		klog.Infoln("=============================== switch in, vs ok ==========================")

		vsInfo.replicasReady, vsInfo.msg, err = vs.ReplicasAreReady(c.StsLister)
		if err != nil {
			return nil, err
		}
		klog.Infoln("=============================== vsInfo ==========================", vsInfo.replicasReady, vsInfo.msg)
	default:
		return nil, fmt.Errorf("unknown resource kind: %s", owner.Kind)
	}
	return vsInfo, nil
}

func (c *VaultController) ensureReadyReplicasCond(vsInfo *vaultserverInfo) error {
	klog.Infoln("=============================== ensureReadyReplicasCond ==========================")
	vsCond := kmapi.Condition{
		Type:    apis.VaultserverReplicaReady,
		Message: vsInfo.msg,
	}

	if vsInfo.replicasReady {
		vsCond.Status = core.ConditionTrue
		vsCond.Reason = apis.AllReplicasAreReady
	} else {
		vsCond.Status = core.ConditionFalse
		vsCond.Reason = apis.SomeReplicasAreNotReady
	}
	klog.Infoln("=============================== ensureReadyReplicasCond ==========================", vsCond.Status, vsCond.Reason)
	// Add "ReplicasReady" condition to the respective vaultserver CR
	return vsInfo.opts.SetCondition(vsCond)
}
