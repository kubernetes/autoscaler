/*
Copyright 2023 The Kubernetes Authors.

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

package provreqclient

import (
	"context"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/autoscaler/cluster-autoscaler/apis/provisioningrequest/autoscaling.x-k8s.io/v1"
	"k8s.io/autoscaler/cluster-autoscaler/provisioningrequest/pods"
	"k8s.io/autoscaler/cluster-autoscaler/provisioningrequest/provreqwrapper"
	. "k8s.io/autoscaler/cluster-autoscaler/utils/test"
)

func TestFetchPodTemplates(t *testing.T) {
	pr1 := ProvisioningRequestWrapperForTesting("namespace", "name-1")
	pr2 := ProvisioningRequestWrapperForTesting("namespace", "name-2")
	mockProvisioningRequests := []*provreqwrapper.ProvisioningRequest{pr1, pr2}

	ctx := context.Background()
	c := NewFakeProvisioningRequestClient(ctx, t, mockProvisioningRequests...)
	got, err := c.FetchPodTemplates(pr1.ProvisioningRequest)
	if err != nil {
		t.Errorf("provisioningRequestClient.ProvisioningRequests() error: %v", err)
	}
	if len(got) != 1 {
		t.Errorf("provisioningRequestClient.ProvisioningRequests() got: %v, want 1 element", err)
	}
	if diff := cmp.Diff(pr1.PodTemplates, got); diff != "" {
		t.Errorf("Template mismatch, diff (-want +got):\n%s", diff)
	}
}

func TestProvisioningRequestsForPods(t *testing.T) {
	checkCapacityProvReq := provreqwrapper.BuildTestProvisioningRequest("ns", "check-capacity", "1m", "100", "", int32(100), false, time.Now(), v1.ProvisioningClassCheckCapacity)
	customProvReq := provreqwrapper.BuildTestProvisioningRequest("ns", "custom", "1m", "100", "", int32(100), false, time.Now(), "custom")
	checkCapacityPods, _ := pods.PodsForProvisioningRequest(checkCapacityProvReq)
	customProvReqPods, _ := pods.PodsForProvisioningRequest(customProvReq)
	regularPod := BuildTestPod("p1", 600, 100)
	client := NewFakeProvisioningRequestClient(context.Background(), t, checkCapacityProvReq, customProvReq)
	testCases := []struct {
		name string
		pods []*apiv1.Pod
		err  bool
		prs  []*provreqwrapper.ProvisioningRequest
	}{
		{
			name: "no pods",
			pods: []*apiv1.Pod{},
		},
		{
			name: "pods from one Provisioning Class",
			pods: checkCapacityPods,
			prs:  []*provreqwrapper.ProvisioningRequest{checkCapacityProvReq},
		},
		{
			name: "pods from different Provisioning Classes",
			pods: append(checkCapacityPods, customProvReqPods...),
			prs:  []*provreqwrapper.ProvisioningRequest{checkCapacityProvReq, customProvReq},
		},
		{
			name: "regular pod",
			pods: []*apiv1.Pod{regularPod},
		},
		{
			name: "provreq pods and regular pod",
			pods: append(checkCapacityPods, regularPod),
			prs:  []*provreqwrapper.ProvisioningRequest{checkCapacityProvReq},
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := ProvisioningRequestsForPods(client, tc.pods)
			assert.ElementsMatch(t, got, tc.prs)
		})
	}
}
