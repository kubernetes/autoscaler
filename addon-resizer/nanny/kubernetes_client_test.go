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
			t.Errorf("expected resource count %v, got %v", expectCount, gotCount)
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

func TestExtractMetricValueForResourceCount(t *testing.T) {
	testCases := []struct {
		name         string
		mf           dto.MetricFamily
		resourceName string
		expectCount  uint64
		expectErr    error
	}{
		{
			name:         "(nodes) empty",
			resourceName: nodeResourceName,
			mf:           dto.MetricFamily{},
			expectErr:    fmt.Errorf("metric name: no valid metric values"),
		},
		{
			name:         "(nodes) with proper value",
			resourceName: nodeResourceName,
			mf: dto.MetricFamily{
				Metric: []*dto.Metric{
					getMetric("resource", "nodes", 4.0),
				},
			},
			expectCount: 4,
		},
		{
			name:         "(nodes) only wrong label",
			resourceName: nodeResourceName,
			mf: dto.MetricFamily{
				Metric: []*dto.Metric{
					getMetric("wrong", "nodes", 4.0),
				},
			},
			expectErr: fmt.Errorf("metric name: no valid metric values"),
		},
		{
			name:         "(nodes) only wrong label value",
			resourceName: nodeResourceName,
			mf: dto.MetricFamily{
				Metric: []*dto.Metric{
					getMetric("resource", "services", 4.0),
				},
			},
			expectErr: fmt.Errorf("metric name: no valid metric values"),
		},
		{
			name:         "(nodes) with negative value",
			resourceName: nodeResourceName,
			mf: dto.MetricFamily{
				Metric: []*dto.Metric{
					getMetric("resource", "nodes", -4.0),
				},
			},
			expectErr: fmt.Errorf("metric name: metric unknown"),
		},
		{
			name:         "(pods) empty",
			resourceName: podResourceName,
			mf:           dto.MetricFamily{},
			expectErr:    fmt.Errorf("metric name: no valid metric values"),
		},
		{
			name:         "(pods) with proper value",
			resourceName: podResourceName,
			mf: dto.MetricFamily{
				Metric: []*dto.Metric{
					getMetric("resource", "pods", 6.0),
				},
			},
			expectCount: 6,
		},
		{
			name:         "(pods) only wrong label",
			resourceName: podResourceName,
			mf: dto.MetricFamily{
				Metric: []*dto.Metric{
					getMetric("wrong", "pods", 4.0),
				},
			},
			expectErr: fmt.Errorf("metric name: no valid metric values"),
		},
		{
			name:         "(pods) only wrong label value",
			resourceName: podResourceName,
			mf: dto.MetricFamily{
				Metric: []*dto.Metric{
					getMetric("resource", "services", 4.0),
				},
			},
			expectErr: fmt.Errorf("metric name: no valid metric values"),
		},
		{
			name:         "(pods) with negative value",
			resourceName: podResourceName,
			mf: dto.MetricFamily{
				Metric: []*dto.Metric{
					getMetric("resource", "pods", -4.0),
				},
			},
			expectErr: fmt.Errorf("metric name: metric unknown"),
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			gotCount, gotErr := extractMetricValueForResourceCount(tc.mf, tc.resourceName, "metric name")
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

func TestGetResourceCountFromDecoder(t *testing.T) {
	preferredMetric := objectCountMetricName
	fallbackMetric := objectCountFallbackMetricName
	testCases := []struct {
		name         string
		metricValues []dto.MetricFamily
		resourceName string
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
			name:         "(nodes) with preferred metric",
			finalResult:  io.EOF,
			resourceName: nodeResourceName,
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
			name:         "(nodes) with fallback metric",
			finalResult:  io.EOF,
			resourceName: nodeResourceName,
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
			name:         "(nodes) with preferred and fallback metrics",
			finalResult:  io.EOF,
			resourceName: nodeResourceName,
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
			name:         "(nodes) falls back on error in preferred metric",
			finalResult:  io.EOF,
			resourceName: nodeResourceName,
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
			name:         "(nodes) reports error on both metrics",
			finalResult:  io.EOF,
			resourceName: nodeResourceName,
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
			name:         "(nodes) multiple metrics in preferred metric family",
			finalResult:  io.EOF,
			resourceName: nodeResourceName,
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
		{
			name:         "(pods) with preferred metric",
			finalResult:  io.EOF,
			resourceName: podResourceName,
			metricValues: []dto.MetricFamily{
				{
					Name: &preferredMetric,
					Metric: []*dto.Metric{
						getMetric("resource", "pods", 10.0),
					},
				},
			},
			expectValue: 10,
		},
		{
			name:         "(pods) with fallback metric",
			finalResult:  io.EOF,
			resourceName: podResourceName,
			metricValues: []dto.MetricFamily{
				{
					Name: &fallbackMetric,
					Metric: []*dto.Metric{
						getMetric("resource", "pods", 10.0),
					},
				},
			},
			expectValue: 10,
		},
		{
			name:         "(pods) with preferred and fallback metrics",
			finalResult:  io.EOF,
			resourceName: podResourceName,
			metricValues: []dto.MetricFamily{
				{
					Name: &fallbackMetric,
					Metric: []*dto.Metric{
						getMetric("resource", "pods", 10.0),
					},
				},
				{
					Name: &preferredMetric,
					Metric: []*dto.Metric{
						getMetric("resource", "pods", 15.0),
					},
				},
			},
			expectValue: 15,
		},
		{
			name:         "(pods) falls back on error in preferred metric",
			finalResult:  io.EOF,
			resourceName: podResourceName,
			metricValues: []dto.MetricFamily{
				{
					Name: &fallbackMetric,
					Metric: []*dto.Metric{
						getMetric("resource", "pods", 10.0),
					},
				},
				{
					Name: &preferredMetric,
					Metric: []*dto.Metric{
						getMetric("wrong", "nodes", 5.0),
					},
				},
			},
			expectValue: 10,
		},
		{
			name:         "(pods) reports error on both metrics",
			finalResult:  io.EOF,
			resourceName: podResourceName,
			metricValues: []dto.MetricFamily{
				{
					Name: &fallbackMetric,
					Metric: []*dto.Metric{
						getMetric("wrong", "pods", 10.0),
					},
				},
				{
					Name: &preferredMetric,
					Metric: []*dto.Metric{
						getMetric("wrong", "pods", 15.0),
					},
				},
			},
			expectErr: fmt.Errorf("at least one metric present but all present metrics have errors: apiserver_storage_objects: no valid metric values, etcd_object_counts: no valid metric values"),
		},
		{
			name:         "(pods) multiple metrics in preferred metric family",
			finalResult:  io.EOF,
			resourceName: podResourceName,
			metricValues: []dto.MetricFamily{
				{
					Name: &preferredMetric,
					Metric: []*dto.Metric{
						getMetric("resource", "pods", 10.0),
					},
				},
				{
					Name: &preferredMetric,
					Metric: []*dto.Metric{
						getMetric("resource", "foobars", 3.0),
					},
				},
			},
			expectValue: 10.0,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fd := fakeDecoder{
				metricValues: tc.metricValues,
				finalResult:  tc.finalResult,
			}
			gotCount, gotErr := getResourceCountFromDecoder(tc.resourceName, &fd)
			expectErrorOrCount(t, tc.expectErr, tc.expectValue, gotErr, gotCount)
		})
	}
}

func TestGetResourceCountFromDecoder_MultpleLabels(t *testing.T) {
	preferredMetric := objectCountMetricName
	value := 3.0
	testCases := []struct {
		name         string
		labelName1   string
		labelValue1  string
		labelName2   string
		labelValue2  string
		resourceName string
	}{
		{
			name:         "(nodes) get metric count with multiple labels",
			labelName1:   resourceLabel,
			labelValue1:  nodeResourceName,
			labelName2:   "flavor",
			labelValue2:  "up",
			resourceName: nodeResourceName,
		},
		{
			name:         "(pods) get metric count with multiple labels",
			labelName1:   resourceLabel,
			labelValue1:  podResourceName,
			labelName2:   "flavor",
			labelValue2:  "up",
			resourceName: podResourceName,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fd := fakeDecoder{
				metricValues: []dto.MetricFamily{
					{
						Name: &preferredMetric,
						Metric: []*dto.Metric{
							{
								Label: []*dto.LabelPair{
									{
										Name:  &tc.labelName1,
										Value: &tc.labelValue1,
									},
									{
										Name:  &tc.labelName2,
										Value: &tc.labelValue2,
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
			gotCount, gotErr := getResourceCountFromDecoder(tc.resourceName, &fd)
			expectErrorOrCount(t, nil, 3.0, gotErr, gotCount)
		})
	}
}
