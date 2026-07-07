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

package checkcapacity

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/autoscaler/cluster-autoscaler/apis/provisioningrequest/autoscaling.x-k8s.io/v1"
	ca_context "k8s.io/autoscaler/cluster-autoscaler/context"
	"k8s.io/autoscaler/cluster-autoscaler/processors/status"
	"k8s.io/autoscaler/cluster-autoscaler/provisioningrequest/conditions"
	"k8s.io/autoscaler/cluster-autoscaler/provisioningrequest/provreqwrapper"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/clustersnapshot"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/clustersnapshot/testsnapshot"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/scheduling"
	"k8s.io/autoscaler/cluster-autoscaler/utils/errors"
	. "k8s.io/autoscaler/cluster-autoscaler/utils/test"
)

func TestCombinedStatusSet(t *testing.T) {
	// TestCombinedStatusSet tests the CombinedStatusSet function.
	testCases := []struct {
		name          string
		statuses      []*status.ScaleUpStatus
		exportedResut status.ScaleUpResult
		exportedError errors.AutoscalerError
		returnedError errors.AutoscalerError
	}{
		{
			name:          "empty",
			statuses:      []*status.ScaleUpStatus{},
			exportedResut: status.ScaleUpNotTried,
		},
		{
			name:          "all successful",
			statuses:      generateStatuses(2, status.ScaleUpSuccessful),
			exportedResut: status.ScaleUpSuccessful,
		},
		{
			name:          "all errors",
			statuses:      generateStatuses(2, status.ScaleUpError),
			exportedResut: status.ScaleUpError,
			exportedError: errors.NewAutoscalerError(errors.InternalError, "error 0 ...and other concurrent errors: [\"error 1\"]"),
			returnedError: errors.NewAutoscalerError(errors.InternalError, "error 0 ...and other concurrent errors: [\"error 1\"]"),
		},
		{
			name:          "all no options available",
			statuses:      generateStatuses(2, status.ScaleUpNoOptionsAvailable),
			exportedResut: status.ScaleUpNoOptionsAvailable,
		},
		{
			name:          "error and successful",
			statuses:      append(generateStatuses(1, status.ScaleUpError), generateStatuses(1, status.ScaleUpSuccessful)...),
			exportedResut: status.ScaleUpSuccessful,
			exportedError: errors.NewAutoscalerError(errors.InternalError, "error 0"),
		},
		{
			name:          "error and no options available",
			statuses:      append(generateStatuses(1, status.ScaleUpError), generateStatuses(1, status.ScaleUpNoOptionsAvailable)...),
			exportedResut: status.ScaleUpError,
			exportedError: errors.NewAutoscalerError(errors.InternalError, "error 0"),
			returnedError: errors.NewAutoscalerError(errors.InternalError, "error 0"),
		},
		{
			name:          "successful and no options available",
			statuses:      append(generateStatuses(1, status.ScaleUpSuccessful), generateStatuses(1, status.ScaleUpNoOptionsAvailable)...),
			exportedResut: status.ScaleUpSuccessful,
		},
		{
			name:          "error, successful and no options available",
			statuses:      append(generateStatuses(1, status.ScaleUpNoOptionsAvailable), append(generateStatuses(1, status.ScaleUpError), generateStatuses(1, status.ScaleUpSuccessful)...)...),
			exportedResut: status.ScaleUpSuccessful,
			exportedError: errors.NewAutoscalerError(errors.InternalError, "error 0"),
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			combinedStatus := NewCombinedStatusSet()

			for _, s := range tc.statuses {
				combinedStatus.Add(s)
			}

			export, retErr := combinedStatus.Export()

			assert.Equal(t, export.Result, tc.exportedResut)

			if tc.exportedError == nil {
				assert.Nil(t, export.ScaleUpError)
			} else {
				assert.Equal(t, tc.exportedError.Error(), (*export.ScaleUpError).Error())
			}

			if tc.returnedError == nil {
				assert.Nil(t, retErr)
			} else {
				assert.Equal(t, tc.returnedError.Error(), retErr.Error())
			}
		})
	}
}

func generateStatuses(n int, result status.ScaleUpResult) []*status.ScaleUpStatus {
	// generateStatuses generates n statuses with the given result.
	statuses := make([]*status.ScaleUpStatus, n)
	for i := 0; i < n; i++ {
		var scaleUpErr *errors.AutoscalerError

		if result == status.ScaleUpError {
			newErr := errors.NewAutoscalerError(errors.InternalError, fmt.Sprintf("error %d", i))
			scaleUpErr = &newErr
		}

		statuses[i] = &status.ScaleUpStatus{Result: result, ScaleUpError: scaleUpErr}
	}
	return statuses
}

func TestGroupPodsByPodSet(t *testing.T) {
	makePod := func(name, prName string) *apiv1.Pod {
		return &apiv1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name: name,
				Annotations: map[string]string{
					v1.ProvisioningRequestPodAnnotationKey: prName,
				},
			},
		}
	}
	makeProvReq := func(name string, podSets ...v1.PodSet) *provreqwrapper.ProvisioningRequest {
		return &provreqwrapper.ProvisioningRequest{
			ProvisioningRequest: &v1.ProvisioningRequest{
				ObjectMeta: metav1.ObjectMeta{Name: name},
				Spec: v1.ProvisioningRequestSpec{
					PodSets: podSets,
				},
			},
		}
	}

	testCases := []struct {
		name     string
		pods     []*apiv1.Pod
		provReq  *provreqwrapper.ProvisioningRequest
		expected map[int][]string // podset index -> pod names
	}{
		{
			name: "groups pods by podset index",
			pods: []*apiv1.Pod{
				makePod("my-pr-0-0", "my-pr"),
				makePod("my-pr-0-1", "my-pr"),
				makePod("my-pr-1-0", "my-pr"),
			},
			provReq: makeProvReq("my-pr",
				v1.PodSet{PodTemplateRef: v1.Reference{Name: "workers"}, Count: 2},
				v1.PodSet{PodTemplateRef: v1.Reference{Name: "ps"}, Count: 1},
			),
			expected: map[int][]string{
				0: {"my-pr-0-0", "my-pr-0-1"},
				1: {"my-pr-1-0"},
			},
		},
		{
			name: "filters pods from other ProvReqs",
			pods: []*apiv1.Pod{
				makePod("my-pr-0-0", "my-pr"),
				makePod("other-pr-0-1", "other-pr"),
			},
			provReq: makeProvReq("my-pr",
				v1.PodSet{PodTemplateRef: v1.Reference{Name: "workers"}, Count: 1},
			),
			expected: map[int][]string{
				0: {"my-pr-0-0"},
			},
		},
		{
			name: "ignores pods with non-matching name pattern",
			pods: []*apiv1.Pod{
				makePod("my-pr-0-0", "my-pr"),
				makePod("no-index-suffix", "my-pr"),
			},
			provReq: makeProvReq("my-pr",
				v1.PodSet{PodTemplateRef: v1.Reference{Name: "workers"}, Count: 1},
			),
			expected: map[int][]string{
				0: {"my-pr-0-0"},
			},
		},
		{
			name: "multi-digit podset indices",
			pods: []*apiv1.Pod{
				makePod("my-pr-0-0", "my-pr"),
				makePod("my-pr-10-0", "my-pr"),
				makePod("my-pr-10-1", "my-pr"),
				makePod("my-pr-2-5", "my-pr"),
			},
			provReq: makeProvReq("my-pr",
				v1.PodSet{PodTemplateRef: v1.Reference{Name: "ps0"}, Count: 1},
				v1.PodSet{PodTemplateRef: v1.Reference{Name: "ps1"}, Count: 1},
				v1.PodSet{PodTemplateRef: v1.Reference{Name: "ps2"}, Count: 1},
				// podsets 3-9 have no pods
				v1.PodSet{PodTemplateRef: v1.Reference{Name: "ps3"}, Count: 1},
				v1.PodSet{PodTemplateRef: v1.Reference{Name: "ps4"}, Count: 1},
				v1.PodSet{PodTemplateRef: v1.Reference{Name: "ps5"}, Count: 1},
				v1.PodSet{PodTemplateRef: v1.Reference{Name: "ps6"}, Count: 1},
				v1.PodSet{PodTemplateRef: v1.Reference{Name: "ps7"}, Count: 1},
				v1.PodSet{PodTemplateRef: v1.Reference{Name: "ps8"}, Count: 1},
				v1.PodSet{PodTemplateRef: v1.Reference{Name: "ps9"}, Count: 1},
				v1.PodSet{PodTemplateRef: v1.Reference{Name: "ps10"}, Count: 2},
			),
			expected: map[int][]string{
				0:  {"my-pr-0-0"},
				2:  {"my-pr-2-5"},
				10: {"my-pr-10-0", "my-pr-10-1"},
			},
		},
		{
			name:     "empty pod list",
			pods:     []*apiv1.Pod{},
			provReq:  makeProvReq("my-pr", v1.PodSet{PodTemplateRef: v1.Reference{Name: "workers"}, Count: 1}),
			expected: map[int][]string{},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			result := groupPodsByPodSet(tc.pods, tc.provReq)
			resultNames := make(map[int][]string)
			for idx, pods := range result {
				names := make([]string, len(pods))
				for i, p := range pods {
					names[i] = p.Name
				}
				resultNames[idx] = names
			}
			assert.Equal(t, tc.expected, resultNames)
		})
	}
}

func TestSetSchedulableDetails(t *testing.T) {
	testCases := []struct {
		name                 string
		podSetNames          []string
		podCounts            map[string]int
		wantPodSetsDetail    string
		wantPodCountsDetail  bool
		wantPodCountsContent string
	}{
		{
			name:                 "sets both details when podCounts is non-nil",
			podSetNames:          []string{"workers", "ps"},
			podCounts:            map[string]int{"workers": 3, "ps": 0},
			wantPodSetsDetail:    `["workers","ps"]`,
			wantPodCountsDetail:  true,
			wantPodCountsContent: `{"ps":0,"workers":3}`,
		},
		{
			name:                "skips podCounts detail when nil",
			podSetNames:         []string{"workers"},
			podCounts:           nil,
			wantPodSetsDetail:   `["workers"]`,
			wantPodCountsDetail: false,
		},
		{
			name:                 "empty podSetNames",
			podSetNames:          []string{},
			podCounts:            map[string]int{},
			wantPodSetsDetail:    `[]`,
			wantPodCountsDetail:  true,
			wantPodCountsContent: `{}`,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			provReq := makeTestProvReq("test-pr")

			err := setSchedulableDetails(provReq, tc.podSetNames, tc.podCounts)
			require.NoError(t, err)

			assert.Equal(t, v1.Detail(tc.wantPodSetsDetail),
				provReq.Status.ProvisioningClassDetails[conditions.SchedulablePodSetsDetailKey])

			if tc.wantPodCountsDetail {
				got := provReq.Status.ProvisioningClassDetails[conditions.SchedulablePodCountsDetailKey]
				// Compare via parsed JSON to avoid map ordering issues.
				var gotMap, wantMap map[string]int
				require.NoError(t, json.Unmarshal([]byte(got), &gotMap))
				require.NoError(t, json.Unmarshal([]byte(tc.wantPodCountsContent), &wantMap))
				assert.Equal(t, wantMap, gotMap)
			} else {
				_, exists := provReq.Status.ProvisioningClassDetails[conditions.SchedulablePodCountsDetailKey]
				assert.False(t, exists)
			}
		})
	}
}

func TestResolvePartialCapacityResult(t *testing.T) {
	testCases := []struct {
		name                   string
		schedulablePodSetNames []string
		totalPodSets           int
		partialCapacityMode    string
		wantProvisionedStatus  metav1.ConditionStatus
		wantProvisionedReason  string
		wantScaleUpResult      status.ScaleUpResult
		wantSnapshotCommitted  bool
		wantDetailsPreserved   bool
	}{
		{
			name:                   "all fit / bookPartial",
			schedulablePodSetNames: []string{"workers", "ps"},
			totalPodSets:           2,
			partialCapacityMode:    PartialCapacityCheckBookPartial,
			wantProvisionedStatus:  metav1.ConditionTrue,
			wantProvisionedReason:  conditions.CapacityIsFoundReason,
			wantScaleUpResult:      status.ScaleUpSuccessful,
			wantSnapshotCommitted:  true,
		},
		{
			name:                   "all fit / checkOnly",
			schedulablePodSetNames: []string{"workers", "ps"},
			totalPodSets:           2,
			partialCapacityMode:    PartialCapacityCheckCheckOnly,
			wantProvisionedStatus:  metav1.ConditionTrue,
			wantProvisionedReason:  conditions.CapacityIsFoundReason,
			wantScaleUpResult:      status.ScaleUpSuccessful,
			wantSnapshotCommitted:  true,
		},
		{
			name:                   "some fit / bookPartial",
			schedulablePodSetNames: []string{"workers"},
			totalPodSets:           2,
			partialCapacityMode:    PartialCapacityCheckBookPartial,
			wantProvisionedStatus:  metav1.ConditionTrue,
			wantProvisionedReason:  conditions.PartialCapacityIsFoundReason,
			wantScaleUpResult:      status.ScaleUpSuccessful,
			wantSnapshotCommitted:  true,
		},
		{
			name:                   "some fit / checkOnly",
			schedulablePodSetNames: []string{"workers"},
			totalPodSets:           2,
			partialCapacityMode:    PartialCapacityCheckCheckOnly,
			wantProvisionedStatus:  metav1.ConditionFalse,
			wantProvisionedReason:  conditions.PartialCapacityIsFoundReason,
			wantScaleUpResult:      status.ScaleUpSuccessful,
			wantSnapshotCommitted:  false,
		},
		{
			name:                   "none fit / bookPartial",
			schedulablePodSetNames: []string{},
			totalPodSets:           2,
			partialCapacityMode:    PartialCapacityCheckBookPartial,
			wantProvisionedReason:  conditions.CapacityIsNotFoundReason,
			wantScaleUpResult:      status.ScaleUpNoOptionsAvailable,
			wantSnapshotCommitted:  false,
			wantDetailsPreserved:   true,
		},
		{
			name:                   "none fit / checkOnly",
			schedulablePodSetNames: []string{},
			totalPodSets:           2,
			partialCapacityMode:    PartialCapacityCheckCheckOnly,
			wantProvisionedReason:  conditions.CapacityIsNotFoundReason,
			wantScaleUpResult:      status.ScaleUpNoOptionsAvailable,
			wantSnapshotCommitted:  false,
			wantDetailsPreserved:   true,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			snapshot := testsnapshot.NewTestSnapshotOrDie(t)
			snapshot.Fork()

			provReq := makeTestProvReqWithPodSets("test-pr", tc.totalPodSets)
			provReq.SetProvisioningClassDetail(conditions.SchedulablePodSetsDetailKey, v1.Detail(`["workers"]`))
			provReq.SetProvisioningClassDetail(conditions.SchedulablePodCountsDetailKey, v1.Detail(`{"workers":2}`))

			combinedStatus := NewCombinedStatusSet()
			o := &checkCapacityProvClass{
				autoscalingCtx: &ca_context.AutoscalingContext{
					ClusterSnapshot: snapshot,
				},
			}

			err := o.resolvePartialCapacityResult(provReq, &combinedStatus, tc.partialCapacityMode, tc.schedulablePodSetNames)
			require.NoError(t, err)

			assert.Equal(t, tc.wantScaleUpResult, combinedStatus.Result)

			cond := findCondition(provReq, v1.Provisioned)
			require.NotNil(t, cond, "expected Provisioned condition to be set")
			assert.Equal(t, tc.wantProvisionedReason, cond.Reason)
			if tc.wantProvisionedStatus != "" {
				assert.Equal(t, tc.wantProvisionedStatus, cond.Status)
			}

			if tc.wantDetailsPreserved {
				_, hasPodSets := provReq.Status.ProvisioningClassDetails[conditions.SchedulablePodSetsDetailKey]
				_, hasPodCounts := provReq.Status.ProvisioningClassDetails[conditions.SchedulablePodCountsDetailKey]
				assert.True(t, hasPodSets, "expected schedulablePodSets detail to be preserved")
				assert.True(t, hasPodCounts, "expected schedulablePodCounts detail to be preserved")
			}

			if tc.wantSnapshotCommitted {
				// After commit the fork is consumed; another Revert should panic or be a no-op.
				// We just verify the method returned without error (commit succeeded).
			} else {
				// Snapshot was reverted inside the method; nothing extra to assert.
			}
		})
	}
}

func TestEvaluatePartialCapacity(t *testing.T) {
	now := time.Now()

	testCases := []struct {
		name                   string
		nodeMilliCPU           int64
		podSets                []v1.PodSet
		podCPUs                []int64
		wantSchedulablePodSets []string
		wantAllSchedulable     bool
	}{
		{
			name:         "all podsets fit",
			nodeMilliCPU: 1000,
			podSets: []v1.PodSet{
				{PodTemplateRef: v1.Reference{Name: "workers"}, Count: 2},
				{PodTemplateRef: v1.Reference{Name: "ps"}, Count: 1},
			},
			podCPUs:                []int64{100, 100, 100},
			wantSchedulablePodSets: []string{"workers", "ps"},
			wantAllSchedulable:     true,
		},
		{
			name:         "some podsets fit",
			nodeMilliCPU: 250,
			podSets: []v1.PodSet{
				{PodTemplateRef: v1.Reference{Name: "workers"}, Count: 2},
				{PodTemplateRef: v1.Reference{Name: "ps"}, Count: 1},
			},
			podCPUs:                []int64{100, 100, 200},
			wantSchedulablePodSets: []string{"workers"},
		},
		{
			name:         "none fit",
			nodeMilliCPU: 10,
			podSets: []v1.PodSet{
				{PodTemplateRef: v1.Reference{Name: "workers"}, Count: 1},
				{PodTemplateRef: v1.Reference{Name: "ps"}, Count: 1},
			},
			podCPUs:                []int64{100, 100},
			wantSchedulablePodSets: []string{},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			node := BuildTestNode("n1", tc.nodeMilliCPU, 1000000)
			SetNodeReadyState(node, true, now.Add(-2*time.Minute))

			snapshot := testsnapshot.NewTestSnapshotOrDie(t)
			clustersnapshot.InitializeClusterSnapshotOrDie(t, snapshot, []*apiv1.Node{node}, []*apiv1.Pod{})

			provReq := makeTestProvReqWithPodSets("test-pr", len(tc.podSets))
			provReq.Spec.PodSets = tc.podSets

			pods := buildAnnotatedPods("test-pr", tc.podSets, tc.podCPUs)

			o := &checkCapacityProvClass{
				autoscalingCtx: &ca_context.AutoscalingContext{
					ClusterSnapshot: snapshot,
				},
				schedulingSimulator: scheduling.NewHintingSimulator(),
			}

			snapshot.Fork()
			names, counts, err := o.evaluatePartialCapacity(pods, provReq)
			require.NoError(t, err)

			assert.Equal(t, tc.wantSchedulablePodSets, names)
			assert.NotNil(t, counts)
			for _, ps := range tc.podSets {
				_, ok := counts[ps.PodTemplateRef.Name]
				assert.True(t, ok, "expected count for podset %s", ps.PodTemplateRef.Name)
			}

			snapshot.Revert()
		})
	}
}

func TestCheckPartialCapacityEndToEnd(t *testing.T) {
	now := time.Now()

	testCases := []struct {
		name                   string
		nodeMilliCPU           int64
		podSets                []v1.PodSet
		podCPUs                []int64
		partialCapacityMode    string
		wantProvisionedReason  string
		wantProvisionedStatus  metav1.ConditionStatus
		wantScaleUpResult      status.ScaleUpResult
		wantSchedulablePodSets []string
		wantDetailPresent      bool
	}{
		{
			name:         "all podsets fit / bookPartial",
			nodeMilliCPU: 1000,
			podSets: []v1.PodSet{
				{PodTemplateRef: v1.Reference{Name: "workers"}, Count: 2},
				{PodTemplateRef: v1.Reference{Name: "ps"}, Count: 1},
			},
			podCPUs:                []int64{100, 100, 100},
			partialCapacityMode:    PartialCapacityCheckBookPartial,
			wantProvisionedReason:  conditions.CapacityIsFoundReason,
			wantProvisionedStatus:  metav1.ConditionTrue,
			wantScaleUpResult:      status.ScaleUpSuccessful,
			wantSchedulablePodSets: []string{"workers", "ps"},
			wantDetailPresent:      true,
		},
		{
			name:         "all podsets fit / checkOnly",
			nodeMilliCPU: 1000,
			podSets: []v1.PodSet{
				{PodTemplateRef: v1.Reference{Name: "workers"}, Count: 2},
				{PodTemplateRef: v1.Reference{Name: "ps"}, Count: 1},
			},
			podCPUs:                []int64{100, 100, 100},
			partialCapacityMode:    PartialCapacityCheckCheckOnly,
			wantProvisionedReason:  conditions.CapacityIsFoundReason,
			wantProvisionedStatus:  metav1.ConditionTrue,
			wantScaleUpResult:      status.ScaleUpSuccessful,
			wantSchedulablePodSets: []string{"workers", "ps"},
			wantDetailPresent:      true,
		},
		{
			name:         "some podsets fit / bookPartial",
			nodeMilliCPU: 250,
			podSets: []v1.PodSet{
				{PodTemplateRef: v1.Reference{Name: "workers"}, Count: 2},
				{PodTemplateRef: v1.Reference{Name: "ps"}, Count: 1},
			},
			podCPUs:                []int64{100, 100, 200},
			partialCapacityMode:    PartialCapacityCheckBookPartial,
			wantProvisionedReason:  conditions.PartialCapacityIsFoundReason,
			wantProvisionedStatus:  metav1.ConditionTrue,
			wantScaleUpResult:      status.ScaleUpSuccessful,
			wantSchedulablePodSets: []string{"workers"},
			wantDetailPresent:      true,
		},
		{
			name:         "some podsets fit / checkOnly",
			nodeMilliCPU: 250,
			podSets: []v1.PodSet{
				{PodTemplateRef: v1.Reference{Name: "workers"}, Count: 2},
				{PodTemplateRef: v1.Reference{Name: "ps"}, Count: 1},
			},
			podCPUs:                []int64{100, 100, 200},
			partialCapacityMode:    PartialCapacityCheckCheckOnly,
			wantProvisionedReason:  conditions.PartialCapacityIsFoundReason,
			wantProvisionedStatus:  metav1.ConditionFalse,
			wantScaleUpResult:      status.ScaleUpSuccessful,
			wantSchedulablePodSets: []string{"workers"},
			wantDetailPresent:      true,
		},
		{
			name:         "none fit / bookPartial",
			nodeMilliCPU: 10,
			podSets: []v1.PodSet{
				{PodTemplateRef: v1.Reference{Name: "workers"}, Count: 1},
				{PodTemplateRef: v1.Reference{Name: "ps"}, Count: 1},
			},
			podCPUs:               []int64{100, 100},
			partialCapacityMode:   PartialCapacityCheckBookPartial,
			wantProvisionedReason: conditions.CapacityIsNotFoundReason,
			wantScaleUpResult:     status.ScaleUpNoOptionsAvailable,
			// schedulablePodSets is preserved as [] so consumers can see per-PodSet counts.
			wantSchedulablePodSets: []string{},
			wantDetailPresent:      true,
		},
		{
			name:         "none fit / checkOnly",
			nodeMilliCPU: 10,
			podSets: []v1.PodSet{
				{PodTemplateRef: v1.Reference{Name: "workers"}, Count: 1},
				{PodTemplateRef: v1.Reference{Name: "ps"}, Count: 1},
			},
			podCPUs:               []int64{100, 100},
			partialCapacityMode:   PartialCapacityCheckCheckOnly,
			wantProvisionedReason: conditions.CapacityIsNotFoundReason,
			wantScaleUpResult:     status.ScaleUpNoOptionsAvailable,
			// schedulablePodSets is preserved as [] so consumers can see per-PodSet counts.
			wantSchedulablePodSets: []string{},
			wantDetailPresent:      true,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			node := BuildTestNode("n1", tc.nodeMilliCPU, 1000000)
			SetNodeReadyState(node, true, now.Add(-2*time.Minute))

			snapshot := testsnapshot.NewTestSnapshotOrDie(t)
			clustersnapshot.InitializeClusterSnapshotOrDie(t, snapshot, []*apiv1.Node{node}, []*apiv1.Pod{})

			provReq := makeTestProvReqWithPodSets("test-pr", len(tc.podSets))
			provReq.Spec.PodSets = tc.podSets
			provReq.Spec.Parameters = map[string]v1.Parameter{
				PartialCapacityCheckKey: v1.Parameter(tc.partialCapacityMode),
			}

			pods := buildAnnotatedPods("test-pr", tc.podSets, tc.podCPUs)

			combinedStatus := NewCombinedStatusSet()
			o := &checkCapacityProvClass{
				autoscalingCtx: &ca_context.AutoscalingContext{
					ClusterSnapshot: snapshot,
				},
				schedulingSimulator: scheduling.NewHintingSimulator(),
			}

			err := o.checkCapacity(pods, provReq, &combinedStatus)
			require.NoError(t, err)

			assert.Equal(t, tc.wantScaleUpResult, combinedStatus.Result)

			cond := findCondition(provReq, v1.Provisioned)
			if cond == nil {
				cond = findCondition(provReq, v1.Failed)
			}
			require.NotNil(t, cond, "expected a Provisioned or Failed condition")
			assert.Equal(t, tc.wantProvisionedReason, cond.Reason)
			if tc.wantProvisionedStatus != "" {
				assert.Equal(t, tc.wantProvisionedStatus, cond.Status)
			}

			if tc.wantDetailPresent {
				detail, ok := provReq.Status.ProvisioningClassDetails[conditions.SchedulablePodSetsDetailKey]
				require.True(t, ok, "expected schedulablePodSets detail to be present")
				var gotPodSets []string
				require.NoError(t, json.Unmarshal([]byte(detail), &gotPodSets))
				assert.Equal(t, tc.wantSchedulablePodSets, gotPodSets)
			}
		})
	}
}

// --- test helpers ---

func makeTestProvReq(name string) *provreqwrapper.ProvisioningRequest {
	return &provreqwrapper.ProvisioningRequest{
		ProvisioningRequest: &v1.ProvisioningRequest{
			ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: "default"},
			Spec: v1.ProvisioningRequestSpec{
				PodSets: []v1.PodSet{
					{PodTemplateRef: v1.Reference{Name: "workers"}, Count: 1},
				},
			},
			Status: v1.ProvisioningRequestStatus{
				Conditions: []metav1.Condition{},
			},
		},
	}
}

func makeTestProvReqWithPodSets(name string, n int) *provreqwrapper.ProvisioningRequest {
	podSets := make([]v1.PodSet, n)
	for i := 0; i < n; i++ {
		podSets[i] = v1.PodSet{
			PodTemplateRef: v1.Reference{Name: fmt.Sprintf("podset-%d", i)},
			Count:          1,
		}
	}
	return &provreqwrapper.ProvisioningRequest{
		ProvisioningRequest: &v1.ProvisioningRequest{
			ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: "default"},
			Spec: v1.ProvisioningRequestSpec{
				PodSets: podSets,
			},
			Status: v1.ProvisioningRequestStatus{
				Conditions: []metav1.Condition{},
			},
		},
	}
}

func buildAnnotatedPods(prName string, podSets []v1.PodSet, cpus []int64) []*apiv1.Pod {
	var pods []*apiv1.Pod
	cpuIdx := 0
	for i, ps := range podSets {
		for j := int32(0); j < ps.Count; j++ {
			cpu := int64(1)
			if cpuIdx < len(cpus) {
				cpu = cpus[cpuIdx]
				cpuIdx++
			}
			podName := fmt.Sprintf("%s-%d-%d", prName, i, j)
			pod := BuildTestPod(podName, cpu, 1, func(p *apiv1.Pod) {
				p.Annotations[v1.ProvisioningRequestPodAnnotationKey] = prName
			})
			pods = append(pods, pod)
		}
	}
	return pods
}

func findCondition(provReq *provreqwrapper.ProvisioningRequest, condType string) *metav1.Condition {
	for i := range provReq.Status.Conditions {
		if provReq.Status.Conditions[i].Type == condType {
			return &provReq.Status.Conditions[i]
		}
	}
	return nil
}
