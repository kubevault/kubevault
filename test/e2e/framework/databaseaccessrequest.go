package framework

import (
	api "kubedb.dev/apimachinery/apis/authorization/v1alpha1"
	patchutil "kubedb.dev/apimachinery/client/clientset/versioned/typed/authorization/v1alpha1/util"
	"kubevault.dev/operator/apis"
)

func (f *Framework) UpdateDatabaseAccessRequestStatus(status *api.DatabaseAccessRequestStatus, dbAReq *api.DatabaseAccessRequest) error {
	_, err := patchutil.UpdateDatabaseAccessRequestStatus(f.DBClient.AuthorizationV1alpha1(), dbAReq, func(s *api.DatabaseAccessRequestStatus) *api.DatabaseAccessRequestStatus {
		s = status
		return s
	}, apis.EnableStatusSubresource)
	return err
}
