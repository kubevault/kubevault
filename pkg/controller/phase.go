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
	"kubevault.dev/apimachinery/apis"
	api "kubevault.dev/apimachinery/apis/kubevault/v1alpha1"

	kmapi "kmodules.xyz/client-go/api/v1"
)

func GetPhase(conditions []kmapi.Condition) api.VaultServerPhase {
	//Todo: Phases from condition array[]:
	//	-Initializing -> at the beginning (till condition initialized is true)
	//	-Unsealing -> unsealing has started but has not completed yet
	//	-Sealed -> unsealed false & initialized true
	//	-Ready -> accepting connection true, unsealed true, initialized true, replicas ready true
	//	-NotReady -> accepting connection false, unsealed true
	//	-Critical -> replica ready false, but accepting connection true

	var phase api.VaultServerPhase

	if kmapi.IsConditionTrue(conditions, apis.VaultServerInitializing) {
		phase = api.VaultServerPhaseInitializing
	}

	if kmapi.IsConditionTrue(conditions, apis.VaultServerUnsealing) {
		phase = api.VaultServerPhaseUnsealing
	}

	if kmapi.IsConditionTrue(conditions, apis.VaultServerInitialized) && kmapi.IsConditionFalse(conditions, apis.VaultServerUnsealed) {
		phase = api.VaultServerPhaseSealed
	}

	if kmapi.IsConditionFalse(conditions, apis.VaultServerAcceptingConnection) && kmapi.IsConditionTrue(conditions, apis.VaultServerUnsealed) {
		phase = api.VaultServerPhaseNotReady
	}

	if kmapi.IsConditionTrue(conditions, apis.VaultServerAcceptingConnection) && kmapi.IsConditionFalse(conditions, apis.AllReplicasAreReady) {
		phase = api.VaultServerPhaseCritical
	}

	if kmapi.IsConditionTrue(conditions, apis.VaultServerInitialized) && kmapi.IsConditionTrue(conditions, apis.VaultServerUnsealed) &&
		kmapi.IsConditionTrue(conditions, apis.VaultServerAcceptingConnection) && kmapi.IsConditionTrue(conditions, apis.AllReplicasAreReady) {
		phase = api.VaultServerPhaseReady
	}

	return phase
}
