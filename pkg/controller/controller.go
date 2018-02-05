package controller

import (
	"fmt"
	"sync"
	"time"

	"github.com/appscode/kutil/tools/queue"
	"github.com/appscode/steward/pkg/eventer"
	"github.com/golang/glog"
	"github.com/hashicorp/vault/api"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
)

type VaultController struct {
	k8sClient       kubernetes.Interface
	informerFactory informers.SharedInformerFactory
	options         Options

	vaultClient *api.Client
	renewer     *time.Ticker

	saQueue    *queue.Worker
	saInformer cache.SharedIndexInformer

	sQueue    *queue.Worker
	sInformer cache.SharedIndexInformer

	dpQueue    *queue.Worker
	dpInformer cache.SharedIndexInformer

	dsQueue    *queue.Worker
	dsInformer cache.SharedIndexInformer

	jobQueue    *queue.Worker
	jobInformer cache.SharedIndexInformer

	rcQueue    *queue.Worker
	rcInformer cache.SharedIndexInformer

	rsQueue    *queue.Worker
	rsInformer cache.SharedIndexInformer

	ssQueue    *queue.Worker
	ssInformer cache.SharedIndexInformer

	recorder record.EventRecorder
	sync.Mutex
}

func New(client kubernetes.Interface, opt Options) *VaultController {
	tweakListOptions := func(opt *metav1.ListOptions) {
		opt.IncludeUninitialized = true
	}
	vc := &VaultController{
		k8sClient:       client,
		informerFactory: informers.NewFilteredSharedInformerFactory(client, opt.ResyncPeriod, core.NamespaceAll, tweakListOptions),
		options:         opt,
		recorder:        eventer.NewEventRecorder(client, "vault-controller"),
	}
	vc.initVault()
	vc.initServiceAccountWatcher()
	vc.initSecretWatcher()
	vc.initDaemonSetWatcher()
	vc.initDeploymentWatcher()
	vc.initJobWatcher()
	vc.initRCWatcher()
	vc.initReplicaSetWatcher()
	vc.initStatefulSetWatcher()
	return vc
}

func (c *VaultController) initVault() (err error) {
	// TODO: unseal vault

	c.renewer = time.NewTicker(c.options.TokenRenewPeriod)

	c.vaultClient, err = api.NewClient(api.DefaultConfig())
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

func (c *VaultController) Run(stopCh chan struct{}) {
	defer runtime.HandleCrash()

	defer c.renewer.Stop()

	glog.Info("Starting Vault controller")
	c.informerFactory.Start(stopCh)

	// Wait for all involved caches to be synced, before processing items from the queue is started
	for _, v := range c.informerFactory.WaitForCacheSync(stopCh) {
		if !v {
			runtime.HandleError(fmt.Errorf("timed out waiting for caches to sync"))
			return
		}
	}

	c.saQueue.Run(stopCh)
	c.sQueue.Run(stopCh)
	c.dsQueue.Run(stopCh)
	c.dpQueue.Run(stopCh)
	c.jobQueue.Run(stopCh)
	c.rcQueue.Run(stopCh)
	c.rsQueue.Run(stopCh)
	c.ssQueue.Run(stopCh)

	go c.renewTokens()

	<-stopCh
	glog.Info("Stopping Vault controller")
}
