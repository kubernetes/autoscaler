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

package options

import (
	"k8s.io/autoscaler/cluster-autoscaler/config"
)

// NodeDeleteOptions contains various options to customize how draining will behave
type NodeDeleteOptions struct {
	// SkipNodesWithSystemPods is true if nodes with kube-system pods should be
	// deleted (except for DaemonSet or mirror pods).
	SkipNodesWithSystemPods bool
	// SkipNodesWithLocalStorage is true if nodes with pods using local storage
	// (e.g. EmptyDir or HostPath) should be deleted.
	SkipNodesWithLocalStorage bool
	// SkipNodesWithCustomControllerPods is true if nodes with
	// custom-controller-owned pods should be skipped.
	SkipNodesWithCustomControllerPods bool
	// MinReplicaCount determines the minimum number of replicas that a replica
	// set or replication controller should have to allow pod deletion during
	// scale down.
	MinReplicaCount int
}

// NewNodeDeleteOptions returns new node delete options extracted from autoscaling options.
func NewNodeDeleteOptions(opts config.AutoscalingOptions) NodeDeleteOptions {
	return NodeDeleteOptions{
		SkipNodesWithSystemPods:           opts.SkipNodesWithSystemPods,
		SkipNodesWithLocalStorage:         opts.SkipNodesWithLocalStorage,
		SkipNodesWithCustomControllerPods: opts.SkipNodesWithCustomControllerPods,
		MinReplicaCount:                   opts.MinReplicaCount,
	}
}
