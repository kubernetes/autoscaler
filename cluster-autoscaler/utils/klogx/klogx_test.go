/*
Copyright 2018 The Kubernetes Authors.

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

package klogx

import (
	"testing"

	"github.com/stretchr/testify/assert"
	klog "k8s.io/klog/v2"
)

// Left, UpTo and Over should work as expected.
func TestLoggingQuota(t *testing.T) {
	q := NewLoggingQuota(3)

	for i := 0; i < 5; i++ {
		assert.Equal(t, 3-i, q.Left())
		assert.Equal(t, i < 3, V(0).UpTo(q).enabled)
		assert.Equal(t, i >= 3, V(0).Over(q).enabled)
	}
}

// Reset should restore left to the original limit.
func TestReset(t *testing.T) {
	q := NewLoggingQuota(3)

	for i := 0; i < 5; i++ {
		assert.Equal(t, i < 3, V(0).UpTo(q).enabled)
	}

	q.Reset()

	assert.Equal(t, 3, q.Left())
	assert.False(t, V(0).Over(q).enabled)
	assert.True(t, V(0).UpTo(q).enabled)
}

// Tests that quota isn't used up by calls limited by verbosity.
func TestVFalse(t *testing.T) {
	// XXX: this is a hack to get a disabled V, since klog v2 no longer
	// provides an easy way to create it
	v := V(10000)
	q := NewLoggingQuota(3)

	assert.False(t, v.UpTo(q).v.Enabled())
	assert.Equal(t, 3, q.Left())
}

// Tests that V limits calls based on verbosity the same way as klog.V.
func TestV(t *testing.T) {
	for i := klog.Level(0); i <= 10; i++ {
		assert.Equal(t, klog.V(i).Enabled(), V(i).v.Enabled())
	}
}
