/*
Copyright 2019 The Kubernetes Authors.

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

package core

import (
	"fmt"
	"testing"

	. "k8s.io/autoscaler/cluster-autoscaler/utils/test"

	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/stretchr/testify/assert"
)

func TestGroupSchedulablePodsForNode(t *testing.T) {
	rc1 := apiv1.ReplicationController{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "rc1",
			Namespace: "default",
			SelfLink:  "api/v1/namespaces/default/replicationcontrollers/rc1",
			UID:       "12345678-1234-1234-1234-123456789012",
		},
	}

	rc2 := apiv1.ReplicationController{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "rc2",
			Namespace: "default",
			SelfLink:  "api/v1/namespaces/default/replicationcontrollers/rc2",
			UID:       "12345678-1234-1234-1234-12345678901a",
		},
	}

	p1 := BuildTestPod("p1", 1500, 200000)
	p2_1 := BuildTestPod("p2_1", 3000, 200000)
	p2_1.OwnerReferences = GenerateOwnerReferences(rc1.Name, "ReplicationController", "extensions/v1beta1", rc1.UID)
	p2_2 := BuildTestPod("p2_2", 3000, 200000)
	p2_2.OwnerReferences = GenerateOwnerReferences(rc1.Name, "ReplicationController", "extensions/v1beta1", rc1.UID)
	p3_1 := BuildTestPod("p3_1", 100, 200000)
	p3_1.OwnerReferences = GenerateOwnerReferences(rc2.Name, "ReplicationController", "extensions/v1beta1", rc2.UID)
	p3_2 := BuildTestPod("p3_2", 100, 200000)
	p3_2.OwnerReferences = GenerateOwnerReferences(rc2.Name, "ReplicationController", "extensions/v1beta1", rc2.UID)
	unschedulablePods := []*apiv1.Pod{p1, p2_1, p2_2, p3_1, p3_2}

	podGroups := groupPodsBySchedulingProperties(unschedulablePods)
	assert.Equal(t, 3, len(podGroups))

	wantedGroups := []struct {
		pods  []*apiv1.Pod
		found bool
	}{
		{pods: []*apiv1.Pod{p1}},
		{pods: []*apiv1.Pod{p2_1, p2_2}},
		{pods: []*apiv1.Pod{p3_1, p3_2}},
	}

	equal := func(a, b []*apiv1.Pod) bool {
		if len(a) != len(b) {
			return false
		}
		ma := map[*apiv1.Pod]bool{}
		for _, ea := range a {
			ma[ea] = true
		}
		for _, eb := range b {
			if !ma[eb] {
				return false
			}
		}
		return true
	}

	for _, g := range podGroups {
		found := false
		for i, wanted := range wantedGroups {
			if equal(g, wanted.pods) {
				wanted.found = true
				wantedGroups[i] = wanted
				found = true
			}
		}
		assert.True(t, found, fmt.Errorf("Unexpected pod group: %+v", g))
	}

	for _, w := range wantedGroups {
		assert.True(t, w.found, fmt.Errorf("Expected pod group: %+v", w))
	}
}
