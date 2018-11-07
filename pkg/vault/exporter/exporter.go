package exporter

import (
	"fmt"

	core_util "github.com/appscode/kutil/core/v1"
	"github.com/kubevault/operator/pkg/vault/util"
	"github.com/pkg/errors"
	core "k8s.io/api/core/v1"
	mona "kmodules.xyz/monitoring-agent-api/api/v1"
)

const (
	VaultExporterStatsdPort       = 9125
	VaultExporterFetchMetricsPort = 9102
	PrometheusExporterPortName    = "prom-http"
)

type Exporter interface {
	Apply(pt *core.PodTemplateSpec, agent *mona.AgentSpec) error
	GetTelemetryConfig() (string, error)
}

type monitor struct {
	image string
}

var telemetryCfg = `telemetry {
  statsd_address = "0.0.0.0:%v"
}`

func NewExporter(image string) (Exporter, error) {
	return monitor{image: image}, nil
}

func (exp monitor) Apply(pt *core.PodTemplateSpec, agent *mona.AgentSpec) error {
	if pt == nil {
		return errors.New("podTempleSpec is nil")
	}

	cont := core.Container{
		Name:            util.VaultExporterContainerName,
		Image:           exp.image,
		ImagePullPolicy: core.PullIfNotPresent,
		Ports: []core.ContainerPort{
			{
				Name:          "udp",
				Protocol:      core.ProtocolUDP,
				ContainerPort: VaultExporterStatsdPort,
			},
			{
				Name:          PrometheusExporterPortName,
				Protocol:      core.ProtocolTCP,
				ContainerPort: VaultExporterFetchMetricsPort,
			},
		},
	}
	if agent != nil {
		cont.Resources = agent.Resources
	}

	pt.Spec.Containers = core_util.UpsertContainer(pt.Spec.Containers, cont)
	return nil
}

func (exp monitor) GetTelemetryConfig() (string, error) {
	return fmt.Sprintf(telemetryCfg, VaultExporterStatsdPort), nil
}
