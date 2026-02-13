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
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

// DefaultBaseURL is the default base URL for the Verda API.
const (
	DefaultBaseURL = "https://api.verda.com/v1"
	trueString     = "true"
)

// Client is the Verda API client.
type Client struct {
	BaseURL         string
	ClientID        string
	ClientSecret    string
	AuthBearerToken string
	UserAgent       string

	HTTPClient *http.Client
	Logger     Logger

	// Middleware management for all requests
	Middleware *Middleware

	// Services
	Auth                 *AuthService
	Instances            *InstanceService
	StartupScripts       *StartupScriptService
	InstanceTypes        *InstanceTypesService
	InstanceAvailability *InstanceAvailabilityService
}

// ClientOption is a function that configures a Client.
type ClientOption func(*Client)

// NewClient creates a new Verda API client with the given options.
func NewClient(options ...ClientOption) (*Client, error) {

	client := &Client{
		BaseURL:    DefaultBaseURL,
		HTTPClient: &http.Client{},
		Logger:     &NoOpLogger{}, // Default: no logging
	}

	for _, option := range options {
		option(client)
	}

	// Enable debug logging via VERDA_DEBUG env var if no custom logger was set
	if verdaDebug := os.Getenv("VERDA_DEBUG"); strings.ToLower(verdaDebug) == trueString {
		if _, isNoOp := client.Logger.(*NoOpLogger); isNoOp {
			client.Logger = NewStdLogger(true)
		}
	}

	client.Middleware = NewDefaultMiddlewareWithUserAgent(client.Logger, client.UserAgent)

	// Wire up debug middleware if VERDA_DEBUG is set
	if verdaDebug := os.Getenv("VERDA_DEBUG"); strings.ToLower(verdaDebug) == trueString {
		client.Middleware.AddRequestMiddleware(DebugLoggingMiddleware(client.Logger))
		client.Middleware.AddResponseMiddleware(DebugResponseLoggingMiddleware(client.Logger))
	}

	if client.ClientID == "" {
		return nil, fmt.Errorf("client ID is required")
	}
	if client.ClientSecret == "" {
		return nil, fmt.Errorf("client secret is required")
	}

	client.Auth = &AuthService{client: client}
	client.Instances = &InstanceService{client: client}
	client.StartupScripts = &StartupScriptService{client: client}
	client.InstanceTypes = &InstanceTypesService{client: client}
	client.InstanceAvailability = &InstanceAvailabilityService{client: client}

	return client, nil
}

// AddRequestMiddleware adds a request middleware to the client.
func (c *Client) AddRequestMiddleware(middleware RequestMiddleware) {
	c.Middleware.AddRequestMiddleware(middleware)
}

// AddResponseMiddleware adds a response middleware to the client.
func (c *Client) AddResponseMiddleware(middleware ResponseMiddleware) {
	c.Middleware.AddResponseMiddleware(middleware)
}

// SetRequestMiddleware sets the request middleware for the client.
func (c *Client) SetRequestMiddleware(middleware []RequestMiddleware) {
	c.Middleware.SetRequestMiddleware(middleware)
}

// SetResponseMiddleware sets the response middleware for the client.
func (c *Client) SetResponseMiddleware(middleware []ResponseMiddleware) {
	c.Middleware.SetResponseMiddleware(middleware)
}

// ClearRequestMiddleware clears all request middleware from the client.
func (c *Client) ClearRequestMiddleware() {
	c.Middleware.ClearRequestMiddleware()
}

// ClearResponseMiddleware clears all response middleware from the client.
func (c *Client) ClearResponseMiddleware() {
	c.Middleware.ClearResponseMiddleware()
}

// WithBaseURL sets the base URL for the client.
func WithBaseURL(baseURL string) ClientOption {
	return func(c *Client) {
		c.BaseURL = baseURL
	}
}

// WithClientID sets the client ID for the client.
func WithClientID(clientID string) ClientOption {
	return func(c *Client) {
		c.ClientID = clientID
	}
}

// WithClientSecret sets the client secret for the client.
func WithClientSecret(clientSecret string) ClientOption {
	return func(c *Client) {
		c.ClientSecret = clientSecret

	}
}

// WithAuthBearerToken sets the authentication bearer token for the client.
func WithAuthBearerToken(token string) ClientOption {
	return func(c *Client) {
		c.AuthBearerToken = token
	}
}

// WithLogger sets the logger for the client.
func WithLogger(logger Logger) ClientOption {
	return func(c *Client) {
		c.Logger = logger
	}
}

// WithDebugLogging enables or disables debug logging for the client.
func WithDebugLogging(enabled bool) ClientOption {
	return func(c *Client) {
		c.Logger = NewStdLogger(enabled)
	}
}

// WithHTTPClient sets the HTTP client for the client.
func WithHTTPClient(httpClient *http.Client) ClientOption {
	return func(c *Client) {
		c.HTTPClient = httpClient
	}
}

// WithUserAgent sets the custom User-Agent string for the client.
func WithUserAgent(userAgent string) ClientOption {
	return func(c *Client) {
		c.UserAgent = userAgent
	}
}

// WithHTTPClient returns a ClientOption that sets the HTTP client.
func (c *Client) WithHTTPClient(httpClient *http.Client) ClientOption {
	return func(c *Client) {
		c.HTTPClient = httpClient
	}
}

// NewRequest creates a new HTTP request with the given method, path, and body.
func (c *Client) NewRequest(ctx context.Context, method, path string, body io.Reader) (*http.Request, error) {
	url := c.BaseURL + path

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	req = req.WithContext(ctx)
	return req, nil
}

// Do executes the request through middleware and returns the parsed response
func (c *Client) Do(req *http.Request, result any) (*Response, error) {
	// Snapshot middleware to avoid race conditions
	requestMiddleware, responseMiddleware := c.Middleware.Snapshot()

	reqCtx := &RequestContext{
		Method:  req.Method,
		Path:    req.URL.Path,
		Body:    nil,
		Headers: req.Header.Clone(),
		Query:   req.URL.Query(),
		Client:  c,
		Request: req,
	}

	requestHandler := c.buildRequestChain(requestMiddleware)
	if err := requestHandler(reqCtx); err != nil {
		return nil, fmt.Errorf("request middleware failed: %w", err)
	}

	// Apply middleware-modified headers back to the request
	for name, values := range reqCtx.Headers {
		req.Header.Del(name)
		for _, value := range values {
			req.Header.Add(name, value)
		}
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	wrappedResp := &Response{Response: resp}

	bodyBytes, readErr := io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	if readErr != nil {
		return wrappedResp, fmt.Errorf("failed to read response body: %w", readErr)
	}

	var parseErr error
	if result != nil {
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			var apiError APIError
			if err := json.Unmarshal(bodyBytes, &apiError); err != nil {
				parseErr = &APIError{
					StatusCode: resp.StatusCode,
					Message:    string(bodyBytes),
				}
			} else {
				apiError.StatusCode = resp.StatusCode
				parseErr = &apiError
			}
		} else {
			trimmed := bytes.TrimSpace(bodyBytes)
			if len(trimmed) > 0 {
				if err := json.Unmarshal(trimmed, result); err != nil {
					parseErr = fmt.Errorf("failed to unmarshal response: %w", err)
				}
			}
		}
	}

	respCtx := &ResponseContext{
		Request:    reqCtx,
		Response:   resp,
		Body:       bodyBytes,
		StatusCode: resp.StatusCode,
		Error:      parseErr,
	}

	responseHandler := c.buildResponseChain(responseMiddleware)
	if middlewareErr := responseHandler(respCtx); middlewareErr != nil {
		return wrappedResp, fmt.Errorf("response middleware failed: %w", middlewareErr)
	}

	if respCtx.Error != nil {
		return wrappedResp, respCtx.Error
	}

	return wrappedResp, nil
}

func (c *Client) buildRequestChain(requestMiddleware []RequestMiddleware) RequestHandler {
	//nolint:revive // ctx unused in default no-op handler
	handler := func(ctx *RequestContext) error {
		return nil
	}

	// Reverse order so last middleware wraps first
	for i := len(requestMiddleware) - 1; i >= 0; i-- {
		handler = requestMiddleware[i](handler)
	}

	return handler
}

func (c *Client) buildResponseChain(responseMiddleware []ResponseMiddleware) ResponseHandler {
	//nolint:revive // ctx unused in default no-op handler
	handler := func(ctx *ResponseContext) error {
		return nil
	}

	// Reverse order so last middleware wraps first
	for i := len(responseMiddleware) - 1; i >= 0; i-- {
		handler = responseMiddleware[i](handler)
	}

	return handler
}

// makeRequest is kept for endpoints with plain text responses that don't fit the normal JSON flow.
// Prefer NewRequest() + Do() for new code to get full middleware support.
func (c *Client) makeRequest(ctx context.Context, method, path string, body any) (*http.Response, error) {
	url := c.BaseURL + path

	var reqBody io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonBody)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	bearerToken, err := c.Auth.GetBearerToken()
	if err != nil {
		return nil, fmt.Errorf("failed to get authentication token: %w", err)
	}
	req.Header.Set("Authorization", bearerToken)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	return resp, nil
}

func (c *Client) handleResponse(resp *http.Response, result any) error {
	defer func() {
		_ = resp.Body.Close()
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var apiError APIError
		if err := json.Unmarshal(body, &apiError); err != nil {
			return &APIError{
				StatusCode: resp.StatusCode,
				Message:    string(body),
			}
		}
		apiError.StatusCode = resp.StatusCode
		return &apiError
	}

	trimmed := bytes.TrimSpace(body)
	if len(trimmed) == 0 {
		return nil
	}

	if result != nil {
		if err := json.Unmarshal(trimmed, result); err != nil {
			return fmt.Errorf("failed to unmarshal response: %w", err)
		}
	}

	return nil
}
