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

package karpenter

import (
	"fmt"
	"sort"
	"strings"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/framework"
	karpenterv1 "sigs.k8s.io/karpenter/pkg/apis/v1"
	karpentercloudprovider "sigs.k8s.io/karpenter/pkg/cloudprovider"
	karpevents "sigs.k8s.io/karpenter/pkg/events"
)

// ConversionResult encapsulates the converted Karpenter primitives and pre-indexed query maps.
type ConversionResult struct {
	NodePools          []*karpenterv1.NodePool
	InstanceTypes      map[string][]*karpentercloudprovider.InstanceType
	OfferingMap        map[string]*karpentercloudprovider.Offering
	PoolITToNodeGroups map[string]map[string][]cloudprovider.NodeGroup
	NodeGroupToPool    map[string]string
}

// NodeGroupsFor returns the CA NodeGroups corresponding to a given virtual NodePool and InstanceType.
func (r *ConversionResult) NodeGroupsFor(nodePoolName, itName string) []cloudprovider.NodeGroup {
	if r == nil || r.PoolITToNodeGroups == nil {
		return nil
	}
	if itMap, ok := r.PoolITToNodeGroups[nodePoolName]; ok {
		return itMap[itName]
	}
	return nil
}

// PoolForNodeGroup returns the virtual NodePool name for a given CA NodeGroup ID.
func (r *ConversionResult) PoolForNodeGroup(ngId string) string {
	if r == nil || r.NodeGroupToPool == nil {
		return ""
	}
	return r.NodeGroupToPool[ngId]
}

// KarpenterConverter translates CA NodeGroups into Karpenter primitives.
type KarpenterConverter interface {
	Convert(nodeGroups []cloudprovider.NodeGroup, nodeInfos map[string]*framework.NodeInfo) (*ConversionResult, error)
}

// SerializeTaints creates a deterministic string key from a slice of taints for node pool grouping.
func SerializeTaints(taints []apiv1.Taint) string {
	var parts []string
	for _, t := range taints {
		parts = append(parts, fmt.Sprintf("%s=%s:%s", t.Key, t.Value, t.Effect))
	}
	sort.Strings(parts)
	return strings.Join(parts, ",")
}

// InstanceTypeNameFromLabels returns the physical instance type name from node labels or the defaultName fallback.
func InstanceTypeNameFromLabels(labels map[string]string, defaultName string) string {
	if val, ok := labels[apiv1.LabelInstanceTypeStable]; ok && val != "" {
		return val
	}
	if val, ok := labels[apiv1.LabelInstanceType]; ok && val != "" {
		return val
	}
	return defaultName
}

// NoopRecorder is a dummy implementation of Karpenter events.Recorder.
type NoopRecorder struct{}

// Publish implements events.Recorder.
func (n *NoopRecorder) Publish(_ ...karpevents.Event) {}
