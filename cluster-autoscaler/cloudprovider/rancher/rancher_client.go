package rancher

import (
	"net/http"
)

type baseClient struct {
	ClusterID         string
	NodePoolID        string
	RancherToken      string
	RancherURI        string
}

func NewBaseClient(cfg *Config) *baseClient{
	return &baseClient{
		ClusterID: cfg.ClusterID,
		NodePoolID: cfg.NodePoolID,
		RancherToken: cfg.RancherToken,
		RancherURI: cfg.RancherURI,
	}
}

func BuildRancherClient(cfg *Config) (*rancherClient, error){
	rancherNodeClient := NewRancherNodeClient(cfg)
	rancherNodePoolClient := NewRancherNodePoolClient(cfg)

	client := &rancherClient{
		nodeClient:         rancherNodeClient,
		nodePoolClient:     rancherNodePoolClient,
	}

	return client, nil
}

type rancherNodeClient struct {
	client    baseClient
}

func NewRancherNodeClient(cfg *Config) *rancherNodeClient{
	return &rancherNodeClient{client: *NewBaseClient(cfg)}
}

// NodeClient defines needed functions for rancher Nodes.
// TODO: Result string->node
type NodeClient interface {
	Get(nodeName string) (result string, err error)
	Delete(nodeName string) (resp *http.Response, err error)
	List(nodePoolName string) (result []string, err error)
}

func (nc *rancherNodeClient) Get(nodeName string) (result string, err error){
	return result, nil
}
func (nc *rancherNodeClient) Delete(nodeName string) (resp *http.Response, err error){
	resp = &http.Response{Status: "200 OK"}
	return resp, nil
}
func (nc *rancherNodeClient) List(nodePoolName string) (result []string, err error){
	return result, nil
}

type rancherNodePoolClient struct {
	client    baseClient
}

func NewRancherNodePoolClient(cfg *Config) *rancherNodePoolClient{
	return &rancherNodePoolClient{client: *NewBaseClient(cfg)}
}

// NodeClient defines needed functions for rancher Nodes.
// TODO: Result string->node
type NodePoolClient interface {
	SetDesiredCapacity(nodePoolName string, size int64) (resp *http.Response, err error)
	Get(nodePoolName string) (result *NodePoolModel, err error)
}

func (nc *rancherNodePoolClient) SetDesiredCapacity(nodePoolName string, size int64) (resp *http.Response, err error){
	return resp, nil
}
func (nc *rancherNodePoolClient) Get(nodePoolName string) (result *NodePoolModel, err error){
	result = MockNodePool()
	return result, nil
}

type rancherClient struct {
	nodeClient     NodeClient
	nodePoolClient NodePoolClient
}

type NodePoolModel struct {
	nodePoolId      string
	quantity        string
	transitioning   string
	nodes           []*NodeModel
}

type NodeModel struct {
	name		 string
	nodeId       string
	clusterId    string
}

func MockNodePool() *NodePoolModel {
	refs :=  MockNodes()
	return &NodePoolModel{
		nodePoolId: "1",
		quantity: "4",
		transitioning: "no",
		nodes: refs,
	}
}

func MockNodes() []*NodeModel{
	refs := make([]*NodeModel, 4)
	refs[0] =
		&NodeModel{
			name: "integration-dev-worker1",
			nodeId: "1",
			clusterId: "1",
		}
	refs[1] =
		&NodeModel{
			name: "integration-dev-worker2",
			nodeId: "1",
			clusterId: "1",
		}
	refs[2] =
		&NodeModel{
			name: "integration-dev-worker3",
			nodeId: "3",
			clusterId: "1",
		}
	refs[3] =
		&NodeModel{
			name: "integration-dev-worker4",
			nodeId: "4",
			clusterId: "1",
		}
	return refs
}