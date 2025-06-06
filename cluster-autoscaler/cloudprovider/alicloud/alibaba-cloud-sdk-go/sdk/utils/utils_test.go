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

package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFirstNotEmpty(t *testing.T) {
	// Test case where the first non-empty string is at the beginning
	result := FirstNotEmpty("hello", "world", "test")
	assert.Equal(t, "hello", result)

	// Test case where the first non-empty string is in the middle
	result = FirstNotEmpty("", "foo", "bar")
	assert.Equal(t, "foo", result)

	// Test case where the first non-empty string is at the end
	result = FirstNotEmpty("", "", "baz")
	assert.Equal(t, "baz", result)

	// Test case where all strings are empty
	result = FirstNotEmpty("", "", "")
	assert.Equal(t, "", result)

	// Test case with no arguments
	result = FirstNotEmpty()
	assert.Equal(t, "", result)
}
