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

	apimeta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/autoscaler/cluster-autoscaler/apis/provisioningrequest/autoscaling.x-k8s.io/v1"
	"k8s.io/autoscaler/cluster-autoscaler/provisioningrequest/provreqclient"
	"k8s.io/autoscaler/cluster-autoscaler/provisioningrequest/provreqwrapper"
	clock "k8s.io/utils/clock/testing"
	"k8s.io/utils/lru"
)

func TestProvisioningRequestPodsInjector(t *testing.T) {
	now := time.Now()
	minAgo := now.Add(-1 * time.Minute).Add(-1 * time.Second)
	hourAgo := now.Add(-1 * time.Hour)

	accepted := metav1.Condition{
		Type:               v1.Accepted,
		Status:             metav1.ConditionTrue,
		LastTransitionTime: metav1.NewTime(minAgo),
	}
	failed := metav1.Condition{
		Type:               v1.Failed,
		Status:             metav1.ConditionTrue,
		LastTransitionTime: metav1.NewTime(hourAgo),
	}
	provisioned := metav1.Condition{
		Type:               v1.Provisioned,
		Status:             metav1.ConditionTrue,
		LastTransitionTime: metav1.NewTime(hourAgo),
	}
	notProvisioned := metav1.Condition{
		Type:               v1.Provisioned,
		Status:             metav1.ConditionFalse,
		LastTransitionTime: metav1.NewTime(hourAgo),
	}
	unknownProvisioned := metav1.Condition{
		Type:               v1.Provisioned,
		Status:             metav1.ConditionUnknown,
		LastTransitionTime: metav1.NewTime(hourAgo),
	}
	notProvisionedRecently := metav1.Condition{
		Type:               v1.Provisioned,
		Status:             metav1.ConditionFalse,
		LastTransitionTime: metav1.NewTime(minAgo),
	}

	podsA := 10
	newProvReqA := testProvisioningRequestWithCondition("new", podsA, v1.ProvisioningClassCheckCapacity)
	newAcceptedProvReqA := testProvisioningRequestWithCondition("new-accepted", podsA, v1.ProvisioningClassCheckCapacity, accepted)

	podsB := 20
	notProvisionedAcceptedProvReqB := testProvisioningRequestWithCondition("provisioned-false-B", podsB, v1.ProvisioningClassBestEffortAtomicScaleUp, notProvisioned, accepted)
	provisionedAcceptedProvReqB := testProvisioningRequestWithCondition("provisioned-and-accepted", podsB, v1.ProvisioningClassBestEffortAtomicScaleUp, provisioned, accepted)
	failedProvReq := testProvisioningRequestWithCondition("failed", podsA, v1.ProvisioningClassBestEffortAtomicScaleUp, failed)
	notProvisionedRecentlyProvReqB := testProvisioningRequestWithCondition("provisioned-false-recently-B", podsB, v1.ProvisioningClassBestEffortAtomicScaleUp, notProvisionedRecently)
	unknownProvisionedProvReqB := testProvisioningRequestWithCondition("provisioned-unknown-B", podsB, v1.ProvisioningClassBestEffortAtomicScaleUp, unknownProvisioned)
	unknownClass := testProvisioningRequestWithCondition("new-accepted", podsA, "unknown-class", accepted)

	testCases := []struct {
		name                             string
		provReqs                         []*provreqwrapper.ProvisioningRequest
		existingUnsUnschedulablePodCount int
		wantUnscheduledPodCount          int
		wantUpdatedConditionName         string
	}{
		{
			name:                     "New ProvisioningRequest, pods are injected and Accepted condition is added",
			provReqs:                 []*provreqwrapper.ProvisioningRequest{newProvReqA, provisionedAcceptedProvReqB},
			wantUnscheduledPodCount:  podsA,
			wantUpdatedConditionName: newProvReqA.Name,
		},
		{
			name:                     "New ProvisioningRequest, pods are injected and Accepted condition is updated",
			provReqs:                 []*provreqwrapper.ProvisioningRequest{newAcceptedProvReqA, provisionedAcceptedProvReqB},
			wantUnscheduledPodCount:  podsA,
			wantUpdatedConditionName: newAcceptedProvReqA.Name,
		},
		{
			name:                     "Provisioned=False, pods are injected",
			provReqs:                 []*provreqwrapper.ProvisioningRequest{notProvisionedAcceptedProvReqB, failedProvReq},
			wantUnscheduledPodCount:  podsB,
			wantUpdatedConditionName: notProvisionedAcceptedProvReqB.Name,
		},
		{
			name:     "Provisioned=True, no pods are injected",
			provReqs: []*provreqwrapper.ProvisioningRequest{provisionedAcceptedProvReqB, failedProvReq},
		},
		{
			name:     "Provisioned=False, ProvReq is backed off, no pods are injected",
			provReqs: []*provreqwrapper.ProvisioningRequest{notProvisionedRecentlyProvReqB},
		},
		{
			name:     "Provisioned=Unknown, no pods are injected",
			provReqs: []*provreqwrapper.ProvisioningRequest{unknownProvisionedProvReqB, failedProvReq},
		},
		{
			name:     "ProvisionedClass is unknown, no pods are injected",
			provReqs: []*provreqwrapper.ProvisioningRequest{unknownClass, failedProvReq},
		},
		{
			name:                             "Provisioned=False, pods are injected but unschedulable pod list is not overwriten",
			provReqs:                         []*provreqwrapper.ProvisioningRequest{newProvReqA},
			existingUnsUnschedulablePodCount: 50,
			wantUnscheduledPodCount:          podsA + 50,
			wantUpdatedConditionName:         newProvReqA.Name,
		},
	}
	for _, tc := range testCases {
		client := provreqclient.NewFakeProvisioningRequestClient(context.Background(), t, tc.provReqs...)
		backoffTime := lru.New(100)
		backoffTime.Add(key(notProvisionedRecentlyProvReqB), 2*time.Minute)
		injector := ProvisioningRequestPodsInjector{1 * time.Minute, 10 * time.Minute, backoffTime, clock.NewFakePassiveClock(now), client}
		getUnscheduledPods, err := injector.Process(nil, provreqwrapper.BuildTestPods("ns", "pod", tc.existingUnsUnschedulablePodCount))
		if err != nil {
			t.Errorf("%s failed: injector.Process return error %v", tc.name, err)
		}
		if len(getUnscheduledPods) != tc.wantUnscheduledPodCount {
			t.Errorf("%s failed: injector.Process return %d unscheduled pods, want %d", tc.name, len(getUnscheduledPods), tc.wantUnscheduledPodCount)
		}
		if tc.wantUpdatedConditionName == "" {
			continue
		}
		pr, _ := client.ProvisioningRequestNoCache("ns", tc.wantUpdatedConditionName)
		accepted := apimeta.FindStatusCondition(pr.Status.Conditions, v1.Accepted)
		if accepted == nil || accepted.LastTransitionTime != metav1.NewTime(now) {
			t.Errorf("%s: injector.Process hasn't update accepted condition for ProvisioningRequest %s", tc.name, tc.wantUpdatedConditionName)
		}
	}

}

func testProvisioningRequestWithCondition(name string, podCount int, class string, conditions ...metav1.Condition) *provreqwrapper.ProvisioningRequest {
	pr := provreqwrapper.BuildTestProvisioningRequest("ns", name, "10", "100", "", int32(podCount), false, time.Now(), class)
	pr.Status.Conditions = conditions
	return pr
}
