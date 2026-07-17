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

package factory

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/autoscaler/cluster-autoscaler/expander"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/framework"
)

type substringTestFilterStrategy struct {
	substring string
}

func newSubstringTestFilterStrategy(substring string) *substringTestFilterStrategy {
	return &substringTestFilterStrategy{
		substring: substring,
	}
}

func (s *substringTestFilterStrategy) BestOptions(expansionOptions []expander.Option, nodeInfo map[string]*framework.NodeInfo) []expander.Option {
	var ret []expander.Option
	for _, option := range expansionOptions {
		if strings.Contains(option.Debug, s.substring) {
			ret = append(ret, option)
		}
	}
	return ret

}

func (s *substringTestFilterStrategy) BestOption(expansionOptions []expander.Option, nodeInfo map[string]*framework.NodeInfo) *expander.Option {
	ret := s.BestOptions(expansionOptions, nodeInfo)
	if len(ret) == 0 {
		return nil
	}
	return &ret[0]
}

func TestChainStrategy_BestOption(t *testing.T) {
	for name, tc := range map[string]struct {
		filters  []expander.Filter
		fallback expander.Strategy
		options  []expander.Option
		expected *expander.Option
	}{
		"selects with no filters": {
			filters:  []expander.Filter{},
			fallback: newSubstringTestFilterStrategy("a"),
			options: []expander.Option{
				*newOption("b"),
				*newOption("a"),
			},
			expected: newOption("a"),
		},
		"filters with one filter": {
			filters: []expander.Filter{
				newSubstringTestFilterStrategy("a"),
			},
			fallback: newSubstringTestFilterStrategy("b"),
			options: []expander.Option{
				*newOption("ab"),
				*newOption("b"),
			},
			expected: newOption("ab"),
		},
		"filters with multiple filters": {
			filters: []expander.Filter{
				newSubstringTestFilterStrategy("a"),
				newSubstringTestFilterStrategy("b"),
			},
			fallback: newSubstringTestFilterStrategy("x"),
			options: []expander.Option{
				*newOption("xab"),
				*newOption("xa"),
				*newOption("x"),
			},
			expected: newOption("xab"),
		},
		"selects from multiple after filters": {
			filters: []expander.Filter{
				newSubstringTestFilterStrategy("x"),
			},
			fallback: newSubstringTestFilterStrategy("a"),
			options: []expander.Option{
				*newOption("xc"),
				*newOption("xaa"),
				*newOption("xab"),
			},
			expected: newOption("xaa"),
		},
		"short circuits": {
			filters: []expander.Filter{
				newSubstringTestFilterStrategy("a"),
				newSubstringTestFilterStrategy("b"),
			},
			fallback: newSubstringTestFilterStrategy("x"),
			options: []expander.Option{
				*newOption("a"),
			},
			expected: newOption("a"),
		},
	} {
		t.Run(name, func(t *testing.T) {
			subject := newChainStrategy(tc.filters, tc.fallback)
			actual := subject.BestOption(tc.options, nil)
			assert.Equal(t, tc.expected, actual)
		})
	}
}

func newOption(debug string) *expander.Option {
	return &expander.Option{
		Debug: debug,
	}
}
