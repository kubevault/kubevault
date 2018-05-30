package main

import (
	"os"

	logs "github.com/appscode/go/log/golog"
	"github.com/kube-vault/operator/pkg/cmds"
)

func main() {
	logs.InitLogs()
	defer logs.FlushLogs()

	if err := cmds.NewRootCmd().Execute(); err != nil {
		os.Exit(1)
	}
	os.Exit(0)
}
