/*
Copyright 2021 The Kubernetes Authors.

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

package common

import (
	"testing"
)

func TestStringPtrsValues(t *testing.T) {
	vals := []string{"a", "b", "c", "d"}
	ptrs := StringPtrs(vals)
	for i := 0; i < len(vals); i++ {
		if *ptrs[i] != vals[i] {
			t.Errorf("[ERROR] value %s != ptr value %s", vals[i], *ptrs[i])
		}
	}
	newVals := StringValues(ptrs)
	for i := 0; i < len(vals); i++ {
		if newVals[i] != vals[i] {
			t.Errorf("[ERROR] new val %s != val %s", newVals[i], vals[i])
		}
	}
}
