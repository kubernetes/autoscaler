package smithyoteltracing

import (
	"context"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/aws/smithy-go/tracing"
	otelcodes "go.opentelemetry.io/otel/codes"
	oteltrace "go.opentelemetry.io/otel/trace"
)

// Adapt wraps a concrete OpenTelemetry SDK TraceProvider for use with Smithy
// SDK clients.
//
// Adapt can be called multiple times on a single TracerProvider.
func Adapt(tp oteltrace.TracerProvider) tracing.TracerProvider {
	return &tracerProvider{tp}
}

type tracerProvider struct {
	otel oteltrace.TracerProvider
}

var _ tracing.TracerProvider = (*tracerProvider)(nil)

func (p *tracerProvider) Tracer(scope string, opts ...tracing.TracerOption) tracing.Tracer {
	var options tracing.TracerOptions
	for _, opt := range opts {
		opt(&options)
	}

	t := p.otel.Tracer(scope, oteltrace.WithInstrumentationAttributes(
		toOTELKeyValues(options.Properties)...,
	))
	return &tracer{t}
}

type tracer struct {
	otel oteltrace.Tracer
}

var _ tracing.Tracer = (*tracer)(nil)

func (t *tracer) StartSpan(ctx context.Context, name string, opts ...tracing.SpanOption) (context.Context, tracing.Span) {
	// We do some context value juggling with our adapted Span to ensure the
	// following:
	//   (1) Our adapted Span is what actually persists on the context and
	//       is what callers are getting and recording to.
	//   (2) OTEL itself sees any pre-existing Span such that the parent-child
	//       relationship of concrete OTEL spans is maintained.
	ours, ok := tracing.GetSpan(ctx)
	if ok {
		ctx = oteltrace.ContextWithSpan(ctx, ours.(*span).otel) // (2)
	}

	var options tracing.SpanOptions
	for _, opt := range opts {
		opt(&options)
	}

	kind := toOTELSpanKind(options.Kind)
	ctx, theirs := t.otel.Start(ctx, name, oteltrace.WithSpanKind(kind))

	ours = &span{
		otel: theirs,
		name: name,
	}
	for k, v := range options.Properties.Values() {
		ours.SetProperty(k, v)
	}
	return tracing.WithSpan(ctx, ours) /* (1) */, ours
}

type span struct {
	otel oteltrace.Span
	name string
}

var _ tracing.Span = (*span)(nil)

func (s *span) Name() string {
	return s.name
}

func (s *span) Context() tracing.SpanContext {
	ctx := s.otel.SpanContext()
	return tracing.SpanContext{
		TraceID:  ctx.TraceID().String(),
		SpanID:   ctx.SpanID().String(),
		IsRemote: ctx.IsRemote(),
	}
}

func (s *span) AddEvent(name string, opts ...tracing.EventOption) {
	var options tracing.EventOptions
	for _, opt := range opts {
		opt(&options)
	}

	s.otel.AddEvent(name, oteltrace.WithAttributes(
		toOTELKeyValues(options.Properties)...,
	))
}

func (s *span) SetProperty(k, v any) {
	s.otel.SetAttributes(toOTELKeyValue(k, v))
}

func (s *span) SetStatus(status tracing.SpanStatus) {
	s.otel.SetStatus(toOTELSpanStatus(status), "")
}

func (s *span) End() {
	s.otel.End()
}

func toOTELSpanKind(v tracing.SpanKind) oteltrace.SpanKind {
	switch v {
	case tracing.SpanKindClient:
		return oteltrace.SpanKindClient
	case tracing.SpanKindServer:
		return oteltrace.SpanKindServer
	case tracing.SpanKindProducer:
		return oteltrace.SpanKindProducer
	case tracing.SpanKindConsumer:
		return oteltrace.SpanKindConsumer
	default:
		return oteltrace.SpanKindInternal
	}
}

func toOTELSpanStatus(v tracing.SpanStatus) otelcodes.Code {
	switch v {
	case tracing.SpanStatusOK:
		return otelcodes.Ok
	case tracing.SpanStatusError:
		return otelcodes.Error
	default:
		return otelcodes.Unset
	}
}
