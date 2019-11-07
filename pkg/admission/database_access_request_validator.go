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

	api "kubevault.dev/operator/apis/engine/v1alpha1"

	"github.com/pkg/errors"
	admission "k8s.io/api/admission/v1beta1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/rest"
	meta_util "kmodules.xyz/client-go/meta"
	hookapi "kmodules.xyz/webhook-runtime/admission/v1beta1"
)

const (
	validatorGroupForDB   = "validators.engine.kubevault.com"
	validatorVersionForDB = "v1alpha1"
)

type DatabaseAccessRequestValidator struct {
	lock        sync.RWMutex
	initialized bool
}

var _ hookapi.AdmissionHook = &DatabaseAccessRequestValidator{}

func (v *DatabaseAccessRequestValidator) Resource() (plural schema.GroupVersionResource, singular string) {
	return schema.GroupVersionResource{
			Group:    validatorGroupForDB,
			Version:  validatorVersionForDB,
			Resource: "databaseaccessrequestvalidators",
		},
		"databaseaccessrequestvalidator"
}

func (v *DatabaseAccessRequestValidator) Initialize(config *rest.Config, stopCh <-chan struct{}) error {
	v.lock.Lock()
	defer v.lock.Unlock()

	v.initialized = true
	return nil
}

func (v *DatabaseAccessRequestValidator) Admit(req *admission.AdmissionRequest) *admission.AdmissionResponse {
	status := &admission.AdmissionResponse{}

	if req.Operation != admission.Update ||
		len(req.SubResource) != 0 ||
		req.Kind.Group != api.SchemeGroupVersion.Group ||
		req.Kind.Kind != api.ResourceKindDatabaseAccessRequest {
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

		dbAReq := obj.(*api.DatabaseAccessRequest).DeepCopy()
		oldDbAReq := oldObject.(*api.DatabaseAccessRequest).DeepCopy()

		isApprovedOrDenied := false

		for _, c := range dbAReq.Status.Conditions {
			if c.Type == api.AccessApproved || c.Type == api.AccessDenied {
				isApprovedOrDenied = true
			}
		}

		if isApprovedOrDenied {
			// once request is approved or denied, .spec can not be changed
			diff := meta_util.Diff(oldDbAReq.Spec, dbAReq.Spec)
			if diff != "" {
				return hookapi.StatusBadRequest(errors.Errorf("once request is approved or denied, .spec can not be changed. Diff: %s", diff))
			}
		}
	}
	status.Allowed = true
	return status
}
