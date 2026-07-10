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
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/framework"
	karpenterv1 "sigs.k8s.io/karpenter/pkg/apis/v1"
	karpentercloudprovider "sigs.k8s.io/karpenter/pkg/cloudprovider"
	karpevents "sigs.k8s.io/karpenter/pkg/events"
)

// KarpenterConverter translates CA NodeGroups into Karpenter primitives.
type KarpenterConverter interface {
	Convert(nodeGroups []cloudprovider.NodeGroup, nodeInfos map[string]*framework.NodeInfo) ([]*karpenterv1.NodePool, map[string][]*karpentercloudprovider.InstanceType)
	ITNameToNodeGroups() map[string][]cloudprovider.NodeGroup
	ITNameToPool() map[string]string
	GetPhysicalITName(labels map[string]string, defaultName string) string
}

// NoopRecorder is a dummy implementation of Karpenter events.Recorder.
type NoopRecorder struct{}

// Publish implements events.Recorder.
func (n *NoopRecorder) Publish(_ ...karpevents.Event) {}
