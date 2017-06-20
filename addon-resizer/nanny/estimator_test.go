/*
Copyright 2016 The Kubernetes Authors.

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

package nanny

import (
	"runtime"
	"testing"

	resource "k8s.io/kubernetes/pkg/api/resource"
	api "k8s.io/kubernetes/pkg/api/v1"
)

var (
	fullEstimator = Estimator{
		Resources: []Resource{
			{
				Base:         resource.MustParse("0.3"),
				ExtraPerNode: resource.MustParse("1"),
				Name:         "cpu",
			},
			{
				Base:         resource.MustParse("30Mi"),
				ExtraPerNode: resource.MustParse("1Mi"),
				Name:         "memory",
			},
			{
				Base:         resource.MustParse("30Gi"),
				ExtraPerNode: resource.MustParse("1Gi"),
				Name:         "storage",
			},
		},
		AcceptanceOffset:     20,
		RecommendationOffset: 10,
	}
	noCPUEstimator = Estimator{
		Resources: []Resource{
			{
				Base:         resource.MustParse("30Mi"),
				ExtraPerNode: resource.MustParse("1Mi"),
				Name:         "memory",
			},
			{
				Base:         resource.MustParse("30Gi"),
				ExtraPerNode: resource.MustParse("1Gi"),
				Name:         "storage",
			},
		},
		AcceptanceOffset:     20,
		RecommendationOffset: 10,
	}
	noMemoryEstimator = Estimator{
		Resources: []Resource{
			{
				Base:         resource.MustParse("0.3"),
				ExtraPerNode: resource.MustParse("1"),
				Name:         "cpu",
			},
			{
				Base:         resource.MustParse("30Gi"),
				ExtraPerNode: resource.MustParse("1Gi"),
				Name:         "storage",
			},
		},
		AcceptanceOffset:     20,
		RecommendationOffset: 10,
	}
	noStorageEstimator = Estimator{
		Resources: []Resource{
			{
				Base:         resource.MustParse("0.3"),
				ExtraPerNode: resource.MustParse("1"),
				Name:         "cpu",
			},
			{
				Base:         resource.MustParse("30Mi"),
				ExtraPerNode: resource.MustParse("1Mi"),
				Name:         "memory",
			},
		},
		AcceptanceOffset:     20,
		RecommendationOffset: 10,
	}
	lessThanMilliEstimator = Estimator{
		Resources: []Resource{
			{
				Base:         resource.MustParse("0.3"),
				ExtraPerNode: resource.MustParse("0.6m"),
				Name:         "cpu",
			},
		},
		AcceptanceOffset:     20,
		RecommendationOffset: 10,
	}
	emptyEstimator = Estimator{
		Resources:            []Resource{},
		AcceptanceOffset:     20,
		RecommendationOffset: 10,
	}
	emptyRecommendedRangeEstimator = Estimator{
		Resources: []Resource{
			{
				Base:         resource.MustParse("0.3"),
				ExtraPerNode: resource.MustParse("1"),
				Name:         "cpu",
			},
			{
				Base:         resource.MustParse("30Mi"),
				ExtraPerNode: resource.MustParse("1Mi"),
				Name:         "memory",
			},
			{
				Base:         resource.MustParse("30Gi"),
				ExtraPerNode: resource.MustParse("1Gi"),
				Name:         "storage",
			},
		},
		AcceptanceOffset:     20,
		RecommendationOffset: 0,
	}

	baseResources = api.ResourceList{
		"cpu":     resource.MustParse("0.3"),
		"memory":  resource.MustParse("30Mi"),
		"storage": resource.MustParse("30Gi"),
	}
	noCPUBaseResources = api.ResourceList{
		"memory":  resource.MustParse("30Mi"),
		"storage": resource.MustParse("30Gi"),
	}
	noMemoryBaseResources = api.ResourceList{
		"cpu":     resource.MustParse("0.3"),
		"storage": resource.MustParse("30Gi"),
	}
	noStorageBaseResources = api.ResourceList{
		"cpu":    resource.MustParse("0.3"),
		"memory": resource.MustParse("30Mi"),
	}
	threeNodeResources = api.ResourceList{
		"cpu":     resource.MustParse("3.3"),
		"memory":  resource.MustParse("33Mi"),
		"storage": resource.MustParse("33Gi"),
	}
	fourNodeResources = api.ResourceList{
		"cpu":     resource.MustParse("4.3"),
		"memory":  resource.MustParse("34Mi"),
		"storage": resource.MustParse("34Gi"),
	}
	fiveNodeResources = api.ResourceList{
		"cpu":     resource.MustParse("5.3"),
		"memory":  resource.MustParse("35Mi"),
		"storage": resource.MustParse("35Gi"),
	}
	twelveNodeResources = api.ResourceList{
		"cpu":     resource.MustParse("12.3"),
		"memory":  resource.MustParse("42Mi"),
		"storage": resource.MustParse("42Gi"),
	}
	fourteenNodeResources = api.ResourceList{
		"cpu":     resource.MustParse("14.3"),
		"memory":  resource.MustParse("44Mi"),
		"storage": resource.MustParse("44Gi"),
	}
	eighteenNodeResources = api.ResourceList{
		"cpu":     resource.MustParse("18.3"),
		"memory":  resource.MustParse("48Mi"),
		"storage": resource.MustParse("48Gi"),
	}
	nineteenNodeResources = api.ResourceList{
		"cpu":     resource.MustParse("19.3"),
		"memory":  resource.MustParse("49Mi"),
		"storage": resource.MustParse("49Gi"),
	}
	twentyNodeResources = api.ResourceList{
		"cpu":     resource.MustParse("20.3"),
		"memory":  resource.MustParse("50Mi"),
		"storage": resource.MustParse("50Gi"),
	}
	twentyOneNodeResources = api.ResourceList{
		"cpu":     resource.MustParse("21.3"),
		"memory":  resource.MustParse("51Mi"),
		"storage": resource.MustParse("51Gi"),
	}
	twentySevenNodeResources = api.ResourceList{
		"cpu":     resource.MustParse("27.3"),
		"memory":  resource.MustParse("57Mi"),
		"storage": resource.MustParse("57Gi"),
	}
	twentyNineNodeResources = api.ResourceList{
		"cpu":     resource.MustParse("29.3"),
		"memory":  resource.MustParse("59Mi"),
		"storage": resource.MustParse("59Gi"),
	}
	twoNodeNoCPUResources = api.ResourceList{
		"memory":  resource.MustParse("32Mi"),
		"storage": resource.MustParse("32Gi"),
	}
	fourNodeNoCPUResources = api.ResourceList{
		"memory":  resource.MustParse("34Mi"),
		"storage": resource.MustParse("34Gi"),
	}
	twoNodeNoMemoryResources = api.ResourceList{
		"cpu":     resource.MustParse("2.3"),
		"storage": resource.MustParse("32Gi"),
	}
	fourNodeNoMemoryResources = api.ResourceList{
		"cpu":     resource.MustParse("4.3"),
		"storage": resource.MustParse("34Gi"),
	}
	twoNodeNoStorageResources = api.ResourceList{
		"cpu":    resource.MustParse("2.3"),
		"memory": resource.MustParse("32Mi"),
	}
	fourNodeNoStorageResources = api.ResourceList{
		"cpu":    resource.MustParse("4.3"),
		"memory": resource.MustParse("34Mi"),
	}
	twoNodeLessThanMilliResources = api.ResourceList{
		"cpu": resource.MustParse("0.3012"),
	}
	fourNodeLessThanMilliResources = api.ResourceList{
		"cpu": resource.MustParse("0.3024"),
	}
	noResources = api.ResourceList{}

	sixteenNodeResources = api.ResourceList{
		"cpu":     resource.MustParse("16.3"),
		"memory":  resource.MustParse("46Mi"),
		"storage": resource.MustParse("46Gi"),
	}
	twentyFourNodeResources = api.ResourceList{
		"cpu":     resource.MustParse("24.3"),
		"memory":  resource.MustParse("54Mi"),
		"storage": resource.MustParse("54Gi"),
	}

	zeroNodesResourcesRange = ResourceListPair{
		lower: baseResources,
		upper: baseResources,
	}
	threeToFiveNodesResourcesRange = ResourceListPair{
		lower: threeNodeResources,
		upper: fiveNodeResources,
	}
	fourteenToEighteenNodesResourcesRange = ResourceListPair{
		lower: fourteenNodeResources,
		upper: eighteenNodeResources,
	}
	twelveToTwentyNodesResourcesRange = ResourceListPair{
		lower: twelveNodeResources,
		upper: twentyNodeResources,
	}
	twentyOneToTwentySevenNodesResourcesRange = ResourceListPair{
		lower: twentyOneNodeResources,
		upper: twentySevenNodeResources,
	}
	nineteenToTwentyNineNodesResourcesRange = ResourceListPair{
		lower: nineteenNodeResources,
		upper: twentyNineNodeResources,
	}
	zeroNodesNoCPUResourcesRange = ResourceListPair{
		lower: noCPUBaseResources,
		upper: noCPUBaseResources,
	}
	twoToFourNodesNoCPUResourcesRange = ResourceListPair{
		lower: twoNodeNoCPUResources,
		upper: fourNodeNoCPUResources,
	}
	zeroNodesNoMemoryResourcesRange = ResourceListPair{
		lower: noMemoryBaseResources,
		upper: noMemoryBaseResources,
	}
	twoToFourNodesNoMemoryResourcesRange = ResourceListPair{
		lower: twoNodeNoMemoryResources,
		upper: fourNodeNoMemoryResources,
	}
	zeroNodesNoStorageResourcesRange = ResourceListPair{
		lower: noStorageBaseResources,
		upper: noStorageBaseResources,
	}
	twoToFourNodesNoStorageResourcesRange = ResourceListPair{
		lower: twoNodeNoStorageResources,
		upper: fourNodeNoStorageResources,
	}
	twoToFourNodesLessThanMilliResourcesRange = ResourceListPair{
		lower: twoNodeLessThanMilliResources,
		upper: fourNodeLessThanMilliResources,
	}
	noResourcesRange = ResourceListPair{
		lower: noResources,
		upper: noResources,
	}
	fourToFourNodesResourcesRange = ResourceListPair{
		lower: fourNodeResources,
		upper: fourNodeResources,
	}
)

func verifyResources(t *testing.T, lineNum int, kind string, got, want api.ResourceList) {
	for res, val := range want {
		actVal, ok := got[res]
		if !ok {
			t.Errorf("[test@line %d] missing resource %s in %s", lineNum, res, kind)
		}
		if val.Cmp(actVal) != 0 {
			t.Errorf("[test@line %d] not equal resource %s in %s, got: %+v, want: %+v", lineNum, res, kind, actVal, val)
		}
	}
}

func verifyRange(t *testing.T, lineNum int, kind string, got, want ResourceListPair) {
	if len(got.lower) != len(want.lower) || len(got.upper) != len(want.upper) {
		t.Errorf("[test@line %d] %s not equal got: %+v want: %+v", lineNum, kind, got, want)
	}
	verifyResources(t, lineNum, kind+" (lower bound)", got.lower, want.lower)
	verifyResources(t, lineNum, kind+" (upper bound)", got.upper, want.upper)
}

func num() int {
	_, _, line, ok := runtime.Caller(1)
	if ok {
		return line
	}
	return -1
}

func TestEstimateResources(t *testing.T) {
	testCases := []struct {
		lineNum         int
		e               ResourceEstimator
		numNodes        uint64
		estimatorResult EstimatorResult
	}{
		{num(), fullEstimator, 0, EstimatorResult{zeroNodesResourcesRange, zeroNodesResourcesRange}},
		{num(), fullEstimator, 4, EstimatorResult{threeToFiveNodesResourcesRange, threeToFiveNodesResourcesRange}},
		{num(), fullEstimator, 16, EstimatorResult{fourteenToEighteenNodesResourcesRange, twelveToTwentyNodesResourcesRange}},
		{num(), fullEstimator, 24, EstimatorResult{twentyOneToTwentySevenNodesResourcesRange, nineteenToTwentyNineNodesResourcesRange}},
		{num(), noCPUEstimator, 0, EstimatorResult{zeroNodesNoCPUResourcesRange, zeroNodesNoCPUResourcesRange}},
		{num(), noCPUEstimator, 3, EstimatorResult{twoToFourNodesNoCPUResourcesRange, twoToFourNodesNoCPUResourcesRange}},
		{num(), noMemoryEstimator, 0, EstimatorResult{zeroNodesNoMemoryResourcesRange, zeroNodesNoMemoryResourcesRange}},
		{num(), noMemoryEstimator, 3, EstimatorResult{twoToFourNodesNoMemoryResourcesRange, twoToFourNodesNoMemoryResourcesRange}},
		{num(), noStorageEstimator, 0, EstimatorResult{zeroNodesNoStorageResourcesRange, zeroNodesNoStorageResourcesRange}},
		{num(), noStorageEstimator, 3, EstimatorResult{twoToFourNodesNoStorageResourcesRange, twoToFourNodesNoStorageResourcesRange}},
		{num(), lessThanMilliEstimator, 3, EstimatorResult{twoToFourNodesLessThanMilliResourcesRange, twoToFourNodesLessThanMilliResourcesRange}},
		{num(), emptyEstimator, 0, EstimatorResult{noResourcesRange, noResourcesRange}},
		{num(), emptyEstimator, 3, EstimatorResult{noResourcesRange, noResourcesRange}},
		{num(), emptyRecommendedRangeEstimator, 0, EstimatorResult{zeroNodesResourcesRange, zeroNodesResourcesRange}},
		{num(), emptyRecommendedRangeEstimator, 4, EstimatorResult{fourToFourNodesResourcesRange, threeToFiveNodesResourcesRange}},
	}

	for _, tc := range testCases {
		got := tc.e.scaleWithNodes(tc.numNodes)
		want := &tc.estimatorResult
		verifyRange(t, tc.lineNum, "AcceptableRange", got.AcceptableRange, want.AcceptableRange)
		verifyRange(t, tc.lineNum, "RecommendedRange", got.RecommendedRange, want.RecommendedRange)
	}
}
