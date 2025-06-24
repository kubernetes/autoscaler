package smithyotelmetrics

import (
	"context"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/aws/smithy-go/metrics"
	otelmetric "go.opentelemetry.io/otel/metric"
)

type otelFloat64Add interface {
	Add(context.Context, float64, ...otelmetric.AddOption)
}

type float64Counter struct {
	otel otelFloat64Add
}

var _ metrics.Float64Counter = (*float64Counter)(nil)
var _ metrics.Float64UpDownCounter = (*float64Counter)(nil)

func (i *float64Counter) Add(ctx context.Context, v float64, opts ...metrics.RecordMetricOption) {
	i.otel.Add(ctx, v, withMetricProps(opts...))
}

type float64Gauge struct {
	otel otelmetric.Float64Gauge
}

var _ metrics.Float64Gauge = (*float64Gauge)(nil)

func (i *float64Gauge) Sample(ctx context.Context, v float64, opts ...metrics.RecordMetricOption) {
	i.otel.Record(ctx, v, withMetricProps(opts...))
}

type float64Histogram struct {
	otel otelmetric.Float64Histogram
}

var _ metrics.Float64Histogram = (*float64Histogram)(nil)

func (i *float64Histogram) Record(ctx context.Context, v float64, opts ...metrics.RecordMetricOption) {
	i.otel.Record(ctx, v, withMetricProps(opts...))
}
