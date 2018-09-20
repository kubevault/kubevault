package controller

import (
	"context"
	"fmt"

	crdutils "github.com/appscode/kutil/apiextensions/v1beta1"
	"github.com/appscode/kutil/tools/queue"
	"github.com/golang/glog"
	api "github.com/kubevault/operator/apis/kubevault/v1alpha1"
	cs "github.com/kubevault/operator/client/clientset/versioned"
	vaultinformers "github.com/kubevault/operator/client/informers/externalversions"
	vault_listers "github.com/kubevault/operator/client/listers/kubevault/v1alpha1"
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

	vsQueue    *queue.Worker
	vsInformer cache.SharedIndexInformer
	vsLister   vault_listers.VaultServerLister
}

func (c *VaultController) ensureCustomResourceDefinitions() error {
	crds := []*crd_api.CustomResourceDefinition{
		api.VaultServer{}.CustomResourceDefinition(),
		api.VaultServerVersion{}.CustomResourceDefinition(),
	}
	return crdutils.RegisterCRDs(c.crdClient, crds)
}

func (c *VaultController) RunInformers(stopCh <-chan struct{}) {
	defer runtime.HandleCrash()

	glog.Info("Starting Vault controller")

	c.extInformerFactory.Start(stopCh)

	for _, v := range c.extInformerFactory.WaitForCacheSync(stopCh) {
		if !v {
			runtime.HandleError(fmt.Errorf("timed out waiting for caches to sync"))
			return
		}
	}

	c.vsQueue.Run(stopCh)

	<-stopCh
	glog.Info("Stopping Vault operator")
}
