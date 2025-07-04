package smithyotelmetrics

import (
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/aws/smithy-go/metrics"
	otelmetric "go.opentelemetry.io/otel/metric"
)

// Adapt wraps a concrete OpenTelemetry SDK MeterProvider for use with Smithy
// SDK clients.
//
// Adapt can be called multiple times on a single MeterProvider.
func Adapt(mp otelmetric.MeterProvider) metrics.MeterProvider {
	return &meterProvider{mp}
}

type meterProvider struct {
	otel otelmetric.MeterProvider
}

var _ metrics.MeterProvider = (*meterProvider)(nil)

func (p *meterProvider) Meter(scope string, opts ...metrics.MeterOption) metrics.Meter {
	var options metrics.MeterOptions
	for _, opt := range opts {
		opt(&options)
	}

	m := p.otel.Meter(scope, otelmetric.WithInstrumentationAttributes(
		toOTELKeyValues(options.Properties)...,
	))
	return &meter{m}
}

type meter struct {
	otel otelmetric.Meter
}

var _ metrics.Meter = (*meter)(nil)

func (m *meter) Int64Counter(name string, opts ...metrics.InstrumentOption) (metrics.Int64Counter, error) {
	unit, desc := toInstrumentOpts(opts...)
	i, err := m.otel.Int64Counter(name, otelmetric.WithUnit(unit), otelmetric.WithDescription(desc))
	if err != nil {
		return nil, err
	}
	return &int64Counter{i}, nil
}

func (m *meter) Int64UpDownCounter(name string, opts ...metrics.InstrumentOption) (metrics.Int64UpDownCounter, error) {
	unit, desc := toInstrumentOpts(opts...)
	i, err := m.otel.Int64UpDownCounter(name, otelmetric.WithUnit(unit), otelmetric.WithDescription(desc))
	if err != nil {
		return nil, err
	}
	return &int64Counter{i}, nil
}

func (m *meter) Int64Gauge(name string, opts ...metrics.InstrumentOption) (metrics.Int64Gauge, error) {
	unit, desc := toInstrumentOpts(opts...)
	i, err := m.otel.Int64Gauge(name, otelmetric.WithUnit(unit), otelmetric.WithDescription(desc))
	if err != nil {
		return nil, err
	}
	return &int64Gauge{i}, nil
}

func (m *meter) Int64Histogram(name string, opts ...metrics.InstrumentOption) (metrics.Int64Histogram, error) {
	unit, desc := toInstrumentOpts(opts...)
	i, err := m.otel.Int64Histogram(name, otelmetric.WithUnit(unit), otelmetric.WithDescription(desc))
	if err != nil {
		return nil, err
	}
	return &int64Histogram{i}, nil
}

func (m *meter) Int64AsyncCounter(name string, callback metrics.Int64Callback, opts ...metrics.InstrumentOption) (metrics.AsyncInstrument, error) {
	unit, desc := toInstrumentOpts(opts...)
	i, err := m.otel.Int64ObservableCounter(name, otelmetric.WithUnit(unit), otelmetric.WithDescription(desc))
	if err != nil {
		return nil, err
	}

	return m.registerAsyncInt64(i, callback)
}

func (m *meter) Int64AsyncUpDownCounter(name string, callback metrics.Int64Callback, opts ...metrics.InstrumentOption) (metrics.AsyncInstrument, error) {
	unit, desc := toInstrumentOpts(opts...)
	i, err := m.otel.Int64ObservableUpDownCounter(name, otelmetric.WithUnit(unit), otelmetric.WithDescription(desc))
	if err != nil {
		return nil, err
	}

	return m.registerAsyncInt64(i, callback)
}

func (m *meter) Int64AsyncGauge(name string, callback metrics.Int64Callback, opts ...metrics.InstrumentOption) (metrics.AsyncInstrument, error) {
	unit, desc := toInstrumentOpts(opts...)
	i, err := m.otel.Int64ObservableGauge(name, otelmetric.WithUnit(unit), otelmetric.WithDescription(desc))
	if err != nil {
		return nil, err
	}

	return m.registerAsyncInt64(i, callback)
}

func (m *meter) Float64Counter(name string, opts ...metrics.InstrumentOption) (metrics.Float64Counter, error) {
	unit, desc := toInstrumentOpts(opts...)
	i, err := m.otel.Float64Counter(name, otelmetric.WithUnit(unit), otelmetric.WithDescription(desc))
	if err != nil {
		return nil, err
	}
	return &float64Counter{i}, nil
}

func (m *meter) Float64UpDownCounter(name string, opts ...metrics.InstrumentOption) (metrics.Float64UpDownCounter, error) {
	unit, desc := toInstrumentOpts(opts...)
	i, err := m.otel.Float64UpDownCounter(name, otelmetric.WithUnit(unit), otelmetric.WithDescription(desc))
	if err != nil {
		return nil, err
	}
	return &float64Counter{i}, nil
}

func (m *meter) Float64Gauge(name string, opts ...metrics.InstrumentOption) (metrics.Float64Gauge, error) {
	unit, desc := toInstrumentOpts(opts...)
	i, err := m.otel.Float64Gauge(name, otelmetric.WithUnit(unit), otelmetric.WithDescription(desc))
	if err != nil {
		return nil, err
	}
	return &float64Gauge{i}, nil
}

func (m *meter) Float64Histogram(name string, opts ...metrics.InstrumentOption) (metrics.Float64Histogram, error) {
	unit, desc := toInstrumentOpts(opts...)
	i, err := m.otel.Float64Histogram(name, otelmetric.WithUnit(unit), otelmetric.WithDescription(desc))
	if err != nil {
		return nil, err
	}
	return &float64Histogram{i}, nil
}

func (m *meter) Float64AsyncCounter(name string, callback metrics.Float64Callback, opts ...metrics.InstrumentOption) (metrics.AsyncInstrument, error) {
	unit, desc := toInstrumentOpts(opts...)
	i, err := m.otel.Float64ObservableCounter(name, otelmetric.WithUnit(unit), otelmetric.WithDescription(desc))
	if err != nil {
		return nil, err
	}

	return m.registerAsyncFloat64(i, callback)
}

func (m *meter) Float64AsyncUpDownCounter(name string, callback metrics.Float64Callback, opts ...metrics.InstrumentOption) (metrics.AsyncInstrument, error) {
	unit, desc := toInstrumentOpts(opts...)
	i, err := m.otel.Float64ObservableUpDownCounter(name, otelmetric.WithUnit(unit), otelmetric.WithDescription(desc))
	if err != nil {
		return nil, err
	}

	return m.registerAsyncFloat64(i, callback)
}

func (m *meter) Float64AsyncGauge(name string, callback metrics.Float64Callback, opts ...metrics.InstrumentOption) (metrics.AsyncInstrument, error) {
	unit, desc := toInstrumentOpts(opts...)
	i, err := m.otel.Float64ObservableGauge(name, otelmetric.WithUnit(unit), otelmetric.WithDescription(desc))
	if err != nil {
		return nil, err
	}

	return m.registerAsyncFloat64(i, callback)
}

func (m *meter) registerAsyncInt64(i otelmetric.Int64Observable, cb metrics.Int64Callback) (metrics.AsyncInstrument, error) {
	r, err := m.otel.RegisterCallback(adaptInt64CB(i, cb), i)
	if err != nil {
		return nil, err
	}

	return &asyncInstrument{r}, nil
}

func (m *meter) registerAsyncFloat64(i otelmetric.Float64Observable, cb metrics.Float64Callback) (metrics.AsyncInstrument, error) {
	r, err := m.otel.RegisterCallback(adaptFloat64CB(i, cb), i)
	if err != nil {
		return nil, err
	}

	return &asyncInstrument{r}, nil
}
