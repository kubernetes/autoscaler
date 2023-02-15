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

package scaleway

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"os"
	"time"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/scaleway/scalewaygo"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	ca_errors "k8s.io/autoscaler/cluster-autoscaler/utils/errors"
	"k8s.io/autoscaler/cluster-autoscaler/utils/gpu"
	"k8s.io/klog/v2"
)

const (
	// GPULabel is the label added to GPU nodes
	GPULabel = "k8s.scaleway.com/gpu"
)

type scalewayCloudProvider struct {
	// client talks to Kapsule API
	client scalewaygo.Client
	// ClusterID is the cluster id where the Autoscaler is running.
	clusterID string
	// nodeGroups is an abstraction around the Pool object returned by the API
	nodeGroups []*NodeGroup

	resourceLimiter *cloudprovider.ResourceLimiter
}

func readConf(config *scalewaygo.Config, configFile io.Reader) error {
	body, err := ioutil.ReadAll(configFile)
	if err != nil {
		return err
	}
	err = json.Unmarshal(body, config)
	return err
}

func newScalewayCloudProvider(configFile io.Reader, defaultUserAgent string, rl *cloudprovider.ResourceLimiter) *scalewayCloudProvider {
	getenvOr := func(key, defaultValue string) string {
		value := os.Getenv(key)
		if value != "" {
			return value
		}
		return defaultValue
	}

	// Config file passed with `cloud-config` flag
	cfg := scalewaygo.Config{}
	if configFile != nil {
		err := readConf(&cfg, configFile)
		if err != nil {
			klog.Errorf("failed to read/parse scaleway config file: %s", err)
		}
	}

	// env takes precedence over config passed by command-line
	cfg.ClusterID = getenvOr("CLUSTER_ID", cfg.ClusterID)
	cfg.SecretKey = getenvOr("SCW_SECRET_KEY", cfg.SecretKey)
	cfg.Region = getenvOr("SCW_REGION", cfg.Region)
	cfg.ApiUrl = getenvOr("SCW_API_URL", cfg.ApiUrl)

	cfg.UserAgent = defaultUserAgent

	client, err := scalewaygo.NewClient(cfg)
	if err != nil {
		klog.Fatalf("failed to create scaleway cloud provider: %v", err)
	}

	klog.V(4).Infof("Scaleway Cloud Provider built; ClusterId=%s,SecretKey=%s-***,Region=%s,ApiURL=%s", cfg.ClusterID, client.Token()[:8], client.Region(), client.ApiURL())

	return &scalewayCloudProvider{
		client:          client,
		clusterID:       cfg.ClusterID,
		resourceLimiter: rl,
	}
}

// BuildScaleway returns CloudProvider implementation for Scaleway.
func BuildScaleway(
	opts config.AutoscalingOptions,
	do cloudprovider.NodeGroupDiscoveryOptions,
	rl *cloudprovider.ResourceLimiter,
) cloudprovider.CloudProvider {
	var configFile io.Reader

	if opts.CloudConfig != "" {
		configFile, err := os.Open(opts.CloudConfig)

		if err != nil {
			klog.Errorf("could not open scaleway configuration %s: %s", opts.CloudConfig, err)
		} else {
			defer func() {
				err = configFile.Close()
				if err != nil {
					klog.Errorf("failed to close scaleway config file: %s", err)
				}
			}()
		}
	}
	return newScalewayCloudProvider(configFile, opts.UserAgent, rl)
}

// Name returns 'scaleway'
func (*scalewayCloudProvider) Name() string {
	return cloudprovider.ScalewayProviderName
}

// NodeGroups returns all node groups configured for this cluster.
// critical endpoint, make it fast
func (scw *scalewayCloudProvider) NodeGroups() []cloudprovider.NodeGroup {

	klog.V(4).Info("NodeGroups,ClusterID=", scw.clusterID)

	nodeGroups := make([]cloudprovider.NodeGroup, len(scw.nodeGroups))
	for i, ng := range scw.nodeGroups {
		nodeGroups[i] = ng
	}
	return nodeGroups
}

func (scw *scalewayCloudProvider) nodeGroupForNode(node *apiv1.Node) (*NodeGroup, error) {
	for _, ng := range scw.nodeGroups {
		if _, ok := ng.nodes[node.Spec.ProviderID]; ok {
			return ng, nil
		}
	}
	return nil, nil
}

// NodeGroupForNode returns the node group for the given node, nil if the node
// should not be processed by cluster autoscaler, or non-nil error if such
// occurred.
// critical endpoint, make it fast
func (scw *scalewayCloudProvider) NodeGroupForNode(node *apiv1.Node) (cloudprovider.NodeGroup, error) {
	klog.V(4).Infof("NodeGroupForNode,NodeSpecProviderID=%s", node.Spec.ProviderID)

	return scw.nodeGroupForNode(node)
}

// HasInstance returns whether a given node has a corresponding instance in this cloud provider
func (scw *scalewayCloudProvider) HasInstance(node *apiv1.Node) (bool, error) {
	return true, cloudprovider.ErrNotImplemented
}

func (scw *scalewayCloudProvider) NodePrice(node *apiv1.Node, startTime time.Time, endTime time.Time) (float64, error) {
	ng, err := scw.nodeGroupForNode(node)
	if err != nil {
		return 0.0, err
	}

	d := endTime.Sub(startTime)
	hours := math.Ceil(d.Hours())

	return hours * float64(ng.specs.NodePricePerHour), nil
}

func (scw *scalewayCloudProvider) PodPrice(pod *apiv1.Pod, startTime time.Time, endTime time.Time) (float64, error) {
	return 0.0, nil
}

// Pricing return pricing model for scaleway.
func (scw *scalewayCloudProvider) Pricing() (cloudprovider.PricingModel, ca_errors.AutoscalerError) {
	klog.V(4).Info("Pricing,called")
	return scw, nil
}

// GetAvailableMachineTypes get all machine types that can be requested from scaleway.
// Not implemented
func (scw *scalewayCloudProvider) GetAvailableMachineTypes() ([]string, error) {
	return []string{}, nil
}

func (scw *scalewayCloudProvider) NewNodeGroup(
	machineType string,
	labels map[string]string,
	systemLabels map[string]string,
	taints []apiv1.Taint,
	extraResources map[string]resource.Quantity,
) (cloudprovider.NodeGroup, error) {
	klog.V(4).Info("NewNodeGroup,called")
	return nil, cloudprovider.ErrNotImplemented
}

// GetResourceLimiter returns struct containing limits (max, min) for resources (cores, memory etc.).
func (scw *scalewayCloudProvider) GetResourceLimiter() (*cloudprovider.ResourceLimiter, error) {
	klog.V(4).Info("GetResourceLimiter,called")
	return scw.resourceLimiter, nil
}

// GPULabel returns the label added to nodes with GPU resource.
func (scw *scalewayCloudProvider) GPULabel() string {
	klog.V(6).Info("GPULabel,called")
	return GPULabel
}

// GetAvailableGPUTypes return all available GPU types cloud provider supports.
// not yet implemented.
func (scw *scalewayCloudProvider) GetAvailableGPUTypes() map[string]struct{} {
	klog.V(4).Info("GetAvailableGPUTypes,called")
	return nil
}

// GetNodeGpuConfig returns the label, type and resource name for the GPU added to node. If node doesn't have
// any GPUs, it returns nil.
func (scw *scalewayCloudProvider) GetNodeGpuConfig(node *apiv1.Node) *cloudprovider.GpuConfig {
	klog.V(6).Info("GetNodeGpuConfig,called")
	return gpu.GetNodeGPUFromCloudProvider(scw, node)
}

// Cleanup cleans up open resources before the cloud provider is destroyed, i.e. go routines etc.
func (scw *scalewayCloudProvider) Cleanup() error {
	klog.V(4).Info("Cleanup,called")
	return nil
}

// Refresh is called before every main loop and can be used to dynamically update cloud provider state.
// In particular the list of node groups returned by NodeGroups can change as a result of CloudProvider.Refresh().
func (scw *scalewayCloudProvider) Refresh() error {
	klog.V(4).Info("Refresh,ClusterID=", scw.clusterID)

	ctx := context.Background()
	resp, err := scw.client.ListPools(ctx, &scalewaygo.ListPoolsRequest{ClusterID: scw.clusterID})

	if err != nil {
		klog.Errorf("Refresh,failed to list pools for cluster %s: %s", scw.clusterID, err)
		return err
	}

	var ng []*NodeGroup

	for _, p := range resp.Pools {

		if p.Pool.Autoscaling == false {
			continue
		}

		nodes, err := nodesFromPool(scw.client, p.Pool)
		if err != nil {
			return fmt.Errorf("Refresh,failed to list nodes for pool %s: %w", p.Pool.ID, err)
		}
		ng = append(ng, &NodeGroup{
			Client: scw.client,
			nodes:  nodes,
			specs:  &p.Specs,
			p:      p.Pool,
		})
	}
	klog.V(4).Infof("Refresh,ClusterID=%s,%d pools found", scw.clusterID, len(ng))

	scw.nodeGroups = ng

	return nil
}
