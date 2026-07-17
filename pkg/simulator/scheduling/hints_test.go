/*
Copyright 2022 The Kubernetes Authors.

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

package scheduling

import (
	"testing"
	"testing/quick"
)

func TestDropOld(t *testing.T) {
	f := func(initial []string, final []string) bool {
		all := chain(initial, final)
		s := NewHints()
		if incorrectKeys(s, []string{}, all) {
			return false
		}
		for _, k := range initial {
			s.Set(HintKey(k), k)
		}
		if incorrectKeys(s, initial, all) {
			return false
		}
		s.DropOld()
		if incorrectKeys(s, initial, all) {
			return false
		}
		for _, k := range final {
			s.Set(HintKey(k), k)
		}
		if incorrectKeys(s, all, all) {
			return false
		}
		s.DropOld()
		if incorrectKeys(s, final, all) {
			return false
		}
		return true
	}
	if err := quick.Check(f, nil); err != nil {
		t.Error(err)
	}
}

func chain(a, b []string) []string {
	return append(append([]string{}, a...), b...)
}

func anyPresent(s *Hints, keys map[string]bool) bool {
	for k := range keys {
		if _, found := s.Get(HintKey(k)); found {
			return true
		}
	}
	return false
}

func allPresentAndCorrect(s *Hints, keys []string) bool {
	for _, k := range keys {
		v, found := s.Get(HintKey(k))
		if !found {
			return false
		}
		if k != v {
			return false
		}
	}
	return true
}

func incorrectKeys(s *Hints, want, all []string) bool {
	dontWant := make(map[string]bool)
	for _, k := range all {
		dontWant[k] = true
	}
	for _, k := range want {
		delete(dontWant, k)
	}
	if !allPresentAndCorrect(s, want) {
		return true
	}
	return anyPresent(s, dontWant)
}
