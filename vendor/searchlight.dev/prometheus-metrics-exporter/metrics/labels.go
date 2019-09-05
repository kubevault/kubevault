package metrics

import (
	"os"

	"github.com/prometheus/prometheus/prompb"
)

const (
	PodNameEnv      = "POD_NAME"
	PodNamespaceEnv = "POD_NAMESPACE"
	PodIPEnv        = "POD_IP"
	ServiceNameEnv  = "SERVICE_NAME"

	PodNameLabel      = "pod"
	PodNamespaceLabel = "namespace"
	PodIPLabel        = "instance"
	ServiceNameLabel  = "service"
)

func GetLabels() []prompb.Label {
	return []prompb.Label{
		{
			Name:  PodNameLabel,
			Value: os.Getenv(PodNameEnv),
		},
		{
			Name:  PodNamespaceLabel,
			Value: os.Getenv(PodNamespaceEnv),
		},
		{
			Name:  PodIPLabel,
			Value: os.Getenv(PodIPEnv),
		},
		{
			Name:  ServiceNameLabel,
			Value: os.Getenv(ServiceNameEnv),
		},
	}
}
