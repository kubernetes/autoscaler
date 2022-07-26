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

package hetzner

import (
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	k8smetrics "k8s.io/component-base/metrics"
	"k8s.io/component-base/metrics/legacyregistry"
)

const subsystemIdentifier = "api"

func instrumentedRoundTripper() http.RoundTripper {
	inFlightRequestsGauge := k8smetrics.NewGauge(&k8smetrics.GaugeOpts{
		Name: fmt.Sprintf("hcloud_%s_in_flight_requests", subsystemIdentifier),
		Help: fmt.Sprintf("A gauge of in-flight requests to the hcloud %s.", subsystemIdentifier),
	})

	requestsPerEndpointCounter := k8smetrics.NewCounterVec(
		&k8smetrics.CounterOpts{
			Name: fmt.Sprintf("hcloud_%s_requests_total", subsystemIdentifier),
			Help: fmt.Sprintf("A counter for requests to the hcloud %s per endpoint.", subsystemIdentifier),
		},
		[]string{"code", "method", "api_endpoint"},
	)

	requestLatencyHistogram := k8smetrics.NewHistogramVec(
		&k8smetrics.HistogramOpts{
			Name:    fmt.Sprintf("hcloud_%s_request_duration_seconds", subsystemIdentifier),
			Help:    fmt.Sprintf("A histogram of request latencies to the hcloud %s .", subsystemIdentifier),
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method"},
	)

	legacyregistry.MustRegister(requestsPerEndpointCounter)
	legacyregistry.MustRegister(requestLatencyHistogram)
	legacyregistry.MustRegister(inFlightRequestsGauge)

	return instrumentRoundTripperInFlight(inFlightRequestsGauge,
		instrumentRoundTripperDuration(requestLatencyHistogram,
			instrumentRoundTripperEndpoint(requestsPerEndpointCounter,
				http.DefaultTransport,
			),
		),
	)
}

type roundTripperFunc func(req *http.Request) (*http.Response, error)

// RoundTrip implements the RoundTripper interface.
func (rt roundTripperFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return rt(r)
}

func instrumentRoundTripperInFlight(gauge *k8smetrics.Gauge, next http.RoundTripper) roundTripperFunc {
	return roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		gauge.Inc()
		defer gauge.Dec()
		return next.RoundTrip(r)
	})
}

func instrumentRoundTripperDuration(obs *k8smetrics.HistogramVec, next http.RoundTripper) roundTripperFunc {
	return roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		start := time.Now()
		resp, err := next.RoundTrip(r)
		if err == nil {
			obs.WithLabelValues(strings.ToLower(resp.Request.Method)).Observe(time.Since(start).Seconds())
		}
		return resp, err
	})
}

func instrumentRoundTripperEndpoint(counter *k8smetrics.CounterVec, next http.RoundTripper) promhttp.RoundTripperFunc {
	return func(r *http.Request) (*http.Response, error) {
		resp, err := next.RoundTrip(r)
		if err == nil {
			statusCode := strconv.Itoa(resp.StatusCode)
			counter.WithLabelValues(statusCode, strings.ToLower(resp.Request.Method), preparePathForLabel(resp.Request.URL.Path)).Inc()
		}
		return resp, err
	}
}

func preparePathForLabel(path string) string {
	path = strings.ToLower(path)

	// replace all numbers and chars that are not a-z, / or _
	reg := regexp.MustCompile("[^a-z/_]+")
	path = reg.ReplaceAllString(path, "")

	// replace all artifacts of number replacement (//)
	path = strings.ReplaceAll(path, "//", "/")

	// replace the /v/ that indicated the API version
	return strings.Replace(path, "/v/", "/", 1)
}
