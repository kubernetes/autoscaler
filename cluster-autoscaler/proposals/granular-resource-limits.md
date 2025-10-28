# Granular Resource Limits in Node Autoscalers

## Objective

Node Autoscalers should allow setting more granular resource limits that would
apply to arbitrary subsets of nodes, beyond the existing limiting mechanisms.

## Background

Cluster Autoscaler supports cluster-wide limits on resources (like total CPU and
memory) and per-node-group node count limits. Karpenter supports
setting [resource limits on a NodePool](https://karpenter.sh/docs/concepts/nodepools/#speclimits).
Also, as mentioned
in [AWS docs](https://docs.aws.amazon.com/eks/latest/best-practices/karpenter.html),
cluster-wide limits are not supported too. This is not flexible enough for many
use cases.

Users often need to configure more granular limits. For instance, a user might
want to limit the total resources consumed by nodes of a specific machine
family, nodes with a particular OS, or nodes with specialized hardware like
GPUs. The current resource limits implementations in both node autoscalers do
not support these scenarios.

This proposal introduces a new API to extend the Node Autoscalers’
functionality, allowing limits to be applied to arbitrary sets of nodes.

## Proposal: The CapacityQuota API

We propose a new Kubernetes custom resource, CapacityQuota, to define
resource limits on specific subsets of nodes. Node subsets are targeted using
standard Kubernetes label selectors, offering a flexible way to group nodes.

A node's eligibility for provisioning operation will be checked against all
CapacityQuota objects that select it. The operation will only be
permitted if it does not violate any of the applicable limits. This should be
compatible with the existing limiting mechanisms, i.e. CAS’ cluster-wide limits
and Karpenter’s NodePool limits. Therefore, if the operation doesn’t violate
CapacityQuota, but violates existing limiting mechanisms, it should
be rejected.

### API Specification

An CapacityQuota object would look as follows:

```yaml
apiVersion: autoscaling.x-k8s.io/v1beta1
kind: CapacityQuota
metadata:
  name: example-resource-quota
spec:
  selector:
    matchLabels:
      example.cloud.com/machine-family: e2
  limits:
    resources:
      cpu: 64
      memory: 256Gi
```

* `selector`: A standard Kubernetes label selector that determines which nodes
  the limits apply to. This allows for fine-grained control based on any label
  present on the nodes, such as zone, region, OS, machine family, or custom
  user-defined labels.
* `limits`: Defines the limits of summed up resources of the selected nodes.
    * `resources`: map of resources (e.g. `cpu`, `memory`) that should be
      limited. This map could be put directly into `limits`, but we put it here
      insteade for the sake of extensibility. For instance, if we were to
      support DRA limits via this API, we would probably define them in a
      separate field under `limits`.

This approach is highly flexible – adding a new dimension for limits only
requires ensuring the nodes are labeled appropriately, with no code changes
needed in the autoscaler.

### Node as a Resource

The CapacityQuota API can be naturally extended to treat the number
of nodes itself as a limitable resource, as shown in one of the examples below.

### CapacityQuota Status

For better observability, the CapacityQuota resource could be
enhanced with a status field. This field, updated by a controller, would display
the current resource usage for the selected nodes, allowing users to quickly
check usage against the defined limits via kubectl describe. The controller can
run in a separate thread as a part of the node autoscaler component.

An example of the status field:

```yaml
status:
  used:
    resources:
      cpu: 32
      memory: 128Gi
      nodes: 50
```

## Alternatives considered

### Minimum limits support

The initial design, besides the maximum limits, also included minimum limits.
Minimum limits were supposed to affect the node consolidation in the node
autoscalers. A consolidation would be allowed only if removing the node wouldn’t
violate any minimum limits. Cluster-wide minimum limits are implemented in CAS
together with the maximum limits, so at first, it seemed logical to include both
limit directions in the design.

Despite being conceptually similar, minimum and maximum limits cover completely
different use cases. Maximum limits can be used to control the cloud provider
costs, to limit scaling certain types of compute, or to control distribution of
compute resources between teams working on the same cluster. Minimum limits’
main use case is ensuring a baseline capacity for users’ workloads, for example
to handle sudden spikes in traffic. However, minimum limits defined as a minimum
amount of resources in the cluster or a subset of nodes do not guarantee that
the workloads will be schedulable on those resources. For example, two nodes
with 2 CPUs each satisfy the minimum limit of 4 CPUs. If a user created a
workload requesting 2 CPUs, that workload would not fit into existing nodes,
making the baseline capacity effectively useless. This scenario will be better
handled by
the [CapacityBuffer API](https://github.com/kubernetes/autoscaler/blob/master/cluster-autoscaler/proposals/buffers.md),
which allows the user to provide an exact shape of their workloads, including
the resource requests. In our example, the user would create a CapacityBuffer
with a pod template requesting 2 CPUs. Such a CapacityBuffer would ensure that a
pod with that shape is always schedulable on the existing nodes.

Therefore, we decided to remove minimum limits from the design of granular
limits, as CapacityBuffers are a better way to provide a baseline capacity for
user workloads.

### Kubernetes LimitRange and ResourceQuota

It has been discussed whether the same result could be accomplished by using the
standard Kubernetes
resources: [LimitRange](https://kubernetes.io/docs/concepts/policy/limit-range/)
and [ResourceQuota](https://kubernetes.io/docs/concepts/policy/resource-quotas/).

LimitRange is a resource used to configure minimum and maximum resource
constraints for a namespace. For example, it can define the default CPU and
memory requests for pods and containers within a namespace, or enforce a minimum
and maximum CPU request for a pod. However, its scope is limited to a single
resource, meaning that it doesn’t look at all pods in the namespace, but just
looks if the pod requests and limits are within defined bounds.

ResourceQuota allows to define and limit the aggregate resource consumption per
namespace. This includes limiting the total CPU, memory, and storage that all
pods and persistent volume claims within a namespace can request or consume. It
also supports limiting the count of various Kubernetes objects, such as pods,
services, and replication controllers. While resource quotas can be used to
limit the resources provisioned by the CA to some degree, it’s not possible to
guarantee that CA won’t scale up above the defined limit. Since the quotas
operate on pod requests, and CA does not guarantee that bin packing will yield
the optimal result, setting the quota to e.g. 64 CPUs does not mean that CA will
stop scaling at 64 CPUs.

Moreover, both of those resources are namespaced, so their scope is limited to
the namespace in which they are defined, while the nodes are global. We can’t
use namespaced resources to limit the creation and deletion of global resources.

### Soft and hard limits

We have discussed a possibility of distinguishing soft and hard limits. That
idea was initially presented for Karpenter node limits in
https://github.com/kubernetes-sigs/karpenter/pull/2525. The currently existing
limits in Karpenter behave like soft limits -- they are best effort, meaning
that they can be exceeded, for example due to race conditions. They are
respected only during the provisioning operations, while they can be exceeded
during the node consolidation. Proposed hard limits would not be exceeded also
during the node consolidation. In Cluster Autoscaler, at this moment there is
no viable use case for such a distinction. Moreover, enforcing hard limits could
be complex to achieve due to concurrency concerns. Because of that, and for the
sake of simplicity of this API, we have decided to support only one type of
limits. The behavior of the limits should be in line with the current
implementation of the limits in the Node Autoscalers.

## User Stories

### Story 1

As a cluster administrator, I want to configure cluster-wide resource limits to
avoid excessive cloud provider costs.

**Note:** This is already supported in CAS, but not in Karpenter.

Example CapacityQuota:

```yaml
apiVersion: autoscaling.x-k8s.io/v1beta1
kind: CapacityQuota
metadata:
  name: cluster-wide-limits
spec:
  limits:
    resources:
      cpu: 128
      memory: 256Gi
```

### Story 2

As a cluster administrator, I want to configure separate resource limits for
specific groups of nodes on top of cluster-wide limits, to avoid a situation
where one group of nodes starves others of resources.

**Note:** A specific group of nodes can be either a NodePool in Karpenter, a
ComputeClass in GKE, or simply a set of nodes grouped by a user-defined label.
This can be useful e.g. for organizations where multiple teams are running
workloads in a shared cluster, and these teams have separate sets of nodes. This
way, a cluster administrator can ensure that each team has a proper limit for
their resources and it doesn’t starve other teams. This story is partly
supported by Karpenter’s NodePool limits.

Example CapacityQuota:

```yaml
apiVersion: autoscaling.x-k8s.io/v1beta1
kind: CapacityQuota
metadata:
  name: team-a-limits
spec:
  selector:
    matchLabels:
      team: a
  limits:
    resources:
      cpu: 32
```

### Story 3

As a cluster administrator, I want to allow scaling up machines that are more
expensive or less suitable for my workloads when better machines are
unavailable, but I want to limit how many of them can be created, so that I can
control extra cloud provider costs, or limit the impact of using non-optimal
machine for my workloads.

Example CapacityQuota:

```yaml
apiVersion: autoscaling.x-k8s.io/v1beta1
kind: CapacityQuota
metadata:
  name: max-e2-resources
spec:
  selector:
    matchLabels:
      example.cloud.com/machine-family: e2
  limits:
    resources:
      cpu: 32
      memory: 64Gi
```

### Story 4

As a cluster administrator, I want to limit the number of nodes in a specific
zone if my cluster is unbalanced for any reason, so that I can avoid exhausting
IP space in that zone, or enforce better balancing across zones.

**Note:** Originally requested
in [https://github.com/kubernetes/autoscaler/issues/6940](https://github.com/kubernetes/autoscaler/issues/6940).

Example CapacityQuota:

```yaml
apiVersion: autoscaling.x-k8s.io/v1beta1
kind: CapacityQuota
metadata:
  name: max-nodes-us-central1-b
spec:
  selector:
    matchLabels:
      topology.kubernetes.io/zone: us-central1-b
  limits:
    resources:
      nodes: 64
```

### Story 5 (obsolete)

As a cluster administrator, I want to ensure there is always a baseline capacity
in my cluster or specific parts of my cluster below which the node autoscaler
won’t consolidate the nodes, so that my workloads can quickly react to sudden
spikes in traffic.

This user story is obsolete. CapacityBuffer API covers this use case in a more
flexible way.

## Other CapacityQuota examples

The following examples illustrate the flexibility of the proposed API and
demonstrate other possible use cases not described in the user stories.

#### **Maximum Windows Nodes**

Limit the total number of nodes running the Windows operating system to 8.

```yaml
apiVersion: autoscaling.x-k8s.io/v1beta1
kind: CapacityQuota
metadata:
  name: max-windows-nodes
spec:
  selector:
    matchLabels:
      kubernetes.io/os: windows
  limits:
    resources:
      nodes: 8
```

#### **Maximum NVIDIA T4 GPUs**

Limit the total number of NVIDIA T4 GPUs in the cluster to 16.

```yaml
apiVersion: autoscaling.x-k8s.io/v1beta1
kind: CapacityQuota
metadata:
  name: max-t4-gpus
spec:
  selector:
    matchLabels:
      example.cloud.com/gpu-type: nvidia-t4
  limits:
    resources:
      nvidia.com/gpu: 16
```

#### **Cluster-wide Limits Excluding Control Plane Nodes**

Apply cluster-wide CPU and memory limits while excluding nodes with the
control-plane role.

```yaml
apiVersion: autoscaling.x-k8s.io/v1beta1
kind: CapacityQuota
metadata:
  name: cluster-limits-no-control-plane
spec:
  selector:
    matchExpressions:
    - key: node-role.kubernetes.io/control-plane
      operator: DoesNotExist
  limits:
    resources:
      cpu: 64
      memory: 128Gi
```
