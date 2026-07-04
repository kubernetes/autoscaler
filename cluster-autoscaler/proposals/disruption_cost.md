# Disruption Cost in Scale-Down Ordering

Author: YurDuiachenko

## Background

When scaling down a cluster, Cluster Autoscaler (CA) is currently not aware of the disruption "cost" the action will have, so it can pick a node that will be more disruptive than others. Users can prevent Cluster Autoscaler from evicting specific pods by using the `safe-to-evict=false` annotation, but it completely blocks scale-down rather than expressing relative disruption cost.

## High level proposal

The proposal is to introduce a new pod-level user-facing annotation:

```text
node-autoscaling.kubernetes.io/disruption-cost
```

Cluster Autoscaler will read this annotation from pods that would need to be rescheduled when a node is removed. For each already-removable non-empty node, CA will sum the disruption costs of the pods and use this total to choose which node to remove next.

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

CA should prefer removing `node-b` before `node-a`.

## Goals

* Allow users to express the relative disruption cost of evicting a pod.
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

CA handles the annotation as follows:

* Cluster Autoscaler only reads the annotation and does not change it.
* The value is a unitless non-negative integer.
* A higher value means the pod is more expensive to disrupt.
* If the annotation is missing or cannot be parsed, CA treats the cost as zero.

Examples:

| Value              | CA sees it as         |
| ------------------ |-----------------------|
| missing annotation | 0                     |
| `"0"`              | 0                     |
| `"10"`             | 10                    |
| `"-1"`             | invalid, treated as 0 |
| `"abc"`            | invalid, treated as 0 |
| `"10.5"`           | invalid, treated as 0 |
| overflow           | invalid, treated as 0 |

### Existing code integration

The main logic will be implemented in `Planner.NodesToDelete()`.

After CA has already calculated removable nodes:

```go
emptyRemovableNodes, needDrainRemovableNodes, unremovableNodes :=
    p.unneededNodes.RemovableAt(...)
```

`needDrainRemovableNodes` are ordered with `sortByRisk`:

```go
needDrainRemovableNodes = sortByRisk(needDrainRemovableNodes)
```

The proposal extends that ordering with disruption cost:

```go
needDrainRemovableNodes = sortByRiskAndDisruptionCost(needDrainRemovableNodes)
```

The updated function should preserve the existing `riskyNodes` and `okNodes` grouping and use disruption cost as an additional ordering signal within each group.

The total disruption cost for a node should be calculated as the sum of annotation values on pods listed in `NodeToBeRemoved.PodsToReschedule`.

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