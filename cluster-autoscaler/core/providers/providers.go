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
	"time"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/context"
	"k8s.io/autoscaler/cluster-autoscaler/processors/nodegroupconfig"
)

// MaxNodeProvisionTimeProvider provides maximum time for stages of node provisioning:
// - end to end
// - node registration only
type MaxNodeProvisionTimeProvider interface {
	// Initialize performs initialization of MaxNodeProvisionTimeProvider
	Initialize(context *context.AutoscalingContext, nodeGroupConfigProcessor nodegroupconfig.NodeGroupConfigProcessor)
	// GetMaxNodeProvisionTime is a time a node has to register since its creation started
	GetMaxNodeProvisionTime(nodeGroup cloudprovider.NodeGroup) (time.Duration, error)
	// GetMaxNodeRegisterTime is a time a node has to register, starting from when its creation finished
	GetMaxNodeRegisterTime(nodeGroup cloudprovider.NodeGroup) (time.Duration, error)
}
