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

package nodegroupchange

import (
	"context"
	"sync"
	"time"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/metrics"
	"k8s.io/autoscaler/cluster-autoscaler/utils/dynamicresources"
	"k8s.io/autoscaler/cluster-autoscaler/utils/gpu"
	"k8s.io/klog/v2"
)

// NodeGroupChangeObserver is an observer of:
// * scale-up(s) for a nodegroup
// * scale-down(s) for a nodegroup
// * scale-up failure(s) for a nodegroup
// * scale-down failure(s) for a nodegroup
type NodeGroupChangeObserver interface {
	// RegisterScaleUp records scale up for a nodegroup.
	RegisterScaleUp(ctx context.Context, nodeGroup cloudprovider.NodeGroup, delta int, currentTime time.Time)
	// RegisterScaleDowns records scale down for a nodegroup.
	RegisterScaleDown(nodeGroup cloudprovider.NodeGroup, nodeName string, currentTime time.Time, expectedDeleteTime time.Time)
	// RegisterFailedScaleUp records failed scale-up for a nodegroup.
	// errorInfo is a wrapper containing the reason for failed scale-up and the actual error message
	RegisterFailedScaleUp(ctx context.Context, nodeGroup cloudprovider.NodeGroup, delta int, errorInfo cloudprovider.InstanceErrorInfo, currentTime time.Time)
	// RegisterFailedScaleDown records failed scale-down for a nodegroup.
	RegisterFailedScaleDown(nodeGroup cloudprovider.NodeGroup, reason string, currentTime time.Time)
}

// NodeGroupChangeObserversList is a slice of observers
// of state of scale up/down in the cluster
type NodeGroupChangeObserversList struct {
	observers []NodeGroupChangeObserver
	// TODO(vadasambar): consider using separate mutexes for functions not related to each other
	mutex sync.Mutex
}

// Register adds new observer to the list.
func (l *NodeGroupChangeObserversList) Register(o NodeGroupChangeObserver) {
	l.observers = append(l.observers, o)
}

// RegisterScaleUp calls RegisterScaleUp for each observer.
func (l *NodeGroupChangeObserversList) RegisterScaleUp(ctx context.Context, nodeGroup cloudprovider.NodeGroup,
	delta int, currentTime time.Time) {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	for _, observer := range l.observers {
		observer.RegisterScaleUp(ctx, nodeGroup, delta, currentTime)
	}
}

// RegisterScaleDown calls RegisterScaleDown for each observer.
func (l *NodeGroupChangeObserversList) RegisterScaleDown(nodeGroup cloudprovider.NodeGroup,
	nodeName string, currentTime time.Time, expectedDeleteTime time.Time) {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	for _, observer := range l.observers {
		observer.RegisterScaleDown(nodeGroup, nodeName, currentTime, expectedDeleteTime)
	}
}

// RegisterFailedScaleUp calls RegisterFailedScaleUp for each observer.
func (l *NodeGroupChangeObserversList) RegisterFailedScaleUp(ctx context.Context, nodeGroup cloudprovider.NodeGroup, delta int, errorInfo cloudprovider.InstanceErrorInfo, currentTime time.Time) {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	for _, observer := range l.observers {
		observer.RegisterFailedScaleUp(ctx, nodeGroup, delta, errorInfo, currentTime)
	}
}

// RegisterFailedScaleDown records failed scale-down for a nodegroup.
func (l *NodeGroupChangeObserversList) RegisterFailedScaleDown(nodeGroup cloudprovider.NodeGroup,
	reason string, currentTime time.Time) {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	for _, observer := range l.observers {
		observer.RegisterFailedScaleDown(nodeGroup, reason, currentTime)
	}
}

// NewNodeGroupChangeObserversList return empty list of scale state observers.
func NewNodeGroupChangeObserversList() *NodeGroupChangeObserversList {
	return &NodeGroupChangeObserversList{}
}

type metricObserver interface {
	RegisterFailedScaleUp(reason metrics.FailedScaleUpReason, gpuResourceName, gpuType, draDriverName string)
	RegisterFailedNodeCreations(reason metrics.FailedScaleUpReason, nodesCount int)
}

// NodeGroupChangeMetricsProducer is an implementation of NodeGroupChangeObserver for reporting the scale up/down metrics
type NodeGroupChangeMetricsProducer struct {
	cloudProvider cloudprovider.CloudProvider
	// metrics is an instance of metricObserver interface which allows to mock and test the nodegroupchange metrics
	metrics metricObserver
}

// RegisterScaleUp calls RegisterScaleUp for each observer.
func (p *NodeGroupChangeMetricsProducer) RegisterScaleUp(ctx context.Context, nodeGroup cloudprovider.NodeGroup,
	delta int, currentTime time.Time) {
}

// RegisterScaleDown calls RegisterScaleDown for each observer.
func (p *NodeGroupChangeMetricsProducer) RegisterScaleDown(nodeGroup cloudprovider.NodeGroup,
	nodeName string, currentTime time.Time, expectedDeleteTime time.Time) {
}

// RegisterFailedScaleUp emits the failed scale up metric.
func (p *NodeGroupChangeMetricsProducer) RegisterFailedScaleUp(ctx context.Context, nodeGroup cloudprovider.NodeGroup, delta int, errorInfo cloudprovider.InstanceErrorInfo, currentTime time.Time) {
	logger := klog.FromContext(ctx)
	availableGPUTypes := p.cloudProvider.GetAvailableGPUTypes()
	gpuResourceName, gpuType, draDriverNames := "", "", ""
	nodeInfo, err := nodeGroup.TemplateNodeInfo()
	if err != nil {
		logger.Info("Failed to get template node info for a node group", "err", err)
	} else if nodeInfo == nil {
		logger.Info("Template node info is nil for node group", "nodeGroupId", nodeGroup.Id())
	} else {
		gpuResourceName, gpuType = gpu.GetGpuInfoForMetrics(ctx, p.cloudProvider.GetNodeGpuConfig(nodeInfo.Node()), availableGPUTypes, nodeInfo.Node(), nodeGroup)
		draDriverNames = dynamicresources.GetDriverNamesForMetricsCompacted(nodeInfo.LocalResourceSlices)
	}
	reason := metrics.FailedScaleUpReason(errorInfo.ErrorCode)
	p.metrics.RegisterFailedScaleUp(reason, gpuResourceName, gpuType, draDriverNames)
	p.metrics.RegisterFailedNodeCreations(reason, delta)
}

// RegisterFailedScaleDown records failed scale-down for a nodegroup.
func (p *NodeGroupChangeMetricsProducer) RegisterFailedScaleDown(nodeGroup cloudprovider.NodeGroup,
	reason string, currentTime time.Time) {
}

// NewNodeGroupChangeMetricsProducer returns a new NodeGroupChangeMetricsProducer.
func NewNodeGroupChangeMetricsProducer(cloudProvider cloudprovider.CloudProvider, metrics metricObserver) *NodeGroupChangeMetricsProducer {
	return &NodeGroupChangeMetricsProducer{cloudProvider: cloudProvider, metrics: metrics}
}
