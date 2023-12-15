# External gRPC Cloud Provider

The External gRPC Cloud Provider provides a plugin system to support out-of-tree cloud provider implementations.

Cluster Autoscaler adds or removes nodes from the cluster by creating or deleting VMs. To separate the autoscaling logic (the same for all clouds) from the API calls required to execute it (different for each cloud), the latter are hidden behind an interface, `CloudProvider`. Each supported cloud has its own implementation in this repository and `--cloud-provider` flag determines which one will be used.

The gRPC Cloud Provider acts as a client for a cloud provider that implements its custom logic separately from the cluster autoscaler, and serves it as a `CloudProvider` gRPC service (similar to the `CloudProvider` interface) without the need to fork this project, follow its development lifecyle, adhere to its rules (e.g. do not use additional external dependencies) or implement the Cluster API.

## Configuration

For the cluster autoscaler parameters, use the `--cloud-provider=externalgrpc` flag and define the cloud configuration file with `--cloud-config=<file location>`, this is yaml file with the following parameters:

| Key | Value | Mandatory | Default |
|-----|-------|-----------|---------|
| address | external gRPC cloud provider service address of the form "host:port", "host%zone:port", "[host]:port" or "[host%zone]:port" | yes | none |
| key | path to file containing the tls key, if using mTLS | no | none |
| cert | path to file containing the tls certificate, if using mTLS | no | none |
| cacert | path to file containing the CA certificate, if using mTLS | no | none |
| grpc_timeout | timeout of invoking a grpc call | no | 5s |

The use of mTLS is recommended, since simple, non-authenticated calls to the external gRPC cloud provider service will result in the creation / deletion of nodes.

Log levels of interest for this provider are:
* 1 (flag: ```--v=1```): basic logging of errors;
* 5 (flag: ```--v=5```): detailed logging of every call;

For the deployment and configuration of an external gRPC cloud provider of choice, see its specific documentation.

## Examples

You can find an example of external gRPC cloud provider service implementation on the [examples/external-grpc-cloud-provider-service](examples/external-grpc-cloud-provider-service) directory: it is actually a server that wraps all the in-tree cloud providers.

A complete example:
* deploy `cert-manager` and the manifests in [examples/certmanager-manifests](examples/certmanager-manifests) to generate certificates for gRPC client and server;
* build the image for the example external gRPC cloud provider service as defined in [examples/external-grpc-cloud-provider-service](examples/external-grpc-cloud-provider-service);
* deploy the example external gRPC cloud provider service using the manifests at [examples/external-grpc-cloud-provider-service-manifests](examples/external-grpc-cloud-provider-service-manifests), change the parameters as needed and test whichever cloud provider you want;
* deploy the cluster autoscaler selecting the External gRPC Cloud Provider using the manifests at [examples/cluster-autoscaler-manifests](examples/cluster-autoscaler-manifests).

## Development

### External gRPC Cloud Provider service Implementation

To build a cloud provider, create a gRPC server for the `CloudProvider` service defined in [protos/externalgrpc.proto](protos/externalgrpc.proto) that implements all its required RPCs.

### Caching

The `CloudProvider` interface was designed with the assumption that its implementation functions would be fast, this may not be true anymore with the added overhead of gRPC. In the interest of performance, some gRPC API responses are cached by this cloud provider:
* `NodeGroupForNode()` caches the node group for a node until `Refresh()` is called;
* `NodeGroups()` caches the current node groups until `Refresh()` is called;
* `GPULabel()` and `GetAvailableGPUTypes()` are cached at first call and never wiped;
* A `NodeGroup` caches `MaxSize()`, `MinSize()` and `Debug()` return values during its creation, and `TemplateNodeInfo()` at its first call, these values will be cached for the lifetime of the `NodeGroup` object.

### Code Generation

To regenerate the gRPC code:

1. install `protoc` and `protoc-gen-go-grpc`:

```bash
go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.31
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.3
```

2. generate gRPC client and server code:

```bash
protoc \
  -I ./cluster-autoscaler \
  -I ./cluster-autoscaler/vendor \
  --go_out=. \
  --go-grpc_out=. \
  ./cluster-autoscaler/cloudprovider/externalgrpc/protos/externalgrpc.proto
```

### General considerations

Abstractions used by Cluster Autoscaler assume nodes belong to "node groups". All node within a group must be of the same machine type (have the same amount of resources), have the same set of labels and taints, and be located in the same availability zone. This doesn't mean a cloud has to have a concept of such node groups, but it helps.

There must be a way to delete a specific node. If your cloud supports instance groups, and you are only able to provide a method to decrease the size of a given group, without guaranteeing which instance will be killed, it won't work well.

There must be a way to match a Kubernetes node to an instance it is running on. This is usually done by kubelet setting node's `ProviderId` field to an instance id which can be used in API calls to cloud.
