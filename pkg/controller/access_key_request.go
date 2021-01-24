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
	"fmt"

	api "kubevault.dev/apimachinery/apis/engine/v1alpha1"
	"kubevault.dev/operator/pkg/vault/credential"

	"github.com/pkg/errors"
	meta_util "kmodules.xyz/client-go/meta"
)

func revokeLease(credM credential.CredentialManager, lease *api.Lease) error {
	if credM == nil {
		return errors.New("credential manager is empty")
	}
	// If lease is not set or leaseID is empty,
	// return nil.
	if lease == nil {
		return nil
	}
	if lease.ID == "" {
		return nil
	}
	return credM.RevokeLease(lease.ID)
}

func getSecretAccessRoleName(kind, namespace, name string) string {
	n := fmt.Sprintf("%s-%s-%s", kind, namespace, name)
	return meta_util.NameWithSuffix(n, "cred-reader")
}
