package rancher

import (
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/utils/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	"github.com/golang/glog"
)

const (
	// ProviderName is the cloud provider name for rancher
	ProviderName = "rancher"
)

type RancherCloudProvider struct{
	rancherManager *RancherManager
	resourceLimiter *cloudprovider.ResourceLimiter
}

// BuildAwsCloudProvider builds CloudProvider implementation for AWS.
func BuildRancherCloudProvider(rancherManager *RancherManager,resourceLimiter *cloudprovider.ResourceLimiter) (cloudprovider.CloudProvider, error) {
	rancher := &RancherCloudProvider{
		rancherManager: rancherManager,
		resourceLimiter: resourceLimiter,
	}
	return rancher, nil
}

func (rancher *RancherCloudProvider) Name() string { return ProviderName }

func (rancher *RancherCloudProvider) NodeGroups() []cloudprovider.NodeGroup {
	asgs := rancher.rancherManager.asgCache.get()
	ngs := make([]cloudprovider.NodeGroup, len(asgs))
	for i, asg := range asgs {
		ngs[i] = asg
	}
	return ngs
	return []cloudprovider.NodeGroup{}
}

func (rancher *RancherCloudProvider) Pricing() (cloudprovider.PricingModel, errors.AutoscalerError) {
	return nil, cloudprovider.ErrNotImplemented
}

func (rancher *RancherCloudProvider) NodeGroupForNode(node *apiv1.Node) (cloudprovider.NodeGroup, error) {
	glog.V(6).Infof("Searching for node group for the node: %s", node.Name)
	ref := &RancherRef{
		Name: node.Name,
	}

	return rancher.rancherManager.GetNodePoolForInstance(ref)
}

func (rancher *RancherCloudProvider) GetAvailableMachineTypes() ([]string, error) {
	return []string{}, cloudprovider.ErrNotImplemented
}

func (rancher *RancherCloudProvider) NewNodeGroup(machineType string, labels map[string]string, systemLabels map[string]string,
	extraResources map[string]resource.Quantity) (cloudprovider.NodeGroup, error) {
	return nil, cloudprovider.ErrNotImplemented
}

// GetResourceLimiter returns struct containing limits (max, min) for resources (cores, memory etc.).
func (rancher *RancherCloudProvider) GetResourceLimiter() (*cloudprovider.ResourceLimiter, error) {
	return rancher.resourceLimiter, nil
}

// Refresh is called before every main loop and can be used to dynamically update cloud provider state.
// In particular the list of node groups returned by NodeGroups can change as a result of CloudProvider.Refresh().
func (rancher *RancherCloudProvider) Refresh() error {
	return rancher.rancherManager.Refresh()
}

// Cleanup cleans up all resources before the cloud provider is removed
func (rancher *RancherCloudProvider) Cleanup() error {
	rancher.rancherManager.Cleanup()
	return nil
}

