/*
Copyright 2026 The Kubernetes Authors.

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

package efficiency

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	apiv1 "k8s.io/api/core/v1"
	cacontext "k8s.io/autoscaler/cluster-autoscaler/context"
	"k8s.io/autoscaler/cluster-autoscaler/core/scaledown/planner"
	"k8s.io/autoscaler/cluster-autoscaler/processors"
)

type scaleDownDependencies struct {
	AutoscalingCtx *cacontext.AutoscalingContext
	Processors     *processors.AutoscalingProcessors
	Planner        *planner.Planner
	Actuator       *fakeActuator
	MetricsTracker *metricsTracker
}

// scaleDownStrategyFactory will define how the scaledown picks/orders nodes for scale down.
// A named strategy will most probably add processors, change sorting and set the inners of scaledown to fulfill its objective.
type scaleDownStrategyFactory func(
	b *testing.B,
	metricsTracker *metricsTracker,
	deps scaleDownDependencies,
)

// vanillaScaleDownStrategy is a bare scaledown evaluation loop without any algorithmic modifications, tailored setups etc.
// With high probability, this function will be rearchitected to smaller functions and logical components.
var vanillaScaleDownStrategy scaleDownStrategyFactory = func(
	b *testing.B,
	metricsTracker *metricsTracker,
	deps scaleDownDependencies,
) {

	autoscalingCtx := deps.AutoscalingCtx
	procs := deps.Processors
	p := deps.Planner
	fakeAct := deps.Actuator

	var scaleDownCandidates []*apiv1.Node
	var podDestinations []*apiv1.Node

	for scaleDownLoop := 1; ; scaleDownLoop++ {
		currentTime := time.Now()
		infos, err := autoscalingCtx.ClusterSnapshot.ListNodeInfos()
		if err != nil {
			b.Fatalf("error retrieving nodes from cluster snapshot, err %s", err.Error())
		}

		if len(infos) == 0 {
			b.Logf("no more nodes in the cluster snapshot, final state reached after %d scaledown loops", scaleDownLoop)
			break
		}

		var currentNodes []*apiv1.Node
		for _, nodeInfo := range infos {
			currentNodes = append(currentNodes, nodeInfo.Node())
		}

		if procs == nil || procs.ScaleDownNodeProcessor == nil {
			scaleDownCandidates = currentNodes
			podDestinations = currentNodes
		} else {
			scaleDownCandidates, err = procs.ScaleDownNodeProcessor.GetScaleDownCandidates(
				autoscalingCtx, currentNodes)
			if err != nil {
				b.Fatalf("error getting scale-down candidates: %v", err)
			}
		}

		err = p.UpdateClusterState(podDestinations, scaleDownCandidates, fakeAct.actuationStatus, currentTime)
		assert.NoError(b, err)

		empty, needDrain := p.NodesToDelete(currentTime.Add(11 * time.Second))
		// No more nodes to delete, final state reached.
		if len(empty) == 0 && len(needDrain) == 0 {
			b.ReportMetric(float64(scaleDownLoop), "loops")
			break
		}

		_, scaleDownNodes, typedErr := fakeAct.startDeletion(empty, needDrain)
		if typedErr != nil {
			b.Errorf("error deleting nodes in fake actuator, err %s", typedErr.Error())
		}

		metricsTracker.Compute(autoscalingCtx.ClusterSnapshot, scaleDownNodes, b)
		metricsTracker.Report(fmt.Sprintf("End of Loop %d", scaleDownLoop), b)
	}
}
