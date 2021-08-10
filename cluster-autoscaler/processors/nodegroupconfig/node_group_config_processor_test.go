/*
Copyright 2020 The Kubernetes Authors.

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

package nodegroupconfig

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/mocks"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	"k8s.io/autoscaler/cluster-autoscaler/context"

	"github.com/stretchr/testify/assert"
)

// This test covers all Get* methods implemented by
// DelegatingNodeGroupConfigProcessor. The implementation and expectations are
// identical hence a single test for the whole bunch.
func TestDelegatingNodeGroupConfigProcessor(t *testing.T) {
	// Different methods covered by this test have identical implementation,
	// but return values of different types.
	// This enum is a generic way to specify test expectations without
	// some reflection magic.
	type Want int
	var NIL Want = 0
	var GLOBAL Want = 1
	var NG Want = 2

	globalOpts := config.NodeGroupAutoscalingOptions{
		ScaleDownUnneededTime:            3 * time.Minute,
		ScaleDownUnreadyTime:             4 * time.Minute,
		ScaleDownGpuUtilizationThreshold: 0.6,
		ScaleDownUtilizationThreshold:    0.5,
	}
	ngOpts := &config.NodeGroupAutoscalingOptions{
		ScaleDownUnneededTime:            10 * time.Minute,
		ScaleDownUnreadyTime:             11 * time.Minute,
		ScaleDownGpuUtilizationThreshold: 0.85,
		ScaleDownUtilizationThreshold:    0.75,
	}

	testUnneededTime := func(t *testing.T, p DelegatingNodeGroupConfigProcessor, c *context.AutoscalingContext, ng cloudprovider.NodeGroup, w Want, we error) {
		res, err := p.GetScaleDownUnneededTime(c, ng)
		assert.Equal(t, err, we)
		results := map[Want]time.Duration{
			NIL:    time.Duration(0),
			GLOBAL: 3 * time.Minute,
			NG:     10 * time.Minute,
		}
		assert.Equal(t, res, results[w])
	}
	testUnreadyTime := func(t *testing.T, p DelegatingNodeGroupConfigProcessor, c *context.AutoscalingContext, ng cloudprovider.NodeGroup, w Want, we error) {
		res, err := p.GetScaleDownUnreadyTime(c, ng)
		assert.Equal(t, err, we)
		results := map[Want]time.Duration{
			NIL:    time.Duration(0),
			GLOBAL: 4 * time.Minute,
			NG:     11 * time.Minute,
		}
		assert.Equal(t, res, results[w])
	}
	testUtilizationThreshold := func(t *testing.T, p DelegatingNodeGroupConfigProcessor, c *context.AutoscalingContext, ng cloudprovider.NodeGroup, w Want, we error) {
		res, err := p.GetScaleDownUtilizationThreshold(c, ng)
		assert.Equal(t, err, we)
		results := map[Want]float64{
			NIL:    0.0,
			GLOBAL: 0.5,
			NG:     0.75,
		}
		assert.Equal(t, res, results[w])
	}
	testGpuThreshold := func(t *testing.T, p DelegatingNodeGroupConfigProcessor, c *context.AutoscalingContext, ng cloudprovider.NodeGroup, w Want, we error) {
		res, err := p.GetScaleDownGpuUtilizationThreshold(c, ng)
		assert.Equal(t, err, we)
		results := map[Want]float64{
			NIL:    0.0,
			GLOBAL: 0.6,
			NG:     0.85,
		}
		assert.Equal(t, res, results[w])
	}

	funcs := map[string]func(*testing.T, DelegatingNodeGroupConfigProcessor, *context.AutoscalingContext, cloudprovider.NodeGroup, Want, error){
		"ScaleDownUnneededTime":            testUnneededTime,
		"ScaleDownUnreadyTime":             testUnreadyTime,
		"ScaleDownUtilizationThreshold":    testUtilizationThreshold,
		"ScaleDownGpuUtilizationThreshold": testGpuThreshold,
		"MultipleOptions": func(t *testing.T, p DelegatingNodeGroupConfigProcessor, c *context.AutoscalingContext, ng cloudprovider.NodeGroup, w Want, we error) {
			testUnneededTime(t, p, c, ng, w, we)
			testUnreadyTime(t, p, c, ng, w, we)
			testUtilizationThreshold(t, p, c, ng, w, we)
			testGpuThreshold(t, p, c, ng, w, we)
		},
		"RepeatingTheSameCallGivesConsistentResults": func(t *testing.T, p DelegatingNodeGroupConfigProcessor, c *context.AutoscalingContext, ng cloudprovider.NodeGroup, w Want, we error) {
			testUnneededTime(t, p, c, ng, w, we)
			testUnneededTime(t, p, c, ng, w, we)
			// throw in a different call
			testGpuThreshold(t, p, c, ng, w, we)
			testUnneededTime(t, p, c, ng, w, we)
		},
	}

	for fname, fn := range funcs {
		cases := map[string]struct {
			globalOptions config.NodeGroupAutoscalingOptions
			ngOptions     *config.NodeGroupAutoscalingOptions
			ngError       error
			want          Want
			wantError     error
		}{
			"NodeGroup.GetOptions not implemented": {
				globalOptions: globalOpts,
				ngError:       cloudprovider.ErrNotImplemented,
				want:          GLOBAL,
			},
			"NodeGroup returns error leads to error": {
				globalOptions: globalOpts,
				ngError:       errors.New("This sentence is false."),
				wantError:     errors.New("This sentence is false."),
			},
			"NodeGroup returns no value fallbacks to default": {
				globalOptions: globalOpts,
				want:          GLOBAL,
			},
			"NodeGroup option overrides global default": {
				globalOptions: globalOpts,
				ngOptions:     ngOpts,
				want:          NG,
			},
		}
		for tn, tc := range cases {
			t.Run(fmt.Sprintf("[%s] %s", fname, tn), func(t *testing.T) {
				context := &context.AutoscalingContext{
					AutoscalingOptions: config.AutoscalingOptions{
						NodeGroupDefaults: tc.globalOptions,
					},
				}
				ng := &mocks.NodeGroup{}
				ng.On("GetOptions", tc.globalOptions).Return(tc.ngOptions, tc.ngError)
				p := DelegatingNodeGroupConfigProcessor{}
				fn(t, p, context, ng, tc.want, tc.wantError)
			})
		}
	}
}
