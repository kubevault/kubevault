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
package controller

import (
	"reflect"
	"testing"

	api "kubevault.dev/operator/apis/engine/v1alpha1"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestDeleteAzureAccessKeyCondition(t *testing.T) {
	type args struct {
		condList []api.AzureAccessKeyRequestCondition
		condType api.RequestConditionType
	}
	tests := []struct {
		name string
		args args
		want []api.AzureAccessKeyRequestCondition
	}{
		{
			name: "test-1",
			args: args{
				condList: []api.AzureAccessKeyRequestCondition{
					{
						Type:           "a",
						Reason:         "a",
						Message:        "a",
						LastUpdateTime: v1.Time{},
					}, {
						Type:           "a",
						Reason:         "b",
						Message:        "b",
						LastUpdateTime: v1.Time{},
					},
				},
				condType: "a",
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := DeleteAzureAccessKeyCondition(tt.args.condList, tt.args.condType); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DeleteAzureAccessKeyCondition() = %v, want %v", got, tt.want)
			}
		})
	}
}
