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

package admission

import (
	"sync"

	api "kubevault.dev/operator/apis/kubevault/v1alpha1"

	admission "k8s.io/api/admission/v1beta1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/rest"
	meta_util "kmodules.xyz/client-go/meta"
	hookapi "kmodules.xyz/webhook-runtime/admission/v1beta1"
)

const (
	authTypeKubernetes = "kubernetes"
)

type VaultServerMutator struct {
	lock        sync.RWMutex
	initialized bool
}

var _ hookapi.AdmissionHook = &VaultServerMutator{}

func (v *VaultServerMutator) Resource() (plural schema.GroupVersionResource, singular string) {
	return schema.GroupVersionResource{
			Group:    mutatorGroup,
			Version:  mutatorVersion,
			Resource: "vaultservermutators",
		},
		"vaultservermutator"
}

func (v *VaultServerMutator) Initialize(config *rest.Config, stopCh <-chan struct{}) error {
	v.lock.Lock()
	defer v.lock.Unlock()

	v.initialized = true
	return nil
}

func (v *VaultServerMutator) Admit(req *admission.AdmissionRequest) *admission.AdmissionResponse {
	status := &admission.AdmissionResponse{}

	if (req.Operation != admission.Create && req.Operation != admission.Update) ||
		len(req.SubResource) != 0 ||
		req.Kind.Group != api.SchemeGroupVersion.Group ||
		req.Kind.Kind != api.ResourceKindVaultServer {
		status.Allowed = true
		return status
	}

	v.lock.RLock()
	defer v.lock.RUnlock()
	if !v.initialized {
		return hookapi.StatusUninitialized()
	}

	obj, err := meta_util.UnmarshalFromJSON(req.Object.Raw, api.SchemeGroupVersion)
	if err != nil {
		return hookapi.StatusBadRequest(err)
	}
	mod := setVaultServerDefaultValues(obj.(*api.VaultServer).DeepCopy())
	patch, err := meta_util.CreateJSONPatch(req.Object.Raw, mod)
	if err != nil {
		return hookapi.StatusInternalServerError(err)
	}
	status.Patch = patch
	patchType := admission.PatchTypeJSONPatch
	status.PatchType = &patchType

	status.Allowed = true
	return status
}

func setVaultServerDefaultValues(vs *api.VaultServer) *api.VaultServer {
	vs.Spec.AuthMethods = upsertAuthMethods(vs.Spec.AuthMethods, api.AuthMethod{
		Type: authTypeKubernetes,
		Path: authTypeKubernetes,
	})
	return vs
}

func upsertAuthMethods(auths []api.AuthMethod, newAuth api.AuthMethod) []api.AuthMethod {
	for _, au := range auths {
		if au.Type == newAuth.Type && au.Path == newAuth.Path {
			return auths
		}
	}
	return append(auths, newAuth)
}
