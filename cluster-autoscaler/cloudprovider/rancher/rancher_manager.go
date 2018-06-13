package rancher

import (
	"time"
	"io"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"gopkg.in/gcfg.v1"
	"github.com/golang/glog"
	"os"
	"fmt"
	"k8s.io/autoscaler/cluster-autoscaler/config/dynamic"
)

const (
	scaleToZeroSupportedStandard = false
	operationWaitTimeout    = 5 * time.Second
	operationPollInterval   = 100 * time.Millisecond
	maxRecordsReturnedByAPI = 100
	refreshInterval         = 1 * time.Minute
)

// Config holds the configuration parsed from the --cloud-config flag
type Config struct {
	ClusterID          string `json:"clusterId" yaml:"clusterId"`
	NodePoolID         string `json:"nodePoolId" yaml:"nodePoolId"`
	RancherToken       string `json:"rancherToken" yaml:"rancherToken"`
	RancherURI         string `json:"rancherUri" yaml:"rancherUri"`
}

// RancherManager is handles rancher communication
type RancherManager struct {
	service               *rancherClient
	asgCache              *asgCache
	lastRefresh           time.Time
	explicitlyConfigured  map[string]bool
}

func BuildRancherManager(configReader io.Reader, discoveryOpts cloudprovider.NodeGroupDiscoveryOptions) (*RancherManager, error) {
	var cfg Config

	if configReader != nil {
		if err := gcfg.ReadInto(&cfg, configReader); err != nil {
			glog.Errorf("Couldn't read config: %v", err)
			return nil, err
		}
	} else {
		cfg.ClusterID = os.Getenv("CLUSTER_ID")
		cfg.NodePoolID = os.Getenv("NODE_POOL_ID")
		cfg.RancherToken = os.Getenv("RANCHER_TOKEN")
		cfg.RancherURI = os.Getenv("RANCHER_URI")
	}

	service, err := BuildRancherClient(&cfg)
	if err != nil {
		return nil, err
	}

	manager := &RancherManager{
		service:      service,
		explicitlyConfigured: make(map[string]bool),
	}

	cache, err := newAsgCache()
	if err != nil {
		return nil, err
	}
	manager.asgCache = cache

	if err := manager.fetchExplicitAsgs(discoveryOpts.NodeGroupSpecs); err != nil {
		return nil, err
	}

	if err := manager.forceRefresh(); err != nil {
		return nil, err
	}

	return manager, nil
}

func (m *RancherManager) fetchExplicitAsgs(specs []string) error {
	changed := false
	glog.V(4).Infof("fetchExplicitAsgs %s", specs)
	for _, spec := range specs {
		asg, err := m.buildAsgFromSpec(spec)
		if err != nil {
			return fmt.Errorf("failed to parse node group spec: %v", err)
		}
		if m.RegisterAsg(asg) {
			changed = true
		}
		m.explicitlyConfigured[asg.Id()] = true
	}

	if changed {
		if err := m.regenerateCache(); err != nil {
			return err
		}
	}
	return nil
}

func (m *RancherManager) buildAsgFromSpec(spec string) (cloudprovider.NodeGroup, error) {
	scaleToZeroSupported := scaleToZeroSupportedStandard

	s, err := dynamic.SpecFromString(spec, scaleToZeroSupported)
	if err != nil {
		return nil, fmt.Errorf("failed to parse node group spec: %v", err)
	}

	return NewNodePool(s, m)
}

// GetNodePoolSize gets NodePool size.
func (m *RancherManager) GetNodePoolSize(nodePool *NodePool) (int64, error) {
	nodePoolSize, err := m.service.nodePoolClient.Get(nodePool.Name)
	if err != nil {
		return -1, err
	}
	result := int64(len(nodePoolSize.nodes))
	return result, nil
}

// SetAsgSize sets ASG size.
func (m *RancherManager) SetNodePoolSize(nodePool *NodePool, size int64) error {
	glog.V(0).Infof("Setting NodePool %s size to %d", nodePool.Id(), size)
	_, err := m.service.nodePoolClient.SetDesiredCapacity(nodePool.Name, size)
	if err != nil {
		return err
	}
	return nil
}

// GetAsgNodes returns Asg nodes.
func (m *RancherManager) GetNodePoolNodes(nodePool *NodePool) ([]string, error) {
	result := make([]string, 0)

	nodes, err := m.service.nodePoolClient.Get(nodePool.Name)
	if err != nil {
		return []string{}, err
	}
	for _, instance := range nodes.nodes {
		result = append(result,
			fmt.Sprintf("%s", instance.name))
	}
	return result, nil
}

// DeleteInstances deletes the given instances. All instances must be controlled by the same ASG.
func (m *RancherManager) DeleteInstances(instances []*RancherRef) error {
	if len(instances) == 0 {
		return nil
	}

	for _, instance := range instances {
		resp, err := m.service.nodeClient.Delete(instance.Name)
		if err != nil {
			return err
		}
		glog.V(4).Infof(resp.Status)
	}

	return nil
}

// Refresh is called before every main loop and can be used to dynamically update cloud provider state.
// In particular the list of node groups returned by NodeGroups can change as a result of CloudProvider.Refresh().
func (m *RancherManager) Refresh() error {
	if m.lastRefresh.Add(refreshInterval).After(time.Now()) {
		return nil
	}
	return m.forceRefresh()
}

func (m *RancherManager) forceRefresh() error {
	m.lastRefresh = time.Now()
	glog.V(2).Infof("Refreshed NodePool list, next refresh after %v", m.lastRefresh.Add(refreshInterval))
	return nil
}

func (m *RancherManager) getAsgs() []cloudprovider.NodeGroup {
	return m.asgCache.get()
}

// GetAsgForInstance returns AsgConfig of the given Instance
func (m *RancherManager) GetNodePoolForInstance(instance *RancherRef) (cloudprovider.NodeGroup, error) {
	return m.asgCache.FindForInstance(instance)
}

// RegisterAsg registers an ASG.
func (m *RancherManager) RegisterAsg(asg cloudprovider.NodeGroup) bool {
	return m.asgCache.Register(asg)
}

// UnregisterAsg unregisters an ASG.
func (m *RancherManager) UnregisterAsg(asg cloudprovider.NodeGroup) bool {
	return m.asgCache.Unregister(asg)
}


func (m *RancherManager) regenerateCache() error {
	m.asgCache.mutex.Lock()
	defer m.asgCache.mutex.Unlock()
	return m.asgCache.regenerate()
}

// Cleanup the ASG cache.
func (m *RancherManager) Cleanup() {
	m.asgCache.Cleanup()
}