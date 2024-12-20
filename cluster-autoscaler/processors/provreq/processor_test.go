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

package provreq

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/autoscaler/cluster-autoscaler/apis/provisioningrequest/autoscaling.x-k8s.io/v1"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	. "k8s.io/autoscaler/cluster-autoscaler/core/test"
	"k8s.io/autoscaler/cluster-autoscaler/provisioningrequest/conditions"
	"k8s.io/autoscaler/cluster-autoscaler/provisioningrequest/provreqclient"
	"k8s.io/autoscaler/cluster-autoscaler/provisioningrequest/provreqwrapper"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/clustersnapshot"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/framework"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/scheduling"
)

func TestRefresh(t *testing.T) {
	now := time.Now()
	dayAgo := now.Add(-1 * 24 * time.Hour)
	weekAgo := now.Add(-1 * defaultExpirationTime).Add(-1 * 5 * time.Minute)

	testCases := []struct {
		name           string
		creationTime   time.Time
		conditions     []metav1.Condition
		wantConditions []metav1.Condition
	}{
		{
			name:         "New ProvisioningRequest, empty conditions",
			creationTime: now,
		},
		{
			name:         "ProvisioningRequest with empty conditions, expired",
			creationTime: weekAgo,
			wantConditions: []metav1.Condition{
				{
					Type:               v1.Failed,
					Status:             metav1.ConditionTrue,
					LastTransitionTime: metav1.NewTime(now),
					Reason:             conditions.ExpiredReason,
					Message:            conditions.ExpiredMsg,
				},
			},
		},
		{
			name:         "ProvisioningRequest wasn't provisioned, expired",
			creationTime: weekAgo,
			conditions: []metav1.Condition{
				{
					Type:               v1.Provisioned,
					Status:             metav1.ConditionFalse,
					LastTransitionTime: metav1.NewTime(dayAgo),
					Reason:             conditions.ExpiredReason,
					Message:            conditions.ExpiredMsg,
				},
			},
			wantConditions: []metav1.Condition{
				{
					Type:               v1.Provisioned,
					Status:             metav1.ConditionFalse,
					LastTransitionTime: metav1.NewTime(dayAgo),
					Reason:             conditions.ExpiredReason,
					Message:            conditions.ExpiredMsg,
				},
				{
					Type:               v1.Failed,
					Status:             metav1.ConditionTrue,
					LastTransitionTime: metav1.NewTime(now),
					Reason:             conditions.ExpiredReason,
					Message:            conditions.ExpiredMsg,
				},
			},
		},
		{
			name:         "BookingCapacity time is expired ",
			creationTime: dayAgo,
			conditions: []metav1.Condition{
				{
					Type:               v1.Provisioned,
					Status:             metav1.ConditionTrue,
					LastTransitionTime: metav1.NewTime(dayAgo),
					Reason:             conditions.ExpiredReason,
					Message:            conditions.ExpiredMsg,
				},
			},
			wantConditions: []metav1.Condition{
				{
					Type:               v1.Provisioned,
					Status:             metav1.ConditionTrue,
					LastTransitionTime: metav1.NewTime(dayAgo),
					Reason:             conditions.ExpiredReason,
					Message:            conditions.ExpiredMsg,
				},
				{
					Type:               v1.BookingExpired,
					Status:             metav1.ConditionTrue,
					LastTransitionTime: metav1.NewTime(now),
					Reason:             conditions.CapacityReservationTimeExpiredReason,
					Message:            conditions.CapacityReservationTimeExpiredMsg,
				},
			},
		},
		{
			name:         "Failed ProvisioningRequest",
			creationTime: dayAgo,
			conditions: []metav1.Condition{
				{
					Type:               v1.Failed,
					Status:             metav1.ConditionTrue,
					LastTransitionTime: metav1.NewTime(dayAgo),
					Reason:             "Failed",
					Message:            "Failed",
				},
			},
			wantConditions: []metav1.Condition{
				{
					Type:               v1.Failed,
					Status:             metav1.ConditionTrue,
					LastTransitionTime: metav1.NewTime(dayAgo),
					Reason:             "Failed",
					Message:            "Failed",
				},
			},
		},
	}
	for _, test := range testCases {
		pr := provreqclient.ProvisioningRequestWrapperForTesting("namespace", "name-1")
		pr.Status.Conditions = test.conditions
		pr.CreationTimestamp = metav1.NewTime(test.creationTime)
		pr.Spec.ProvisioningClassName = v1.ProvisioningClassCheckCapacity

		additionalPr := provreqclient.ProvisioningRequestWrapperForTesting("namespace", "additional")
		additionalPr.CreationTimestamp = metav1.NewTime(weekAgo)
		additionalPr.Spec.ProvisioningClassName = v1.ProvisioningClassCheckCapacity

		processor := provReqProcessor{func() time.Time { return now }, 1, provreqclient.NewFakeProvisioningRequestClient(nil, t, pr, additionalPr), nil}
		processor.refresh([]*provreqwrapper.ProvisioningRequest{pr, additionalPr})

		assert.ElementsMatch(t, test.wantConditions, pr.Status.Conditions)
		if len(test.conditions) == len(test.wantConditions) {
			assert.ElementsMatch(t, []metav1.Condition{
				{
					Type:               v1.Failed,
					Status:             metav1.ConditionTrue,
					LastTransitionTime: metav1.NewTime(now),
					Reason:             conditions.ExpiredReason,
					Message:            conditions.ExpiredMsg,
				},
			}, additionalPr.Status.Conditions)
		} else {
			assert.ElementsMatch(t, []metav1.Condition{}, additionalPr.Status.Conditions)
		}
	}
}

func TestDeleteOldProvReqs(t *testing.T) {
	now := time.Now()
	tenDaysAgo := now.Add(-1 * 10 * 24 * time.Hour)
	pr := provreqclient.ProvisioningRequestWrapperForTesting("namespace", "name-1")
	additionalPr := provreqclient.ProvisioningRequestWrapperForTesting("namespace", "additional")

	oldFailedPr := provreqclient.ProvisioningRequestWrapperForTesting("namespace", "failed")
	oldExpiredPr := provreqclient.ProvisioningRequestWrapperForTesting("namespace", "expired")
	oldFailedPr.CreationTimestamp = metav1.NewTime(tenDaysAgo)
	oldExpiredPr.CreationTimestamp = metav1.NewTime(tenDaysAgo)
	oldFailedPr.Status.Conditions = []metav1.Condition{
		{
			Type:               v1.Failed,
			Status:             metav1.ConditionTrue,
			LastTransitionTime: metav1.NewTime(tenDaysAgo),
			Reason:             "Failed",
			Message:            "Failed",
		},
	}
	oldFailedPr.Spec.ProvisioningClassName = v1.ProvisioningClassCheckCapacity
	oldExpiredPr.Status.Conditions = []metav1.Condition{
		{
			Type:               v1.Provisioned,
			Status:             metav1.ConditionTrue,
			LastTransitionTime: metav1.NewTime(tenDaysAgo),
			Reason:             "Provisioned",
			Message:            "",
		},
		{
			Type:               v1.BookingExpired,
			Status:             metav1.ConditionTrue,
			LastTransitionTime: metav1.NewTime(tenDaysAgo),
			Reason:             "Capacity is expired",
			Message:            "",
		},
	}
	oldExpiredPr.Spec.ProvisioningClassName = v1.ProvisioningClassCheckCapacity

	client := provreqclient.NewFakeProvisioningRequestClient(nil, t, pr, additionalPr, oldFailedPr, oldExpiredPr)

	processor := provReqProcessor{func() time.Time { return now }, 1, client, nil}
	processor.refresh([]*provreqwrapper.ProvisioningRequest{pr, additionalPr, oldFailedPr, oldExpiredPr})

	_, err := client.ProvisioningRequestNoCache(oldFailedPr.Namespace, oldFailedPr.Name)
	assert.Error(t, err)
	_, err = client.ProvisioningRequestNoCache(oldExpiredPr.Namespace, oldExpiredPr.Name)
	assert.Error(t, err)
	_, err = client.ProvisioningRequestNoCache(pr.Namespace, pr.Name)
	assert.NoError(t, err)
	_, err = client.ProvisioningRequestNoCache(additionalPr.Namespace, additionalPr.Name)
	assert.NoError(t, err)
}

type fakeInjector struct {
	pods []*apiv1.Pod
}

func (f *fakeInjector) TrySchedulePods(clusterSnapshot clustersnapshot.ClusterSnapshot, pods []*apiv1.Pod, isNodeAcceptable func(*framework.NodeInfo) bool, breakOnFailure bool) ([]scheduling.Status, int, error) {
	f.pods = pods
	return nil, 0, nil
}

func TestBookCapacity(t *testing.T) {
	testCases := []struct {
		name             string
		conditions       []string
		provReq          *provreqwrapper.ProvisioningRequest
		capacityIsBooked bool
	}{
		{
			name:             "ProvReq is new, check-capacity class",
			provReq:          provreqwrapper.BuildTestProvisioningRequest("ns", "pr", "2", "100m", "", 10, false, time.Now(), v1.ProvisioningClassCheckCapacity),
			capacityIsBooked: false,
		},
		{
			name:             "ProvReq is Failed, best-effort-atomic class",
			conditions:       []string{v1.Failed},
			provReq:          provreqwrapper.BuildTestProvisioningRequest("ns", "pr", "2", "100m", "", 10, false, time.Now(), v1.ProvisioningClassBestEffortAtomicScaleUp),
			capacityIsBooked: false,
		},
		{
			name:             "ProvReq is Provisioned, unknown class",
			conditions:       []string{v1.Provisioned},
			provReq:          provreqwrapper.BuildTestProvisioningRequest("ns", "pr", "2", "100m", "", 10, false, time.Now(), "unknown"),
			capacityIsBooked: false,
		},
		{
			name:             "ProvReq is Provisioned, capacity should be booked, check-capacity class",
			conditions:       []string{v1.Provisioned},
			provReq:          provreqwrapper.BuildTestProvisioningRequest("ns", "pr", "2", "100m", "", 10, false, time.Now(), v1.ProvisioningClassCheckCapacity),
			capacityIsBooked: true,
		},
		{
			name:             "ProvReq is Provisioned, capacity should be booked, best-effort-atomic class",
			conditions:       []string{v1.Provisioned},
			provReq:          provreqwrapper.BuildTestProvisioningRequest("ns", "pr", "2", "100m", "", 10, false, time.Now(), v1.ProvisioningClassBestEffortAtomicScaleUp),
			capacityIsBooked: true,
		},
		{
			name:             "ProvReq has BookingExpired, capacity should not be booked, best-effort-atomic class",
			conditions:       []string{v1.Provisioned, v1.BookingExpired},
			provReq:          provreqwrapper.BuildTestProvisioningRequest("ns", "pr", "2", "100m", "", 10, false, time.Now(), v1.ProvisioningClassBestEffortAtomicScaleUp),
			capacityIsBooked: false,
		},
	}
	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			test := test
			injector := &fakeInjector{pods: []*apiv1.Pod{}}
			for _, condition := range test.conditions {
				conditions.AddOrUpdateCondition(test.provReq, condition, metav1.ConditionTrue, "", "", metav1.Now())
			}

			processor := &provReqProcessor{
				now:        func() time.Time { return time.Now() },
				client:     provreqclient.NewFakeProvisioningRequestClient(context.Background(), t, test.provReq),
				maxUpdated: 20,
				injector:   injector,
			}
			ctx, _ := NewScaleTestAutoscalingContext(config.AutoscalingOptions{}, nil, nil, nil, nil, nil)
			processor.bookCapacity(&ctx)
			if (test.capacityIsBooked && len(injector.pods) == 0) || (!test.capacityIsBooked && len(injector.pods) > 0) {
				t.Fail()
			}
		})
	}
}
