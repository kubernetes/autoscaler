/*
Copyright 2020 The Kubernetes Authors.

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

package aws

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSmallestInstanceTypeSameFamily(t *testing.T) {
	instanceTypes := []string{
		"c5.large",
		"c5.xlarge",
	}
	smallest := smallestInstanceType(instanceTypes)
	assert.Equal(t, smallest, "c5.large")
}

func TestSmallestInstanceTypeSameCPUDifferentMem(t *testing.T) {
	instanceTypes := []string{
		"c4.large",
		"c5.large",
	}
	smallest := smallestInstanceType(instanceTypes)
	assert.Equal(t, smallest, "c4.large")
}

func TestSmallestInstanceTypeSameMemDifferentCPU(t *testing.T) {
	instanceTypes := []string{
		"c5.xlarge",
		"m4.large",
	}
	smallest := smallestInstanceType(instanceTypes)
	assert.Equal(t, smallest, "m4.large")
}

func TestSmallestInstanceTypeSameFamilyDifferentProcessor(t *testing.T) {
	{
		instanceTypes := []string{
			"m5a.large",
			"m5.large",
		}
		smallest := smallestInstanceType(instanceTypes)
		assert.Equal(t, smallest, instanceTypes[0])
	}
	{

		instanceTypes := []string{
			"m5.large",
			"m5a.large",
		}
		smallest := smallestInstanceType(instanceTypes)
		assert.Equal(t, smallest, instanceTypes[0])
	}
}

func TestSmallestInstanceEqualButHasStorage(t *testing.T) {
	{
		instanceTypes := []string{
			"c5d.large",
			"c5.large",
		}
		smallest := smallestInstanceType(instanceTypes)
		assert.Equal(t, smallest, instanceTypes[0])
	}

	{
		instanceTypes := []string{
			"c5.large",
			"c5d.large",
		}
		smallest := smallestInstanceType(instanceTypes)
		assert.Equal(t, smallest, instanceTypes[0])
	}
}

func TestSmallestInstanceTypeCompareGPU(t *testing.T) {
	instanceTypes := []string{
		"p3.2xlarge",
		"r3.2xlarge",
	}
	smallest := smallestInstanceType(instanceTypes)
	assert.Equal(t, smallest, "r3.2xlarge")
}
