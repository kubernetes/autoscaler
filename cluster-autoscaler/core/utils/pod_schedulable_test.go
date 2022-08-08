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

package utils

import (
	"fmt"
	"k8s.io/autoscaler/cluster-autoscaler/simulator"
	"testing"

	. "k8s.io/autoscaler/cluster-autoscaler/utils/test"

	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/stretchr/testify/assert"
)

func TestPodSchedulableMap(t *testing.T) {
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

	pMap := NewPodSchedulableMap()

	podInRc1_1 := BuildTestPod("podInRc1_1", 500, 1000)
	podInRc1_1.OwnerReferences = GenerateOwnerReferences(rc1.Name, "ReplicationController", "extensions/v1beta1", rc1.UID)

	podInRc2 := BuildTestPod("podInRc2", 500, 1000)
	podInRc2.OwnerReferences = GenerateOwnerReferences(rc2.Name, "ReplicationController", "extensions/v1beta1", rc2.UID)

	// Basic sanity checks
	_, found := pMap.Get(podInRc1_1)
	assert.False(t, found)
	pMap.Set(podInRc1_1, nil)
	err, found := pMap.Get(podInRc1_1)
	assert.True(t, found)
	assert.Nil(t, err)

	cpuErr := simulator.GenericPredicateError()

	// Pod in different RC
	_, found = pMap.Get(podInRc2)
	assert.False(t, found)
	pMap.Set(podInRc2, cpuErr)
	err, found = pMap.Get(podInRc2)
	assert.True(t, found)
	assert.Equal(t, cpuErr, err)

	// Another replica in rc1
	podInRc1_2 := BuildTestPod("podInRc1_2", 500, 1000)
	podInRc1_2.OwnerReferences = GenerateOwnerReferences(rc1.Name, "ReplicationController", "extensions/v1beta1", rc1.UID)
	err, found = pMap.Get(podInRc1_2)
	assert.True(t, found)
	assert.Nil(t, err)

	// A replica in rc1 with a projected volume
	podInRc1ProjectedVol := BuildTestPod("podInRc1_ProjectedVol", 500, 1000)
	podInRc1ProjectedVol.OwnerReferences = GenerateOwnerReferences(rc1.Name, "ReplicationController", "extensions/v1beta1", rc1.UID)
	podInRc1ProjectedVol.Spec.Volumes = []apiv1.Volume{{Name: "kube-api-access-nz94b", VolumeSource: apiv1.VolumeSource{Projected: BuildServiceTokenProjectedVolumeSource("path")}}}
	err, found = pMap.Get(podInRc1ProjectedVol)
	assert.True(t, found)
	assert.Nil(t, err)

	// A replica in rc1 with a non-projected volume
	podInRc1FlexVol := BuildTestPod("podInRc1_FlexVol", 500, 1000)
	podInRc1FlexVol.OwnerReferences = GenerateOwnerReferences(rc1.Name, "ReplicationController", "extensions/v1beta1", rc1.UID)
	podInRc1FlexVol.Spec.Volumes = []apiv1.Volume{{Name: "volume-mo25i", VolumeSource: apiv1.VolumeSource{FlexVolume: &apiv1.FlexVolumeSource{Driver: "testDriver"}}}}
	err, found = pMap.Get(podInRc1FlexVol)
	assert.False(t, found)
	assert.Nil(t, err)

	// A pod in rc1, but with different requests
	differentPodInRc1 := BuildTestPod("differentPodInRc1", 1000, 1000)
	differentPodInRc1.OwnerReferences = GenerateOwnerReferences(rc1.Name, "ReplicationController", "extensions/v1beta1", rc1.UID)
	_, found = pMap.Get(differentPodInRc1)
	assert.False(t, found)
	pMap.Set(differentPodInRc1, cpuErr)
	err, found = pMap.Get(differentPodInRc1)
	assert.True(t, found)
	assert.Equal(t, cpuErr, err)

	// A non-replicated pod
	nonReplicatedPod := BuildTestPod("nonReplicatedPod", 1000, 1000)
	_, found = pMap.Get(nonReplicatedPod)
	assert.False(t, found)
	pMap.Set(nonReplicatedPod, err)
	_, found = pMap.Get(nonReplicatedPod)
	assert.False(t, found)

	// Verify information about first pod has not been overwritten by adding
	// other pods
	err, found = pMap.Get(podInRc1_1)
	assert.True(t, found)
	assert.Nil(t, err)
}

func TestPodSchedulableMapSizeLimiting(t *testing.T) {
	rc := apiv1.ReplicationController{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "rc",
			Namespace: "default",
			SelfLink:  "api/v1/namespaces/default/replicationcontrollers/rc",
			UID:       "12345678-1234-1234-1234-123456789012",
		},
	}
	pMap := NewPodSchedulableMap()
	pods := make([]*apiv1.Pod, 0, maxPodsPerOwnerRef+1)
	for i := 0; i < maxPodsPerOwnerRef+1; i += 1 {
		p := BuildTestPod(fmt.Sprintf("p%d", i), 3000, 200000)
		p.OwnerReferences = GenerateOwnerReferences(rc.Name, "ReplicationController", "extensions/v1beta1", rc.UID)
		p.Labels = map[string]string{"uniqueLabel": fmt.Sprintf("l%d", i)}
		pods = append(pods, p)
		_, found := pMap.Get(p)
		assert.False(t, found)
	}
	for _, p := range pods {
		pMap.Set(p, nil)
	}
	for i, p := range pods {
		_, found := pMap.Get(p)
		if i != len(pods)-1 {
			assert.True(t, found)
		} else {
			assert.False(t, found)
		}
	}
	assert.Equal(t, 1, pMap.OverflowingControllerCount())
}

func TestPodSchedulableMapIgnoreDaemonSets(t *testing.T) {
	ds := appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "ds",
			Namespace: "default",
			SelfLink:  "api/v1/namespaces/default/daemonsets/ds",
			UID:       "12345678-1234-1234-1234-123456789012",
		},
	}
	pMap := NewPodSchedulableMap()
	pod := BuildTestPod("pod", 3000, 200000)
	pod.OwnerReferences = GenerateOwnerReferences(ds.Name, "DaemonSet", "apps/v1", ds.UID)
	pMap.Set(pod, nil)
	_, found := pMap.Get(pod)
	assert.False(t, found)
}
