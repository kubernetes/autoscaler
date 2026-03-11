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
	"sync"
)

// Middleware manages request and response middleware chains with thread-safe operations
type Middleware struct {
	mu                 sync.RWMutex
	requestMiddleware  []RequestMiddleware
	responseMiddleware []ResponseMiddleware
}

// NewMiddleware creates a new Middleware manager with optional default middleware
func NewMiddleware(requestMiddleware []RequestMiddleware, responseMiddleware []ResponseMiddleware) *Middleware {
	return &Middleware{
		requestMiddleware:  append([]RequestMiddleware{}, requestMiddleware...),
		responseMiddleware: append([]ResponseMiddleware{}, responseMiddleware...),
	}
}

// NewDefaultMiddleware creates a Middleware manager with the standard default middleware
func NewDefaultMiddleware(logger Logger) *Middleware {
	return NewDefaultMiddlewareWithUserAgent(logger, "")
}

// NewDefaultMiddlewareWithUserAgent creates a Middleware manager with custom User-Agent support
func NewDefaultMiddlewareWithUserAgent(logger Logger, customUserAgent string) *Middleware {
	userAgent := BuildUserAgent(customUserAgent)

	requestMiddleware := []RequestMiddleware{
		AuthenticationMiddleware(),
		JSONContentTypeMiddleware(),
		UserAgentMiddleware(userAgent),
	}

	responseMiddleware := []ResponseMiddleware{
		ErrorHandlingMiddleware(),
	}

	// Add logging middleware if logger supports debug (not NoOpLogger)
	if _, isNoOp := logger.(*NoOpLogger); !isNoOp {
		requestMiddleware = append(requestMiddleware, LoggingMiddleware(logger))
		responseMiddleware = append(responseMiddleware, ResponseLoggingMiddleware(logger))
	}

	return NewMiddleware(requestMiddleware, responseMiddleware)
}

// Snapshot returns thread-safe copies of the current middleware chains
func (m *Middleware) Snapshot() ([]RequestMiddleware, []ResponseMiddleware) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	requestCopy := append([]RequestMiddleware{}, m.requestMiddleware...)
	responseCopy := append([]ResponseMiddleware{}, m.responseMiddleware...)

	return requestCopy, responseCopy
}

// AddRequestMiddleware adds a request middleware to the chain.
func (m *Middleware) AddRequestMiddleware(middleware RequestMiddleware) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.requestMiddleware = append(m.requestMiddleware, middleware)
}

// SetRequestMiddleware replaces the request middleware chain.
func (m *Middleware) SetRequestMiddleware(middleware []RequestMiddleware) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.requestMiddleware = append([]RequestMiddleware{}, middleware...)
}

// ClearRequestMiddleware clears all request middleware.
func (m *Middleware) ClearRequestMiddleware() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.requestMiddleware = []RequestMiddleware{}
}

// LenRequestMiddleware returns the number of request middleware.
func (m *Middleware) LenRequestMiddleware() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.requestMiddleware)
}

// AddResponseMiddleware adds a response middleware to the chain.
func (m *Middleware) AddResponseMiddleware(middleware ResponseMiddleware) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.responseMiddleware = append(m.responseMiddleware, middleware)
}

// SetResponseMiddleware replaces the response middleware chain.
func (m *Middleware) SetResponseMiddleware(middleware []ResponseMiddleware) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.responseMiddleware = append([]ResponseMiddleware{}, middleware...)
}

// ClearResponseMiddleware clears all response middleware.
func (m *Middleware) ClearResponseMiddleware() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.responseMiddleware = []ResponseMiddleware{}
}

// LenResponseMiddleware returns the number of response middleware.
func (m *Middleware) LenResponseMiddleware() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.responseMiddleware)
}

// Clear clears all middleware.
func (m *Middleware) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.requestMiddleware = []RequestMiddleware{}
	m.responseMiddleware = []ResponseMiddleware{}
}

// Len returns the number of request and response middleware.
func (m *Middleware) Len() (requestCount, responseCount int) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.requestMiddleware), len(m.responseMiddleware)
}
