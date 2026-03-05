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
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/autoscaler/cluster-autoscaler/processors/status"
	"k8s.io/autoscaler/cluster-autoscaler/utils/errors"
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
		{
			name:          "all partial capacity",
			statuses:      generateStatuses(2, status.ScaleUpPartialCapacityAvailable),
			exportedResut: status.ScaleUpPartialCapacityAvailable,
		},
		{
			name:          "successful and partial capacity",
			statuses:      append(generateStatuses(1, status.ScaleUpPartialCapacityAvailable), generateStatuses(1, status.ScaleUpSuccessful)...),
			exportedResut: status.ScaleUpSuccessful,
		},
		{
			name:          "partial capacity and no options available",
			statuses:      append(generateStatuses(1, status.ScaleUpPartialCapacityAvailable), generateStatuses(1, status.ScaleUpNoOptionsAvailable)...),
			exportedResut: status.ScaleUpPartialCapacityAvailable,
		},
		{
			name:          "error and partial capacity",
			statuses:      append(generateStatuses(1, status.ScaleUpError), generateStatuses(1, status.ScaleUpPartialCapacityAvailable)...),
			exportedResut: status.ScaleUpPartialCapacityAvailable,
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

func TestSortPodsFromProvReq(t *testing.T) {
	testCases := []struct {
		name               string
		input              []*apiv1.Pod
		expectedSortedPods []types.NamespacedName
	}{
		{
			name: "single PodSet with multiple pods",
			input: []*apiv1.Pod{
				{ObjectMeta: metav1.ObjectMeta{Name: "workload-0-2", Namespace: "default"}},
				{ObjectMeta: metav1.ObjectMeta{Name: "workload-0-0", Namespace: "default"}},
				{ObjectMeta: metav1.ObjectMeta{Name: "workload-0-1", Namespace: "default"}},
			},
			expectedSortedPods: []types.NamespacedName{
				{Namespace: "default", Name: "workload-0-0"},
				{Namespace: "default", Name: "workload-0-1"},
				{Namespace: "default", Name: "workload-0-2"},
			},
		},
		{
			name: "multiple PodSets",
			input: []*apiv1.Pod{
				{ObjectMeta: metav1.ObjectMeta{Name: "workload-1-0", Namespace: "default"}},
				{ObjectMeta: metav1.ObjectMeta{Name: "workload-0-1", Namespace: "default"}},
				{ObjectMeta: metav1.ObjectMeta{Name: "workload-2-0", Namespace: "default"}},
				{ObjectMeta: metav1.ObjectMeta{Name: "workload-0-0", Namespace: "default"}},
				{ObjectMeta: metav1.ObjectMeta{Name: "workload-1-1", Namespace: "default"}},
			},
			expectedSortedPods: []types.NamespacedName{
				{Namespace: "default", Name: "workload-0-0"},
				{Namespace: "default", Name: "workload-0-1"},
				{Namespace: "default", Name: "workload-1-0"},
				{Namespace: "default", Name: "workload-1-1"},
				{Namespace: "default", Name: "workload-2-0"},
			},
		},
		{
			name: "mixed with non-matching pattern - fallback to lexicographic",
			input: []*apiv1.Pod{
				{ObjectMeta: metav1.ObjectMeta{Name: "workload-0-1", Namespace: "default"}},
				{ObjectMeta: metav1.ObjectMeta{Name: "other-pod", Namespace: "default"}},
				{ObjectMeta: metav1.ObjectMeta{Name: "workload-0-0", Namespace: "default"}},
			},
			expectedSortedPods: []types.NamespacedName{
				{Namespace: "default", Name: "other-pod"},
				{Namespace: "default", Name: "workload-0-0"},
				{Namespace: "default", Name: "workload-0-1"},
			},
		},
		{
			name: "different namespaces with same indices - namespace used as tiebreaker",
			input: []*apiv1.Pod{
				{ObjectMeta: metav1.ObjectMeta{Name: "workload-0-0", Namespace: "ns-b"}},
				{ObjectMeta: metav1.ObjectMeta{Name: "workload-0-1", Namespace: "ns-a"}},
				{ObjectMeta: metav1.ObjectMeta{Name: "workload-0-0", Namespace: "ns-a"}},
			},
			expectedSortedPods: []types.NamespacedName{
				{Namespace: "ns-a", Name: "workload-0-0"},
				{Namespace: "ns-b", Name: "workload-0-0"},
				{Namespace: "ns-a", Name: "workload-0-1"},
			},
		},
		{
			name: "complex PodSet indices",
			input: []*apiv1.Pod{
				{ObjectMeta: metav1.ObjectMeta{Name: "app-10-5", Namespace: "default"}},
				{ObjectMeta: metav1.ObjectMeta{Name: "app-2-10", Namespace: "default"}},
				{ObjectMeta: metav1.ObjectMeta{Name: "app-2-2", Namespace: "default"}},
				{ObjectMeta: metav1.ObjectMeta{Name: "app-10-0", Namespace: "default"}},
			},
			expectedSortedPods: []types.NamespacedName{
				{Namespace: "default", Name: "app-2-2"},
				{Namespace: "default", Name: "app-2-10"},
				{Namespace: "default", Name: "app-10-0"},
				{Namespace: "default", Name: "app-10-5"},
			},
		},
		{
			name:               "empty list",
			input:              []*apiv1.Pod{},
			expectedSortedPods: []types.NamespacedName{},
		},
		{
			name: "single pod",
			input: []*apiv1.Pod{
				{ObjectMeta: metav1.ObjectMeta{Name: "workload-0-0", Namespace: "default"}},
			},
			expectedSortedPods: []types.NamespacedName{
				{Namespace: "default", Name: "workload-0-0"},
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			sorted := sortPodsFromProvReq(tc.input)

			sortedNamespacedNames := make([]types.NamespacedName, len(sorted))
			for i, pod := range sorted {
				sortedNamespacedNames[i] = types.NamespacedName{
					Namespace: pod.Namespace,
					Name:      pod.Name,
				}
			}

			assert.Equal(t, tc.expectedSortedPods, sortedNamespacedNames, "Pods should be sorted in the correct order")

			// Verify we didn't modify the number of pods
			assert.Equal(t, len(tc.input), len(sorted), "Should have same number of pods")
		})
	}
}
