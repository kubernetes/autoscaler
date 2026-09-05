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
	"reflect"
	"testing"
	"time"

	api "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

var (
	// ResourcesLists to compose test cases.
	standard = api.ResourceList{
		"cpu":     resource.MustParse("0.3"),
		"memory":  resource.MustParse("200Mi"),
		"storage": resource.MustParse("10Gi"),
	}
	noStorage = api.ResourceList{
		"cpu":    resource.MustParse("0.3"),
		"memory": resource.MustParse("200Mi"),
	}
	smallMemoryNoStorage = api.ResourceList{
		"cpu":    resource.MustParse("0.3"),
		"memory": resource.MustParse("100Mi"),
	}
	noMemory = api.ResourceList{
		"cpu":     resource.MustParse("0.3"),
		"storage": resource.MustParse("10Gi"),
	}
	noCPU = api.ResourceList{
		"memory":  resource.MustParse("200Mi"),
		"storage": resource.MustParse("10Gi"),
	}
	smallStorage = api.ResourceList{
		"cpu":     resource.MustParse("0.3"),
		"memory":  resource.MustParse("200Mi"),
		"storage": resource.MustParse("1Gi"),
	}
	smallMemory = api.ResourceList{
		"cpu":     resource.MustParse("0.3"),
		"memory":  resource.MustParse("100Mi"),
		"storage": resource.MustParse("10Gi"),
	}
	smallCPU = api.ResourceList{
		"cpu":     resource.MustParse("0.1"),
		"memory":  resource.MustParse("200Mi"),
		"storage": resource.MustParse("10Gi"),
	}
	bigStorage = api.ResourceList{
		"cpu":     resource.MustParse("0.3"),
		"memory":  resource.MustParse("200Mi"),
		"storage": resource.MustParse("50Gi"),
	}
	bigMemory = api.ResourceList{
		"cpu":     resource.MustParse("0.3"),
		"memory":  resource.MustParse("900Mi"),
		"storage": resource.MustParse("10Gi"),
	}
	bigCPU = api.ResourceList{
		"cpu":     resource.MustParse("0.9"),
		"memory":  resource.MustParse("200Mi"),
		"storage": resource.MustParse("10Gi"),
	}
	belowStandard = api.ResourceList{
		"cpu":     resource.MustParse("0.2"),
		"memory":  resource.MustParse("150Mi"),
		"storage": resource.MustParse("8Gi"),
	}
	wayBelowStandard = api.ResourceList{
		"cpu":     resource.MustParse("0.1"),
		"memory":  resource.MustParse("100Mi"),
		"storage": resource.MustParse("4Gi"),
	}
	aboveStandard = api.ResourceList{
		"cpu":     resource.MustParse("0.4"),
		"memory":  resource.MustParse("250Mi"),
		"storage": resource.MustParse("12Gi"),
	}
	wayAboveStandard = api.ResourceList{
		"cpu":     resource.MustParse("0.5"),
		"memory":  resource.MustParse("300Mi"),
		"storage": resource.MustParse("16Gi"),
	}
	belowStandardNoStorage = api.ResourceList{
		"cpu":    resource.MustParse("0.2"),
		"memory": resource.MustParse("150Mi"),
	}
	aboveStandardNoStorage = api.ResourceList{
		"cpu":    resource.MustParse("0.4"),
		"memory": resource.MustParse("250Mi"),
	}

	standardRecommended = &EstimatorResult{
		AcceptableRange:  ResourceListPair{belowStandard, aboveStandard},
		RecommendedRange: ResourceListPair{belowStandard, aboveStandard},
	}
	standardAcceptableBelowRecommended = &EstimatorResult{
		AcceptableRange:  ResourceListPair{belowStandard, wayAboveStandard},
		RecommendedRange: ResourceListPair{aboveStandard, wayAboveStandard},
	}
	standardAcceptableAboveRecommended = &EstimatorResult{
		AcceptableRange:  ResourceListPair{wayBelowStandard, aboveStandard},
		RecommendedRange: ResourceListPair{wayBelowStandard, belowStandard},
	}
	standardBelowAcceptable = &EstimatorResult{
		AcceptableRange:  ResourceListPair{aboveStandard, wayAboveStandard},
		RecommendedRange: ResourceListPair{aboveStandard, wayAboveStandard},
	}
	standardAboveAcceptable = &EstimatorResult{
		AcceptableRange:  ResourceListPair{wayBelowStandard, belowStandard},
		RecommendedRange: ResourceListPair{wayBelowStandard, belowStandard},
	}
	standardRecommendedNoStorage = &EstimatorResult{
		AcceptableRange:  ResourceListPair{belowStandardNoStorage, aboveStandardNoStorage},
		RecommendedRange: ResourceListPair{belowStandardNoStorage, aboveStandardNoStorage},
	}
	noDelay        = time.Duration(0)
	oneSecondDelay = time.Second
	oneMinuteDelay = time.Minute
)

func TestCheckResources(t *testing.T) {
	testCases := []struct {
		x       api.ResourceList
		e       *EstimatorResult
		res     api.ResourceName
		wantRes *api.ResourceList
		wantOp  operation
	}{
		// Test for the CPU resource type.
		{standard, standardRecommended, "cpu", nil, unknown},
		{belowStandard, standardRecommended, "cpu", nil, unknown},
		{aboveStandard, standardRecommended, "cpu", nil, unknown},
		{standard, standardAcceptableAboveRecommended, "cpu", nil, unknown},
		{standard, standardAcceptableBelowRecommended, "cpu", nil, unknown},
		{standard, standardAboveAcceptable, "cpu", &belowStandard, scaleDown},
		{standard, standardBelowAcceptable, "cpu", &aboveStandard, scaleUp},
		{noStorage, standardRecommended, "cpu", nil, unknown},
		{noMemory, standardRecommended, "cpu", nil, unknown},
		{noCPU, standardRecommended, "cpu", &belowStandard, unknown},
		{smallStorage, standardRecommended, "cpu", nil, unknown},
		{smallMemory, standardRecommended, "cpu", nil, unknown},
		{smallCPU, standardRecommended, "cpu", &belowStandard, scaleUp},
		{bigStorage, standardRecommended, "cpu", nil, unknown},
		{bigMemory, standardRecommended, "cpu", nil, unknown},
		{bigCPU, standardRecommended, "cpu", &aboveStandard, scaleDown},

		// Test for the memory resource type.
		{standard, standardRecommended, "memory", nil, unknown},
		{belowStandard, standardRecommended, "memory", nil, unknown},
		{aboveStandard, standardRecommended, "memory", nil, unknown},
		{standard, standardAcceptableAboveRecommended, "memory", nil, unknown},
		{standard, standardAcceptableBelowRecommended, "memory", nil, unknown},
		{standard, standardAboveAcceptable, "memory", &belowStandard, scaleDown},
		{standard, standardBelowAcceptable, "memory", &aboveStandard, scaleUp},
		{noStorage, standardRecommended, "memory", nil, unknown},
		{noMemory, standardRecommended, "memory", &belowStandard, unknown},
		{noCPU, standardRecommended, "memory", nil, unknown},
		{smallStorage, standardRecommended, "memory", nil, unknown},
		{smallMemory, standardRecommended, "memory", &belowStandard, scaleUp},
		{smallCPU, standardRecommended, "memory", nil, unknown},
		{bigStorage, standardRecommended, "memory", nil, unknown},
		{bigMemory, standardRecommended, "memory", &aboveStandard, scaleDown},
		{bigCPU, standardRecommended, "memory", nil, unknown},

		// Test for the storage resource type.
		{standard, standardRecommended, "storage", nil, unknown},
		{belowStandard, standardRecommended, "storage", nil, unknown},
		{aboveStandard, standardRecommended, "storage", nil, unknown},
		{standard, standardAcceptableAboveRecommended, "storage", nil, unknown},
		{standard, standardAcceptableBelowRecommended, "storage", nil, unknown},
		{standard, standardAboveAcceptable, "storage", &belowStandard, scaleDown},
		{standard, standardBelowAcceptable, "storage", &aboveStandard, scaleUp},
		{noStorage, standardRecommended, "storage", &belowStandard, unknown},
		{noMemory, standardRecommended, "storage", nil, unknown},
		{noCPU, standardRecommended, "storage", nil, unknown},
		{smallStorage, standardRecommended, "storage", &belowStandard, scaleUp},
		{smallMemory, standardRecommended, "storage", nil, unknown},
		{smallCPU, standardRecommended, "storage", nil, unknown},
		{bigStorage, standardRecommended, "storage", &aboveStandard, scaleDown},
		{bigMemory, standardRecommended, "storage", nil, unknown},
		{bigCPU, standardRecommended, "storage", nil, unknown},

		// Test successful comparison when not all ResourceNames are present.
		{smallMemoryNoStorage, standardRecommendedNoStorage, "memory", &belowStandardNoStorage, scaleUp},
	}

	for i, tc := range testCases {
		gotRes, gotOp := checkResource(tc.e, tc.x, tc.res)
		if !reflect.DeepEqual(tc.wantRes, gotRes) || !reflect.DeepEqual(tc.wantOp, gotOp) {
			t.Errorf("checkResource got (%v, %v), want (%v, %v) for test case %d.", gotRes, gotOp, tc.wantRes, tc.wantOp, i)
		}
	}
}

func TestShouldOverwriteResources(t *testing.T) {
	testCases := []struct {
		res             api.ResourceList
		e               *EstimatorResult
		wantRes         *api.ResourceRequirements
		wantOp          operation
		cpuRequestsOnly bool
	}{
		{standard, standardRecommended, nil, unknown, false},
		{belowStandard, standardRecommended, nil, unknown, false},
		{aboveStandard, standardRecommended, nil, unknown, false},
		{standard, standardAcceptableAboveRecommended, nil, unknown, false},
		{standard, standardAcceptableBelowRecommended, nil, unknown, false},
		{standard, standardAboveAcceptable, &api.ResourceRequirements{belowStandard, belowStandard}, scaleDown, false},
		{standard, standardBelowAcceptable, &api.ResourceRequirements{aboveStandard, aboveStandard}, scaleUp, false},
		{noStorage, standardRecommended, &api.ResourceRequirements{belowStandard, belowStandard}, unknown, false},
		{noMemory, standardRecommended, &api.ResourceRequirements{belowStandard, belowStandard}, unknown, false},
		{noCPU, standardRecommended, &api.ResourceRequirements{belowStandard, belowStandard}, unknown, false},
		{smallStorage, standardRecommended, &api.ResourceRequirements{belowStandard, belowStandard}, scaleUp, false},
		{smallMemory, standardRecommended, &api.ResourceRequirements{belowStandard, belowStandard}, scaleUp, false},
		{smallCPU, standardRecommended, &api.ResourceRequirements{belowStandard, belowStandard}, scaleUp, false},
		{smallCPU, standardRecommended, &api.ResourceRequirements{Limits: smallCPU, Requests: belowStandard}, scaleUp, true},
		{smallCPU, &EstimatorResult{}, nil, unknown, true},
		{bigStorage, standardRecommended, &api.ResourceRequirements{aboveStandard, aboveStandard}, scaleDown, false},
		{bigMemory, standardRecommended, &api.ResourceRequirements{aboveStandard, aboveStandard}, scaleDown, false},
		{bigCPU, standardRecommended, &api.ResourceRequirements{aboveStandard, aboveStandard}, scaleDown, false},

		// Test successful comparison when not all ResourceNames are present.
		{smallMemoryNoStorage, standardRecommendedNoStorage, &api.ResourceRequirements{belowStandardNoStorage, belowStandardNoStorage}, scaleUp, false},
	}
	for i, tc := range testCases {
		gotRes, gotOp := shouldOverwriteResources(tc.e, tc.res, tc.res, tc.cpuRequestsOnly)
		if !reflect.DeepEqual(tc.wantRes, gotRes) || !reflect.DeepEqual(tc.wantOp, gotOp) {
			t.Errorf("shouldOverwriteResources got (%v, %v), want (%v, %v) for test case %d.", gotRes, gotOp, tc.wantRes, tc.wantOp, i)
		}
	}
}

func TestUpdateResources(t *testing.T) {
	now := time.Now()
	tenSecondsAgo := now.Add(-10 * time.Second)
	oneMinuteAgo := now.Add(-time.Minute)
	oneHourAgo := now.Add(-time.Hour)
	testCases := []struct {
		res             api.ResourceList
		e               *EstimatorResult
		lc              time.Time
		sud             time.Duration
		sdd             time.Duration
		wantRes         *api.ResourceRequirements
		want            updateResult
		cpuRequestsOnly bool
	}{
		// No changes to the resources
		{standard, standardRecommended, now, noDelay, noDelay, nil, noChange, false},
		{standard, standardRecommended, oneHourAgo, noDelay, noDelay, nil, noChange, false},
		{standard, standardRecommended, oneHourAgo, oneMinuteDelay, noDelay, nil, noChange, false},
		{standard, standardRecommended, oneHourAgo, noDelay, oneMinuteDelay, nil, noChange, false},
		{standard, standardAcceptableAboveRecommended, now, noDelay, noDelay, nil, noChange, false},
		{standard, standardAcceptableBelowRecommended, now, noDelay, noDelay, nil, noChange, false},
		// Delay has not passed
		{smallCPU, standardRecommended, tenSecondsAgo, oneMinuteDelay, noDelay, nil, postpone, false},
		{smallCPU, standardRecommended, tenSecondsAgo, oneMinuteDelay, oneSecondDelay, nil, postpone, false},
		{bigCPU, standardRecommended, tenSecondsAgo, noDelay, oneMinuteDelay, nil, postpone, false},
		{bigCPU, standardRecommended, tenSecondsAgo, oneSecondDelay, oneMinuteDelay, nil, postpone, false},
		// Delay has passed
		{smallCPU, standardRecommended, oneMinuteAgo, oneMinuteDelay, noDelay, &api.ResourceRequirements{belowStandard, belowStandard}, overwrite, false},
		{smallCPU, standardRecommended, oneMinuteAgo, oneMinuteDelay, noDelay, &api.ResourceRequirements{Limits: smallCPU, Requests: belowStandard}, overwrite, true},
		{smallCPU, &EstimatorResult{}, oneMinuteAgo, oneMinuteDelay, noDelay, nil, noChange, true},
		{bigCPU, standardRecommended, oneMinuteAgo, noDelay, oneMinuteDelay, &api.ResourceRequirements{aboveStandard, aboveStandard}, overwrite, false},
		{smallCPU, standardRecommended, oneHourAgo, oneMinuteDelay, noDelay, &api.ResourceRequirements{belowStandard, belowStandard}, overwrite, false},
		{bigCPU, standardRecommended, oneHourAgo, noDelay, oneMinuteDelay, &api.ResourceRequirements{aboveStandard, aboveStandard}, overwrite, false},
	}
	for i, tc := range testCases {
		k8s := newFakeKubernetesClient(10, tc.res, tc.res)
		est := newFakeResourceEstimator(tc.e)
		got := updateResources(k8s, est, now, tc.lc, tc.sdd, tc.sud, noChange, tc.cpuRequestsOnly)
		if tc.want != got {
			t.Errorf("updateResources got %d, want %d for test case %d.", got, tc.want, i)
		}
		if tc.want == overwrite && got == overwrite && !reflect.DeepEqual(tc.wantRes, k8s.newResources) {
			t.Errorf("updateResources got %v, want %v for test case %d.", k8s.newResources, tc.wantRes, i)
		}
	}
}

type fakeKubernetesClient struct {
	nodes        uint64
	resources    *api.ResourceRequirements
	newResources *api.ResourceRequirements
}

func newFakeKubernetesClient(nodes uint64, limits, reqs api.ResourceList) *fakeKubernetesClient {
	return &fakeKubernetesClient{
		nodes: 10,
		resources: &api.ResourceRequirements{
			Limits:   limits,
			Requests: reqs,
		},
	}
}

func (f *fakeKubernetesClient) CountNodes() (uint64, error) {
	return f.nodes, nil
}

func (f *fakeKubernetesClient) ContainerResources() (*api.ResourceRequirements, error) {
	return f.resources, nil
}

func (f *fakeKubernetesClient) UpdateDeployment(resources *api.ResourceRequirements) error {
	f.newResources = resources
	return nil
}

func (f *fakeKubernetesClient) Stop() {
}

type fakeResourceEstimator struct {
	result *EstimatorResult
}

func newFakeResourceEstimator(result *EstimatorResult) *fakeResourceEstimator {
	return &fakeResourceEstimator{
		result: result,
	}
}

func (f *fakeResourceEstimator) scaleWithNodes(numNodes uint64) *EstimatorResult {
	return f.result
}
