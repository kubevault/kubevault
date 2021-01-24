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

package engines

import (
	"kubevault.dev/operator/pkg/vault/secret"
	"kubevault.dev/operator/pkg/vault/secret/engines/aws"
	"kubevault.dev/operator/pkg/vault/secret/engines/azure"
	"kubevault.dev/operator/pkg/vault/secret/engines/database"
	"kubevault.dev/operator/pkg/vault/secret/engines/gcp"
	"kubevault.dev/operator/pkg/vault/secret/engines/kv"
	"kubevault.dev/operator/pkg/vault/secret/engines/pki"

	"github.com/pkg/errors"
)

func NewSecretManager(engineName string) (secret.SecretManager, error) {
	switch engineName {
	case aws.UID:
		return aws.NewSecretManager(), nil
	case pki.UID:
		return pki.NewSecretManager(), nil
	case kv.UID:
		return kv.NewSecretManager(), nil
	case database.UID:
		return database.NewSecretManager(), nil
	case gcp.UID:
		return gcp.NewSecretManager(), nil
	case azure.UID:
		return azure.NewSecretManager(), nil
	default:
		return nil, errors.Errorf("unsupported/invalid secret engine '%s'", engineName)
	}
}
