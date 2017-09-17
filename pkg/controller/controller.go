package controller

import (
	"fmt"
	"os"
	"sync"
	"time"

	stringz "github.com/appscode/go/strings"
	"github.com/appscode/steward/pkg/eventer"
	"github.com/golang/glog"
	"github.com/hashicorp/vault/api"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	rt "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
)

type VaultController struct {
	k8sClient kubernetes.Interface
	options   Options

	vaultClient *api.Client

	podQueue    workqueue.RateLimitingInterface
	podIndexer  cache.Indexer
	podInformer cache.Controller

	saQueue    workqueue.RateLimitingInterface
	saIndexer  cache.Indexer
	saInformer cache.Controller

	sQueue    workqueue.RateLimitingInterface
	sIndexer  cache.Indexer
	sInformer cache.Controller

	recorder record.EventRecorder
	sync.Mutex
}

func New(client kubernetes.Interface, opt Options) *VaultController {
	vc := &VaultController{
		k8sClient: client,
		options:   opt,
		recorder:  eventer.NewEventRecorder(client, "vault-controller"),
	}
	vc.initPodWatcher()
	vc.initServiceAccountWatcher()
	vc.initSecretWatcher()
	vc.initVault()
	return vc
}

func (c *VaultController) initPodWatcher() {
	lw := &cache.ListWatch{
		ListFunc: func(options metav1.ListOptions) (rt.Object, error) {
			options.IncludeUninitialized = true
			return c.k8sClient.CoreV1().Pods(v1.NamespaceAll).List(options)
		},
		WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
			options.IncludeUninitialized = true
			return c.k8sClient.CoreV1().Pods(v1.NamespaceAll).Watch(options)
		},
	}

	// create the workqueue
	c.podQueue = workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "pod")

	// Bind the workqueue to a cache with the help of an informer. This way we make sure that
	// whenever the cache is updated, the pod key is added to the workqueue.
	// Note that when we finally process the item from the workqueue, we might see a newer version
	// of the Pod than the version which was responsible for triggering the update.
	c.podIndexer, c.podInformer = cache.NewIndexerInformer(lw, &v1.Pod{}, c.options.ResyncPeriod, cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(obj)
			if err == nil {
				c.podQueue.Add(key)
			}
		},
		UpdateFunc: func(old interface{}, new interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(new)
			if err == nil {
				c.podQueue.Add(key)
			}
		},
		DeleteFunc: func(obj interface{}) {
			// IndexerInformer uses a delta queue, therefore for deletes we have to use this
			// key function.
			key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
			if err == nil {
				c.podQueue.Add(key)
			}
		},
	}, cache.Indexers{})
}

func (c *VaultController) initServiceAccountWatcher() {
	lw := &cache.ListWatch{
		ListFunc: func(options metav1.ListOptions) (rt.Object, error) {
			return c.k8sClient.CoreV1().ServiceAccounts(v1.NamespaceAll).List(options)
		},
		WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
			return c.k8sClient.CoreV1().ServiceAccounts(v1.NamespaceAll).Watch(options)
		},
	}

	// create the workqueue
	c.saQueue = workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "serviceAccount")

	// Bind the workqueue to a cache with the help of an informer. This way we make sure that
	// whenever the cache is updated, the pod key is added to the workqueue.
	// Note that when we finally process the item from the workqueue, we might see a newer version
	// of the ServiceAccount than the version which was responsible for triggering the update.
	c.saIndexer, c.saInformer = cache.NewIndexerInformer(lw, &v1.ServiceAccount{}, c.options.ResyncPeriod, cache.ResourceEventHandlerFuncs{
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
			return c.k8sClient.CoreV1().Secrets(v1.NamespaceAll).List(options)
		},
		WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
			return c.k8sClient.CoreV1().Secrets(v1.NamespaceAll).Watch(options)
		},
	}

	// create the workqueue
	c.sQueue = workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "secret")

	// Bind the workqueue to a cache with the help of an informer. This way we make sure that
	// whenever the cache is updated, the pod key is added to the workqueue.
	// Note that when we finally process the item from the workqueue, we might see a newer version
	// of the Secret than the version which was responsible for triggering the update.
	c.sIndexer, c.sInformer = cache.NewIndexerInformer(lw, &v1.Secret{}, c.options.ResyncPeriod, cache.ResourceEventHandlerFuncs{
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

func (c *VaultController) initVault() (err error) {
	// TODO: unseal vault

	c.vaultClient, err = api.NewClient(api.DefaultConfig())
	if err != nil {
		return
	}
	c.vaultClient.SetWrappingLookupFunc(func(operation, path string) string {
		if (operation == "PUT" || operation == "POST") &&
			(path == "sys/wrapping/wrap" || path == "auth/approle/login") {
			return stringz.Val(os.Getenv(api.EnvVaultWrapTTL), api.DefaultWrappingTTL)
		}
		return ""
	})

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

func (c *VaultController) Run(threadiness int, stopCh chan struct{}) {
	defer runtime.HandleCrash()

	// Let the workers stop when we are done
	defer c.podQueue.ShutDown()
	defer c.saQueue.ShutDown()
	defer c.sQueue.ShutDown()
	glog.Info("Starting Vault controller")

	go c.podInformer.Run(stopCh)
	go c.saInformer.Run(stopCh)
	go c.sInformer.Run(stopCh)

	// Wait for all involved caches to be synced, before processing items from the queue is started
	if !cache.WaitForCacheSync(stopCh, c.podInformer.HasSynced) {
		runtime.HandleError(fmt.Errorf("timed out waiting for caches to sync"))
		return
	}
	if !cache.WaitForCacheSync(stopCh, c.saInformer.HasSynced) {
		runtime.HandleError(fmt.Errorf("timed out waiting for caches to sync"))
		return
	}
	if !cache.WaitForCacheSync(stopCh, c.sInformer.HasSynced) {
		runtime.HandleError(fmt.Errorf("timed out waiting for caches to sync"))
		return
	}

	for i := 0; i < threadiness; i++ {
		go wait.Until(c.runPodWatcher, time.Second, stopCh)
		go wait.Until(c.runServiceAccountWatcher, time.Second, stopCh)
		go wait.Until(c.runSecretWatcher, time.Second, stopCh)
	}

	<-stopCh
	glog.Info("Stopping Vault controller")
}
