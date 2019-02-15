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

package target

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	vpa_types_v1beta1 "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1beta1"
	vpa_types_v1beta2 "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1beta2"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/test"
)

// VerticalPodAutoscalerBuilder helps building test instances of VerticalPodAutoscaler.
type VerticalPodAutoscalerBuilder interface {
	WithName(vpaName string) VerticalPodAutoscalerBuilder
	WithNamespace(namespace string) VerticalPodAutoscalerBuilder
	WithSelector(labelSelector string) VerticalPodAutoscalerBuilder
	GetV1Beta1() *vpa_types_v1beta1.VerticalPodAutoscaler
	GetV1Beta2() *vpa_types_v1beta2.VerticalPodAutoscaler
}

// VerticalPodAutoscaler returns a new VerticalPodAutoscalerBuilder.
func VerticalPodAutoscaler() VerticalPodAutoscalerBuilder {
	return &verticalPodAutoscalerBuilder{
		namespace: "default",
	}
}

type verticalPodAutoscalerBuilder struct {
	vpaName       string
	namespace     string
	labelSelector *metav1.LabelSelector
}

func (b *verticalPodAutoscalerBuilder) WithName(vpaName string) VerticalPodAutoscalerBuilder {
	c := *b
	c.vpaName = vpaName
	return &c
}

func (b *verticalPodAutoscalerBuilder) WithNamespace(namespace string) VerticalPodAutoscalerBuilder {
	c := *b
	c.namespace = namespace
	return &c
}

func (b *verticalPodAutoscalerBuilder) WithSelector(labelSelector string) VerticalPodAutoscalerBuilder {
	c := *b
	if labelSelector, err := metav1.ParseToLabelSelector(labelSelector); err != nil {
		panic(err)
	} else {
		c.labelSelector = labelSelector
	}
	return &c
}

func (b *verticalPodAutoscalerBuilder) GetV1Beta1() *vpa_types_v1beta1.VerticalPodAutoscaler {
	return &vpa_types_v1beta1.VerticalPodAutoscaler{
		ObjectMeta: metav1.ObjectMeta{
			Name:      b.vpaName,
			Namespace: b.namespace,
		},
		Spec: vpa_types_v1beta1.VerticalPodAutoscalerSpec{
			Selector: b.labelSelector,
		},
	}
}

func (b *verticalPodAutoscalerBuilder) GetV1Beta2() *vpa_types_v1beta2.VerticalPodAutoscaler {
	if b.labelSelector != nil {
		panic("v1beta2 doesn't support selector")
	}
	return &vpa_types_v1beta2.VerticalPodAutoscaler{
		ObjectMeta: metav1.ObjectMeta{
			Name:      b.vpaName,
			Namespace: b.namespace,
		},
	}
}

func parseLabelSelector(selector string) labels.Selector {
	labelSelector, _ := metav1.ParseToLabelSelector(selector)
	parsedSelector, _ := metav1.LabelSelectorAsSelector(labelSelector)
	return parsedSelector
}

func TestLegacySelector(t *testing.T) {

	type testCase struct {
		vpa                 *vpa_types_v1beta2.VerticalPodAutoscaler
		vpaInStore          *vpa_types_v1beta1.VerticalPodAutoscaler
		expectedSelector    labels.Selector
		expectedError       bool
		expectedErrorString string
	}

	testCases := []testCase{
		{
			vpa:                 nil,
			vpaInStore:          nil,
			expectedSelector:    nil,
			expectedError:       true,
			expectedErrorString: "Failed to list v1beta1 VPAs. Reason: Cannot list",
		}, {
			vpa:                 VerticalPodAutoscaler().WithName("a").GetV1Beta2(),
			vpaInStore:          nil,
			expectedSelector:    nil,
			expectedError:       true,
			expectedErrorString: "Failed to list v1beta1 VPAs. Reason: Cannot list",
		}, {
			vpa:                 VerticalPodAutoscaler().WithName("a").GetV1Beta2(),
			vpaInStore:          VerticalPodAutoscaler().WithName("b").GetV1Beta1(),
			expectedSelector:    nil,
			expectedError:       true,
			expectedErrorString: "v1beta1 VPA not found",
		}, {
			vpa:                 VerticalPodAutoscaler().WithName("a").GetV1Beta2(),
			vpaInStore:          VerticalPodAutoscaler().WithName("a").GetV1Beta1(),
			expectedSelector:    nil,
			expectedError:       true,
			expectedErrorString: "v1beta1 selector not found",
		}, {
			vpa:              VerticalPodAutoscaler().WithName("a").GetV1Beta2(),
			vpaInStore:       VerticalPodAutoscaler().WithName("a").WithSelector("app = t").GetV1Beta1(),
			expectedSelector: parseLabelSelector("app = t"),
			expectedError:    false,
		},
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("test case number: %d", i), func(t *testing.T) {
			vpaLister := &test.VerticalPodAutoscalerV1Beta1ListerMock{}
			if tc.vpaInStore == nil {
				vpaLister.On("List").Return(nil, fmt.Errorf("Cannot list"))
			} else {
				vpaLister.On("List").Return([]*vpa_types_v1beta1.VerticalPodAutoscaler{tc.vpaInStore}, nil)
			}

			fetcher := beta1TargetSelectorFetcher{
				vpaLister: vpaLister,
			}

			selector, err := fetcher.Fetch(tc.vpa)

			if tc.expectedError {
				assert.Equal(t, fmt.Errorf(tc.expectedErrorString), err)
			} else {
				assert.Nil(t, err)
			}

			if tc.expectedSelector == nil {
				assert.Nil(t, selector)
			} else {
				assert.Equal(t, tc.expectedSelector.String(), selector.String())
			}
		})
	}

}
