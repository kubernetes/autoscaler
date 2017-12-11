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

package gce

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetNormalizedGpuCount(t *testing.T) {
	gpus, err := getNormalizedGpuCount(int64(0))
	assert.Equal(t, err, nil)
	assert.Equal(t, gpus, int64(1))

	gpus, err = getNormalizedGpuCount(int64(1))
	assert.Equal(t, err, nil)
	assert.Equal(t, gpus, int64(1))

	gpus, err = getNormalizedGpuCount(int64(2))
	assert.Equal(t, err, nil)
	assert.Equal(t, gpus, int64(2))

	gpus, err = getNormalizedGpuCount(int64(3))
	assert.Equal(t, err, nil)
	assert.Equal(t, gpus, int64(4))

	gpus, err = getNormalizedGpuCount(int64(7))
	assert.Equal(t, err, nil)
	assert.Equal(t, gpus, int64(8))

	gpus, err = getNormalizedGpuCount(int64(8))
	assert.Equal(t, err, nil)
	assert.Equal(t, gpus, int64(8))

	gpus, err = getNormalizedGpuCount(int64(9))
	assert.Error(t, err)
}

func TestValidateGpuConfig(t *testing.T) {
	// valid configs
	err := validateGpuConfig("nvidia-tesla-k80", int64(1), "europe-west1-b", "n1-standard-1")
	assert.Equal(t, err, nil)
	err = validateGpuConfig("nvidia-tesla-p100", int64(1), "europe-west1-b", "n1-standard-1")
	assert.Equal(t, err, nil)
	err = validateGpuConfig("nvidia-tesla-k80", int64(4), "europe-west1-b", "n1-standard-32")
	assert.Equal(t, err, nil)

	// invalid gpu
	err = validateGpuConfig("duke-igthorn", int64(1), "europe-west1-b", "n1-standard-1")
	assert.Error(t, err)

	// invalid zone
	err = validateGpuConfig("nvidia-tesla-k80", int64(1), "castle-drekmore", "n1-standard-1")
	assert.Error(t, err)

	// invalid machine type
	err = validateGpuConfig("nvidia-tesla-k80", int64(1), "europe-west1-b", "toadie-the-ogre")
	assert.Error(t, err)

	// 1 gpu with large machine
	err = validateGpuConfig("nvidia-tesla-k80", int64(1), "europe-west1-b", "n1-standard-32")
	assert.Error(t, err)
}
