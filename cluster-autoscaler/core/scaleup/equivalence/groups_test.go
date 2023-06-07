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

package equivalence

import (
	"fmt"
	"testing"

	. "k8s.io/autoscaler/cluster-autoscaler/utils/test"

	appsv1 "k8s.io/api/apps/v1"
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

	rc3 := apiv1.ReplicationController{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "rc3",
			Namespace: "default",
			SelfLink:  "api/v1/namespaces/default/replicationcontrollers/rc3",
			UID:       "12345678-1234-1234-1234-12345678901e",
		},
	}

	rc4 := apiv1.ReplicationController{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "rc4",
			Namespace: "default",
			SelfLink:  "api/v1/namespaces/default/replicationcontrollers/rc4",
			UID:       "12345678-1234-1234-1234-12345678901f",
		},
	}

	projectedSAVol := BuildServiceTokenProjectedVolumeSource("path")
	p1 := BuildTestPod("p1", 1500, 200000)
	p2_1 := BuildTestPod("p2_1", 3000, 200000)
	p2_1.OwnerReferences = GenerateOwnerReferences(rc1.Name, "ReplicationController", "extensions/v1beta1", rc1.UID)
	p2_2 := BuildTestPod("p2_2", 3000, 200000)
	p2_2.OwnerReferences = GenerateOwnerReferences(rc1.Name, "ReplicationController", "extensions/v1beta1", rc1.UID)
	p3_1 := BuildTestPod("p3_1", 100, 200000)
	p3_1.OwnerReferences = GenerateOwnerReferences(rc2.Name, "ReplicationController", "extensions/v1beta1", rc2.UID)
	p3_2 := BuildTestPod("p3_2", 100, 200000)
	p3_2.OwnerReferences = GenerateOwnerReferences(rc2.Name, "ReplicationController", "extensions/v1beta1", rc2.UID)
	// Two pods with projected volume sources should be in the same equivalence group
	p4_1 := BuildTestPod("p4_1", 100, 200000)
	p4_1.OwnerReferences = GenerateOwnerReferences(rc3.Name, "ReplicationController", "extensions/v1beta1", rc3.UID)
	p4_1.Spec.Volumes = []apiv1.Volume{{Name: "kube-api-access-nz94b", VolumeSource: apiv1.VolumeSource{Projected: projectedSAVol}}}
	p4_2 := BuildTestPod("p4_2", 100, 200000)
	p4_2.OwnerReferences = GenerateOwnerReferences(rc3.Name, "ReplicationController", "extensions/v1beta1", rc3.UID)
	p4_2.Spec.Volumes = []apiv1.Volume{{Name: "kube-api-access-mo25i", VolumeSource: apiv1.VolumeSource{Projected: projectedSAVol}}}
	// Two pods with flex volume sources should be in different equivalence groups
	p5_1 := BuildTestPod("p5_1", 100, 200000)
	p5_1.Spec.Volumes = []apiv1.Volume{{Name: "volume-nz94b", VolumeSource: apiv1.VolumeSource{FlexVolume: &apiv1.FlexVolumeSource{Driver: "testDriver"}}}}
	p5_1.OwnerReferences = GenerateOwnerReferences(rc4.Name, "ReplicationController", "extensions/v1beta1", rc4.UID)
	p5_2 := BuildTestPod("p5_2", 100, 200000)
	p5_2.Spec.Volumes = []apiv1.Volume{{Name: "volume-mo25i", VolumeSource: apiv1.VolumeSource{FlexVolume: &apiv1.FlexVolumeSource{Driver: "testDriver"}}}}
	p5_2.OwnerReferences = GenerateOwnerReferences(rc4.Name, "ReplicationController", "extensions/v1beta1", rc4.UID)
	unschedulablePods := []*apiv1.Pod{p1, p2_1, p2_2, p3_1, p3_2, p4_1, p4_2, p5_1, p5_2}

	podGroups := groupPodsBySchedulingProperties(unschedulablePods)
	assert.Equal(t, 6, len(podGroups))

	wantedGroups := []struct {
		pods  []*apiv1.Pod
		found bool
	}{
		{pods: []*apiv1.Pod{p1}},
		{pods: []*apiv1.Pod{p2_1, p2_2}},
		{pods: []*apiv1.Pod{p3_1, p3_2}},
		{pods: []*apiv1.Pod{p4_1, p4_2}},
		{pods: []*apiv1.Pod{p5_1}},
		{pods: []*apiv1.Pod{p5_2}},
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

func TestEquivalenceGroupSizeLimiting(t *testing.T) {
	rc := apiv1.ReplicationController{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "rc",
			Namespace: "default",
			SelfLink:  "api/v1/namespaces/default/replicationcontrollers/rc",
			UID:       "12345678-1234-1234-1234-123456789012",
		},
	}
	pods := make([]*apiv1.Pod, 0, maxEquivalenceGroupsByController+1)
	for i := 0; i < maxEquivalenceGroupsByController+1; i += 1 {
		p := BuildTestPod(fmt.Sprintf("p%d", i), 3000, 200000)
		p.OwnerReferences = GenerateOwnerReferences(rc.Name, "ReplicationController", "extensions/v1beta1", rc.UID)
		label := fmt.Sprintf("l%d", i)
		if i > maxEquivalenceGroupsByController {
			label = fmt.Sprintf("l%d", maxEquivalenceGroupsByController)
		}
		p.Labels = map[string]string{"uniqueLabel": label}
		pods = append(pods, p)
	}
	podGroups := groupPodsBySchedulingProperties(pods)
	assert.Equal(t, len(pods), len(podGroups))
	for i := range podGroups {
		assert.Equal(t, 1, len(podGroups[i]))
	}
}

func TestEquivalenceGroupIgnoresDaemonSets(t *testing.T) {
	ds := appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "ds",
			Namespace: "default",
			SelfLink:  "api/v1/namespaces/default/daemonsets/ds",
			UID:       "12345678-1234-1234-1234-123456789012",
		},
	}
	pods := make([]*apiv1.Pod, 2)
	pods[0] = BuildTestPod("p1", 3000, 200000)
	pods[0].OwnerReferences = GenerateOwnerReferences(ds.Name, "DaemonSet", "apps/v1", ds.UID)
	pods[1] = BuildTestPod("p2", 3000, 200000)
	pods[1].OwnerReferences = GenerateOwnerReferences(ds.Name, "DaemonSet", "apps/v1", ds.UID)
	podGroups := groupPodsBySchedulingProperties(pods)
	assert.Equal(t, 2, len(podGroups))
}
