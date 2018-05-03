package test

import (
	"testing"
	"github.com/appscode/kutil/tools/clientcmd"
	"github.com/golang/glog"
	"github.com/soter/vault-operator/pkg/controller"
	"k8s.io/client-go/kubernetes"
	crd_cs "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/typed/apiextensions/v1beta1"
	cs "github.com/soter/vault-operator/client/clientset/versioned"
	"time"
)

func TestRun(t *testing.T) {
	clientConfig,err := clientcmd.BuildConfigFromContext("/home/ac/.kube/config", "minikube")
	if err!=nil {
		glog.Fatal(err)
	}

	/*_,err= rest.UnversionedRESTClientFor(clientConfig)
	if err!=nil {
		glog.Fatal(err)
	}*/

	ctrlConfig := controller.NewConfig(clientConfig)

	ctrlConfig.MaxNumRequeues = 5
	ctrlConfig.NumThreads = 1
	ctrlConfig.ResyncPeriod = 10*time.Second

	if ctrlConfig.KubeClient, err = kubernetes.NewForConfig(ctrlConfig.ClientConfig); err != nil {
		glog.Fatal(err)
	}
	if ctrlConfig.ExtClient, err = cs.NewForConfig(ctrlConfig.ClientConfig); err != nil {
		glog.Fatal(err)
	}
	if ctrlConfig.CRDClient, err = crd_cs.NewForConfig(ctrlConfig.ClientConfig); err != nil {
		glog.Fatal(err)
	}

	vaultCtrl, err := ctrlConfig.New()
	if err!=nil {
		glog.Fatal(err)
	}

	stopCh := make(chan struct{})

	vaultCtrl.RunInformers(stopCh)

}