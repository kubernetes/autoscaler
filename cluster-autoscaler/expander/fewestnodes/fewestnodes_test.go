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

package fewestnodes

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"k8s.io/autoscaler/cluster-autoscaler/expander"
)

func TestFewestNodesExpander(t *testing.T) {
	cases := []struct {
		name    string
		options []expander.Option
		want    expander.Option
	}{
		{
			name:    "SingleOption",
			options: []expander.Option{{NodeCount: 42}},
			want:    expander.Option{NodeCount: 42},
		},
		{
			name: "OptionsWithEqualNodeCounts",
			options: []expander.Option{
				{NodeCount: 42},
				{NodeCount: 42},
				{NodeCount: 42},
			},
			// We hope this is a random choice of one of the above. :)
			want: expander.Option{NodeCount: 42},
		},
		{
			name: "OptionsWithDifferentNodeCounts",
			options: []expander.Option{
				{NodeCount: 1},
				{NodeCount: 42},
				{NodeCount: 8},
			},
			want: expander.Option{NodeCount: 1},
		},
	}

	e := NewStrategy()

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := e.BestOption(tc.options, nil)
			assert.True(t, assert.ObjectsAreEqual(tc.want, *got), "\nwant: %#v\ngot: %#v", tc.want, *got)
		})
	}
}
