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
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

// AuthService handles authentication with the Verda API.
type AuthService struct {
	client *Client
	mu     sync.RWMutex
	token  *TokenResponse
}

// TokenRequest represents an OAuth token request.
type TokenRequest struct {
	GrantType    string `json:"grant_type"`
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	RefreshToken string `json:"refresh_token,omitempty"`
}

// TokenResponse represents an OAuth token response.
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	Scope        string `json:"scope"`
	ExpiresAt    time.Time
}

// Authenticate authenticates with the Verda API using client credentials.
func (s *AuthService) Authenticate() (*TokenResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.doTokenRequest(TokenRequest{
		GrantType:    "client_credentials",
		ClientID:     s.client.ClientID,
		ClientSecret: s.client.ClientSecret,
	})
}

// RefreshToken refreshes the authentication token.
func (s *AuthService) RefreshToken() (*TokenResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.token == nil || s.token.RefreshToken == "" {
		return s.authenticateWithoutLock()
	}

	return s.doTokenRequest(TokenRequest{
		GrantType:    "refresh_token",
		RefreshToken: s.token.RefreshToken,
		ClientID:     s.client.ClientID,
		ClientSecret: s.client.ClientSecret,
	})
}

func (s *AuthService) authenticateWithoutLock() (*TokenResponse, error) {
	return s.doTokenRequest(TokenRequest{
		GrantType:    "client_credentials",
		ClientID:     s.client.ClientID,
		ClientSecret: s.client.ClientSecret,
	})
}

// doTokenRequest tries JSON first (production), falls back to form-encoded (staging quirk)
func (s *AuthService) doTokenRequest(body TokenRequest) (*TokenResponse, error) {
	payload, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal token request: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, s.client.BaseURL+"/oauth2/token", bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("failed to create token request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	if s.client.AuthBearerToken != "" {
		req.Header.Set("Authorization", "Bearer "+s.client.AuthBearerToken)
	}

	resp, err := s.client.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("authentication request failed: %w", err)
	}

	var tokenResp TokenResponse
	err = s.client.handleResponse(resp, &tokenResp)
	if err == nil {
		tokenResp.ExpiresAt = time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)
		s.token = &tokenResp
		return &tokenResp, nil
	}

	// Staging environments may require form-encoded instead of JSON
	if apiErr, ok := err.(*APIError); ok {
		msg := strings.ToLower(apiErr.Message)
		if apiErr.StatusCode == 400 && (strings.Contains(msg, "grant_type") && strings.Contains(msg, "not specified") || strings.Contains(msg, "unsupported grant type") || strings.Contains(msg, "not valid json")) {
			form := url.Values{}
			form.Set("grant_type", body.GrantType)
			if body.ClientID != "" {
				form.Set("client_id", body.ClientID)
			}
			if body.ClientSecret != "" {
				form.Set("client_secret", body.ClientSecret)
			}
			if body.RefreshToken != "" {
				form.Set("refresh_token", body.RefreshToken)
			}

			req2, err2 := http.NewRequest(http.MethodPost, s.client.BaseURL+"/oauth2/token", strings.NewReader(form.Encode()))
			if err2 != nil {
				return nil, fmt.Errorf("failed to create token request (form): %w", err2)
			}
			req2.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			req2.Header.Set("Accept", "application/json")
			if s.client.AuthBearerToken != "" {
				req2.Header.Set("Authorization", "Bearer "+s.client.AuthBearerToken)
			}
			resp2, err2 := s.client.HTTPClient.Do(req2)
			if err2 != nil {
				return nil, fmt.Errorf("authentication request failed (form): %w", err2)
			}
			var tokenResp2 TokenResponse
			if err3 := s.client.handleResponse(resp2, &tokenResp2); err3 == nil {
				tokenResp2.ExpiresAt = time.Now().Add(time.Duration(tokenResp2.ExpiresIn) * time.Second)
				s.token = &tokenResp2
				return &tokenResp2, nil
			}
		}
	}
	return nil, fmt.Errorf("authentication failed: %w", err)
}

// GetValidToken returns a valid token, refreshing if necessary.
func (s *AuthService) GetValidToken() (*TokenResponse, error) {
	s.mu.RLock()
	token := s.token
	s.mu.RUnlock()

	if token == nil {
		return s.Authenticate()
	}

	// Refresh 30s before expiry to avoid races
	if time.Now().Add(30 * time.Second).After(token.ExpiresAt) {
		return s.RefreshToken()
	}

	return token, nil
}

// IsExpired returns true if the current token is expired or about to expire.
func (s *AuthService) IsExpired() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.token == nil {
		return true
	}

	// Consider expired 30s before actual expiry
	return time.Now().Add(30 * time.Second).After(s.token.ExpiresAt)
}

// GetBearerToken returns a valid bearer token string for use in Authorization headers.
func (s *AuthService) GetBearerToken() (string, error) {
	token, err := s.GetValidToken()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("Bearer %s", token.AccessToken), nil
}
