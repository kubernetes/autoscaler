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
	v1 "k8s.io/api/resource/v1"
)

func TestSampleSize(t *testing.T) {
	tests := map[string]struct {
		nodesToSample  []string
		wantSampleSize int
	}{
		"FewerThanLimit": {
			nodesToSample:  []string{"node-1", "node-2"},
			wantSampleSize: 2,
		},
		"ExactlyLimit": {
			nodesToSample:  []string{"node-1", "node-2", "node-3", "node-4", "node-5"},
			wantSampleSize: 5,
		},
		"MoreThanLimit": {
			nodesToSample:  []string{"node-1", "node-2", "node-3", "node-4", "node-5", "node-6"},
			wantSampleSize: 5,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			s := newLoggingSampler()
			for _, node := range tc.nodesToSample {
				s.Sample(node, nil)
			}
			if len(s.sampledNodes) != tc.wantSampleSize {
				t.Errorf("Sample() got size %d, want %d", len(s.sampledNodes), tc.wantSampleSize)
			}
		})
	}
}

func TestLogSampled(t *testing.T) {
	tests := map[string]struct {
		samples []struct {
			nodeName string
			deltas   []resourceDelta
		}
		wantOutput string
	}{
		"Empty": {
			samples:    nil,
			wantOutput: "",
		},
		"SingleNodeSingleDelta": {
			samples: []struct {
				nodeName string
				deltas   []resourceDelta
			}{
				{
					nodeName: "node-1",
					deltas: []resourceDelta{
						{
							Driver:               "driver-1",
							TemplateResourcePool: "pool-1",
							NodeResourcePool:     "pool-1",
							DeviceCountDelta:     1,
							TemplateSignatureMap: attributesMap{v1.QualifiedName("attr-1"): {}},
							NodeSignatureMap:     attributesMap{v1.QualifiedName("attr-1"): {}},
						},
					},
				},
			},
			wantOutput: `DRA Resource Discrepancies detected between node templates and actual nodes:
- node-1: ResourceDelta{Type="Mismatch", Driver="driver-1", TemplatePool="pool-1", NodePool="pool-1", DeviceCountDelta="1", TemplateSignature="attr-1", NodeSignature="attr-1", MissingAttributes="", ExtraAttributes=""}`,
		},
		"MultipleNodesAndDeltas": {
			samples: []struct {
				nodeName string
				deltas   []resourceDelta
			}{
				{
					nodeName: "node-1",
					deltas: []resourceDelta{
						{
							Driver:               "driver-1",
							TemplateResourcePool: "pool-1",
							NodeResourcePool:     "pool-1",
							DeviceCountDelta:     2,
							TemplateSignatureMap: attributesMap{v1.QualifiedName("A"): {}},
							NodeSignatureMap:     attributesMap{v1.QualifiedName("B"): {}},
						},
					},
				},
				{
					nodeName: "node-2",
					deltas: []resourceDelta{
						{
							Driver:               "driver-2",
							TemplateResourcePool: "pool-2",
							NodeResourcePool:     "pool-2",
							DeviceCountDelta:     0,
							TemplateSignatureMap: attributesMap{v1.QualifiedName("X"): {}, v1.QualifiedName("Y"): {}},
							NodeSignatureMap:     attributesMap{v1.QualifiedName("X"): {}, v1.QualifiedName("Z"): {}},
						},
						{
							Driver:               "driver-3",
							TemplateResourcePool: "pool-3",
							NodeResourcePool:     "pool-3",
							DeviceCountDelta:     -1,
						},
					},
				},
			},
			wantOutput: `DRA Resource Discrepancies detected between node templates and actual nodes:
- node-1: ResourceDelta{Type="Mismatch", Driver="driver-1", TemplatePool="pool-1", NodePool="pool-1", DeviceCountDelta="2", TemplateSignature="A", NodeSignature="B", MissingAttributes="A", ExtraAttributes="B"}
- node-2: ResourceDelta{Type="Mismatch", Driver="driver-2", TemplatePool="pool-2", NodePool="pool-2", DeviceCountDelta="0", TemplateSignature="X;Y", NodeSignature="X;Z", MissingAttributes="Y", ExtraAttributes="Z"}, ResourceDelta{Type="Mismatch", Driver="driver-3", TemplatePool="pool-3", NodePool="pool-3", DeviceCountDelta="-1", TemplateSignature="", NodeSignature="", MissingAttributes="", ExtraAttributes=""}`,
		},
		"MissingAndExtraPools": {
			samples: []struct {
				nodeName string
				deltas   []resourceDelta
			}{
				{
					nodeName: "node-1",
					deltas: []resourceDelta{
						{
							Driver:               "driver-1",
							TemplateResourcePool: "pool-1",
							DeviceCountDelta:     1,
						},
						{
							Driver:           "driver-1",
							NodeResourcePool: "pool-2",
							DeviceCountDelta: -1,
						},
					},
				},
			},
			wantOutput: `DRA Resource Discrepancies detected between node templates and actual nodes:
- node-1: ResourceDelta{Type="Missing", Driver="driver-1", TemplatePool="pool-1", NodePool="", DeviceCountDelta="1", TemplateSignature="", NodeSignature="", MissingAttributes="", ExtraAttributes=""}, ResourceDelta{Type="Extra", Driver="driver-1", TemplatePool="", NodePool="pool-2", DeviceCountDelta="-1", TemplateSignature="", NodeSignature="", MissingAttributes="", ExtraAttributes=""}`,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			var gotOutput string
			logger := func(format string, args ...any) {
				gotOutput = fmt.Sprintf(format, args...)
			}
			s := newLoggingSamplerWithLogger(logger)
			for _, sample := range tc.samples {
				s.Sample(sample.nodeName, sample.deltas)
			}
			s.LogSampled()

			if diff := cmp.Diff(tc.wantOutput, gotOutput); diff != "" {
				t.Errorf("LogSampled() diff (-want +got):\n%s", diff)
			}
		})
	}
}

func TestWriteAttributes(t *testing.T) {
	tests := map[string]struct {
		m          attributesMap
		wantOutput string
	}{
		"Empty": {
			m:          attributesMap{},
			wantOutput: "",
		},
		"SingleKey": {
			m:          attributesMap{v1.QualifiedName("A"): {}},
			wantOutput: "A",
		},
		"MultipleKeysSorted": {
			m:          attributesMap{v1.QualifiedName("C"): {}, v1.QualifiedName("A"): {}, v1.QualifiedName("B"): {}},
			wantOutput: "A;B;C",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			sampler := newLoggingSampler()
			sampler.writeAttributes(tc.m)
			if got := sampler.summaryBuilder.String(); got != tc.wantOutput {
				t.Errorf("writeAttributes() = %q, want %q", got, tc.wantOutput)
			}
		})
	}
}

func TestWriteAttributesDifference(t *testing.T) {
	tests := map[string]struct {
		base       attributesMap
		subtract   attributesMap
		wantOutput string
	}{
		"Empty": {
			base:       attributesMap{},
			subtract:   attributesMap{},
			wantOutput: "",
		},
		"NoDifference": {
			base:       attributesMap{v1.QualifiedName("A"): {}},
			subtract:   attributesMap{v1.QualifiedName("A"): {}},
			wantOutput: "",
		},
		"PartialDifference": {
			base:       attributesMap{v1.QualifiedName("A"): {}, v1.QualifiedName("B"): {}},
			subtract:   attributesMap{v1.QualifiedName("B"): {}},
			wantOutput: "A",
		},
		"NoDifference_NonExistentKey": {
			base:       attributesMap{v1.QualifiedName("A"): {}, v1.QualifiedName("B"): {}},
			subtract:   attributesMap{v1.QualifiedName("C"): {}},
			wantOutput: "A;B",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			sampler := newLoggingSampler()
			sampler.writeAttributesDifference(tc.base, tc.subtract)
			if got := sampler.summaryBuilder.String(); got != tc.wantOutput {
				t.Errorf("writeMapKeyDifference() = %q, want %q", got, tc.wantOutput)
			}
		})
	}
}
