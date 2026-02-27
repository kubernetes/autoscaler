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

package spec

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetPodSpecsReturnsNoResults(t *testing.T) {
	// given
	tc := newEmptySpecClientTestCase()
	client := tc.createFakeSpecClient()

	// when
	podSpecs, err := client.GetPodSpecs()

	// then
	assert.NoError(t, err)
	assert.Empty(t, podSpecs)
}

func TestGetPodSpecsReturnsSpecs(t *testing.T) {
	// given
	tc := newSpecClientTestCase()
	client := tc.createFakeSpecClient()

	// when
	podSpecs, err := client.GetPodSpecs()

	// then
	assert.NoError(t, err)
	assert.Equal(t, len(tc.podSpecs), len(podSpecs), "SpecClient returned different number of results then expected")
	for _, podSpec := range podSpecs {
		assert.Contains(t, tc.podSpecs, podSpec, "One of returned BasicPodSpec is different than expected")
	}
}

func TestGetPodSpecsNativeSidecar(t *testing.T) {
	tc := newNativeSidecarSpecClientTestCase()
	client := tc.createFakeSpecClient()

	podSpecs, err := client.GetPodSpecs()
	assert.NoError(t, err)
	if !assert.Len(t, podSpecs, 1) {
		return
	}
	pod := podSpecs[0]
	// The native sidecar init container should be marked
	if assert.Len(t, pod.InitContainers, 2, "expected 2 init containers") {
		assert.True(t, pod.InitContainers[0].IsNativeSidecar, "native sidecar should have IsNativeSidecar=true")
		assert.False(t, pod.InitContainers[1].IsNativeSidecar, "regular init container should have IsNativeSidecar=false")
	}
	// Regular containers should not be marked as native sidecar
	for _, c := range pod.Containers {
		assert.False(t, c.IsNativeSidecar, "regular container %q should have IsNativeSidecar=false", c.ID.ContainerName)
	}
}
