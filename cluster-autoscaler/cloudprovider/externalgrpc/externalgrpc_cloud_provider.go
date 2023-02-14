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

package externalgrpc

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"net"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"gopkg.in/yaml.v2"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/externalgrpc/protos"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	"k8s.io/autoscaler/cluster-autoscaler/utils/errors"
	"k8s.io/autoscaler/cluster-autoscaler/utils/gpu"
	klog "k8s.io/klog/v2"
)

const (
	grpcTimeout = 5 * time.Second
)

// externalGrpcCloudProvider implements CloudProvider interface.
type externalGrpcCloudProvider struct {
	resourceLimiter *cloudprovider.ResourceLimiter
	client          protos.CloudProviderClient

	mutex                 sync.Mutex
	nodeGroupForNodeCache map[string]cloudprovider.NodeGroup // used to cache NodeGroupForNode grpc calls. Discarded at each Refresh()
	nodeGroupsCache       []cloudprovider.NodeGroup          // used to cache NodeGroups grpc calls. Discarded at each Refresh()
	gpuLabelCache         *string                            // used to cache GPULabel grpc calls
	gpuTypesCache         map[string]struct{}                // used to cache GetAvailableGPUTypes grpc calls
}

// Name returns name of the cloud provider.
func (e *externalGrpcCloudProvider) Name() string {
	return cloudprovider.ExternalGrpcProviderName
}

// NodeGroups returns all node groups configured for this cloud provider.
func (e *externalGrpcCloudProvider) NodeGroups() []cloudprovider.NodeGroup {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	if e.nodeGroupsCache != nil {
		klog.V(5).Info("Returning cached NodeGroups")
		return e.nodeGroupsCache
	}
	nodeGroups := make([]cloudprovider.NodeGroup, 0)
	ctx, cancel := context.WithTimeout(context.Background(), grpcTimeout)
	defer cancel()
	klog.V(5).Info("Performing gRPC call NodeGroups")
	res, err := e.client.NodeGroups(ctx, &protos.NodeGroupsRequest{})
	if err != nil {
		klog.V(1).Infof("Error on gRPC call NodeGroups: %v", err)
		return nodeGroups
	}
	for _, pbNg := range res.GetNodeGroups() {
		ng := &NodeGroup{
			id:      pbNg.Id,
			minSize: int(pbNg.MinSize),
			maxSize: int(pbNg.MaxSize),
			debug:   pbNg.Debug,
			client:  e.client,
		}
		nodeGroups = append(nodeGroups, ng)
	}
	e.nodeGroupsCache = nodeGroups
	return nodeGroups
}

// NodeGroupForNode returns the node group for the given node, nil if the node
// should not be processed by cluster autoscaler, or non-nil error if such
// occurred. Must be implemented.
func (e *externalGrpcCloudProvider) NodeGroupForNode(node *apiv1.Node) (cloudprovider.NodeGroup, error) {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	if node == nil {
		return nil, fmt.Errorf("node in NodeGroupForNode call cannot be nil")
	}
	nodeID := node.Name + node.Spec.ProviderID //ProviderID is empty in some edge cases
	// lookup cache
	if ng, ok := e.nodeGroupForNodeCache[nodeID]; ok {
		klog.V(5).Infof("Returning cached information for NodeGroupForNode for node %v - %v", node.Name, node.Spec.ProviderID)
		return ng, nil
	}
	// perform grpc call
	ctx, cancel := context.WithTimeout(context.Background(), grpcTimeout)
	defer cancel()
	klog.V(5).Infof("Performing gRPC call NodeGroupForNode for node %v - %v", node.Name, node.Spec.ProviderID)
	res, err := e.client.NodeGroupForNode(ctx, &protos.NodeGroupForNodeRequest{
		Node: externalGrpcNode(node),
	})
	if err != nil {
		klog.V(1).Infof("Error on gRPC call NodeGroupForNode: %v", err)
		return nil, err
	}
	pbNg := res.GetNodeGroup()
	if pbNg.GetId() == "" { // if id == "" then the node should not be processed by cluster autoscaler, do not cache this
		return nil, nil
	}
	ng := &NodeGroup{
		id:      pbNg.GetId(),
		maxSize: int(pbNg.GetMaxSize()),
		minSize: int(pbNg.GetMinSize()),
		debug:   pbNg.GetDebug(),
		client:  e.client,
	}
	e.nodeGroupForNodeCache[nodeID] = ng
	return ng, nil
}

// HasInstance returns whether a given node has a corresponding instance in this cloud provider
func (e *externalGrpcCloudProvider) HasInstance(node *apiv1.Node) (bool, error) {
	return true, cloudprovider.ErrNotImplemented
}

// pricingModel implements cloudprovider.PricingModel interface.
type pricingModel struct {
	client protos.CloudProviderClient
}

// NodePrice returns a price of running the given node for a given period of time.
func (m *pricingModel) NodePrice(node *apiv1.Node, startTime time.Time, endTime time.Time) (float64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), grpcTimeout)
	defer cancel()
	klog.V(5).Infof("Performing gRPC call PricingNodePrice for node %v", node.Name)
	start := metav1.NewTime(startTime)
	end := metav1.NewTime(endTime)
	res, err := m.client.PricingNodePrice(ctx, &protos.PricingNodePriceRequest{
		Node:      externalGrpcNode(node),
		StartTime: &start,
		EndTime:   &end,
	})
	if err != nil {
		klog.V(1).Infof("Error on gRPC call PricingNodePrice: %v", err)
		return 0, err
	}
	return res.GetPrice(), nil
}

// PodPrice returns a theoretical minimum price of running a pod for a given
// period of time on a perfectly matching machine.
func (m *pricingModel) PodPrice(pod *apiv1.Pod, startTime time.Time, endTime time.Time) (float64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), grpcTimeout)
	defer cancel()
	klog.V(5).Infof("Performing gRPC call PricingPodPrice for pod %v", pod.Name)
	start := metav1.NewTime(startTime)
	end := metav1.NewTime(endTime)
	res, err := m.client.PricingPodPrice(ctx, &protos.PricingPodPriceRequest{
		Pod:       pod,
		StartTime: &start,
		EndTime:   &end,
	})
	if err != nil {
		klog.V(1).Infof("Error on gRPC call PricingPodPrice: %v", err)
		return 0, err
	}
	return res.GetPrice(), nil
}

// Pricing returns pricing model for this cloud provider or error if not available.
// Implementation optional.
//
// The external gRPC provider will always return a pricing model without errors,
// even if a cloud provider does not actually support this feature, errors will be returned
// by subsequent calls to the pricing model if this is the case.
func (e *externalGrpcCloudProvider) Pricing() (cloudprovider.PricingModel, errors.AutoscalerError) {
	return &pricingModel{
		client: e.client,
	}, nil
}

// GetAvailableMachineTypes get all machine types that can be requested from the cloud provider.
// Implementation optional.
func (e *externalGrpcCloudProvider) GetAvailableMachineTypes() ([]string, error) {
	return []string{}, cloudprovider.ErrNotImplemented
}

// NewNodeGroup builds a theoretical node group based on the node definition provided. The node group is not automatically
// created on the cloud provider side. The node group is not returned by NodeGroups() until it is created.
// Implementation optional.
func (e *externalGrpcCloudProvider) NewNodeGroup(machineType string, labels map[string]string, systemLabels map[string]string,
	taints []apiv1.Taint, extraResources map[string]resource.Quantity) (cloudprovider.NodeGroup, error) {
	return nil, cloudprovider.ErrNotImplemented
}

// GetResourceLimiter returns struct containing limits (max, min) for resources (cores, memory etc.).
func (e *externalGrpcCloudProvider) GetResourceLimiter() (*cloudprovider.ResourceLimiter, error) {
	return e.resourceLimiter, nil
}

// GPULabel returns the label added to nodes with GPU resource.
func (e *externalGrpcCloudProvider) GPULabel() string {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	if e.gpuLabelCache != nil {
		klog.V(5).Info("Returning cached GPULabel")
		return *e.gpuLabelCache
	}
	ctx, cancel := context.WithTimeout(context.Background(), grpcTimeout)
	defer cancel()
	klog.V(5).Info("Performing gRPC call GPULabel")
	res, err := e.client.GPULabel(ctx, &protos.GPULabelRequest{})
	if err != nil {
		klog.V(1).Infof("Error on gRPC call GPULabel: %v", err)
		return ""
	}
	gpuLabel := res.GetLabel()
	e.gpuLabelCache = &gpuLabel
	return gpuLabel
}

// GetAvailableGPUTypes return all available GPU types cloud provider supports.
func (e *externalGrpcCloudProvider) GetAvailableGPUTypes() map[string]struct{} {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	if e.gpuTypesCache != nil {
		klog.V(5).Info("Returning cached GetAvailableGPUTypes")
		return e.gpuTypesCache
	}
	ctx, cancel := context.WithTimeout(context.Background(), grpcTimeout)
	defer cancel()
	klog.V(5).Info("Performing gRPC call GetAvailableGPUTypes")
	res, err := e.client.GetAvailableGPUTypes(ctx, &protos.GetAvailableGPUTypesRequest{})
	if err != nil {
		klog.V(1).Infof("Error on gRPC call GetAvailableGPUTypes: %v", err)
		return nil
	}
	gpuTypes := make(map[string]struct{})
	var empty struct{}
	for k := range res.GetGpuTypes() {
		gpuTypes[k] = empty
	}
	e.gpuTypesCache = gpuTypes
	return gpuTypes
}

// GetNodeGpuConfig returns the label, type and resource name for the GPU added to node. If node doesn't have
// any GPUs, it returns nil.
func (e *externalGrpcCloudProvider) GetNodeGpuConfig(node *apiv1.Node) *cloudprovider.GpuConfig {
	return gpu.GetNodeGPUFromCloudProvider(e, node)
}

// Cleanup cleans up open resources before the cloud provider is destroyed, i.e. go routines etc.
func (e *externalGrpcCloudProvider) Cleanup() error {
	ctx, cancel := context.WithTimeout(context.Background(), grpcTimeout)
	defer cancel()
	klog.V(5).Info("Performing gRPC call Cleanup")
	_, err := e.client.Cleanup(ctx, &protos.CleanupRequest{})
	if err != nil {
		klog.V(1).Infof("Error on gRPC call Cleanup: %v", err)
		return err
	}
	return nil
}

// Refresh is called before every main loop and can be used to dynamically update cloud provider state.
// In particular the list of node groups returned by NodeGroups can change as a result of CloudProvider.Refresh().
func (e *externalGrpcCloudProvider) Refresh() error {
	// invalidate cache
	e.mutex.Lock()
	e.nodeGroupForNodeCache = make(map[string]cloudprovider.NodeGroup)
	e.nodeGroupsCache = nil
	e.mutex.Unlock()
	ctx, cancel := context.WithTimeout(context.Background(), grpcTimeout)
	defer cancel()
	klog.V(5).Info("Performing gRPC call Refresh")
	_, err := e.client.Refresh(ctx, &protos.RefreshRequest{})
	if err != nil {
		klog.V(1).Infof("Error on gRPC call Refresh: %v", err)
		return err
	}
	return nil
}

// BuildExternalGrpc builds the externalgrpc cloud provider.
func BuildExternalGrpc(
	opts config.AutoscalingOptions,
	do cloudprovider.NodeGroupDiscoveryOptions,
	rl *cloudprovider.ResourceLimiter,
) cloudprovider.CloudProvider {
	if opts.CloudConfig == "" {
		klog.Fatal("No config file provided, please specify it via the --cloud-config flag")
	}
	config, err := ioutil.ReadFile(opts.CloudConfig)
	if err != nil {
		klog.Fatalf("Could not open cloud provider configuration file %q: %v", opts.CloudConfig, err)
	}
	client, err := newExternalGrpcCloudProviderClient(config)
	if err != nil {
		klog.Fatalf("Could not create gRPC client: %v", err)
	}
	return newExternalGrpcCloudProvider(client, rl)
}

// cloudConfig is the struct hoding the configs to connect to the external cluster autoscaler provider service.
type cloudConfig struct {
	Address string `yaml:"address"` // external cluster autoscaler provider address of the form "host:port", "host%zone:port", "[host]:port" or "[host%zone]:port"
	Key     string `yaml:"key"`     // path to file containing the tls key
	Cert    string `yaml:"cert"`    // path to file containing the tls certificate
	Cacert  string `yaml:"cacert"`  // path to file containing the CA certificate
}

func newExternalGrpcCloudProviderClient(config []byte) (protos.CloudProviderClient, error) {
	var yamlConfig cloudConfig
	err := yaml.Unmarshal([]byte(config), &yamlConfig)
	if err != nil {
		return nil, fmt.Errorf("can't parse YAML: %v", err)
	}
	host, _, err := net.SplitHostPort(yamlConfig.Address)
	if err != nil {
		return nil, fmt.Errorf("failed to parse address: %v", err)
	}
	var dialOpt grpc.DialOption
	if len(yamlConfig.Cert) == 0 {
		klog.V(5).Info("No certs specified in external gRPC provider config, using insecure mode")
		dialOpt = grpc.WithInsecure()
	} else {
		certFile, err := ioutil.ReadFile(yamlConfig.Cert)
		if err != nil {
			return nil, fmt.Errorf("could not open Cert configuration file %q: %v", yamlConfig.Cert, err)
		}
		keyFile, err := ioutil.ReadFile(yamlConfig.Key)
		if err != nil {
			return nil, fmt.Errorf("could not open Key configuration file %q: %v", yamlConfig.Key, err)
		}
		cacertFile, err := ioutil.ReadFile(yamlConfig.Cacert)
		if err != nil {
			return nil, fmt.Errorf("could not open Cacert configuration file %q: %v", yamlConfig.Cacert, err)
		}
		cert, err := tls.X509KeyPair(certFile, keyFile)
		if err != nil {
			return nil, fmt.Errorf("failed to parse cert key pair: %v", err)
		}
		certPool := x509.NewCertPool()
		ok := certPool.AppendCertsFromPEM(cacertFile)
		if !ok {
			return nil, fmt.Errorf("failed to parse ca: %v", err)
		}
		transportCreds := credentials.NewTLS(&tls.Config{
			ServerName:   host,
			Certificates: []tls.Certificate{cert},
			RootCAs:      certPool,
		})
		dialOpt = grpc.WithTransportCredentials(transportCreds)
	}
	conn, err := grpc.Dial(yamlConfig.Address, dialOpt)
	if err != nil {
		return nil, fmt.Errorf("failed to dial server: %v", err)
	}
	return protos.NewCloudProviderClient(conn), nil
}

func newExternalGrpcCloudProvider(client protos.CloudProviderClient, rl *cloudprovider.ResourceLimiter) cloudprovider.CloudProvider {
	return &externalGrpcCloudProvider{
		resourceLimiter:       rl,
		client:                client,
		nodeGroupForNodeCache: make(map[string]cloudprovider.NodeGroup),
	}
}

// externalGrpcNode converts an apiv1.Node to a protos.ExternalGrpcNode.
func externalGrpcNode(apiv1Node *apiv1.Node) *protos.ExternalGrpcNode {
	return &protos.ExternalGrpcNode{
		ProviderID:  apiv1Node.Spec.ProviderID,
		Name:        apiv1Node.Name,
		Labels:      apiv1Node.Labels,
		Annotations: apiv1Node.Annotations,
	}
}
