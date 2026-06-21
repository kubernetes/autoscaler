# Disruption Cost in Scale-Down Ordering

Author: YurDuiachenko

## Background

When scaling down a cluster, CA is currently not aware of the disruption "cost" action will have, so it can pick node which will be more disruptive than others. Users can prevent Cluster Autoscaler from evicting specific pods by using `safe-to-evict=false` annotation, but it rather complitely blocks scaling-down rather than expressing relative disruption cost.

## High level proposal

The proposal is to introduce a new pod-level user-facing annotation:

```text
node-autoscaling.kubernetes.io/disruption-cost
```

Cluster Autoscaler will read this annotation from pods that would need to be rescheduled when a node is removed. For each already-removable non-empty node, CA will sum the disruption costs of pods and use this total to pick up which node remove next.

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

* Add another layer of ordering on scale-down process
* Protect more long running processes from shutting it down and make them to starting over.

## Detailed design

### Annotation

The annotation is placed on Pods. For deployment it can be to set through workload pod templates:

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

Cluster Autoscaler only reads the annotation. It does not write or mutate it.

The annotation value is a unit-less non-negative integer. Higher values mean that disrupting the pod is more expensive. Missing or invalid values are treated as zero.

The overall behavior:


| Value              | is                    |
| ------------------ | --------------------- |
| missing annotation | 0                     |
| "0"                | 0                     |
| "10"               | 10                    |
| "-1"               | invalid, treated as 0 |
| "abc"              | invalid, treated as 0 |
| "10.5"             | invalid, treated as 0 |
| overflow           | invalid, treated as 0 |


### Candidate ordering

## Blocking behavior

safe-to-evict
Daemonset
On-complition

### Existing code integration

## Monitoring

```text
cluster_autoscaler_invalid_disruption_cost_annotations_total
```

Potential optional debug metric:

```text
cluster_autoscaler_selected_node_disruption_cost
```

## Testing
