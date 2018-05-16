package controller

import (
	"context"
	"fmt"
	"time"

	crdutils "github.com/appscode/kutil/apiextensions/v1beta1"
	"github.com/appscode/kutil/tools/queue"
	"github.com/golang/glog"
	vaultapi "github.com/hashicorp/vault/api"
	api "github.com/soter/vault-operator/apis/vault/v1alpha1"
	cs "github.com/soter/vault-operator/client/clientset/versioned"
	vaultinformers "github.com/soter/vault-operator/client/informers/externalversions"
	vault_listers "github.com/soter/vault-operator/client/listers/vault/v1alpha1"
	crd_api "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	crd_cs "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/typed/apiextensions/v1beta1"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
)

type VaultController struct {
	config
	restConfig *rest.Config

	// ctxCancels stores vault clusters' contexts that are used to
	// cancel their goroutines when they are deleted
	ctxCancels map[string]context.CancelFunc

	kubeClient kubernetes.Interface
	extClient  cs.Interface
	crdClient  crd_cs.ApiextensionsV1beta1Interface
	recorder   record.EventRecorder

	kubeInformerFactory informers.SharedInformerFactory
	extInformerFactory  vaultinformers.SharedInformerFactory

	vaultClient *vaultapi.Client
	renewer     *time.Ticker

	saQueue    *queue.Worker
	saInformer cache.SharedIndexInformer

	sQueue    *queue.Worker
	sInformer cache.SharedIndexInformer

	vsQueue    *queue.Worker
	vsInformer cache.SharedIndexInformer
	vsLister   vault_listers.VaultServerLister
}

func (c *VaultController) ensureCustomResourceDefinitions() error {
	crds := []*crd_api.CustomResourceDefinition{
		api.VaultServer{}.CustomResourceDefinition(),
	}
	return crdutils.RegisterCRDs(c.crdClient, crds)
}

func (c *VaultController) initVault() (err error) {
	// TODO: unseal vault

	c.renewer = time.NewTicker(c.TokenRenewPeriod)

	c.vaultClient, err = vaultapi.NewClient(vaultapi.DefaultConfig())
	if err != nil {
		return
	}
	//var approleLoginPathRegex = regexp.MustCompile(`auth/.+/login`)
	//c.vaultClient.SetWrappingLookupFunc(func(operation, path string) string {
	//	if (operation == "PUT" || operation == "POST") &&
	//		(path == "sys/wrapping/wrap" || approleLoginPathRegex.MatchString(path)) {
	//		return stringz.Val(os.Getenv(api.EnvVaultWrapTTL), api.DefaultWrappingTTL)
	//	}
	//	return ""
	//})

	err = c.mountSecretBackend()
	if err != nil {
		return
	}
	err = c.mountAuthBackend()
	if err != nil {
		return
	}
	return
}

func (c *VaultController) RunInformers(stopCh <-chan struct{}) {
	defer runtime.HandleCrash()

	// TODO : uncomment later
	// defer c.renewer.Stop()

	glog.Info("Starting Vault controller")

	// TODO : uncomment later
	// c.kubeInformerFactory.Start(stopCh)
	c.extInformerFactory.Start(stopCh)

	// TODO : uncomment later
	// Wait for all involved caches to be synced, before processing items from the queue is started
	//for _, v := range c.kubeInformerFactory.WaitForCacheSync(stopCh) {
	//	if !v {
	//		runtime.HandleError(fmt.Errorf("timed out waiting for caches to sync"))
	//		return
	//	}
	//}
	for _, v := range c.extInformerFactory.WaitForCacheSync(stopCh) {
		if !v {
			runtime.HandleError(fmt.Errorf("timed out waiting for caches to sync"))
			return
		}
	}

	// TODO : uncomment later
	// c.saQueue.Run(stopCh)
	// c.sQueue.Run(stopCh)
	c.vsQueue.Run(stopCh)

	// TODO : uncomment later
	//go c.renewTokens()

	<-stopCh
	glog.Info("Stopping Vault operator")
}
