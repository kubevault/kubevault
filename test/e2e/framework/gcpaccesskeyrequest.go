package framework

import (
	"github.com/kubevault/operator/apis"
	api "github.com/kubevault/operator/apis/engine/v1alpha1"
	patchutil "github.com/kubevault/operator/client/clientset/versioned/typed/engine/v1alpha1/util"
)

func (f *Framework) UpdateGCPAccessKeyRequestStatus(status *api.GCPAccessKeyRequestStatus, dbAReq *api.GCPAccessKeyRequest) error {
	_, err := patchutil.UpdateGCPAccessKeyRequestStatus(f.CSClient.EngineV1alpha1(), dbAReq, func(s *api.GCPAccessKeyRequestStatus) *api.GCPAccessKeyRequestStatus {
		s = status
		return s
	}, apis.EnableStatusSubresource)
	return err
}
