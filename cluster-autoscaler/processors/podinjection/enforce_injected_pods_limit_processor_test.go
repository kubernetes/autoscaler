/*
Copyright 2024 The Kubernetes Authors.

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

package podinjection

import (
	"testing"

	"github.com/stretchr/testify/assert"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
)

func TestEnforceInjectedPodsLimitProcessor(t *testing.T) {

	samplePod := buildTestPod("default", "test-pod")
	ownerUid := types.UID("sample uid")

	testCases := []struct {
		name                                          string
		podLimit                                      int
		unschedulablePods                             []*apiv1.Pod
		expectedNumberOfResultedUnschedulablePods     int
		expectedNumberOfResultedUnschedulableFakePods int
		expectedNumberOfResultedUnschedulableRealPods int
	}{
		{
			name:              "Real pods = 0 && fake pods < PodLimit",
			podLimit:          10,
			unschedulablePods: makeFakePods(ownerUid, samplePod, 5),
			expectedNumberOfResultedUnschedulablePods:     5,
			expectedNumberOfResultedUnschedulableFakePods: 5,
			expectedNumberOfResultedUnschedulableRealPods: 0,
		},
		{
			name:              "Real pods = 0 && fake pods > PodLimit",
			podLimit:          10,
			unschedulablePods: makeFakePods(ownerUid, samplePod, 15),
			expectedNumberOfResultedUnschedulablePods:     10,
			expectedNumberOfResultedUnschedulableFakePods: 10,
			expectedNumberOfResultedUnschedulableRealPods: 0,
		},
		{
			name:              "Real pods > PodLimit && some fake pods",
			podLimit:          10,
			unschedulablePods: append(makeTestingPods(11), makeFakePods(ownerUid, samplePod, 5)...),
			expectedNumberOfResultedUnschedulablePods:     11,
			expectedNumberOfResultedUnschedulableFakePods: 0,
			expectedNumberOfResultedUnschedulableRealPods: 11,
		},
		{
			name:              "Real pods = PodLimit && some fake pods",
			podLimit:          10,
			unschedulablePods: append(makeTestingPods(10), makeFakePods(ownerUid, samplePod, 5)...),
			expectedNumberOfResultedUnschedulablePods:     10,
			expectedNumberOfResultedUnschedulableFakePods: 0,
			expectedNumberOfResultedUnschedulableRealPods: 10,
		},
		{
			name:              "Real pods < PodLimit && real pods + fake pods > PodLimit",
			podLimit:          10,
			unschedulablePods: append(makeTestingPods(3), makeFakePods(ownerUid, samplePod, 10)...),
			expectedNumberOfResultedUnschedulablePods:     10,
			expectedNumberOfResultedUnschedulableFakePods: 7,
			expectedNumberOfResultedUnschedulableRealPods: 3,
		},
		{
			name:              "Real pods < PodLimit && real pods + fake pods < PodLimit",
			podLimit:          10,
			unschedulablePods: append(makeTestingPods(3), makeFakePods(ownerUid, samplePod, 4)...),
			expectedNumberOfResultedUnschedulablePods:     7,
			expectedNumberOfResultedUnschedulableFakePods: 4,
			expectedNumberOfResultedUnschedulableRealPods: 3,
		},
		{
			name:              "Real pods < PodLimit && real pods + fake pods = PodLimit",
			podLimit:          10,
			unschedulablePods: append(makeTestingPods(3), makeFakePods(ownerUid, samplePod, 7)...),
			expectedNumberOfResultedUnschedulablePods:     10,
			expectedNumberOfResultedUnschedulableFakePods: 7,
			expectedNumberOfResultedUnschedulableRealPods: 3,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			p := NewEnforceInjectedPodsLimitProcessor(tc.podLimit)
			pods, _ := p.Process(nil, tc.unschedulablePods)
			assert.EqualValues(t, tc.expectedNumberOfResultedUnschedulablePods, len(pods))
			numberOfFakePods := numberOfFakePods(pods)
			assert.EqualValues(t, tc.expectedNumberOfResultedUnschedulableFakePods, numberOfFakePods)
			assert.EqualValues(t, tc.expectedNumberOfResultedUnschedulableRealPods, len(pods)-numberOfFakePods)
		})
	}
}

func numberOfFakePods(pods []*apiv1.Pod) int {
	numberOfFakePods := 0
	for _, pod := range pods {
		if IsFake(pod) {
			numberOfFakePods += 1
		}
	}
	return numberOfFakePods
}

func makeTestingPods(numberOfRealTestPods int) []*apiv1.Pod {
	var testingPods []*apiv1.Pod
	for range numberOfRealTestPods {
		testingPods = append(testingPods, buildTestPod("default", "test-pod"))
	}
	return testingPods
}
