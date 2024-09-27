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

package backoff

import (
	"time"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/framework"
)

// Status contains information about back off status.
type Status struct {
	IsBackedOff bool
	ErrorInfo   cloudprovider.InstanceErrorInfo
}

// Backoff allows time-based backing off of node groups considered in scale up algorithm
type Backoff interface {
	// Backoff execution for the given node group. Returns time till execution is backed off.
	Backoff(nodeGroup cloudprovider.NodeGroup, nodeInfo *framework.NodeInfo, errorInfo cloudprovider.InstanceErrorInfo, currentTime time.Time) time.Time
	// BackoffStatus returns whether the execution is backed off for the given node group and error info when the node group is backed off.
	BackoffStatus(nodeGroup cloudprovider.NodeGroup, nodeInfo *framework.NodeInfo, currentTime time.Time) Status
	// RemoveBackoff removes backoff data for the given node group.
	RemoveBackoff(nodeGroup cloudprovider.NodeGroup, nodeInfo *framework.NodeInfo)
	// RemoveStaleBackoffData removes stale backoff data.
	RemoveStaleBackoffData(currentTime time.Time)
}
