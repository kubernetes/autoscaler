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

package leastnodes

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"k8s.io/autoscaler/cluster-autoscaler/expander"
)

func TestLeastNodes(t *testing.T) {
	e := NewFilter()

	eo0 := expander.Option{Debug: "EO0", NodeCount: 2}
	ret := e.BestOptions([]expander.Option{eo0}, nil)
	assert.Equal(t, ret, []expander.Option{eo0})

	eo1 := expander.Option{Debug: "EO1", NodeCount: 1}
	ret = e.BestOptions([]expander.Option{eo0, eo1}, nil)
	assert.Equal(t, ret, []expander.Option{eo1})

	eo1b := expander.Option{Debug: "EO1b", NodeCount: 1}
	ret = e.BestOptions([]expander.Option{eo0, eo1, eo1b}, nil)
	assert.NotEqual(t, ret, []expander.Option{eo1})
	assert.ObjectsAreEqual(ret, []expander.Option{eo1, eo1b})
}
