package metrics

import (
	"fmt"

	"github.com/golang/glog"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/prometheus/prompb"
	"searchlight.dev/prometheus-metrics-exporter/metrics"
)

// This will periodically send metrics
// If 'registry' is nil, then 'prometheus.DefaultRegisterer' will be used
// non-blocking
func RunMetricsExporter(conf *metrics.MetricsExporterConfigs, registry *prometheus.Registry, stopCh <-chan struct{}) error {
	if registry == nil {
		return fmt.Errorf("invalid registry")
	}

	mExporter, err := metrics.NewMetricsExporter(conf, registry)
	if err != nil {
		return err
	}
	// for 'up' metrics
	mExporter.Register(metrics.NewHealthCollector())

	if err := mExporter.Run(stopCh, []prompb.Label{
		{
			Name:  "operator",
			Value: "kubevault",
		},
	}); err != nil {
		return err
	}
	glog.Infoln("metrics exporter is started...")
	return nil
}
