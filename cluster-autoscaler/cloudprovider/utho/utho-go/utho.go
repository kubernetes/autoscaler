/*
Copyright 2025 The Kubernetes Authors.

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

package utho

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// BaseUrl is the base endpoint for the Utho API.
const BaseUrl = "https://api.utho.com/v2/"

var defaultHTTPClient = &http.Client{Timeout: time.Second * 300}

// Client is the interface for interacting with the Utho API client.
type Client interface {
	// NewRequest creates a new API request.
	NewRequest(method, url string, body ...interface{}) (*http.Request, error)
	// Do sends an API request and decodes the response.
	Do(req *http.Request, v interface{}) (*http.Response, error)

	// Kubernetes returns the KubernetesService for Kubernetes-related API calls.
	Kubernetes() *KubernetesService
}

type service struct {
	client Client
}

type client struct {
	client  *http.Client
	baseURL *url.URL
	token   string

	kubernetes *KubernetesService
}

// NewClient creates a new Utho client.
// Because the token supplied will be used for all authenticated requests,
// the created client should not be used across different users
func NewClient(token string, options ...UthoOption) (Client, error) {
	if token == "" {
		return nil, errors.New("you must provide an API token")
	}

	defaultBaseURL, err := toURLWithEndingSlash(BaseUrl)
	if err != nil {
		return nil, err
	}

	client := &client{
		client:  defaultHTTPClient,
		baseURL: defaultBaseURL,
		token:   token,
	}

	for _, option := range options {
		if err = option(client); err != nil {
			return nil, err
		}
	}

	commonService := &service{client: client}
	client.kubernetes = (*KubernetesService)(commonService)

	return client, nil
}

// toURLWithEndingSlash parses the given URL string and ensures the path ends with a slash.
func toURLWithEndingSlash(u string) (*url.URL, error) {
	baseURL, err := url.Parse(u)
	if err != nil {
		return nil, err
	}

	if !strings.HasSuffix(baseURL.Path, "/") {
		baseURL.Path += "/"
	}

	return baseURL, err
}

// NewRequest creates an API request.
// A relative URL `url` can be specified which is resolved relative to the baseURL of the client.
// Relative URLs should be specified without a preceding slash.
// The `body` parameter can be used to pass a body to the request. If no body is required, the parameter can be omitted.
func (c *client) NewRequest(method, url string, body ...interface{}) (*http.Request, error) {
	fullUrl, err := c.baseURL.Parse(url)
	if err != nil {
		return nil, err
	}

	var buf io.ReadWriter
	if len(body) > 0 && body[0] != nil {
		buf = &bytes.Buffer{}
		enc := json.NewEncoder(buf)
		enc.SetEscapeHTML(false)
		err := enc.Encode(body[0])
		if err != nil {
			return nil, err
		}
	}

	req, err := http.NewRequest(method, fullUrl.String(), buf)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept-Encoding", "application/json")

	return req, nil
}

// Do will send the given request using the client `c` on which it is called.
// If the response contains a body, it will be unmarshalled in `v`.
func (c *client) Do(req *http.Request, v interface{}) (*http.Response, error) {
	req.Header.Set("Authorization", "Bearer "+c.token)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	err = checkForErrors(resp)
	if err != nil {
		return resp, err
	}

	if resp.Body != nil && v != nil {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return resp, err
		}

		err = json.Unmarshal(body, &v)
		if err != nil {
			return resp, err
		}
	}

	return resp, nil
}

// checkForErrors inspects the HTTP response and returns an error if the status code indicates failure.
func checkForErrors(resp *http.Response) error {
	if c := resp.StatusCode; c >= 200 && c < 400 {
		return nil
	}

	errorResponse := &ErrorResponse{Response: resp}

	data, err := io.ReadAll(resp.Body)
	if err == nil && data != nil {
		// it's ok if we cannot unmarshal to Utho's error response
		_ = json.Unmarshal(data, errorResponse)
	}

	return errorResponse
}

// Kubernetes returns the KubernetesService for Kubernetes-related API calls.
func (c *client) Kubernetes() *KubernetesService {
	return c.kubernetes
}
