# Cluster Autoscaler on Cluster API

The cluster autoscaler on [Cluster API](https://cluster-api.sigs.k8s.io/) uses
the [cluster-api project](https://github.com/kubernetes-sigs/cluster-api) to
manage the provisioning and de-provisioning of nodes within a Kubernetes
cluster.

## Kubernetes Version

The cluster-api provider requires Kubernetes v1.16 or greater to run the
v1alpha3 version of the API.

## Cluster API Prerequisites

Please be aware that currently the cluster autoscaler only supports CAPI
clusters that have joined their management and workload clusters into a single
cluster. For more information about this please see the
[Cluster API Concepts documentations](https://cluster-api.sigs.k8s.io/user/concepts.html)
and the [`clusterctl move` command documentation](https://cluster-api.sigs.k8s.io/user/concepts.html).

## Starting the Autoscaler

To enable the Cluster API provider, you must first specify it in the command
line arguments to the cluster autoscaler binary. For example:

```
cluster-autoscaler --cloud-provider=clusterapi
```

Please note, this example only shows the cloud provider options, you will
most likely need other command line flags. For more information you can invoke
`cluster-autoscaler --help` to see a full list of options.

## Enabling Autoscaling

To enable the automatic scaling of components in your cluster-api managed
cloud there are a few annotations you need to provide. These annotations
must be applied to either [MachineSet](https://cluster-api.sigs.k8s.io/developer/architecture/controllers/machine-set.html)
or [MachineDeployment](https://cluster-api.sigs.k8s.io/developer/architecture/controllers/machine-deployment.html)
resources depending on the type of cluster-api mechanism that you are using.

There are two annotations that control how a cluster resource should be scaled:

* `cluster.k8s.io/cluster-api-autoscaler-node-group-min-size` - This specifies
  the minimum number of nodes for the associated resource group. The autoscaler
  will not scale the group below this number. Please note that currently the
  cluster-api provider will not scale down to zero nodes.

* `cluster.k8s.io/cluster-api-autoscaler-node-group-max-size` - This specifies
  the maximum number of nodes for the associated resource group. The autoscaler
  will not scale the group above this number.

The autoscaler will monitor any `MachineSet` or `MachineDeployment` containing
both of these annotations.

## Specifying a Custom Resource Group

By default all Kubernetes resources consumed by the Cluster API provider will
use the group `cluster.x-k8s.io`, with a dynamically acquired version. In
some situations, such as testing or prototyping, you may wish to change this
group variable. For these situations you may use the environment variable
`CAPI_GROUP` to change the group that the provider will use.
