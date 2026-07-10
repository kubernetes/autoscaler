/*
Copyright 2023 The Kubernetes Authors.

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

package metrics

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"go.opentelemetry.io/otel/attribute"
	api "go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/noop"

	"sigs.k8s.io/cloud-provider-azure/pkg/azclient/policy/ratelimit"
	"sigs.k8s.io/cloud-provider-azure/pkg/azclient/policy/retryrepectthrottled"
)

var (
	armRequestLatency    api.Float64Histogram
	armRequestErrors     api.Int64Counter
	armRequestRateLimits api.Int64Counter
	armRequestThrottles  api.Int64Counter
)

// ARMContext is the context for ARM metrics.
type ARMContext struct {
	startedAt  time.Time
	attributes []attribute.KeyValue
}

// BeginARMRequest creates a new ARMContext for an ARM request.
func BeginARMRequest(subscriptionID, resourceGroup, resource, method string) *ARMContext {
	return BeginARMRequestWithAttributes(
		attribute.String("subscription_id", subscriptionID),
		attribute.String("resource_group", strings.ToLower(resourceGroup)),
		attribute.String("resource", resource),
		attribute.String("method", method),
	)
}

// BeginARMRequest creates a new ARMContext for an ARM request.
func BeginARMRequestWithAttributes(attributes ...attribute.KeyValue) *ARMContext {
	return &ARMContext{
		startedAt:  time.Now(),
		attributes: attributes,
	}
}

// Observe observes the result of the ARM request.
// It's a convenience method for calling Done, Errored, RateLimited, or Throttled.
// You should not call this method after calling one of the other methods.
func (c *ARMContext) Observe(ctx context.Context, err error) {
	if err == nil {
		c.Done(ctx)
		return
	}

	if errors.Is(err, ratelimit.ErrRateLimitReached) {
		c.RateLimited(ctx)
		return
	}

	if errors.Is(err, retryrepectthrottled.ErrTooManyRequest) {
		c.Throttled(ctx)
		return
	}

	c.Errored(ctx, err)
}

// Done finishes the ARMContext and records the latency.
func (c *ARMContext) Done(ctx context.Context) {
	elapsed := time.Since(c.startedAt).Seconds()
	ARMRequestLatency().Record(ctx, elapsed, api.WithAttributes(c.attributes...))
}

// Errored finishes the ARMContext and records the error.
func (c *ARMContext) Errored(ctx context.Context, err error) {
	c.Done(ctx)
	var respErr *azcore.ResponseError
	if errors.As(err, &respErr) {
		attributes := append(c.attributes,
			attribute.Int("status_code", respErr.StatusCode),
			attribute.String("error_code", respErr.ErrorCode),
		)
		ARMRequestErrors().Add(ctx, 1, api.WithAttributes(attributes...))
	} else {
		ARMRequestErrors().Add(ctx, 1, api.WithAttributes(c.attributes...)) // error without status code
	}
}

// RateLimited finishes the ARMContext and records the rate limit.
func (c *ARMContext) RateLimited(ctx context.Context) {
	c.Done(ctx)
	ARMRequestRateLimits().Add(ctx, 1, api.WithAttributes(c.attributes...))
}

// Throttled finishes the ARMContext and records the throttle.
func (c *ARMContext) Throttled(ctx context.Context) {
	c.Done(ctx)
	ARMRequestThrottles().Add(ctx, 1, api.WithAttributes(c.attributes...))
}

// ARMRequestLatency returns the histogram for ARM request latency.
func ARMRequestLatency() api.Float64Histogram {
	if armRequestLatency == nil {
		return noop.Float64Histogram{}
	}
	return armRequestLatency
}

// ARMRequestErrors returns the counter for ARM request errors.
func ARMRequestErrors() api.Int64Counter {
	if armRequestErrors == nil {
		return noop.Int64Counter{}
	}
	return armRequestErrors
}

// ARMRequestRateLimits returns the counter for ARM request rate limits.
func ARMRequestRateLimits() api.Int64Counter {
	if armRequestRateLimits == nil {
		return noop.Int64Counter{}
	}
	return armRequestRateLimits
}

// ARMRequestThrottles returns the counter for ARM request throttles.
func ARMRequestThrottles() api.Int64Counter {
	if armRequestThrottles == nil {
		return noop.Int64Counter{}
	}
	return armRequestThrottles
}

// Setup sets up the ARM metrics.
func Setup(meter api.Meter) error {
	setups := []func(api.Meter) error{
		setupARMRequestLatency,
		setupARMRequestErrors,
		setupARMRequestRateLimits,
		setupARMRequestThrottles,
	}

	for _, setup := range setups {
		if err := setup(meter); err != nil {
			return fmt.Errorf("setup azclient metrics: %w", err)
		}
	}
	return nil
}

func setupARMRequestLatency(meter api.Meter) error {
	m, err := meter.Float64Histogram(
		"arm.request.duration",
		api.WithUnit("s"),
		api.WithDescription("Measures the duration of Azure ARM API calls."),
		api.WithExplicitBucketBoundaries(.1, .25, .5, 1, 2.5, 5, 10, 60, 300, 600),
	)

	if err != nil {
		return fmt.Errorf("create arm.request.duration histogram: %w", err)
	}

	armRequestLatency = m

	return nil
}

func setupARMRequestErrors(meter api.Meter) error {
	c, err := meter.Int64Counter(
		"arm.request.errors.counter",
		api.WithDescription("Measures the number of errors in Azure ARM API calls."),
	)

	if err != nil {
		return fmt.Errorf("create arm.request.errors.counter counter: %w", err)
	}

	armRequestErrors = c

	return nil
}

func setupARMRequestRateLimits(meter api.Meter) error {
	c, err := meter.Int64Counter(
		"arm.request.rate_limit.counter",
		api.WithDescription("Measures the number of rate-limited Azure ARM API calls."),
	)

	if err != nil {
		return fmt.Errorf("create arm.request.rate_limit.counter counter: %w", err)
	}

	armRequestRateLimits = c

	return nil
}

func setupARMRequestThrottles(meter api.Meter) error {
	c, err := meter.Int64Counter(
		"arm.request.throttle.counter",
		api.WithDescription("Measures the number of throttled Azure ARM API calls."),
	)

	if err != nil {
		return fmt.Errorf("create arm.request.throttle.counter counter: %w", err)
	}

	armRequestThrottles = c

	return nil
}
