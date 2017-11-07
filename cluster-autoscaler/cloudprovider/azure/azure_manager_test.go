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

package azure

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFixEndiannessUUID(t *testing.T) {
	var toFix = "60D7F925-4C67-DF44-A144-A3FE111ECDE3"
	var expected = ("25F9D760-674C-44DF-A144-A3FE111ECDE3")
	var result = fixEndiannessUUID(toFix)
	assert.Equal(t, result, expected)
}

func TestDoubleFixShouldProduceSameString(t *testing.T) {
	var toFix = "60D7F925-4C67-DF44-A144-A3FE111ECDE3"
	var tmp = fixEndiannessUUID(toFix)
	var result = fixEndiannessUUID(tmp)
	assert.Equal(t, result, toFix)
}

func TestFixEndiannessUUIDFailsOnInvalidUUID(t *testing.T) {
	assert.Panics(t, func() {
		var toFix = "60D7F925-4C67-DF44-A144-A3FE111ECDE3XXXX"
		_ = fixEndiannessUUID(toFix)
	}, "Calling with invalid UUID should panic")

}

func TestFixEndiannessUUIDFailsOnInvalidUUID2(t *testing.T) {
	assert.Panics(t, func() {
		var toFix = "60D7-F925-4C67-DF44-A144-A3FE-111E-CDE3-XXXX"
		_ = fixEndiannessUUID(toFix)
	}, "Calling with invalid UUID should panic")

}

func TestReverseBytes(t *testing.T) {
	assert.Equal(t, "CDAB", reverseBytes("ABCD"))
}
