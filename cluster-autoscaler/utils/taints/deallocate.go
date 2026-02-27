/*
Copyright 2024 The Kubernetes Authors.

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

package taints

import (
	apiv1 "k8s.io/api/core/v1"
	cloudproviderapi "k8s.io/cloud-provider/api"
)

// HasShutdownTaint returns true if cloudprovider node shutdown taint is applied on the node.
func HasShutdownTaint(node *apiv1.Node) bool {
	return HasTaint(node, cloudproviderapi.TaintNodeShutdown)
}

// HasUnreachableTaint returns true if unreachable taint is applied on the node.
func HasUnreachableTaint(node *apiv1.Node) bool {
	return HasTaint(node, apiv1.TaintNodeUnreachable)
}
