package engines

import (
	"github.com/pkg/errors"
	"kubevault.dev/operator/pkg/vault/secret"
	"kubevault.dev/operator/pkg/vault/secret/engines/aws"
	"kubevault.dev/operator/pkg/vault/secret/engines/azure"
	"kubevault.dev/operator/pkg/vault/secret/engines/database"
	"kubevault.dev/operator/pkg/vault/secret/engines/gcp"
	"kubevault.dev/operator/pkg/vault/secret/engines/kv"
	"kubevault.dev/operator/pkg/vault/secret/engines/pki"
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
