/*
Copyright AppsCode Inc. and Contributors

Licensed under the AppsCode Community License 1.0.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    https://github.com/appscode/licenses/raw/1.0.0/AppsCode-Community-1.0.0.md

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package framework

import (
	"context"

	api "kubevault.dev/apimachinery/apis/engine/v1alpha1"
	patchutil "kubevault.dev/apimachinery/client/clientset/versioned/typed/engine/v1alpha1/util"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (f *Framework) UpdateGCPAccessKeyRequestStatus(status *api.GCPAccessKeyRequestStatus, dbAReq *api.GCPAccessKeyRequest) error {
	_, err := patchutil.UpdateGCPAccessKeyRequestStatus(context.TODO(), f.CSClient.EngineV1alpha1(), dbAReq.ObjectMeta, func(s *api.GCPAccessKeyRequestStatus) *api.GCPAccessKeyRequestStatus {
		return status
	}, metav1.UpdateOptions{})
	return err
}
