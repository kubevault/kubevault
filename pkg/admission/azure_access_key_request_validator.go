package admission

import (
	"sync"

	api "kubevault.dev/operator/apis/engine/v1alpha1"

	"github.com/pkg/errors"
	admission "k8s.io/api/admission/v1beta1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/rest"
	meta_util "kmodules.xyz/client-go/meta"
	hookapi "kmodules.xyz/webhook-runtime/admission/v1beta1"
)

type AzureAccessKeyRequestValidator struct {
	lock        sync.RWMutex
	initialized bool
}

var _ hookapi.AdmissionHook = &AzureAccessKeyRequestValidator{}

func (v *AzureAccessKeyRequestValidator) Resource() (plural schema.GroupVersionResource, singular string) {
	return schema.GroupVersionResource{
			Group:    validatorGroupForEngine,
			Version:  validatorVersionForEngine,
			Resource: "azureaccesskeyrequestvalidators",
		},
		"azureaccesskeyrequestvalidator"
}

func (v *AzureAccessKeyRequestValidator) Initialize(config *rest.Config, stopCh <-chan struct{}) error {
	v.lock.Lock()
	defer v.lock.Unlock()

	v.initialized = true
	return nil
}

func (v *AzureAccessKeyRequestValidator) Admit(req *admission.AdmissionRequest) *admission.AdmissionResponse {
	status := &admission.AdmissionResponse{}

	if req.Operation != admission.Update ||
		len(req.SubResource) != 0 ||
		req.Kind.Group != api.SchemeGroupVersion.Group ||
		req.Kind.Kind != api.ResourceKindAzureAccessKeyRequest {
		status.Allowed = true
		return status
	}

	v.lock.RLock()
	defer v.lock.RUnlock()
	if !v.initialized {
		return hookapi.StatusUninitialized()
	}

	if req.Operation == admission.Update {
		obj, err := meta_util.UnmarshalFromJSON(req.Object.Raw, api.SchemeGroupVersion)
		if err != nil {
			return hookapi.StatusBadRequest(err)
		}
		// validate changes made by user
		oldObject, err := meta_util.UnmarshalFromJSON(req.OldObject.Raw, api.SchemeGroupVersion)
		if err != nil {
			return hookapi.StatusBadRequest(err)
		}

		azureAKReq := obj.(*api.AzureAccessKeyRequest).DeepCopy()
		oldAzureAKReq := oldObject.(*api.AzureAccessKeyRequest).DeepCopy()

		isApprovedOrDenied := false

		for _, c := range azureAKReq.Status.Conditions {
			if c.Type == api.AccessApproved || c.Type == api.AccessDenied {
				isApprovedOrDenied = true
			}
		}

		if isApprovedOrDenied {
			// once request is approved or denied, .spec can not be changed
			diff := meta_util.Diff(oldAzureAKReq.Spec, azureAKReq.Spec)
			if diff != "" {
				return hookapi.StatusBadRequest(errors.Errorf("once request is approved or denied, .spec can not be changed. Diff: %s", diff))
			}
		}
	}
	status.Allowed = true
	return status
}
