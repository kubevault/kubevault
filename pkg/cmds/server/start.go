package server

import (
	"fmt"
	"io"
	"net"

	"kubevault.dev/operator/pkg/controller"
	"kubevault.dev/operator/pkg/metrics"
	"kubevault.dev/operator/pkg/server"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/spf13/pflag"
	admissionv1beta1 "k8s.io/api/admission/v1beta1"
	genericapiserver "k8s.io/apiserver/pkg/server"
	genericoptions "k8s.io/apiserver/pkg/server/options"
	"kmodules.xyz/client-go/meta"
	"kmodules.xyz/client-go/tools/clientcmd"
	metricsutil "searchlight.dev/prometheus-metrics-exporter/metrics"
)

const defaultEtcdPathPrefix = "/registry/kubevault.com"

type VaultServerOptions struct {
	RecommendedOptions *genericoptions.RecommendedOptions
	ExtraOptions       *ExtraOptions
	MetricsExporterCfg *metricsutil.MetricsExporterConfigs

	StdOut io.Writer
	StdErr io.Writer
}

func NewVaultServerOptions(out, errOut io.Writer) *VaultServerOptions {
	o := &VaultServerOptions{
		// TODO we will nil out the etcd storage options.  This requires a later level of k8s.io/apiserver
		RecommendedOptions: genericoptions.NewRecommendedOptions(
			defaultEtcdPathPrefix,
			server.Codecs.LegacyCodec(admissionv1beta1.SchemeGroupVersion),
			genericoptions.NewProcessInfo("vault-operator", meta.Namespace()),
		),
		ExtraOptions:       NewExtraOptions(),
		MetricsExporterCfg: metricsutil.NewMetricsExporterConfigs(),
		StdOut:             out,
		StdErr:             errOut,
	}
	o.RecommendedOptions.Etcd = nil
	o.RecommendedOptions.Admission = nil

	return o
}

func (o VaultServerOptions) AddFlags(fs *pflag.FlagSet) {
	o.RecommendedOptions.AddFlags(fs)
	o.ExtraOptions.AddFlags(fs)
	o.MetricsExporterCfg.AddFlags(fs)
}

func (o VaultServerOptions) Validate(args []string) error {
	if err := o.MetricsExporterCfg.Validate(); err != nil {
		return err
	}
	return nil
}

func (o *VaultServerOptions) Complete() error {
	return nil
}

func (o VaultServerOptions) Config() (*server.VaultServerConfig, error) {
	// TODO have a "real" external address
	if err := o.RecommendedOptions.SecureServing.MaybeDefaultWithSelfSignedCerts("localhost", nil, []net.IP{net.ParseIP("127.0.0.1")}); err != nil {
		return nil, fmt.Errorf("error creating self-signed certificates: %v", err)
	}

	serverConfig := genericapiserver.NewRecommendedConfig(server.Codecs)
	serverConfig.EnableMetrics = true
	if err := o.RecommendedOptions.ApplyTo(serverConfig); err != nil {
		return nil, err
	}
	// Fixes https://github.com/Azure/AKS/issues/522
	clientcmd.Fix(serverConfig.ClientConfig)

	extraConfig := controller.NewConfig(serverConfig.ClientConfig)
	if err := o.ExtraOptions.ApplyTo(extraConfig); err != nil {
		return nil, err
	}

	config := &server.VaultServerConfig{
		GenericConfig: serverConfig,
		ExtraConfig:   extraConfig,
	}
	return config, nil
}

func (o VaultServerOptions) Run(stopCh <-chan struct{}) error {
	config, err := o.Config()
	if err != nil {
		return err
	}

	s, err := config.Complete().New()
	if err != nil {
		return err
	}

	registry, ok := prometheus.DefaultRegisterer.(*prometheus.Registry)
	if !ok {
		return fmt.Errorf("failed to convert  prometheus.DefaultRegisterer to *prometheus.Registry")
	}
	// non-blocking
	err = metrics.RunMetricsExporter(o.MetricsExporterCfg, registry, stopCh)
	if err != nil {
		return err
	}

	return s.Run(stopCh)
}
