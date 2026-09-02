/*
Copyright 2023 The Kubernetes Authors.

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

package upcloud

import (
	"context"
	"fmt"
	"os"
	"time"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/upcloud/pkg/github.com/upcloudltd/upcloud-go-api/v8/upcloud/client"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/upcloud/pkg/github.com/upcloudltd/upcloud-go-api/v8/upcloud/service"
	"k8s.io/client-go/informers"
	"k8s.io/klog/v2"
	"sigs.k8s.io/cluster-autoscaler/pkg/cloudprovider"
	"sigs.k8s.io/cluster-autoscaler/pkg/cloudprovider/builder"
	"sigs.k8s.io/cluster-autoscaler/pkg/config"
	coreoptions "sigs.k8s.io/cluster-autoscaler/pkg/core/options"
	"sigs.k8s.io/cluster-autoscaler/pkg/utils/errors"
	"sigs.k8s.io/cluster-autoscaler/pkg/utils/gpu"
)

const (
	timeoutProviderInit         time.Duration = time.Second * 15
	timeoutGetRequest           time.Duration = time.Second * 10
	timeoutModifyNodeGroup      time.Duration = time.Second * 20
	timeoutNodeGroupStateChange time.Duration = time.Minute * 20
	timeoutDeleteNode           time.Duration = time.Second * 20
	timeoutWaitNodeGroupState   time.Duration = time.Minute * 20

	nodeGroupMinSize int = 1
	nodeGroupMaxSize int = 20

	logInfo  klog.Level = 4
	logDebug klog.Level = 5

	envUpCloudToken     string = "UPCLOUD_TOKEN"
	envUpCloudUsername  string = "UPCLOUD_USERNAME"
	envUpCloudPassword  string = "UPCLOUD_PASSWORD"
	envUpCloudClusterID string = "UPCLOUD_CLUSTER_ID"

	// ProviderName is the cloud provider name for this provider.
	ProviderName = "upcloud"
)

type upCloudConfig struct {
	ClusterID string
	Token     string
	Username  string
	Password  string
	UserAgent string
}

// upCloudCloudProvider implements cloudprovide.CloudProvider interfaces
type upCloudCloudProvider struct {
	manager         *manager
	resourceLimiter *cloudprovider.ResourceLimiter
}

// Name returns name of the cloud provider.
func (u *upCloudCloudProvider) Name() string {
	klog.V(logDebug).Info("UpCloud CloudProvider.Name called")
	return ProviderName
}

// NodeGroups returns all node groups configured for this cloud provider.
func (u *upCloudCloudProvider) NodeGroups(ctx context.Context) []cloudprovider.NodeGroup {
	klog.V(logDebug).Info("UpCloud CloudProvider.NodeGroups called")
	nodeGroups := make([]cloudprovider.NodeGroup, len(u.manager.nodeGroups))
	for i, ng := range u.manager.nodeGroups {
		nodeGroups[i] = ng
	}
	return nodeGroups
}

// NodeGroupForNode returns the node group for the given node, nil if the node
// should not be processed by cluster autoscaler, or non-nil error if such
// occurred. Must be implemented.
func (u *upCloudCloudProvider) NodeGroupForNode(ctx context.Context, node *apiv1.Node) (cloudprovider.NodeGroup, error) {
	klog.V(logDebug).Info("UpCloud CloudProvider.NodeGroupForNode called")
	providerID := node.Spec.ProviderID
	for _, group := range u.manager.nodeGroups {
		nodes, err := group.Nodes(ctx)
		if err != nil {
			return nil, err
		}
		for _, n := range nodes {
			if n.Id == providerID {
				return group, nil
			}
		}
	}
	klog.V(logInfo).Infof("couldn't find node group for node with provider ID %s", providerID)
	return nil, nil
}

// HasInstance returns whether the node has corresponding instance in cloud provider,
// true if the node has an instance, false if it no longer exists
func (u *upCloudCloudProvider) HasInstance(ctx context.Context, _ *apiv1.Node) (bool, error) {
	klog.V(logDebug).Info("UpCloud CloudProvider.HasInstance called")
	return true, cloudprovider.ErrNotImplemented
}

// GetResourceLimiter returns struct containing limits (max, min) for resources (cores, memory etc.).
func (u *upCloudCloudProvider) GetResourceLimiter(ctx context.Context) (*cloudprovider.ResourceLimiter, error) {
	klog.V(logDebug).Info("UpCloud CloudProvider.GetResourceLimiter called")
	return u.resourceLimiter, nil
}

// GetAvailableGPUTypes return all available GPU types cloud provider supports.
func (u *upCloudCloudProvider) GetAvailableGPUTypes(ctx context.Context) map[string]struct{} {
	klog.V(logDebug).Info("UpCloud CloudProvider.GetAvailableGPUTypes called")
	return nil
}

// GPULabel returns the label added to nodes with GPU resource.
func (u *upCloudCloudProvider) GPULabel(ctx context.Context) string {
	klog.V(logDebug).Info("UpCloud CloudProvider.GPULabel called")
	return ""
}

// GetNodeGpuConfig returns the label, type and resource name for the GPU added to node. If node doesn't have
// any GPUs, it returns nil.
func (u *upCloudCloudProvider) GetNodeGpuConfig(ctx context.Context, node *apiv1.Node) *cloudprovider.GpuConfig {
	klog.V(logDebug).Info("UpCloud CloudProvider.GetNodeGpuConfig called")
	return gpu.GetNodeGPUFromCloudProvider(ctx, u, node)
}

// Refresh is called before every main loop and can be used to dynamically update cloud provider state.
// In particular the list of node groups returned by NodeGroups can change as a result of CloudProvider.Refresh().
func (u *upCloudCloudProvider) Refresh(ctx context.Context) error {
	klog.V(logDebug).Info("UpCloud CloudProvider.Refresh called")
	return u.manager.refresh()
}

// Pricing returns pricing model for this cloud provider or error if not available.
// Implementation optional.
func (u *upCloudCloudProvider) Pricing(ctx context.Context) (cloudprovider.PricingModel, errors.AutoscalerError) {
	klog.V(logDebug).Info("UpCloud CloudProvider.Pricing called")
	return nil, cloudprovider.ErrNotImplemented
}

// GetAvailableMachineTypes get all machine types that can be requested from the cloud provider.
// Implementation optional.
func (u *upCloudCloudProvider) GetAvailableMachineTypes(ctx context.Context) ([]string, error) {
	klog.V(logDebug).Info("UpCloud CloudProvider.GetAvailableMachineTypes called")
	return nil, cloudprovider.ErrNotImplemented
}

// NewNodeGroup builds a theoretical node group based on the node definition provided. The node group is not automatically
// created on the cloud provider side. The node group is not returned by NodeGroups() until it is created.
// Implementation optional.
func (u *upCloudCloudProvider) NewNodeGroup(ctx context.Context, _ string, _ map[string]string, _ map[string]string, _ []apiv1.Taint, _ map[string]resource.Quantity) (cloudprovider.NodeGroup, error) {
	klog.V(logDebug).Info("UpCloud CloudProvider.NewNodeGroup called")
	return nil, cloudprovider.ErrNotImplemented
}

// Cleanup cleans up open resources before the cloud provider is destroyed, i.e. go routines etc.
func (u *upCloudCloudProvider) Cleanup(ctx context.Context) error {
	klog.V(logDebug).Info("UpCloud CloudProvider.Cleanup called")
	return nil
}

// BuildUpCloud builds UpCloud's cloud provider implementation
func BuildUpCloud(opts *coreoptions.AutoscalerOptions, do cloudprovider.NodeGroupDiscoveryOptions, rl *cloudprovider.ResourceLimiter) cloudprovider.CloudProvider {
	ctx, cancel := context.WithTimeout(context.Background(), timeoutProviderInit)
	defer cancel()

	cfg, err := buildCloudConfig(opts.AutoscalingOptions)
	if err != nil {
		klog.Fatalf("failed to initialize UpCloud config: %v", err)
	}
	svc, err := newUpCloudService(cfg)
	if err != nil {
		klog.Fatalf("failed to initialize UpCloud service: %v", err)
	}
	manager, err := newManager(ctx, svc, cfg, opts.AutoscalingOptions, do)
	if err != nil {
		klog.Fatalf("failed to initialize manager: %v", err)
	}

	klog.V(logInfo).Infof("%s cloud provider initialized successfully", opts.CloudProviderName)
	if len(manager.nodeGroupSpecs) > 0 {
		for _, v := range manager.nodeGroupSpecs {
			klog.Infof("using custom %s node group spec: %s min=%d max=%d", opts.CloudProviderName, v.Name, v.MinSize, v.MaxSize)
		}
	}
	return &upCloudCloudProvider{
		manager:         manager,
		resourceLimiter: rl,
	}
}

// buildCloudConfig builds cloud config for UpCloud provider.
func buildCloudConfig(opts config.AutoscalingOptions) (upCloudConfig, error) {
	return cloudConfigFromEnv(opts)
}

func newUpCloudService(cfg upCloudConfig) (upCloudService, error) {
	if cfg.Token == "" && (cfg.Username == "" || cfg.Password == "") {
		return nil, errors.NewAutoscalerError(errors.ConfigurationError, "UpCloud API credentials not configured")
	}
	upClient := client.New(cfg.Username, cfg.Password)
	if cfg.Token != "" {
		upClient = client.New("", "", client.WithBearerAuth(cfg.Token))
	}
	if cfg.UserAgent != "" {
		upClient.UserAgent = cfg.UserAgent
	}
	return service.New(upClient), nil
}

func cloudConfigFromEnv(opts config.AutoscalingOptions) (upCloudConfig, error) {
	cfg := upCloudConfig{}

	if cfg.ClusterID = os.Getenv(envUpCloudClusterID); cfg.ClusterID == "" {
		return cfg, fmt.Errorf("environment variable %s not set", envUpCloudClusterID)
	}
	if cfg.Token = os.Getenv(envUpCloudToken); cfg.Token == "" {
		if cfg.Username = os.Getenv(envUpCloudUsername); cfg.Username == "" {
			return cfg, fmt.Errorf("environment variable %s not set", envUpCloudUsername)
		}
		if cfg.Password = os.Getenv(envUpCloudPassword); cfg.Password == "" {
			return cfg, fmt.Errorf("environment variable %s not set", envUpCloudPassword)
		}
	}
	if opts.UserAgent != "" {
		cfg.UserAgent = opts.UserAgent
	}

	return cfg, nil
}

func init() {
	builder.RegisterCloudProvider(ProviderName, func(opts *coreoptions.AutoscalerOptions, do cloudprovider.NodeGroupDiscoveryOptions, rl *cloudprovider.ResourceLimiter, informerFactory informers.SharedInformerFactory) cloudprovider.CloudProvider {
		return BuildUpCloud(opts, do, rl)
	})
	builder.SetDefaultCloudProvider(ProviderName)
}
