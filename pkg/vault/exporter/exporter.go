/*
Copyright The KubeVault Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package exporter

import (
	"fmt"

	api "kubevault.dev/operator/apis/kubevault/v1alpha1"
	"kubevault.dev/operator/pkg/vault/util"

	"github.com/pkg/errors"
	core "k8s.io/api/core/v1"
	core_util "kmodules.xyz/client-go/core/v1"
)

const (
	VaultExporterStatsdPort       = 9125
	VaultExporterFetchMetricsPort = 9102
	PrometheusExporterPortName    = "prom-http"
)

type Exporter interface {
	Apply(pt *core.PodTemplateSpec, vs *api.VaultServer) error
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

func (exp monitor) Apply(pt *core.PodTemplateSpec, vs *api.VaultServer) error {
	if pt == nil {
		return errors.New("podTempleSpec is nil")
	}

	agent := vs.Spec.Monitor

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

	if vs.Spec.TLS != nil && vs.Spec.TLS.CABundle != nil {
		cont.Args = append(cont.Args, fmt.Sprintf("--vault.tls-cacert=%s", vs.Spec.TLS.CABundle))
	}

	if agent != nil {
		cont.Args = append(cont.Args, agent.Args...)
		cont.Env = agent.Env
		cont.Resources = agent.Resources
		cont.SecurityContext = agent.SecurityContext
	}

	pt.Spec.Containers = core_util.UpsertContainer(pt.Spec.Containers, cont)
	return nil
}

func (exp monitor) GetTelemetryConfig() (string, error) {
	return fmt.Sprintf(telemetryCfg, VaultExporterStatsdPort), nil
}
