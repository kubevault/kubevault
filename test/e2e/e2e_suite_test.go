package e2e_test

import (
	"testing"
	"time"

	logs "github.com/appscode/go/log/golog"
	"github.com/appscode/kutil/tools/clientcmd"
	"github.com/kubevault/operator/pkg/controller"
	"github.com/kubevault/operator/test/e2e/framework"
	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/reporters"
	. "github.com/onsi/gomega"
)

const (
	TIMEOUT = 20 * time.Minute
)

var (
	ctrl *controller.VaultController
	root *framework.Framework
)

func TestE2e(t *testing.T) {
	logs.InitLogs()
	RegisterFailHandler(Fail)
	SetDefaultEventuallyTimeout(TIMEOUT)
	junitReporter := reporters.NewJUnitReporter("junit.xml")
	RunSpecsWithDefaultAndCustomReporters(t, "e2e Suite", []Reporter{junitReporter})
}

var _ = BeforeSuite(func() {
	clientConfig, err := clientcmd.BuildConfigFromContext(options.KubeConfig, options.KubeContext)
	Expect(err).NotTo(HaveOccurred())

	ctrlConfig := controller.NewConfig(clientConfig)
	ctrlConfig.MaxNumRequeues = 5
	ctrlConfig.NumThreads = 1
	ctrlConfig.ResyncPeriod = 10 * time.Second

	err = options.ApplyTo(ctrlConfig)
	Expect(err).NotTo(HaveOccurred())

	ctrl, err := ctrlConfig.New()
	Expect(err).NotTo(HaveOccurred())

	root = framework.New(ctrlConfig.KubeClient, ctrlConfig.ExtClient, nil, options.StartAPIServer, clientConfig)
	err = root.CreateNamespace()
	Expect(err).NotTo(HaveOccurred())
	By("Using test namespace " + root.Namespace())

	if options.StartAPIServer {
		// still not implemented
		By("Not implemented")
	} else {
		// Now let's start the controller
		go ctrl.RunInformers(nil)
	}

})

var _ = AfterSuite(func() {
	if options.StartAPIServer {
		//By("Cleaning API server and Webhook stuff")
		//root.KubeClient.AdmissionregistrationV1beta1().MutatingWebhookConfigurations().Delete("admission.stash.appscode.com", meta.DeleteInBackground())
		//root.KubeClient.AdmissionregistrationV1beta1().ValidatingWebhookConfigurations().Delete("admission.stash.appscode.com", meta.DeleteInBackground())
		//root.KubeClient.CoreV1().Endpoints(root.Namespace()).Delete("stash-local-apiserver", meta.DeleteInBackground())
		//root.KubeClient.CoreV1().Services(root.Namespace()).Delete("stash-local-apiserver", meta.DeleteInBackground())
		//root.KAClient.ApiregistrationV1beta1().APIServices().Delete("v1alpha1.admission.stash.appscode.com", meta.DeleteInBackground())
	}
	root.DeleteNamespace()
})
