package framework

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/appscode/go/crypto/rand"
	aggregator "github.com/appscode/go/util/errors"
	"github.com/appscode/kutil/tools/certstore"
	cs "github.com/kubevault/operator/client/clientset/versioned"
	. "github.com/onsi/gomega"
	"github.com/spf13/afero"
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
	VaultAppRef     *appcat.AppReference
}

func New(kubeClient kubernetes.Interface, extClient cs.Interface, appc appcat_cs.AppcatalogV1alpha1Interface, kaClient ka.Interface, webhookEnabled bool, clientConfig *rest.Config, runDynamoDBTest bool) *Framework {
	store, err := certstore.NewCertStore(afero.NewMemMapFs(), filepath.Join("", "pki"))
	Expect(err).NotTo(HaveOccurred())

	err = store.InitCA()
	Expect(err).NotTo(HaveOccurred())

	return &Framework{
		KubeClient:      kubeClient,
		CSClient:        extClient,
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
	f.VaultAppRef, err = f.DeployVault()
	if err != nil {
		return err
	}
	fmt.Println(f.VaultAppRef)
	return nil
}

func (f *Framework) Cleanup() error {
	errs := []error{}
	err := f.DeleteVault()
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
