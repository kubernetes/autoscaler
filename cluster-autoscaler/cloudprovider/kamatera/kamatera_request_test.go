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

package kamatera

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestKamateraHTTPClientTimeout(t *testing.T) {
	assert.NotNil(t, kamateraHTTPClient, "kamateraHTTPClient should not be nil")
	assert.Equal(t, 5*time.Minute, kamateraHTTPClient.Timeout, "HTTP client timeout should be 5 minutes")
}
