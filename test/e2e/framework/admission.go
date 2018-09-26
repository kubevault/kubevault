package framework

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"

	discovery_util "github.com/appscode/kutil/discovery"
	shell "github.com/codeskyblue/go-sh"
	api "github.com/kubevault/operator/apis/kubevault/v1alpha1"
	srvr "github.com/kubevault/operator/pkg/cmds/server"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	genericapiserver "k8s.io/apiserver/pkg/server"
	"k8s.io/client-go/discovery"
	restclient "k8s.io/client-go/rest"
	kapi "k8s.io/kube-aggregator/pkg/apis/apiregistration/v1beta1"
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
		api.EnableStatusSubresource = true
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

func (f *Framework) EventuallyAPIServerReady() GomegaAsyncAssertion {
	return Eventually(
		func() error {
			apiservice, err := f.KAClient.ApiregistrationV1beta1().APIServices().Get("v1alpha1.admission.kubevault.com", metav1.GetOptions{})
			if err != nil {
				return err
			}
			for _, cond := range apiservice.Status.Conditions {
				if cond.Type == kapi.Available && cond.Status == kapi.ConditionTrue && cond.Reason == "Passed" {
					return nil
				}
			}
			return fmt.Errorf("ApiService not ready yet")
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
