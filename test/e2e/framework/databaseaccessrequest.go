package framework

import (
	api "kubevault.dev/operator/apis/engine/v1alpha1"
	patchutil "kubevault.dev/operator/client/clientset/versioned/typed/engine/v1alpha1/util"
)

func (f *Framework) UpdateDatabaseAccessRequestStatus(status *api.DatabaseAccessRequestStatus, dbAReq *api.DatabaseAccessRequest) error {
	_, err := patchutil.UpdateDatabaseAccessRequestStatus(f.CSClient.EngineV1alpha1(), dbAReq, func(s *api.DatabaseAccessRequestStatus) *api.DatabaseAccessRequestStatus {
		return status
	})
	return err
}
