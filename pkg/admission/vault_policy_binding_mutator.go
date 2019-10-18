package admission

import (
	"sync"

	admission "k8s.io/api/admission/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/rest"
	meta_util "kmodules.xyz/client-go/meta"
	hookapi "kmodules.xyz/webhook-runtime/admission/v1beta1"
	api "kubevault.dev/operator/apis/policy/v1alpha1"
)

const (
	mutatorGroup   = "mutators.kubevault.com"
	mutatorVersion = "v1alpha1"
)

type PolicyBindingMutator struct {
	lock        sync.RWMutex
	initialized bool
}

var _ hookapi.AdmissionHook = &PolicyBindingMutator{}

func (a *PolicyBindingMutator) Resource() (plural schema.GroupVersionResource, singular string) {
	return schema.GroupVersionResource{
			Group:    mutatorGroup,
			Version:  mutatorVersion,
			Resource: "vaultpolicybindingmutators",
		},
		"vaultpolicybindingmutator"
}

func (a *PolicyBindingMutator) Initialize(config *rest.Config, stopCh <-chan struct{}) error {
	a.lock.Lock()
	defer a.lock.Unlock()

	a.initialized = true
	return nil
}

func (a *PolicyBindingMutator) Admit(req *admission.AdmissionRequest) *admission.AdmissionResponse {
	status := &admission.AdmissionResponse{}

	// N.B.: No Mutating for delete
	if (req.Operation != admission.Create && req.Operation != admission.Update) ||
		len(req.SubResource) != 0 ||
		req.Kind.Group != api.SchemeGroupVersion.Group ||
		req.Kind.Kind != api.ResourceKindVaultPolicyBinding {
		status.Allowed = true
		return status
	}

	a.lock.RLock()
	defer a.lock.RUnlock()
	if !a.initialized {
		return hookapi.StatusUninitialized()
	}
	obj, err := meta_util.UnmarshalFromJSON(req.Object.Raw, api.SchemeGroupVersion)
	if err != nil {
		return hookapi.StatusBadRequest(err)
	}
	mod := setDefaultValues(obj.(*api.VaultPolicyBinding).DeepCopy())
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

// setDefaultValues provides the defaulting that is performed in mutating stage of creating/updating a VaultPolicyBinding
func setDefaultValues(pb *api.VaultPolicyBinding) runtime.Object {
	pb.SetDefaults()
	return pb
}
