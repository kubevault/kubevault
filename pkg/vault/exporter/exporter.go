/*
Copyright AppsCode Inc. and Contributors

Licensed under the AppsCode Community License 1.0.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    https://github.com/appscode/licenses/raw/1.0.0/AppsCode-Community-1.0.0.md

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package exporter

import (
	"context"
	"fmt"

	conapi "kubevault.dev/apimachinery/apis"
	capi "kubevault.dev/apimachinery/apis/catalog/v1alpha1"
	api "kubevault.dev/apimachinery/apis/kubevault/v1alpha1"
	"kubevault.dev/operator/pkg/vault/util"

	"github.com/pkg/errors"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
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
	config  capi.VaultServerVersionExporter
	kClient kubernetes.Interface
}

var telemetryCfg = `telemetry {
  statsd_address = "0.0.0.0:%v"
}`

func NewExporter(config *capi.VaultServerVersion, kClient kubernetes.Interface) (Exporter, error) {
	return monitor{config: config.Spec.Exporter, kClient: kClient}, nil
}

func (exp monitor) Apply(pt *core.PodTemplateSpec, vs *api.VaultServer) error {
	if pt == nil {
		return errors.New("podTempleSpec is nil")
	}

	agent := vs.Spec.Monitor

	c := core.Container{
		Name:            util.VaultExporterContainerName,
		Image:           exp.config.Image,
		ImagePullPolicy: exp.config.ImagePullPolicy,
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

	if vs.Spec.TLS != nil && vs.Spec.TLS.Certificates != nil {
		// Get k8s secret
		secretName := vs.GetCertSecretName("vault")
		secret, err := exp.kClient.CoreV1().Secrets(vs.Namespace).Get(context.TODO(), secretName, metav1.GetOptions{})
		if err != nil {
			return errors.Wrap(err, "failed to get secret")
		}

		// Read ca.crt from the secret
		// caCert := conapi.TLSCACertKey
		byt, ok := secret.Data[conapi.TLSCACertKey]
		if !ok {
			return errors.New("missing ca.crt in vault secret")
		}

		// export ca.crt as tls-cacert
		// c.Args = append(c.Args, fmt.Sprintf("--vault.tls-cacert=%s", vs.Spec.TLS.Certificates))
		c.Args = append(c.Args, fmt.Sprintf("--vault.tls-cacert=%s", string(byt)))
	}

	if agent != nil && agent.Prometheus != nil {
		c.Args = append(c.Args, agent.Prometheus.Exporter.Args...)
		c.Env = agent.Prometheus.Exporter.Env
		c.Resources = agent.Prometheus.Exporter.Resources
		c.SecurityContext = agent.Prometheus.Exporter.SecurityContext
	}

	pt.Spec.Containers = core_util.UpsertContainer(pt.Spec.Containers, c)
	return nil
}

func (exp monitor) GetTelemetryConfig() (string, error) {
	return fmt.Sprintf(telemetryCfg, VaultExporterStatsdPort), nil
}
