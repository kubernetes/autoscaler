/*
Copyright 2025 The Kubernetes Authors.

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

package provider

import (
	storagev1 "k8s.io/api/storage/v1"
	"k8s.io/apimachinery/pkg/labels"
	csisnapshot "k8s.io/autoscaler/cluster-autoscaler/simulator/csi/snapshot"
	informers "k8s.io/client-go/informers"
	v1storagelister "k8s.io/client-go/listers/storage/v1"
)

// Provider provides access to CSI node information for the cluster.
type Provider struct {
	csiNodesLister v1storagelister.CSINodeLister
}

// NewCSINodeProvider creates a new Provider with the given CSI node lister.
func NewCSINodeProvider(csiNodesLister v1storagelister.CSINodeLister) *Provider {
	return &Provider{csiNodesLister: csiNodesLister}
}

// NewCSINodeProviderFromInformers creates a new Provider from an informer factory.
func NewCSINodeProviderFromInformers(informerFactory informers.SharedInformerFactory) *Provider {
	return NewCSINodeProvider(informerFactory.Storage().V1().CSINodes().Lister())
}

// Snapshot returns a snapshot of the CSI node information.
func (p *Provider) Snapshot() (*csisnapshot.Snapshot, error) {
	csiNodes, err := p.csiNodesLister.List(labels.Everything())
	if err != nil {
		return nil, err
	}

	csiNodesMap := make(map[string]*storagev1.CSINode)
	for _, csiNode := range csiNodes {
		csiNodesMap[csiNode.Name] = csiNode
	}
	return csisnapshot.NewSnapshot(csiNodesMap), nil

}
