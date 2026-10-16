/*
Copyright The Kubernetes Authors.

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

package clusterapi

import (
	"context"
	"fmt"

	"k8s.io/client-go/discovery"
	"k8s.io/client-go/discovery/cached/memory"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
	"k8s.io/client-go/scale"

	"sigs.k8s.io/cluster-autoscaler/pkg/cloudprovider"
	ca_processors "sigs.k8s.io/cluster-autoscaler/pkg/processors"
	"sigs.k8s.io/cluster-autoscaler/pkg/processors/scaledowncandidates"
)

// ProviderOptions configures a Cluster API cloud provider built from in-memory REST configs.
// The configs are copied before use, client side rate limits (QPS, Burst) must be set on them by the caller.
type ProviderOptions struct {
	// ManagementConfig is the REST config of the cluster holding the Cluster API objects.
	ManagementConfig *rest.Config
	// WorkloadConfig is the REST config of the cluster whose nodes are autoscaled.
	WorkloadConfig *rest.Config
	// DiscoveryOptions selects the node groups, same as the --nodes and --node-group-auto-discovery flags.
	DiscoveryOptions cloudprovider.NodeGroupDiscoveryOptions
	// ResourceLimiter holds the cluster wide resource limits, it may be nil.
	ResourceLimiter *cloudprovider.ResourceLimiter
}

// providerClients are the clients a Cluster API provider needs.
type providerClients struct {
	management dynamic.Interface
	workload   kubernetes.Interface
	discovery  discovery.DiscoveryInterface
	scale      scale.ScalesGetter
}

// NewProvider builds a Cluster API cloud provider whose informers stop when ctx is done
// or when Cleanup is called. Unlike BuildClusterAPI it returns errors instead of exiting.
// Callers must call Cleanup to release the provider, it blocks until the informers have exited.
func NewProvider(ctx context.Context, opts ProviderOptions) (cloudprovider.CloudProvider, error) {
	if opts.ManagementConfig == nil {
		return nil, fmt.Errorf("management cluster config must not be nil")
	}
	if opts.WorkloadConfig == nil {
		return nil, fmt.Errorf("workload cluster config must not be nil")
	}

	clients, err := newProviderClients(rest.CopyConfig(opts.ManagementConfig), rest.CopyConfig(opts.WorkloadConfig))
	if err != nil {
		return nil, err
	}

	p, err := newProviderFromClients(ctx, clients, opts)
	if err != nil {
		return nil, err
	}

	return p, nil
}

func newProviderClients(managementConfig, workloadConfig *rest.Config) (providerClients, error) {
	// Grab a dynamic interface that we can create informers from
	managementClient, err := dynamic.NewForConfig(managementConfig)
	if err != nil {
		return providerClients{}, fmt.Errorf("could not generate dynamic client for config: %w", err)
	}

	workloadClient, err := kubernetes.NewForConfig(workloadConfig)
	if err != nil {
		return providerClients{}, fmt.Errorf("create kube clientset failed: %w", err)
	}

	managementDiscoveryClient, err := discovery.NewDiscoveryClientForConfig(managementConfig)
	if err != nil {
		return providerClients{}, fmt.Errorf("create discovery client failed: %w", err)
	}

	cachedDiscovery := memory.NewMemCacheClient(managementDiscoveryClient)
	managementScaleClient, err := scale.NewForConfig(
		managementConfig,
		restmapper.NewDeferredDiscoveryRESTMapper(cachedDiscovery),
		dynamic.LegacyAPIPathResolverFunc,
		scale.NewDiscoveryScaleKindResolver(managementDiscoveryClient))
	if err != nil {
		return providerClients{}, fmt.Errorf("create scale client failed: %w", err)
	}

	return providerClients{
		management: managementClient,
		workload:   workloadClient,
		discovery:  managementDiscoveryClient,
		scale:      managementScaleClient,
	}, nil
}

func newProviderFromClients(ctx context.Context, clients providerClients, opts ProviderOptions) (*provider, error) {
	ctx, cancel := context.WithCancel(ctx)

	controller, err := newMachineController(clients.management, clients.workload, clients.discovery, clients.scale, opts.DiscoveryOptions, ctx.Done())
	if err != nil {
		cancel()
		return nil, err
	}

	if err := controller.run(); err != nil {
		cancel()
		controller.shutdown()
		return nil, err
	}

	return &provider{
		providerName:    ProviderName,
		resourceLimiter: opts.ResourceLimiter,
		controller:      controller,
		stop:            cancel,
	}, nil
}

// RegisterProcessors wires the Cluster API specific processors into processors.
// cp must be a provider returned by NewProvider or BuildClusterAPI.
func RegisterProcessors(cp cloudprovider.CloudProvider, processors *ca_processors.AutoscalingProcessors) error {
	p, ok := cp.(*provider)
	if !ok {
		return fmt.Errorf("expected a %s cloud provider, got %T", ProviderName, cp)
	}
	if processors == nil {
		return fmt.Errorf("processors must not be nil")
	}

	return scaledowncandidates.RegisterCombinedScaleDownCandidateProcessor(processors.ScaleDownNodeProcessor, NewScaleDownNodeUpgradeProcessor(p.controller))
}
