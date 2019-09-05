package metrics

import (
	"context"
	"time"

	"github.com/appscode/go/wait"
	"github.com/gogo/protobuf/proto"
	"github.com/golang/glog"
	"github.com/golang/snappy"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/prompb"
)

type RemoteWriter struct {
	client      *RemoteClient
	interval    time.Duration
	gatherer    prometheus.Gatherer
	extraLabels []prompb.Label
}

func NewRemoteWriter(cl *RemoteClient, g prometheus.Gatherer, interval time.Duration, extraLabels []prompb.Label) (*RemoteWriter, error) {
	if cl == nil {
		return nil, errors.New("remote storage client can not be nil")
	}
	if g == nil {
		return nil, errors.New("prometheus metrics gatherer can not be nil")
	}
	return &RemoteWriter{
		client:      cl,
		interval:    interval,
		gatherer:    g,
		extraLabels: extraLabels,
	}, nil
}

func (w *RemoteWriter) Run(stopCh <-chan struct{}) {
	wait.JitterUntil(func() {
		err := w.remoteWrite(context.TODO())
		if err != nil {
			glog.Errorf("metrics exporter: %v", err)
		}
	}, w.interval, 0, false, stopCh)
}

// it will write metrics to remote storage
func (w *RemoteWriter) remoteWrite(ctx context.Context) error {
	mfs, err := w.gatherer.Gather()
	if err != nil {
		return errors.Wrap(err, "failed to collect metrics")
	}

	samples, err := metricFamilyToTimeseries(mfs, w.extraLabels)
	if err != nil {
		return errors.Wrap(err, "failed to convert metric family to time series")
	}

	req, err := buildWriteRequest(samples)
	if err != nil {
		return errors.Wrap(err, "failed to build prometheus write request")
	}

	err = w.client.Store(ctx, req)
	if err != nil {
		return errors.Wrap(err, "failed to writes metrics to remote client")
	}
	return nil
}

func metricFamilyToTimeseries(mfs []*dto.MetricFamily, extraLabels []prompb.Label) ([]prompb.TimeSeries, error) {
	var ts []prompb.TimeSeries
	for _, mf := range mfs {
		vec, err := expfmt.ExtractSamples(&expfmt.DecodeOptions{
			Timestamp: model.Now(),
		}, mf)
		if err != nil {
			return nil, err
		}

		for _, s := range vec {
			if s != nil {
				ts = append(ts, prompb.TimeSeries{
					Labels: metricToLabels(s.Metric, extraLabels),
					Samples: []prompb.Sample{
						{
							Value:     float64(s.Value),
							Timestamp: int64(s.Timestamp),
						},
					},
				})
			}
		}
	}
	return ts, nil
}

func metricToLabels(m model.Metric, extraLabels []prompb.Label) []prompb.Label {
	var lables []prompb.Label
	for k, v := range m {
		lables = append(lables, prompb.Label{
			Name:  string(k),
			Value: string(v),
		})
	}
	lables = append(lables, extraLabels...)
	return lables
}

// https://github.com/prometheus/prometheus/blob/84df210c410a0684ec1a05479bfa54458562695e/storage/remote/queue_manager.go#L759
func buildWriteRequest(samples []prompb.TimeSeries) ([]byte, error) {
	req := &prompb.WriteRequest{
		Timeseries: samples,
	}

	data, err := proto.Marshal(req)
	if err != nil {
		return nil, err
	}

	compressed := snappy.Encode(nil, data)
	return compressed, nil
}
