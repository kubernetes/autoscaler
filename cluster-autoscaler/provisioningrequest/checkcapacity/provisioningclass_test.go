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
