package framework

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"

	shell "github.com/codeskyblue/go-sh"
	"github.com/golang/glog"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	genericapiserver "k8s.io/apiserver/pkg/server"
	"k8s.io/client-go/discovery"
	restclient "k8s.io/client-go/rest"
	kapi "k8s.io/kube-aggregator/pkg/apis/apiregistration/v1beta1"
	discovery_util "kmodules.xyz/client-go/discovery"
	"kubevault.dev/operator/apis"
	srvr "kubevault.dev/operator/pkg/cmds/server"
)

func (f *Framework) NewTestVaultServerOptions(kubeConfigPath string, controllerOptions *srvr.ExtraOptions) *srvr.VaultServerOptions {
	opt := srvr.NewVaultServerOptions(os.Stdout, os.Stderr)
	opt.RecommendedOptions.Authentication.RemoteKubeConfigFile = kubeConfigPath
	opt.RecommendedOptions.Authorization.RemoteKubeConfigFile = kubeConfigPath
	opt.RecommendedOptions.CoreAPI.CoreAPIKubeconfigPath = kubeConfigPath
	opt.RecommendedOptions.Authentication.SkipInClusterLookup = true
	opt.RecommendedOptions.SecureServing.BindPort = 8443
	opt.RecommendedOptions.SecureServing.BindAddress = net.ParseIP("127.0.0.1")
	opt.ExtraOptions = controllerOptions
	opt.StdErr = os.Stderr
	opt.StdOut = os.Stdout

	return opt
}

func (f *Framework) StartAPIServerAndOperator(config *restclient.Config, kubeConfigPath string, ctrlOptions *srvr.ExtraOptions) {
	defer GinkgoRecover()

	discClient, err := discovery.NewDiscoveryClientForConfig(config)
	Expect(err).NotTo(HaveOccurred())
	serverVersion, err := discovery_util.GetBaseVersion(discClient)
	Expect(err).NotTo(HaveOccurred())
	if strings.Compare(serverVersion, "1.11") >= 0 {
		apis.EnableStatusSubresource = true
	}

	sh := shell.NewSession()
	args := []interface{}{"--test=true"}
	SetupServer := filepath.Join("..", "..", "hack", "dev", "run.sh")

	By("Creating API server and webhook stuffs")
	cmd := sh.Command(SetupServer, args...)
	err = cmd.Run()
	Expect(err).ShouldNot(HaveOccurred())

	By("Starting Server and Operator")
	stopCh := genericapiserver.SetupSignalHandler()
	vsOptions := f.NewTestVaultServerOptions(kubeConfigPath, ctrlOptions)
	err = vsOptions.Run(stopCh)
	Expect(err).ShouldNot(HaveOccurred())
}

func (f *Framework) isApiSvcReady(apiSvcName string) error {
	apiSvc, err := f.KAClient.ApiregistrationV1beta1().APIServices().Get(apiSvcName, metav1.GetOptions{})
	if err != nil {
		return err
	}
	for _, cond := range apiSvc.Status.Conditions {
		if cond.Type == kapi.Available && cond.Status == kapi.ConditionTrue {
			glog.Infof("APIService %v status is true", apiSvcName)
			return nil
		}
	}
	glog.Errorf("APIService %v not ready yet", apiSvcName)
	return fmt.Errorf("APIService %v not ready yet", apiSvcName)
}

func (f *Framework) EventuallyAPIServerReady() GomegaAsyncAssertion {
	return Eventually(
		func() error {
			if err := f.isApiSvcReady("v1alpha1.mutators.kubevault.com"); err != nil {
				return err
			}
			if err := f.isApiSvcReady("v1alpha1.validators.kubevault.com"); err != nil {
				return err
			}
			if err := f.isApiSvcReady("v1alpha1.validators.authorization.kubedb.com"); err != nil {
				return err
			}
			time.Sleep(time.Second * 5) // let the resource become available
			return nil
		},
		time.Minute*5,
		time.Second*2,
	)
}

func (f *Framework) CleanAdmissionConfigs() {
	// delete validating webhook
	if err := f.KubeClient.AdmissionregistrationV1beta1().ValidatingWebhookConfigurations().DeleteCollection(deleteInForeground(), metav1.ListOptions{
		LabelSelector: "app=vault-operator",
	}); err != nil && !kerr.IsNotFound(err) {
		fmt.Printf("error in deletion of Validating Webhook. Error: %v", err)
	}

	// Delete APIService
	if err := f.KAClient.ApiregistrationV1beta1().APIServices().DeleteCollection(deleteInBackground(), metav1.ListOptions{
		LabelSelector: "app=vault-operator",
	}); err != nil && !kerr.IsNotFound(err) {
		fmt.Printf("error in deletion of APIService. Error: %v", err)
	}

	// Delete Service
	if err := f.KubeClient.CoreV1().Services("default").Delete("vault-operator", &metav1.DeleteOptions{}); err != nil && !kerr.IsNotFound(err) {
		fmt.Printf("error in deletion of Service. Error: %v", err)
	}

	// Delete EndPoints
	if err := f.KubeClient.CoreV1().Endpoints("default").DeleteCollection(deleteInBackground(), metav1.ListOptions{
		LabelSelector: "app=vault-operator",
	}); err != nil && !kerr.IsNotFound(err) {
		fmt.Printf("error in deletion of Endpoints. Error: %v", err)
	}

	time.Sleep(time.Second * 1) // let the vault-server know it!!
}
