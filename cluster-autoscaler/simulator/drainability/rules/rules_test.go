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

package rules

import (
	"testing"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/drainability"
)

type fakeSkipRule struct{}

func (r fakeSkipRule) Drainable(*drainability.DrainContext, *apiv1.Pod) drainability.Status {
	return drainability.NewSkipStatus()
}

type fakeDrainRule struct{}

func (r fakeDrainRule) Drainable(*drainability.DrainContext, *apiv1.Pod) drainability.Status {
	return drainability.NewDrainableStatus()
}

func TestDrainable(t *testing.T) {
	for desc, tc := range map[string]struct {
		rules      Rules
		wantStatus drainability.Status
	}{
		"no rules": {
			wantStatus: drainability.NewUndefinedStatus(),
		},
		"outcome priority is respected": {
			rules: Rules{
				fakeDrainRule{},
				fakeSkipRule{},
				fakeDrainRule{},
			},
			wantStatus: drainability.NewSkipStatus(),
		},
	} {
		t.Run(desc, func(t *testing.T) {
			got := tc.rules.Drainable(nil, nil)
			if got != tc.wantStatus {
				t.Errorf("Drainable(): got status: %v, want status: %v", got, tc.wantStatus)
			}
		})
	}
}
