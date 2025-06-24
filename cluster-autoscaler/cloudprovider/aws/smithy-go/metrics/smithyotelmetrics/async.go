package smithyotelmetrics

import (
	"context"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/aws/smithy-go/metrics"
	otelmetric "go.opentelemetry.io/otel/metric"
)

type asyncInstrument struct {
	otel otelmetric.Registration
}

var _ metrics.AsyncInstrument = (*asyncInstrument)(nil)

func (i *asyncInstrument) Stop() {
	i.otel.Unregister()
}

// int64Observer wraps an untyped, multi-instrument OTEL Observer to Observe()
// against a single int64 instrument.
type int64Observer struct {
	observer   otelmetric.Observer
	instrument otelmetric.Int64Observable
}

var _ metrics.Int64Observer = (*int64Observer)(nil)

func (o *int64Observer) Observe(_ context.Context, v int64, opts ...metrics.RecordMetricOption) {
	o.observer.ObserveInt64(o.instrument, v, withMetricProps(opts...))
}

// adaptInt64CB wraps an OTEL async instrument callback, binding it to a single
// int64 instrument.
func adaptInt64CB(io otelmetric.Int64Observable, cb metrics.Int64Callback) otelmetric.Callback {
	return func(ctx context.Context, o otelmetric.Observer) error {
		cb(ctx, &int64Observer{o, io})
		return nil
	}
}

// float64Observer wraps an untyped, multi-instrument OTEL Observer to Observe()
// against a single float64 instrument.
type float64Observer struct {
	observer   otelmetric.Observer
	instrument otelmetric.Float64Observable
}

var _ metrics.Float64Observer = (*float64Observer)(nil)

func (o *float64Observer) Observe(_ context.Context, v float64, opts ...metrics.RecordMetricOption) {
	o.observer.ObserveFloat64(o.instrument, v, withMetricProps(opts...))
}

// adaptFloat64CB wraps an OTEL async instrument callback, binding it to a single
// float64 instrument.
func adaptFloat64CB(io otelmetric.Float64Observable, cb metrics.Float64Callback) otelmetric.Callback {
	return func(ctx context.Context, o otelmetric.Observer) error {
		cb(ctx, &float64Observer{o, io})
		return nil
	}
}
