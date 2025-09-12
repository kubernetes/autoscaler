package smithyotelmetrics

import (
	"context"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/aws/smithy-go/metrics"
	otelmetric "go.opentelemetry.io/otel/metric"
)

type int64Counter struct {
	otel interface {
		Add(context.Context, int64, ...otelmetric.AddOption)
	}
}

var _ metrics.Int64Counter = (*int64Counter)(nil)
var _ metrics.Int64UpDownCounter = (*int64Counter)(nil)

func (i *int64Counter) Add(ctx context.Context, v int64, opts ...metrics.RecordMetricOption) {
	i.otel.Add(ctx, v, withMetricProps(opts...))
}

type int64Gauge struct {
	otel otelmetric.Int64Gauge
}

var _ metrics.Int64Gauge = (*int64Gauge)(nil)

func (i *int64Gauge) Sample(ctx context.Context, v int64, opts ...metrics.RecordMetricOption) {
	i.otel.Record(ctx, v, withMetricProps(opts...))
}

type int64Histogram struct {
	otel otelmetric.Int64Histogram
}

var _ metrics.Int64Histogram = (*int64Histogram)(nil)

func (i *int64Histogram) Record(ctx context.Context, v int64, opts ...metrics.RecordMetricOption) {
	i.otel.Record(ctx, v, withMetricProps(opts...))
}
