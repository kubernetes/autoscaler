/*
Copyright The Kubernetes Authors.

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

package comparator

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	resourceapi "k8s.io/api/resource/v1"
	v1 "k8s.io/api/resource/v1"
	"k8s.io/autoscaler/cluster-autoscaler/metrics"
)

func TestLogSampledDiscrepancies(t *testing.T) {
	tests := map[string]struct {
		sampledNodes  []string
		sampledDeltas [][]resourceDelta
		want          string
	}{
		"Empty": {
			sampledNodes:  []string{},
			sampledDeltas: [][]resourceDelta{},
			want:          "",
		},
		"SingleNodeSingleDelta": {
			sampledNodes: []string{"node-1"},
			sampledDeltas: [][]resourceDelta{
				{
					{
						Driver:               "test-driver",
						TemplateResourcePool: "pool-1",
						DeviceCountDelta:     1,
						TemplateSignatureMap: signatureMap{v1.QualifiedName("attr1"): {}},
					},
				},
			},
			want: "DRA Resource Discrepancies detected between node templates and actual nodes:\n- node-1: MissingResource{Driver=\"test-driver\", ResourcePool=\"pool-1\", DeviceCount=\"1\", AttributeSignature=\"attr1\"}",
		},
		"MultipleNodesMultipleDeltas": {
			sampledNodes: []string{"node-1", "node-2"},
			sampledDeltas: [][]resourceDelta{
				{
					{
						Driver:           "test-driver",
						NodeResourcePool: "pool-1",
						DeviceCountDelta: -1,
						NodeSignatureMap: signatureMap{v1.QualifiedName("attr2"): {}},
					},
				},
				{
					{
						Driver:               "test-driver",
						TemplateResourcePool: "pool-1",
						NodeResourcePool:     "pool-2",
						DeviceCountDelta:     0,
						TemplateSignatureMap: signatureMap{v1.QualifiedName("attr1"): {}},
						NodeSignatureMap:     signatureMap{v1.QualifiedName("attr2"): {}},
					},
				},
			},
			want: "DRA Resource Discrepancies detected between node templates and actual nodes:\n- node-1: ExtraResource{Driver=\"test-driver\", ResourcePool=\"pool-1\", DeviceCount=\"1\", AttributeSignature=\"attr2\"}\n- node-2: MismatchResource{Driver=\"test-driver\", TemplatePool=\"pool-1\", NodePool=\"pool-2\", DeviceCountDelta=\"0\", MissingAttributes=\"attr1\", ExtraAttributes=\"attr2\"}",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			var got string
			c := &NodeResourcesComparator{
				logger: func(format string, args ...any) {
					got = fmt.Sprintf(format, args...)
				},
			}

			c.logSampledDiscrepancies(tc.sampledNodes, tc.sampledDeltas)

			if got != tc.want {
				t.Errorf("logSampledDiscrepancies() = %q, want %q", got, tc.want)
			}
		})
	}
}


type fakeMetricsEmitter struct {
	metrics map[string]uint32
}

func (f *fakeMetricsEmitter) SetNodeTemplateResourcesMismatch(driver string, mismatchType metrics.ResourceMismatchType, value uint32) {
	if f.metrics == nil {
		f.metrics = make(map[string]uint32)
	}
	key := fmt.Sprintf("%s/%s", driver, mismatchType)
	f.metrics[key] = value
}

func TestEmitMetrics(t *testing.T) {
	tests := map[string]struct {
		aggregatedDiscrepancies map[string]driverDiscrepancy
		wantMetrics             map[string]uint32
	}{
		"Empty": {
			aggregatedDiscrepancies: map[string]driverDiscrepancy{},
			wantMetrics:             nil,
		},
		"MissingDiscrepancy": {
			aggregatedDiscrepancies: map[string]driverDiscrepancy{
				"driver-1": {missing: 1},
			},
			wantMetrics: map[string]uint32{
				"driver-1/missing": 1,
			},
		},
		"ExtraDiscrepancy": {
			aggregatedDiscrepancies: map[string]driverDiscrepancy{
				"driver-1": {extra: 2},
			},
			wantMetrics: map[string]uint32{
				"driver-1/extra": 2,
			},
		},
		"MismatchDiscrepancy": {
			aggregatedDiscrepancies: map[string]driverDiscrepancy{
				"driver-1": {mismatch: 3},
			},
			wantMetrics: map[string]uint32{
				"driver-1/mismatch": 3,
			},
		},
		"UnknownDiscrepancy": {
			aggregatedDiscrepancies: map[string]driverDiscrepancy{
				"driver-1": {unknown: 4},
			},
			wantMetrics: map[string]uint32{
				"driver-1/unknown": 4,
			},
		},
		"AllMetricsCombined": {
			aggregatedDiscrepancies: map[string]driverDiscrepancy{
				"driver-1": {missing: 1, extra: 2, mismatch: 3, unknown: 4},
			},
			wantMetrics: map[string]uint32{
				"driver-1/missing":  1,
				"driver-1/extra":    2,
				"driver-1/mismatch": 3,
				"driver-1/unknown":  4,
			},
		},
		"MultipleDriversMixedMetrics": {
			aggregatedDiscrepancies: map[string]driverDiscrepancy{
				"driver-1": {missing: 1},
				"driver-2": {extra: 5, mismatch: 2},
			},
			wantMetrics: map[string]uint32{
				"driver-1/missing":  1,
				"driver-2/extra":    5,
				"driver-2/mismatch": 2,
			},
		},
		"ZeroValueMetrics": {
			aggregatedDiscrepancies: map[string]driverDiscrepancy{
				"driver-1": {missing: 0, extra: 0},
			},
			wantMetrics: nil,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			fake := &fakeMetricsEmitter{}
			c := &NodeResourcesComparator{metrics: fake}
			c.emitMetrics(tc.aggregatedDiscrepancies)

			if diff := cmp.Diff(tc.wantMetrics, fake.metrics); diff != "" {
				t.Errorf("emitMetrics() diff (-want +got):\n%s", diff)
			}
		})
	}
}

type noOpMetricsEmitter struct{}

func (n *noOpMetricsEmitter) SetNodeTemplateResourcesMismatch(driver string, mismatchType metrics.ResourceMismatchType, value uint32) {}


func BenchmarkReportResourceDiscrepancies_1Node_0Discrepancies(b *testing.B) {
	comparator := NewNodeResourcesComparator(&noOpMetricsEmitter{})
	comparator.logger = func(format string, args ...any) {}

	data := []NodeComparisonData{
		{
			NodeName:       "node-1",
			TemplateSlices: []*resourceapi.ResourceSlice{makeSingleResourceSlice("driver", "pool", poolDevices{deviceCount: 1, shape: deviceShapeA})},
			NodeSlices:     []*resourceapi.ResourceSlice{makeSingleResourceSlice("driver", "pool", poolDevices{deviceCount: 1, shape: deviceShapeA})},
		},
	}

	for b.Loop() {
		comparator.ReportResourceDiscrepancies(data)
	}
}

func BenchmarkReportResourceDiscrepancies_1Node_NoDRA(b *testing.B) {
	comparator := NewNodeResourcesComparator(&noOpMetricsEmitter{})
	comparator.logger = func(format string, args ...any) {}

	data := []NodeComparisonData{
		{
			NodeName:       "node-1",
			TemplateSlices: nil,
			NodeSlices:     nil,
		},
	}

	for b.Loop() {
		comparator.ReportResourceDiscrepancies(data)
	}
}

func BenchmarkReportResourceDiscrepancies_1Node_10Discrepancies(b *testing.B) {
	comparator := NewNodeResourcesComparator(&noOpMetricsEmitter{})
	comparator.logger = func(format string, args ...any) {}

	templateSlices := make([]*resourceapi.ResourceSlice, 10)
	for i := 0; i < 10; i++ {
		templateSlices[i] = makeSingleResourceSlice("driver", fmt.Sprintf("pool-%d", i), poolDevices{deviceCount: 1, shape: deviceShapeA})
	}
	data := []NodeComparisonData{
		{
			NodeName:       "node-1",
			TemplateSlices: templateSlices,
			NodeSlices:     nil,
		},
	}

	for b.Loop() {
		comparator.ReportResourceDiscrepancies(data)
	}
}

func BenchmarkReportResourceDiscrepancies_10Nodes_10DiscrepanciesEach(b *testing.B) {
	comparator := NewNodeResourcesComparator(&noOpMetricsEmitter{})
	comparator.logger = func(format string, args ...any) {}

	data := make([]NodeComparisonData, 10)
	for n := 0; n < 10; n++ {
		templateSlices := make([]*resourceapi.ResourceSlice, 10)
		for i := 0; i < 10; i++ {
			templateSlices[i] = makeSingleResourceSlice("driver", fmt.Sprintf("pool-%d", i), poolDevices{deviceCount: 1, shape: deviceShapeA})
		}
		data[n] = NodeComparisonData{
			NodeName:       fmt.Sprintf("node-%d", n),
			TemplateSlices: templateSlices,
			NodeSlices:     nil,
		}
	}

	for b.Loop() {
		comparator.ReportResourceDiscrepancies(data)
	}
}

func BenchmarkReportResourceDiscrepancies_10Nodes_NoDRA(b *testing.B) {
	comparator := NewNodeResourcesComparator(&noOpMetricsEmitter{})
	comparator.logger = func(format string, args ...any) {}

	data := make([]NodeComparisonData, 10)
	for n := 0; n < 10; n++ {
		data[n] = NodeComparisonData{
			NodeName:       fmt.Sprintf("node-%d", n),
			TemplateSlices: nil,
			NodeSlices:     nil,
		}
	}

	for b.Loop() {
		comparator.ReportResourceDiscrepancies(data)
	}
}

func BenchmarkReportResourceDiscrepancies_10Nodes_0Discrepancies(b *testing.B) {
	comparator := NewNodeResourcesComparator(&noOpMetricsEmitter{})
	comparator.logger = func(format string, args ...any) {}

	data := make([]NodeComparisonData, 10)
	for n := 0; n < 10; n++ {
		slice := makeSingleResourceSlice("driver", "pool", poolDevices{deviceCount: 1, shape: deviceShapeA})
		data[n] = NodeComparisonData{
			NodeName:       fmt.Sprintf("node-%d", n),
			TemplateSlices: []*resourceapi.ResourceSlice{slice},
			NodeSlices:     []*resourceapi.ResourceSlice{slice},
		}
	}

	for b.Loop() {
		comparator.ReportResourceDiscrepancies(data)
	}
}
func BenchmarkReportResourceDiscrepancies_1Node_10Drivers_10Discrepancies(b *testing.B) {
	comparator := NewNodeResourcesComparator(&noOpMetricsEmitter{})
	comparator.logger = func(format string, args ...any) {}

	templateSlices := make([]*resourceapi.ResourceSlice, 10)
	for i := 0; i < 10; i++ {
		templateSlices[i] = makeSingleResourceSlice(fmt.Sprintf("driver-%d", i), "pool", poolDevices{deviceCount: 1, shape: deviceShapeA})
	}
	data := []NodeComparisonData{
		{
			NodeName:       "node-1",
			TemplateSlices: templateSlices,
			NodeSlices:     nil,
		},
	}

	for b.Loop() {
		comparator.ReportResourceDiscrepancies(data)
	}
}

func BenchmarkReportResourceDiscrepancies_10Nodes_10Drivers_10DiscrepanciesEach(b *testing.B) {
	comparator := NewNodeResourcesComparator(&noOpMetricsEmitter{})
	comparator.logger = func(format string, args ...any) {}

	data := make([]NodeComparisonData, 10)
	for n := 0; n < 10; n++ {
		templateSlices := make([]*resourceapi.ResourceSlice, 10)
		for i := 0; i < 10; i++ {
			templateSlices[i] = makeSingleResourceSlice(fmt.Sprintf("driver-%d", i), "pool", poolDevices{deviceCount: 1, shape: deviceShapeA})
		}
		data[n] = NodeComparisonData{
			NodeName:       fmt.Sprintf("node-%d", n),
			TemplateSlices: templateSlices,
			NodeSlices:     nil,
		}
	}

	for b.Loop() {
		comparator.ReportResourceDiscrepancies(data)
	}
}
