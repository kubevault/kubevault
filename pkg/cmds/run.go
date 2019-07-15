package cmds

import (
	"io"

	"github.com/appscode/go/log"
	v "github.com/appscode/go/version"
	"github.com/spf13/cobra"
	"kmodules.xyz/client-go/tools/cli"
	"kubevault.dev/operator/pkg/cmds/server"
)

func NewCmdRun(out, errOut io.Writer, stopCh <-chan struct{}) *cobra.Command {
	o := server.NewVaultServerOptions(out, errOut)

	cmd := &cobra.Command{
		Use:               "run",
		Short:             "Launch Vault operator",
		DisableAutoGenTag: true,
		PreRun: func(c *cobra.Command, args []string) {
			cli.SendPeriodicAnalytics(c, v.Version.Version)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			log.Infof("Starting operator version %s+%s ...", v.Version.Version, v.Version.CommitHash)

			if err := o.Complete(); err != nil {
				return err
			}
			if err := o.Validate(args); err != nil {
				return err
			}
			if err := o.Run(stopCh); err != nil {
				return err
			}
			return nil
		},
	}

	o.AddFlags(cmd.Flags())

	return cmd
}
