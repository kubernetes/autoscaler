/*
Copyright 2023 The Kubernetes Authors.

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

package recommender

import (
	"errors"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
)

// fakeRoundTripper is a simple http.RoundTripper that always returns 200 OK.
type fakeRoundTripper struct{}

func (fakeRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	return &http.Response{
		StatusCode: http.StatusOK,
		Body:       http.NoBody,
		Request:    req,
	}, nil
}

func TestMetricsHelpers(t *testing.T) {
	reg := prometheus.NewPedanticRegistry()
	reg.MustRegister(recommendationLatency, aggregateContainerStatesCount, metricServerResponses,
		prometheusClientRequestsCount, prometheusClientRequestsDuration)

	t.Run("ObserveRecommendationLatency", func(t *testing.T) {
		// Get initial sample count
		initialCh := make(chan prometheus.Metric, 1)
		recommendationLatency.Collect(initialCh)
		initialMetric := <-initialCh
		initialDto := &dto.Metric{}
		if err := initialMetric.Write(initialDto); err != nil {
			t.Fatalf("failed to read initial metric: %v", err)
		}
		initialCount := initialDto.GetHistogram().GetSampleCount()

		ObserveRecommendationLatency(time.Now().Add(-2 * time.Second))

		m := &dto.Metric{}
		ch := make(chan prometheus.Metric, 1)
		recommendationLatency.Collect(ch)
		metric := <-ch
		if err := metric.Write(m); err != nil {
			t.Fatalf("failed to read metric: %v", err)
		}
		h := m.GetHistogram()
		if h.GetSampleCount() != initialCount+1 {
			t.Errorf("expected %d samples, got %d", initialCount+1, h.GetSampleCount())
		}
	})

	t.Run("RecordAggregateContainerStatesCount", func(t *testing.T) {
		RecordAggregateContainerStatesCount(5)

		m := &dto.Metric{}
		ch := make(chan prometheus.Metric, 1)
		aggregateContainerStatesCount.Collect(ch)
		metric := <-ch
		if err := metric.Write(m); err != nil {
			t.Fatalf("failed to read metric: %v", err)
		}
		if v := m.GetGauge().GetValue(); v != 5 {
			t.Errorf("expected gauge 5, got %f", v)
		}
	})

	t.Run("RecordMetricsServerResponse", func(t *testing.T) {
		// Get initial counter values
		initialCounts := map[string]float64{}
		m := &dto.Metric{}
		ch := make(chan prometheus.Metric, 10) // larger buffer for any existing metrics
		go func() {
			metricServerResponses.Collect(ch)
			close(ch)
		}()
		for metric := range ch {
			if err := metric.Write(m); err != nil {
				t.Fatalf("failed to read metric: %v", err)
			}
			key := labelsToKey(m.GetLabel())
			initialCounts[key] = m.GetCounter().GetValue()
		}

		RecordMetricsServerResponse(nil, "test")
		RecordMetricsServerResponse(errors.New("boom"), "test")

		ch = make(chan prometheus.Metric, 10)
		go func() {
			metricServerResponses.Collect(ch)
			close(ch)
		}()
		finalCounts := map[string]float64{}
		for metric := range ch {
			if err := metric.Write(m); err != nil {
				t.Fatalf("failed to read metric: %v", err)
			}
			key := labelsToKey(m.GetLabel())
			finalCounts[key] = m.GetCounter().GetValue()
		}

		successKey := "client_name=test,is_error=false,"
		errorKey := "client_name=test,is_error=true,"

		expectedSuccess := initialCounts[successKey] + 1
		expectedError := initialCounts[errorKey] + 1

		if finalCounts[successKey] != expectedSuccess {
			t.Errorf("expected success counter %f, got %f", expectedSuccess, finalCounts[successKey])
		}
		if finalCounts[errorKey] != expectedError {
			t.Errorf("expected error counter %f, got %f", expectedError, finalCounts[errorKey])
		}
	})

	t.Run("NewPrometheusRoundTripperCounter", func(t *testing.T) {
		// Get initial counter values
		initialCh := make(chan prometheus.Metric, 10)
		go func() {
			prometheusClientRequestsCount.Collect(initialCh)
			close(initialCh)
		}()
		initialCount := uint64(0)
		m := &dto.Metric{}
		for metric := range initialCh {
			if err := metric.Write(m); err != nil {
				t.Fatalf("failed to read initial metric: %v", err)
			}
			initialCount += uint64(m.GetCounter().GetValue())
		}

		rt := NewPrometheusRoundTripperCounter(fakeRoundTripper{})
		req := &http.Request{Method: http.MethodGet, URL: &url.URL{Scheme: "http", Host: "example.com"}}
		if _, err := rt.RoundTrip(req); err != nil {
			t.Fatalf("RoundTrip failed: %v", err)
		}

		ch := make(chan prometheus.Metric, 10)
		go func() {
			prometheusClientRequestsCount.Collect(ch)
			close(ch)
		}()
		finalCount := uint64(0)
		for metric := range ch {
			if err := metric.Write(m); err != nil {
				t.Fatalf("failed to read metric: %v", err)
			}
			finalCount += uint64(m.GetCounter().GetValue())
		}
		if finalCount != initialCount+1 {
			t.Errorf("expected counter increase by 1, initial: %d, final: %d", initialCount, finalCount)
		}
	})

	t.Run("NewPrometheusRoundTripperDuration", func(t *testing.T) {
		// Get initial histogram sample count
		initialCh := make(chan prometheus.Metric, 10)
		go func() {
			prometheusClientRequestsDuration.Collect(initialCh)
			close(initialCh)
		}()
		initialTotalCount := uint64(0)
		m := &dto.Metric{}
		for metric := range initialCh {
			if err := metric.Write(m); err != nil {
				t.Fatalf("failed to read initial metric: %v", err)
			}
			initialTotalCount += m.GetHistogram().GetSampleCount()
		}

		rt := NewPrometheusRoundTripperDuration(fakeRoundTripper{})
		req := &http.Request{Method: http.MethodGet, URL: &url.URL{Scheme: "http", Host: "example.com"}}
		if _, err := rt.RoundTrip(req); err != nil {
			t.Fatalf("RoundTrip failed: %v", err)
		}

		ch := make(chan prometheus.Metric, 10)
		go func() {
			prometheusClientRequestsDuration.Collect(ch)
			close(ch)
		}()
		finalTotalCount := uint64(0)
		for metric := range ch {
			if err := metric.Write(m); err != nil {
				t.Fatalf("failed to read metric: %v", err)
			}
			finalTotalCount += m.GetHistogram().GetSampleCount()
		}
		if finalTotalCount != initialTotalCount+1 {
			t.Errorf("expected histogram count increase by 1, initial: %d, final: %d", initialTotalCount, finalTotalCount)
		}
	})
}
