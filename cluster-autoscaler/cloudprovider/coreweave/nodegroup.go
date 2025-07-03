package coreweave

import (
    "k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
    apiv1 "k8s.io/api/core/v1"
    "k8s.io/autoscaler/cluster-autoscaler/config"
    "k8s.io/autoscaler/cluster-autoscaler/simulator/framework"
)

type CoreWeaveNodeGroup struct{}

func (ng *CoreWeaveNodeGroup) MaxSize() int                                 { return 0 }
func (ng *CoreWeaveNodeGroup) MinSize() int                                 { return 0 }
func (ng *CoreWeaveNodeGroup) TargetSize() (int, error)                     { return 0, cloudprovider.ErrNotImplemented }
func (ng *CoreWeaveNodeGroup) IncreaseSize(delta int) error                 { return cloudprovider.ErrNotImplemented }
func (ng *CoreWeaveNodeGroup) AtomicIncreaseSize(delta int) error           { return cloudprovider.ErrNotImplemented }
func (ng *CoreWeaveNodeGroup) DeleteNodes(nodes []*apiv1.Node) error        { return cloudprovider.ErrNotImplemented }
func (ng *CoreWeaveNodeGroup) ForceDeleteNodes(nodes []*apiv1.Node) error   { return cloudprovider.ErrNotImplemented }
func (ng *CoreWeaveNodeGroup) DecreaseTargetSize(delta int) error           { return cloudprovider.ErrNotImplemented }
func (ng *CoreWeaveNodeGroup) Id() string                                   { return "" }
func (ng *CoreWeaveNodeGroup) Debug() string                                { return "" }
func (ng *CoreWeaveNodeGroup) Nodes() ([]cloudprovider.Instance, error)     { return nil, cloudprovider.ErrNotImplemented }
func (ng *CoreWeaveNodeGroup) TemplateNodeInfo() (*framework.NodeInfo, error) { return nil, cloudprovider.ErrNotImplemented }
func (ng *CoreWeaveNodeGroup) Exist() bool                                  { return false }
func (ng *CoreWeaveNodeGroup) Create() (cloudprovider.NodeGroup, error)     { return nil, cloudprovider.ErrNotImplemented }
func (ng *CoreWeaveNodeGroup) Delete() error                                { return cloudprovider.ErrNotImplemented }
func (ng *CoreWeaveNodeGroup) Autoprovisioned() bool                        { return false }
func (ng *CoreWeaveNodeGroup) GetOptions(defaults config.NodeGroupAutoscalingOptions) (*config.NodeGroupAutoscalingOptions, error) {
    return nil, cloudprovider.ErrNotImplemented
}