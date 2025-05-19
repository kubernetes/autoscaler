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
		recommendationLatency.Reset()
		ObserveRecommendationLatency(time.Now().Add(-2 * time.Second))

		m := &dto.Metric{}
		ch := make(chan prometheus.Metric, 1)
		recommendationLatency.Collect(ch)
		metric := <-ch
		if err := metric.Write(m); err != nil {
			t.Fatalf("failed to read metric: %v", err)
		}
		h := m.GetHistogram()
		if h.GetSampleCount() != 1 {
			t.Errorf("expected 1 sample, got %d", h.GetSampleCount())
		}
		recommendationLatency.Reset()
	})

	t.Run("RecordAggregateContainerStatesCount", func(t *testing.T) {
		aggregateContainerStatesCount.Reset()
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
		aggregateContainerStatesCount.Reset()
	})

	t.Run("RecordMetricsServerResponse", func(t *testing.T) {
		metricServerResponses.Reset()
		RecordMetricsServerResponse(nil, "test")
		RecordMetricsServerResponse(errors.New("boom"), "test")

		m := &dto.Metric{}
		ch := make(chan prometheus.Metric, 2)
		go func() {
			metricServerResponses.Collect(ch)
			close(ch)
		}()
		counts := map[string]float64{}
		for metric := range ch {
			if err := metric.Write(m); err != nil {
				t.Fatalf("failed to read metric: %v", err)
			}
			key := labelsToKey(m.GetLabel())
			counts[key] = m.GetCounter().GetValue()
		}

		if counts["client_name=test,is_error=false,"] != 1 {
			t.Errorf("expected success counter 1, got %f", counts["client_name=test,is_error=false,"])
		}
		if counts["client_name=test,is_error=true,"] != 1 {
			t.Errorf("expected error counter 1, got %f", counts["client_name=test,is_error=true,"])
		}
		metricServerResponses.Reset()
	})

	t.Run("NewPrometheusRoundTripperCounter", func(t *testing.T) {
		prometheusClientRequestsCount.Reset()
		rt := NewPrometheusRoundTripperCounter(fakeRoundTripper{})
		req := &http.Request{Method: http.MethodGet, URL: &url.URL{Scheme: "http", Host: "example.com"}}
		if _, err := rt.RoundTrip(req); err != nil {
			t.Fatalf("RoundTrip failed: %v", err)
		}

		ch := make(chan prometheus.Metric, 1)
		prometheusClientRequestsCount.Collect(ch)
		m := &dto.Metric{}
		metric := <-ch
		if err := metric.Write(m); err != nil {
			t.Fatalf("failed to read metric: %v", err)
		}
		if c := m.GetCounter().GetValue(); c != 1 {
			t.Errorf("expected counter 1, got %f", c)
		}
		prometheusClientRequestsCount.Reset()
	})

	t.Run("NewPrometheusRoundTripperDuration", func(t *testing.T) {
		prometheusClientRequestsDuration.Reset()
		rt := NewPrometheusRoundTripperDuration(fakeRoundTripper{})
		req := &http.Request{Method: http.MethodGet, URL: &url.URL{Scheme: "http", Host: "example.com"}}
		if _, err := rt.RoundTrip(req); err != nil {
			t.Fatalf("RoundTrip failed: %v", err)
		}

		ch := make(chan prometheus.Metric, 1)
		prometheusClientRequestsDuration.Collect(ch)
		m := &dto.Metric{}
		metric := <-ch
		if err := metric.Write(m); err != nil {
			t.Fatalf("failed to read metric: %v", err)
		}
		if h := m.GetHistogram().GetSampleCount(); h != 1 {
			t.Errorf("expected histogram count 1, got %d", h)
		}
		prometheusClientRequestsDuration.Reset()
	})
}
