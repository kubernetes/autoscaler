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
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

var (
	// ResourcesLists to compose test cases.
	standard = corev1.ResourceList{
		"cpu":     resource.MustParse("0.3"),
		"memory":  resource.MustParse("200Mi"),
		"storage": resource.MustParse("10Gi"),
	}
	siStandard = corev1.ResourceList{
		"cpu":     resource.MustParse("0.3"),
		"memory":  resource.MustParse("200M"),
		"storage": resource.MustParse("10G"),
	}
	noStorage = corev1.ResourceList{
		"cpu":    resource.MustParse("0.3"),
		"memory": resource.MustParse("200Mi"),
	}
	siNoStorage = corev1.ResourceList{
		"cpu":    resource.MustParse("0.3"),
		"memory": resource.MustParse("200M"),
	}
	smallMemoryNoStorage = corev1.ResourceList{
		"cpu":    resource.MustParse("0.3"),
		"memory": resource.MustParse("100Mi"),
	}
	noMemory = corev1.ResourceList{
		"cpu":     resource.MustParse("0.3"),
		"storage": resource.MustParse("10Gi"),
	}
	noCPU = corev1.ResourceList{
		"memory":  resource.MustParse("200Mi"),
		"storage": resource.MustParse("10Gi"),
	}
	smallStorage = corev1.ResourceList{
		"cpu":     resource.MustParse("0.3"),
		"memory":  resource.MustParse("200Mi"),
		"storage": resource.MustParse("1Gi"),
	}
	smallMemory = corev1.ResourceList{
		"cpu":     resource.MustParse("0.3"),
		"memory":  resource.MustParse("100Mi"),
		"storage": resource.MustParse("10Gi"),
	}
	smallCPU = corev1.ResourceList{
		"cpu":     resource.MustParse("0.1"),
		"memory":  resource.MustParse("200Mi"),
		"storage": resource.MustParse("10Gi"),
	}
	bigStorage = corev1.ResourceList{
		"cpu":     resource.MustParse("0.3"),
		"memory":  resource.MustParse("200Mi"),
		"storage": resource.MustParse("20Gi"),
	}
	bigMemory = corev1.ResourceList{
		"cpu":     resource.MustParse("0.3"),
		"memory":  resource.MustParse("300Mi"),
		"storage": resource.MustParse("10Gi"),
	}
	bigCPU = corev1.ResourceList{
		"cpu":     resource.MustParse("0.5"),
		"memory":  resource.MustParse("200Mi"),
		"storage": resource.MustParse("10Gi"),
	}
	noDelay        = time.Duration(0)
	oneSecondDelay = time.Second
	oneMinuteDelay = time.Minute
)

func TestCheckResources(t *testing.T) {
	testCases := []struct {
		th            int64
		x, y          corev1.ResourceList
		res           corev1.ResourceName
		wantOverwrite bool
		wantOp        operation
	}{
		// Test no threshold for the CPU resource type.
		{0, standard, standard, "cpu", false, unknown},
		{0, standard, siStandard, "cpu", false, unknown},
		{0, standard, noStorage, "cpu", false, unknown},
		{0, standard, noMemory, "cpu", false, unknown},
		{0, standard, noCPU, "cpu", true, unknown},
		{0, standard, smallStorage, "cpu", false, unknown},
		{0, standard, smallMemory, "cpu", false, unknown},
		{0, standard, smallCPU, "cpu", true, scaleDown},
		{0, standard, bigCPU, "cpu", true, scaleUp},

		// Test no threshold for the memory resource type.
		{0, standard, standard, "memory", false, unknown},
		{0, standard, siStandard, "memory", true, scaleDown},
		{0, standard, noStorage, "memory", false, unknown},
		{0, standard, noMemory, "memory", true, unknown},
		{0, standard, noCPU, "memory", false, unknown},
		{0, standard, smallStorage, "memory", false, unknown},
		{0, standard, smallMemory, "memory", true, scaleDown},
		{0, standard, bigMemory, "memory", true, scaleUp},
		{0, standard, smallCPU, "memory", false, unknown},

		// Test no threshold for the storage resource type.
		{0, standard, standard, "storage", false, unknown},
		{0, standard, siStandard, "storage", true, scaleDown},
		{0, standard, noStorage, "storage", true, unknown},
		{0, standard, noMemory, "storage", false, unknown},
		{0, standard, noCPU, "storage", false, unknown},
		{0, standard, smallStorage, "storage", true, scaleDown},
		{0, standard, bigStorage, "storage", true, scaleUp},
		{0, standard, smallMemory, "storage", false, unknown},
		{0, standard, smallCPU, "storage", false, unknown},

		// Test large threshold for the CPU resource type.
		{10, standard, standard, "cpu", false, unknown},
		{10, standard, siStandard, "cpu", false, unknown},
		{10, standard, noStorage, "cpu", false, unknown},
		{10, standard, noMemory, "cpu", false, unknown},
		{10, standard, noCPU, "cpu", true, unknown},
		{10, standard, smallStorage, "cpu", false, unknown},
		{10, standard, smallMemory, "cpu", false, unknown},
		{10, standard, smallCPU, "cpu", true, scaleDown},
		{10, standard, bigCPU, "cpu", true, scaleUp},

		// Test large threshold for the memory resource type.
		{10, standard, standard, "memory", false, unknown},
		{10, standard, siStandard, "memory", false, unknown},
		{10, standard, noStorage, "memory", false, unknown},
		{10, standard, noMemory, "memory", true, unknown},
		{10, standard, noCPU, "memory", false, unknown},
		{10, standard, smallStorage, "memory", false, unknown},
		{10, standard, smallMemory, "memory", true, scaleDown},
		{10, standard, bigMemory, "memory", true, scaleUp},
		{10, standard, smallCPU, "memory", false, unknown},

		// Test large threshold for the storage resource type.
		{10, standard, standard, "storage", false, unknown},
		{10, standard, siStandard, "storage", false, unknown},
		{10, standard, noStorage, "storage", true, unknown},
		{10, standard, noMemory, "storage", false, unknown},
		{10, standard, noCPU, "storage", false, unknown},
		{10, standard, smallStorage, "storage", true, scaleDown},
		{10, standard, bigStorage, "storage", true, scaleUp},
		{10, standard, smallMemory, "storage", false, unknown},
		{10, standard, smallCPU, "storage", false, unknown},

		// Test successful comparison when not all ResourceNames are present.
		{0, noStorage, siNoStorage, "cpu", false, unknown},
		{0, noStorage, siNoStorage, "memory", true, scaleDown},
		{10, noStorage, siNoStorage, "cpu", false, unknown},
		{10, noStorage, siNoStorage, "memory", false, unknown},
		{10, noStorage, smallMemoryNoStorage, "memory", true, scaleDown},
	}

	for i, tc := range testCases {
		gotOverwrite, gotOp := checkResource(tc.th, tc.x, tc.y, tc.res)
		if tc.wantOverwrite != gotOverwrite || tc.wantOp != gotOp {
			t.Errorf("checkResource got (%t, %v), want (%t, %v) for test case %d.", gotOverwrite, gotOp, tc.wantOverwrite, tc.wantOp, i)
		}
	}
}

func TestShouldOverwriteResources(t *testing.T) {
	testCases := []struct {
		th            int64
		x, y          corev1.ResourceList
		wantOverwrite bool
		wantOp        operation
	}{
		// Test no threshold.
		{0, standard, standard, false, unknown}, // A threshold of 0 should be exact.
		{0, standard, siStandard, true, scaleDown},
		{0, standard, noStorage, true, unknown}, // Overwrite on qualitative differences.
		{0, standard, noMemory, true, unknown},
		{0, standard, noCPU, true, unknown},
		{0, standard, smallStorage, true, scaleDown}, // Overwrite past the threshold.
		{0, standard, smallMemory, true, scaleDown},
		{0, standard, smallCPU, true, scaleDown},
		{0, standard, bigStorage, true, scaleUp},
		{0, standard, bigMemory, true, scaleUp},
		{0, standard, bigCPU, true, scaleUp},

		// Test a large threshold.
		{10, standard, standard, false, unknown},
		{10, standard, siStandard, false, unknown}, // A threshold of 10 gives leeway.
		{10, standard, noStorage, true, unknown},
		{10, standard, noMemory, true, unknown},
		{10, standard, noCPU, true, unknown},
		{10, standard, smallStorage, true, scaleDown}, // The differences are larger than the threshold.
		{10, standard, smallMemory, true, scaleDown},
		{10, standard, smallCPU, true, scaleDown},
		{10, standard, bigStorage, true, scaleUp},
		{10, standard, bigMemory, true, scaleUp},
		{10, standard, bigCPU, true, scaleUp},

		// Test successful comparison when not all ResourceNames are present.
		{10, noStorage, siNoStorage, false, unknown},
	}
	for i, tc := range testCases {

		gotOverwrite, gotOp := shouldOverwriteResources(tc.th, tc.x, tc.x, tc.y, tc.x)
		if tc.wantOverwrite != gotOverwrite || tc.wantOp != gotOp {
			t.Errorf("shouldOverwriteResources got (%t, %v), want (%t, %v) for test case %d.", gotOverwrite, gotOp, tc.wantOverwrite, tc.wantOp, i)
		}
	}
}

func TestUpdateResources(t *testing.T) {
	now := time.Now()
	tenSecondsAgo := now.Add(-10 * time.Second)
	oneMinuteAgo := now.Add(-time.Minute)
	oneHourAgo := now.Add(-time.Hour)
	testCases := []struct {
		th   uint64
		x, y corev1.ResourceList
		lc   time.Time
		sud  time.Duration
		sdd  time.Duration
		want updateResult
	}{
		// No changes to the resources
		{0, standard, standard, now, noDelay, noDelay, noChange},
		{0, standard, standard, oneHourAgo, noDelay, noDelay, noChange},
		{0, standard, standard, oneHourAgo, oneMinuteDelay, noDelay, noChange},
		{0, standard, standard, oneHourAgo, noDelay, oneMinuteDelay, noChange},
		{10, standard, siStandard, now, noDelay, noDelay, noChange},
		// Delay has not passed
		{0, standard, bigCPU, tenSecondsAgo, oneMinuteDelay, noDelay, postpone},
		{0, standard, bigCPU, tenSecondsAgo, oneMinuteDelay, oneSecondDelay, postpone},
		{0, standard, smallCPU, tenSecondsAgo, noDelay, oneMinuteDelay, postpone},
		{0, standard, smallCPU, tenSecondsAgo, oneSecondDelay, oneMinuteDelay, postpone},
		// Delay has passed
		{0, standard, bigCPU, oneMinuteAgo, oneMinuteDelay, noDelay, overwrite},
		{0, standard, smallCPU, oneMinuteAgo, noDelay, oneMinuteDelay, overwrite},
		{0, standard, bigCPU, oneHourAgo, oneMinuteDelay, noDelay, overwrite},
		{0, standard, smallCPU, oneHourAgo, noDelay, oneMinuteDelay, overwrite},
	}
	for i, tc := range testCases {
		k8s := newFakeKubernetesClient(10, tc.x, tc.x)
		est := newFakeResourceEstimator(tc.y, tc.x)
		got := updateResources(k8s, est, now, tc.lc, tc.sdd, tc.sud, tc.th, noChange)
		if tc.want != got {
			t.Errorf("updateResources got %d, want %d for test case %d.", got, tc.want, i)
		}
		if tc.want == overwrite && got == overwrite && k8s.newResources != est.resources {
			t.Errorf("updateResources got %v, want %v for test case %d.", k8s.newResources, est.resources, i)
		}
	}
}

type fakeKubernetesClient struct {
	nodes        uint64
	resources    *corev1.ResourceRequirements
	newResources *corev1.ResourceRequirements
}

func newFakeKubernetesClient(nodes uint64, limits, reqs corev1.ResourceList) *fakeKubernetesClient {
	return &fakeKubernetesClient{
		nodes: 10,
		resources: &corev1.ResourceRequirements{
			Limits:   limits,
			Requests: reqs,
		},
	}
}

func (f *fakeKubernetesClient) CountNodes() (uint64, error) {
	return f.nodes, nil
}

func (f *fakeKubernetesClient) ContainerResources() (*corev1.ResourceRequirements, error) {
	return f.resources, nil
}

func (f *fakeKubernetesClient) UpdateDeployment(resources *corev1.ResourceRequirements) error {
	f.newResources = resources
	return nil
}

type fakeResourceEstimator struct {
	nodes     uint64
	resources *corev1.ResourceRequirements
}

func newFakeResourceEstimator(limits, reqs corev1.ResourceList) *fakeResourceEstimator {
	return &fakeResourceEstimator{
		resources: &corev1.ResourceRequirements{
			Limits:   limits,
			Requests: reqs,
		},
	}
}

func (f *fakeResourceEstimator) scaleWithNodes(numNodes uint64) *corev1.ResourceRequirements {
	return f.resources
}
