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
	gocmp "cmp"
	"fmt"
	"slices"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	resourceapi "k8s.io/api/resource/v1"
	v1 "k8s.io/api/resource/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	deviceShapeA     = map[string]struct{}{"A": {}}
	deviceShapeB     = map[string]struct{}{"B": {}}
	deviceShapeAB    = map[string]struct{}{"A": {}, "B": {}}
	deviceShapeEmpty = map[string]struct{}{}
	deviceShapeABC   = map[string]struct{}{"A": {}, "B": {}, "C": {}}
	deviceShapeABCD  = map[string]struct{}{"A": {}, "B": {}, "C": {}, "D": {}}
	deviceShapeA_BC  = map[string]struct{}{"A": {}, "BC": {}}
	deviceShapeAB_C  = map[string]struct{}{"AB": {}, "C": {}}
)

func TestCompareDraResources(t *testing.T) {
	tests := map[string]struct {
		templateSlices []*resourceapi.ResourceSlice
		nodeSlices     []*resourceapi.ResourceSlice
		wantReports    []resourceDelta
	}{
		"Empty": {
			templateSlices: []*resourceapi.ResourceSlice{},
			nodeSlices:     []*resourceapi.ResourceSlice{},
			wantReports:    []resourceDelta{},
		},
		"TemplateOnly": {
			templateSlices: []*resourceapi.ResourceSlice{makeSingleResourceSlice("driver", "pool", poolDevices{deviceCount: 1, shape: deviceShapeA})},
			nodeSlices:     []*resourceapi.ResourceSlice{},
			wantReports: []resourceDelta{
				{
					Driver:               "driver",
					TemplateResourcePool: "pool",
					TemplateSignatureMap: signatureMap{v1.QualifiedName("A"): {}},
					NodeSignatureMap:     signatureMap{},
					DeviceCountDelta:     1,
				},
			},
		},
		"NoDriverInTemplate": {
			templateSlices: []*resourceapi.ResourceSlice{},
			nodeSlices:     []*resourceapi.ResourceSlice{makeSingleResourceSlice("driver", "pool", poolDevices{deviceCount: 1, shape: deviceShapeA})},
			wantReports:    []resourceDelta{},
		},
		"ShapeMismatch": {
			templateSlices: []*resourceapi.ResourceSlice{makeSingleResourceSlice("driver", "pool", poolDevices{deviceCount: 1, shape: deviceShapeA})},
			nodeSlices:     []*resourceapi.ResourceSlice{makeSingleResourceSlice("driver", "pool", poolDevices{deviceCount: 1, shape: deviceShapeB})},
			wantReports: []resourceDelta{
				{
					Driver:               "driver",
					TemplateResourcePool: "pool",
					NodeResourcePool:     "pool",
					TemplateSignatureMap: signatureMap{v1.QualifiedName("A"): {}},
					NodeSignatureMap:     signatureMap{v1.QualifiedName("B"): {}},
					DeviceCountDelta:     0,
				},
			},
		},
		"ExactMatch": {
			templateSlices: []*resourceapi.ResourceSlice{makeSingleResourceSlice("driver", "pool", poolDevices{deviceCount: 2, shape: deviceShapeA})},
			nodeSlices:     []*resourceapi.ResourceSlice{makeSingleResourceSlice("driver", "pool", poolDevices{deviceCount: 2, shape: deviceShapeA})},
			wantReports:    []resourceDelta{},
		},
		"ExactMatchDifferentPoolName": {
			templateSlices: []*resourceapi.ResourceSlice{makeSingleResourceSlice("driver", "template-pool", poolDevices{deviceCount: 1, shape: deviceShapeA})},
			nodeSlices:     []*resourceapi.ResourceSlice{makeSingleResourceSlice("driver", "node-pool", poolDevices{deviceCount: 1, shape: deviceShapeA})},
			wantReports:    []resourceDelta{},
		},
		"MissingHardware": {
			templateSlices: []*resourceapi.ResourceSlice{makeSingleResourceSlice("driver", "pool", poolDevices{deviceCount: 3, shape: deviceShapeA})},
			nodeSlices:     []*resourceapi.ResourceSlice{makeSingleResourceSlice("driver", "pool", poolDevices{deviceCount: 2, shape: deviceShapeA})},
			wantReports: []resourceDelta{
				{
					Driver:               "driver",
					TemplateResourcePool: "pool",
					NodeResourcePool:     "pool",
					TemplateSignatureMap: signatureMap{v1.QualifiedName("A"): {}},
					NodeSignatureMap: signatureMap{v1.QualifiedName("A"): {}},
					DeviceCountDelta:     1,
				},
			},
		},
		"FuzzyMatchSubset": {
			templateSlices: []*resourceapi.ResourceSlice{makeSingleResourceSlice("driver", "pool", poolDevices{deviceCount: 1, shape: deviceShapeAB})},
			nodeSlices:     []*resourceapi.ResourceSlice{makeSingleResourceSlice("driver", "pool", poolDevices{deviceCount: 1, shape: deviceShapeA})},
			wantReports: []resourceDelta{
				{
					Driver:               "driver",
					TemplateResourcePool: "pool",
					NodeResourcePool:     "pool",
					TemplateSignatureMap: signatureMap{v1.QualifiedName("A"): {}, v1.QualifiedName("B"): {}},
					NodeSignatureMap:     signatureMap{v1.QualifiedName("A"): {}},
				},
			},
		},
		"FuzzyMatchSuperset": {
			templateSlices: []*resourceapi.ResourceSlice{makeSingleResourceSlice("driver", "pool", poolDevices{deviceCount: 1, shape: deviceShapeA})},
			nodeSlices:     []*resourceapi.ResourceSlice{makeSingleResourceSlice("driver", "pool", poolDevices{deviceCount: 1, shape: deviceShapeAB})},
			wantReports: []resourceDelta{
				{
					Driver:               "driver",
					TemplateResourcePool: "pool",
					NodeResourcePool:     "pool",
					TemplateSignatureMap: signatureMap{v1.QualifiedName("A"): {}},
					NodeSignatureMap:     signatureMap{v1.QualifiedName("A"): {}, v1.QualifiedName("B"): {}},
				},
			},
		},
		"HeterogeneousPoolExactMatch": {
			templateSlices: []*resourceapi.ResourceSlice{makeSingleResourceSlice("driver", "pool", poolDevices{1, deviceShapeA}, poolDevices{1, deviceShapeB})},
			nodeSlices:     []*resourceapi.ResourceSlice{makeSingleResourceSlice("driver", "pool", poolDevices{1, deviceShapeB}, poolDevices{1, deviceShapeA})},
			wantReports:    []resourceDelta{},
		},
		"ExtraNodeResourcePool": {
			templateSlices: []*resourceapi.ResourceSlice{makeSingleResourceSlice("driver", "pool", poolDevices{1, deviceShapeA})},
			nodeSlices: []*resourceapi.ResourceSlice{
				makeSingleResourceSlice("driver", "pool", poolDevices{1, deviceShapeA}),
				makeSingleResourceSlice("driver", "pool-extra", poolDevices{1, deviceShapeB}),
			},
			wantReports: []resourceDelta{
				{
					Driver:           "driver",
					NodeResourcePool: "pool-extra",
					NodeSignatureMap: signatureMap{v1.QualifiedName("B"): {}},
					DeviceCountDelta: -1,
				},
			},
		},
		"FragmentedNodeSlices": {
			templateSlices: []*resourceapi.ResourceSlice{
				makeSingleResourceSlice("driver", "pool-a", poolDevices{deviceCount: 4, shape: deviceShapeA}),
			},
			nodeSlices: []*resourceapi.ResourceSlice{
				makeResourceSlice("driver", "pool-a", 2, poolDevices{deviceCount: 2, shape: deviceShapeA}),
				makeResourceSlice("driver", "pool-a", 2, poolDevices{deviceCount: 2, shape: deviceShapeA}),
			},
			wantReports: []resourceDelta{},
		},
		"MultiplePoolsWithMixedDeltas": {
			templateSlices: []*resourceapi.ResourceSlice{
				makeSingleResourceSlice("driver", "pool-1", poolDevices{deviceCount: 2, shape: deviceShapeA}),
				makeSingleResourceSlice("driver", "pool-2", poolDevices{deviceCount: 1, shape: deviceShapeB}),
			},
			nodeSlices: []*resourceapi.ResourceSlice{
				// pool-1 is missing a device (count 1 instead of 2)
				makeSingleResourceSlice("driver", "pool-1", poolDevices{deviceCount: 1, shape: deviceShapeA}),
				// pool-2 has an extra attribute (shape AB instead of B)
				makeSingleResourceSlice("driver", "pool-2", poolDevices{deviceCount: 1, shape: deviceShapeAB}),
			},
			wantReports: []resourceDelta{
				{
					Driver:               "driver",
					TemplateResourcePool: "pool-1",
					NodeResourcePool:     "pool-1",
					TemplateSignatureMap: signatureMap{v1.QualifiedName("A"): {}},
					NodeSignatureMap:     signatureMap{v1.QualifiedName("A"): {}},
					DeviceCountDelta:     1,
				},
				{
					Driver:               "driver",
					TemplateResourcePool: "pool-2",
					NodeResourcePool:     "pool-2",
					TemplateSignatureMap: signatureMap{v1.QualifiedName("B"): {}},
					NodeSignatureMap:     signatureMap{v1.QualifiedName("A"): {}, v1.QualifiedName("B"): {}},
				},
			},
		},
		"CrossDriver": {
			templateSlices: []*resourceapi.ResourceSlice{
				makeSingleResourceSlice("driver-alpha", "pool-a", poolDevices{deviceCount: 1, shape: deviceShapeA}),
				makeSingleResourceSlice("driver-beta", "pool-b", poolDevices{deviceCount: 1, shape: deviceShapeB}),
			},
			nodeSlices: []*resourceapi.ResourceSlice{
				// alpha matches perfectly
				makeSingleResourceSlice("driver-alpha", "pool-a", poolDevices{deviceCount: 1, shape: deviceShapeA}),
				// beta has a mismatch
				makeSingleResourceSlice("driver-beta", "pool-b", poolDevices{deviceCount: 1, shape: deviceShapeAB}),
				// gamma should be completely ignored (not in template)
				makeSingleResourceSlice("driver-gamma", "pool-c", poolDevices{deviceCount: 1, shape: deviceShapeA}),
			},
			wantReports: []resourceDelta{
				{
					Driver:               "driver-beta",
					TemplateResourcePool: "pool-b",
					NodeResourcePool:     "pool-b",
					TemplateSignatureMap: signatureMap{v1.QualifiedName("B"): {}},
					NodeSignatureMap:     signatureMap{v1.QualifiedName("A"): {}, v1.QualifiedName("B"): {}},
				},
			},
		},
		"HashDelimiterAntiCollision": { // "A"+"BC" and "AB"+"C" both concat to "ABC".
			templateSlices: []*resourceapi.ResourceSlice{makeSingleResourceSlice("driver", "pool", poolDevices{deviceCount: 1, shape: deviceShapeA_BC})},
			nodeSlices:     []*resourceapi.ResourceSlice{makeSingleResourceSlice("driver", "pool", poolDevices{deviceCount: 1, shape: deviceShapeAB_C})},
			wantReports: []resourceDelta{
				{
					Driver:               "driver",
					TemplateResourcePool: "pool",
					NodeResourcePool:     "pool",
					TemplateSignatureMap: signatureMap{v1.QualifiedName("A"): {}, v1.QualifiedName("BC"): {}},
					NodeSignatureMap:     signatureMap{v1.QualifiedName("AB"): {}, v1.QualifiedName("C"): {}},
				},
			},
		},
		"FuzzyPriorityOverlapOverCount": {
			templateSlices: []*resourceapi.ResourceSlice{
				makeSingleResourceSlice("driver", "pool-t", poolDevices{deviceCount: 5, shape: deviceShapeABC}),
			},
			nodeSlices: []*resourceapi.ResourceSlice{
				makeSingleResourceSlice("driver", "pool-n1", poolDevices{deviceCount: 5, shape: deviceShapeAB}),
				makeSingleResourceSlice("driver", "pool-n2", poolDevices{deviceCount: 3, shape: deviceShapeABCD}),
			},
			wantReports: []resourceDelta{
				{
					Driver:               "driver",
					TemplateResourcePool: "pool-t",
					NodeResourcePool:     "pool-n2",
					TemplateSignatureMap: signatureMap{v1.QualifiedName("A"): {}, v1.QualifiedName("B"): {}, v1.QualifiedName("C"): {}},
					NodeSignatureMap:     signatureMap{v1.QualifiedName("A"): {}, v1.QualifiedName("B"): {}, v1.QualifiedName("C"): {}, v1.QualifiedName("D"): {}},
					DeviceCountDelta:     2,
				},
				{
					Driver:           "driver",
					NodeResourcePool: "pool-n1",
					NodeSignatureMap: signatureMap{v1.QualifiedName("A"): {}, v1.QualifiedName("B"): {}},
					DeviceCountDelta: -5,
				},
			},
		},
		"EmptyAttributesBoundary": {
			// Verifies that devices with zero attributes are processed and diffed correctly
			templateSlices: []*resourceapi.ResourceSlice{
				makeSingleResourceSlice("driver", "pool", poolDevices{deviceCount: 2, shape: deviceShapeEmpty}),
			},
			nodeSlices: []*resourceapi.ResourceSlice{
				makeSingleResourceSlice("driver", "pool", poolDevices{deviceCount: 1, shape: deviceShapeEmpty}),
				makeSingleResourceSlice("driver", "missing", poolDevices{deviceCount: 1, shape: deviceShapeA}),
			},
			wantReports: []resourceDelta{
				{
					Driver:               "driver",
					TemplateResourcePool: "pool",
					NodeResourcePool:     "pool",
					TemplateSignatureMap: signatureMap{},
					NodeSignatureMap:     signatureMap{},
					DeviceCountDelta:     1,
				},
				{
					Driver:           "driver",
					NodeResourcePool: "missing",
					NodeSignatureMap: signatureMap{v1.QualifiedName("A"): {}},
					DeviceCountDelta: -1,
				},
			},
		},
		"DriverWithPoolInFlux_GenerationMismatch": {
			templateSlices: []*resourceapi.ResourceSlice{
				{
					Spec: resourceapi.ResourceSliceSpec{
						Driver: "driver",
						Pool: resourceapi.ResourcePool{
							Name: "pool",
						},
						Devices: []resourceapi.Device{
							{
								Attributes: makeAttributesFromShape(deviceShapeA),
							},
						},
					},
				},
			},
			nodeSlices: []*resourceapi.ResourceSlice{
				{
					Spec: resourceapi.ResourceSliceSpec{
						Driver: "driver",
						Pool: resourceapi.ResourcePool{
							Name:       "pool",
							Generation: 1,
						},
						Devices: []resourceapi.Device{
							{
								Attributes: makeAttributesFromShape(deviceShapeB),
							},
						},
					},
				},
				{
					Spec: resourceapi.ResourceSliceSpec{
						Driver: "driver",
						Pool: resourceapi.ResourcePool{
							Name:       "pool",
							Generation: 2,
						},
						Devices: []resourceapi.Device{
							{
								Attributes: makeAttributesFromShape(deviceShapeABC),
							},
						},
					},
				},
			},
			wantReports: []resourceDelta{},
		},
		"DriverWithPoolInFlux_IncompletePool": {
			templateSlices: []*resourceapi.ResourceSlice{
				{
					Spec: resourceapi.ResourceSliceSpec{
						Driver: "driver",
						Pool: resourceapi.ResourcePool{
							Name: "pool",
						},
						Devices: []resourceapi.Device{
							{
								Attributes: makeAttributesFromShape(deviceShapeA),
							},
						},
					},
				},
			},
			nodeSlices: []*resourceapi.ResourceSlice{
				{
					Spec: resourceapi.ResourceSliceSpec{
						Driver: "driver",
						Pool: resourceapi.ResourcePool{
							Name:               "pool",
							ResourceSliceCount: 2,
						},
						Devices: []resourceapi.Device{
							{
								Attributes: makeAttributesFromShape(deviceShapeB),
							},
						},
					},
				},
			},
			wantReports: []resourceDelta{},
		},
		"AllInOne": {
			templateSlices: []*resourceapi.ResourceSlice{
				// DRIVER 1: GPU
				makeSingleResourceSlice("gpu-driver", "gpu-expected-pool", poolDevices{deviceCount: 2, shape: deviceShapeAB}),
				makeSingleResourceSlice("gpu-driver", "gpu-expected-pool", poolDevices{deviceCount: 2, shape: deviceShapeAB}),
				// DRIVER 2: CUSTOM
				makeSingleResourceSlice("custom-driver", "custom-expected-pool",
					poolDevices{deviceCount: 2, shape: deviceShapeA},
					poolDevices{deviceCount: 2, shape: deviceShapeAB},
				),
				// DRIVER 3: MISSING
				makeSingleResourceSlice("missing-driver", "missing-pool", poolDevices{deviceCount: 1, shape: deviceShapeA}),
			},
			nodeSlices: []*resourceapi.ResourceSlice{
				// DRIVER 1: GPU
				makeSingleResourceSlice("gpu-driver", "gpu-actual-pool", poolDevices{deviceCount: 3, shape: deviceShapeAB}),
				makeSingleResourceSlice("gpu-driver", "gpu-actual-pool", poolDevices{deviceCount: 1, shape: deviceShapeAB}),
				// DRIVER 2: CUSTOM
				makeSingleResourceSlice("custom-driver", "custom-actual-exact", poolDevices{deviceCount: 2, shape: deviceShapeA}),
				makeSingleResourceSlice("custom-driver", "custom-actual-fuzzy-1", poolDevices{deviceCount: 2, shape: deviceShapeB}),
				makeSingleResourceSlice("custom-driver", "custom-actual-fuzzy-2", poolDevices{deviceCount: 5, shape: deviceShapeA}),
				// DRIVER 4: ROGUE
				makeSingleResourceSlice("rogue-driver", "rogue-pool", poolDevices{deviceCount: 99, shape: deviceShapeAB}),
			},
			wantReports: []resourceDelta{
				{
					Driver:               "custom-driver",
					TemplateResourcePool: "custom-expected-pool",
					NodeResourcePool:     "custom-actual-fuzzy-1",
					TemplateSignatureMap: signatureMap{v1.QualifiedName("A"): {}, v1.QualifiedName("B"): {}},
					NodeSignatureMap:     signatureMap{v1.QualifiedName("B"): {}},
				},
				{
					Driver:               "missing-driver",
					TemplateResourcePool: "missing-pool",
					TemplateSignatureMap: signatureMap{v1.QualifiedName("A"): {}},
					NodeSignatureMap:     signatureMap{},
					DeviceCountDelta:     1,
				},
				{
					Driver:           "custom-driver",
					NodeResourcePool: "custom-actual-fuzzy-2",
					NodeSignatureMap: signatureMap{v1.QualifiedName("A"): {}},
					DeviceCountDelta: -5,
				},
			},
		},
	}

	cmdOpt := cmpopts.EquateEmpty()
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			reports := compareDraResources(test.templateSlices, test.nodeSlices, nil)
			reports = normalizeReports(reports)
			test.wantReports = normalizeReports(test.wantReports)
			if diff := cmp.Diff(test.wantReports, reports, cmdOpt); diff != "" {
				t.Errorf("CompareDraResources() diff (-want +got):\n%s", diff)
			}
		})
	}
}

func BenchmarkCompareDraResourcesExact(b *testing.B) {
	templateSlices := []*resourceapi.ResourceSlice{
		makeResourceSlice("gpu-driver", "gpu-expected-pool", 1, poolDevices{deviceCount: 10, shape: deviceShapeAB}),
	}
	nodeSlices := []*resourceapi.ResourceSlice{
		makeResourceSlice("gpu-driver", "gpu-actual-pool", 1, poolDevices{deviceCount: 10, shape: deviceShapeAB}),
	}
	for i := 0; i < b.N; i++ {
		compareDraResources(templateSlices, nodeSlices, nil)
	}
}

func BenchmarkCompareDraResourcesFuzzy(b *testing.B) {
	templateSlices := []*resourceapi.ResourceSlice{
		makeResourceSlice("gpu-driver", "gpu-expected-pool", 1, poolDevices{deviceCount: 10, shape: deviceShapeAB}),
	}
	nodeSlices := []*resourceapi.ResourceSlice{
		makeResourceSlice("gpu-driver", "gpu-actual-pool", 1, poolDevices{deviceCount: 10, shape: deviceShapeABC}),
	}
	for i := 0; i < b.N; i++ {
		compareDraResources(templateSlices, nodeSlices, nil)
	}
}

func BenchmarkCompareDraResourcesRankingFuzzy(b *testing.B) {
	templateSlices := []*resourceapi.ResourceSlice{
		makeResourceSlice("gpu-driver", "gpu-expected-pool", 1, poolDevices{deviceCount: 10, shape: deviceShapeAB}),
	}
	nodeSlices := []*resourceapi.ResourceSlice{
		makeResourceSlice("gpu-driver", "gpu-actual-pool", 3, poolDevices{deviceCount: 9, shape: deviceShapeABC}),
		makeResourceSlice("gpu-driver", "gpu-actual-pool", 3, poolDevices{deviceCount: 11, shape: deviceShapeA}),
		makeResourceSlice("gpu-driver", "gpu-actual-pool", 3, poolDevices{deviceCount: 10, shape: deviceShapeB}),
	}
	for i := 0; i < b.N; i++ {
		compareDraResources(templateSlices, nodeSlices, nil)
	}
}

func normalizeReports(reports []resourceDelta) []resourceDelta {
	slices.SortStableFunc(reports, func(a, b resourceDelta) int {
		if diff := gocmp.Compare(a.Driver, b.Driver); diff != 0 {
			return diff
		}
		if diff := gocmp.Compare(a.TemplateResourcePool, b.TemplateResourcePool); diff != 0 {
			return diff
		}
		if diff := gocmp.Compare(a.NodeResourcePool, b.NodeResourcePool); diff != 0 {
			return diff
		}
		return 0
	})

	return reports
}

type poolDevices struct {
	deviceCount int
	shape       map[string]struct{}
}

func makeAttributesFromShape(shape map[string]struct{}) map[resourceapi.QualifiedName]resourceapi.DeviceAttribute {
	attributes := make(map[resourceapi.QualifiedName]resourceapi.DeviceAttribute)
	for attr := range shape {
		attributes[resourceapi.QualifiedName(attr)] = resourceapi.DeviceAttribute{}
	}
	return attributes
}

func makeResourceSlice(driver string, pool string, slicesCount int, devices ...poolDevices) *resourceapi.ResourceSlice {
	totalDevices := 0
	for _, shape := range devices {
		totalDevices += shape.deviceCount
	}

	sliceDevices := make([]resourceapi.Device, totalDevices)
	createdDevices := 0
	for i, shape := range devices {
		for j := 0; j < shape.deviceCount; j++ {
			sliceDevices[createdDevices] = resourceapi.Device{
				Name:       fmt.Sprintf("dev-%d-%d", i, j),
				Attributes: makeAttributesFromShape(shape.shape),
			}
			createdDevices++
		}
	}

	return &resourceapi.ResourceSlice{
		ObjectMeta: metav1.ObjectMeta{
			Name: fmt.Sprintf("slice-%s-%d", pool, slicesCount),
		},
		Spec: resourceapi.ResourceSliceSpec{
			Driver: driver,
			Pool: resourceapi.ResourcePool{
				Name:               pool,
				ResourceSliceCount: int64(slicesCount),
				Generation:         1,
			},
			Devices: sliceDevices,
		},
	}
}

func makeSingleResourceSlice(driver string, pool string, devices ...poolDevices) *resourceapi.ResourceSlice {
	return makeResourceSlice(driver, pool, 1, devices...)
}
