package e2e_test

import (
	"flag"
	"path/filepath"

	"github.com/appscode/go/flags"
	logs "github.com/appscode/go/log/golog"
	clientSetScheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/util/homedir"
	appcatscheme "kmodules.xyz/custom-resources/client/clientset/versioned/scheme"
	dbscheme "kubedb.dev/apimachinery/client/clientset/versioned/scheme"
	"kubevault.dev/operator/client/clientset/versioned/scheme"
	"kubevault.dev/operator/pkg/cmds/server"
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
		ExtraOptions:   server.NewExtraOptions(),
		KubeConfig:     filepath.Join(homedir.HomeDir(), ".kube", "config"),
		StartAPIServer: false,
	}
)

func init() {
	scheme.AddToScheme(clientSetScheme.Scheme)
	appcatscheme.AddToScheme(clientSetScheme.Scheme)
	dbscheme.AddToScheme(clientSetScheme.Scheme)

	options.AddGoFlags(flag.CommandLine)
	flag.StringVar(&options.KubeConfig, "kubeconfig", options.KubeConfig, "Path to kubeconfig file with authorization information (the master location is set by the master flag).")
	flag.StringVar(&options.KubeContext, "kube-context", "", "Name of kube context")
	flag.BoolVar(&options.StartAPIServer, "webhook", options.StartAPIServer, "Start API server for webhook")
	flag.BoolVar(&options.RunDynamoDBTest, "run-dynamodb-test", options.RunDynamoDBTest, "Run dynamoDB test")
	enableLogging()
	flag.Parse()
}

func enableLogging() {
	defer func() {
		logs.InitLogs()
		defer logs.FlushLogs()
	}()
	flag.Set("logtostderr", "true")
	logLevelFlag := flag.Lookup("v")
	if logLevelFlag != nil {
		if len(logLevelFlag.Value.String()) > 0 && logLevelFlag.Value.String() != "0" {
			return
		}
	}
	flags.SetLogLevel(2)
}
