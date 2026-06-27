# Disruption Cost in Scale-Down Ordering

Author: YurDuiachenko

## Background

When scaling down a cluster, Cluster Autoscaler (CA) is currently not aware of the disruption "cost" the action will have, so it can pick a node that will be more disruptive than others. Users can prevent Cluster Autoscaler from evicting specific pods by using the safe-to-evict=false annotation, but it completely blocks scale-down rather than expressing relative disruption cost.

## High level proposal

The proposal is to introduce a new pod-level user-facing annotation:

```text
node-autoscaling.kubernetes.io/disruption-cost
```

Cluster Autoscaler will read this annotation from pods that would need to be rescheduled when a node is removed. For each already-removable non-empty node, CA will sum the disruption costs of the pods and use this total to choose which node to remove next.

If there is two removable nodes:

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

Cluster Autoscaler only reads the annotation and never changes it.

The value is a unit-less non-negative integer. A higher value means the pod is more expensive to disrupt. If the annotation is missing or cannot be parsed, CA treats the cost as zero.

| Value              | CA sees it as         |
| ------------------ |-----------------------|
| missing annotation | 0                     |
| `"0"`              | 0                     |
| `"10"`             | 10                    |
| `"-1"`             | invalid, treated as 0 |
| `"abc"`            | invalid, treated as 0 |
| `"10.5"`           | invalid, treated as 0 |
| overflow           | invalid, treated as 0 |

### Candidate ordering


### Blocking behavior

`safe-to-evict`

PodDisruptionBudgets

DaemonSet pods

Completed pods




### Existing code refactoring

### Relationship to Karpenter

`karpenter.sh/disruption-cost`


## Monitoring



```text
cluster_autoscaler_invalid_disruption_cost_annotations_total
```



```text
cluster_autoscaler_selected_node_disruption_cost
```

## Testing
