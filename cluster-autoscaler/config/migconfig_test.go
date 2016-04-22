/*
Copyright 2016 The Kubernetes Authors All rights reserved.

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

package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSet(t *testing.T) {
	migConfigFlag := MigConfigFlag{}
	assert.Error(t, migConfigFlag.Set("a"))
	assert.Error(t, migConfigFlag.Set("a:b:c"))
	assert.Error(t, migConfigFlag.Set("1:2:x"))
	assert.Error(t, migConfigFlag.Set("1:2:"))
	assert.NoError(t, migConfigFlag.Set("111:222:https://content.googleapis.com/compute/v1/projects/test-project/zones/test-zone/instanceGroups/test-name"))
	assert.Equal(t, 111, migConfigFlag[0].MinSize)
	assert.Equal(t, 222, migConfigFlag[0].MaxSize)
	assert.Equal(t, "test-zone", migConfigFlag[0].Zone)
	assert.Equal(t, "test-name", migConfigFlag[0].Name)
	assert.Contains(t, migConfigFlag.String(), "111")
	assert.Contains(t, migConfigFlag.String(), "222")
	assert.Contains(t, migConfigFlag.String(), "test-zone")
	assert.Contains(t, migConfigFlag.String(), "test-name")
}
