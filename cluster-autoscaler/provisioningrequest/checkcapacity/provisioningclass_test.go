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
	v1 "k8s.io/autoscaler/cluster-autoscaler/apis/provisioningrequest/autoscaling.x-k8s.io/v1"
	"k8s.io/autoscaler/cluster-autoscaler/processors/status"
	"k8s.io/autoscaler/cluster-autoscaler/provisioningrequest/provreqwrapper"
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
			name:    "empty pod list",
			pods:    []*apiv1.Pod{},
			provReq: makeProvReq("my-pr", v1.PodSet{PodTemplateRef: v1.Reference{Name: "workers"}, Count: 1}),
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
