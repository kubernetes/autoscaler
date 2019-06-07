/*
Copyright 2019 The Kubernetes Authors.

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

package api

import (
	"math"
	"testing"

	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/stretchr/testify/assert"
)

func mustParseToPointer(str string) *resource.Quantity {
	val := resource.MustParse(str)
	return &val
}

func TestGetProportionalResourceLimit(t *testing.T) {
	tests := []struct {
		name               string
		originalLimit      *resource.Quantity
		originalRequest    *resource.Quantity
		recommendedRequest *resource.Quantity
		defaultLimit       *resource.Quantity
		expectLimit        *resource.Quantity
		expectAnnotation   bool
	}{
		{
			name:               "scale proportionally",
			originalLimit:      mustParseToPointer("2"),
			originalRequest:    mustParseToPointer("1"),
			recommendedRequest: mustParseToPointer("10"),
			expectLimit:        mustParseToPointer("20"),
		},
		{
			name:               "scale proportionally with default",
			originalRequest:    mustParseToPointer("1"),
			recommendedRequest: mustParseToPointer("10"),
			defaultLimit:       mustParseToPointer("2"),
			expectLimit:        mustParseToPointer("20"),
		},
		{
			name:               "no original limit",
			originalRequest:    mustParseToPointer("1"),
			recommendedRequest: mustParseToPointer("10"),
			expectLimit:        nil,
		},
		{
			name:               "no original request",
			originalLimit:      mustParseToPointer("2"),
			recommendedRequest: mustParseToPointer("10"),
			expectLimit:        mustParseToPointer("10"),
		},
		{
			name:               "limit equal to request",
			originalLimit:      mustParseToPointer("1"),
			originalRequest:    mustParseToPointer("1"),
			recommendedRequest: mustParseToPointer("10"),
			expectLimit:        mustParseToPointer("10"),
		},
		{
			name:               "go over milli cap",
			originalLimit:      mustParseToPointer("10G"),
			originalRequest:    mustParseToPointer("1m"),
			recommendedRequest: mustParseToPointer("10G"),
			expectLimit:        resource.NewMilliQuantity(math.MaxInt64, resource.DecimalExponent),
			expectAnnotation:   true,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			gotLimit, gotAnnotation := getProportionalResourceLimit(core.ResourceCPU, tc.originalLimit, tc.originalRequest, tc.recommendedRequest, tc.defaultLimit)
			if tc.expectLimit == nil {
				assert.Nil(t, gotLimit)
			} else {
				if assert.NotNil(t, gotLimit) {
					assert.Equal(t, gotLimit.MilliValue(), tc.expectLimit.MilliValue())
				}
			}
			assert.Equal(t, gotAnnotation != "", tc.expectAnnotation)
		})
	}
}
