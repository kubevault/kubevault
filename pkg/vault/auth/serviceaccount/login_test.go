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

package serviceaccount

import (
	"encoding/json"
	"fmt"
	"testing"

	config "kubevault.dev/apimachinery/apis/config/v1alpha1"
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
