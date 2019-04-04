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
)

func TestCheckResources(t *testing.T) {
	testCases := []struct {
		x    api.ResourceList
		e    *EstimatorResult
		res  api.ResourceName
		want *api.ResourceList
	}{
		// Test for the CPU resource type.
		{standard, standardRecommended, "cpu", nil},
		{belowStandard, standardRecommended, "cpu", nil},
		{aboveStandard, standardRecommended, "cpu", nil},
		{standard, standardAcceptableAboveRecommended, "cpu", nil},
		{standard, standardAcceptableBelowRecommended, "cpu", nil},
		{standard, standardAboveAcceptable, "cpu", &belowStandard},
		{standard, standardBelowAcceptable, "cpu", &aboveStandard},
		{noStorage, standardRecommended, "cpu", nil},
		{noMemory, standardRecommended, "cpu", nil},
		{noCPU, standardRecommended, "cpu", &belowStandard},
		{smallStorage, standardRecommended, "cpu", nil},
		{smallMemory, standardRecommended, "cpu", nil},
		{smallCPU, standardRecommended, "cpu", &belowStandard},
		{bigStorage, standardRecommended, "cpu", nil},
		{bigMemory, standardRecommended, "cpu", nil},
		{bigCPU, standardRecommended, "cpu", &aboveStandard},

		// Test for the memory resource type.
		{standard, standardRecommended, "memory", nil},
		{belowStandard, standardRecommended, "memory", nil},
		{aboveStandard, standardRecommended, "memory", nil},
		{standard, standardAcceptableAboveRecommended, "memory", nil},
		{standard, standardAcceptableBelowRecommended, "memory", nil},
		{standard, standardAboveAcceptable, "memory", &belowStandard},
		{standard, standardBelowAcceptable, "memory", &aboveStandard},
		{noStorage, standardRecommended, "memory", nil},
		{noMemory, standardRecommended, "memory", &belowStandard},
		{noCPU, standardRecommended, "memory", nil},
		{smallStorage, standardRecommended, "memory", nil},
		{smallMemory, standardRecommended, "memory", &belowStandard},
		{smallCPU, standardRecommended, "memory", nil},
		{bigStorage, standardRecommended, "memory", nil},
		{bigMemory, standardRecommended, "memory", &aboveStandard},
		{bigCPU, standardRecommended, "memory", nil},

		// Test for the storage resource type.
		{standard, standardRecommended, "storage", nil},
		{belowStandard, standardRecommended, "storage", nil},
		{aboveStandard, standardRecommended, "storage", nil},
		{standard, standardAcceptableAboveRecommended, "storage", nil},
		{standard, standardAcceptableBelowRecommended, "storage", nil},
		{standard, standardAboveAcceptable, "storage", &belowStandard},
		{standard, standardBelowAcceptable, "storage", &aboveStandard},
		{noStorage, standardRecommended, "storage", &belowStandard},
		{noMemory, standardRecommended, "storage", nil},
		{noCPU, standardRecommended, "storage", nil},
		{smallStorage, standardRecommended, "storage", &belowStandard},
		{smallMemory, standardRecommended, "storage", nil},
		{smallCPU, standardRecommended, "storage", nil},
		{bigStorage, standardRecommended, "storage", &aboveStandard},
		{bigMemory, standardRecommended, "storage", nil},
		{bigCPU, standardRecommended, "storage", nil},

		// Test successful comparison when not all ResourceNames are present.
		{smallMemoryNoStorage, standardRecommendedNoStorage, "memory", &belowStandardNoStorage},
	}

	for i, tc := range testCases {
		got := checkResource(tc.e, tc.x, tc.res)
		if !reflect.DeepEqual(tc.want, got) {
			t.Errorf("checkResource got %v, want %v for test case %d.", got, tc.want, i)
		}
	}
}

func TestShouldOverwriteResources(t *testing.T) {
	testCases := []struct {
		res  api.ResourceList
		e    *EstimatorResult
		want *api.ResourceRequirements
	}{
		{standard, standardRecommended, nil},
		{belowStandard, standardRecommended, nil},
		{aboveStandard, standardRecommended, nil},
		{standard, standardAcceptableAboveRecommended, nil},
		{standard, standardAcceptableBelowRecommended, nil},
		{standard, standardAboveAcceptable, &api.ResourceRequirements{belowStandard, belowStandard}},
		{standard, standardBelowAcceptable, &api.ResourceRequirements{aboveStandard, aboveStandard}},
		{noStorage, standardRecommended, &api.ResourceRequirements{belowStandard, belowStandard}},
		{noMemory, standardRecommended, &api.ResourceRequirements{belowStandard, belowStandard}},
		{noCPU, standardRecommended, &api.ResourceRequirements{belowStandard, belowStandard}},
		{smallStorage, standardRecommended, &api.ResourceRequirements{belowStandard, belowStandard}},
		{smallMemory, standardRecommended, &api.ResourceRequirements{belowStandard, belowStandard}},
		{smallCPU, standardRecommended, &api.ResourceRequirements{belowStandard, belowStandard}},
		{bigStorage, standardRecommended, &api.ResourceRequirements{aboveStandard, aboveStandard}},
		{bigMemory, standardRecommended, &api.ResourceRequirements{aboveStandard, aboveStandard}},
		{bigCPU, standardRecommended, &api.ResourceRequirements{aboveStandard, aboveStandard}},

		// Test successful comparison when not all ResourceNames are present.
		{smallMemoryNoStorage, standardRecommendedNoStorage, &api.ResourceRequirements{belowStandardNoStorage, belowStandardNoStorage}},
	}
	for i, tc := range testCases {
		got := shouldOverwriteResources(tc.e, tc.res, tc.res)
		if !reflect.DeepEqual(tc.want, got) {
			t.Errorf("shouldOverwriteResources got %v, want %v for test case %d.", got, tc.want, i)
		}
	}
}
