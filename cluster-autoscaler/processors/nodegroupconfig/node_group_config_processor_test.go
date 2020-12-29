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
	"testing"
	"time"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/mocks"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	"k8s.io/autoscaler/cluster-autoscaler/context"

	"github.com/stretchr/testify/assert"
)

func TestApplyingDefaults(t *testing.T) {
	defaultOptions := config.NodeGroupAutoscalingOptions{
		ScaleDownUnneededTime: 3 * time.Minute,
	}

	cases := map[string]struct {
		globalOptions         config.NodeGroupAutoscalingOptions
		ngOptions             *config.NodeGroupAutoscalingOptions
		ngError               error
		wantScaleDownUnneeded time.Duration
		wantError             error
	}{
		"NodeGroup.GetOptions not implemented": {
			globalOptions:         defaultOptions,
			ngError:               cloudprovider.ErrNotImplemented,
			wantScaleDownUnneeded: 3 * time.Minute,
		},
		"NodeGroup returns error leads to error": {
			globalOptions: defaultOptions,
			ngError:       errors.New("This sentence is false."),
			wantError:     errors.New("This sentence is false."),
		},
		"NodeGroup returns no value fallbacks to default": {
			globalOptions:         defaultOptions,
			wantScaleDownUnneeded: 3 * time.Minute,
		},
		"NodeGroup option overrides global default": {
			globalOptions: defaultOptions,
			ngOptions: &config.NodeGroupAutoscalingOptions{
				ScaleDownUnneededTime: 10 * time.Minute,
			},
			wantScaleDownUnneeded: 10 * time.Minute,
		},
	}
	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			context := &context.AutoscalingContext{
				AutoscalingOptions: config.AutoscalingOptions{
					NodeGroupAutoscalingOptions: tc.globalOptions,
				},
			}
			ng := &mocks.NodeGroup{}
			ng.On("GetOptions", tc.globalOptions).Return(tc.ngOptions, tc.ngError).Once()
			p := NewDefaultNodeGroupConfigProcessor()
			res, err := p.GetScaleDownUnneededTime(context, ng)
			assert.Equal(t, res, tc.wantScaleDownUnneeded)
			assert.Equal(t, err, tc.wantError)
		})
	}
}
