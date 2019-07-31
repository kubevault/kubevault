package cmds

import (
	"flag"
	"os"

	"github.com/appscode/go/flags"
	v "github.com/appscode/go/version"
	"github.com/spf13/cobra"
	genericapiserver "k8s.io/apiserver/pkg/server"
	clientsetscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/kubernetes/pkg/api/legacyscheme"
	"kmodules.xyz/client-go/logs"
	"kmodules.xyz/client-go/tools/cli"
	appcatscheme "kmodules.xyz/custom-resources/client/clientset/versioned/scheme"
	"kubevault.dev/operator/client/clientset/versioned/scheme"
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
			scheme.AddToScheme(legacyscheme.Scheme)
		},
	}
	rootCmd.PersistentFlags().AddGoFlagSet(flag.CommandLine)
	logs.ParseFlags()
	rootCmd.PersistentFlags().BoolVar(&cli.EnableAnalytics, "enable-analytics", cli.EnableAnalytics, "Send analytical events to Google Analytics")

	rootCmd.AddCommand(v.NewCmdVersion())
	stopCh := genericapiserver.SetupSignalHandler()
	rootCmd.AddCommand(NewCmdRun(os.Stdout, os.Stderr, stopCh))

	return rootCmd
}
