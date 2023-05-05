/*
Copyright 2023 The Kubernetes Authors.

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

package scaledowncandidates

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/test"
	testprovider "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/test"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	"k8s.io/autoscaler/cluster-autoscaler/context"
)

func TestGetScaleDownCandidates(t *testing.T) {
	n1 := &v1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: "n1",
		},
	}

	n2 := &v1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: "n2",
		},
	}

	n3 := &v1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: "n3",
		},
	}

	ctx := context.AutoscalingContext{
		AutoscalingOptions: config.AutoscalingOptions{
			ScaleDownDelayAfterAdd:     time.Minute * 10,
			ScaleDownDelayAfterDelete:  time.Minute * 10,
			ScaleDownDelayAfterFailure: time.Minute * 10,
			ScaleDownDelayTypeLocal:    true,
		},
	}

	testCases := map[string]struct {
		autoscalingContext context.AutoscalingContext
		candidates         []*v1.Node
		expected           []*v1.Node
		setupProcessor     func(p *ScaleDownCandidatesDelayProcessor) *ScaleDownCandidatesDelayProcessor
	}{
		// Expectation: no nodegroups should be filtered out
		"no scale ups - no scale downs - no scale down failures": {
			autoscalingContext: ctx,
			candidates:         []*v1.Node{n1, n2, n3},
			expected:           []*v1.Node{n1, n2, n3},
			setupProcessor:     nil,
		},
		// Expectation: only nodegroups in cool-down should be filtered out
		"no scale ups - 2 scale downs - no scale down failures": {
			autoscalingContext: ctx,
			candidates:         []*v1.Node{n1, n2, n3},
			expected:           []*v1.Node{n1, n3},
			setupProcessor: func(p *ScaleDownCandidatesDelayProcessor) *ScaleDownCandidatesDelayProcessor {
				// fake nodegroups for calling `RegisterScaleDown`
				ng2 := test.NewTestNodeGroup("ng-2", 0, 0, 0, false, false, "", nil, nil)
				ng3 := test.NewTestNodeGroup("ng-3", 0, 0, 0, false, false, "", nil, nil)
				// in cool down
				p.RegisterScaleDown(ng2, "n2", time.Now().Add(-time.Minute*5), time.Time{})
				// not in cool down anymore
				p.RegisterScaleDown(ng3, "n3", time.Now().Add(-time.Minute*11), time.Time{})

				return p
			},
		},
		// Expectation: only nodegroups in cool-down should be filtered out
		"1 scale up - no scale down - no scale down failures": {
			autoscalingContext: ctx,
			candidates:         []*v1.Node{n1, n2, n3},
			expected:           []*v1.Node{n1, n3},
			setupProcessor: func(p *ScaleDownCandidatesDelayProcessor) *ScaleDownCandidatesDelayProcessor {
				// fake nodegroups for calling `RegisterScaleUp`
				ng2 := test.NewTestNodeGroup("ng-2", 0, 0, 0, false, false, "", nil, nil)
				ng3 := test.NewTestNodeGroup("ng-3", 0, 0, 0, false, false, "", nil, nil)

				// in cool down
				p.RegisterScaleUp(ng2, 0, time.Now().Add(-time.Minute*5))
				// not in cool down anymore
				p.RegisterScaleUp(ng3, 0, time.Now().Add(-time.Minute*11))
				return p
			},
		},
		// Expectation: only nodegroups in cool-down should be filtered out
		"no scale up - no scale down - 1 scale down failure": {
			autoscalingContext: ctx,
			candidates:         []*v1.Node{n1, n2, n3},
			expected:           []*v1.Node{n1, n3},
			setupProcessor: func(p *ScaleDownCandidatesDelayProcessor) *ScaleDownCandidatesDelayProcessor {
				// fake nodegroups for calling `RegisterScaleUp`
				ng2 := test.NewTestNodeGroup("ng-2", 0, 0, 0, false, false, "", nil, nil)
				ng3 := test.NewTestNodeGroup("ng-3", 0, 0, 0, false, false, "", nil, nil)

				// in cool down
				p.RegisterFailedScaleDown(ng2, "", time.Now().Add(-time.Minute*5))
				// not in cool down anymore
				p.RegisterFailedScaleDown(ng3, "", time.Now().Add(-time.Minute*11))
				return p
			},
		},
	}

	for description, testCase := range testCases {
		t.Run(description, func(t *testing.T) {
			provider := testprovider.NewTestCloudProvider(nil, nil)

			p := NewScaleDownCandidatesDelayProcessor()

			if testCase.setupProcessor != nil {
				p = testCase.setupProcessor(p)
			}

			provider.AddNodeGroup("ng-1", 1, 3, 2)
			provider.AddNode("ng-1", n1)
			provider.AddNodeGroup("ng-2", 1, 3, 2)
			provider.AddNode("ng-2", n2)
			provider.AddNodeGroup("ng-3", 1, 3, 2)
			provider.AddNode("ng-3", n3)

			testCase.autoscalingContext.CloudProvider = provider

			no, err := p.GetScaleDownCandidates(&testCase.autoscalingContext, testCase.candidates)

			assert.NoError(t, err)
			assert.Equal(t, testCase.expected, no)
		})
	}
}
