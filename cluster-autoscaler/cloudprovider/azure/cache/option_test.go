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

package cache

import (
	"testing"

	. "github.com/onsi/gomega"

	"github.com/go-logr/logr"
)

func Test_WithKeyCanonicalizer_whenPassedToNew_ConfiguresKeyCanonicalizer(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	canonicalizer := func(key string) string {
		return "canonical-" + key
	}

	cache := New[string, string](
		logr.Discard(),
		WithKeyCanonicalizer(canonicalizer))

	// Check the requested key canonicalizer is correctly in place
	g.Expect(cache.keyCanonicalizer("example")).To(Equal("canonical-example"))
}
