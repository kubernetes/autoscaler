/*
Copyright 2021 The Kubernetes Authors.

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

/*
  This hack completes the other sub-processor that add storageclass/local-data
  custom resources requests to pods requesting local-data.

  Use case being addressed here is: when a node with local-storage just joined,
  its PV isn't immediately available, and pods that triggered the upscale will
  remain pending for a moment. The autoscaler might run during that window,
  evaluate those pods against the just joined nodes, and would consider them
  unschedulable as the fresh real nodes don't have the custom local-data
  resource allocatable (as op. to virtual nodes built from ASG templates).
  Which would cause a spurious re-upscale.

  Decorating those fresh nodes with the requested resources they were expected
  to offer when created will help the autoscaler to consider them usable for
  still pending pods, which then won't be flagged as unschedulable and won't
  trigger an other upscale.

  Theorically all nodes labeled local-storage:true could be injected that custom
  resource, for a better accuracy. But side effects of that generalisation
  needs to be evaluated carefully. For instance, absence of the custom resource
  prevents the autoscaler to even try to repack stateful pods (as they wouldn't
  fit anywhere). That's why we limit the change to recent local-data nodes for
  now (later: evaluate if we can safely apply to all local-data real nodes).

  This replaces a previous workaround that did set those new nodes as "unready"
  until an LVP pod was running there for 90s, but in a way that don't require
  altering core, and better mimic normal nodes behaviour.
*/

package pods

import (
	"time"

	"k8s.io/autoscaler/cluster-autoscaler/context"
	"k8s.io/autoscaler/cluster-autoscaler/processors/datadog/common"

	apiv1 "k8s.io/api/core/v1"
)

const (
	// NodeReadyGraceDelay is time during which we inject custom resource after a node becomes ready
	NodeReadyGraceDelay = 5 * time.Minute
)

type transformDataNodes struct{}

// NewTransformDataNodes returns a processor injecting local data custom resource
func NewTransformDataNodes() *transformDataNodes {
	return &transformDataNodes{}
}

// CleanUp tears down a transformDataNodes processor
func (p *transformDataNodes) CleanUp() {}

// Process injects local data custom resources to nodes offering local-data storage, that became ready since less than 5mn
func (p *transformDataNodes) Process(ctx *context.AutoscalingContext, pods []*apiv1.Pod) ([]*apiv1.Pod, error) {
	nodeInfos, err := ctx.ClusterSnapshot.NodeInfos().List()
	if err != nil {
		return pods, err
	}

	for _, nodeInfo := range nodeInfos {
		node := nodeInfo.Node()
		if !common.NodeHasLocalData(node) {
			continue
		}

		// TODO: evaluate if that age check is really needed
		readyTimestamp := time.Now()
		for _, condition := range node.Status.Conditions {
			if condition.Type == apiv1.NodeReady && condition.Status == apiv1.ConditionTrue {
				readyTimestamp = condition.LastTransitionTime.Time
			}
		}
		if readyTimestamp.Add(NodeReadyGraceDelay).Before(time.Now()) {
			continue
		}

		common.SetNodeLocalDataResource(nodeInfo)
	}

	return pods, nil
}
