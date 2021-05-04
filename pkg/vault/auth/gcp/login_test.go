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

package gcp

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	authtype "kubevault.dev/operator/pkg/vault/auth/types"

	"github.com/stretchr/testify/assert"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	appcat "kmodules.xyz/custom-resources/apis/appcatalog/v1alpha1"
)

func TestLogin(t *testing.T) {
	addr := os.Getenv("VAULT_ADDR")
	credentialaddr := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
	role := os.Getenv("VAULT_GCP_ROLE_NAME")
	if addr == "" || credentialaddr == "" || role == "" {
		t.Skip()
	}

	jsonBytes, err := ioutil.ReadFile(credentialaddr)
	if err != nil {
		klog.Fatal(err)
	}

	au, err := New(&authtype.AuthInfo{
		VaultApp: &appcat.AppBinding{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "gcp",
				Namespace: "default",
			},
			Spec: appcat.AppBindingSpec{
				ClientConfig: appcat.ClientConfig{
					URL:                   &addr,
					InsecureSkipTLSVerify: true,
				},
				Secret: &core.LocalObjectReference{
					Name: "gcp",
				},
				Parameters: &runtime.RawExtension{
					Raw: []byte(fmt.Sprintf(`{ "VaultRole" : "%s" }`, role)),
				},
			},
		},
		ServiceAccountRef: nil,
		Secret: &core.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "gcp",
				Namespace: "default",
			},
			Data: map[string][]byte{
				"sa.json": []byte(jsonBytes),
			},
		},
		VaultRole: "",
		Path:      "",
	})

	if err != nil {
		klog.Println("New failed!")
	}
	if au == nil {
		klog.Println("au nil!")
		t.Skip()
	}

	if au.signedJwt == "" || au.role == "" {
		t.Skip()
	}
	token, err := au.Login()
	if assert.Nil(t, err) {
		fmt.Println(token)
	}
}
