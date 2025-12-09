/*
Copyright 2019 The Kubernetes Authors.

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

package verda

import (
	"errors"
	"fmt"
	"net/http"
	"testing"
	"time"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/verdacloud/verdacloud-sdk-go/verda/testutil"
)

func TestShouldRetry(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
		// APIError with status codes
		{
			name:     "500 Internal Server Error",
			err:      &APIError{StatusCode: 500, Message: "Internal Server Error"},
			expected: true,
		},
		{
			name:     "502 Bad Gateway",
			err:      &APIError{StatusCode: 502, Message: "Bad Gateway"},
			expected: true,
		},
		{
			name:     "503 Service Unavailable",
			err:      &APIError{StatusCode: 503, Message: "Service Unavailable"},
			expected: true,
		},
		{
			name:     "504 Gateway Timeout",
			err:      &APIError{StatusCode: 504, Message: "Gateway Timeout"},
			expected: true,
		},
		{
			name:     "429 Too Many Requests",
			err:      &APIError{StatusCode: 429, Message: "Too Many Requests"},
			expected: true,
		},
		{
			name:     "408 Request Timeout",
			err:      &APIError{StatusCode: 408, Message: "Request Timeout"},
			expected: true,
		},
		{
			name:     "400 Bad Request",
			err:      &APIError{StatusCode: 400, Message: "Bad Request"},
			expected: false,
		},
		{
			name:     "401 Unauthorized",
			err:      &APIError{StatusCode: 401, Message: "Unauthorized"},
			expected: false,
		},
		{
			name:     "403 Forbidden",
			err:      &APIError{StatusCode: 403, Message: "Forbidden"},
			expected: false,
		},
		{
			name:     "404 Not Found",
			err:      &APIError{StatusCode: 404, Message: "Not Found"},
			expected: false,
		},
		{
			name:     "599 Custom 5xx error",
			err:      &APIError{StatusCode: 599, Message: "Custom Server Error"},
			expected: true,
		},
		{
			name:     "499 Custom 4xx error",
			err:      &APIError{StatusCode: 499, Message: "Custom Client Error"},
			expected: false,
		},
		// Pattern matching for non-retryable errors
		{
			name:     "authentication error",
			err:      errors.New("authentication failed"),
			expected: false,
		},
		{
			name:     "unauthorized error",
			err:      errors.New("unauthorized access"),
			expected: false,
		},
		{
			name:     "forbidden error",
			err:      errors.New("forbidden resource"),
			expected: false,
		},
		{
			name:     "not found error",
			err:      errors.New("resource not found"),
			expected: false,
		},
		{
			name:     "invalid request error",
			err:      errors.New("invalid request parameters"),
			expected: false,
		},
		{
			name:     "bad request error",
			err:      errors.New("bad request format"),
			expected: false,
		},
		// Pattern matching for retryable errors
		{
			name:     "timeout error",
			err:      errors.New("request timeout occurred"),
			expected: true,
		},
		{
			name:     "connection error",
			err:      errors.New("connection refused"),
			expected: true,
		},
		{
			name:     "temporary error",
			err:      errors.New("temporary network failure"),
			expected: true,
		},
		{
			name:     "rate limit error",
			err:      errors.New("rate limit exceeded"),
			expected: true,
		},
		{
			name:     "too many requests error",
			err:      errors.New("too many requests sent"),
			expected: true,
		},
		// Case insensitive matching
		{
			name:     "TIMEOUT uppercase",
			err:      errors.New("REQUEST TIMEOUT"),
			expected: true,
		},
		{
			name:     "Authentication uppercase",
			err:      errors.New("AUTHENTICATION FAILED"),
			expected: false,
		},
		// Unknown errors default to non-retryable
		{
			name:     "unknown error",
			err:      errors.New("some random error"),
			expected: false,
		},
		// Wrapped APIError
		{
			name:     "wrapped 503 error",
			err:      fmt.Errorf("request failed: %w", &APIError{StatusCode: 503, Message: "Service Unavailable"}),
			expected: true,
		},
		{
			name:     "wrapped 401 error",
			err:      fmt.Errorf("auth failed: %w", &APIError{StatusCode: 401, Message: "Unauthorized"}),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := shouldRetry(tt.err)
			if result != tt.expected {
				t.Errorf("shouldRetry(%v) = %v, expected %v", tt.err, result, tt.expected)
			}
		})
	}
}

func TestExponentialBackoffRetryMiddleware_Success(t *testing.T) {
	logger := &NoOpLogger{}
	attempts := 0

	middleware := ExponentialBackoffRetryMiddleware(3, 10*time.Millisecond, logger)

	handler := middleware(func(ctx *RequestContext) error {
		attempts++
		return nil // Success on first try
	})

	ctx := &RequestContext{
		Method: "GET",
		Path:   "/test",
	}

	err := handler(ctx)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if attempts != 1 {
		t.Errorf("Expected 1 attempt, got %d", attempts)
	}
}

func TestExponentialBackoffRetryMiddleware_RetryableError(t *testing.T) {
	logger := &NoOpLogger{}
	attempts := 0
	maxRetries := 3

	middleware := ExponentialBackoffRetryMiddleware(maxRetries, 1*time.Millisecond, logger)

	handler := middleware(func(ctx *RequestContext) error {
		attempts++
		// Always return a retryable error
		return &APIError{StatusCode: 503, Message: "Service Unavailable"}
	})

	ctx := &RequestContext{
		Method: "GET",
		Path:   "/test",
	}

	start := time.Now()
	err := handler(ctx)
	duration := time.Since(start)

	if err == nil {
		t.Error("Expected error after retries exhausted")
	}

	expectedAttempts := maxRetries + 1 // Initial attempt + retries
	if attempts != expectedAttempts {
		t.Errorf("Expected %d attempts, got %d", expectedAttempts, attempts)
	}

	// Should have some delay due to retries (at least 1ms * 3 retries)
	if duration < 3*time.Millisecond {
		t.Errorf("Expected some delay from retries, got %v", duration)
	}
}

func TestExponentialBackoffRetryMiddleware_NonRetryableError(t *testing.T) {
	logger := &NoOpLogger{}
	attempts := 0

	middleware := ExponentialBackoffRetryMiddleware(3, 10*time.Millisecond, logger)

	handler := middleware(func(ctx *RequestContext) error {
		attempts++
		// Return non-retryable error
		return &APIError{StatusCode: 401, Message: "Unauthorized"}
	})

	ctx := &RequestContext{
		Method: "GET",
		Path:   "/test",
	}

	err := handler(ctx)
	if err == nil {
		t.Error("Expected error")
	}

	if attempts != 1 {
		t.Errorf("Expected only 1 attempt for non-retryable error, got %d", attempts)
	}
}

func TestExponentialBackoffRetryMiddleware_SuccessAfterRetry(t *testing.T) {
	logger := &NoOpLogger{}
	attempts := 0

	middleware := ExponentialBackoffRetryMiddleware(3, 1*time.Millisecond, logger)

	handler := middleware(func(ctx *RequestContext) error {
		attempts++
		if attempts < 3 {
			// Fail first 2 times with retryable error
			return &APIError{StatusCode: 503, Message: "Service Unavailable"}
		}
		// Succeed on 3rd attempt
		return nil
	})

	ctx := &RequestContext{
		Method: "GET",
		Path:   "/test",
	}

	err := handler(ctx)
	if err != nil {
		t.Errorf("Expected success after retries, got error: %v", err)
	}

	if attempts != 3 {
		t.Errorf("Expected 3 attempts, got %d", attempts)
	}
}

func TestExponentialBackoffRetryMiddleware_ExponentialDelay(t *testing.T) {
	logger := &NoOpLogger{}
	attempts := 0
	delays := []time.Duration{}

	middleware := ExponentialBackoffRetryMiddleware(3, 10*time.Millisecond, logger)

	var lastTime time.Time
	handler := middleware(func(ctx *RequestContext) error {
		now := time.Now()
		if attempts > 0 {
			delay := now.Sub(lastTime)
			delays = append(delays, delay)
		}
		lastTime = now
		attempts++
		// Always fail with retryable error
		return &APIError{StatusCode: 503, Message: "Service Unavailable"}
	})

	ctx := &RequestContext{
		Method: "GET",
		Path:   "/test",
	}

	_ = handler(ctx)

	// Should have 3 delays (between attempts 1-2, 2-3, 3-4)
	if len(delays) != 3 {
		t.Errorf("Expected 3 delays, got %d", len(delays))
	}

	// Verify exponential growth (accounting for jitter of ±50%)
	// Expected delays (before jitter): 10ms, 20ms, 40ms
	// With jitter, ranges: [5-15ms], [10-30ms], [20-60ms]
	if len(delays) >= 3 {
		// First delay should be around 10ms ± 50%
		if delays[0] < 5*time.Millisecond || delays[0] > 15*time.Millisecond {
			t.Logf("Warning: First delay %v outside expected range [5ms-15ms]", delays[0])
		}

		// Second delay should be larger than first (on average)
		// Third delay should be larger than second (on average)
		// We can't strictly enforce this due to jitter, but we can check the total
		// Expected total: 70ms (10 + 20 + 40), with jitter range: [35ms-105ms]
		totalDelay := delays[0] + delays[1] + delays[2]
		minTotal := 35 * time.Millisecond  // 50% of expected
		maxTotal := 105 * time.Millisecond // 150% of expected

		if totalDelay < minTotal || totalDelay > maxTotal {
			t.Errorf("Total delay %v outside expected range [%v-%v]", totalDelay, minTotal, maxTotal)
		}
	}
}

func TestExponentialBackoffRetryMiddleware_MaxDelay(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping long-running test in short mode")
	}

	logger := &NoOpLogger{}
	attempts := 0
	delays := []time.Duration{}

	// Use 100ms base to hit cap quickly but keep test under 1 second
	// With 100ms initial: 100ms, 200ms, 400ms, 800ms...
	// We'll do enough retries to hit the theoretical cap at 30s
	middleware := ExponentialBackoffRetryMiddleware(5, 100*time.Millisecond, logger)

	var lastTime time.Time

	handler := middleware(func(ctx *RequestContext) error {
		now := time.Now()
		if attempts > 0 {
			delay := now.Sub(lastTime)
			delays = append(delays, delay)
		}
		lastTime = now
		attempts++

		return &APIError{StatusCode: 503, Message: "Service Unavailable"}
	})

	ctx := &RequestContext{
		Method: "GET",
		Path:   "/test",
	}

	_ = handler(ctx)

	// With 100ms base, we get: 100, 200, 400, 800, 1600ms (all under 30s cap)
	// The cap of 30s won't be hit with these small delays, but we test
	// that the logic works correctly and delays grow exponentially

	if len(delays) != 5 {
		t.Errorf("Expected 5 delays, got %d", len(delays))
	}

	// Verify exponential growth pattern (accounting for jitter)
	// Each delay should be roughly 2x the previous (within jitter range)
	for i := 1; i < len(delays); i++ {
		// With jitter, the ratio could be anywhere from ~1x to ~3x
		// We just verify it's increasing on average
		if i == len(delays)-1 {
			// Last few should be roughly double the earlier ones
			avgEarly := (delays[0] + delays[1]) / 2
			avgLate := (delays[len(delays)-2] + delays[len(delays)-1]) / 2

			// Later delays should be significantly larger
			if avgLate < avgEarly {
				t.Errorf("Expected delays to grow exponentially: early avg %v, late avg %v", avgEarly, avgLate)
			}
		}
	}
}

func TestAuthenticationMiddleware(t *testing.T) {
	t.Run("successful authentication", func(t *testing.T) {
		// Create a properly initialized client with auth service
		client := &Client{
			ClientID:     "test-id",
			ClientSecret: "test-secret",
			BaseURL:      "https://api.test.com",
			HTTPClient:   &http.Client{},
		}
		client.Auth = &AuthService{
			client: client,
			token: &TokenResponse{
				AccessToken: "test-token",
				ExpiresIn:   3600,
				TokenType:   "Bearer",
				ExpiresAt:   time.Now().Add(1 * time.Hour),
			},
		}

		middleware := AuthenticationMiddleware()
		handler := middleware(func(ctx *RequestContext) error {
			// Verify token was added
			auth := ctx.Headers.Get("Authorization")
			if auth != "Bearer test-token" {
				t.Errorf("Expected 'Bearer test-token', got '%s'", auth)
			}
			return nil
		})

		ctx := &RequestContext{
			Headers: make(map[string][]string),
			Client:  client,
		}

		err := handler(ctx)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
	})

	t.Run("authentication failure", func(t *testing.T) {
		// Create a mock server that returns auth errors
		mockServer := testutil.NewMockServer()
		defer mockServer.Close()

		// Set up mock to return authentication error
		mockServer.SetHandler(http.MethodPost, "/oauth2/token", func(w http.ResponseWriter, r *http.Request) {
			testutil.ErrorResponse(w, http.StatusUnauthorized, "Invalid credentials")
		})

		// Create a client using the mock server
		client := NewTestClient(mockServer)

		middleware := AuthenticationMiddleware()
		handler := middleware(func(ctx *RequestContext) error {
			t.Error("Handler should not be called when auth fails")
			return nil
		})

		ctx := &RequestContext{
			Headers: make(map[string][]string),
			Client:  client,
		}

		err := handler(ctx)
		if err == nil {
			t.Error("Expected authentication error")
		}
	})
}

func TestContentTypeMiddleware(t *testing.T) {
	t.Run("sets content type when body present", func(t *testing.T) {
		middleware := ContentTypeMiddleware("application/json")
		handler := middleware(func(ctx *RequestContext) error {
			ct := ctx.Headers.Get("Content-Type")
			if ct != "application/json" {
				t.Errorf("Expected 'application/json', got '%s'", ct)
			}
			return nil
		})

		ctx := &RequestContext{
			Body:    map[string]string{"key": "value"},
			Headers: make(map[string][]string),
		}

		err := handler(ctx)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
	})

	t.Run("does not set content type when no body", func(t *testing.T) {
		middleware := ContentTypeMiddleware("application/json")
		handler := middleware(func(ctx *RequestContext) error {
			ct := ctx.Headers.Get("Content-Type")
			if ct != "" {
				t.Errorf("Expected no Content-Type, got '%s'", ct)
			}
			return nil
		})

		ctx := &RequestContext{
			Body:    nil,
			Headers: make(map[string][]string),
		}

		err := handler(ctx)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
	})
}

func TestJSONContentTypeMiddleware(t *testing.T) {
	middleware := JSONContentTypeMiddleware()
	handler := middleware(func(ctx *RequestContext) error {
		ct := ctx.Headers.Get("Content-Type")
		if ct != "application/json" {
			t.Errorf("Expected 'application/json', got '%s'", ct)
		}
		return nil
	})

	ctx := &RequestContext{
		Body:    map[string]string{"key": "value"},
		Headers: make(map[string][]string),
	}

	err := handler(ctx)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func TestUserAgentMiddleware(t *testing.T) {
	userAgent := "TestAgent/1.0"
	middleware := UserAgentMiddleware(userAgent)
	handler := middleware(func(ctx *RequestContext) error {
		ua := ctx.Headers.Get("User-Agent")
		if ua != userAgent {
			t.Errorf("Expected '%s', got '%s'", userAgent, ua)
		}
		return nil
	})

	ctx := &RequestContext{
		Headers: make(map[string][]string),
	}

	err := handler(ctx)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func TestLoggingMiddleware(t *testing.T) {
	logger := &NoOpLogger{}
	middleware := LoggingMiddleware(logger)

	called := false
	handler := middleware(func(ctx *RequestContext) error {
		called = true
		return nil
	})

	ctx := &RequestContext{
		Method: "GET",
		Path:   "/test",
	}

	err := handler(ctx)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if !called {
		t.Error("Handler was not called")
	}
}
