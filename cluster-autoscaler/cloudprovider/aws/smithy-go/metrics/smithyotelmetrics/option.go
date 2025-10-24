package smithyotelmetrics

import (
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/aws/smithy-go/metrics"
	otelmetric "go.opentelemetry.io/otel/metric"
)

func toInstrumentOpts(opts ...metrics.InstrumentOption) (unit, desc string) {
	var o metrics.InstrumentOptions
	for _, opt := range opts {
		opt(&o)
	}
	return o.UnitLabel, o.Description
}

func withMetricProps(opts ...metrics.RecordMetricOption) otelmetric.MeasurementOption {
	var o metrics.RecordMetricOptions
	for _, opt := range opts {
		opt(&o)
	}
	return otelmetric.WithAttributes(toOTELKeyValues(o.Properties)...)

}
