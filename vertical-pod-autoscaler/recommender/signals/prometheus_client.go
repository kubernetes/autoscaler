/*
Copyright 2017 The Kubernetes Authors.
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
)

// A client that talks to Prometheus using its HTTP API.
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

func NewPrometheusClient(httpClient *http.Client, address string) PrometheusClient {
	return &prometheusClient{httpClient: httpClient, address: address}
}

// Changes Prometheus address and query into a full escaped URL to call.
func getUrlWithQuery(address, query string) (string, error) {
	url, err := url.Parse(address)
	if err != nil {
		return "", err
	}
	queryValues := url.Query()
	queryValues.Set("query", query)
	url.RawQuery = queryValues.Encode()
	return url.String(), nil
}

func (c *prometheusClient) GetTimeseries(query string) ([]Timeseries, error) {
	url, err := getUrlWithQuery(c.address, query)
	if err != nil {
		return nil, fmt.Errorf("couldn't construct url to Prometheus: %v", err)
	}
	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("error getting data from Prometheus: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad HTTP status: %v %s", resp.StatusCode, http.StatusText(resp.StatusCode))
	}
	return decodeTimeseriesFromResponse(resp.Body)
}
