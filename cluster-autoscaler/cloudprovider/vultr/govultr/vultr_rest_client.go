/*
Copyright 2022 The Kubernetes Authors.

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

package govultr

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
)

// Client that is used for HTTP requests
type Client struct {
	httpClient *http.Client
	baseURL    *url.URL
	userAgent  string
}

// NewClient returns a client struct
func NewClient(client *http.Client) *Client {
	// do something better here
	u, _ := url.Parse("https://api.vultr.com/v2")
	return &Client{
		httpClient: client,
		baseURL:    u,
		userAgent:  "kubernetes/cluster-autoscaler",
	}
}

// SetBaseUrl sets the base URL
func (c *Client) SetBaseUrl(baseURL string) (*Client, error) {
	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, err
	}

	c.baseURL = u

	return c, nil
}

// SetUserAgent sets the user-agent for HTTP requests
func (c *Client) SetUserAgent(userAgent string) *Client {
	c.userAgent = userAgent
	return c
}

func (c *Client) newRequest(ctx context.Context, method, uri string, body interface{}) (*http.Request, error) {
	resolvedURL, err := c.baseURL.Parse(uri)
	if err != nil {
		return nil, err
	}

	buf := new(bytes.Buffer)
	if body != nil {
		if err = json.NewEncoder(buf).Encode(body); err != nil {
			return nil, err
		}
	}

	req, err := http.NewRequest(method, resolvedURL.String(), buf)
	if err != nil {
		return nil, err
	}

	req.Header.Add("User-Agent", c.userAgent)
	req.Header.Add("Content-Type", "application/json")

	return req, nil
}

func (c *Client) doWithContext(ctx context.Context, r *http.Request, data interface{}) error {
	req := r.WithContext(ctx)
	res, err := c.httpClient.Do(req)

	if err != nil {
		return err
	}

	//todo handle this
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}

	if res.StatusCode >= http.StatusOK && res.StatusCode <= http.StatusNoContent {
		if data != nil {
			if err := json.Unmarshal(body, data); err != nil {
				return err
			}
		}
		return nil
	}

	//todo make into errors struct?
	return errors.New(string(body))
}
