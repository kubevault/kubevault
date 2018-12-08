package cmds

import (
	"flag"
	"os"

	"github.com/appscode/go/flags"
	v "github.com/appscode/go/version"
	"github.com/appscode/kutil/tools/cli"
	dbscheme "github.com/kubedb/apimachinery/client/clientset/versioned/scheme"
	"github.com/kubevault/operator/client/clientset/versioned/scheme"
	"github.com/spf13/cobra"
	genericapiserver "k8s.io/apiserver/pkg/server"
	clientsetscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/kubernetes/pkg/api/legacyscheme"
	appcatscheme "kmodules.xyz/custom-resources/client/clientset/versioned/scheme"
)

func NewRootCmd() *cobra.Command {
	var rootCmd = &cobra.Command{
		Use:               "vault-operator [command]",
		Short:             `Vault Operator by AppsCode - HashiCorp Vault Operator for Kubernetes`,
		DisableAutoGenTag: true,
		PersistentPreRun: func(c *cobra.Command, args []string) {
			flags.DumpAll(c.Flags())
			cli.SendAnalytics(c, v.Version.Version)

			scheme.AddToScheme(clientsetscheme.Scheme)
			appcatscheme.AddToScheme(clientsetscheme.Scheme)
			dbscheme.AddToScheme(clientsetscheme.Scheme)
			scheme.AddToScheme(legacyscheme.Scheme)
		},
	}
	rootCmd.PersistentFlags().AddGoFlagSet(flag.CommandLine)
	// ref: https://github.com/kubernetes/kubernetes/issues/17162#issuecomment-225596212
	flag.CommandLine.Parse([]string{})
	rootCmd.PersistentFlags().BoolVar(&cli.EnableAnalytics, "enable-analytics", cli.EnableAnalytics, "Send analytical events to Google Analytics")

	rootCmd.AddCommand(v.NewCmdVersion())
	stopCh := genericapiserver.SetupSignalHandler()
	rootCmd.AddCommand(NewCmdRun(os.Stdout, os.Stderr, stopCh))

	return rootCmd
}
