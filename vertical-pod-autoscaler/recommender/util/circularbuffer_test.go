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

package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCircularBuffer(t *testing.T) {
	b := NewCircularBuffer(3)
	overflow, discarded := b.Push(1.0)
	assert.False(t, overflow)
	assert.Equal(t, 1.0, *b.Head())
	assert.Equal(t, []float64{1.0}, b.Contents())

	overflow, discarded = b.Push(2.0)
	assert.False(t, overflow)
	assert.Equal(t, 2.0, *b.Head())
	assert.Equal(t, []float64{1.0, 2.0}, b.Contents())

	overflow, discarded = b.Push(3.0)
	assert.False(t, overflow)
	assert.Equal(t, 3.0, *b.Head())
	assert.Equal(t, []float64{1.0, 2.0, 3.0}, b.Contents())

	overflow, discarded = b.Push(4.0)
	assert.True(t, overflow)
	assert.Equal(t, discarded, 1.0)
	assert.Equal(t, 4.0, *b.Head())
	assert.Equal(t, []float64{2.0, 3.0, 4.0}, b.Contents())
}
