# Disruption Cost Based Scale-Down Ordering

Author: YurDuiachenko

## Background

Cluster Autoscaler can scale down non-empty nodes when their pods can be safely moved elsewhere. During scale-down, CA already evaluates whether a node can be removed by considering node utilization, node group limits, drainability rules, PodDisruptionBudgets, scheduling simulation, and other existing blockers.

However, once multiple non-empty nodes are already removable, CA has limited workload-level information to choose which removable node would be least disruptive to remove first.

The existing `cluster-autoscaler.kubernetes.io/safe-to-evict` annotation is useful for expressing whether a pod can be evicted, but it is mostly binary. It can prevent scale-down entirely for a pod, but it does not allow users to express that one removable pod is more expensive to disrupt than another.

This proposal introduces a soft disruption cost signal that can influence the ordering of already-removable scale-down candidates without changing the existing removability checks.

## High level proposal

Introduce a pod-level disruption cost annotation.

The preferred annotation name is:

```text
node-autoscaling.kubernetes.io/disruption-cost
```

The original issue suggested:

```text
cluster-autoscaler.kubernetes.io/disruption-cost
```

This proposal treats the exact annotation prefix as an open question, because related Karpenter and node-autoscaling work is discussing similar disruption/consolidation ordering semantics.

The annotation value is a unit-less non-negative integer. Higher values mean that disrupting the pod is more expensive. Missing or invalid values are treated as zero.

Cluster Autoscaler will read this annotation from pods that would need to be rescheduled when a node is removed. For each already-removable non-empty node, CA will sum the disruption costs of its reschedulable pods and use this total as a soft ordering signal.

This signal must not make an otherwise non-removable node removable. It only affects ordering among nodes that have already passed existing scale-down safety checks.

### Example

Consider two removable nodes:

```text
node-a:
  pod api-0     disruption-cost=100
  pod worker-0  disruption-cost=20
  total cost=120

node-b:
  pod stateless-0 disruption-cost=0
  pod batch-0     disruption-cost=5
  total cost=5
```

If both nodes are removable, CA should prefer removing `node-b` before `node-a`.

## Goals

* Goals

## Non-goals

* Non-goals

## Annotation semantics

### Scope

The annotation is placed on Pods. In practice, users are expected to set it through workload pod templates.

Example:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: expensive-api
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

### Value

The value is a unit-less non-negative integer.

Suggested initial behavior:

| table              | table                 |
|--------------------|-----------------------|
| missing annotation | 0                     |
| "0"                | 0                     |
| "10"               | 10                    |
| "-1"               | invalid, treated as 0 |
| "abc"              | invalid, treated as 0 | 
| "10.5"             | invalid, treated as 0 |
| overflow           | invalid, treated as 0 |



Invalid values should not block scale-down. They should be ignored and treated as zero. The implementation may log invalid values and may later expose a metric for observability.


## Detailed design

### Existing code integration


### Annotation parsing


### Node cost aggregation


### Candidate ordering


## Relationship with Karpenter and related consolidation work


## Interaction with existing scale-down blockers


## DaemonSet behavior


## On-completion pod behavior


## Monitoring

The initial implementation may rely on logs for invalid annotation values.

Potential follow-up metric:

```text
cluster_autoscaler_invalid_disruption_cost_annotations_total
```

Potential optional debug metric:

```text
cluster_autoscaler_selected_node_disruption_cost
```

Metrics are not required for the first implementation unless maintainers request them.


## Testing



## Risks and mitigations

### Risk: 