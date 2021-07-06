/*
Copyright AppsCode Inc. and Contributors

Licensed under the AppsCode Community License 1.0.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    https://github.com/appscode/licenses/raw/1.0.0/AppsCode-Community-1.0.0.md

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"fmt"

	catalogapi "kubevault.dev/apimachinery/apis/catalog/v1alpha1"
	engineapi "kubevault.dev/apimachinery/apis/engine/v1alpha1"
	vaultapi "kubevault.dev/apimachinery/apis/kubevault/v1alpha1"
	policyapi "kubevault.dev/apimachinery/apis/policy/v1alpha1"
	cs "kubevault.dev/apimachinery/client/clientset/versioned"
	vaultinformers "kubevault.dev/apimachinery/client/informers/externalversions"
	engine_listers "kubevault.dev/apimachinery/client/listers/engine/v1alpha1"
	vault_listers "kubevault.dev/apimachinery/client/listers/kubevault/v1alpha1"
	policy_listers "kubevault.dev/apimachinery/client/listers/policy/v1alpha1"

	pcm "github.com/prometheus-operator/prometheus-operator/pkg/client/versioned/typed/monitoring/v1"
	auditlib "go.bytebuilders.dev/audit/lib"
	crd_cs "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	appslister "k8s.io/client-go/listers/apps/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog/v2"
	reg_util "kmodules.xyz/client-go/admissionregistration/v1beta1"
	"kmodules.xyz/client-go/apiextensions"
	"kmodules.xyz/client-go/tools/queue"
	appcat "kmodules.xyz/custom-resources/apis/appcatalog/v1alpha1"
	appcat_cs "kmodules.xyz/custom-resources/client/clientset/versioned/typed/appcatalog/v1alpha1"
)

type VaultController struct {
	config
	clientConfig *rest.Config

	// ctxCancels stores vault clusters' contexts that are used to
	// cancel their goroutines when they are deleted
	ctxCancels map[string]CtxWithCancel

	// Todo: Dynamic client
	dynamicClient dynamic.Interface

	kubeClient       kubernetes.Interface
	extClient        cs.Interface
	appCatalogClient appcat_cs.AppcatalogV1alpha1Interface
	crdClient        crd_cs.Interface
	recorder         record.EventRecorder
	auditor          *auditlib.EventPublisher
	promClient       pcm.MonitoringV1Interface

	kubeInformerFactory informers.SharedInformerFactory
	extInformerFactory  vaultinformers.SharedInformerFactory

	// for VaultServer
	vsQueue    *queue.Worker
	vsInformer cache.SharedIndexInformer
	vsLister   vault_listers.VaultServerLister

	// Todo: StatefulSet Watcher
	StsQueue    *queue.Worker
	StsInformer cache.SharedIndexInformer
	StsLister   appslister.StatefulSetLister

	// for VaultPolicy
	vplcyQueue    *queue.Worker
	vplcyInformer cache.SharedIndexInformer
	vplcyLister   policy_listers.VaultPolicyLister

	// for VaultPolicyBinding
	vplcyBindingQueue    *queue.Worker
	vplcyBindingInformer cache.SharedIndexInformer
	vplcyBindingLister   policy_listers.VaultPolicyBindingLister

	// PostgresRole
	pgRoleQueue    *queue.Worker
	pgRoleInformer cache.SharedIndexInformer
	pgRoleLister   engine_listers.PostgresRoleLister

	// MySQLRole
	myRoleQueue    *queue.Worker
	myRoleInformer cache.SharedIndexInformer
	myRoleLister   engine_listers.MySQLRoleLister

	// MongoDBRole
	mgRoleQueue    *queue.Worker
	mgRoleInformer cache.SharedIndexInformer
	mgRoleLister   engine_listers.MongoDBRoleLister

	// ElasticsearchRole
	esRoleQueue    *queue.Worker
	esRoleInformer cache.SharedIndexInformer
	esRoleLister   engine_listers.ElasticsearchRoleLister

	// AWSRole
	awsRoleQueue    *queue.Worker
	awsRoleInformer cache.SharedIndexInformer
	awsRoleLister   engine_listers.AWSRoleLister

	// DatabaseAccessRequest
	dbAccessQueue    *queue.Worker
	dbAccessInformer cache.SharedIndexInformer
	dbAccessLister   engine_listers.DatabaseAccessRequestLister

	// AWSAccessKeyRequest
	awsAccessQueue    *queue.Worker
	awsAccessInformer cache.SharedIndexInformer
	awsAccessLister   engine_listers.AWSAccessKeyRequestLister

	// GCPRole
	gcpRoleQueue    *queue.Worker
	gcpRoleInformer cache.SharedIndexInformer
	gcpRoleLister   engine_listers.GCPRoleLister

	// GCPAccessKeyRequest
	gcpAccessQueue    *queue.Worker
	gcpAccessInformer cache.SharedIndexInformer
	gcpAccessLister   engine_listers.GCPAccessKeyRequestLister

	// AzureRole
	azureRoleQueue    *queue.Worker
	azureRoleInformer cache.SharedIndexInformer
	azureRoleLister   engine_listers.AzureRoleLister

	// AzureAccessKeyRequest
	azureAccessQueue    *queue.Worker
	azureAccessInformer cache.SharedIndexInformer
	azureAccessLister   engine_listers.AzureAccessKeyRequestLister

	// SecretEngine
	secretEngineQueue    *queue.Worker
	secretEngineInformer cache.SharedIndexInformer
	secretEngineLister   engine_listers.SecretEngineLister

	// Contain the currently processing finalizer
	finalizerInfo *mapFinalizer

	// authMethodCtx stores auth method controller contexts that are used to
	// cancel their goroutines when they are not needed
	authMethodCtx map[string]CtxWithCancel
}

func (c *VaultController) ensureCustomResourceDefinitions() error {
	crds := []*apiextensions.CustomResourceDefinition{
		vaultapi.VaultServer{}.CustomResourceDefinition(),
		catalogapi.VaultServerVersion{}.CustomResourceDefinition(),
		policyapi.VaultPolicy{}.CustomResourceDefinition(),
		policyapi.VaultPolicyBinding{}.CustomResourceDefinition(),
		appcat.AppBinding{}.CustomResourceDefinition(),
		engineapi.AWSAccessKeyRequest{}.CustomResourceDefinition(),
		engineapi.AWSRole{}.CustomResourceDefinition(),
		engineapi.AzureAccessKeyRequest{}.CustomResourceDefinition(),
		engineapi.AzureRole{}.CustomResourceDefinition(),
		engineapi.DatabaseAccessRequest{}.CustomResourceDefinition(),
		engineapi.GCPAccessKeyRequest{}.CustomResourceDefinition(),
		engineapi.GCPRole{}.CustomResourceDefinition(),
		engineapi.MongoDBRole{}.CustomResourceDefinition(),
		engineapi.MySQLRole{}.CustomResourceDefinition(),
		engineapi.PostgresRole{}.CustomResourceDefinition(),
		engineapi.ElasticsearchRole{}.CustomResourceDefinition(),
		engineapi.SecretEngine{}.CustomResourceDefinition(),
	}
	return apiextensions.RegisterCRDs(c.crdClient, crds)
}

func (c *VaultController) Run(stopCh <-chan struct{}) {
	go c.RunInformers(stopCh)

	if c.EnableMutatingWebhook {
		cancel, _ := reg_util.SyncMutatingWebhookCABundle(c.clientConfig, mutatingWebhook)
		defer cancel()
	}
	if c.EnableValidatingWebhook {
		cancel, _ := reg_util.SyncValidatingWebhookCABundle(c.clientConfig, validatingWebhook)
		defer cancel()
	}

	<-stopCh
}

func (c *VaultController) RunInformers(stopCh <-chan struct{}) {
	defer runtime.HandleCrash()

	klog.Info("Starting Vault controller")

	c.extInformerFactory.Start(stopCh)
	// Todo: StatefulSet Informer
	c.kubeInformerFactory.Start(stopCh)
	// Todo: Health Checker
	c.RunHealthChecker(stopCh)
	for _, v := range c.extInformerFactory.WaitForCacheSync(stopCh) {
		if !v {
			runtime.HandleError(fmt.Errorf("timed out waiting for caches to sync"))
			return
		}
	}

	// Todo: For StatefulSet
	go c.StsQueue.Run(stopCh)

	// For VaultServer
	go c.vsQueue.Run(stopCh)

	//For VaultPolicy
	go c.vplcyQueue.Run(stopCh)

	//For VaultPolicyBinding
	go c.vplcyBindingQueue.Run(stopCh)

	// For DB role
	go c.pgRoleQueue.Run(stopCh)
	go c.myRoleQueue.Run(stopCh)
	go c.mgRoleQueue.Run(stopCh)
	go c.esRoleQueue.Run(stopCh)

	// For AWSRole
	go c.awsRoleQueue.Run(stopCh)

	// For DB access request
	go c.dbAccessQueue.Run(stopCh)

	// For AWS access key request
	go c.awsAccessQueue.Run(stopCh)

	// For GCPRole
	go c.gcpRoleQueue.Run(stopCh)

	// For GCP access key request
	go c.gcpAccessQueue.Run(stopCh)

	// For AzureRole
	go c.azureRoleQueue.Run(stopCh)

	// For Azure access key request
	go c.azureAccessQueue.Run(stopCh)

	// For Secret Engine
	go c.secretEngineQueue.Run(stopCh)

	<-stopCh
	klog.Info("Stopping Vault operator")
}
