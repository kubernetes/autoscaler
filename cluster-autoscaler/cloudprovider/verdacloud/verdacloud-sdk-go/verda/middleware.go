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
	"bytes"
	"crypto/rand"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// RequestContext holds information about an HTTP request in the middleware chain.
type RequestContext struct {
	Method  string
	Path    string
	Body    interface{}
	Headers http.Header
	Query   url.Values
	Request *http.Request
	Client  *Client
}

// ResponseContext holds information about an HTTP response in the middleware chain.
type ResponseContext struct {
	Request    *RequestContext
	Response   *http.Response
	Body       []byte
	Error      error
	StatusCode int
}

// RequestMiddleware is a function that wraps a RequestHandler.
type RequestMiddleware func(next RequestHandler) RequestHandler

// ResponseMiddleware is a function that wraps a ResponseHandler.
type ResponseMiddleware func(next ResponseHandler) ResponseHandler

// RequestHandler handles a request context.
type RequestHandler func(ctx *RequestContext) error

// ResponseHandler handles a response context.
type ResponseHandler func(ctx *ResponseContext) error

// AuthenticationMiddleware returns a middleware that adds authentication headers.
func AuthenticationMiddleware() RequestMiddleware {
	return func(next RequestHandler) RequestHandler {
		return func(ctx *RequestContext) error {
			bearerToken, err := ctx.Client.Auth.GetBearerToken()
			if err != nil {
				return fmt.Errorf("failed to get authentication token: %w", err)
			}

			if bearerToken == "" {
				return fmt.Errorf("empty authentication token")
			}

			ctx.Headers.Set("Authorization", bearerToken)

			return next(ctx)
		}
	}
}

// ContentTypeMiddleware returns a middleware that sets the Content-Type header.
func ContentTypeMiddleware(contentType string) RequestMiddleware {
	return func(next RequestHandler) RequestHandler {
		return func(ctx *RequestContext) error {
			if ctx.Body != nil {
				ctx.Headers.Set("Content-Type", contentType)
			}
			return next(ctx)
		}
	}
}

// JSONContentTypeMiddleware returns a middleware that sets the Content-Type to application/json.
func JSONContentTypeMiddleware() RequestMiddleware {
	return ContentTypeMiddleware("application/json")
}

// cryptoRandFloat64 uses crypto/rand for jitter to prevent thundering herd on retries
func cryptoRandFloat64() float64 {
	var b [8]byte
	if _, err := rand.Read(b[:]); err != nil {
		return 0
	}
	// Convert random bytes to float64 in [0,1) using mantissa bits
	return float64(binary.BigEndian.Uint64(b[:])&((1<<53)-1)) / (1 << 53)
}

// LoggingMiddleware returns a middleware that logs request duration.
func LoggingMiddleware(logger Logger) RequestMiddleware {
	return func(next RequestHandler) RequestHandler {
		return func(ctx *RequestContext) error {
			start := time.Now()
			logger.Debug("Starting %s request to %s", ctx.Method, ctx.Path)

			err := next(ctx)

			duration := time.Since(start)
			if err != nil {
				logger.Debug("Request %s %s failed after %v: %v", ctx.Method, ctx.Path, duration, err)
			} else {
				logger.Debug("Request %s %s completed in %v", ctx.Method, ctx.Path, duration)
			}

			return err
		}
	}
}

// UserAgentMiddleware returns a middleware that sets the User-Agent header.
func UserAgentMiddleware(userAgent string) RequestMiddleware {
	return func(next RequestHandler) RequestHandler {
		return func(ctx *RequestContext) error {
			ctx.Headers.Set("User-Agent", userAgent)
			return next(ctx)
		}
	}
}

// ExponentialBackoffRetryMiddleware returns a middleware that retries failed requests with exponential backoff.
func ExponentialBackoffRetryMiddleware(maxRetries int, initialDelay time.Duration, logger Logger) RequestMiddleware {
	const maxDelay = 30 * time.Second
	const jitterPercent = 0.5

	return func(next RequestHandler) RequestHandler {
		return func(ctx *RequestContext) error {
			var lastErr error

			for attempt := 0; attempt <= maxRetries; attempt++ {
				if attempt > 0 {
					// Exponential backoff: initialDelay * 2^(attempt-1), capped at 30s
					baseDelay := float64(initialDelay) * math.Pow(2, float64(attempt-1))
					cappedDelay := time.Duration(math.Min(baseDelay, float64(maxDelay)))

					// Add ±50% jitter to avoid thundering herd
					jitter := (cryptoRandFloat64()*2 - 1) * jitterPercent
					actualDelay := time.Duration(float64(cappedDelay) * (1 + jitter))

					logger.Debug("Retrying request %s %s (attempt %d/%d) after %v",
						ctx.Method, ctx.Path, attempt+1, maxRetries+1, actualDelay)
					time.Sleep(actualDelay)
				}

				lastErr = next(ctx)
				if lastErr == nil {
					return nil
				}

				if !shouldRetry(lastErr) {
					logger.Debug("Request %s %s failed with non-retryable error: %v", ctx.Method, ctx.Path, lastErr)
					break
				}
			}

			return fmt.Errorf("request failed after %d retries: %w", maxRetries, lastErr)
		}
	}
}

// shouldRetry decides if we should retry based on status code - never retry auth/client errors
func shouldRetry(err error) bool {
	if err == nil {
		return false
	}

	var apiErr *APIError
	if errors.As(err, &apiErr) {
		statusCode := apiErr.StatusCode

		switch statusCode {
		case http.StatusInternalServerError,
			http.StatusBadGateway,
			http.StatusServiceUnavailable,
			http.StatusGatewayTimeout,
			http.StatusTooManyRequests,
			http.StatusRequestTimeout:
			return true
		case http.StatusBadRequest,
			http.StatusUnauthorized,
			http.StatusForbidden,
			http.StatusNotFound:
			return false
		default:
			if statusCode >= 500 && statusCode < 600 {
				return true
			}
			if statusCode >= 400 && statusCode < 500 {
				return false
			}
		}
	}

	errStr := strings.ToLower(err.Error())

	nonRetryablePatterns := []string{
		"authentication",
		"unauthorized",
		"forbidden",
		"not found",
		"invalid",
		"bad request",
	}
	for _, pattern := range nonRetryablePatterns {
		if strings.Contains(errStr, pattern) {
			return false
		}
	}

	retryablePatterns := []string{
		"timeout",
		"connection",
		"temporary",
		"rate limit",
		"too many requests",
	}
	for _, pattern := range retryablePatterns {
		if strings.Contains(errStr, pattern) {
			return true
		}
	}

	return false
}

// ErrorHandlingMiddleware returns a middleware that handles HTTP error responses.
func ErrorHandlingMiddleware() ResponseMiddleware {
	return func(next ResponseHandler) ResponseHandler {
		return func(ctx *ResponseContext) error {
			if ctx.StatusCode < 200 || ctx.StatusCode >= 300 {
				var apiError APIError
				if len(ctx.Body) > 0 {
					apiError = APIError{
						StatusCode: ctx.StatusCode,
						Message:    string(ctx.Body),
					}
				} else {
					apiError = APIError{
						StatusCode: ctx.StatusCode,
						Message:    http.StatusText(ctx.StatusCode),
					}
				}
				ctx.Error = &apiError
			}

			return next(ctx)
		}
	}
}

// ResponseLoggingMiddleware returns a middleware that logs response details.
func ResponseLoggingMiddleware(logger Logger) ResponseMiddleware {
	return func(next ResponseHandler) ResponseHandler {
		return func(ctx *ResponseContext) error {
			logger.Debug("Response for %s %s: Status %d, Body length: %d bytes",
				ctx.Request.Method, ctx.Request.Path, ctx.StatusCode, len(ctx.Body))

			if ctx.Error != nil {
				logger.Debug("Response error: %v", ctx.Error)
			}

			return next(ctx)
		}
	}
}

// MetricsMiddleware is a placeholder for collecting request/response metrics
func MetricsMiddleware(logger Logger) ResponseMiddleware {
	return func(next ResponseHandler) ResponseHandler {
		return func(ctx *ResponseContext) error {
			logger.Debug("Metrics: %s %s -> %d (%d bytes)",
				ctx.Request.Method, ctx.Request.Path, ctx.StatusCode, len(ctx.Body))

			return next(ctx)
		}
	}
}

// CacheMiddleware is a placeholder for response caching
func CacheMiddleware() ResponseMiddleware {
	return func(next ResponseHandler) ResponseHandler {
		return func(ctx *ResponseContext) error {
			return next(ctx)
		}
	}
}

// DebugLoggingMiddleware logs full request details - redacts auth headers and skips token endpoints
func DebugLoggingMiddleware(logger Logger) RequestMiddleware {
	return func(next RequestHandler) RequestHandler {
		return func(ctx *RequestContext) error {
			// Don't log sensitive token refresh endpoints
			if strings.Contains(ctx.Path, "/token") {
				return next(ctx)
			}

			logger.Info("=== API Request ===")
			logger.Info("Method: %s", ctx.Method)
			logger.Info("Path: %s", ctx.Path)

			if len(ctx.Query) > 0 {
				logger.Info("Query Parameters: %s", ctx.Query.Encode())
			}

			logger.Info("Headers:")
			for name, values := range ctx.Headers {
				if name == "Authorization" {
					logger.Info("  %s: [REDACTED]", name)
				} else {
					for _, value := range values {
						logger.Info("  %s: %s", name, value)
					}
				}
			}

			// Log request body if present
			if ctx.Request != nil && ctx.Request.Body != nil {
				// Read the body
				bodyBytes, err := io.ReadAll(ctx.Request.Body)
				if err != nil {
					logger.Info("Request Body: [error reading body: %v]", err)
				} else {
					// Restore the body for the actual request
					ctx.Request.Body = io.NopCloser(bytes.NewReader(bodyBytes))

					if len(bodyBytes) > 0 {
						bodyStr := string(bodyBytes)
						if len(bodyStr) > 1000 {
							logger.Info("Request Body (truncated): %s...", bodyStr[:1000])
						} else {
							logger.Info("Request Body: %s", bodyStr)
						}
					} else {
						logger.Info("Request Body: [empty]")
					}
				}
			} else {
				logger.Info("Request Body: [none]")
			}

			return next(ctx)
		}
	}
}

// DebugResponseLoggingMiddleware returns a middleware that logs detailed response information.
func DebugResponseLoggingMiddleware(logger Logger) ResponseMiddleware {
	return func(next ResponseHandler) ResponseHandler {
		return func(ctx *ResponseContext) error {
			if ctx.Request != nil && strings.Contains(ctx.Request.Path, "/token") {
				return next(ctx)
			}

			logger.Info("=== API Response ===")
			logger.Info("Status Code: %d", ctx.StatusCode)

			if ctx.Response != nil {
				logger.Info("Response Headers:")
				for name, values := range ctx.Response.Header {
					for _, value := range values {
						logger.Info("  %s: %s", name, value)
					}
				}
			}

			if len(ctx.Body) > 0 {
				bodyStr := string(ctx.Body)
				if len(bodyStr) > 1000 {
					logger.Info("Response Body (truncated): %s...", bodyStr[:1000])
				} else {
					logger.Info("Response Body: %s", bodyStr)
				}
			} else {
				logger.Info("Response Body: [empty]")
			}

			if ctx.Error != nil {
				logger.Info("Error: %v", ctx.Error)
			}
			logger.Info("==================")

			return next(ctx)
		}
	}
}

// AddDetailedDebugLogging adds verbose request/response debug logging to the client.
// Call this after creating your client if you want to see full HTTP requests/responses.
func AddDetailedDebugLogging(client *Client) {
	client.AddRequestMiddleware(DebugLoggingMiddleware(client.Logger))
	client.AddResponseMiddleware(DebugResponseLoggingMiddleware(client.Logger))
}
