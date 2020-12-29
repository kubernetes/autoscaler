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
	"time"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/context"
)

// NodeGroupConfigProcessor provides config values for a particular NodeGroup.
type NodeGroupConfigProcessor interface {
	// Process processes a map of nodeInfos for node groups.
	GetScaleDownUnneededTime(context *context.AutoscalingContext, nodeGroup cloudprovider.NodeGroup) (time.Duration, error)
	// CleanUp cleans up processor's internal structures.
	CleanUp()
}

// DelegatingNodeGroupConfigProcessor calls NodeGroup.GetOptions to get config
// for each NodeGroup. If NodeGroup doesn't return a value default config is
// used instead.
type DelegatingNodeGroupConfigProcessor struct {
}

// GetScaleDownUnneededTime returns ScaleDownUnneededTime value that should be used for a given NodeGroup.
func (p *DelegatingNodeGroupConfigProcessor) GetScaleDownUnneededTime(context *context.AutoscalingContext, nodeGroup cloudprovider.NodeGroup) (time.Duration, error) {
	ngConfig, err := nodeGroup.GetOptions(context.NodeGroupAutoscalingOptions)
	if err != nil && err != cloudprovider.ErrNotImplemented {
		return time.Duration(0), err
	}
	if ngConfig == nil || err == cloudprovider.ErrNotImplemented {
		return context.ScaleDownUnneededTime, nil
	}
	return ngConfig.ScaleDownUnneededTime, nil
}

// CleanUp cleans up processor's internal structures.
func (p *DelegatingNodeGroupConfigProcessor) CleanUp() {
}

// NewDefaultNodeGroupConfigProcessor returns a default instance of NodeGroupConfigProcessor.
func NewDefaultNodeGroupConfigProcessor() NodeGroupConfigProcessor {
	return &DelegatingNodeGroupConfigProcessor{}
}
