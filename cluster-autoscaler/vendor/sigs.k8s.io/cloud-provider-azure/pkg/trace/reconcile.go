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
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	apimetric "go.opentelemetry.io/otel/metric"
	apitrace "go.opentelemetry.io/otel/trace"

	"sigs.k8s.io/cloud-provider-azure/pkg/trace/metrics"
)

// ReconcileSpan is a helper for tracing reconcile operations.
type ReconcileSpan struct {
	span       apitrace.Span
	startedAt  time.Time
	attributes []attribute.KeyValue
}

// BeginReconcile starts a new ReconcileSpan and returns the context with the span.
func BeginReconcile(ctx context.Context, tracer apitrace.Tracer, name string, attributes ...attribute.KeyValue) (context.Context, *ReconcileSpan) {
	ctx, span := tracer.Start(ctx, name, apitrace.WithAttributes(attributes...))

	return ctx, &ReconcileSpan{
		span:       span,
		startedAt:  time.Now(),
		attributes: append([]attribute.KeyValue{attribute.String("name", name)}, attributes...),
	}
}

func (c *ReconcileSpan) recordLatency(ctx context.Context) {
	elapsed := time.Since(c.startedAt).Seconds()
	metrics.ReconcileLatency().Record(ctx, elapsed, apimetric.WithAttributes(c.attributes...))
}

// Inner returns the inner span.
func (c *ReconcileSpan) Inner() apitrace.Span {
	return c.span
}

// Observe observes the result of the reconcile operation.
// It's a convenience method for calling Done, or Errored.
// You should not call this method after calling one of the other methods.
func (c *ReconcileSpan) Observe(ctx context.Context, err error) {
	if err == nil {
		c.Done(ctx)
		return
	}

	c.Errored(ctx, err)
}

// Done finishes the ReconcileSpan and records the latency.
func (c *ReconcileSpan) Done(ctx context.Context) {
	c.recordLatency(ctx)

	c.span.End()
}

// Errored finishes the ReconcileSpan and records the error.
func (c *ReconcileSpan) Errored(ctx context.Context, err error) {
	c.recordLatency(ctx)
	metrics.ReconcileErrors().Add(ctx, 1, apimetric.WithAttributes(c.attributes...))

	c.span.RecordError(err)
	c.span.SetStatus(codes.Error, err.Error())
	c.span.End()
}
