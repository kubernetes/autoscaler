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

package profile

import (
	"math/rand"
	"testing"
	"time"
)

func TestExponentialBackoff(t *testing.T) {
	expected := []time.Duration{
		1 * time.Second,
		2 * time.Second,
		4 * time.Second,
		8 * time.Second,
		16 * time.Second,
		32 * time.Second,
		64 * time.Second,
		128 * time.Second,
	}
	for i := 0; i < len(expected); i++ {
		if ExponentialBackoff(i) != expected[i] {
			t.Fatalf("unexpected retry time, %+v expected, got %+v", expected[i], ExponentialBackoff(i))
		}
	}
}

func TestConstantDurationFunc(t *testing.T) {
	wanted := time.Duration(rand.Int()%100) * time.Second
	actual := ConstantDurationFunc(wanted)(rand.Int())
	if actual != wanted {
		t.Fatalf("unexpected retry time, %+v expected, got %+v", wanted, actual)
	}
}
