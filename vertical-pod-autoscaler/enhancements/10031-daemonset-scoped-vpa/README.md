# AEP-10031: DaemonSet-Scoped Vertical Pod Autoscaler Recommendations

<!--
Greenfield AEP describing the design implemented in
https://github.com/kubernetes/autoscaler/pull/10012

Related discussion / earlier draft:
https://github.com/kubernetes/autoscaler/pull/7942

Replace NNNN with the tracking issue number before opening an AEP PR.
Copy this directory to vertical-pod-autoscaler/enhancements/<NNNN>-daemonset-scope/
-->

<!-- toc -->

- [Summary](#summary)
- [Motivation](#motivation)
  - [Problem](#problem)
  - [Example: Falco on GPU and non-GPU nodes](#example-falco-on-gpu-and-non-gpu-nodes)
  - [Goals](#goals)
  - [Non-Goals](#non-goals)
- [Proposal](#proposal)
- [Design Details](#design-details)
  - [API Changes](#api-changes)
  - [End-to-End Flow](#end-to-end-flow)
  - [Recommender](#recommender)
  - [Admission Controller and Updater](#admission-controller-and-updater)
  - [Validation](#validation)
  - [Feature Gate](#feature-gate)
  - [Absent Labels and Cold Start](#absent-labels-and-cold-start)
  - [Checkpoints and Restart Behavior](#checkpoints-and-restart-behavior)
  - [Performance Considerations](#performance-considerations)
  - [Test Plan](#test-plan)
  - [Feature Enablement and Rollback](#feature-enablement-and-rollback)
  - [Graduation Criteria](#graduation-criteria)
  - [Version Skew](#version-skew)
  - [Kubernetes Version Compatibility](#kubernetes-version-compatibility)
- [Implementation History](#implementation-history)
- [Alternatives](#alternatives)
- [Relationship to AEP-7942](#relationship-to-aep-7942)
<!-- /toc -->

## Summary

Vertical Pod Autoscaler currently computes **one** recommendation per VPA and
applies it to every matching pod. For DaemonSets whose resource needs vary with
the node they run on, that single recommendation is systematically wrong for
part of the fleet.

This AEP adds **DaemonSet scope**: an optional `spec.scope` field naming a
**node label key**. The recommender maintains independent recommendations for
each distinct value of that label and publishes them as `status.groups`.
Admission and updater select the group that matches the pod's node.

The feature is alpha, default-off, behind the `DaemonSetScope` feature gate.

## Motivation

### Problem

DaemonSet pods are not interchangeable replicas. Their CPU and memory usage is
often a function of:

- what workloads run on the node,
- node hardware class (GPU vs CPU, instance type, disk class),
- local traffic or telemetry volume.

VPA's global average recommendation therefore tends to:

- **waste resources** on quiet nodes, and/or
- **starve** the agent on busy or specialized nodes.

Operators today work around this by:

- disabling VPA for those DaemonSets,
- hand-tuning oversized requests everywhere,
- or maintaining multiple DaemonSets / VPAs per node pool — fragile under node
  churn.

### Example: Falco on GPU and non-GPU nodes

[Falco](https://falco.org/) and similar runtime security agents are commonly
deployed as a cluster-wide DaemonSet. On GPU nodes used for ML training or
inference, Falco typically sees much higher process and syscall rates than on
ordinary worker nodes. In production this often means **several times** more
CPU and memory on GPU nodes than on CPU-only nodes.

A single VPA recommendation cannot fit both classes:

| Node class       | What a global VPA tends to do                                |
| ---------------- | ------------------------------------------------------------ |
| CPU-only workers | Over-request Falco (capacity wasted on every node)           |
| GPU workers      | Under-request Falco (throttling, OOMRisk, gaps in detection) |

Clusters already label GPU capacity, for example:

- `nvidia.com/gpu.present=true`
- `gpu=true` (custom)
- `node.kubernetes.io/instance-type=<gpu SKU>`

**Scoped VPA** lets the operator set:

```yaml
spec:
  targetRef:
    kind: DaemonSet
    name: falco
  scope: nvidia.com/gpu.present
```

The recommender then learns separate targets for `true` vs missing/`false`, and
each Falco pod receives the recommendation for its node's class.

The same pattern applies to other node-local agents (Prometheus collectors,
log shippers, CNI / policy agents) whenever a node label correlates with load.

### Goals

1. Partition DaemonSet VPA recommendations by a user-chosen node label key.
2. Expose per-value recommendations in VPA status without breaking clients that
   expect at most one global `status.recommendation`.
3. Apply the matching group at admission and updater time.
4. Preserve legacy behavior when `spec.scope` is empty.
5. Guard the feature with an alpha feature gate and admission validation.

### Non-Goals

1. Scoped recommendations for Deployments, StatefulSets, or Job-like workloads.
2. Cross-group bootstrapping (using other groups to invent an initial
   recommendation for a never-seen label value).
3. Requiring a fixed taxonomy of scopes (`Node`, `NodePool`, cloud-specific
   enums). Any node label key is allowed.
4. Changing how recommendations are _calculated_ (histograms, percentile
   logic, policies) — only _how samples are grouped_.
5. Solving pure "scale with node size" without usage data (see issue #5928).

## Proposal

When `DaemonSetScope` is enabled and a VPA targets a DaemonSet with non-empty
`spec.scope`:

1. **Recommender** groups matching pods by `node.labels[spec.scope]` (or a
   sentinel if the label is absent) and writes one recommendation per group to
   `status.groups`.
2. **`status.recommendation` is cleared** for that VPA so consumers cannot
   mistake a global average for the scoped result.
3. **Admission / updater** resolve the pod's scope value from its node and
   apply the matching group’s container targets.
4. If no group exists yet for that value, VPA does not override the pod’s
   resources from recommendation data (DaemonSet defaults / last applied
   values remain).

Unscoped VPAs are unchanged.

## Design Details

### API Changes

#### `spec.scope`

```go
// VerticalPodAutoscalerSpec
Scope VerticalPodAutoscalerScopeType `json:"scope,omitempty"`

type VerticalPodAutoscalerScopeType string

const (
    ScopeNode VerticalPodAutoscalerScopeType = "kubernetes.io/hostname"
)
```

- Type is a string alias so arbitrary label keys work.
- `ScopeNode` documents the well-known per-node key; it is not the only valid
  value.

#### `status.groups`

```go
type VerticalPodAutoscalerStatus struct {
    Recommendation *RecommendedPodResources           `json:"recommendation,omitempty"`
    Groups         []RecommendedPodResourcesGroup     `json:"groups,omitempty"`
    // Conditions, ObservedGeneration unchanged
}

type RecommendedPodResourcesGroup struct {
    ScopeValue               string                           `json:"scopeValue"`
    ContainerRecommendations []RecommendedContainerResources  `json:"containerRecommendations,omitempty"`
}
```

Alpha payload compaction: for scoped groups, recommender stores **Target only**
(LowerBound / UpperBound / UncappedTarget omitted) to limit object size when the
number of groups is large.

#### Example (Falco)

```yaml
apiVersion: autoscaling.k8s.io/v1
kind: VerticalPodAutoscaler
metadata:
  name: falco
  namespace: falco
spec:
  targetRef:
    apiVersion: apps/v1
    kind: DaemonSet
    name: falco
  scope: nvidia.com/gpu.present
  updatePolicy:
    updateMode: "Auto"
status:
  # recommendation omitted when scoped
  groups:
    - scopeValue: "true"
      containerRecommendations:
        - containerName: falco
          target:
            cpu: 500m
            memory: 512Mi
    - scopeValue: "__absent__"
      containerRecommendations:
        - containerName: falco
          target:
            cpu: 100m
            memory: 128Mi
```

### End-to-End Flow

```text
Node labels                     DaemonSet pods
     |                                |
     v                                v
 cluster feeder copies node label
 into a synthetic pod aggregation label
     |
     v
 recommender aggregates usage per (scopeValue, container)
     |
     v
 writes status.groups[]; clears status.recommendation
     |
     v
 admission/updater: read pod.nodeName -> node.labels[scope]
     |
     v
 select matching status.groups[].containerRecommendations
     |
     v
 mutate / update pod requests (existing capping / policies apply)
```

### Recommender

#### Grouping mechanism

The implementation keeps **a single in-memory VPA** per user object (no N
shadow VPAs in ClusterState).

For each pod matching a scoped DaemonSet VPA, the cluster feeder adds:

```text
podLabels["__vpa_scope_" + hex(fnv32(scopeKey))] = <labelValue or "__absent__">
```

Aggregate container state keys therefore diverge per scope value. The
recommender:

1. Builds `map[scopeValue]ContainerNameToAggregateStateMap`.
2. Runs the existing pod-resource recommender per scope value.
3. Runs recommendation post-processors per group.
4. Strips non-target fields from group recommendations.
5. Updates VPA status: `Groups = ...`, `Recommendation = nil`.
6. Caches `(vpaID, scopedGeneration) -> (recommendation, groups)` so unchanged
   scoped inputs do not rebuild all groups every loop.

#### Global recommendation field

Even for scoped VPAs the recommender still computes an internal global
aggregation for bookkeeping consistent with today’s VPA model, but **does not
publish** it on `status.recommendation` when scope is active.

### Admission Controller and Updater

Selection logic (shared capping / recommendation path):

1. If not a scoped DaemonSet (gate off, empty scope, or non-DaemonSet): use
   `status.recommendation` as today.
2. Resolve scope value:
   - if `spec.scope == kubernetes.io/hostname`, use `pod.spec.nodeName`;
   - else GET Node (cached) and read `labels[spec.scope]`, or `__absent__`.
3. Look up `status.groups` by `scopeValue`.
4. On miss: return empty container recommendations (no speculative fallback).

Node label resolution is cached in admission to avoid a GET per pod on every
webhook call.

### Validation

When creating or updating a VPA:

| Condition                                        | Result                                    |
| ------------------------------------------------ | ----------------------------------------- |
| Gate on, `scope` set, `targetRef.kind=DaemonSet` | Allowed                                   |
| Gate on, `scope` set, other kind                 | Forbidden / invalid                       |
| Gate off, new object with `scope`                | Forbidden                                 |
| Gate off, object that already had `scope`        | Grandfathered (allowed to keep the field) |

### Feature Gate

```text
Name:       DaemonSetScope
Maturity:   Alpha
Default:    false
Since:      VPA 1.8
Components: admission-controller, recommender, updater
```

Enablement example:

```text
--feature-gates=DaemonSetScope=true
```

Must be set on all three components for full behavior.

### Absent Labels and Cold Start

| Case                          | Scope value used   | Recommendation                        |
| ----------------------------- | ------------------ | ------------------------------------- |
| Label present                 | label value string | Matching `status.groups` entry        |
| Label missing                 | `__absent__`       | Matching `__absent__` group if any    |
| Never-seen value / no samples | (no group yet)     | No recommendation applied from groups |
| Pod not bound to a node yet   | empty / unresolved | No group selection                    |

This matches the product decision that alpha will **not** invent an initial
recommendation from other groups (avoids scheduling oversized agents onto small
nodes — a concern raised in AEP-7942 discussions).

### Checkpoints and Restart Behavior

Live aggregation is per scope value (via the synthetic label).

The existing checkpoint writer still keys checkpoints by `(VPA, containerName)`
and merges aggregate states that share a container name. **Alpha limitation:**
after a recommender restart, per-scope histogram separation may not fully
survive checkpoint reload.

**Beta requirement:** checkpoint and restore per
`(VPA, containerName, scopeValue)`, including naming and garbage collection for
stale scope values.

### Performance Considerations

Cardinality is controlled by the operator’s choice of label:

| Scope key                          | Typical #groups | Guidance                            |
| ---------------------------------- | --------------- | ----------------------------------- |
| `nvidia.com/gpu.present`           | 2               | Preferred for Falco-like splits     |
| `node.kubernetes.io/instance-type` | tens            | Good default for hardware classes   |
| `kubernetes.io/hostname`           | ≈ node count    | Highest cost; use when truly needed |

Mitigations already in the implementation:

- scoped recommendation cache across recommender iterations;
- compact group status (targets only);
- admission node-label cache;
- benchmarks for ~1000 groups on the ProcessVPAUpdate / aggregation paths.

### Test Plan

**Unit**

- `IsScopedDaemonSet` respects the feature gate.
- Feeder injects synthetic labels; uses `__absent__` when node label missing.
- Recommender populates `status.groups` and clears `status.recommendation`.
- Admission selects the correct group for GPU vs non-GPU labels.
- Validation matrix for gate / kind / grandfathering.

**Benchmarks**

- DaemonSet scope with 1000 groups (update + aggregation).

**e2e (alpha)**

- Kind cluster with two node label classes and a DaemonSet VPA scoped on that
  label; assert distinct mutated requests.
- Gate-off rejection of `spec.scope` on create.

### Feature Enablement and Rollback

- **Enable:** set `DaemonSetScope=true` on admission, recommender, updater;
  create scoped VPAs.
- **Disable:** stop writing/consuming groups; validation blocks new scoped
  specs (existing fields remain stored). Running pods keep last-applied
  requests until recreated.
- **Re-enable:** recommender rebuilds groups from live metrics (checkpoint
  fidelity limited in alpha — see above).

### Graduation Criteria

**Alpha**

- Gate + API fields + unit tests + documented examples (including GPU label).
- Known checkpoint limitation documented.

**Beta**

- Per-scope checkpoints.
- e2e on Testgrid.
- At least one production-like report for a heterogeneous DaemonSet (e.g.
  security or observability agent on mixed GPU clusters).
- No breaking changes to `status.groups` shape.

**GA**

- Two releases in beta.
- Stable performance envelope documented for hostname scope at supported
  cluster sizes.
- User-facing docs in kubernetes/autoscaler (and optionally kubernetes.io).

### Version Skew

| Recommender | Admission  | Effect                                             |
| ----------- | ---------- | -------------------------------------------------- |
| scoped on   | scoped on  | Full feature                                       |
| scoped on   | scoped off | Groups written, ignored; pods keep defaults (safe) |
| scoped off  | scoped on  | No groups; no scoped mutation (safe)               |

Because scoped mode clears `status.recommendation`, skew does not accidentally
apply a misleading global average.

### Kubernetes Version Compatibility

Depends only on core Pod/Node APIs. No minimum Kubernetes feature gate beyond
what VPA already requires.

## Implementation History

- 2026-07-21: Implementation opened as
  [PR #10012](https://github.com/kubernetes/autoscaler/pull/10012).
- 2026-07-22: This greenfield AEP draft written to mirror #10012 exactly,
  including the Falco / GPU-label motivating example.

## Alternatives

1. **Multiple DaemonSets / VPAs per node pool** — operational overhead; poor
   fit for dynamic GPU node provisioning.
2. **Only support `kubernetes.io/hostname`** — solves per-node variance but
   forces maximum cardinality even when a 2-way GPU split is enough (Falco).
3. **Fixed enum scopes (`Node`, `NodePool`)** — fights cloud-specific and
   custom labels; rejected in favor of arbitrary node label keys (also the
   direction AEP-7942 converged on).
4. **Keep publishing a global `status.recommendation` alongside groups** —
   unsafe for naive clients; rejected.
5. **Shadow VPA objects per scope value in ClusterState** — more churn; the
   synthetic-label approach reuses existing aggregation machinery.
6. **Cross-group initial recommendation** — deferred; risk of unschedulable
   pods on smaller nodes.

## Relationship to AEP-7942

[AEP PR #7942](https://github.com/kubernetes/autoscaler/pull/7942) introduced
the same high-level idea (`spec.scope` as a node label key for DaemonSets).
This document is a self-contained rewrite that:

- pins the **status API** (`status.groups`, cleared `status.recommendation`);
- describes the **actual recommender/admission design** from PR #10012;
- adds the **Falco / GPU label** motivating example;
- states alpha/beta expectations for **checkpoints**.

If maintainers prefer a single lineage, the content here can be merged into
#7942 instead of landing as a separate AEP number.
