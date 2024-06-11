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

	v1 "k8s.io/api/core/v1"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/autoscaler/cluster-autoscaler/apis/provisioningrequest/autoscaling.x-k8s.io/v1beta1"
	"k8s.io/autoscaler/cluster-autoscaler/provisioningrequest/provreqclient"
	"k8s.io/autoscaler/cluster-autoscaler/provisioningrequest/provreqwrapper"
	clock "k8s.io/utils/clock/testing"
)

func TestProvisioningRequestPodsInjector(t *testing.T) {
	now := time.Now()
	minAgo := now.Add(-1 * time.Minute)
	hourAgo := now.Add(-1 * time.Hour)

	accepted := metav1.Condition{
		Type:               v1beta1.Accepted,
		Status:             metav1.ConditionTrue,
		LastTransitionTime: metav1.NewTime(minAgo),
	}
	failed := metav1.Condition{
		Type:               v1beta1.Failed,
		Status:             metav1.ConditionTrue,
		LastTransitionTime: metav1.NewTime(hourAgo),
	}
	provisioned := metav1.Condition{
		Type:               v1beta1.Provisioned,
		Status:             metav1.ConditionTrue,
		LastTransitionTime: metav1.NewTime(hourAgo),
	}
	notProvisioned := metav1.Condition{
		Type:               v1beta1.Provisioned,
		Status:             metav1.ConditionFalse,
		LastTransitionTime: metav1.NewTime(hourAgo),
	}
	unknownProvisioned := metav1.Condition{
		Type:               v1beta1.Provisioned,
		Status:             metav1.ConditionUnknown,
		LastTransitionTime: metav1.NewTime(hourAgo),
	}
	notProvisionedRecently := metav1.Condition{
		Type:               v1beta1.Provisioned,
		Status:             metav1.ConditionFalse,
		LastTransitionTime: metav1.NewTime(minAgo),
	}

	podsA := 10
	newProvReqA := testProvisioningRequestWithCondition("new", podsA)
	newAcceptedProvReqA := testProvisioningRequestWithCondition("new-accepted", podsA, accepted)

	podsB := 20
	notProvisionedAcceptedProvReqB := testProvisioningRequestWithCondition("provisioned-false-B", podsB, notProvisioned, accepted)
	provisionedAcceptedProvReqB := testProvisioningRequestWithCondition("provisioned-and-accepted", podsB, provisioned, accepted)
	failedProvReq := testProvisioningRequestWithCondition("failed", podsA, failed)
	notProvisionedRecentlyProvReqB := testProvisioningRequestWithCondition("provisioned-false-recently-B", podsB, notProvisionedRecently)
	unknownProvisionedProvReqB := testProvisioningRequestWithCondition("provisioned-unknown-B", podsB, unknownProvisioned)

	testCases := []struct {
		name                     string
		provReqs                 []*provreqwrapper.ProvisioningRequest
		wantUnscheduledPodCount  int
		wantUpdatedConditionName string
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
			provReqs: []*provreqwrapper.ProvisioningRequest{provisionedAcceptedProvReqB, failedProvReq, notProvisionedRecentlyProvReqB},
		},
		{
			name:     "Provisioned=Unknown, no pods are injected",
			provReqs: []*provreqwrapper.ProvisioningRequest{unknownProvisionedProvReqB, failedProvReq, notProvisionedRecentlyProvReqB},
		},
	}
	for _, tc := range testCases {
		client := provreqclient.NewFakeProvisioningRequestClient(context.Background(), t, tc.provReqs...)
		injector := ProvisioningRequestPodsInjector{client, clock.NewFakePassiveClock(now)}
		getUnscheduledPods, err := injector.Process(nil, []*v1.Pod{})
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
		accepted := apimeta.FindStatusCondition(pr.Status.Conditions, v1beta1.Accepted)
		if accepted == nil || accepted.LastTransitionTime != metav1.NewTime(now) {
			t.Errorf("%s: injector.Process hasn't update accepted condition for ProvisioningRequest %s", tc.name, tc.wantUpdatedConditionName)
		}
	}

}

func testProvisioningRequestWithCondition(name string, podCount int, conditions ...metav1.Condition) *provreqwrapper.ProvisioningRequest {
	pr := provreqwrapper.BuildTestProvisioningRequest("ns", name, "10", "100", "", int32(podCount), false, time.Now(), "ProvisioningClass")
	pr.Status.Conditions = conditions
	return pr
}
