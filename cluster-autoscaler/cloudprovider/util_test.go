/*
Copyright 2016 The Kubernetes Authors.

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

package cloudprovider

import (
	"testing"

	"github.com/stretchr/testify/assert"
	apiv1 "k8s.io/api/core/v1"
)

func TestBuildReadyConditions(t *testing.T) {
	conditions := BuildReadyConditions()
	foundReady := false
	for _, condition := range conditions {
		if condition.Type == apiv1.NodeReady && condition.Status == apiv1.ConditionTrue {
			foundReady = true
		}
	}
	assert.True(t, foundReady)
}

func TestBuildKubeProxy(t *testing.T) {

	pod := BuildKubeProxy("kube-proxy")
	assert.NotNil(t, pod)
	assert.Equal(t, 1, len(pod.Spec.Containers))
	cpu := pod.Spec.Containers[0].Resources.Requests[apiv1.ResourceCPU]
	assert.Equal(t, int64(100), cpu.MilliValue())
}

func TestJoinStringMaps(t *testing.T) {
	map1 := map[string]string{"1": "a", "2": "b"}
	map2 := map[string]string{"3": "c", "2": "d"}
	map3 := map[string]string{"5": "e"}
	result := JoinStringMaps(map1, map2, map3)
	assert.Equal(t, map[string]string{"1": "a", "2": "d", "3": "c", "5": "e"}, result)
}
