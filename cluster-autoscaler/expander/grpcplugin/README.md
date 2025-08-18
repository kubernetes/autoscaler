# gRPC Expander for Cluster Autoscaler

## Introduction
This expander functions as a gRPC client, and will pass expansion options to an external gRPC server.
The external server will use this information to make a decision on which Node Group to expand, and return an option to expand.

## Motivation

This expander gives users very fine grained control over which option they'd like to expand.
The gRPC server must be implemented by the user, but the logic can be developed out of band with Cluster Autoscaler.
There are a wide variety of use cases here. Some examples are as follows:
* A tiered weighted random strategy can be implemented, instead of a static priority ladder offered by the priority expander.
* A strategy to encapsulate business logic specific to a user but not all users of Cluster Autoscaler
* A strategy to take into account the dynamic fluctuating prices of the spot instance market

## Configuration options
As using this expander requires communication with another service, users must specify a few options as CLI arguments.

```yaml
--grpcExpanderUrl
```
URL of the gRPC Expander server, for CA to communicate with.
```yaml
--grpcExpanderCert
```
Location of the volume mounted certificate of the gRPC server if it is configured to communicate over TLS

## gRPC Expander Server Setup
The gRPC server can be set up in many ways, but a simple example is described below.
An example of a barebones gRPC Exapnder Server can be found in the `example` directory under `fake_grpc_server.go` file. This is meant to be copied elsewhere and deployed as a separate
service. Note that the `protos/expander.pb.go` generated protobuf code will also need to be copied and used to serialize/deserizle the Options passed from CA.
Communication between Cluster Autoscaler and the gRPC Server will occur over native kube-proxy. To use this, note the Service and Namespace the gRPC server is deployed in.

Deploy the gRPC Expander Server as a separate app, listening on a specifc port number.
Start Cluster Autoscaler with the `--grpcExapnderURl=SERVICE_NAME.NAMESPACE_NAME.svc.cluster.local:PORT_NUMBER` flag, as well as `--grpcExpanderCert` pointed at the location of the volume mounted certificate of the gRPC server.

## Details

The gRPC client currently transforms nodeInfo objects passed into the expander to v1.Node objects to save rpc call throughput. As such, the gRPC server will not have access to daemonsets and static pods running on each node.


