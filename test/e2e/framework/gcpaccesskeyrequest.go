package framework

import (
	api "kubevault.dev/operator/apis/engine/v1alpha1"
	patchutil "kubevault.dev/operator/client/clientset/versioned/typed/engine/v1alpha1/util"
)

func (f *Framework) UpdateGCPAccessKeyRequestStatus(status *api.GCPAccessKeyRequestStatus, dbAReq *api.GCPAccessKeyRequest) error {
	_, err := patchutil.UpdateGCPAccessKeyRequestStatus(f.CSClient.EngineV1alpha1(), dbAReq, func(s *api.GCPAccessKeyRequestStatus) *api.GCPAccessKeyRequestStatus {
		s = status
		return s
	})
	return err
}
