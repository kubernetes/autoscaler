/*
Copyright 2018 Google LLC

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

https://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cloud

import (
	"context"
	"time"
)

// RateLimitKey is a key identifying the operation to be rate limited. The rate limit
// queue will be determined based on the contents of RateKey.
//
// This type will be removed in a future release. Please change all
// references to CallContextKey.
type RateLimitKey = CallContextKey

// RateLimiter is the interface for a rate limiting policy.
type RateLimiter interface {
	// Accept uses the RateLimitKey to derive a sleep time for the calling
	// goroutine. This call will block until the operation is ready for
	// execution.
	//
	// Accept returns an error if the given context ctx was canceled
	// while waiting for acceptance into the queue.
	Accept(ctx context.Context, key *RateLimitKey) error
	// Observe uses the RateLimitKey to handle response results, which may affect
	// the sleep time for the Accept function.
	Observe(ctx context.Context, err error, key *RateLimitKey)
}

// acceptor is an object which blocks within Accept until a call is allowed to run.
// Accept is a behavior of the flowcontrol.RateLimiter interface.
type acceptor interface {
	// Accept blocks until a call is allowed to run.
	Accept()
}

// AcceptRateLimiter wraps an Acceptor with RateLimiter parameters.
type AcceptRateLimiter struct {
	// Acceptor is the underlying rate limiter.
	Acceptor acceptor
}

// Accept wraps an Acceptor and blocks on Accept or context.Done(). Key is ignored.
func (rl *AcceptRateLimiter) Accept(ctx context.Context, _ *RateLimitKey) error {
	ch := make(chan struct{})
	go func() {
		rl.Acceptor.Accept()
		close(ch)
	}()
	select {
	case <-ch:
		break
	case <-ctx.Done():
		return ctx.Err()
	}
	return nil
}

// Observe does nothing.
func (rl *AcceptRateLimiter) Observe(context.Context, error, *RateLimitKey) {
}

// NopRateLimiter is a rate limiter that performs no rate limiting.
type NopRateLimiter struct {
}

// Accept everything immediately.
func (*NopRateLimiter) Accept(context.Context, *RateLimitKey) error {
	return nil
}

// Observe does nothing.
func (*NopRateLimiter) Observe(context.Context, error, *RateLimitKey) {
}

// MinimumRateLimiter wraps a RateLimiter and will only call its Accept until the minimum
// duration has been met or the context is cancelled.
type MinimumRateLimiter struct {
	// RateLimiter is the underlying ratelimiter which is called after the mininum time is reacehd.
	RateLimiter RateLimiter
	// Minimum is the minimum wait time before the underlying ratelimiter is called.
	Minimum time.Duration
}

// Accept blocks on the minimum duration and context. Once the minimum duration is met,
// the func is blocked on the underlying ratelimiter.
func (m *MinimumRateLimiter) Accept(ctx context.Context, key *RateLimitKey) error {
	select {
	case <-time.After(m.Minimum):
		return m.RateLimiter.Accept(ctx, key)
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Observe just passes error to the underlying ratelimiter.
func (m *MinimumRateLimiter) Observe(ctx context.Context, err error, key *RateLimitKey) {
	m.RateLimiter.Observe(ctx, err, key)
}
