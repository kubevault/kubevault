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

package controller

import (
	"testing"

	api "kubevault.dev/apimachinery/apis/kubevault/v1alpha1"

	kmapi "kmodules.xyz/client-go/api/v1"
)

func TestGetPhase(t *testing.T) {
	testCases := []struct {
		name          string
		conditions    []kmapi.Condition
		expectedPhase api.ClusterPhase
	}{
		{},
		{},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if got := GetPhase(tc.conditions); got != tc.expectedPhase {
				t.Errorf("Expected: %s Found: %s", tc.expectedPhase, got)
			}
		})
	}
}
