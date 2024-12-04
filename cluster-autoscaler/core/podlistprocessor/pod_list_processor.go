/*
Copyright 2022 The Kubernetes Authors.

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

package podlistprocessor

import (
	"k8s.io/autoscaler/cluster-autoscaler/processors/pods"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/framework"
)

// NewDefaultPodListProcessor returns a default implementation of the pod list
// processor, which wraps and sequentially runs other sub-processors.
func NewDefaultPodListProcessor(nodeFilter func(*framework.NodeInfo) bool) *pods.CombinedPodListProcessor {
	return pods.NewCombinedPodListProcessor([]pods.PodListProcessor{
		NewClearTPURequestsPodListProcessor(),
		NewFilterOutExpendablePodListProcessor(),
		NewCurrentlyDrainedNodesPodListProcessor(),
		NewFilterOutSchedulablePodListProcessor(nodeFilter),
		NewFilterOutDaemonSetPodListProcessor(),
	})
}
