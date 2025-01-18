/*
Copyright 2025 The Kubernetes Authors.

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

package pods

import (
	"testing"

	apiv1 "k8s.io/api/core/v1"
	testutils "k8s.io/autoscaler/cluster-autoscaler/utils/test"
)

func TestDefaultScaleUpEnforcer(t *testing.T) {
	p1 := testutils.BuildTestPod("p1", 40, 0)
	unschedulablePods := []*apiv1.Pod{p1}
	scaleUpEnforcer := NewDefaultScaleUpEnforcer()
	forceScaleUp := scaleUpEnforcer.ShouldForceScaleUp(unschedulablePods)
	if forceScaleUp {
		t.Errorf("Error: scaleUpEnforcer should not force scale up by default")
	}
}
