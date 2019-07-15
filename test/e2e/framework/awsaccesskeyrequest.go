package framework

import (
	"kubevault.dev/operator/apis"
	api "kubevault.dev/operator/apis/engine/v1alpha1"
	patchutil "kubevault.dev/operator/client/clientset/versioned/typed/engine/v1alpha1/util"
)

func (f *Framework) UpdateAWSAccessKeyRequestStatus(status *api.AWSAccessKeyRequestStatus, dbAReq *api.AWSAccessKeyRequest) error {
	_, err := patchutil.UpdateAWSAccessKeyRequestStatus(f.CSClient.EngineV1alpha1(), dbAReq, func(s *api.AWSAccessKeyRequestStatus) *api.AWSAccessKeyRequestStatus {
		s = status
		return s
	}, apis.EnableStatusSubresource)
	return err
}
