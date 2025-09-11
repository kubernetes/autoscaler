package provider

import (
	storagev1 "k8s.io/api/storage/v1"
	"k8s.io/apimachinery/pkg/labels"
	csisnapshot "k8s.io/autoscaler/cluster-autoscaler/simulator/csi/snapshot"
	informers "k8s.io/client-go/informers"
	v1storagelister "k8s.io/client-go/listers/storage/v1"
)

type Provider struct {
	csINodesLister v1storagelister.CSINodeLister
}

func NewCSINodeProvider(csINodesLister v1storagelister.CSINodeLister) *Provider {
	return &Provider{csINodesLister: csINodesLister}
}

func NewCSINodeProviderFromInformers(informerFactory informers.SharedInformerFactory) *Provider {
	return NewCSINodeProvider(informerFactory.Storage().V1().CSINodes().Lister())
}

func (p *Provider) Snapshot() (*csisnapshot.Snapshot, error) {
	csiNodes, err := p.csINodesLister.List(labels.Everything())
	if err != nil {
		return nil, err
	}

	csiNodesMap := make(map[string]*storagev1.CSINode)
	for _, csiNode := range csiNodes {
		csiNodesMap[csiNode.Name] = csiNode
	}
	return csisnapshot.NewSnapshot(csiNodesMap), nil

}
