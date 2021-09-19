# Cluster Autoscaler for Linode

The cluster autoscaler for Linode scales nodes in a LKE cluster.

## Linode Kubernetes Engine

Linode Kubernetes Engine ([LKE](https://www.linode.com/docs/guides/deploy-and-manage-a-cluster-with-linode-kubernetes-engine-a-tutorial/)) is the managed K8s solution provided by Linode.

LKE lets users create Node Pools, i.e. groups of nodes (also called Linodes) each of the same type.

The size of a Node Pool can be configured at any moment. The user cannot select specific nodes to be deleted when downsizing a Node Pool, rather, Linode will randomly select nodes to be deleted to reach the defined size, even if a node is not healthy or has been manually deleted.

Nodes in a Node Pool are considered disposable: they can be deleted and recreated at any moment, deleting a single node or using the *recycle* feature, on these cases the node will be recreated by Linode after a small amount of time.

Node Pools do not support user defined tags or labels.

There is no limitation on the number of Node Pool a LKE Cluster can have, limited to the maximum number of nodes an LKE Cluster can have.

## Cluster Autoscaler

A cluster autoscaler node group is composed of multiple LKE Node Pools of the same Linode type (e.g. g6-standard-1, g6-standard-2), each holding a *single* Linode.

At every scan interval the cluster autoscaler reviews every LKE Node Pool and if:
* it is not among the ones to be excluded as defined in the configuration file;
* it holds a single Linode;

then it becomes part of the node group holding LKE Node Pools of the Linode type it has, or a node group is created with this LKE Node Pool inside if there are no node groups with that Linode type at the moment.

Scaling is achieved adding LKE Node Pools to node groups, *not* increasing the size of a LKE Node Pool, that must stay 1. The reason behind this is that Linode does not provide a way to selectively delete a Linode from a LKE Node Pool and decrease the size of the pool with it.

This is also the reason we cannot use the standard `nodes` and `node-group-auto-discovery` cluster autoscaler flag (no labels could be used to select a specific node group), and the reason why there can be no node group of the same type.

## Configuration

The cluster autoscaler automatically select every LKE Node Pool that is part of a LKE cluster, so there is no need define the `node-group-auto-discovery` or `nodes` flags, see [examples/cluster-autoscaler-autodiscover.yaml](examples/cluster-autoscaler-autodiscover.yaml) for an example of a kubernetes deployment.

It is mandatory to define the cloud configuration file `cloud-config`.
You can see an example of the cloud config file at [examples/cluster-autoscaler-secret.yaml](examples/cluster-autoscaler-secret.yaml), it is an INI file with the following fields:

| Key | Value | Mandatory | Default |
|-----|-------|-----------|---------|
| global/linode-token | Linode API Token with Read/Write permission for Kubernetes and Linodes | yes | none |
| global/lke-cluster-id | ID of the LKE cluster (numeric of the form: 12345, you can get this via `linode-cli` or looking at the first number of a linode in a pool, e.g. for lke15989-19461-5fec9212fad2 the lke-cluster-id is "15989") | yes | none |
| global/defaut-min-size-per-linode-type | minimum size of a node group (must be > 0) | no | 1 |
| global/defaut-max-size-per-linode-type | maximum size of a node group | no | 254 |
| global/do-not-import-pool-id | Pool id (numeric of the form: 12345) that will be excluded from the pools managed by the cluster autoscaler; can be repeated | no | none
| nodegroup \"linode_type\"/min-size" | minimum size for a specific node group | no | global/defaut-min-size-per-linode-type |
| nodegroup \"linode_type\"/max-size" | maximum size for a specific node group | no | global/defaut-min-size-per-linode-type |

Log levels of intertest for the Linode provider are:
* 1 (flag: ```--v=1```): basic logging at start;
* 2 (flag: ```--v=2```): logging of the node group composition at every scan;

## Development

Make sure you are inside the `cluster-autoscaler` path of the [autoscaler repository](https://github.com/kubernetes/autoscaler).

Create the docker image:
```
make container
```
tag the generated docker image and push it to a registry.
