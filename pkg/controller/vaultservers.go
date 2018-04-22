package controller

import (
	"fmt"

	"github.com/appscode/kubernetes-webhook-util/admission"
	hooks "github.com/appscode/kubernetes-webhook-util/admission/v1beta1"
	webhook "github.com/appscode/kubernetes-webhook-util/admission/v1beta1/generic"
	"github.com/appscode/kutil/tools/queue"
	"github.com/golang/glog"
	"github.com/soter/vault-operator/apis/vault"
	api "github.com/soter/vault-operator/apis/vault/v1alpha1"
	//apps "k8s.io/api/apps/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/tools/cache"
)

func (c *VaultController) NewVaultServerWebhook() hooks.AdmissionHook {
	return webhook.NewGenericWebhook(
		schema.GroupVersionResource{
			Group:    "admission.vault.soter.ac",
			Version:  "v1alpha1",
			Resource: "vaultservers",
		},
		"vaultserver",
		[]string{vault.GroupName},
		api.SchemeGroupVersion.WithKind("VaultServer"),
		nil,
		&admission.ResourceHandlerFuncs{
			CreateFunc: func(obj runtime.Object) (runtime.Object, error) {
				return nil, obj.(*api.VaultServer).IsValid()
			},
			UpdateFunc: func(oldObj, newObj runtime.Object) (runtime.Object, error) {
				return nil, newObj.(*api.VaultServer).IsValid()
			},
		},
	)
}

func (c *VaultController) initVaultServerWatcher() {
	c.vsInformer = c.extInformerFactory.Vault().V1alpha1().VaultServers().Informer()
	c.vsQueue = queue.New("VaultServer", c.MaxNumRequeues, c.NumThreads, c.runVaultServerInjector)
	c.vsInformer.AddEventHandler(queue.DefaultEventHandler(c.vsQueue.GetQueue()))
	c.vsLister = c.extInformerFactory.Vault().V1alpha1().VaultServers().Lister()
}

// syncToStdout is the business logic of the controller. In this controller it simply prints
// information about the deployment to stdout. In case an error happened, it has to simply return the error.
// The retry logic should not be part of the business logic.
func (c *VaultController) runVaultServerInjector(key string) error {
	obj, exists, err := c.vsInformer.GetIndexer().GetByKey(key)
	if err != nil {
		glog.Errorf("Fetching object with key %s from store failed with %v", key, err)
		return err
	}

	if !exists {
		// Below we will warm up our cache with a VaultServer, so that we will see a delete for one d
		glog.Warningf("VaultServer %s does not exist anymore\n", key)

		namespace, name, err := cache.SplitMetaNamespaceKey(key)
		if err != nil {
			return err
		}
		fmt.Println(namespace, name)
	} else {
		restic := obj.(*api.VaultServer)
		glog.Infof("Sync/Add/Update for VaultServer %s\n", restic.GetName())

		fmt.Println(restic.Namespace, restic.Name)
	}
	return nil
}
