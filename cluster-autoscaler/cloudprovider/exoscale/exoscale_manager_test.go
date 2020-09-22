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

package exoscale

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewManager(t *testing.T) {
	os.Setenv("EXOSCALE_API_KEY", "KEY")
	os.Setenv("EXOSCALE_API_SECRET", "SECRET")
	os.Setenv("EXOSCALE_API_ENDPOINT", "url")

	manager, err := newManager()
	assert.NoError(t, err)
	assert.NotNil(t, manager)
}

func TestNewManagerFailure(t *testing.T) {
	os.Unsetenv("EXOSCALE_API_KEY")
	os.Unsetenv("EXOSCALE_API_SECRET")
	os.Setenv("EXOSCALE_API_ENDPOINT", "url")

	manager, err := newManager()
	assert.Error(t, err)
	assert.Nil(t, manager)
}

func TestComputeInstanceLimit(t *testing.T) {
	ts := newTestServer(
		testHTTPResponse{200, testMockResourceLimit},
	)

	os.Setenv("EXOSCALE_API_KEY", "KEY")
	os.Setenv("EXOSCALE_API_SECRET", "SECRET")
	os.Setenv("EXOSCALE_API_ENDPOINT", ts.URL)

	manager, err := newManager()
	assert.NoError(t, err)
	assert.NotNil(t, manager)

	limit, err := manager.computeInstanceLimit()
	assert.NoError(t, err)
	assert.Equal(t, testMockResourceLimitMax, limit)
}
