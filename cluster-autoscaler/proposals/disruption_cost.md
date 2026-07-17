# Disruption Cost in Scale-Down Ordering

Author: YurDuiachenko

## Background

Currently, there is no common workload-level API for expressing disruption cost of a pod across different node autoscalers:

- Cluster Autoscaler does not consider pod disruption cost in the scale-down process, so it can pick nodes whose removal will be more disruptive than others. 
- Karpenter uses `controller.kubernetes.io/pod-deletion-cost` for node disruption scoring, but it overloads a ReplicaSet deletion signal with node disruption semantics.

Users can also prevent specific pods from eviction by using autoscaler-specific annotations such as `cluster-autoscaler.kubernetes.io/safe-to-evict` or `karpenter.sh/do-not-disrupt`, 
but it delays or completely blocks the scale-down process rather than ordering it.

## High level proposal

The proposal is to introduce a new user-facing autoscaler-agnostic annotation:

```text
node-autoscaling.kubernetes.io/disruption-cost
```

Node autoscalers will read this annotation from pods, with higher values indicating a higher relative disruption cost. 
Each autoscaler can use its own algorithms for aggregation and candidate selection.

Cluster Autoscaler, for example, will sum the disruption costs of the pods that need to be rescheduled and use this total to choose which node to remove next.

If there are two removable nodes:

```text
node-a:
  pod api-0     disruption-cost=100
  pod worker-0  disruption-cost=20
  total cost=120

node-b:
  pod api-1     disruption-cost=5
  pod worker-1  disruption-cost=10
  total cost=15
```

`node-b` should be removed before `node-a`.

## Goals

* Allow users to express the relative disruption cost of evicting a pod.
* Define a common API that can be consumed by different node autoscalers.
* Add another layer of ordering to the scale-down process.
* Protect long-running processes from being shut down and made to start over.

## Detailed design

### Annotation

The annotation is placed on Pods. In practice, users are expected to set it through workload pod templates.

For example:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: app
spec:
  template:
    metadata:
      annotations:
        node-autoscaling.kubernetes.io/disruption-cost: "100"
    spec:
      containers:
      - name: app
        image: example/app
```

The common annotation semantics are:

* The annotation is set by user.
* Node autoscalers only read the annotation and never change it.
* A higher value represents that disrupting the pod is relatively more expensive.
* The annotation is a best-effort preference rather than a hard disruption constraint. 

The annotation value is a unitless non-negative integer in the range `[0, MaxInt32]`, so missing, unparsable, negative or out-of-range value should be treated as zero.

Examples:

| Value              | seen as               |
| ------------------ |-----------------------|
| missing annotation | 0                     |
| `"0"`              | 0                     |
| `"10"`             | 10                    |
| `"-1"`             | invalid, treated as 0 |
| `"abc"`            | invalid, treated as 0 |
| `"10.5"`           | invalid, treated as 0 |
| overflow           | invalid, treated as 0 |

### Difference with Pod Deletion Cost

Kubernetes already defines the `controller.kubernetes.io/pod-deletion-cost` annotation, so why is it preferred to introduce a new annotation over reusing the existing one?

While both annotations are placed on pods and have similar naming, during the scale-down process `controller.kubernetes.io/pod-deletion-cost` is used to hint at the deletion cost of a pod compared to other pods within the same **_ReplicaSet_**,
whereas `node-autoscaling.kubernetes.io/disruption-cost` is used to hint at the disruption cost of a pod compared across different **_nodes_**.

Therefore, reusing the existing annotation would overload it with two independent meanings and consumers.
The ReplicaSet controller needs a signal for which pods should be deleted first, and node autoscalers need a signal for which pods are more expensive to disrupt during node removal or consolidation. 
These values are not necessarily the same and may even point in opposite directions.

### Implementation in Cluster Autoscaler

Disruption cost should influence two stages of the Cluster Autoscaler scale-down process:

1. the order in which scale-down candidates are evaluated before node-removal simulation;
2. the final ordering of already-removable non-empty nodes.

Applying disruption cost only during final node deletion ordering is not sufficient, 
because candidate evaluation order affects which nodes are marked as unneeded.

For example, consider that either node A or node B can be removed, but not both:

```text
node-a:
  total disruption cost=10

node-b:
  total disruption cost=100
```

If node B is evaluated first, the scale-down planner may determine that its pods can be moved to node A and mark node B as unneeded. Node A may then never be marked as unneeded. By the time final removable-node ordering is performed, node B may be the only available candidate, leaving no alternative for disruption-cost ordering to prefer.

#### Candidate evaluation ordering

TODO

#### Final removable-node ordering

The final logic will be implemented in `Planner.NodesToDelete()`.

After CA has already calculated removable nodes:

```go
emptyRemovableNodes, needDrainRemovableNodes, unremovableNodes :=
    p.unneededNodes.RemovableAt(...)
```

`needDrainRemovableNodes` are ordered with `sortByRisk`:

```go
needDrainRemovableNodes = sortByRisk(needDrainRemovableNodes)
```

That ordering should be extended with disruption cost:

```go
needDrainRemovableNodes = sortByRiskAndDisruptionCost(needDrainRemovableNodes)
```

The updated function should preserve the existing `riskyNodes` and `okNodes` grouping and use disruption cost as an additional ordering signal within each group.

The total disruption cost for a node should be calculated as the sum of annotation values on pods listed in `NodeToBeRemoved.PodsToReschedule`.


### Relationship to Karpenter

TODO

### Corner cases

* An annotation set on `DaemonSetPods` has no impact on the node disruption cost.
* `PodDisruptionBudget`s keep their current behavior.
* Pods with `safe-to-evict=false` should keep blocking scale-down.

## Testing

The following unit test scenarios should be added:

* [TC1] Missing, invalid, negative, non-integer, and overflowing annotation values are treated as zero.
* [TC2] Valid annotation values are parsed and summed for pods in `NodeToBeRemoved.PodsToReschedule`.
* [TC3] Nodes with lower total disruption cost are preferred within the same existing ordering group.
* [TC4] Existing `riskyNodes` and `okNodes` ordering is preserved.