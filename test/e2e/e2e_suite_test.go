package e2e_test

import (
	"testing"
	"time"

	logs "github.com/appscode/go/log/golog"
	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/reporters"
	. "github.com/onsi/gomega"
	ka "k8s.io/kube-aggregator/pkg/client/clientset_generated/clientset"
	"kmodules.xyz/client-go/tools/clientcmd"
	"kubevault.dev/operator/pkg/controller"
	"kubevault.dev/operator/test/e2e/framework"
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
	By("Using kubeconfig from " + options.KubeConfig)
	clientConfig, err := clientcmd.BuildConfigFromContext(options.KubeConfig, options.KubeContext)
	Expect(err).NotTo(HaveOccurred())
	// raise throttling time. ref: https://github.com/appscode/voyager/issues/640
	clientConfig.Burst = 100
	clientConfig.QPS = 100

	ctrlConfig := controller.NewConfig(clientConfig)
	ctrlConfig.MaxNumRequeues = 5
	ctrlConfig.NumThreads = 1
	ctrlConfig.ResyncPeriod = 10 * time.Second

	err = options.ApplyTo(ctrlConfig)
	Expect(err).NotTo(HaveOccurred())

	kaClient := ka.NewForConfigOrDie(clientConfig)
	Expect(err).NotTo(HaveOccurred())

	root = framework.New(ctrlConfig.KubeClient, ctrlConfig.ExtClient, ctrlConfig.AppCatalogClient, ctrlConfig.DbClient, kaClient, options.StartAPIServer, clientConfig, options.RunDynamoDBTest)
	err = root.CreateNamespace()
	Expect(err).NotTo(HaveOccurred())
	By("Using test namespace " + root.Namespace())
	if options.StartAPIServer {
		go root.StartAPIServerAndOperator(clientConfig, options.KubeConfig, options.ExtraOptions)
		root.EventuallyAPIServerReady().Should(Succeed())

		// let's API server be warmed up
		time.Sleep(time.Second * 5)
	} else if !framework.SelfHostedOperator {
		ctrl, err := ctrlConfig.New()
		Expect(err).NotTo(HaveOccurred())
		// Now let's start the controller
		go ctrl.RunInformers(nil)
	}

	By("Creating vault server version")
	err = root.CreateVaultserverVersion()
	Expect(err).NotTo(HaveOccurred())

	By("Deploying vault, mongodb, mysql, postgres...")
	err = root.InitialSetup()
	Expect(err).NotTo(HaveOccurred())

})

var _ = AfterSuite(func() {
	if options.StartAPIServer {
		By("Cleaning API server and Webhook stuff")
		root.CleanAdmissionConfigs()
	}

	Expect(root.Cleanup()).NotTo(HaveOccurred())
	By("Deleting Namespace...")
	root.DeleteNamespace()
	err := root.DeleteVaultserverVersion()
	Expect(err).NotTo(HaveOccurred())
})
