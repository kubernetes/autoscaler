package nanny

import (
	"fmt"
	"io"
	"testing"

	dto "github.com/prometheus/client_model/go"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

func TestMergeResources(t *testing.T) {
	testCases := []struct {
		name    string
		current *corev1.ResourceRequirements
		new     *corev1.ResourceRequirements
		want    *corev1.ResourceRequirements
	}{
		{
			name:    "Overwrite empty",
			current: &corev1.ResourceRequirements{},
			new: &corev1.ResourceRequirements{
				Requests: corev1.ResourceList{"cpu": resource.MustParse("200m")},
				Limits:   corev1.ResourceList{"cpu": resource.MustParse("600m")},
			},
			want: &corev1.ResourceRequirements{
				Requests: corev1.ResourceList{"cpu": resource.MustParse("200m")},
				Limits:   corev1.ResourceList{"cpu": resource.MustParse("600m")},
			},
		}, {
			name: "Overwrite all",
			current: &corev1.ResourceRequirements{
				Requests: corev1.ResourceList{"cpu": resource.MustParse("100m")},
				Limits:   corev1.ResourceList{"cpu": resource.MustParse("200m")},
			},
			new: &corev1.ResourceRequirements{
				Requests: corev1.ResourceList{"cpu": resource.MustParse("200m")},
				Limits:   corev1.ResourceList{"cpu": resource.MustParse("600m")},
			},
			want: &corev1.ResourceRequirements{
				Requests: corev1.ResourceList{"cpu": resource.MustParse("200m")},
				Limits:   corev1.ResourceList{"cpu": resource.MustParse("600m")},
			},
		}, {
			name: "Add limits",
			current: &corev1.ResourceRequirements{
				Requests: corev1.ResourceList{"cpu": resource.MustParse("100m")},
			},
			new: &corev1.ResourceRequirements{
				Requests: corev1.ResourceList{"cpu": resource.MustParse("100m")},
				Limits:   corev1.ResourceList{"cpu": resource.MustParse("600m")},
			},
			want: &corev1.ResourceRequirements{
				Requests: corev1.ResourceList{"cpu": resource.MustParse("100m")},
				Limits:   corev1.ResourceList{"cpu": resource.MustParse("600m")},
			},
		}, {
			name: "Don't remove existing when new is empty",
			current: &corev1.ResourceRequirements{
				Requests: corev1.ResourceList{"cpu": resource.MustParse("100m")},
				Limits:   corev1.ResourceList{"cpu": resource.MustParse("200m")},
			},
			new: &corev1.ResourceRequirements{},
			want: &corev1.ResourceRequirements{
				Requests: corev1.ResourceList{"cpu": resource.MustParse("100m")},
				Limits:   corev1.ResourceList{"cpu": resource.MustParse("200m")},
			},
		}, {
			name: "Don't remove additional existing",
			current: &corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					"cpu":    resource.MustParse("100m"),
					"memory": resource.MustParse("300Mi"),
				},
				Limits: corev1.ResourceList{
					"cpu":    resource.MustParse("200m"),
					"memory": resource.MustParse("900Mi"),
				},
			},
			new: &corev1.ResourceRequirements{
				Requests: corev1.ResourceList{"cpu": resource.MustParse("200m")},
				Limits:   corev1.ResourceList{"cpu": resource.MustParse("600m")},
			},
			want: &corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					"cpu":    resource.MustParse("200m"),
					"memory": resource.MustParse("300Mi"),
				},
				Limits: corev1.ResourceList{
					"cpu":    resource.MustParse("600m"),
					"memory": resource.MustParse("900Mi"),
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := mergeResources(tc.current, tc.new)
			verifyResources(t, "limits", got.Limits, tc.want.Limits)
			verifyResources(t, "requests", got.Requests, tc.want.Requests)
		})
	}
}

func expectErrorOrCount(t *testing.T, expectErr error, expectCount uint64, gotErr error, gotCount uint64) {
	if expectErr == nil {
		if gotErr != nil {
			t.Errorf("expected no error, got %v", gotErr)
		}
		if gotCount != expectCount {
			t.Errorf("expected node count %v, got %v", expectCount, gotCount)
		}
		return
	}
	if gotErr == nil {
		t.Errorf("expected error %v, got nil", expectErr)
		return
	}
	if gotErr.Error() != expectErr.Error() {
		t.Errorf("expected error %v, got %v", expectErr, gotErr)
	}
}

func getMetric(labelName, labelValue string, value float64) *dto.Metric {
	return &dto.Metric{
		Label: []*dto.LabelPair{
			{
				Name:  &labelName,
				Value: &labelValue,
			},
		},
		Gauge: &dto.Gauge{
			Value: &value,
		},
	}
}

func TestExtractMetricValueForNodeCount(t *testing.T) {
	testCases := []struct {
		name        string
		mf          dto.MetricFamily
		expectCount uint64
		expectErr   error
	}{
		{
			name:      "empty",
			mf:        dto.MetricFamily{},
			expectErr: fmt.Errorf("metric name: no valid metric values"),
		},
		{
			name: "with proper value",
			mf: dto.MetricFamily{
				Metric: []*dto.Metric{
					getMetric("resource", "nodes", 4.0),
				},
			},
			expectCount: 4,
		},
		{
			name: "only wrong label",
			mf: dto.MetricFamily{
				Metric: []*dto.Metric{
					getMetric("wrong", "nodes", 4.0),
				},
			},
			expectErr: fmt.Errorf("metric name: no valid metric values"),
		},
		{
			name: "only wrong label value",
			mf: dto.MetricFamily{
				Metric: []*dto.Metric{
					getMetric("resource", "pods", 4.0),
				},
			},
			expectErr: fmt.Errorf("metric name: no valid metric values"),
		},
		{
			name: "with negative value",
			mf: dto.MetricFamily{
				Metric: []*dto.Metric{
					getMetric("resource", "nodes", -4.0),
				},
			},
			expectErr: fmt.Errorf("metric name: metric unknown"),
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			gotCount, gotErr := extractMetricValueForNodeCount(tc.mf, "metric name")
			expectErrorOrCount(t, tc.expectErr, tc.expectCount, gotErr, gotCount)
		})
	}
}

type fakeDecoder struct {
	metricValues []dto.MetricFamily
	finalResult  error
}

func (fd *fakeDecoder) Decode(output *dto.MetricFamily) error {
	if len(fd.metricValues) == 0 {
		return fd.finalResult
	}
	*output = fd.metricValues[len(fd.metricValues)-1]
	fd.metricValues = fd.metricValues[:len(fd.metricValues)-1]
	return nil
}

func TestGetNodeCountFromDecoder(t *testing.T) {
	preferredMetric := objectCountMetricName
	fallbackMetric := objectCountFallbackMetricName
	testCases := []struct {
		name         string
		metricValues []dto.MetricFamily
		finalResult  error
		expectValue  uint64
		expectErr    error
	}{
		{
			name:        "empty",
			finalResult: io.EOF,
			expectErr:   fmt.Errorf("no metric set"),
		},
		{
			name:        "with preferred metric",
			finalResult: io.EOF,
			metricValues: []dto.MetricFamily{
				{
					Name: &preferredMetric,
					Metric: []*dto.Metric{
						getMetric("resource", "nodes", 4.0),
					},
				},
			},
			expectValue: 4,
		},
		{
			name:        "with fallback metric",
			finalResult: io.EOF,
			metricValues: []dto.MetricFamily{
				{
					Name: &fallbackMetric,
					Metric: []*dto.Metric{
						getMetric("resource", "nodes", 4.0),
					},
				},
			},
			expectValue: 4,
		},
		{
			name:        "with preferred and fallback metrics",
			finalResult: io.EOF,
			metricValues: []dto.MetricFamily{
				{
					Name: &fallbackMetric,
					Metric: []*dto.Metric{
						getMetric("resource", "nodes", 4.0),
					},
				},
				{
					Name: &preferredMetric,
					Metric: []*dto.Metric{
						getMetric("resource", "nodes", 5.0),
					},
				},
			},
			expectValue: 5,
		},
		{
			name:        "with error",
			finalResult: fmt.Errorf("oops"),
			expectErr:   fmt.Errorf("decoding error: oops"),
		},
		{
			name:        "fallbacks on error preferred on preferred metric",
			finalResult: io.EOF,
			metricValues: []dto.MetricFamily{
				{
					Name: &fallbackMetric,
					Metric: []*dto.Metric{
						getMetric("resource", "nodes", 4.0),
					},
				},
				{
					Name: &preferredMetric,
					Metric: []*dto.Metric{
						getMetric("wrong", "nodes", 5.0),
					},
				},
			},
			expectValue: 4,
		},
		{
			name:        "reports error on both metrics",
			finalResult: io.EOF,
			metricValues: []dto.MetricFamily{
				{
					Name: &fallbackMetric,
					Metric: []*dto.Metric{
						getMetric("wrong", "nodes", 4.0),
					},
				},
				{
					Name: &preferredMetric,
					Metric: []*dto.Metric{
						getMetric("wrong", "nodes", 5.0),
					},
				},
			},
			expectErr: fmt.Errorf("at least one metric present but all present metrics have errors: apiserver_storage_objects: no valid metric values, etcd_object_counts: no valid metric values"),
		},
		{
			name:        "multiple metrics in preferred metric family",
			finalResult: io.EOF,
			metricValues: []dto.MetricFamily{
				{
					Name: &preferredMetric,
					Metric: []*dto.Metric{
						getMetric("resource", "nodes", 5.0),
					},
				},
				{
					Name: &preferredMetric,
					Metric: []*dto.Metric{
						getMetric("resource", "foobars", 3.0),
					},
				},
			},
			expectValue: 5.0,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fd := fakeDecoder{
				metricValues: tc.metricValues,
				finalResult:  tc.finalResult,
			}
			gotCount, gotErr := getNodeCountFromDecoder(&fd)
			expectErrorOrCount(t, tc.expectErr, tc.expectValue, gotErr, gotCount)
		})
	}
}

func TestGetNodeCountFromDecoder_MultpleLabels(t *testing.T) {
	preferredMetric := objectCountMetricName
	labelName1 := resourceLabel
	labelValue1 := nodeResourceName
	labelName2 := "flavor"
	labelValue2 := "up"
	value := 3.0
	fd := fakeDecoder{
		metricValues: []dto.MetricFamily{
			{
				Name: &preferredMetric,
				Metric: []*dto.Metric{
					{
						Label: []*dto.LabelPair{
							{
								Name:  &labelName1,
								Value: &labelValue1,
							},
							{
								Name:  &labelName2,
								Value: &labelValue2,
							},
						},
						Gauge: &dto.Gauge{
							Value: &value,
						},
					},
				},
			},
		},
		finalResult: io.EOF,
	}
	gotCount, gotErr := getNodeCountFromDecoder(&fd)
	expectErrorOrCount(t, nil, 3.0, gotErr, gotCount)
}
