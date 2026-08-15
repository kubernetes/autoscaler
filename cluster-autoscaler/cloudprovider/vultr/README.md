# Cluster Autoscaler for Vultr

The cluster autoscaler for Vultr scales nodes in a VKE cluster.

## Vultr Kubernetes Engine

Vultr Kubernetes Engine ([VKE](https://www.vultr.com/docs/vultr-kubernetes-engine/)) is the managed kubernetes solution provided by Vultr.

VKE lets users create Node Pools, i.e. groups of nodes each of the same type.

The size of a Node Pool can be configured at any moment. The user cannot select specific nodes to be deleted when downsizing a Node Pool, rather, VKE will randomly select nodes to be deleted to reach the defined size, even if a node is not healthy or has been manually deleted.

Nodes in a Node Pool are considered disposable: they can be deleted and recreated at any moment, deleting a single node outside of VKE will be recreated by Vultr after a small amount of time.


## Configuration

It is mandatory to define the cloud configuration file `cloud-config`.  You can see an example of the cloud config file at [examples/cluster-autoscaler-secret.yaml](examples/cluster-autoscaler-secret.yaml), it is an INI file with the following fields:

The (JSON) configuration file of the Vultr cloud provider supports the following values:

- `cluster_id`: the ID of the VKE cluster.
- `token`: the Vultr API key literally defined.


Configuring the autoscaler such as if it should be monitoring node pools or what the minimum and maximum values. Should be configured through the [Vultr API](https://www.vultr.com/api/#tag/kubernetes).
The autoscaler will pick up any changes and adjust accordingly.

## Development

Make sure you are inside the `cluster-autoscaler` path of the [autoscaler repository](https://github.com/kubernetes/autoscaler).

Create the docker image:
```
make container
```
tag the generated docker image and push it to a registry.
