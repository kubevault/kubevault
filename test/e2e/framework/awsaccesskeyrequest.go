package framework

import (
	"github.com/kubevault/operator/apis"
	api "github.com/kubevault/operator/apis/engine/v1alpha1"
	patchutil "github.com/kubevault/operator/client/clientset/versioned/typed/engine/v1alpha1/util"
)

func (f *Framework) UpdateAWSAccessKeyRequestStatus(status *api.AWSAccessKeyRequestStatus, dbAReq *api.AWSAccessKeyRequest) error {
	_, err := patchutil.UpdateAWSAccessKeyRequestStatus(f.CSClient.EngineV1alpha1(), dbAReq, func(s *api.AWSAccessKeyRequestStatus) *api.AWSAccessKeyRequestStatus {
		s = status
		return s
	}, apis.EnableStatusSubresource)
	return err
}
