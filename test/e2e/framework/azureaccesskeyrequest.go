package framework

import (
	"github.com/kubevault/operator/apis"
	api "github.com/kubevault/operator/apis/engine/v1alpha1"
	patchutil "github.com/kubevault/operator/client/clientset/versioned/typed/engine/v1alpha1/util"
)

func (f *Framework) UpdateAzureAccessKeyRequestStatus(status *api.AzureAccessKeyRequestStatus, azureAReq *api.AzureAccessKeyRequest) error {
	_, err := patchutil.UpdateAzureAccessKeyRequestStatus(f.CSClient.EngineV1alpha1(), azureAReq, func(s *api.AzureAccessKeyRequestStatus) *api.AzureAccessKeyRequestStatus {
		s = status
		return s
	}, apis.EnableStatusSubresource)
	return err
}
