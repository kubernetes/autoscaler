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

type injectable struct {
	ClientToken *string
}

type uninjectable struct {
}

func TestSafeInjectClientToken(t *testing.T) {
	defer func() {
		if e := recover(); e != nil {
			t.Fatalf("panic on injecting client token: %+v", e)
		}
	}()

	injectableVal := new(injectable)
	safeInjectClientToken(injectableVal)
	if injectableVal.ClientToken == nil || len(*injectableVal.ClientToken) == 0 {
		t.Fatalf("no client token injected: %+v", injectableVal)
	}

	uninjectableVal := new(uninjectable)
	safeInjectClientToken(uninjectableVal)
}

var (
	exists = make(map[string]struct{})
)

func BenchmarkGenerateClientToken(b *testing.B) {
	for i := 0; i < b.N; i++ {
		token := randomClientToken()
		if _, conflict := exists[token]; conflict {
			b.Fatalf("conflict with generated token: %s, %d", token, i)
		}
		exists[token] = struct{}{}
	}
}
