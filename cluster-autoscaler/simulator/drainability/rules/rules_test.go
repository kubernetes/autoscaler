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

	"github.com/google/go-cmp/cmp"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/drainability"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/framework"
	"k8s.io/autoscaler/cluster-autoscaler/utils/drain"
)

func TestDrainable(t *testing.T) {
	for desc, tc := range map[string]struct {
		rules Rules
		want  drainability.Status
	}{
		"no rules": {
			want: drainability.NewUndefinedStatus(),
		},
		"first non-undefined rule returned": {
			rules: Rules{
				fakeRule{drainability.NewUndefinedStatus()},
				fakeRule{drainability.NewDrainableStatus()},
				fakeRule{drainability.NewSkipStatus()},
			},
			want: drainability.NewDrainableStatus(),
		},
		"override match": {
			rules: Rules{
				fakeRule{drainability.Status{
					Outcome:   drainability.DrainOk,
					Overrides: []drainability.OutcomeType{drainability.BlockDrain},
				}},
				fakeRule{drainability.NewBlockedStatus(drain.NotEnoughPdb, nil)},
			},
			want: drainability.Status{
				Outcome:   drainability.DrainOk,
				Overrides: []drainability.OutcomeType{drainability.BlockDrain},
			},
		},
		"override no match": {
			rules: Rules{
				fakeRule{drainability.Status{
					Outcome:   drainability.DrainOk,
					Overrides: []drainability.OutcomeType{drainability.SkipDrain},
				}},
				fakeRule{drainability.NewBlockedStatus(drain.NotEnoughPdb, nil)},
			},
			want: drainability.NewBlockedStatus(drain.NotEnoughPdb, nil),
		},
		"override unreachable": {
			rules: Rules{
				fakeRule{drainability.NewSkipStatus()},
				fakeRule{drainability.Status{
					Outcome:   drainability.DrainOk,
					Overrides: []drainability.OutcomeType{drainability.BlockDrain},
				}},
				fakeRule{drainability.NewBlockedStatus(drain.NotEnoughPdb, nil)},
			},
			want: drainability.NewSkipStatus(),
		},
		"multiple overrides all run": {
			rules: Rules{
				fakeRule{drainability.Status{
					Outcome:   drainability.DrainOk,
					Overrides: []drainability.OutcomeType{drainability.SkipDrain},
				}},
				fakeRule{drainability.Status{
					Outcome:   drainability.SkipDrain,
					Overrides: []drainability.OutcomeType{drainability.BlockDrain},
				}},
				fakeRule{drainability.NewBlockedStatus(drain.NotEnoughPdb, nil)},
			},
			want: drainability.Status{
				Outcome:   drainability.SkipDrain,
				Overrides: []drainability.OutcomeType{drainability.BlockDrain},
			},
		},
		"multiple overrides respects order": {
			rules: Rules{
				fakeRule{drainability.Status{
					Outcome:   drainability.SkipDrain,
					Overrides: []drainability.OutcomeType{drainability.BlockDrain},
				}},
				fakeRule{drainability.Status{
					Outcome:   drainability.DrainOk,
					Overrides: []drainability.OutcomeType{drainability.BlockDrain},
				}},
				fakeRule{drainability.NewBlockedStatus(drain.NotEnoughPdb, nil)},
			},
			want: drainability.Status{
				Outcome:   drainability.SkipDrain,
				Overrides: []drainability.OutcomeType{drainability.BlockDrain},
			},
		},
	} {
		t.Run(desc, func(t *testing.T) {
			got := tc.rules.Drainable(nil, &apiv1.Pod{}, nil)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("Drainable(): got status diff (-want +got):\n%s", diff)
			}
		})
	}
}

type fakeRule struct {
	status drainability.Status
}

func (r fakeRule) Name() string {
	return "FakeRule"
}

func (r fakeRule) Drainable(*drainability.DrainContext, *apiv1.Pod, *framework.NodeInfo) drainability.Status {
	return r.status
}
