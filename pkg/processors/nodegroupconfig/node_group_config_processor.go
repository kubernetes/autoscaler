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
	"context"
	"time"

	"sigs.k8s.io/cluster-autoscaler/pkg/cloudprovider"
	"sigs.k8s.io/cluster-autoscaler/pkg/config"
)

// NodeGroupConfigProcessor provides config values for a particular NodeGroup.
type NodeGroupConfigProcessor interface {
	// GetScaleDownUnneededTime returns ScaleDownUnneededTime value that should be used for a given NodeGroup.
	GetScaleDownUnneededTime(ctx context.Context, nodeGroup cloudprovider.NodeGroup) (time.Duration, error)
	// GetScaleDownUnreadyTime returns ScaleDownUnreadyTime value that should be used for a given NodeGroup.
	GetScaleDownUnreadyTime(ctx context.Context, nodeGroup cloudprovider.NodeGroup) (time.Duration, error)
	// GetScaleDownUtilizationThreshold returns ScaleDownUtilizationThreshold value that should be used for a given NodeGroup.
	GetScaleDownUtilizationThreshold(ctx context.Context, nodeGroup cloudprovider.NodeGroup) (float64, error)
	// GetScaleDownGpuUtilizationThreshold returns ScaleDownGpuUtilizationThreshold value that should be used for a given NodeGroup.
	GetScaleDownGpuUtilizationThreshold(ctx context.Context, nodeGroup cloudprovider.NodeGroup) (float64, error)
	// GetMaxNodeProvisionTime return MaxNodeProvisionTime value that should be used for a given NodeGroup.
	GetMaxNodeProvisionTime(ctx context.Context, nodeGroup cloudprovider.NodeGroup) (time.Duration, error)
	// GetMaxNodeStartupTime return MaxNodeStartupTime value that should be used for a given NodeGroup.
	GetMaxNodeStartupTime(ctx context.Context, nodeGroup cloudprovider.NodeGroup) (time.Duration, error)
	// GetIgnoreDaemonSetsUtilization returns IgnoreDaemonSetsUtilization value that should be used for a given NodeGroup.
	GetIgnoreDaemonSetsUtilization(ctx context.Context, nodeGroup cloudprovider.NodeGroup) (bool, error)
	// CleanUp cleans up processor's internal structures.
	CleanUp()
}

// DelegatingNodeGroupConfigProcessor calls NodeGroup.GetOptions to get config
// for each NodeGroup. If NodeGroup doesn't return a value default config is
// used instead.
type DelegatingNodeGroupConfigProcessor struct {
	nodeGroupDefaults config.NodeGroupAutoscalingOptions
}

// GetScaleDownUnneededTime returns ScaleDownUnneededTime value that should be used for a given NodeGroup.
func (p *DelegatingNodeGroupConfigProcessor) GetScaleDownUnneededTime(ctx context.Context, nodeGroup cloudprovider.NodeGroup) (time.Duration, error) {
	ngConfig, err := nodeGroup.GetOptions(ctx, p.nodeGroupDefaults)
	if err != nil && err != cloudprovider.ErrNotImplemented {
		return time.Duration(0), err
	}
	if ngConfig == nil || err == cloudprovider.ErrNotImplemented {
		return p.nodeGroupDefaults.ScaleDownUnneededTime, nil
	}
	return ngConfig.ScaleDownUnneededTime, nil
}

// GetScaleDownUnreadyTime returns ScaleDownUnreadyTime value that should be used for a given NodeGroup.
func (p *DelegatingNodeGroupConfigProcessor) GetScaleDownUnreadyTime(ctx context.Context, nodeGroup cloudprovider.NodeGroup) (time.Duration, error) {
	ngConfig, err := nodeGroup.GetOptions(ctx, p.nodeGroupDefaults)
	if err != nil && err != cloudprovider.ErrNotImplemented {
		return time.Duration(0), err
	}
	if ngConfig == nil || err == cloudprovider.ErrNotImplemented {
		return p.nodeGroupDefaults.ScaleDownUnreadyTime, nil
	}
	return ngConfig.ScaleDownUnreadyTime, nil
}

// GetScaleDownUtilizationThreshold returns ScaleDownUtilizationThreshold value that should be used for a given NodeGroup.
func (p *DelegatingNodeGroupConfigProcessor) GetScaleDownUtilizationThreshold(ctx context.Context, nodeGroup cloudprovider.NodeGroup) (float64, error) {
	ngConfig, err := nodeGroup.GetOptions(ctx, p.nodeGroupDefaults)
	if err != nil && err != cloudprovider.ErrNotImplemented {
		return 0.0, err
	}
	if ngConfig == nil || err == cloudprovider.ErrNotImplemented {
		return p.nodeGroupDefaults.ScaleDownUtilizationThreshold, nil
	}
	return ngConfig.ScaleDownUtilizationThreshold, nil
}

// GetScaleDownGpuUtilizationThreshold returns ScaleDownGpuUtilizationThreshold value that should be used for a given NodeGroup.
func (p *DelegatingNodeGroupConfigProcessor) GetScaleDownGpuUtilizationThreshold(ctx context.Context, nodeGroup cloudprovider.NodeGroup) (float64, error) {
	ngConfig, err := nodeGroup.GetOptions(ctx, p.nodeGroupDefaults)
	if err != nil && err != cloudprovider.ErrNotImplemented {
		return 0.0, err
	}
	if ngConfig == nil || err == cloudprovider.ErrNotImplemented {
		return p.nodeGroupDefaults.ScaleDownGpuUtilizationThreshold, nil
	}
	return ngConfig.ScaleDownGpuUtilizationThreshold, nil
}

// GetMaxNodeProvisionTime returns MaxNodeProvisionTime value that should be used for a given NodeGroup.
func (p *DelegatingNodeGroupConfigProcessor) GetMaxNodeProvisionTime(ctx context.Context, nodeGroup cloudprovider.NodeGroup) (time.Duration, error) {
	ngConfig, err := nodeGroup.GetOptions(ctx, p.nodeGroupDefaults)
	if err != nil && err != cloudprovider.ErrNotImplemented {
		return time.Duration(0), err
	}
	if ngConfig == nil || err == cloudprovider.ErrNotImplemented {
		return p.nodeGroupDefaults.MaxNodeProvisionTime, nil
	}
	return ngConfig.MaxNodeProvisionTime, nil
}

// GetMaxNodeStartupTime returns MaxNodeStartupTime value that should be used for a given NodeGroup.
func (p *DelegatingNodeGroupConfigProcessor) GetMaxNodeStartupTime(ctx context.Context, nodeGroup cloudprovider.NodeGroup) (time.Duration, error) {
	ngConfig, err := nodeGroup.GetOptions(ctx, p.nodeGroupDefaults)
	if err != nil && err != cloudprovider.ErrNotImplemented {
		return p.nodeGroupDefaults.MaxNodeStartupTime, err
	}
	if ngConfig == nil || err == cloudprovider.ErrNotImplemented {
		return p.nodeGroupDefaults.MaxNodeStartupTime, nil
	}
	return ngConfig.MaxNodeStartupTime, nil
}

// GetIgnoreDaemonSetsUtilization returns IgnoreDaemonSetsUtilization value that should be used for a given NodeGroup.
func (p *DelegatingNodeGroupConfigProcessor) GetIgnoreDaemonSetsUtilization(ctx context.Context, nodeGroup cloudprovider.NodeGroup) (bool, error) {
	ngConfig, err := nodeGroup.GetOptions(ctx, p.nodeGroupDefaults)
	if err != nil && err != cloudprovider.ErrNotImplemented {
		return false, err
	}
	if ngConfig == nil || err == cloudprovider.ErrNotImplemented {
		return p.nodeGroupDefaults.IgnoreDaemonSetsUtilization, nil
	}
	return ngConfig.IgnoreDaemonSetsUtilization, nil
}

// CleanUp cleans up processor's internal structures.
func (p *DelegatingNodeGroupConfigProcessor) CleanUp() {
}

// NewDefaultNodeGroupConfigProcessor returns a default instance of NodeGroupConfigProcessor.
func NewDefaultNodeGroupConfigProcessor(nodeGroupDefaults config.NodeGroupAutoscalingOptions) NodeGroupConfigProcessor {
	return &DelegatingNodeGroupConfigProcessor{
		nodeGroupDefaults: nodeGroupDefaults,
	}
}
