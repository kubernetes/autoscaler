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

package limitrange

import (
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes/fake"

	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/test"
)

const testNamespace = "test-namespace"

func TestNewNoopLimitsChecker(t *testing.T) {
	nlc := NewNoopLimitsCalculator()
	limitRange, err := nlc.GetContainerLimitRangeItem(testNamespace)
	assert.NoError(t, err)
	assert.Nil(t, limitRange)
}

func TestNoLimitRange(t *testing.T) {
	cs := fake.NewClientset()
	factory := informers.NewSharedInformerFactory(cs, 0)
	lc, err := NewLimitsRangeCalculator(factory)

	if assert.NoError(t, err) {
		limitRange, err := lc.GetContainerLimitRangeItem(testNamespace)
		assert.NoError(t, err)
		assert.Nil(t, limitRange)
	}
}

func TestGetContainerLimitRangeItem(t *testing.T) {
	baseContainerLimitRange := test.LimitRange().WithName("test-lr").WithNamespace(testNamespace).WithType(corev1.LimitTypeContainer)
	containerLimitRangeWithMax := baseContainerLimitRange.WithMax(test.Resources("2", "2")).Get()
	containerLimitRangeWithDefault := baseContainerLimitRange.WithDefault(test.Resources("2", "2")).Get()
	containerLimitRangeWithMin := baseContainerLimitRange.WithMin(test.Resources("2", "2")).Get()
	testCases := []struct {
		name           string
		limitRanges    []runtime.Object
		expectedErr    error
		expectedLimits *corev1.LimitRangeItem
	}{
		{
			name: "no matching limit ranges",
			limitRanges: []runtime.Object{
				test.LimitRange().WithName("different-namespace").WithNamespace("different").WithType(corev1.LimitTypeContainer).WithMax(test.Resources("2", "2")).Get(),
				test.LimitRange().WithName("different-type").WithNamespace(testNamespace).WithType(corev1.LimitTypePersistentVolumeClaim).WithMax(test.Resources("2", "2")).Get(),
			},
			expectedErr:    nil,
			expectedLimits: nil,
		},
		{
			name: "matching container limit range",
			limitRanges: []runtime.Object{
				containerLimitRangeWithMax,
			},
			expectedErr:    nil,
			expectedLimits: &containerLimitRangeWithMax.Spec.Limits[0],
		},
		{
			name: "with default value",
			limitRanges: []runtime.Object{
				containerLimitRangeWithDefault,
			},
			expectedErr:    nil,
			expectedLimits: &containerLimitRangeWithDefault.Spec.Limits[0],
		},
		{
			name: "respects min",
			limitRanges: []runtime.Object{
				containerLimitRangeWithMin,
			},
			expectedErr:    nil,
			expectedLimits: &containerLimitRangeWithMin.Spec.Limits[0],
		},
		{
			name: "multiple items",
			limitRanges: []runtime.Object{
				baseContainerLimitRange.WithMax(test.Resources("2", "2")).WithDefault(test.Resources("1.5", "1.5")).
					WithMin(test.Resources("1", "1")).Get(),
			},
			expectedErr: nil,
			expectedLimits: &corev1.LimitRangeItem{
				Type:    corev1.LimitTypeContainer,
				Min:     test.Resources("1", "1"),
				Max:     test.Resources("2", "2"),
				Default: test.Resources("1.5", "1.5"),
			},
		},
		{
			name: "takes lowest max",
			limitRanges: []runtime.Object{
				baseContainerLimitRange.WithMax(test.Resources("1.5", "1.5")).WithMax(test.Resources("2.", "2.")).Get(),
			},
			expectedErr: nil,
			expectedLimits: &corev1.LimitRangeItem{
				Type: corev1.LimitTypeContainer,
				Max:  test.Resources("1.5", "1.5"),
			},
		},
		{
			name: "takes highest min",
			limitRanges: []runtime.Object{
				baseContainerLimitRange.WithMin(test.Resources("1.5", "1.5")).WithMin(test.Resources("1.", "1.")).Get(),
			},
			expectedErr: nil,
			expectedLimits: &corev1.LimitRangeItem{
				Type: corev1.LimitTypeContainer,
				Min:  test.Resources("1.5", "1.5"),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cs := fake.NewClientset(tc.limitRanges...)
			factory := informers.NewSharedInformerFactory(cs, 0)
			lc, err := NewLimitsRangeCalculator(factory)

			factory.Start(t.Context().Done())
			_ = factory.WaitForCacheSync(t.Context().Done())
			if assert.NoError(t, err) {
				limitRange, err := lc.GetContainerLimitRangeItem(testNamespace)
				if tc.expectedErr == nil {
					assert.NoError(t, err)
				} else {
					assert.Error(t, err)
				}
				assert.Equal(t, tc.expectedLimits, limitRange)
			}
		})

	}
}
