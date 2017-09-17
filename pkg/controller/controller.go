package controller

import (
	"fmt"
	"sync"
	"time"

	"github.com/appscode/steward/pkg/eventer"
	"github.com/golang/glog"
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
	client  kubernetes.Interface
	options Options

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
		client:   client,
		options:  opt,
		recorder: eventer.NewEventRecorder(client, "vault-controller"),
	}
	vc.initPodWatcher()
	vc.initServiceAccountWatcher()
	vc.initSecretWatcher()
	return vc
}

func (vc *VaultController) initPodWatcher() {
	lw := &cache.ListWatch{
		ListFunc: func(options metav1.ListOptions) (rt.Object, error) {
			options.IncludeUninitialized = true
			return vc.client.CoreV1().Pods(v1.NamespaceAll).List(options)
		},
		WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
			options.IncludeUninitialized = true
			return vc.client.CoreV1().Pods(v1.NamespaceAll).Watch(options)
		},
	}

	// create the workqueue
	vc.podQueue = workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "pod")

	// Bind the workqueue to a cache with the help of an informer. This way we make sure that
	// whenever the cache is updated, the pod key is added to the workqueue.
	// Note that when we finally process the item from the workqueue, we might see a newer version
	// of the Pod than the version which was responsible for triggering the update.
	vc.podIndexer, vc.podInformer = cache.NewIndexerInformer(lw, &v1.Pod{}, vc.options.ResyncPeriod, cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(obj)
			if err == nil {
				vc.podQueue.Add(key)
			}
		},
		UpdateFunc: func(old interface{}, new interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(new)
			if err == nil {
				vc.podQueue.Add(key)
			}
		},
		DeleteFunc: func(obj interface{}) {
			// IndexerInformer uses a delta queue, therefore for deletes we have to use this
			// key function.
			key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
			if err == nil {
				vc.podQueue.Add(key)
			}
		},
	}, cache.Indexers{})
}

func (vc *VaultController) initServiceAccountWatcher() {
	lw := &cache.ListWatch{
		ListFunc: func(options metav1.ListOptions) (rt.Object, error) {
			return vc.client.CoreV1().ServiceAccounts(v1.NamespaceAll).List(options)
		},
		WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
			return vc.client.CoreV1().ServiceAccounts(v1.NamespaceAll).Watch(options)
		},
	}

	// create the workqueue
	vc.saQueue = workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "serviceAccount")

	// Bind the workqueue to a cache with the help of an informer. This way we make sure that
	// whenever the cache is updated, the pod key is added to the workqueue.
	// Note that when we finally process the item from the workqueue, we might see a newer version
	// of the ServiceAccount than the version which was responsible for triggering the update.
	vc.saIndexer, vc.saInformer = cache.NewIndexerInformer(lw, &v1.ServiceAccount{}, vc.options.ResyncPeriod, cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(obj)
			if err == nil {
				vc.saQueue.Add(key)
			}
		},
		UpdateFunc: func(old interface{}, new interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(new)
			if err == nil {
				vc.saQueue.Add(key)
			}
		},
		DeleteFunc: func(obj interface{}) {
			// IndexerInformer uses a delta queue, therefore for deletes we have to use this
			// key function.
			key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
			if err == nil {
				vc.saQueue.Add(key)
			}
		},
	}, cache.Indexers{})
}

func (vc *VaultController) initSecretWatcher() {
	lw := &cache.ListWatch{
		ListFunc: func(options metav1.ListOptions) (rt.Object, error) {
			return vc.client.CoreV1().Secrets(v1.NamespaceAll).List(options)
		},
		WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
			return vc.client.CoreV1().Secrets(v1.NamespaceAll).Watch(options)
		},
	}

	// create the workqueue
	vc.sQueue = workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "secret")

	// Bind the workqueue to a cache with the help of an informer. This way we make sure that
	// whenever the cache is updated, the pod key is added to the workqueue.
	// Note that when we finally process the item from the workqueue, we might see a newer version
	// of the Secret than the version which was responsible for triggering the update.
	vc.sIndexer, vc.sInformer = cache.NewIndexerInformer(lw, &v1.Secret{}, vc.options.ResyncPeriod, cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(obj)
			if err == nil {
				vc.sQueue.Add(key)
			}
		},
		UpdateFunc: func(old interface{}, new interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(new)
			if err == nil {
				vc.sQueue.Add(key)
			}
		},
		DeleteFunc: func(obj interface{}) {
			// IndexerInformer uses a delta queue, therefore for deletes we have to use this
			// key function.
			key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
			if err == nil {
				vc.sQueue.Add(key)
			}
		},
	}, cache.Indexers{})
}

func (vc *VaultController) Run(threadiness int, stopCh chan struct{}) {
	defer runtime.HandleCrash()

	// Let the workers stop when we are done
	defer vc.podQueue.ShutDown()
	defer vc.saQueue.ShutDown()
	defer vc.sQueue.ShutDown()
	glog.Info("Starting Vault controller")

	go vc.podInformer.Run(stopCh)
	go vc.saInformer.Run(stopCh)
	go vc.sInformer.Run(stopCh)

	// Wait for all involved caches to be synced, before processing items from the queue is started
	if !cache.WaitForCacheSync(stopCh, vc.podInformer.HasSynced) {
		runtime.HandleError(fmt.Errorf("timed out waiting for caches to sync"))
		return
	}
	if !cache.WaitForCacheSync(stopCh, vc.saInformer.HasSynced) {
		runtime.HandleError(fmt.Errorf("timed out waiting for caches to sync"))
		return
	}
	if !cache.WaitForCacheSync(stopCh, vc.sInformer.HasSynced) {
		runtime.HandleError(fmt.Errorf("timed out waiting for caches to sync"))
		return
	}

	for i := 0; i < threadiness; i++ {
		go wait.Until(vc.runPodWatcher, time.Second, stopCh)
		go wait.Until(vc.runServiceAccountWatcher, time.Second, stopCh)
		go wait.Until(vc.runSecretWatcher, time.Second, stopCh)
	}

	<-stopCh
	glog.Info("Stopping Vault controller")
}
