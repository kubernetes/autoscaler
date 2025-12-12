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
	"encoding/json"
	"net/http"
	"strings"
	"testing"
	"time"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/verdacloud/verdacloud-sdk-go/verda/testutil"
)

func TestAuthService_Authenticate(t *testing.T) {
	mockServer := testutil.NewMockServer()
	defer mockServer.Close()

	client := NewTestClient(mockServer)

	t.Run("successful authentication", func(t *testing.T) {
		token, err := client.Auth.Authenticate()
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if token == nil {
			t.Fatal("expected token, got nil")
		}

		if token.AccessToken != "mock_access_token_12345" {
			t.Errorf("expected access token 'mock_access_token_12345', got '%s'", token.AccessToken)
		}

		if token.RefreshToken != "mock_refresh_token_67890" {
			t.Errorf("expected refresh token 'mock_refresh_token_67890', got '%s'", token.RefreshToken)
		}

		if token.TokenType != "Bearer" {
			t.Errorf("expected token type 'Bearer', got '%s'", token.TokenType)
		}

		if token.ExpiresIn != 3600 {
			t.Errorf("expected expires_in 3600, got %d", token.ExpiresIn)
		}

		if token.ExpiresAt.IsZero() {
			t.Error("expected ExpiresAt to be set")
		}
	})

	t.Run("authentication error", func(t *testing.T) {
		// Set up mock server to return error for authentication
		mockServer.SetHandler(http.MethodPost, "/oauth2/token", func(w http.ResponseWriter, r *http.Request) {
			testutil.ErrorResponse(w, http.StatusUnauthorized, "Invalid credentials")
		})

		// Use valid client but server will return error
		client := NewTestClient(mockServer)
		_, err := client.Auth.Authenticate()
		if err == nil {
			t.Error("expected error, got nil")
		}
	})
}

func TestAuthService_RefreshToken(t *testing.T) {
	mockServer := testutil.NewMockServer()
	defer mockServer.Close()

	client := NewTestClient(mockServer)

	t.Run("successful token refresh", func(t *testing.T) {
		// First authenticate to get initial token
		_, err := client.Auth.Authenticate()
		if err != nil {
			t.Fatalf("authentication failed: %v", err)
		}

		// Now test refresh
		token, err := client.Auth.RefreshToken()
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if token == nil {
			t.Fatal("expected token, got nil")
		}

		if token.AccessToken == "" {
			t.Error("expected access token to be set")
		}
	})

	t.Run("refresh without initial token", func(t *testing.T) {
		// Create new client without prior authentication
		newClient := NewTestClient(mockServer)

		token, err := newClient.Auth.RefreshToken()
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		// Should fall back to authentication
		if token == nil {
			t.Fatal("expected token, got nil")
		}
	})

	t.Run("refresh token failure fallback", func(t *testing.T) {
		// Create client and authenticate first
		fallbackClient := NewTestClient(mockServer)
		_, err := fallbackClient.Auth.Authenticate()
		if err != nil {
			t.Fatalf("authentication failed: %v", err)
		}

		// Set up mock server to fail refresh but succeed authentication
		mockServer.SetHandler(http.MethodPost, "/oauth2/token", func(w http.ResponseWriter, r *http.Request) {
			if err := r.ParseForm(); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			if r.FormValue("grant_type") == "refresh_token" {
				// Fail refresh
				testutil.ErrorResponse(w, http.StatusBadRequest, "Invalid refresh token")
				return
			}
			// Succeed authentication
			response := map[string]interface{}{
				"access_token":  "new_access_token",
				"refresh_token": "new_refresh_token",
				"token_type":    "Bearer",
				"expires_in":    3600,
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(response)
		})

		token, err := fallbackClient.Auth.RefreshToken()
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if token == nil {
			t.Fatal("expected token, got nil")
		}
	})
}

func TestAuthService_GetValidToken(t *testing.T) {
	mockServer := testutil.NewMockServer()
	defer mockServer.Close()

	client := NewTestClient(mockServer)

	t.Run("get valid token - initial authentication", func(t *testing.T) {
		token, err := client.Auth.GetValidToken()
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if token == nil {
			t.Fatal("expected token, got nil")
		}

		if token.AccessToken == "" {
			t.Error("expected access token to be set")
		}
	})

	t.Run("get valid token - use existing valid token", func(t *testing.T) {
		// First call to authenticate
		firstToken, err := client.Auth.GetValidToken()
		if err != nil {
			t.Fatalf("first authentication failed: %v", err)
		}

		// Second call should return the same token without re-authentication
		secondToken, err := client.Auth.GetValidToken()
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if secondToken.AccessToken != firstToken.AccessToken {
			t.Error("expected same token to be returned")
		}
	})

	t.Run("get valid token - refresh expired token", func(t *testing.T) {
		// First authenticate
		_, err := client.Auth.Authenticate()
		if err != nil {
			t.Fatalf("authentication failed: %v", err)
		}

		// Manually set token to be expired
		client.Auth.mu.Lock()
		client.Auth.token.ExpiresAt = time.Now().Add(-1 * time.Hour)
		client.Auth.mu.Unlock()

		// This should trigger a refresh
		token, err := client.Auth.GetValidToken()
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if token == nil {
			t.Fatal("expected token, got nil")
		}
	})
}

func TestAuthService_IsExpired(t *testing.T) {
	mockServer := testutil.NewMockServer()
	defer mockServer.Close()

	client := NewTestClient(mockServer)

	t.Run("no token - should be expired", func(t *testing.T) {
		if !client.Auth.IsExpired() {
			t.Error("expected token to be expired when no token exists")
		}
	})

	t.Run("valid token - should not be expired", func(t *testing.T) {
		_, err := client.Auth.Authenticate()
		if err != nil {
			t.Fatalf("authentication failed: %v", err)
		}

		if client.Auth.IsExpired() {
			t.Error("expected token to not be expired")
		}
	})

	t.Run("expired token - should be expired", func(t *testing.T) {
		_, err := client.Auth.Authenticate()
		if err != nil {
			t.Fatalf("authentication failed: %v", err)
		}

		// Manually set token to be expired
		client.Auth.mu.Lock()
		client.Auth.token.ExpiresAt = time.Now().Add(-1 * time.Hour)
		client.Auth.mu.Unlock()

		if !client.Auth.IsExpired() {
			t.Error("expected token to be expired")
		}
	})
}

func TestAuthService_GetBearerToken(t *testing.T) {
	mockServer := testutil.NewMockServer()
	defer mockServer.Close()

	client := NewTestClient(mockServer)

	t.Run("get bearer token", func(t *testing.T) {
		bearerToken, err := client.Auth.GetBearerToken()
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		expectedPrefix := "Bearer "
		if !strings.HasPrefix(bearerToken, expectedPrefix) {
			t.Errorf("expected bearer token to start with '%s', got '%s'", expectedPrefix, bearerToken)
		}

		// Should contain the access token
		if !strings.Contains(bearerToken, "mock_access_token_12345") {
			t.Errorf("expected bearer token to contain access token")
		}
	})
}
