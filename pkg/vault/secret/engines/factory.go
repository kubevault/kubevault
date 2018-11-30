package engines

import (
	"github.com/kubevault/operator/pkg/vault/secret"
	"github.com/kubevault/operator/pkg/vault/secret/engines/aws"
	"github.com/kubevault/operator/pkg/vault/secret/engines/database"
	"github.com/kubevault/operator/pkg/vault/secret/engines/kv"
	"github.com/kubevault/operator/pkg/vault/secret/engines/pki"
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
	default:
		return nil, errors.Errorf("unsupported/invalid secret engine '%s'", engineName)
	}
}
