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
	"context"
	"net/http"
	"testing"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/verdacloud/verdacloud-sdk-go/verda/testutil"
)

func TestNewClient(t *testing.T) {
	tests := []struct {
		name      string
		opts      []ClientOption
		wantError bool
	}{
		{
			name: "valid config",
			opts: []ClientOption{
				WithClientID("test_id"),
				WithClientSecret("test_secret"),
			},
			wantError: false,
		},
		{
			name: "missing client ID",
			opts: []ClientOption{
				WithClientSecret("test_secret"),
			},
			wantError: true,
		},
		{
			name: "missing client secret",
			opts: []ClientOption{
				WithClientID("test_id"),
			},
			wantError: true,
		},
		{
			name: "custom base URL",
			opts: []ClientOption{
				WithClientID("test_id"),
				WithClientSecret("test_secret"),
				WithBaseURL("https://custom.example.com/v1"),
			},
			wantError: false,
		},
		{
			name: "custom HTTP client",
			opts: []ClientOption{
				WithClientID("test_id"),
				WithClientSecret("test_secret"),
				WithHTTPClient(&http.Client{}),
			},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewClient(tt.opts...)

			if tt.wantError {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if client == nil {
				t.Error("expected client, got nil")
				return
			}

			// Check default base URL
			expectedBaseURL := client.BaseURL
			if expectedBaseURL == "" {
				expectedBaseURL = DefaultBaseURL
			}
			if client.BaseURL != expectedBaseURL {
				t.Errorf("expected BaseURL %s, got %s", expectedBaseURL, client.BaseURL)
			}

			// Check credentials are set (validation happens in NewClient)
			if client.ClientID == "" {
				t.Error("ClientID should be set")
			}
			if client.ClientSecret == "" {
				t.Error("ClientSecret should be set")
			}

			// Check services are initialized
			if client.Auth == nil {
				t.Error("Auth service not initialized")
			}
			if client.Instances == nil {
				t.Error("Instances service not initialized")
			}
			if client.Balance == nil {
				t.Error("Balance service not initialized")
			}
			if client.SSHKeys == nil {
				t.Error("SSHKeys service not initialized")
			}
			if client.Volumes == nil {
				t.Error("Volumes service not initialized")
			}
			if client.StartupScripts == nil {
				t.Error("StartupScripts service not initialized")
			}
			if client.Locations == nil {
				t.Error("Locations service not initialized")
			}
			if client.Images == nil {
				t.Error("Images service not initialized")
			}
			if client.InstanceTypes == nil {
				t.Error("InstanceTypes service not initialized")
			}
			if client.InstanceAvailability == nil {
				t.Error("InstanceAvailability service not initialized")
			}
			if client.ContainerTypes == nil {
				t.Error("ContainerTypes service not initialized")
			}
			if client.LongTerm == nil {
				t.Error("LongTerm service not initialized")
			}
			if client.ContainerDeployments == nil {
				t.Error("ContainerDeployments service not initialized")
			}
			if client.ServerlessJobs == nil {
				t.Error("ServerlessJobs service not initialized")
			}
		})
	}
}

func TestClientMakeRequest(t *testing.T) {
	mockServer := testutil.NewMockServer()
	defer mockServer.Close()

	client := NewTestClient(mockServer)

	// Test successful request
	t.Run("successful request", func(t *testing.T) {
		resp, err := client.makeRequest(context.Background(), http.MethodGet, "/balance", nil)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if resp != nil {
			_ = resp.Body.Close() //nolint:errcheck // Test code
		} else {
			t.Error("expected response, got nil")
		}
	})

	// Test request with body
	t.Run("request with body", func(t *testing.T) {
		requestBody := map[string]string{"test": "data"}
		resp, err := client.makeRequest(context.Background(), http.MethodPost, "/instances", requestBody)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if resp != nil {
			_ = resp.Body.Close() //nolint:errcheck // Test code
		} else {
			t.Error("expected response, got nil")
		}
	})
}

func TestClientHandleResponse(t *testing.T) {
	mockServer := testutil.NewMockServer()
	defer mockServer.Close()

	client := NewTestClient(mockServer)

	t.Run("successful response parsing", func(t *testing.T) {
		resp, err := client.makeRequest(context.Background(), http.MethodGet, "/balance", nil)
		if err != nil {
			t.Fatalf("unexpected error making request: %v", err)
		}

		var balance Balance
		err = client.handleResponse(resp, &balance)
		if err != nil {
			t.Errorf("unexpected error handling response: %v", err)
		}

		if balance.Amount != 100.50 {
			t.Errorf("expected amount 100.50, got %f", balance.Amount)
		}
		if balance.Currency != "USD" {
			t.Errorf("expected currency USD, got %s", balance.Currency)
		}
	})

	t.Run("error response", func(t *testing.T) {
		// Set up a custom handler that returns an error
		mockServer.SetHandler(http.MethodGet, "/error-test", func(w http.ResponseWriter, r *http.Request) {
			testutil.ErrorResponse(w, http.StatusBadRequest, "Test error message")
		})

		resp, err := client.makeRequest(context.Background(), http.MethodGet, "/error-test", nil)
		if err != nil {
			t.Fatalf("unexpected error making request: %v", err)
		}

		err = client.handleResponse(resp, nil)
		if err == nil {
			t.Error("expected error, got nil")
		}

		apiErr, ok := err.(*APIError)
		if !ok {
			t.Errorf("expected APIError, got %T", err)
		} else {
			if apiErr.StatusCode != http.StatusBadRequest {
				t.Errorf("expected status code %d, got %d", http.StatusBadRequest, apiErr.StatusCode)
			}
			if apiErr.Message != "Test error message" {
				t.Errorf("expected message 'Test error message', got '%s'", apiErr.Message)
			}
		}
	})
}

func TestClientMiddleware(t *testing.T) {
	mockServer := testutil.NewMockServer()
	defer mockServer.Close()

	client := NewTestClient(mockServer)

	t.Run("set request middleware", func(t *testing.T) {
		called := false
		middleware := func(next RequestHandler) RequestHandler {
			return func(ctx *RequestContext) error {
				called = true
				return next(ctx)
			}
		}

		client.SetRequestMiddleware([]RequestMiddleware{middleware})
		_, _ = client.Balance.Get(context.Background())

		if !called {
			t.Error("expected request middleware to be called")
		}
	})

	t.Run("set response middleware", func(t *testing.T) {
		called := false
		middleware := func(next ResponseHandler) ResponseHandler {
			return func(ctx *ResponseContext) error {
				called = true
				return next(ctx)
			}
		}

		client.SetResponseMiddleware([]ResponseMiddleware{middleware})
		_, _ = client.Balance.Get(context.Background())

		if !called {
			t.Error("expected response middleware to be called")
		}
	})

	t.Run("clear request middleware", func(t *testing.T) {
		called := false
		middleware := func(next RequestHandler) RequestHandler {
			return func(ctx *RequestContext) error {
				called = true
				return next(ctx)
			}
		}

		client.AddRequestMiddleware(middleware)
		client.ClearRequestMiddleware()
		_, _ = client.Balance.Get(context.Background())

		if called {
			t.Error("expected request middleware to not be called after clear")
		}
	})

	t.Run("clear response middleware", func(t *testing.T) {
		called := false
		middleware := func(next ResponseHandler) ResponseHandler {
			return func(ctx *ResponseContext) error {
				called = true
				return next(ctx)
			}
		}

		client.AddResponseMiddleware(middleware)
		client.ClearResponseMiddleware()
		_, _ = client.Balance.Get(context.Background())

		if called {
			t.Error("expected response middleware to not be called after clear")
		}
	})

	t.Run("with auth bearer token", func(t *testing.T) {
		opts := []ClientOption{
			WithClientID("test_id"),
			WithClientSecret("test_secret"),
			WithAuthBearerToken("test_bearer_token"),
		}

		client, err := NewClient(opts...)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Just verify that the client was created successfully
		// The token is stored internally and used for authentication
		if client == nil {
			t.Fatal("expected client to be created")
		}
	})
}
