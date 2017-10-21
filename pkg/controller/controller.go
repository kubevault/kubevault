package controller

import (
	"fmt"
	"sync"
	"time"

	"github.com/appscode/steward/pkg/eventer"
	"github.com/golang/glog"
	"github.com/hashicorp/vault/api"
	apps "k8s.io/api/apps/v1beta1"
	batch "k8s.io/api/batch/v1"
	core "k8s.io/api/core/v1"
	extensions "k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	rt "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
)

type VaultController struct {
	k8sClient kubernetes.Interface
	options   Options

	vaultClient *api.Client
	renewer     *time.Ticker

	saQueue    workqueue.RateLimitingInterface
	saIndexer  cache.Indexer
	saInformer cache.Controller

	sQueue    workqueue.RateLimitingInterface
	sIndexer  cache.Indexer
	sInformer cache.Controller

	dpQueue    workqueue.RateLimitingInterface
	dpIndexer  cache.Indexer
	dpInformer cache.Controller

	dsQueue    workqueue.RateLimitingInterface
	dsIndexer  cache.Indexer
	dsInformer cache.Controller

	jQueue    workqueue.RateLimitingInterface
	jIndexer  cache.Indexer
	jInformer cache.Controller

	rcQueue    workqueue.RateLimitingInterface
	rcIndexer  cache.Indexer
	rcInformer cache.Controller

	rsQueue    workqueue.RateLimitingInterface
	rsIndexer  cache.Indexer
	rsInformer cache.Controller

	ssQueue    workqueue.RateLimitingInterface
	ssIndexer  cache.Indexer
	ssInformer cache.Controller

	recorder record.EventRecorder
	sync.Mutex
}

func New(client kubernetes.Interface, opt Options) *VaultController {
	vc := &VaultController{
		k8sClient: client,
		options:   opt,
		recorder:  eventer.NewEventRecorder(client, "vault-controller"),
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

func (c *VaultController) initServiceAccountWatcher() {
	lw := &cache.ListWatch{
		ListFunc: func(options metav1.ListOptions) (rt.Object, error) {
			return c.k8sClient.CoreV1().ServiceAccounts(core.NamespaceAll).List(options)
		},
		WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
			return c.k8sClient.CoreV1().ServiceAccounts(core.NamespaceAll).Watch(options)
		},
	}

	// create the workqueue
	c.saQueue = workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "serviceAccount")

	// Bind the workqueue to a cache with the help of an informer. This way we make sure that
	// whenever the cache is updated, the pod key is added to the workqueue.
	// Note that when we finally process the item from the workqueue, we might see a newer version
	// of the ServiceAccount than the version which was responsible for triggering the update.
	c.saIndexer, c.saInformer = cache.NewIndexerInformer(lw, &core.ServiceAccount{}, c.options.ResyncPeriod, cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(obj)
			if err == nil {
				c.saQueue.Add(key)
			}
		},
		UpdateFunc: func(old interface{}, new interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(new)
			if err == nil {
				c.saQueue.Add(key)
			}
		},
		DeleteFunc: func(obj interface{}) {
			// IndexerInformer uses a delta queue, therefore for deletes we have to use this
			// key function.
			key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
			if err == nil {
				c.saQueue.Add(key)
			}
		},
	}, cache.Indexers{})
}

func (c *VaultController) initSecretWatcher() {
	lw := &cache.ListWatch{
		ListFunc: func(options metav1.ListOptions) (rt.Object, error) {
			return c.k8sClient.CoreV1().Secrets(core.NamespaceAll).List(options)
		},
		WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
			return c.k8sClient.CoreV1().Secrets(core.NamespaceAll).Watch(options)
		},
	}

	// create the workqueue
	c.sQueue = workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "secret")

	// Bind the workqueue to a cache with the help of an informer. This way we make sure that
	// whenever the cache is updated, the pod key is added to the workqueue.
	// Note that when we finally process the item from the workqueue, we might see a newer version
	// of the Secret than the version which was responsible for triggering the update.
	c.sIndexer, c.sInformer = cache.NewIndexerInformer(lw, &core.Secret{}, c.options.ResyncPeriod, cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(obj)
			if err == nil {
				c.sQueue.Add(key)
			}
		},
		UpdateFunc: func(old interface{}, new interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(new)
			if err == nil {
				c.sQueue.Add(key)
			}
		},
		DeleteFunc: func(obj interface{}) {
			// IndexerInformer uses a delta queue, therefore for deletes we have to use this
			// key function.
			key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
			if err == nil {
				c.sQueue.Add(key)
			}
		},
	}, cache.Indexers{})
}

func (c *VaultController) initDaemonSetWatcher() {
	lw := &cache.ListWatch{
		ListFunc: func(options metav1.ListOptions) (rt.Object, error) {
			options.IncludeUninitialized = true
			return c.k8sClient.ExtensionsV1beta1().DaemonSets(core.NamespaceAll).List(options)
		},
		WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
			options.IncludeUninitialized = true
			return c.k8sClient.ExtensionsV1beta1().DaemonSets(core.NamespaceAll).Watch(options)
		},
	}

	// create the workqueue
	c.dsQueue = workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "daemonset")

	// Bind the workqueue to a cache with the help of an informer. This way we make sure that
	// whenever the cache is updated, the pod key is added to the workqueue.
	// Note that when we finally process the item from the workqueue, we might see a newer version
	// of the DaemonSet than the version which was responsible for triggering the update.
	c.dsIndexer, c.dsInformer = cache.NewIndexerInformer(lw, &extensions.DaemonSet{}, c.options.ResyncPeriod, cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(obj)
			if err == nil {
				c.dsQueue.Add(key)
			}
		},
		UpdateFunc: func(old interface{}, new interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(new)
			if err == nil {
				c.dsQueue.Add(key)
			}
		},
		DeleteFunc: func(obj interface{}) {
			// IndexerInformer uses a delta queue, therefore for deletes we have to use this
			// key function.
			key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
			if err == nil {
				c.dsQueue.Add(key)
			}
		},
	}, cache.Indexers{})
}

func (c *VaultController) initDeploymentWatcher() {
	lw := &cache.ListWatch{
		ListFunc: func(options metav1.ListOptions) (rt.Object, error) {
			options.IncludeUninitialized = true
			return c.k8sClient.AppsV1beta1().Deployments(core.NamespaceAll).List(options)
		},
		WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
			options.IncludeUninitialized = true
			return c.k8sClient.AppsV1beta1().Deployments(core.NamespaceAll).Watch(options)
		},
	}

	// create the workqueue
	c.dpQueue = workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "deployment")

	// Bind the workqueue to a cache with the help of an informer. This way we make sure that
	// whenever the cache is updated, the pod key is added to the workqueue.
	// Note that when we finally process the item from the workqueue, we might see a newer version
	// of the Deployment than the version which was responsible for triggering the update.
	c.dpIndexer, c.dpInformer = cache.NewIndexerInformer(lw, &apps.Deployment{}, c.options.ResyncPeriod, cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(obj)
			if err == nil {
				c.dpQueue.Add(key)
			}
		},
		UpdateFunc: func(old interface{}, new interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(new)
			if err == nil {
				c.dpQueue.Add(key)
			}
		},
		DeleteFunc: func(obj interface{}) {
			// IndexerInformer uses a delta queue, therefore for deletes we have to use this
			// key function.
			key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
			if err == nil {
				c.dpQueue.Add(key)
			}
		},
	}, cache.Indexers{})
}

func (c *VaultController) initJobWatcher() {
	lw := &cache.ListWatch{
		ListFunc: func(options metav1.ListOptions) (rt.Object, error) {
			options.IncludeUninitialized = true
			return c.k8sClient.BatchV1().Jobs(core.NamespaceAll).List(options)
		},
		WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
			options.IncludeUninitialized = true
			return c.k8sClient.BatchV1().Jobs(core.NamespaceAll).Watch(options)
		},
	}

	// create the workqueue
	c.jQueue = workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "job")

	// Bind the workqueue to a cache with the help of an informer. This way we make sure that
	// whenever the cache is updated, the pod key is added to the workqueue.
	// Note that when we finally process the item from the workqueue, we might see a newer version
	// of the Job than the version which was responsible for triggering the update.
	c.jIndexer, c.jInformer = cache.NewIndexerInformer(lw, &batch.Job{}, c.options.ResyncPeriod, cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(obj)
			if err == nil {
				c.jQueue.Add(key)
			}
		},
		UpdateFunc: func(old interface{}, new interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(new)
			if err == nil {
				c.jQueue.Add(key)
			}
		},
		DeleteFunc: func(obj interface{}) {
			// IndexerInformer uses a delta queue, therefore for deletes we have to use this
			// key function.
			key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
			if err == nil {
				c.jQueue.Add(key)
			}
		},
	}, cache.Indexers{})
}

func (c *VaultController) initRCWatcher() {
	lw := &cache.ListWatch{
		ListFunc: func(options metav1.ListOptions) (rt.Object, error) {
			options.IncludeUninitialized = true
			return c.k8sClient.CoreV1().ReplicationControllers(core.NamespaceAll).List(options)
		},
		WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
			options.IncludeUninitialized = true
			return c.k8sClient.CoreV1().ReplicationControllers(core.NamespaceAll).Watch(options)
		},
	}

	// create the workqueue
	c.rcQueue = workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "rc")

	// Bind the workqueue to a cache with the help of an informer. This way we make sure that
	// whenever the cache is updated, the pod key is added to the workqueue.
	// Note that when we finally process the item from the workqueue, we might see a newer version
	// of the ReplicationController than the version which was responsible for triggering the update.
	c.rcIndexer, c.rcInformer = cache.NewIndexerInformer(lw, &core.ReplicationController{}, c.options.ResyncPeriod, cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(obj)
			if err == nil {
				c.rcQueue.Add(key)
			}
		},
		UpdateFunc: func(old interface{}, new interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(new)
			if err == nil {
				c.rcQueue.Add(key)
			}
		},
		DeleteFunc: func(obj interface{}) {
			// IndexerInformer uses a delta queue, therefore for deletes we have to use this
			// key function.
			key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
			if err == nil {
				c.rcQueue.Add(key)
			}
		},
	}, cache.Indexers{})
}

func (c *VaultController) initReplicaSetWatcher() {
	lw := &cache.ListWatch{
		ListFunc: func(options metav1.ListOptions) (rt.Object, error) {
			options.IncludeUninitialized = true
			return c.k8sClient.ExtensionsV1beta1().ReplicaSets(core.NamespaceAll).List(options)
		},
		WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
			options.IncludeUninitialized = true
			return c.k8sClient.ExtensionsV1beta1().ReplicaSets(core.NamespaceAll).Watch(options)
		},
	}

	// create the workqueue
	c.rsQueue = workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "replicaset")

	// Bind the workqueue to a cache with the help of an informer. This way we make sure that
	// whenever the cache is updated, the pod key is added to the workqueue.
	// Note that when we finally process the item from the workqueue, we might see a newer version
	// of the ReplicaSet than the version which was responsible for triggering the update.
	c.rsIndexer, c.rsInformer = cache.NewIndexerInformer(lw, &extensions.ReplicaSet{}, c.options.ResyncPeriod, cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(obj)
			if err == nil {
				c.rsQueue.Add(key)
			}
		},
		UpdateFunc: func(old interface{}, new interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(new)
			if err == nil {
				c.rsQueue.Add(key)
			}
		},
		DeleteFunc: func(obj interface{}) {
			// IndexerInformer uses a delta queue, therefore for deletes we have to use this
			// key function.
			key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
			if err == nil {
				c.rsQueue.Add(key)
			}
		},
	}, cache.Indexers{})
}

func (c *VaultController) initStatefulSetWatcher() {
	lw := &cache.ListWatch{
		ListFunc: func(options metav1.ListOptions) (rt.Object, error) {
			options.IncludeUninitialized = true
			return c.k8sClient.AppsV1beta1().StatefulSets(core.NamespaceAll).List(options)
		},
		WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
			options.IncludeUninitialized = true
			return c.k8sClient.AppsV1beta1().StatefulSets(core.NamespaceAll).Watch(options)
		},
	}

	// create the workqueue
	c.ssQueue = workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "statefulset")

	// Bind the workqueue to a cache with the help of an informer. This way we make sure that
	// whenever the cache is updated, the pod key is added to the workqueue.
	// Note that when we finally process the item from the workqueue, we might see a newer version
	// of the StatefulSet than the version which was responsible for triggering the update.
	c.ssIndexer, c.ssInformer = cache.NewIndexerInformer(lw, &apps.StatefulSet{}, c.options.ResyncPeriod, cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(obj)
			if err == nil {
				c.ssQueue.Add(key)
			}
		},
		UpdateFunc: func(old interface{}, new interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(new)
			if err == nil {
				c.ssQueue.Add(key)
			}
		},
		DeleteFunc: func(obj interface{}) {
			// IndexerInformer uses a delta queue, therefore for deletes we have to use this
			// key function.
			key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
			if err == nil {
				c.ssQueue.Add(key)
			}
		},
	}, cache.Indexers{})
}

func (c *VaultController) Run(threadiness int, stopCh chan struct{}) {
	defer runtime.HandleCrash()

	// Let the workers stop when we are done
	defer c.saQueue.ShutDown()
	defer c.sQueue.ShutDown()
	defer c.dsQueue.ShutDown()
	defer c.dpQueue.ShutDown()
	defer c.jQueue.ShutDown()
	defer c.rcQueue.ShutDown()
	defer c.rsQueue.ShutDown()
	defer c.ssQueue.ShutDown()
	defer c.renewer.Stop()
	glog.Info("Starting Vault controller")

	go c.saInformer.Run(stopCh)
	go c.sInformer.Run(stopCh)
	go c.dsInformer.Run(stopCh)
	go c.dpInformer.Run(stopCh)
	go c.jInformer.Run(stopCh)
	go c.rcInformer.Run(stopCh)
	go c.rsInformer.Run(stopCh)
	go c.ssInformer.Run(stopCh)

	// Wait for all involved caches to be synced, before processing items from the queue is started
	if !cache.WaitForCacheSync(stopCh, c.saInformer.HasSynced) {
		runtime.HandleError(fmt.Errorf("timed out waiting for caches to sync"))
		return
	}
	if !cache.WaitForCacheSync(stopCh, c.sInformer.HasSynced) {
		runtime.HandleError(fmt.Errorf("timed out waiting for caches to sync"))
		return
	}
	if !cache.WaitForCacheSync(stopCh, c.dsInformer.HasSynced) {
		runtime.HandleError(fmt.Errorf("timed out waiting for caches to sync"))
		return
	}
	if !cache.WaitForCacheSync(stopCh, c.dpInformer.HasSynced) {
		runtime.HandleError(fmt.Errorf("timed out waiting for caches to sync"))
		return
	}
	if !cache.WaitForCacheSync(stopCh, c.jInformer.HasSynced) {
		runtime.HandleError(fmt.Errorf("timed out waiting for caches to sync"))
		return
	}
	if !cache.WaitForCacheSync(stopCh, c.rcInformer.HasSynced) {
		runtime.HandleError(fmt.Errorf("timed out waiting for caches to sync"))
		return
	}
	if !cache.WaitForCacheSync(stopCh, c.rsInformer.HasSynced) {
		runtime.HandleError(fmt.Errorf("timed out waiting for caches to sync"))
		return
	}
	if !cache.WaitForCacheSync(stopCh, c.ssInformer.HasSynced) {
		runtime.HandleError(fmt.Errorf("timed out waiting for caches to sync"))
		return
	}

	for i := 0; i < threadiness; i++ {
		go wait.Until(c.runServiceAccountWatcher, time.Second, stopCh)
		go wait.Until(c.runSecretWatcher, time.Second, stopCh)
		go wait.Until(c.runDaemonSetWatcher, time.Second, stopCh)
		go wait.Until(c.runDeploymentWatcher, time.Second, stopCh)
		go wait.Until(c.runJobWatcher, time.Second, stopCh)
		go wait.Until(c.runRCWatcher, time.Second, stopCh)
		go wait.Until(c.runReplicaSetWatcher, time.Second, stopCh)
		go wait.Until(c.runSecretWatcher, time.Second, stopCh)
	}
	go c.renewTokens()

	<-stopCh
	glog.Info("Stopping Vault controller")
}
