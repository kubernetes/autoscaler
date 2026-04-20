package fort

import (
	"fmt"

	"k8s.io/client-go/tools/cache"
)

type TPod struct {
	NodeName string
	Label    string
}

type TService struct {
	Name     string
	SomeData string
}

type TServiceNode struct {
	Node    string
	Service string
	Count   int64
}

func (s *TService) Matches(pod *TPod) bool {
	if len(pod.Label) == 0 || len(s.SomeData) == 0 {
		return false
	}
	return pod.Label[0] == s.SomeData[0]
}

type TNode struct {
	Name    string
	Domains []string
}

type TNodeDomain struct {
	Name   string
	Domain string
}

type TServiceDomain struct {
	Service string
	Domain  string
	Count   int64
}

type DomainCount struct {
	Domain string
	Count  int64
}

type TServiceInfo struct {
	Service      string
	DomainCounts []DomainCount
}

type PodSpreadLiteInfo struct {
	ServiceNodes   CloneableSharedInformerQuery
	NodeDomains    CloneableSharedInformerQuery
	ServiceDomains CloneableSharedInformerQuery
	ServiceInfo    CloneableSharedInformerQuery

	PodUpdates     ManualSharedInformer
	ServiceUpdates ManualSharedInformer
	NodeUpdates    ManualSharedInformer
}

func NewPodSpreadLiteInfo(podInformer, serviceInformer, nodeInformer ManualSharedInformer) *PodSpreadLiteInfo {
	d := &PodSpreadLiteInfo{
		PodUpdates:     podInformer,
		ServiceUpdates: serviceInformer,
		NodeUpdates:    nodeInformer,
	}

	lock := podInformer.GetLockGroup()

	// Compute the number of pods on each node that match each service.
	d.ServiceNodes = QueryInformer(&GroupByJoin[*TServiceNode, *TPod, *TService]{
		Lock: lock,
		Select: func(fields []GroupField) (*TServiceNode, error) {
			return &TServiceNode{
				Service: fields[0].(string),
				Node:    fields[1].(string),
				Count:   fields[2].(int64),
			}, nil
		},

		From: podInformer,
		Join: serviceInformer,

		Where: func(pod *TPod, service *TService) bool {
			return service.Matches(pod)
		},

		GroupBy: func(pod *TPod, service *TService) (any, []GroupField) {
			return [2]string{service.Name, pod.NodeName},
				[]GroupField{
					AnyValue(service.Name),
					AnyValue(pod.NodeName),
					Count(),
				}
		},
	})

	// Create one entry for each node / domain pair.
	d.NodeDomains = QueryInformer(&FlatMap[*TNodeDomain, *TNode]{
		Lock: lock,
		Map: func(node *TNode) ([]*TNodeDomain, error) {
			ret := []*TNodeDomain{}
			for _, d := range node.Domains {
				ret = append(ret, &TNodeDomain{Name: node.Name, Domain: d})
			}
			return ret, nil
		},
		Over: nodeInformer,
	})

	// Join the service/node pod counts with the node/domain pairs.
	d.ServiceDomains = QueryInformer(&GroupByJoin[*TServiceDomain, *TServiceNode, *TNodeDomain]{
		Lock: lock,
		Select: func(fields []GroupField) (*TServiceDomain, error) {
			return &TServiceDomain{
				Service: fields[0].(string),
				Domain:  fields[1].(string),
				Count:   fields[2].(int64),
			}, nil
		},
		From: d.ServiceNodes,
		Join: d.NodeDomains,
		On: func(service *TServiceNode, node *TNodeDomain) any {
			if service != nil {
				return [1]string{service.Node}
			} else {
				return [1]string{node.Name}
			}
		},
		GroupBy: func(service *TServiceNode, node *TNodeDomain) (any, []GroupField) {
			return [2]string{service.Service, node.Domain},
				[]GroupField{
					AnyValue(service.Service),
					AnyValue(node.Domain),
					Sum(service.Count),
				}
		},
	})

	// Flatten the results into one entry per service, with a sub array of domains and their counts.
	d.ServiceInfo = QueryInformer(&GroupBy[*TServiceInfo, *TServiceDomain]{
		Lock: lock,
		Select: func(fields []GroupField) (*TServiceInfo, error) {
			rawDistincts := fields[1].([]any)
			domainCounts := make([]DomainCount, len(rawDistincts))
			for i, d := range rawDistincts {
				domainCounts[i] = d.(DomainCount)
			}
			return &TServiceInfo{
				Service:      fields[0].(string),
				DomainCounts: domainCounts,
			}, nil
		},
		From: d.ServiceDomains,
		GroupBy: func(serviceDomain *TServiceDomain) (any, []GroupField) {
			return [1]string{serviceDomain.Service},
				[]GroupField{
					AnyValue(serviceDomain.Service),
					Distinct(DomainCount{Domain: serviceDomain.Domain, Count: serviceDomain.Count}),
				}
		},
	})

	return d
}

func (d *PodSpreadLiteInfo) Clone() *PodSpreadLiteInfo {
	nd := &PodSpreadLiteInfo{}

	// To create a consistent snapshot of the entire query DAG, we must first 
	// acquire an exclusive lock on the entire Domain. 
	// SnapshotLockDomain identifies the shared LockGroup from the provided informers.
	locks := SnapshotLockDomain(d.ServiceNodes, d.NodeDomains, d.ServiceDomains, d.ServiceInfo)
	defer locks.Unlock()

	// 1. Manually clone the leaf sources.
	nd.PodUpdates = d.PodUpdates.Clone(nil).(ManualSharedInformer)
	nd.ServiceUpdates = d.ServiceUpdates.Clone(nil).(ManualSharedInformer)
	nd.NodeUpdates = d.NodeUpdates.Clone(nil).(ManualSharedInformer)

	// 2. Prepare a memo map with the leaf replacements. 
	// This map ensures that intermediate query stages are only cloned once,
	// correctly handling "Diamond DAGs" where multiple queries share a source.
	memo := map[cache.SharedInformer]cache.SharedInformer{
		d.PodUpdates:     nd.PodUpdates,
		d.ServiceUpdates: nd.ServiceUpdates,
		d.NodeUpdates:    nd.NodeUpdates,
	}

	// 3. Use ClonePipeline to recursively perform O(1) structural clones of all
	// intermediate query stages, attaching them to the new sources.
	nd.ServiceNodes = ClonePipeline(d.ServiceNodes, memo).(CloneableSharedInformerQuery)
	nd.NodeDomains = ClonePipeline(d.NodeDomains, memo).(CloneableSharedInformerQuery)
	nd.ServiceDomains = ClonePipeline(d.ServiceDomains, memo).(CloneableSharedInformerQuery)
	nd.ServiceInfo = ClonePipeline(d.ServiceInfo, memo).(CloneableSharedInformerQuery)

	return nd
}

func (p *TPod) String() string {
	return fmt.Sprintf("Pod(%s, %s)", p.NodeName, p.Label)
}

func (s *TService) String() string {
	return fmt.Sprintf("Service(%s, %s)", s.Name, s.SomeData)
}

func (n *TNode) String() string {
	return fmt.Sprintf("Node(%s, %v)", n.Name, n.Domains)
}
