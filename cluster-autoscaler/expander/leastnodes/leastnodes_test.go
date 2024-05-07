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
	for _, tc := range []struct {
		name                     string
		expansionOptions         []expander.Option
		expectedExpansionOptions []expander.Option
	}{
		{
			name:                     "no options",
			expansionOptions:         nil,
			expectedExpansionOptions: nil,
		},
		{
			name: "no valid options",
			expansionOptions: []expander.Option{
				{Debug: "EO0", NodeCount: 0},
			},
			expectedExpansionOptions: nil,
		},
		{
			name: "1 valid option",
			expansionOptions: []expander.Option{
				{Debug: "EO0", NodeCount: 2},
			},
			expectedExpansionOptions: []expander.Option{
				{Debug: "EO0", NodeCount: 2},
			},
		},
		{
			name: "2 valid options, not equal",
			expansionOptions: []expander.Option{
				{Debug: "EO0", NodeCount: 2},
				{Debug: "EO1", NodeCount: 1},
			},
			expectedExpansionOptions: []expander.Option{
				{Debug: "EO1", NodeCount: 1},
			},
		},
		{
			name: "3 valid options, 2 equal",
			expansionOptions: []expander.Option{
				{Debug: "EO0", NodeCount: 6},
				{Debug: "EO1", NodeCount: 2},
				{Debug: "EO2", NodeCount: 2},
			},
			expectedExpansionOptions: []expander.Option{
				{Debug: "EO1", NodeCount: 2},
				{Debug: "EO2", NodeCount: 2},
			},
		},
		{
			name: "3 valid options, all equal",
			expansionOptions: []expander.Option{
				{Debug: "EO0", NodeCount: 8},
				{Debug: "EO1", NodeCount: 8},
				{Debug: "EO2", NodeCount: 8},
			},
			expectedExpansionOptions: []expander.Option{
				{Debug: "EO0", NodeCount: 8},
				{Debug: "EO1", NodeCount: 8},
				{Debug: "EO2", NodeCount: 8},
			},
		},
		{
			name: "6 valid options, 1 invalid option, 3 equal",
			expansionOptions: []expander.Option{
				{Debug: "EO0", NodeCount: 23},
				{Debug: "EO1", NodeCount: 0},
				{Debug: "EO2", NodeCount: 5},
				{Debug: "EO3", NodeCount: 8},
				{Debug: "EO4", NodeCount: 5},
				{Debug: "EO5", NodeCount: 5},
				{Debug: "EO6", NodeCount: 22},
			},
			expectedExpansionOptions: []expander.Option{
				{Debug: "EO2", NodeCount: 5},
				{Debug: "EO4", NodeCount: 5},
				{Debug: "EO5", NodeCount: 5},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			e := NewFilter()
			ret := e.BestOptions(tc.expansionOptions, nil)
			assert.Equal(t, ret, tc.expectedExpansionOptions)
		})
	}
}
