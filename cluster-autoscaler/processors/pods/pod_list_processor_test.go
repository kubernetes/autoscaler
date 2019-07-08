/*
Copyright 2018 The Kubernetes Authors.

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

	"k8s.io/autoscaler/cluster-autoscaler/context"
	. "k8s.io/autoscaler/cluster-autoscaler/utils/test"
)

func TestPodListProcessor(t *testing.T) {
	context := &context.AutoscalingContext{}
	p1 := BuildTestPod("p1", 40, 0)
	p2 := BuildTestPod("p2", 400, 0)
	n1 := BuildTestNode("n1", 100, 1000)
	n2 := BuildTestNode("n1", 100, 1000)
	unschedulablePods := []*apiv1.Pod{p1}
	allScheduledPods := []*apiv1.Pod{p2}
	allNodes := []*apiv1.Node{n1, n2}
	readyNodes := []*apiv1.Node{n1, n2}
	podListProcessor := NewDefaultPodListProcessor()
	gotUnschedulablePods, gotAllScheduled, err := podListProcessor.Process(context, unschedulablePods, allScheduledPods, allNodes, readyNodes, []*apiv1.Node{})
	if len(gotUnschedulablePods) != 1 || len(gotAllScheduled) != 1 || err != nil {
		t.Errorf("Error podListProcessor.Process() = %v, %v, %v want %v, %v, nil ",
			gotUnschedulablePods, gotAllScheduled, err, unschedulablePods, allScheduledPods)
	}

}
