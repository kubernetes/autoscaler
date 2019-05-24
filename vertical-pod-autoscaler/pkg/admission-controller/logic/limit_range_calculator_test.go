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

package logic

import (
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/informers"

	//"fmt"
	"testing"

	//"k8s.io/apimachinery/pkg/runtime"
	//"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes/fake"

	apiv1 "k8s.io/api/core/v1"
	//"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/test"
)

func getPod() *apiv1.Pod {
	labels := map[string]string{"app": "testingApp"}
	return test.Pod().WithName("test_uninitialized").AddContainer(test.BuildTestContainer("container1", "", "")).WithLabels(labels).Get()
}

func TestNewNoopLimitsChecker(t *testing.T) {
	nlc := NewNoopLimitsCalculator()
	limitRange, err := nlc.GetContainerLimitRangeItem(getPod().Namespace)
	if assert.NoError(t, err) {
		assert.Nil(t, limitRange)
	}
}

func TestNoLimitRange(t *testing.T) {
	cs := fake.NewSimpleClientset()
	factory := informers.NewSharedInformerFactory(cs, 0)
	lc, err := NewLimitsRangeCalculator(factory)

	if assert.NoError(t, err) {
		limitRange, err := lc.GetContainerLimitRangeItem(getPod().Namespace)
		if assert.NoError(t, err) {
			assert.Nil(t, limitRange)
		}
	}
}

func TestUpdateResourceLimits(t *testing.T) {

	testCases := []struct {
		name         string
		pod          *apiv1.Pod
		limitRanges  []runtime.Object
		expectErr    error
		expectLimits *apiv1.LimitRangeItem
	}{
		{
			name: "no matching limit ranges",
			pod:  getPod(),
			limitRanges: []runtime.Object{
				test.LimitRange().WithName("different-namespace").WithNamespace("different").WithType(apiv1.LimitTypePod).WithMax(test.Resources("2", "2")).Get(),
				test.LimitRange().WithName("differen-type").WithNamespace("default").WithType(apiv1.LimitTypePersistentVolumeClaim).WithMax(test.Resources("2", "2")).Get(),
			},
			expectErr:    nil,
			expectLimits: nil,
		},
		{
			name: "matching container limit range",
			pod:  getPod(),
			limitRanges: []runtime.Object{
				test.LimitRange().WithName("default").WithNamespace("default").WithType(apiv1.LimitTypeContainer).WithMax(test.Resources("2", "2")).Get(),
			},
			expectErr: nil,
			expectLimits: &apiv1.LimitRangeItem{
				Max: test.Resources("2", "2"),
			},
		},
		{
			name: "with default value",
			pod:  getPod(),
			limitRanges: []runtime.Object{
				test.LimitRange().WithName("default").WithNamespace("default").WithType(apiv1.LimitTypeContainer).WithDefault(test.Resources("2", "2")).Get(),
			},
			expectErr: nil,
			expectLimits: &apiv1.LimitRangeItem{
				Default: test.Resources("2", "2"),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cs := fake.NewSimpleClientset(tc.limitRanges...)
			factory := informers.NewSharedInformerFactory(cs, 0)
			lc, err := NewLimitsRangeCalculator(factory)
			if assert.NoError(t, err) {
				labels := map[string]string{"app": "testingApp"}
				pod := test.Pod().WithName("test_uninitialized").AddContainer(test.BuildTestContainer("container1", "", "")).WithLabels(labels).Get()
				limitRange, err := lc.GetContainerLimitRangeItem(pod.Namespace)
				if tc.expectErr == nil {
					assert.NoError(t, err)
				} else {
					assert.Error(t, err)
				}
				assert.Equal(t, tc.expectLimits, limitRange)
			}
		})

	}
}
