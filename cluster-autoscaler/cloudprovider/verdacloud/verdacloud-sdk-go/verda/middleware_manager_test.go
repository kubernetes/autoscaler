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
	"testing"
)

func TestMiddlewareManager(t *testing.T) {
	t.Run("NewDefaultMiddleware", func(t *testing.T) {
		// Test with NoOpLogger (should not add logging middleware)
		middleware := NewDefaultMiddleware(&NoOpLogger{})

		reqCount, respCount := middleware.Len()
		if reqCount != 2 {
			t.Errorf("Expected 2 default request middleware, got %d", reqCount)
		}
		if respCount != 1 {
			t.Errorf("Expected 1 default response middleware, got %d", respCount)
		}

		// Test with debug logger (should add logging middleware)
		debugMiddleware := NewDefaultMiddleware(NewStdLogger(true))
		reqCount, respCount = debugMiddleware.Len()
		if reqCount != 3 { // Auth + JSON + Logging
			t.Errorf("Expected 3 request middleware with debug logger, got %d", reqCount)
		}
		if respCount != 2 { // Error + ResponseLogging
			t.Errorf("Expected 2 response middleware with debug logger, got %d", respCount)
		}
	})

	t.Run("AddRequestMiddleware", func(t *testing.T) {
		middleware := NewMiddleware(nil, nil)

		testMiddleware := func(next RequestHandler) RequestHandler {
			return func(ctx *RequestContext) error {
				return next(ctx)
			}
		}

		middleware.AddRequestMiddleware(testMiddleware)

		if middleware.LenRequestMiddleware() != 1 {
			t.Errorf("Expected 1 request middleware, got %d", middleware.LenRequestMiddleware())
		}
	})

	t.Run("AddResponseMiddleware", func(t *testing.T) {
		middleware := NewMiddleware(nil, nil)

		testMiddleware := func(next ResponseHandler) ResponseHandler {
			return func(ctx *ResponseContext) error {
				return next(ctx)
			}
		}

		middleware.AddResponseMiddleware(testMiddleware)

		if middleware.LenResponseMiddleware() != 1 {
			t.Errorf("Expected 1 response middleware, got %d", middleware.LenResponseMiddleware())
		}
	})

	t.Run("SetRequestMiddleware", func(t *testing.T) {
		middleware := NewDefaultMiddleware(&NoOpLogger{})

		newMiddleware := []RequestMiddleware{
			func(next RequestHandler) RequestHandler {
				return func(ctx *RequestContext) error {
					return next(ctx)
				}
			},
		}

		middleware.SetRequestMiddleware(newMiddleware)

		if middleware.LenRequestMiddleware() != 1 {
			t.Errorf("Expected 1 request middleware after set, got %d", middleware.LenRequestMiddleware())
		}
	})

	t.Run("ClearMiddleware", func(t *testing.T) {
		middleware := NewDefaultMiddleware(&NoOpLogger{})

		middleware.Clear()

		reqCount, respCount := middleware.Len()
		if reqCount != 0 {
			t.Errorf("Expected 0 request middleware after clear, got %d", reqCount)
		}
		if respCount != 0 {
			t.Errorf("Expected 0 response middleware after clear, got %d", respCount)
		}
	})

	t.Run("Snapshot", func(t *testing.T) {
		middleware := NewDefaultMiddleware(&NoOpLogger{})

		// Get snapshot
		reqSnapshot, respSnapshot := middleware.Snapshot()

		// Verify snapshot has correct length
		if len(reqSnapshot) != 2 {
			t.Errorf("Expected 2 request middleware in snapshot, got %d", len(reqSnapshot))
		}
		if len(respSnapshot) != 1 {
			t.Errorf("Expected 1 response middleware in snapshot, got %d", len(respSnapshot))
		}

		// Modify original - snapshot should be unaffected
		middleware.ClearRequestMiddleware()

		// Snapshot should still have original values
		if len(reqSnapshot) != 2 {
			t.Errorf("Snapshot was affected by original modification")
		}

		// Original should be cleared
		if middleware.LenRequestMiddleware() != 0 {
			t.Errorf("Expected 0 request middleware after clear, got %d", middleware.LenRequestMiddleware())
		}
	})
}

func TestMiddlewareThreadSafety(t *testing.T) {
	middleware := NewDefaultMiddleware(&NoOpLogger{})

	// Test concurrent access
	var wg sync.WaitGroup
	numGoroutines := 10

	// Add middleware concurrently
	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()

			testMiddleware := func(next RequestHandler) RequestHandler {
				return func(ctx *RequestContext) error {
					return next(ctx)
				}
			}

			middleware.AddRequestMiddleware(testMiddleware)
		}()
	}

	// Take snapshots concurrently
	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			_, _ = middleware.Snapshot()
		}()
	}

	wg.Wait()

	// Should have original 2 + 10 added = 12 request middleware
	if middleware.LenRequestMiddleware() != 12 {
		t.Errorf("Expected 12 request middleware after concurrent adds, got %d", middleware.LenRequestMiddleware())
	}
}

func TestClientMiddlewareIntegration(t *testing.T) {
	client, err := NewClient(
		WithClientID("test"),
		WithClientSecret("test"),
	)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Test that client has default middleware
	reqCount, respCount := client.Middleware.Len()
	if reqCount != 2 {
		t.Errorf("Expected 2 default request middleware in client, got %d", reqCount)
	}
	if respCount != 1 {
		t.Errorf("Expected 1 default response middleware in client, got %d", respCount)
	}

	// Test adding middleware through client methods
	testReqMiddleware := func(next RequestHandler) RequestHandler {
		return func(ctx *RequestContext) error {
			return next(ctx)
		}
	}

	testRespMiddleware := func(next ResponseHandler) ResponseHandler {
		return func(ctx *ResponseContext) error {
			return next(ctx)
		}
	}

	client.AddRequestMiddleware(testReqMiddleware)
	client.AddResponseMiddleware(testRespMiddleware)

	// Verify middleware were added
	reqCount, respCount = client.Middleware.Len()
	if reqCount != 3 {
		t.Errorf("Expected 3 request middleware after adding one, got %d", reqCount)
	}
	if respCount != 2 {
		t.Errorf("Expected 2 response middleware after adding one, got %d", respCount)
	}

	// Test that snapshots get copies
	req1Middleware, _ := client.Middleware.Snapshot()
	req2Middleware, _ := client.Middleware.Snapshot()

	// Should be equal length but different slices
	if len(req1Middleware) != len(req2Middleware) {
		t.Error("Snapshots should have equal length")
	}

	// Verify they are different slice instances (not same underlying array)
	if &req1Middleware[0] == &req2Middleware[0] {
		t.Error("Snapshots should be independent copies")
	}
}
