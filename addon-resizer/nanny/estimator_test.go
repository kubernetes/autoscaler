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
	"testing"

	"k8s.io/apimachinery/pkg/api/resource"
	corev1 "k8s.io/api/core/v1"
)

var (
	fullEstimator = LinearEstimator{
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
	}
	noCPUEstimator = LinearEstimator{
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
	}
	noMemoryEstimator = LinearEstimator{
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
	}
	noStorageEstimator = LinearEstimator{
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
	}
	lessThanMilliEstimator = LinearEstimator{
		Resources: []Resource{
			{
				Base:         resource.MustParse("0.3"),
				ExtraPerNode: resource.MustParse("0.5m"),
				Name:         "cpu",
			},
		},
	}
	emptyEstimator = LinearEstimator{
		Resources: []Resource{},
	}

	exponentialEstimator = ExponentialEstimator{
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
		ScaleFactor: 1.5,
	}
	exponentialLessThanMilliEstimator = ExponentialEstimator{
		Resources: []Resource{
			{
				Base:         resource.MustParse("0.3"),
				ExtraPerNode: resource.MustParse("0.5m"),
				Name:         "cpu",
			},
		},
		ScaleFactor: 1.5,
	}

	baseResources = corev1.ResourceList{
		"cpu":     resource.MustParse("0.3"),
		"memory":  resource.MustParse("30Mi"),
		"storage": resource.MustParse("30Gi"),
	}

	noCPUBaseResources = corev1.ResourceList{
		"memory":  resource.MustParse("30Mi"),
		"storage": resource.MustParse("30Gi"),
	}
	noMemoryBaseResources = corev1.ResourceList{
		"cpu":     resource.MustParse("0.3"),
		"storage": resource.MustParse("30Gi"),
	}
	noStorageBaseResources = corev1.ResourceList{
		"cpu":    resource.MustParse("0.3"),
		"memory": resource.MustParse("30Mi"),
	}
	threeNodeResources = corev1.ResourceList{
		"cpu":     resource.MustParse("3.3"),
		"memory":  resource.MustParse("33Mi"),
		"storage": resource.MustParse("33Gi"),
	}
	threeNodeNoCPUResources = corev1.ResourceList{
		"memory":  resource.MustParse("33Mi"),
		"storage": resource.MustParse("33Gi"),
	}
	threeNodeNoMemoryResources = corev1.ResourceList{
		"cpu":     resource.MustParse("3.3"),
		"storage": resource.MustParse("33Gi"),
	}
	threeNodeNoStorageResources = corev1.ResourceList{
		"cpu":    resource.MustParse("3.3"),
		"memory": resource.MustParse("33Mi"),
	}
	threeNodeLessThanMilliResources = corev1.ResourceList{
		"cpu": resource.MustParse("0.3015"),
	}
	threeNodeLessThanMilliExpResources = corev1.ResourceList{
		"cpu": resource.MustParse("0.308"),
	}
	noResources = corev1.ResourceList{}

	sixteenNodeResources = corev1.ResourceList{
		"cpu":     resource.MustParse("16.3"),
		"memory":  resource.MustParse("46Mi"),
		"storage": resource.MustParse("46Gi"),
	}
	twentyFourNodeResources = corev1.ResourceList{
		"cpu":     resource.MustParse("24.3"),
		"memory":  resource.MustParse("54Mi"),
		"storage": resource.MustParse("54Gi"),
	}
)

func verifyResources(t *testing.T, kind string, got, want corev1.ResourceList) {
	if len(got) != len(want) {
		t.Errorf("%s not equal got: %+v want: %+v", kind, got, want)
	}
	for res, val := range want {
		actVal, ok := got[res]
		if !ok {
			t.Errorf("missing resource %s in %s", res, kind)
		}
		if val.Cmp(actVal) != 0 {
			t.Errorf("not equal resource %s in %s, got: %+v, want: %+v", res, kind, actVal, val)
		}
	}
}

func TestEstimateResources(t *testing.T) {
	testCases := []struct {
		e        ResourceEstimator
		numNodes uint64
		limits   corev1.ResourceList
		requests corev1.ResourceList
	}{
		{fullEstimator, 0, baseResources, baseResources},
		{fullEstimator, 3, threeNodeResources, threeNodeResources},
		{fullEstimator, 16, sixteenNodeResources, sixteenNodeResources},
		{fullEstimator, 24, twentyFourNodeResources, twentyFourNodeResources},
		{noCPUEstimator, 0, noCPUBaseResources, noCPUBaseResources},
		{noCPUEstimator, 3, threeNodeNoCPUResources, threeNodeNoCPUResources},
		{noMemoryEstimator, 0, noMemoryBaseResources, noMemoryBaseResources},
		{noMemoryEstimator, 3, threeNodeNoMemoryResources, threeNodeNoMemoryResources},
		{noStorageEstimator, 0, noStorageBaseResources, noStorageBaseResources},
		{noStorageEstimator, 3, threeNodeNoStorageResources, threeNodeNoStorageResources},
		{lessThanMilliEstimator, 3, threeNodeLessThanMilliResources, threeNodeLessThanMilliResources},
		{emptyEstimator, 0, noResources, noResources},
		{emptyEstimator, 3, noResources, noResources},
		{exponentialEstimator, 0, sixteenNodeResources, sixteenNodeResources},
		{exponentialEstimator, 3, sixteenNodeResources, sixteenNodeResources},
		{exponentialEstimator, 10, sixteenNodeResources, sixteenNodeResources},
		{exponentialEstimator, 16, sixteenNodeResources, sixteenNodeResources},
		{exponentialEstimator, 17, twentyFourNodeResources, twentyFourNodeResources},
		{exponentialEstimator, 20, twentyFourNodeResources, twentyFourNodeResources},
		{exponentialEstimator, 24, twentyFourNodeResources, twentyFourNodeResources},
		{exponentialLessThanMilliEstimator, 3, threeNodeLessThanMilliExpResources, threeNodeLessThanMilliExpResources},
	}

	for _, tc := range testCases {
		got := tc.e.scaleWithNodes(tc.numNodes)
		want := &corev1.ResourceRequirements{
			Limits:   tc.limits,
			Requests: tc.requests,
		}
		verifyResources(t, "limits", got.Limits, want.Limits)
		verifyResources(t, "requests", got.Requests, want.Limits)
	}
}
