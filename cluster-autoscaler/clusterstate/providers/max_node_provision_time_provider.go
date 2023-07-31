/*
Copyright 2016 The Kubernetes Authors.

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

package providers

import (
	"fmt"
	"time"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/context"
	"k8s.io/autoscaler/cluster-autoscaler/processors/nodegroupconfig"
)

// NewDefaultMaxNodeProvisionTimeProvider returns the default maxNodeProvisionTimeProvider which uses the NodeGroupConfigProcessor.
func NewDefaultMaxNodeProvisionTimeProvider() *defultMaxNodeProvisionTimeProvider {
	return &defultMaxNodeProvisionTimeProvider{}
}

type defultMaxNodeProvisionTimeProvider struct {
	initialized              bool
	context                  *context.AutoscalingContext
	nodeGroupConfigProcessor nodegroupconfig.NodeGroupConfigProcessor
}

// Initialize initializes defultMaxNodeProvisionTimeProvider
func (p *defultMaxNodeProvisionTimeProvider) Initialize(context *context.AutoscalingContext, nodeGroupConfigProcessor nodegroupconfig.NodeGroupConfigProcessor) {
	p.context = context
	p.nodeGroupConfigProcessor = nodeGroupConfigProcessor
	p.initialized = true
}

// GetMaxNodeProvisionTime is a time a node has to register since its creation started
func (p *defultMaxNodeProvisionTimeProvider) GetMaxNodeProvisionTime(nodeGroup cloudprovider.NodeGroup) (time.Duration, error) {
	if !p.initialized {
		return 0, fmt.Errorf("defultMaxNodeProvisionTimeProvider uninitialized")
	}
	return p.nodeGroupConfigProcessor.GetMaxNodeProvisionTime(p.context, nodeGroup)
}

// GetMaxNodeRegisterTime is a time a node has to register, starting from when its creation finished
func (p *defultMaxNodeProvisionTimeProvider) GetMaxNodeRegisterTime(nodeGroup cloudprovider.NodeGroup) (time.Duration, error) {
	if !p.initialized {
		return 0, fmt.Errorf("defultMaxNodeProvisionTimeProvider uninitialized")
	}
	return p.nodeGroupConfigProcessor.GetMaxNodeProvisionTime(p.context, nodeGroup)
}
