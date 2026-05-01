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
	"k8s.io/autoscaler/cluster-autoscaler/metrics"
)

var noOpLogger logger = func(format string, args ...any) {}

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

func (f *fakeMetricsEmitter) Reset() {
	clear(f.metrics)
}

// nodeComparisonData is used to simulate node resource topology for testing,
// has no difference from passing raw slices, but drastically improves readability
// by grouping related data together.
type nodeComparisonData struct {
	NodeName       string
	TemplateSlices []*resourceapi.ResourceSlice
	NodeSlices     []*resourceapi.ResourceSlice
}

func nodeNamesArray(data []nodeComparisonData) []string {
	result := make([]string, len(data))
	for i, d := range data {
		result[i] = d.NodeName
	}
	return result
}

func nodeSlicesArray(data []nodeComparisonData) [][]*resourceapi.ResourceSlice {
	result := make([][]*resourceapi.ResourceSlice, len(data))
	for i, d := range data {
		result[i] = d.NodeSlices
	}
	return result
}

func templateSlicesArray(data []nodeComparisonData) [][]*resourceapi.ResourceSlice {
	result := make([][]*resourceapi.ResourceSlice, len(data))
	for i, d := range data {
		result[i] = d.TemplateSlices
	}
	return result
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
				"driver-1/missing":  1,
				"driver-1/extra":    0,
				"driver-1/mismatch": 0,
				"driver-1/unknown":  0,
			},
		},
		"ExtraDiscrepancy": {
			aggregatedDiscrepancies: map[string]driverDiscrepancy{
				"driver-1": {extra: 2},
			},
			wantMetrics: map[string]uint32{
				"driver-1/extra":    2,
				"driver-1/missing":  0,
				"driver-1/mismatch": 0,
				"driver-1/unknown":  0,
			},
		},
		"MismatchDiscrepancy": {
			aggregatedDiscrepancies: map[string]driverDiscrepancy{
				"driver-1": {mismatch: 3},
			},
			wantMetrics: map[string]uint32{
				"driver-1/mismatch": 3,
				"driver-1/missing":  0,
				"driver-1/extra":    0,
				"driver-1/unknown":  0,
			},
		},
		"UnknownDiscrepancy": {
			aggregatedDiscrepancies: map[string]driverDiscrepancy{
				"driver-1": {unknown: 4},
			},
			wantMetrics: map[string]uint32{
				"driver-1/unknown":  4,
				"driver-1/missing":  0,
				"driver-1/extra":    0,
				"driver-1/mismatch": 0,
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
				"driver-1/unknown":  0,
				"driver-1/extra":    0,
				"driver-1/mismatch": 0,

				"driver-2/missing":  0,
				"driver-2/unknown":  0,
				"driver-2/extra":    5,
				"driver-2/mismatch": 2,
			},
		},
		"ZeroValueMetrics": {
			aggregatedDiscrepancies: map[string]driverDiscrepancy{
				"driver-1": {missing: 0, extra: 0},
			},
			wantMetrics: map[string]uint32{
				"driver-1/missing":  0,
				"driver-1/extra":    0,
				"driver-1/mismatch": 0,
				"driver-1/unknown":  0,
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			fake := &fakeMetricsEmitter{}
			c := &NodeResourcesComparator{metrics: fake, discrepanciesPerDriver: tc.aggregatedDiscrepancies}
			c.emitMetrics()

			if diff := cmp.Diff(tc.wantMetrics, fake.metrics); diff != "" {
				t.Errorf("emitMetrics() diff (-want +got):\n%s", diff)
			}
		})
	}
}

type noOpMetricsEmitter struct{}

func (n *noOpMetricsEmitter) SetNodeTemplateResourcesMismatch(driver string, mismatchType metrics.ResourceMismatchType, value uint32) {
}

func TestReportResourceDiscrepancies(t *testing.T) {
	tests := map[string]struct {
		data        []nodeComparisonData
		wantMetrics map[string]uint32
	}{
		"EmptyData": {
			data:        nil,
			wantMetrics: nil,
		},
		"NoSlicesToCompare": {
			data: []nodeComparisonData{
				{
					NodeName:       "node-1",
					TemplateSlices: nil,
					NodeSlices:     nil,
				},
			},
			wantMetrics: nil,
		},
		"PerfectMatch": {
			data: []nodeComparisonData{
				{
					NodeName:       "node-1",
					TemplateSlices: []*resourceapi.ResourceSlice{makeSingleResourceSlice("driver", "pool", poolDevices{deviceCount: 2, shape: deviceShapeA})},
					NodeSlices:     []*resourceapi.ResourceSlice{makeSingleResourceSlice("driver", "pool", poolDevices{deviceCount: 2, shape: deviceShapeA})},
				},
			},
			wantMetrics: nil,
		},
		"MissingResourcePool": {
			data: []nodeComparisonData{
				{
					NodeName:       "node-1",
					TemplateSlices: []*resourceapi.ResourceSlice{makeSingleResourceSlice("driver", "pool", poolDevices{deviceCount: 1, shape: deviceShapeA})},
					NodeSlices:     nil,
				},
			},
			wantMetrics: map[string]uint32{
				"driver/missing":  1,
				"driver/extra":    0,
				"driver/mismatch": 0,
				"driver/unknown":  0,
			},
		},
		"ExtraResourcePool": {
			data: []nodeComparisonData{
				{
					NodeName:       "node-1",
					TemplateSlices: []*resourceapi.ResourceSlice{makeSingleResourceSlice("driver", "pool-1", poolDevices{deviceCount: 1, shape: deviceShapeA})},
					NodeSlices: []*resourceapi.ResourceSlice{
						makeSingleResourceSlice("driver", "pool-1", poolDevices{deviceCount: 1, shape: deviceShapeA}),
						makeSingleResourceSlice("driver", "pool-2", poolDevices{deviceCount: 1, shape: deviceShapeB}),
					},
				},
			},
			wantMetrics: map[string]uint32{
				"driver/extra":    1,
				"driver/missing":  0,
				"driver/mismatch": 0,
				"driver/unknown":  0,
			},
		},
		"FuzzyMismatch_Attributes": {
			data: []nodeComparisonData{
				{
					NodeName:       "node-1",
					TemplateSlices: []*resourceapi.ResourceSlice{makeSingleResourceSlice("driver", "pool", poolDevices{deviceCount: 1, shape: deviceShapeA})},
					NodeSlices:     []*resourceapi.ResourceSlice{makeSingleResourceSlice("driver", "pool", poolDevices{deviceCount: 1, shape: deviceShapeB})},
				},
			},
			wantMetrics: map[string]uint32{
				"driver/mismatch": 1,
				"driver/missing":  0,
				"driver/extra":    0,
				"driver/unknown":  0,
			},
		},
		"FuzzyMismatch_DeviceCount": {
			data: []nodeComparisonData{
				{
					NodeName:       "node-1",
					TemplateSlices: []*resourceapi.ResourceSlice{makeSingleResourceSlice("driver", "pool", poolDevices{deviceCount: 4, shape: deviceShapeA})},
					NodeSlices:     []*resourceapi.ResourceSlice{makeSingleResourceSlice("driver", "pool", poolDevices{deviceCount: 2, shape: deviceShapeA})},
				},
			},
			wantMetrics: map[string]uint32{
				"driver/mismatch": 1,
				"driver/missing":  0,
				"driver/extra":    0,
				"driver/unknown":  0,
			},
		},
		"IgnoredDriver": {
			data: []nodeComparisonData{
				{
					NodeName: "node-1",
					TemplateSlices: []*resourceapi.ResourceSlice{
						makeSingleResourceSlice("known-driver", "pool", poolDevices{deviceCount: 1, shape: deviceShapeA}),
					},
					NodeSlices: []*resourceapi.ResourceSlice{
						makeSingleResourceSlice("known-driver", "pool", poolDevices{deviceCount: 1, shape: deviceShapeA}),
						makeSingleResourceSlice("node-only-driver", "pool", poolDevices{deviceCount: 5, shape: deviceShapeB}),
					},
				},
			},
			wantMetrics: map[string]uint32{
				"node-only-driver/missing":  0,
				"node-only-driver/extra":    1,
				"node-only-driver/mismatch": 0,
				"node-only-driver/unknown":  0,
			},
		},
		"MultiNodeMultiDriver": {
			data: []nodeComparisonData{
				{
					// Node 1: driver-A is missing a pool
					NodeName: "node-1",
					TemplateSlices: []*resourceapi.ResourceSlice{
						makeSingleResourceSlice("driver-A", "pool-1", poolDevices{deviceCount: 1, shape: deviceShapeA}),
						makeSingleResourceSlice("driver-B", "pool-1", poolDevices{deviceCount: 1, shape: deviceShapeB}),
					},
					NodeSlices: []*resourceapi.ResourceSlice{
						makeSingleResourceSlice("driver-B", "pool-1", poolDevices{deviceCount: 1, shape: deviceShapeB}),
					},
				},
				{
					// Node 2: driver-A has a fuzzy mismatch, driver-B has an extra pool
					NodeName: "node-2",
					TemplateSlices: []*resourceapi.ResourceSlice{
						makeSingleResourceSlice("driver-A", "pool-1", poolDevices{deviceCount: 2, shape: deviceShapeA}),
						makeSingleResourceSlice("driver-B", "pool-1", poolDevices{deviceCount: 1, shape: deviceShapeB}),
					},
					NodeSlices: []*resourceapi.ResourceSlice{
						makeSingleResourceSlice("driver-A", "pool-1", poolDevices{deviceCount: 1, shape: deviceShapeA}),
						makeSingleResourceSlice("driver-B", "pool-1", poolDevices{deviceCount: 1, shape: deviceShapeB}),
						makeSingleResourceSlice("driver-B", "pool-2", poolDevices{deviceCount: 1, shape: deviceShapeA}),
					},
				},
			},
			wantMetrics: map[string]uint32{
				"driver-A/missing":  1,
				"driver-A/mismatch": 1,
				"driver-A/extra":    0,
				"driver-A/unknown":  0,

				"driver-B/missing":  0,
				"driver-B/extra":    1,
				"driver-B/mismatch": 0,
				"driver-B/unknown":  0,
			},
		},
		"MultipleMissing": {
			data: []nodeComparisonData{
				{
					NodeName: "node-1",
					TemplateSlices: []*resourceapi.ResourceSlice{
						makeSingleResourceSlice("driver", "pool1", poolDevices{deviceCount: 1, shape: deviceShapeA}),
						makeSingleResourceSlice("driver", "pool2", poolDevices{deviceCount: 1, shape: deviceShapeB}),
						makeSingleResourceSlice("driver", "pool3", poolDevices{deviceCount: 1, shape: deviceShapeABC}),
					},
					NodeSlices: []*resourceapi.ResourceSlice{},
				},
			},
			wantMetrics: map[string]uint32{
				"driver/missing":  3,
				"driver/extra":    0,
				"driver/mismatch": 0,
				"driver/unknown":  0,
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			fake := &fakeMetricsEmitter{}
			c := NewNodeResourcesComparator(fake)
			c.ReportResourceDiscrepancies(nodeNamesArray(tc.data), templateSlicesArray(tc.data), nodeSlicesArray(tc.data))
			if diff := cmp.Diff(tc.wantMetrics, fake.metrics); diff != "" {
				t.Errorf("ReportResourceDiscrepancies() metrics diff (-want +got):\n%s", diff)
			}
		})
	}
}

func TestMissingNodesGaugesAreReset(t *testing.T) {
	nodeNames := []string{"node-1"}

	templateSlices := [][]*resourceapi.ResourceSlice{
		{
			makeSingleResourceSlice("driver1", "pool", poolDevices{deviceCount: 1, shape: deviceShapeA}),
			makeSingleResourceSlice("driver2", "pool", poolDevices{deviceCount: 4, shape: deviceShapeA}),
			makeSingleResourceSlice("driver3", "pool", poolDevices{deviceCount: 1, shape: deviceShapeABC}),
		},
	}

	nodeSlices := [][]*resourceapi.ResourceSlice{
		{
			makeSingleResourceSlice("driver1", "pool1", poolDevices{deviceCount: 1, shape: deviceShapeB}),
			makeSingleResourceSlice("driver2", "pool", poolDevices{deviceCount: 2, shape: deviceShapeA}),
			makeSingleResourceSlice("driver1", "pool2", poolDevices{deviceCount: 1, shape: deviceShapeABC}),
		},
	}

	wantMetricsBeforeScaleDown := map[string]uint32{
		"driver1/mismatch": 1,
		"driver1/unknown":  0,
		"driver1/extra":    1,
		"driver1/missing":  0,

		"driver2/mismatch": 1,
		"driver2/unknown":  0,
		"driver2/extra":    0,
		"driver2/missing":  0,

		"driver3/mismatch": 0,
		"driver3/unknown":  0,
		"driver3/extra":    0,
		"driver3/missing":  1,
	}

	wantMetricsAfterScaleDown := map[string]uint32{
		"driver1/mismatch": 0,
		"driver1/unknown":  0,
		"driver1/extra":    0,
		"driver1/missing":  0,

		"driver2/mismatch": 0,
		"driver2/unknown":  0,
		"driver2/extra":    0,
		"driver2/missing":  0,

		"driver3/mismatch": 0,
		"driver3/unknown":  0,
		"driver3/extra":    0,
		"driver3/missing":  0,
	}

	fake := &fakeMetricsEmitter{}
	c := NewNodeResourcesComparator(fake)
	c.ReportResourceDiscrepancies(nodeNames, templateSlices, nodeSlices)
	if diff := cmp.Diff(wantMetricsBeforeScaleDown, fake.metrics); diff != "" {
		t.Errorf("ReportResourceDiscrepancies() metrics diff (-want +got):\n%s", diff)
	}

	fake.Reset()
	c.ReportResourceDiscrepancies([]string{}, [][]*resourceapi.ResourceSlice{}, [][]*resourceapi.ResourceSlice{})
	if diff := cmp.Diff(wantMetricsAfterScaleDown, fake.metrics); diff != "" {
		t.Errorf("ReportResourceDiscrepancies() metrics diff (-want +got):\n%s", diff)
	}

	fake.Reset()
	c.ReportResourceDiscrepancies([]string{}, [][]*resourceapi.ResourceSlice{}, [][]*resourceapi.ResourceSlice{})
	if len(fake.metrics) != 0 {
		t.Errorf("Expected no metrics after second reset, while reported: %v", fake.metrics)
	}
}

func BenchmarkReportResourceDiscrepancies_1Node_0Discrepancies(b *testing.B) {
	comparator := NewNodeResourcesComparator(&noOpMetricsEmitter{})
	nodeNames := []string{"node-1"}
	templateSlices := [][]*resourceapi.ResourceSlice{
		{makeSingleResourceSlice("driver", "pool", poolDevices{deviceCount: 1, shape: deviceShapeA})},
	}
	nodeSlices := [][]*resourceapi.ResourceSlice{
		{makeSingleResourceSlice("driver", "pool", poolDevices{deviceCount: 1, shape: deviceShapeA})},
	}

	for b.Loop() {
		comparator.ReportResourceDiscrepancies(nodeNames, templateSlices, nodeSlices)
	}
}

func BenchmarkReportResourceDiscrepancies_1Node_NoDRA(b *testing.B) {
	comparator := newNodeResourcesComparatorWithLogger(&noOpMetricsEmitter{}, noOpLogger)
	nodeNames := []string{"node-1"}
	templateSlices := [][]*resourceapi.ResourceSlice{nil}
	nodeSlices := [][]*resourceapi.ResourceSlice{nil}

	for b.Loop() {
		comparator.ReportResourceDiscrepancies(nodeNames, templateSlices, nodeSlices)
	}
}

func BenchmarkReportResourceDiscrepancies_1Node_10Discrepancies(b *testing.B) {
	comparator := newNodeResourcesComparatorWithLogger(&noOpMetricsEmitter{}, noOpLogger)

	templateSlice := make([]*resourceapi.ResourceSlice, 10)
	for i := 0; i < 10; i++ {
		templateSlice[i] = makeSingleResourceSlice("driver", fmt.Sprintf("pool-%d", i), poolDevices{deviceCount: 1, shape: deviceShapeA})
	}

	nodeNames := []string{"node-1"}
	nodeSlices := [][]*resourceapi.ResourceSlice{{}}
	templateSlices := [][]*resourceapi.ResourceSlice{templateSlice}
	for b.Loop() {
		comparator.ReportResourceDiscrepancies(nodeNames, templateSlices, nodeSlices)
	}
}

func BenchmarkReportResourceDiscrepancies_10Nodes_10DiscrepanciesEach(b *testing.B) {
	comparator := newNodeResourcesComparatorWithLogger(&noOpMetricsEmitter{}, noOpLogger)
	templateSlices := make([][]*resourceapi.ResourceSlice, 10)
	nodeSlices := make([][]*resourceapi.ResourceSlice, 10)
	nodeNames := make([]string, 10)

	for n := 0; n < 10; n++ {
		templateSlice := make([]*resourceapi.ResourceSlice, 10)
		for i := 0; i < 10; i++ {
			templateSlice[i] = makeSingleResourceSlice("driver", fmt.Sprintf("pool-%d", i), poolDevices{deviceCount: 1, shape: deviceShapeA})
		}

		nodeNames[n] = fmt.Sprintf("node-%d", n)
		templateSlices[n] = templateSlice
		nodeSlices[n] = nil
	}

	for b.Loop() {
		comparator.ReportResourceDiscrepancies(nodeNames, templateSlices, nodeSlices)
	}
}

func BenchmarkReportResourceDiscrepancies_10Nodes_NoDRA(b *testing.B) {
	comparator := newNodeResourcesComparatorWithLogger(&noOpMetricsEmitter{}, noOpLogger)
	nodeNames := make([]string, 10)
	templateSlices := make([][]*resourceapi.ResourceSlice, 10)
	nodeSlices := make([][]*resourceapi.ResourceSlice, 10)

	for n := 0; n < 10; n++ {
		nodeNames[n] = fmt.Sprintf("node-%d", n)
		templateSlices[n] = nil
		nodeSlices[n] = nil
	}

	for b.Loop() {
		comparator.ReportResourceDiscrepancies(nodeNames, templateSlices, nodeSlices)
	}
}

func BenchmarkReportResourceDiscrepancies_10Nodes_0Discrepancies(b *testing.B) {
	comparator := newNodeResourcesComparatorWithLogger(&noOpMetricsEmitter{}, noOpLogger)
	names := make([]string, 10)
	tpl := make([][]*resourceapi.ResourceSlice, 10)
	nodes := make([][]*resourceapi.ResourceSlice, 10)

	for n := 0; n < 10; n++ {
		slice := makeSingleResourceSlice("driver", "pool", poolDevices{deviceCount: 1, shape: deviceShapeA})
		names[n] = fmt.Sprintf("node-%d", n)
		tpl[n] = []*resourceapi.ResourceSlice{slice}
		nodes[n] = []*resourceapi.ResourceSlice{slice}
	}

	for b.Loop() {
		comparator.ReportResourceDiscrepancies(names, tpl, nodes)
	}
}
func BenchmarkReportResourceDiscrepancies_1Node_10Drivers_10Discrepancies(b *testing.B) {
	comparator := newNodeResourcesComparatorWithLogger(&noOpMetricsEmitter{}, noOpLogger)
	templateSlices := make([]*resourceapi.ResourceSlice, 10)
	for i := 0; i < 10; i++ {
		templateSlices[i] = makeSingleResourceSlice(fmt.Sprintf("driver-%d", i), "pool", poolDevices{deviceCount: 1, shape: deviceShapeA})
	}

	names := []string{"node-1"}
	tpl := [][]*resourceapi.ResourceSlice{templateSlices}
	nodes := [][]*resourceapi.ResourceSlice{nil}

	for b.Loop() {
		comparator.ReportResourceDiscrepancies(names, tpl, nodes)
	}
}

func BenchmarkReportResourceDiscrepancies_10Nodes_10Drivers_10DiscrepanciesEach(b *testing.B) {
	comparator := newNodeResourcesComparatorWithLogger(&noOpMetricsEmitter{}, noOpLogger)

	names := make([]string, 10)
	tpl := make([][]*resourceapi.ResourceSlice, 10)
	nodes := make([][]*resourceapi.ResourceSlice, 10)
	for i := 0; i < 10; i++ {
		templateSlices := make([]*resourceapi.ResourceSlice, 10)
		for j := 0; j < 10; j++ {
			templateSlices[j] = makeSingleResourceSlice(fmt.Sprintf("driver-%d", j), "pool", poolDevices{deviceCount: 1, shape: deviceShapeA})
		}
		names[i] = fmt.Sprintf("node-%d", i)
		tpl[i] = templateSlices
		nodes[i] = nil
	}

	for b.Loop() {
		comparator.ReportResourceDiscrepancies(names, tpl, nodes)
	}
}
