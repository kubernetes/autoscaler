/*
Copyright 2017 The Kubernetes Authors.

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

package model

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var (
	testTimestamp, _ = time.Parse(TimeLayout, "2017-04-18 17:35:05")
	testPodID        = PodID{"namespace-1", "pod-1"}
	testContainerID  = ContainerID{testPodID, "container-1"}
)

func makeTestUsageSample() *ContainerUsageSampleWithKey {
	return &ContainerUsageSampleWithKey{ContainerUsageSample{testTimestamp, 1.0, 1.0}, testContainerID}
}

func TestClusterAddSample(t *testing.T) {
	// Create a pod with a single container.
	cluster := NewCluster()
	labels := make(map[string]string)
	cluster.AddOrUpdatePod(testPodID, labels)
	assert.NoError(t, cluster.AddOrUpdateContainer(testContainerID))

	// Add a usage sample to the container.
	cluster.AddSample(makeTestUsageSample())

	// Verify that the sample was aggregated into the container stats.
	containerStats := cluster.Pods[testPodID].Containers["container-1"]
	assert.Equal(t, testTimestamp, containerStats.lastSampleStart)
}

// Verifies that AddSample and AddOrUpdateContainer methods return a proper
// KeyError when refering to a non-existent pod.
func TestMissingKeys(t *testing.T) {
	cluster := NewCluster()
	err := cluster.AddSample(makeTestUsageSample())
	assert.EqualError(t, err, "KeyError: {namespace-1 pod-1}")

	err = cluster.AddOrUpdateContainer(testContainerID)
	assert.EqualError(t, err, "KeyError: {namespace-1 pod-1}")
}
