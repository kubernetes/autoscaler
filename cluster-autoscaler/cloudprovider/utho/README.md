# Cluster Autoscaler for Utho

The cluster autoscaler for Utho scales nodes in a Utho Kubernetes cluster.

## Utho Kubernetes Engine

Utho Kubernetes Engine https://utho.com/kubernetes is the managed kubernetes solution provided by Utho.

Utho Kubernetes lets users create Node Pools, i.e. groups of nodes each of the same type.

The size of a Node Pool can be configured at any moment. The user cannot select specific nodes to be deleted when downsizing a Node Pool, rather, Utho Kubernetes will randomly select nodes to be deleted to reach the defined size, even if a node is not healthy or has been manually deleted.

Nodes in a Node Pool are considered disposable: they can be deleted and recreated at any moment, deleting a single node outside of Utho Kubernetes will be recreated by Utho after a small amount of time.

## Configuration

It is mandatory to define the cloud configuration file `cloud-config`.  You can see an example of the cloud config file at [examples/cluster-autoscaler-secret.yaml](examples/cluster-autoscaler-secret.yaml), it is an INI file with the following fields:

The (JSON) configuration file of the Utho cloud provider supports the following values:

- `cluster_id`: the ID of the Utho Kubernetes cluster.
- `token`: the Utho API key literally defined.


Configuring the autoscaler such as if it should be monitoring node pools or what the minimum and maximum values. Should be configured through the [Utho API](https://utho.com/api-docs/#api-Kubernetes).
The autoscaler will pick up any changes and adjust accordingly.

## Development

Make sure you are inside the `cluster-autoscaler` path of the [autoscaler repository](https://github.com/kubernetes/autoscaler).

Create the docker image:
```
make container
```
tag the generated docker image and push it to a registry.
