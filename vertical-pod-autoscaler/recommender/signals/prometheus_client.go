/*
Copyright 2018 The Kubernetes Authors.

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

package signals

import (
	"fmt"
	"net/http"
	"net/url"
	"time"
)

// PrometheusClient talks to Prometheus using its HTTP API.
type PrometheusClient interface {
	// Given a particular query (that's supposed to return range vectors
	// in Prometheus terminology), gets the results from Prometheus.
	GetTimeseries(query string) ([]Timeseries, error)
}

// An implementation of PrometheusClient.
type prometheusClient struct {
	httpClient *http.Client
	address    string
}

// NewPrometheusClient constructs a prometheusClient.
func NewPrometheusClient(httpClient *http.Client, address string) PrometheusClient {
	return &prometheusClient{httpClient: httpClient, address: address}
}

// Changes Prometheus address and query into a full escaped URL to call.
func getUrlWithQuery(address, query string) (string, error) {
	url, err := url.Parse(address)
	if err != nil {
		return "", err
	}
	url.Path = "api/v1/query"
	queryValues := url.Query()
	queryValues.Set("query", query)
	url.RawQuery = queryValues.Encode()
	return url.String(), nil
}

func retry(callback func() error, attempts int, delay time.Duration) error {
	for i := 1; ; i++ {
		err := callback()
		if err == nil {
			return nil
		}
		if i >= attempts {
			return fmt.Errorf("tried %d times, last error: %v", attempts, err)
		}
		time.Sleep(delay)
	}
}

func (c *prometheusClient) GetTimeseries(query string) ([]Timeseries, error) {
	url, err := getUrlWithQuery(c.address, query)
	if err != nil {
		return nil, fmt.Errorf("couldn't construct url to Prometheus: %v", err)
	}
	var resp *http.Response
	err = retry(func() error {
		resp, err = c.httpClient.Get(url)
		if err != nil {
			return fmt.Errorf("error getting data from Prometheus: %v", err)
		}
		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("bad HTTP status: %v %s", resp.StatusCode, http.StatusText(resp.StatusCode))
		}
		return nil
	}, 10, 3*time.Second)
	return decodeTimeseriesFromResponse(resp.Body)
}
