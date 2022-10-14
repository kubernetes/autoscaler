# Cluster Autoscaler on Cluster API

The cluster autoscaler on [Cluster API](https://cluster-api.sigs.k8s.io/) uses
the [cluster-api project](https://github.com/kubernetes-sigs/cluster-api) to
manage the provisioning and de-provisioning of nodes within a Kubernetes
cluster.

## Kubernetes Version

The cluster-api provider requires Kubernetes v1.16 or greater to run the
v1alpha3 version of the API.

## Starting the Autoscaler

To enable the Cluster API provider, you must first specify it in the command
line arguments to the cluster autoscaler binary. For example:

```
cluster-autoscaler --cloud-provider=clusterapi
```

Please note, this example only shows the cloud provider options, you will
most likely need other command line flags. For more information you can invoke
`cluster-autoscaler --help` to see a full list of options.

## Configuring node group auto discovery

If you do not configure node group auto discovery, cluster autoscaler will attempt
to match nodes against any scalable resources found in any namespace and belonging
to any Cluster.

Limiting cluster autoscaler to only match against resources in the blue namespace

```
--node-group-auto-discovery=clusterapi:namespace=blue
```

Limiting cluster autoscaler to only match against resources belonging to Cluster test1

```
--node-group-auto-discovery=clusterapi:clusterName=test1
```

Limiting cluster autoscaler to only match against resources matching the provided labels

```
--node-group-auto-discovery=clusterapi:color=green,shape=square
```

These can be mixed and matched in any combination, for example to only match resources
in the staging namespace, belonging to the purple cluster, with the label owner=jim:

```
--node-group-auto-discovery=clusterapi:namespace=staging,clusterName=purple,owner=jim
```

## Connecting cluster-autoscaler to Cluster API management and workload Clusters

You will also need to provide the path to the kubeconfig(s) for the management
and workload cluster you wish cluster-autoscaler to run against. To specify the
kubeconfig path for the workload cluster to monitor, use the `--kubeconfig`
option and supply the path to the kubeconfig. If the `--kubeconfig` option is
not specified, cluster-autoscaler will attempt to use an in-cluster configuration.
To specify the kubeconfig path for the management cluster to monitor, use the
`--cloud-config` option and supply the path to the kubeconfig. If the
`--cloud-config` option is not specified it will fall back to using the kubeconfig
that was provided with the `--kubeconfig` option.

### Autoscaler running in a joined cluster using service account credentials
```
+-----------------+
| mgmt / workload |
| --------------- |
|    autoscaler   |
+-----------------+
```
Use in-cluster config for both management and workload cluster:
```
cluster-autoscaler --cloud-provider=clusterapi
```

### Autoscaler running in workload cluster using service account credentials, with separate management cluster
```
+--------+              +------------+
|  mgmt  |              |  workload  |
|        | cloud-config | ---------- |
|        |<-------------+ autoscaler |
+--------+              +------------+
```

Use in-cluster config for workload cluster, specify kubeconfig for management cluster:
```
cluster-autoscaler --cloud-provider=clusterapi \
                   --cloud-config=/mnt/kubeconfig
```

### Autoscaler running in management cluster using service account credentials, with separate workload cluster
```
+------------+             +----------+
|    mgmt    |             | workload |
| ---------- | kubeconfig  |          |
| autoscaler +------------>|          |
+------------+             +----------+
```

Use in-cluster config for management cluster, specify kubeconfig for workload cluster:
```
cluster-autoscaler --cloud-provider=clusterapi \
                   --kubeconfig=/mnt/kubeconfig \
                   --clusterapi-cloud-config-authoritative
```

### Autoscaler running anywhere, with separate kubeconfigs for management and workload clusters
```
+--------+               +------------+             +----------+
|  mgmt  |               |     ?      |             | workload |
|        |  cloud-config | ---------- | kubeconfig  |          |
|        |<--------------+ autoscaler +------------>|          |
+--------+               +------------+             +----------+
```

Use separate kubeconfigs for both management and workload cluster:
```
cluster-autoscaler --cloud-provider=clusterapi \
                   --kubeconfig=/mnt/workload.kubeconfig \
                   --cloud-config=/mnt/management.kubeconfig
```

### Autoscaler running anywhere, with a common kubeconfig for management and workload clusters
```
+---------------+             +------------+
| mgmt/workload |             |     ?      |
|               |  kubeconfig | ---------- |
|               |<------------+ autoscaler |
+---------------+             +------------+
```

Use a single provided kubeconfig for both management and workload cluster:
```
cluster-autoscaler --cloud-provider=clusterapi \
                   --kubeconfig=/mnt/workload.kubeconfig
```

## Enabling Autoscaling

To enable the automatic scaling of components in your cluster-api managed
cloud there are a few annotations you need to provide. These annotations
must be applied to either [MachineSet](https://cluster-api.sigs.k8s.io/developer/architecture/controllers/machine-set.html)
or [MachineDeployment](https://cluster-api.sigs.k8s.io/developer/architecture/controllers/machine-deployment.html)
resources depending on the type of cluster-api mechanism that you are using.

There are two annotations that control how a cluster resource should be scaled:

* `cluster.x-k8s.io/cluster-api-autoscaler-node-group-min-size` - This specifies
  the minimum number of nodes for the associated resource group. The autoscaler
  will not scale the group below this number. Please note that the cluster-api
  provider will not scale down to, or from, zero unless that capability is enabled
  (see [Scale from zero support](#scale-from-zero-support)).

* `cluster.x-k8s.io/cluster-api-autoscaler-node-group-max-size` - This specifies
  the maximum number of nodes for the associated resource group. The autoscaler
  will not scale the group above this number.

The autoscaler will monitor any `MachineSet` or `MachineDeployment` containing
both of these annotations.

### Scale from zero support

The Cluster API community has defined an opt-in method for infrastructure
providers to enable scaling from zero-sized node groups in the
[Opt-in Autoscaling from Zero enhancement](https://github.com/kubernetes-sigs/cluster-api/blob/main/docs/proposals/20210310-opt-in-autoscaling-from-zero.md).
As defined in the enhancement, each provider may add support for scaling from
zero to their provider, but they are not required to do so. If you are expecting
built-in support for scaling from zero, please check with the Cluster API
infrastructure providers that you are using.

If your Cluster API provider does not have support for scaling from zero, you
may still use this feature through the capacity annotations. You may add these
annotations to your MachineDeployments, or MachineSets if you are not using
MachineDeployments (it is not needed on both), to instruct the cluster
autoscaler about the sizing of the nodes in the node group. At the minimum,
you must specify the CPU and memory annotations, these annotations should
match the expected capacity of the nodes created from the infrastructure.

For example, if my MachineDeployment will create nodes that have "16000m" CPU,
"128G" memory, 2 NVidia GPUs, and can support 200 max pods, the folllowing
annotations will instruct the autoscaler how to expand the node group from
zero replicas:

```yaml
apiVersion: cluster.x-k8s.io/v1alpha4
kind: MachineDeployment
metadata:
  annotations:
    cluster.x-k8s.io/cluster-api-autoscaler-node-group-max-size: "5"
    cluster.x-k8s.io/cluster-api-autoscaler-node-group-min-size: "0"
    capacity.cluster-autoscaler.kubernetes.io/memory: "128G"
    capacity.cluster-autoscaler.kubernetes.io/cpu: "16"
    capacity.cluster-autoscaler.kubernetes.io/gpu-type: "nvidia.com/gpu"
    capacity.cluster-autoscaler.kubernetes.io/gpu-count: "2"
    capacity.cluster-autoscaler.kubernetes.io/maxPods: "200"
```

*Note* the `maxPods` annotation will default to `110` if it is not supplied.
This value is inspired by the Kubernetes best practices
[Considerations for large clusters](https://kubernetes.io/docs/setup/best-practices/cluster-large/).

#### RBAC changes for scaling from zero

If you are using the opt-in support for scaling from zero as defined by the
Cluster API infrastructure provider, you will need to add the infrastructure
machine template types to your role permissions for the service account
associated with the cluster autoscaler deployment. The service account will
need permission to `get` and `list` the infrastructure machine templates for
your infrastructure provider.

For example, when using the [Kubemark provider](https://github.com/kubernetes-sigs/cluster-api-provider-kubemark)
you will need to set the following permissions:

```yaml
rules:
  - apiGroups:
    - infrastructure.cluster.x-k8s.io
    resources:
    - kubemarkmachinetemplates
    verbs:
    - get
    - list
```

#### Pre-defined labels and taints on nodes scaled from zero

The Cluster API provider currently does not support the addition of pre-defined
labels and taints for node groups that are scaling from zero. This work is on-going
and will be included in a future release once the API for specifying those
labels and taints has been accepted by the community.

## Specifying a Custom Resource Group

By default all Kubernetes resources consumed by the Cluster API provider will
use the group `cluster.x-k8s.io`, with a dynamically acquired version. In
some situations, such as testing or prototyping, you may wish to change this
group variable. For these situations you may use the environment variable
`CAPI_GROUP` to change the group that the provider will use.

Please note that setting the `CAPI_GROUP` environment variable will also cause the
annotations for minimum and maximum size to change.
This behavior will also affect the machine annotation on nodes, the machine deletion annotation,
and the cluster name label. For example, if `CAPI_GROUP=test.k8s.io`
then the minimum size annotation key will be `test.k8s.io/cluster-api-autoscaler-node-group-min-size`,
the machine annotation on nodes will be `test.k8s.io/machine`, the machine deletion
annotation will be `test.k8s.io/delete-machine`, and the cluster name label will be
`test.k8s.io/cluster-name`.

## Specifying a Custom Resource Version

When determining the group version for the Cluster API types, by default the autoscaler
will look for the latest version of the group. For example, if `MachineDeployments`
exist in the `cluster.x-k8s.io` group at versions `v1alpha1` and `v1beta1`, the
autoscaler will choose `v1beta1`.

In some cases it may be desirable to specify which version of the API the cluster
autoscaler should use. This can be useful in debugging scenarios, or in situations
where you have deployed multiple API versions and wish to ensure that the autoscaler
uses a specific version.

Setting the `CAPI_VERSION` environment variable will instruct the autoscaler to use
the version specified. This works in a similar fashion as the API group environment
variable with the exception that there is no default value. When this variable is not
set, the autoscaler will use the behavior described above.

## Sample manifest

A sample manifest that will create a deployment running the autoscaler is
available. It can be deployed by passing it through `envsubst`, providing
these environment variables to set the namespace to deploy into as well as the image and tag to use:

```
export AUTOSCALER_NS=kube-system
export AUTOSCALER_IMAGE=us.gcr.io/k8s-artifacts-prod/autoscaling/cluster-autoscaler:v1.20.0
envsubst < examples/deployment.yaml | kubectl apply -f-
```

### A note on permissions
The `cluster-autoscaler-management` role for accessing cluster api scalable resources is scoped to `ClusterRole`.
This may not be ideal for all environments (eg. Multi tenant environments).
In such cases, it is recommended to scope it to a `Role` mapped to a specific namespace.


## Autoscaling with ClusterClass and Managed Topologies

For users using [ClusterClass and Managed Topologies](https://cluster-api.sigs.k8s.io/tasks/experimental-features/cluster-class/index.html) the Cluster Topology controller attempts to set MachineDeployment replicas based on the `spec.topology.workers.machineDeployments[].replicas` field. In order to use the Cluster Autoscaler this field can be left unset in the Cluster definition.

The below Cluster definition shows which field to leave unset:

```yaml
apiVersion: cluster.x-k8s.io/v1beta1
kind: Cluster
metadata:
  name: "my-cluster"
  namespace: default
spec:
  clusterNetwork:
    services:
      cidrBlocks: ["10.128.0.0/12"]
    pods:
      cidrBlocks: ["192.168.0.0/16"]
    serviceDomain: "cluster.local"
  topology:
    class: "quick-start"
    version: v1.24.0
    controlPlane:
      replicas: 1 
    workers:
      machineDeployments:
        - class: default-worker
          name: linux
       ## replicas field is not set. 
       ## replicas: 1
```

**Warning**: If the Autoscaler is enabled **and** the replicas field is set for a `MachineDeployment` or `MachineSet` the Cluster may enter a broken state where replicas become unpredictable.

If the replica field is unset in the Cluster definition Autoscaling can be enabled [as described above](#enabling-autoscaling)
