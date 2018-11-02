package serviceaccount

import (
	"encoding/json"
	"fmt"
	"testing"

	config "github.com/kubevault/operator/apis/config/v1alpha1"
)

func TestTry(t *testing.T) {
	st := `
{
   "apiVersion":"config.kubevault.com/v1alpha1",
   "authMethodControllerRole":"k8s.-.default.example-auth-method-controller",
   "authPath":"kubernetes",
   "kind":"VaultServerConfiguration",
   "policyControllerRole":"example-policy-controller",
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
