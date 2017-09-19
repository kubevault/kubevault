package cmds

import (
	"os"
	"time"

	"github.com/appscode/steward/pkg/controller"
	"github.com/golang/glog"
	"github.com/hashicorp/vault/api"
	"github.com/spf13/cobra"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	master     string
	kubeconfig string

	opts = controller.Options{
		ClusterName:      "kubernetes",
		VaultAddress:     os.Getenv(api.EnvVaultAddress),
		VaultToken:       os.Getenv(api.EnvVaultToken),
		CACertFile:       os.Getenv(api.EnvVaultCAPath),
		ResyncPeriod:     5 * time.Minute,
		TokenRenewPeriod: 60 * time.Minute,
		MaxNumRequeues:   5,
	}
	// support unseal using secrets
)

func NewCmdRun() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "run",
		Short:             "Run operator",
		DisableAutoGenTag: true,
		Run: func(cmd *cobra.Command, args []string) {
			run()
		},
	}

	cmd.Flags().StringVar(&opts.ClusterName, "cluster-name", opts.ClusterName, "Name of Kubernetes cluster used to create backends")
	cmd.Flags().StringVar(&master, "master", master, "The address of the Kubernetes API server (overrides any value in kubeconfig)")
	cmd.Flags().StringVar(&kubeconfig, "kubeconfig", kubeconfig, "Path to kubeconfig file with authorization information (the master location is set by the master flag).")
	cmd.Flags().DurationVar(&opts.ResyncPeriod, "resync-period", opts.ResyncPeriod, "If non-zero, will re-list this often. Otherwise, re-list will be delayed aslong as possible (until the upstream source closes the watch or times out.")

	cmd.Flags().StringVar(&opts.VaultAddress, "vault-address", opts.VaultAddress, "Address of Vault server")
	cmd.Flags().StringVar(&opts.VaultToken, "vault-token", opts.VaultToken, "Vault token used by operator.")
	cmd.Flags().StringVar(&opts.CACertFile, "ca-cert-file", opts.CACertFile, "File containing CA certificate used by Vault server.")
	cmd.Flags().DurationVar(&opts.TokenRenewPeriod, "token-renew-period", opts.TokenRenewPeriod, "Interval between consecutive attempts at renewing vault tokens.")

	return cmd
}

func run() {
	// creates the connection
	config, err := clientcmd.BuildConfigFromFlags(master, kubeconfig)
	if err != nil {
		glog.Fatal(err)
	}

	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		glog.Fatal(err)
	}

	controller := controller.New(clientset, opts)

	// Now let's start the controller
	stop := make(chan struct{})
	defer close(stop)
	go controller.Run(1, stop)

	// Wait forever
	select {}
}
