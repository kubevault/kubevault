package admission

import (
	"sync"

	"github.com/pkg/errors"
	admission "k8s.io/api/admission/v1beta1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/rest"
	meta_util "kmodules.xyz/client-go/meta"
	hookapi "kmodules.xyz/webhook-runtime/admission/v1beta1"
	api "kubevault.dev/operator/apis/engine/v1alpha1"
)

const (
	validatorGroupForEngine   = "validators.engine.kubevault.com"
	validatorVersionForEngine = "v1alpha1"
)

type AWSAccessKeyRequestValidator struct {
	lock        sync.RWMutex
	initialized bool
}

var _ hookapi.AdmissionHook = &AWSAccessKeyRequestValidator{}

func (v *AWSAccessKeyRequestValidator) Resource() (plural schema.GroupVersionResource, singular string) {
	return schema.GroupVersionResource{
			Group:    validatorGroupForEngine,
			Version:  validatorVersionForEngine,
			Resource: "awsaccesskeyrequestvalidators",
		},
		"awsaccesskeyrequestvalidator"
}

func (v *AWSAccessKeyRequestValidator) Initialize(config *rest.Config, stopCh <-chan struct{}) error {
	v.lock.Lock()
	defer v.lock.Unlock()

	v.initialized = true
	return nil
}

func (v *AWSAccessKeyRequestValidator) Admit(req *admission.AdmissionRequest) *admission.AdmissionResponse {
	status := &admission.AdmissionResponse{}

	if req.Operation != admission.Update ||
		len(req.SubResource) != 0 ||
		req.Kind.Group != api.SchemeGroupVersion.Group ||
		req.Kind.Kind != api.ResourceKindAWSAccessKeyRequest {
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

		awsAKReq := obj.(*api.AWSAccessKeyRequest).DeepCopy()
		oldAwsAKReq := oldObject.(*api.AWSAccessKeyRequest).DeepCopy()

		isApprovedOrDenied := false

		for _, c := range awsAKReq.Status.Conditions {
			if c.Type == api.AccessApproved || c.Type == api.AccessDenied {
				isApprovedOrDenied = true
			}
		}

		if isApprovedOrDenied {
			// once request is approved or denied, .spec can not be changed
			diff := meta_util.Diff(oldAwsAKReq.Spec, awsAKReq.Spec)
			if diff != "" {
				return hookapi.StatusBadRequest(errors.Errorf("once request is approved or denied, .spec can not be changed. Diff: %s", diff))
			}
		}
	}
	status.Allowed = true
	return status
}
