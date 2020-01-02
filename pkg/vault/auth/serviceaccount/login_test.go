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

package serviceaccount

import (
	"encoding/json"
	"fmt"
	"testing"

	config "kubevault.dev/operator/apis/config/v1alpha1"
)

func TestTry(t *testing.T) {
	st := `
{
   "apiVersion":"config.kubevault.com/v1alpha1",
   "authMethodControllerRole":"k8s.-.default.example-auth-method-controller",
   "authPath":"kubernetes",
   "kind":"VaultServerConfiguration",
   "vaultRole":"example-policy-controller",
   "serviceAccountName":"example",
   "tokenReviewerServiceAccountName":"example-k8s-token-reviewer",
   "usePodServiceAccountForCsiDriver":true
}
`
	var cf config.VaultServerConfiguration
	err := json.Unmarshal([]byte([]byte(st)), &cf)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(cf)
	}
}
