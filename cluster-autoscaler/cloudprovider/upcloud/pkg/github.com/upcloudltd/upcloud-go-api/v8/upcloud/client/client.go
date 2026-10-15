package client

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

const (
	Version    string = "8.35.0"
	APIVersion string = "1.3"
	APIBaseURL string = "https://api.upcloud.com"

	EnvDebugAPIBaseURL            string = "UPCLOUD_DEBUG_API_BASE_URL"
	EnvDebugSkipCertificateVerify string = "UPCLOUD_DEBUG_SKIP_CERTIFICATE_VERIFY"
	EnvToken                      string = "UPCLOUD_TOKEN"
	EnvUsername                   string = "UPCLOUD_USERNAME"
	EnvPassword                   string = "UPCLOUD_PASSWORD"
)

// LogFn is a function that logs a message with context and optional key-value pairs, e.g., slog.DebugContext
type LogFn func(context.Context, string, ...any)

type config struct {
	username   string
	password   string
	token      string
	baseURL    string
	httpClient *http.Client
	logger     LogFn
}

// Client represents an API client
type Client struct {
	UserAgent string
	config    config
}

// NewFromEnv creates a new client from environment variables and
// validates that only one authentication method is provided.
func NewFromEnv(c ...ConfigFn) (*Client, error) {
	token := os.Getenv(EnvToken)
	username := os.Getenv(EnvUsername)
	password := os.Getenv(EnvPassword)

	if token != "" && (username != "" || password != "") {
		return nil, errors.New("only one authentication method (token or basic auth) can be provided")
	}

	if token == "" && (username == "" || password == "") {
		return nil, errors.New("authentication credentials must be provided via environment variables")
	}

	config := config{
		baseURL:    clientBaseURL(os.Getenv(EnvDebugAPIBaseURL)),
		httpClient: NewDefaultHTTPClient(),
	}

	if token != "" {
		config.token = token
	} else {
		config.username = username
		config.password = password
	}

	// If set, replace http client transport with one skipping tls verification
	if os.Getenv(EnvDebugSkipCertificateVerify) == "1" {
		c = append(c, WithInsecureSkipVerify())
	}

	for _, fn := range c {
		fn(&config)
	}

	return &Client{
		UserAgent: userAgent(),
		config:    config,
	}, nil
}

// Get performs a GET request to the specified path and returns the response body.
func (c *Client) Get(ctx context.Context, path string) ([]byte, error) {
	r, err := c.createRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	return c.Do(r)
}

// GetStream performs a GET request to the specified path and returns the response body reader.
func (c *Client) GetStream(ctx context.Context, path string) (io.ReadCloser, error) {
	r, err := c.createRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	return c.DoStream(r)
}

// Post performs a POST request to the specified path and returns the response body.
func (c *Client) Post(ctx context.Context, path string, body []byte) ([]byte, error) {
	r, err := c.createRequest(ctx, http.MethodPost, path, body)
	if err != nil {
		return nil, err
	}
	return c.Do(r)
}

// Put performs a PUT request to the specified path and returns the response body.
func (c *Client) Put(ctx context.Context, path string, body []byte) ([]byte, error) {
	r, err := c.createRequest(ctx, http.MethodPut, path, body)
	if err != nil {
		return nil, err
	}
	return c.Do(r)
}

// Patch performs a PATCH request to the specified path and returns the response body.
func (c *Client) Patch(ctx context.Context, path string, body []byte) ([]byte, error) {
	r, err := c.createRequest(ctx, http.MethodPatch, path, body)
	if err != nil {
		return nil, err
	}
	return c.Do(r)
}

// Delete performs a DELETE request to the specified path and returns the response body.
func (c *Client) Delete(ctx context.Context, path string) ([]byte, error) {
	r, err := c.createRequest(ctx, http.MethodDelete, path, nil)
	if err != nil {
		return nil, err
	}
	return c.Do(r)
}

// Do performs HTTP request and returns the response body.
func (c *Client) Do(r *http.Request) ([]byte, error) {
	response, err := c.config.httpClient.Do(r) //gosec:disable G704 -- request is constructed by trusted internal callers
	if err != nil {
		return nil, err
	}

	return c.handleResponse(response)
}

// DoStream performs HTTP request and returns the response body reader.
func (c *Client) DoStream(r *http.Request) (io.ReadCloser, error) {
	response, err := c.config.httpClient.Do(r) //gosec:disable G704 -- request is constructed by trusted internal callers
	if err != nil {
		return nil, err
	}

	return c.prepareResponse(response)
}

func (c *Client) createRequest(ctx context.Context, method, path string, body []byte) (*http.Request, error) {
	var bodyReader io.Reader

	if body != nil {
		bodyReader = bytes.NewBuffer(body)
	}
	req, err := http.NewRequestWithContext(ctx, method, c.createRequestURL(path), bodyReader)
	if err != nil {
		return nil, err
	}
	c.addDefaultHeaders(req)
	c.logRequest(req, body)
	return req, err
}

// prepareResponse prepares the response and returns either the response body or an error.
func (c *Client) prepareResponse(response *http.Response) (io.ReadCloser, error) {
	// Return an error on unsuccessful requests
	if response.StatusCode < 200 || response.StatusCode > 299 {
		defer response.Body.Close()

		errorBody, _ := io.ReadAll(response.Body)
		var errorType ErrorType
		switch response.Header.Get("Content-Type") {
		case "application/problem+json":
			errorType = ErrorTypeProblem
		default:
			errorType = ErrorTypeError
		}
		c.logResponse(response, errorBody)
		return nil, &Error{response.StatusCode, response.Status, errorBody, errorType}
	}

	return response.Body, nil
}

// handleResponse parses the response and returns either the response body or an error.
func (c *Client) handleResponse(response *http.Response) ([]byte, error) {
	body, err := c.prepareResponse(response)
	if err != nil {
		return nil, err
	}
	defer body.Close()

	responseBody, err := io.ReadAll(body)
	c.logResponse(response, responseBody)

	return responseBody, err
}

func (c *Client) addDefaultHeaders(r *http.Request) {
	const (
		accept        string = "Accept"
		contentType   string = "Content-Type"
		userAgent     string = "User-Agent"
		authorization string = "Authorization"
	)
	if _, ok := r.Header[accept]; !ok {
		r.Header.Set(accept, "application/json")
	}
	if _, ok := r.Header[contentType]; !ok {
		r.Header.Set(contentType, "application/json")
	}
	if _, ok := r.Header[userAgent]; !ok {
		r.Header.Set(userAgent, c.UserAgent)
	}
	if _, ok := r.Header[authorization]; !ok && strings.HasPrefix(r.URL.String(), c.config.baseURL) {
		if c.config.token != "" {
			r.Header.Set(authorization, "Bearer "+c.config.token)
		} else {
			r.SetBasicAuth(c.config.username, c.config.password)
		}
	}
}

// createRequestURL creates and returns a complete request URL for the specified API location using a newer API version.
func (c *Client) createRequestURL(location string) string {
	return fmt.Sprintf("%s%s", c.getBaseURL(), location)
}

// Returns the base URL to use for API requests
func (c *Client) getBaseURL() string {
	if c.config.baseURL == "" {
		c.config.baseURL = clientBaseURL(os.Getenv(EnvDebugAPIBaseURL))
	}
	return fmt.Sprintf("%s/%s", c.config.baseURL, APIVersion)
}

// Pretty prints given JSON bytes. If the JSON is not valid, returns the original bytes as string.
func prettyJSON(i []byte) string {
	var o bytes.Buffer
	if err := json.Indent(&o, i, "", "  "); err != nil {
		return string(i)
	}
	return o.String()
}

func (c *Client) logRequest(r *http.Request, body []byte) {
	const authorization string = "Authorization"

	if c.config.logger != nil {
		headers := r.Header.Clone()
		if _, ok := headers[authorization]; ok {
			auth := strings.Split(headers.Get(authorization), " ")
			// Redact the token part of the Authorization header or the whole value if there is no space to separate scheme from parameters.
			if len(auth) > 1 {
				headers.Set(authorization, fmt.Sprintf("%s [REDACTED]", auth[0]))
			} else {
				headers.Set(authorization, "[REDACTED]")
			}
		}

		c.config.logger(r.Context(), "Sending request to UpCloud API",
			"url", r.URL.Redacted(),
			"method", r.Method,
			"headers", headers,
			"body", prettyJSON(body),
		)
	}
}

func (c *Client) logResponse(r *http.Response, body []byte) {
	if c.config.logger != nil {
		c.config.logger(r.Request.Context(), "Received response from UpCloud API",
			"url", r.Request.URL.Redacted(),
			"status", r.Status,
			"headers", r.Header,
			"body", prettyJSON(body),
		)
	}
}

type ConfigFn func(o *config)

// WithBaseURL modifies the client baseURL
func WithBaseURL(baseURL string) ConfigFn {
	return func(c *config) {
		c.baseURL = baseURL
	}
}

// WithInsecureSkipVerify modifies the client's httpClient to skip verifying
// the server's certificate chain and host name. This should be used only for testing.
func WithInsecureSkipVerify() ConfigFn {
	return func(c *config) {
		if c.httpClient != nil {
			if t, ok := c.httpClient.Transport.(*http.Transport); ok {
				var cfg *tls.Config
				if t.TLSClientConfig == nil {
					cfg = NewDefaultTLSClientConfig()
				} else {
					cfg = t.TLSClientConfig.Clone()
				}
				cfg.InsecureSkipVerify = true
				t.TLSClientConfig = cfg
			}
		}
	}
}

// WithHTTPClient replaces the client's default httpClient with the specified one
func WithHTTPClient(httpClient *http.Client) ConfigFn {
	return func(c *config) {
		c.httpClient = httpClient
	}
}

// WithBasicAuth configures the client to use basic auth credentials for authentication
func WithBasicAuth(username, password string) ConfigFn {
	return func(c *config) {
		c.username = username
		c.password = password
		c.token = ""
	}
}

// WithBearerAuth (EXPERIMENTAL) configures the client to use bearer token for authentication
func WithBearerAuth(apiToken string) ConfigFn {
	return func(c *config) {
		c.token = apiToken
		c.username = ""
		c.password = ""
	}
}

// WithTimeout modifies the client's httpClient timeout
func WithTimeout(timeout time.Duration) ConfigFn {
	return func(c *config) {
		c.httpClient.Timeout = timeout
	}
}

// WithLogger configures logging function that logs requests and responses
func WithLogger(logger LogFn) ConfigFn {
	return func(c *config) {
		c.logger = logger
	}
}

// New creates and returns a new client configured with the specified user and password and optional
// config functions.
// TODO: we should get rid of username, password here, but it's a breaking change. Credentials can be now set with
// configurators client.WithBasicAuth("user", "pass") or client.WithBearerAuth("ucat_token")
func New(username, password string, c ...ConfigFn) *Client {
	config := config{
		username:   username,
		password:   password,
		baseURL:    clientBaseURL(os.Getenv(EnvDebugAPIBaseURL)),
		httpClient: NewDefaultHTTPClient(),
	}

	// If set, replace http client transport with one skipping tls verification
	if os.Getenv(EnvDebugSkipCertificateVerify) == "1" {
		c = append(c, WithInsecureSkipVerify())
	}

	for _, fn := range c {
		fn(&config)
	}
	return &Client{
		UserAgent: userAgent(),
		config:    config,
	}
}

func userAgent() string {
	return fmt.Sprintf("upcloud-go-api/%s", Version)
}

func clientBaseURL(URL string) string {
	if URL == "" {
		return APIBaseURL
	}

	if u, err := url.Parse(URL); err != nil || u.Scheme == "" || u.Host == "" {
		return APIBaseURL
	}

	return URL
}

// NewDefaultHTTPClient returns new default http.Client.
func NewDefaultHTTPClient() *http.Client {
	transport := NewDefaultHTTPTransport()
	return &http.Client{
		Transport: transport,
	}
}

// NewDefaultHTTPTransport return new HTTP client transport round tripper.
func NewDefaultHTTPTransport() http.RoundTripper {
	return &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		DisableKeepAlives:     true,
		MaxIdleConnsPerHost:   -1,
		TLSClientConfig:       NewDefaultTLSClientConfig(),
	}
}

func NewDefaultTLSClientConfig() *tls.Config {
	// Note: would be nice to be able to set this only if the default is lower,
	// but there does not seem to be a good way to check the default as of Go 1.24.
	return &tls.Config{
		MinVersion: tls.VersionTLS13,
	}
}
