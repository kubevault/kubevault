package controller

import (
	"context"
	"time"

	reg_util "github.com/appscode/kutil/admissionregistration/v1beta1"
	cs "github.com/kubevault/operator/client/clientset/versioned"
	vaultinformers "github.com/kubevault/operator/client/informers/externalversions"
	"github.com/kubevault/operator/pkg/eventer"
	core "k8s.io/api/core/v1"
	crd_cs "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/typed/apiextensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

const (
	validatingWebhook = "admission.kubevault.com"
)

var (
	AnalyticsClientID string
	EnableAnalytics   = true
)

type config struct {
	EnableRBAC     bool
	DockerRegistry string
	MaxNumRequeues int
	NumThreads     int
	OpsAddress     string
	ResyncPeriod   time.Duration
}

type Config struct {
	config

	ClientConfig *rest.Config
	KubeClient   kubernetes.Interface
	ExtClient    cs.Interface
	CRDClient    crd_cs.ApiextensionsV1beta1Interface
}

func NewConfig(clientConfig *rest.Config) *Config {
	return &Config{
		ClientConfig: clientConfig,
	}
}

func (c *Config) New() (*VaultController, error) {
	tweakListOptions := func(opt *metav1.ListOptions) {
		opt.IncludeUninitialized = true
	}
	ctrl := &VaultController{
		config:              c.config,
		clientConfig:        c.ClientConfig,
		ctxCancels:          make(map[string]context.CancelFunc),
		kubeClient:          c.KubeClient,
		extClient:           c.ExtClient,
		crdClient:           c.CRDClient,
		kubeInformerFactory: informers.NewFilteredSharedInformerFactory(c.KubeClient, c.ResyncPeriod, core.NamespaceAll, tweakListOptions),
		extInformerFactory:  vaultinformers.NewSharedInformerFactory(c.ExtClient, c.ResyncPeriod),
		recorder:            eventer.NewEventRecorder(c.KubeClient, "vault-controller"),
	}

	if err := ctrl.ensureCustomResourceDefinitions(); err != nil {
		return nil, err
	}
	if err := reg_util.UpdateValidatingWebhookCABundle(ctrl.clientConfig, validatingWebhook); err != nil {
		return nil, err
	}

	ctrl.initVaultServerWatcher()
	return ctrl, nil
}
