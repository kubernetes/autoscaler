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

package actionablecluster

import (
	"time"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/autoscaler/cluster-autoscaler/clusterstate/utils"
	"k8s.io/autoscaler/cluster-autoscaler/context"
	"k8s.io/autoscaler/cluster-autoscaler/metrics"
	"k8s.io/autoscaler/cluster-autoscaler/utils/errors"
	"k8s.io/klog/v2"
)

// NodesNotReadyAfterStartTimeout How long should Cluster Autoscaler wait for nodes to become ready after start.
const NodesNotReadyAfterStartTimeout = 10 * time.Minute

// ActionableClusterProcessor defines interface for whether action can be taken on the cluster or not
type ActionableClusterProcessor interface {
	// ShouldAbort is the func that will return where the cluster is in an actionable state or not
	ShouldAbort(context *context.AutoscalingContext, allNodes []*apiv1.Node, readyNodes []*apiv1.Node, currentTime time.Time) (abortLoop bool, err errors.AutoscalerError)
	CleanUp()
}

// EmptyClusterProcessor implements ActionableClusterProcessor by using ScaleUpFromZero flag
// and other node readiness conditions to check whether the cluster is actionable or not
type EmptyClusterProcessor struct {
	nodesNotReadyAfterStartTimeout time.Duration
	startTime                      time.Time
}

// ShouldAbort give the decision on whether CA can act on the cluster
func (e *EmptyClusterProcessor) ShouldAbort(context *context.AutoscalingContext, allNodes []*apiv1.Node, readyNodes []*apiv1.Node, currentTime time.Time) (bool, errors.AutoscalerError) {
	if context.AutoscalingOptions.ScaleUpFromZero {
		return false, nil
	}
	if len(allNodes) == 0 {
		OnEmptyCluster(context, "Cluster has no nodes.", true)
		return true, nil
	}
	if len(readyNodes) == 0 {
		// Cluster Autoscaler may start running before nodes are ready.
		// Timeout ensures no ClusterUnhealthy events are published immediately in this case.
		OnEmptyCluster(context, "Cluster has no ready nodes.", currentTime.After(e.startTime.Add(e.nodesNotReadyAfterStartTimeout)))
		return true, nil
	}
	// the cluster is not empty
	return false, nil
}

// OnEmptyCluster runs actions if the cluster is empty
func OnEmptyCluster(context *context.AutoscalingContext, status string, emitEvent bool) {
	klog.Warningf(status)
	context.ProcessorCallbacks.ResetUnneededNodes()
	// updates metrics related to empty cluster's state.
	metrics.UpdateClusterSafeToAutoscale(false)
	metrics.UpdateNodesCount(0, 0, 0, 0, 0)
	if context.WriteStatusConfigMap {
		utils.WriteStatusConfigMap(context.ClientSet, context.ConfigNamespace, status, context.LogRecorder, context.StatusConfigMapName)
	}
	if emitEvent {
		context.LogRecorder.Eventf(apiv1.EventTypeWarning, "ClusterUnhealthy", status)
	}
}

// CleanUp cleans up the Processor
func (e *EmptyClusterProcessor) CleanUp() {
}

// NewDefaultActionableClusterProcessor returns a new Processor instance
func NewDefaultActionableClusterProcessor() ActionableClusterProcessor {
	return NewCustomActionableClusterProcessor(time.Now(), NodesNotReadyAfterStartTimeout)
}

// NewCustomActionableClusterProcessor returns a new instance with custom values
func NewCustomActionableClusterProcessor(startTime time.Time, nodeNotReadyTimeout time.Duration) ActionableClusterProcessor {
	return &EmptyClusterProcessor{
		nodesNotReadyAfterStartTimeout: nodeNotReadyTimeout,
		startTime:                      startTime,
	}
}
