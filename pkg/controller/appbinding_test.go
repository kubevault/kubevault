package controller

import (
	"encoding/json"
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/appscode/kutil/tools/clientcmd"
	"github.com/golang/glog"
	vaultconfig "github.com/kubevault/operator/apis/config/v1alpha1"
	api "github.com/kubevault/operator/apis/kubevault/v1alpha1"
	cs "github.com/kubevault/operator/client/clientset/versioned"
	"github.com/kubevault/operator/pkg/vault"
	"github.com/stretchr/testify/assert"
	core "k8s.io/api/core/v1"
	crd_cs "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/typed/apiextensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	appcat "kmodules.xyz/custom-resources/apis/appcatalog/v1alpha1"
	appcat_cs "kmodules.xyz/custom-resources/client/clientset/versioned/typed/appcatalog/v1alpha1"
)

func TestRun(t *testing.T) {
	// look this carefully
	clientConfig, err := clientcmd.BuildConfigFromContext("/home/ac/.kube/config",
		"minikube")
	if err != nil {
		glog.Fatal(err)
	}

	/*_,err= rest.UnversionedRESTClientFor(clientConfig)
	if err!=nil {
		glog.Fatal(err)
	}*/

	ctrlConfig := NewConfig(clientConfig)

	ctrlConfig.MaxNumRequeues = 5
	ctrlConfig.NumThreads = 1
	ctrlConfig.ResyncPeriod = 10 * time.Minute

	if ctrlConfig.KubeClient, err = kubernetes.NewForConfig(ctrlConfig.ClientConfig); err != nil {
		glog.Fatal(err)
	}
	if ctrlConfig.ExtClient, err = cs.NewForConfig(ctrlConfig.ClientConfig); err != nil {
		glog.Fatal(err)
	}
	if ctrlConfig.CRDClient, err = crd_cs.NewForConfig(ctrlConfig.ClientConfig); err != nil {
		glog.Fatal(err)
	}
	if ctrlConfig.AppCatalogClient, err = appcat_cs.NewForConfig(ctrlConfig.ClientConfig); err != nil {
		glog.Fatal(err)
	}

	ctrl, err := ctrlConfig.New()
	if err != nil {
		glog.Fatal(err)
	}

	_, err = vault.NewClient(ctrl.kubeClient, ctrl.appCatalogClient, &appcat.AppReference{
		Name:      "vault",
		Namespace: "default",
	})
	fmt.Println(err)
	return

	err = ctrl.ensureAppBindings(&api.VaultServer{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "default",
			UID:       "123",
		},
	}, &vaultFake{
		sr: &core.Secret{
			Data: map[string][]byte{
				"ca.crt": []byte("ca"),
			},
		},
		svc: &core.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test",
				Namespace: "default",
			},
		},
	})

	assert.Nil(t, err)
}

func TestJson(t *testing.T) {
	k8sConf, err := json.Marshal(vaultconfig.KubernetesAuthConfiguration{
		Role: "test",
	})
	if err != nil {
		assert.Nil(t, err, "marshal k8s conf")
	}
	// http://goinbigdata.com/how-to-correctly-serialize-json-string-in-golang/
	params, err := json.Marshal(string(k8sConf))
	if err != nil {
		assert.Nil(t, err, "marshal string")
	}
	//params := k8sConf
	fmt.Println("1:", string(params))
	fmt.Printf("%s\n", string(params))

	s, err := strconv.Unquote(string(params))
	if err != nil {
		assert.Nil(t, err, "unquote")
	}

	var m vaultconfig.KubernetesAuthConfiguration
	err = json.Unmarshal([]byte(s), &m)
	if err != nil {
		assert.Nil(t, err, "unmarshal")
	} else {
		fmt.Println(m)
	}
}
