package e2e_test

import (
	"flag"
	"log"
	"os"
	"path/filepath"

	"github.com/appscode/go/flags"
	logs "github.com/appscode/go/log/golog"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientsetscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/util/homedir"
	appcatscheme "kmodules.xyz/custom-resources/client/clientset/versioned/scheme"
	"kubevault.dev/operator/client/clientset/versioned/scheme"
	dbscheme "kubevault.dev/operator/client/clientset/versioned/scheme"
	"kubevault.dev/operator/pkg/cmds/server"
	"kubevault.dev/operator/test/e2e/framework"
)

type E2EOptions struct {
	*server.ExtraOptions

	KubeContext     string
	KubeConfig      string
	StartAPIServer  bool
	RunDynamoDBTest bool
}

var (
	options = &E2EOptions{
		ExtraOptions: server.NewExtraOptions(),
		KubeConfig: func() string {
			kubecfg := os.Getenv("KUBECONFIG")
			if kubecfg != "" {
				return kubecfg
			}
			return filepath.Join(homedir.HomeDir(), ".kube", "config")
		}(),
		StartAPIServer: false,
	}
)

func init() {
	utilruntime.Must(scheme.AddToScheme(clientsetscheme.Scheme))
	utilruntime.Must(appcatscheme.AddToScheme(clientsetscheme.Scheme))
	utilruntime.Must(dbscheme.AddToScheme(clientsetscheme.Scheme))

	options.AddGoFlags(flag.CommandLine)
	flag.StringVar(&options.KubeConfig, "kubeconfig", options.KubeConfig, "Path to kubeconfig file with authorization information (the master location is set by the master flag).")
	flag.StringVar(&options.KubeContext, "kube-context", "", "Name of kube context")
	flag.BoolVar(&options.StartAPIServer, "webhook", options.StartAPIServer, "Start API server for webhook")
	flag.BoolVar(&options.RunDynamoDBTest, "run-dynamodb-test", options.RunDynamoDBTest, "Run dynamoDB test")
	flag.BoolVar(&framework.SelfHostedOperator, "selfhosted-operator", framework.SelfHostedOperator, "Enable this for self-hosted operator")
	flag.StringVar(&framework.UnsealerImage, "unsealer-image", framework.UnsealerImage, "vault unsealer image")
	enableLogging()
	flag.Parse()
	framework.DockerRegistry = options.DockerRegistry
}

func enableLogging() {
	defer func() {
		logs.InitLogs()
		defer logs.FlushLogs()
	}()
	err := flag.Set("logtostderr", "true")
	if err != nil {
		log.Printf("Set flag failed with :%v\n", err)
	}
	logLevelFlag := flag.Lookup("v")
	if logLevelFlag != nil {
		if len(logLevelFlag.Value.String()) > 0 && logLevelFlag.Value.String() != "0" {
			return
		}
	}
	flags.SetLogLevel(2)
}
