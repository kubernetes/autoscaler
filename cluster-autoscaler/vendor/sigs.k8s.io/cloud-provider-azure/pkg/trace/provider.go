/*
Copyright 2024 The Kubernetes Authors.

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

package trace

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	prometheusexporter "go.opentelemetry.io/otel/exporters/prometheus"
	apimetric "go.opentelemetry.io/otel/metric"
	metricnoop "go.opentelemetry.io/otel/metric/noop"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	apitrace "go.opentelemetry.io/otel/trace"
	tracenoop "go.opentelemetry.io/otel/trace/noop"
	"k8s.io/component-base/metrics/legacyregistry"
)

const (
	DefaultTracerName = "cloud-provider-azure"
	DefaultMeterName  = "cloud-provider-azure"
)

var (
	globalProvider *Provider
)

func SetGlobalProvider(p *Provider) {
	globalProvider = p
}

func DefaultTracer() apitrace.Tracer {
	if globalProvider == nil {
		return tracenoop.Tracer{}
	}
	return globalProvider.defaultTracer
}

func DefaultMeter() apimetric.Meter {
	if globalProvider == nil {
		return metricnoop.Meter{}
	}
	return globalProvider.defaultMeter
}

type Provider struct {
	traceProvider *sdktrace.TracerProvider
	defaultTracer apitrace.Tracer

	meterProvider      *sdkmetric.MeterProvider
	defaultMeter       apimetric.Meter
	prometheusRegistry *prometheus.Registry
}

func New() (*Provider, error) {
	// TODO: support tracing
	var (
		traceProvider = sdktrace.NewTracerProvider(
			sdktrace.WithSampler(sdktrace.NeverSample()),
		)
		defaultTracer = traceProvider.Tracer(DefaultTracerName)
	)

	var (
		meterProvider      *sdkmetric.MeterProvider
		prometheusRegistry = prometheus.NewRegistry()
	)
	{
		exporter, err := prometheusexporter.New(
			prometheusexporter.WithRegisterer(prometheusRegistry),
		)
		if err != nil {
			return nil, fmt.Errorf("initialize prometheus exporter: %w", err)
		}
		meterProvider = sdkmetric.NewMeterProvider(sdkmetric.WithReader(exporter))
	}

	defaultMeter := meterProvider.Meter(DefaultMeterName)

	return &Provider{
		traceProvider:      traceProvider,
		defaultTracer:      defaultTracer,
		meterProvider:      meterProvider,
		defaultMeter:       defaultMeter,
		prometheusRegistry: prometheusRegistry,
	}, nil
}

// TracerProvider returns the tracer provider.
func (p *Provider) TracerProvider() *sdktrace.TracerProvider {
	return p.traceProvider
}

// MeterProvider returns the meter provider.
func (p *Provider) MeterProvider() *sdkmetric.MeterProvider {
	return p.meterProvider
}

// DefaultMeter returns the default meter.
func (p *Provider) DefaultMeter() apimetric.Meter {
	return p.defaultMeter
}

// MetricsHTTPHandler returns an HTTP handler for local Prometheus metrics.
func (p *Provider) MetricsHTTPHandler() http.Handler {
	return promhttp.InstrumentMetricHandler(
		p.prometheusRegistry,
		promhttp.HandlerFor(prometheus.Gatherers{
			p.prometheusRegistry,
			legacyregistry.DefaultGatherer, // Merge with the metrics from legacy registerer
		}, promhttp.HandlerOpts{}),
	)
}

// Stop stops the provider.
func (p *Provider) Stop(ctx context.Context) error {
	errC := make(chan error, 2)

	go func() {
		errC <- p.meterProvider.Shutdown(ctx)
	}()

	go func() {
		errC <- p.traceProvider.Shutdown(ctx)
	}()

	errs := make([]error, 0, 2)
	for i := 0; i < 2; i++ {
		if err := <-errC; err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	return nil
}
