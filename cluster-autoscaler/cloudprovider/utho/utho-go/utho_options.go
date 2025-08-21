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
	"errors"
	"net/http"
)

// UthoOption describes a functional parameter for the utho client constructor
type UthoOption func(*client) error

// WithHTTPClient allows the overriding of the http client
func WithHTTPClient(httpClient *http.Client) UthoOption {
	return func(c *client) error {
		if httpClient == nil {
			return errors.New("http client can't be nil")
		}

		c.client = httpClient
		return nil
	}
}

// WithBaseURL allows the overriding of the base URL
func WithBaseURL(rawURL string) UthoOption {
	return func(c *client) error {
		if rawURL == "" {
			return errors.New("base url can't be empty")
		}

		baseURL, err := toURLWithEndingSlash(rawURL)
		if err != nil {
			return err
		}

		c.baseURL = baseURL
		return nil
	}
}
