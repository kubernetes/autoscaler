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

	"k8s.io/klog/v2"
	"sigs.k8s.io/cluster-autoscaler/pkg/cloudprovider"
	"sigs.k8s.io/cluster-autoscaler/pkg/metrics"
	"sigs.k8s.io/cluster-autoscaler/pkg/simulator/framework"
	"sigs.k8s.io/cluster-autoscaler/pkg/utils/dynamicresources"
	"sigs.k8s.io/cluster-autoscaler/pkg/utils/gpu"
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
	RegisterScaleUp(delta int, gpuResourceName, gpuType, draDriverName string)
	RegisterFailedScaleUp(reason metrics.FailedScaleUpReason, gpuResourceName, gpuType, draDriverName string)
	RegisterFailedNodeCreations(reason metrics.FailedScaleUpReason, nodesCount int)
}

type nodeInfoRegistry interface {
	GetNodeInfo(id string) (*framework.NodeInfo, bool)
}

// NodeGroupChangeMetricsProducer is an implementation of NodeGroupChangeObserver for reporting the scale up/down metrics
type NodeGroupChangeMetricsProducer struct {
	cloudProvider cloudprovider.CloudProvider
	// metrics is an instance of metricObserver interface which allows to mock and test the nodegroupchange metrics
	metrics          metricObserver
	nodeInfoRegistry nodeInfoRegistry
}

// RegisterScaleUp emits the scale up metric.
func (p *NodeGroupChangeMetricsProducer) RegisterScaleUp(ctx context.Context, nodeGroup cloudprovider.NodeGroup,
	delta int, currentTime time.Time) {
	gpuResourceName, gpuType, draDriverNames := retrieveMetricsDataFromNodeGroup(ctx, nodeGroup, p.cloudProvider, p.nodeInfoRegistry)
	p.metrics.RegisterScaleUp(delta, gpuResourceName, gpuType, draDriverNames)
}

// RegisterScaleDown is a no-op for NodeGroupChangeMetricsProducer.
func (p *NodeGroupChangeMetricsProducer) RegisterScaleDown(nodeGroup cloudprovider.NodeGroup,
	nodeName string, currentTime time.Time, expectedDeleteTime time.Time) {
}

// RegisterFailedScaleUp emits the failed scale up metric.
func (p *NodeGroupChangeMetricsProducer) RegisterFailedScaleUp(ctx context.Context, nodeGroup cloudprovider.NodeGroup, delta int, errorInfo cloudprovider.InstanceErrorInfo, currentTime time.Time) {
	gpuResourceName, gpuType, draDriverNames := retrieveMetricsDataFromNodeGroup(ctx, nodeGroup, p.cloudProvider, p.nodeInfoRegistry)
	reason := metrics.FailedScaleUpReason(errorInfo.ErrorCode)
	p.metrics.RegisterFailedScaleUp(reason, gpuResourceName, gpuType, draDriverNames)
	p.metrics.RegisterFailedNodeCreations(reason, delta)
}

// RegisterFailedScaleDown is a no-op for NodeGroupChangeMetricsProducer.
func (p *NodeGroupChangeMetricsProducer) RegisterFailedScaleDown(nodeGroup cloudprovider.NodeGroup,
	reason string, currentTime time.Time) {
}

// NewNodeGroupChangeMetricsProducer returns a new NodeGroupChangeMetricsProducer.
func NewNodeGroupChangeMetricsProducer(cloudProvider cloudprovider.CloudProvider, metrics metricObserver, nodeInfoRegistry nodeInfoRegistry) *NodeGroupChangeMetricsProducer {
	return &NodeGroupChangeMetricsProducer{cloudProvider: cloudProvider, metrics: metrics, nodeInfoRegistry: nodeInfoRegistry}
}

func retrieveMetricsDataFromNodeGroup(ctx context.Context, nodeGroup cloudprovider.NodeGroup, cloudProvider cloudprovider.CloudProvider, registry nodeInfoRegistry) (string, string, string) {
	logger := klog.FromContext(ctx)
	availableGPUTypes := cloudProvider.GetAvailableGPUTypes(ctx)
	gpuResourceName, gpuType, draDriverNames := "", "", ""
	var nodeInfo *framework.NodeInfo
	var found bool
	if registry != nil {
		nodeInfo, found = registry.GetNodeInfo(nodeGroup.Id())
	}
	if !found {
		var err error
		nodeInfo, err = nodeGroup.TemplateNodeInfo(ctx)
		if err != nil {
			logger.Info("Failed to get template node info for a node group", "err", err)
		} else if nodeInfo == nil {
			logger.Info("Template node info for node group is nil", "nodeGroupId", nodeGroup.Id())
		}
	}
	if nodeInfo != nil {
		gpuResourceName, gpuType = gpu.GetGpuInfoForMetrics(ctx, cloudProvider.GetNodeGpuConfig(ctx, nodeInfo.Node()), availableGPUTypes, nodeInfo.Node(), nodeGroup)
		draDriverNames = dynamicresources.GetDriverNamesForMetricsCompacted(nodeInfo.LocalResourceSlices)
	}
	return gpuResourceName, gpuType, draDriverNames
}
