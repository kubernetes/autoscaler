# Plugable Cloud Provider over gRPC

## Motivation

CA is released as a bundle which includes a hardcoded list of supported cloud providers.
Whenever users want to implement the logic of their own cloud provider, they need to fork the CA and add their own implementation.

In particular users need to follow these steps to support a custom private cloud:

* Write a client for your private cloud in Go, implementing CloudProvider interface.

* Add constructing it to cloud provider builder.

* Build a custom image of Cluster Autoscaler that includes those changes and configure it to start with your cloud provider.

This is a concern that has been raised in the past [PR953](https://github.com/kubernetes/autoscaler/issues/953) and [PR1060](https://github.com/kubernetes/autoscaler/issues/1060).

Therefore a new implemetation should be added to CA in order to extend it without breaking any backwards compatibility or 
the current cloud provider implementations.

## Goals

* Support custom cloud provider implementations without changing the current `CloudProvider` interface.
* Make CA extendable, so users do not need to fork the CA repository.

## Proposal

There are couple of examples of plugable designs using Go SDKs that would guide us on how to extend CA
to support custom providers as plugins.
This approach is inspired based on [Hashicorp go-plugin](https://github.com/hashicorp/go-plugin) and [Grafana Go SDK for plugins](https://github.com/grafana/grafana-plugin-sdk-go).

There are two alternatives on how to plug custom cloud providers:

* **Option1:** Install the plugin as a binary that would be mounted into the CA container
and invoked by CA server.
In this option CA server launches each provider plugin as a subprocess and communicates with it over gRPC.

* **Option2:** A custom cloud provider server is deployed along side CA and both communicates via gRPC with TLS/SSL.

Regardless of the chosen option, both solutions have to expose a common gRPC API
with the following operations:

```go
type clusterAutoscalerCustomProviderServer struct {
        ...
}

func (s *clusterAutoscalerCustomProviderServer) NodeGroups(ctx context.Context) ([]pb.NodeGroup, error) {
        ...
}
...

func (s *clusterAutoscalerCustomProviderServer) NodeGroupForNode(ctx context.Context, *apiv1.Node) (*pb.NodeGroup, error) {
        ...
}
...

func (s *clusterAutoscalerCustomProviderServer) Refresh() error {
        ...
}
...

func (s *clusterAutoscalerCustomProviderServer) GetAvailableMachineTypes() ([]string, error) {
        ...
}

...

```

Obviously, these API calls implement the `CloudProvider` interface of the [CA](https://github.com/kubernetes/autoscaler/blob/master/cluster-autoscaler/cloudprovider/cloud_provider.go#L50):

```go
type CloudProvider interface {
	Name() string

	NodeGroups() []NodeGroup

	NodeGroupForNode(*apiv1.Node) (NodeGroup, error)

	Pricing() (PricingModel, errors.AutoscalerError)

	GetAvailableMachineTypes() ([]string, error)

	NewNodeGroup(machineType string, labels map[string]string, systemLabels map[string]string,
		taints []apiv1.Taint, extraResources map[string]resource.Quantity) (NodeGroup, error)

	GetResourceLimiter() (*ResourceLimiter, error)

	GPULabel() string

	GetAvailableGPUTypes() map[string]struct{}

	Cleanup() error

	Refresh() error

  ...
}
```

In order to talk to the custom cloud provider server, this new cloud provider has to be registered
when bootstrapping the CA. 
Consequently, the CA needs to expose new flags to specify the cloud provider and all the required properties
to reach the remote gRPC server.

A new flag, named
`--cloud-provider-url=https://local.svc.io/mycloudprovider/server`, determines the URL to reach the custom provider implementation.
In addition this approach reuses the existing flag that defines the name of the cloud provider `--cloud-provider=mycloudprovider`
https://github.com/kubernetes/autoscaler/tree/master/cluster-autoscaler/cloudprovider.

To connect the CA core with this new remote cloud provider, this approach needs to implement a new generic cloud provider 
as part of the CA core code.
This new provider, named `CustomCloudProvider` simply makes gRPC calls to the remote functions exposed by the custom cloud provider server. In other words, it forwards the calls and handle the errors analogously how done in other existing providers.

Obviously, this new apprach needs to use TLS to ensure a secure communication between CA and this CA provider server.
The flags need to be defined [TODO].

## User Stories

### Story1

When using CA only a reduced list of cloud providers are supported, if users want to use their own private cloud provider (e.g. Openstack, OpenNebula,...), they need to implement its cloud provider interface for that environment.
This locked design limits the extensibility of CA and goes against certain native Kubernetes primitives.

This approach aims to make CA plugable, so any user can implement the logic of their own provider
following a pre-defined gRPC interface.
