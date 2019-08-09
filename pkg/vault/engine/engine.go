package engine

import (
	vaultapi "github.com/hashicorp/vault/api"
	"github.com/pkg/errors"
	"k8s.io/client-go/kubernetes"
	appcat "kmodules.xyz/custom-resources/apis/appcatalog/v1alpha1"
	appcat_cs "kmodules.xyz/custom-resources/client/clientset/versioned/typed/appcatalog/v1alpha1"
	api "kubevault.dev/operator/apis/engine/v1alpha1"
	"kubevault.dev/operator/pkg/vault"
	"kubevault.dev/operator/pkg/vault/role/aws"
	"kubevault.dev/operator/pkg/vault/role/azure"
	"kubevault.dev/operator/pkg/vault/role/database"
	"kubevault.dev/operator/pkg/vault/role/gcp"
)

type SecretEngine struct {
	appClient    appcat_cs.AppcatalogV1alpha1Interface
	secretEngine *api.SecretEngine
	vaultClient  *vaultapi.Client
	kubeClient   kubernetes.Interface
	path         string
}

func NewSecretEngine(kClient kubernetes.Interface, appClient appcat_cs.AppcatalogV1alpha1Interface, engine *api.SecretEngine) (*SecretEngine, error) {
	vAppRef := &appcat.AppReference{
		Namespace: engine.Namespace,
		Name:      engine.Spec.VaultRef.Name,
	}

	vClient, err := vault.NewClient(kClient, appClient, vAppRef)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create vault api client")
	}
	// If path is not provided then set path to
	// default secret engine path (i.e. "gcp", "aws", "azure", "database")
	path := GetSecretEnginePath(engine)

	return &SecretEngine{
		appClient:    appClient,
		kubeClient:   kClient,
		vaultClient:  vClient,
		secretEngine: engine,
		path:         path,
	}, nil
}

func GetSecretEnginePath(engine *api.SecretEngine) string {
	if engine.Spec.Path != "" {
		return engine.Spec.Path
	}
	if engine.Spec.GCP != nil {
		return gcp.DefaultGCPPath
	}
	if engine.Spec.AWS != nil {
		return aws.DefaultAWSPath
	}
	if engine.Spec.Azure != nil {
		return azure.DefaultAzurePath
	}
	return database.DefaultDatabasePath
}

// checks whether SecretEngine is enabled or not
func (secretEngineClient *SecretEngine) IsSecretEngineEnabled() (bool, error) {
	mnt, err := secretEngineClient.vaultClient.Sys().ListMounts()
	if err != nil {
		return false, errors.Wrap(err, "failed to list mounted secrets engines")
	}

	mntPath := secretEngineClient.path + "/"
	for k := range mnt {
		if k == mntPath {
			return true, nil
		}
	}
	return false, nil
}

// It enables secret engine
// It first checks whether secret engine is enabled or not
func (secretEngineClient *SecretEngine) EnableSecretEngine() error {
	enabled, err := secretEngineClient.IsSecretEngineEnabled()
	if err != nil {
		return err
	}

	if enabled {
		return nil
	}
	var engineType string
	if secretEngineClient.secretEngine.Spec.AWS != nil {
		engineType = api.EngineTypeAWS
	} else if secretEngineClient.secretEngine.Spec.GCP != nil {
		engineType = api.EngineTypeGCP
	} else if secretEngineClient.secretEngine.Spec.Azure != nil {
		engineType = api.EngineTypeAzure
	} else if secretEngineClient.secretEngine.Spec.MongoDB != nil ||
		secretEngineClient.secretEngine.Spec.Postgres != nil ||
		secretEngineClient.secretEngine.Spec.MySQL != nil {
		engineType = api.EngineTypeDatabase
	} else {
		return errors.New("Failed to enable secret engine: Unknown secret engine type")
	}

	err = secretEngineClient.vaultClient.Sys().Mount(secretEngineClient.path, &vaultapi.MountInput{
		Type: engineType,
	})
	if err != nil {
		return err
	}
	return nil
}
