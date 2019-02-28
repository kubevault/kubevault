package clusterid

import (
	"flag"

	"github.com/spf13/pflag"
)

var clusterName = ""

func AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&clusterName, "cluster-name", clusterName, "Name of cluster used in a multi-cluster setup")
}

func AddGoFlags(fs *flag.FlagSet) {
	fs.StringVar(&clusterName, "cluster-name", clusterName, "Name of cluster used in a multi-cluster setup")
}

func ClusterName() string {
	return clusterName
}
