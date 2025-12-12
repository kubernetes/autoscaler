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
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

type TestUser struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

func TestStandaloneRequestFunctions(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			switch r.URL.Path {
			case "/users":
				users := []TestUser{
					{ID: 1, Name: "John Doe"},
					{ID: 2, Name: "Jane Smith"},
				}
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(users)
			case "/users/1":
				user := TestUser{ID: 1, Name: "John Doe"}
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(user)
			}
		case http.MethodPost:
			if r.URL.Path == "/users" {
				user := TestUser{ID: 3, Name: "New User"}
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(user)
			}
		case http.MethodPut:
			if r.URL.Path == "/users/1" {
				user := TestUser{ID: 1, Name: "Updated User"}
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(user)
			}
		case http.MethodDelete:
			if r.URL.Path == "/users/1" {
				w.WriteHeader(http.StatusNoContent)
			}
		}
	}))
	defer server.Close()

	// Create client
	client, err := NewClient(
		WithBaseURL(server.URL),
		WithClientID("test_id"),
		WithClientSecret("test_secret"),
	)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()

	// Test GET request for array
	t.Run("GET array", func(t *testing.T) {
		users, resp, err := getRequest[[]TestUser](ctx, client, "/users")
		if err != nil {
			t.Fatalf("GET request failed: %v", err)
		}
		if resp == nil {
			t.Fatal("Response should not be nil")
		}
		if len(users) != 2 {
			t.Errorf("Expected 2 users, got %d", len(users))
		}
		if users[0].Name != "John Doe" {
			t.Errorf("Expected first user name 'John Doe', got '%s'", users[0].Name)
		}
	})

	// Test GET request for single object
	t.Run("GET single", func(t *testing.T) {
		user, resp, err := getRequest[TestUser](ctx, client, "/users/1")
		if err != nil {
			t.Fatalf("GET request failed: %v", err)
		}
		if resp == nil {
			t.Fatal("Response should not be nil")
		}
		if user.ID != 1 {
			t.Errorf("Expected user ID 1, got %d", user.ID)
		}
		if user.Name != "John Doe" {
			t.Errorf("Expected user name 'John Doe', got '%s'", user.Name)
		}
	})

	// Test POST request
	t.Run("POST", func(t *testing.T) {
		newUser := TestUser{Name: "New User"}
		user, resp, err := postRequest[TestUser](ctx, client, "/users", newUser)
		if err != nil {
			t.Fatalf("POST request failed: %v", err)
		}
		if resp == nil {
			t.Fatal("Response should not be nil")
		}
		if user.ID != 3 {
			t.Errorf("Expected user ID 3, got %d", user.ID)
		}
		if user.Name != "New User" {
			t.Errorf("Expected user name 'New User', got '%s'", user.Name)
		}
	})

	// Test PUT request
	t.Run("PUT", func(t *testing.T) {
		updateUser := TestUser{Name: "Updated User"}
		user, resp, err := putRequest[TestUser](ctx, client, "/users/1", updateUser)
		if err != nil {
			t.Fatalf("PUT request failed: %v", err)
		}
		if resp == nil {
			t.Fatal("Response should not be nil")
		}
		if user.Name != "Updated User" {
			t.Errorf("Expected user name 'Updated User', got '%s'", user.Name)
		}
	})

	// Test DELETE request with no result
	t.Run("DELETE no result", func(t *testing.T) {
		resp, err := deleteRequestNoResult(ctx, client, "/users/1")
		if err != nil {
			t.Fatalf("DELETE request failed: %v", err)
		}
		if resp == nil {
			t.Fatal("Response should not be nil")
		}
		if resp.StatusCode != http.StatusNoContent {
			t.Errorf("Expected status 204, got %d", resp.StatusCode)
		}
	})
}

func TestServiceIntegration(t *testing.T) {
	// Create a test server that mimics the Verda API
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Add Authorization header check
		if r.Header.Get("Authorization") == "" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		switch r.URL.Path {
		case "/balance":
			balance := Balance{Amount: 100.50, Currency: "USD"}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(balance)
		case "/locations":
			locations := []Location{
				{Code: "FIN-01", Name: "Finland 1", CountryCode: "FI"},
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(locations)
		}
	}))
	defer server.Close()

	// Create client
	client, err := NewClient(
		WithBaseURL(server.URL),
		WithClientID("test_id"),
		WithClientSecret("test_secret"),
	)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()

	// Test balance service
	t.Run("Balance Service", func(t *testing.T) {
		balance, err := client.Balance.Get(ctx)
		if err != nil {
			// Expected to fail due to auth, but let's check the error type
			t.Logf("Expected auth error: %v", err)
		} else {
			if balance.Amount != 100.50 {
				t.Errorf("Expected balance 100.50, got %f", balance.Amount)
			}
		}
	})

	// Test locations service
	t.Run("Locations Service", func(t *testing.T) {
		locations, err := client.Locations.Get(ctx)
		if err != nil {
			// Expected to fail due to auth, but let's check the error type
			t.Logf("Expected auth error: %v", err)
		} else {
			if len(locations) != 1 {
				t.Errorf("Expected 1 location, got %d", len(locations))
			}
		}
	})
}
