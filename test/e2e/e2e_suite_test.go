/*
Copyright The KubeVault Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package e2e_test

import (
	"testing"
	"time"

	"kubevault.dev/operator/pkg/controller"
	"kubevault.dev/operator/test/e2e/framework"

	logs "github.com/appscode/go/log/golog"
	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/reporters"
	. "github.com/onsi/gomega"
	ka "k8s.io/kube-aggregator/pkg/client/clientset_generated/clientset"
	"kmodules.xyz/client-go/tools/clientcmd"
)

const (
	TIMEOUT = 20 * time.Minute
)

var (
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
	err := root.DeleteNamespace()
	Expect(err).NotTo(HaveOccurred())
	err = root.DeleteVaultserverVersion()
	Expect(err).NotTo(HaveOccurred())
})
