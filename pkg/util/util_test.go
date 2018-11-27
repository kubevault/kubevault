package util

import (
	"fmt"
	"testing"
	"time"

	"github.com/appscode/kutil/tools/clientcmd"
	"github.com/golang/glog"
	"k8s.io/client-go/kubernetes"
)

func TestTryGetJwtTokenSecretNameFromServiceAccount(t *testing.T) {
	t.Skip()
	// look this carefully
	clientConfig, err := clientcmd.BuildConfigFromContext("/home/ac/.kube/config",
		"minikube")
	if err != nil {
		glog.Fatal(err)
	}

	kc, err := kubernetes.NewForConfig(clientConfig)
	if err != nil {
		glog.Fatal(err)
	}

	s, err := TryGetJwtTokenSecretNameFromServiceAccount(kc, "ddd", "default", time.Second, time.Second*3)
	fmt.Println(err)
	fmt.Println(s)
}
