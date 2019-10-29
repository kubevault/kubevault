package framework

import (
	"log"
	"path/filepath"
	"time"

	cs "kubevault.dev/operator/client/clientset/versioned"
	db_cs "kubevault.dev/operator/client/clientset/versioned"

	"github.com/appscode/go/crypto/rand"
	aggregator "github.com/appscode/go/util/errors"
	. "github.com/onsi/gomega"
	"github.com/spf13/afero"
	"gomodules.xyz/cert/certstore"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	ka "k8s.io/kube-aggregator/pkg/client/clientset_generated/clientset"
	appcat "kmodules.xyz/custom-resources/apis/appcatalog/v1alpha1"
	appcat_cs "kmodules.xyz/custom-resources/client/clientset/versioned/typed/appcatalog/v1alpha1"
)

const (
	timeOut         = 10 * time.Minute
	pollingInterval = 10 * time.Second
)

var (
	SelfHostedOperator = true
	DockerRegistry     = "kubevault"
	UnsealerImage      = "vault-unsealer:v0.3.0"
)

type Framework struct {
	KubeClient      kubernetes.Interface
	CSClient        cs.Interface
	AppcatClient    appcat_cs.AppcatalogV1alpha1Interface
	KAClient        ka.Interface
	namespace       string
	CertStore       *certstore.CertStore
	WebhookEnabled  bool
	ClientConfig    *rest.Config
	RunDynamoDBTest bool
	DBClient        db_cs.Interface

	VaultAppRef    *appcat.AppReference
	MongoAppRef    *appcat.AppReference
	MysqlAppRef    *appcat.AppReference
	PostgresAppRef *appcat.AppReference
}

func New(kubeClient kubernetes.Interface, extClient cs.Interface, appc appcat_cs.AppcatalogV1alpha1Interface, dbClient db_cs.Interface, kaClient ka.Interface, webhookEnabled bool, clientConfig *rest.Config, runDynamoDBTest bool) *Framework {
	store, err := certstore.NewCertStore(afero.NewMemMapFs(), filepath.Join("", "pki"))
	Expect(err).NotTo(HaveOccurred())

	err = store.InitCA()
	Expect(err).NotTo(HaveOccurred())

	return &Framework{
		KubeClient:      kubeClient,
		CSClient:        extClient,
		DBClient:        dbClient,
		AppcatClient:    appc,
		KAClient:        kaClient,
		namespace:       rand.WithUniqSuffix("test-vault"),
		CertStore:       store,
		WebhookEnabled:  webhookEnabled,
		ClientConfig:    clientConfig,
		RunDynamoDBTest: runDynamoDBTest,
	}
}

func (f *Framework) InitialSetup() error {
	var err error
	if !SelfHostedOperator {
		log.Println("Deploying vault...")
		f.VaultAppRef, err = f.DeployVault()
		if err != nil {
			return err
		}
	} else {
		log.Println("Deploying vault...")
		f.VaultAppRef, err = f.DeployVaultServer()
		if err != nil {
			return err
		}
	}

	log.Println("Deploying Mongodb...")
	f.MongoAppRef, err = f.DeployMongodb()
	if err != nil {
		return err
	}

	log.Println("Deploying Mysql... ")
	f.MysqlAppRef, err = f.DeployMysql()
	if err != nil {
		return err
	}

	log.Println("Deploying postgres...")
	f.PostgresAppRef, err = f.DeployPostgres()
	if err != nil {
		return err
	}
	return nil
}

func (f *Framework) Cleanup() error {
	errs := []error{}
	if !SelfHostedOperator {
		err := f.DeleteVault()
		if err != nil {
			errs = append(errs, err)
		}
	} else {
		err := f.CleanUpVaultServer()
		if err != nil {
			errs = append(errs, err)
		}
	}

	err := f.DeleteMongodb()
	if err != nil {
		errs = append(errs, err)
	}

	err = f.DeleteMysql()
	if err != nil {
		errs = append(errs, err)
	}

	err = f.DeletePostgres()
	if err != nil {
		errs = append(errs, err)
	}

	if len(errs) != 0 {
		return aggregator.NewAggregate(errs)
	}
	return nil
}

func (f *Framework) Invoke() *Invocation {
	return &Invocation{
		Framework: f,
		app:       rand.WithUniqSuffix("vault-e2e"),
	}
}

type Invocation struct {
	*Framework
	app string
}
