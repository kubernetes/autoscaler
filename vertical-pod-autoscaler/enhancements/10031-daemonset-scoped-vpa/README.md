# AEP-10031: DaemonSet-Scoped Vertical Pod Autoscaler Recommendations

<!--
AEP describing the design implemented in
https://github.com/kubernetes/autoscaler/pull/10012
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
    - [<code>spec.scope</code>](#specscope)
    - [<code>status.recommendationGroups</code>](#statusrecommendationgroups)
    - [<code>VerticalPodAutoscalerCheckpoint</code>](#verticalpodautoscalercheckpoint)
    - [Example (Falco)](#example-falco)
  - [End-to-End Flow](#end-to-end-flow)
  - [Recommender](#recommender)
    - [Grouping mechanism](#grouping-mechanism)
    - [Cache lifecycle](#cache-lifecycle)
  - [Admission Controller and Updater](#admission-controller-and-updater)
  - [Validation](#validation)
  - [Feature Gate](#feature-gate)
  - [Missing Scope Label and Cold Start](#missing-scope-label-and-cold-start)
  - [Checkpoints and Restart Behavior](#checkpoints-and-restart-behavior)
  - [Performance Considerations](#performance-considerations)
  - [Test Plan](#test-plan)
  - [Feature Enablement and Rollback](#feature-enablement-and-rollback)
  - [Graduation Criteria](#graduation-criteria)
  - [Version Skew](#version-skew)
  - [Kubernetes Version Compatibility](#kubernetes-version-compatibility)
- [Implementation History](#implementation-history)
- [Alternatives](#alternatives)
<!-- /toc -->

## Summary

Vertical Pod Autoscaler currently computes **one** recommendation per VPA and
applies it to every matching pod. For DaemonSets whose resource needs vary with
the node they run on, that single recommendation is systematically wrong for
part of the fleet.

This AEP adds **DaemonSet scope**: an optional `spec.scope` field naming a
**node label key**. The recommender maintains independent recommendations for
each distinct value of that label and publishes them as
`status.recommendationGroups`, while keeping the aggregate
`status.recommendation` as a fallback. Admission and updater select the group
that matches the pod's node.

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

| Node class       | What a global VPA tends to do                                 |
| ---------------- | ------------------------------------------------------------- |
| CPU-only workers | Over-request Falco (capacity wasted on every node)            |
| GPU workers      | Under-request Falco (throttling, OOM risk, gaps in detection) |

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

The recommender then learns separate targets for `true` vs missing, and
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
   These are out of scope for this AEP; support for other controllers, if
   needed, would be proposed as a separate AEP behind its own feature gate.
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
   `status.recommendationGroups`.
2. **`status.recommendation` keeps the aggregate recommendation** for the whole
   DaemonSet. It is a safe fallback for clients that do not consume
   `recommendationGroups` (for example after the feature gate is disabled during
   a rollback), so scoped mode never leaves pods without a recommendation.
3. **Admission / updater** resolve the pod's scope value from its node and
   apply the matching group's container targets, falling back to the global
   recommendation when no group matches or the gate is off.
4. If no group exists yet for that value and the fallback recommendation is not
   available either, VPA does not override the pod's resources from
   recommendation data (DaemonSet defaults / last applied values remain).

Unscoped VPAs are unchanged.

Worked example: a DaemonSet runs on three nodes — node A labeled `gpu=true`,
node B labeled `gpu=false`, and node C without the `gpu` label — with
`spec.scope: gpu`. The recommender produces three groups:

- `scopeValue: "true"` for node A's pod,
- `scopeValue: "false"` for node B's pod,
- `scopeValue: "__absent__"` for node C's pod (the label key is missing).

Each pod receives the target from its node's group, while
`status.recommendation` holds the aggregate across all three pods.

## Design Details

### API Changes

#### `spec.scope`

```go
// VerticalPodAutoscalerSpec
Scope VerticalPodAutoscalerScopeType `json:"scope,omitempty"`

type VerticalPodAutoscalerScopeType string
```

- Type is a string alias so any node label key works (cluster-specific labels
  such as GPU labels are first-class).
- The well-known per-node key is `kubernetes.io/hostname`; it is not special-
  cased in the API, only handled as a fast path when resolving the scope value
  (see [Admission Controller and Updater](#admission-controller-and-updater)).

#### `status.recommendationGroups`

```go
type VerticalPodAutoscalerStatus struct {
    Recommendation       *RecommendedPodResources       `json:"recommendation,omitempty"`
    RecommendationGroups []RecommendedPodResourcesGroup `json:"recommendationGroups,omitempty"`
    // Conditions, ObservedGeneration unchanged
}

type RecommendedPodResourcesGroup struct {
    // ScopeValue is the value of the node label named by spec.scope shared by
    // the pods in this group. The sentinel "__absent__" is used when the node
    // does not carry that label key.
    ScopeValue               string                          `json:"scopeValue"`
    ContainerRecommendations []RecommendedContainerResources `json:"containerRecommendations,omitempty"`
}
```

To keep the status object small when there are many groups, each group carries
only the effective `target` for its containers; `lowerBound`, `upperBound` and
`uncappedTarget` are omitted from groups. The global `status.recommendation`
still carries the full bounds.

#### `VerticalPodAutoscalerCheckpoint`

Scope-aware history requires each checkpoint to record which scope value it
belongs to, so a restarted recommender restores per-scope aggregates instead of
cold-starting:

```go
type VerticalPodAutoscalerCheckpointSpec struct {
    VPAObjectName string `json:"vpaObjectName,omitempty"`
    ContainerName string `json:"containerName,omitempty"`
    // ScopeValue is the value of the spec.scope node label key this checkpoint
    // belongs to when the VPA uses DaemonSet scoping. Empty for non-scoped VPAs,
    // so existing checkpoints keep their meaning.
    ScopeValue string `json:"scopeValue,omitempty"`
}
```

- **Backward compatible:** the field is optional and empty for every existing
  (non-scoped) checkpoint, so old checkpoints load unchanged.
- **Object naming:** non-scoped checkpoints keep the historical
  `<vpaName>-<container>` name; scoped checkpoints append a short FNV-32 hash of
  the scope value (`<vpaName>-<container>-<hash>`), so arbitrary label values
  cannot produce an invalid or over-long (RFC 1123) object name.
- **Persisted sentinels:** an empty scope value is stored as `__empty__` (to
  tell it apart from a non-scoped checkpoint); a node missing the scope label
  key uses `__absent__`, consistent with `status.recommendationGroups`.

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
    updateMode: "Recreate"
status:
  recommendation:
    containerRecommendations:
      - containerName: falco
        target:
          cpu: 250m
          memory: 300Mi
  recommendationGroups:
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
 writes status.recommendationGroups[]; keeps global status.recommendation
     |
     v
 admission/updater: resolve node (pod.spec.nodeName or the metadata.name
 nodeAffinity matchField) -> node.labels[scope]
     |
     v
 select matching status.recommendationGroups[].containerRecommendations,
 or fall back to status.recommendation
     |
     v
 mutate / update pod requests (existing capping / policies apply)
```

### Recommender

#### Grouping mechanism

The implementation keeps **a single in-memory VPA** per user object (no N
shadow VPAs in ClusterState). Grouping reuses the existing aggregation
machinery without changing how recommendations are calculated.

`scopeKey` is the value of `spec.scope` — a node label key such as
`nvidia.com/gpu.present`. To partition samples by that node label without
colliding with the pod's own labels, the cluster feeder adds a **synthetic pod
label** for each pod matching a scoped DaemonSet VPA:

```text
podLabels["__vpa_scope_" + hex(fnv32(scopeKey))] = <node label value or "__absent__">
```

- The `__vpa_scope_` prefix plus a hash of `scopeKey` yields a stable, short
  key that cannot clash with user labels.
- The value is the node's value for `spec.scope`, or the sentinel `__absent__`
  when the node does not carry that label key.

Aggregate container state keys therefore diverge per scope value. The
recommender:

1. Builds `map[scopeValue]ContainerNameToAggregateStateMap`.
2. Runs the existing pod-resource recommender per scope value.
3. Runs recommendation post-processors per group.
4. Strips non-target fields from group recommendations (targets only).
5. Updates VPA status: sets `RecommendationGroups`, and keeps
   `Recommendation` populated with the aggregate over all matching pods.

#### Cache lifecycle

- Aggregations for pods on removed nodes age out through the existing VPA
  aggregate garbage collection (`DeleteRemovedPods` plus normal aggregate
  ageing), so scope values whose nodes disappear stop contributing and their
  groups fall away.
- Admission's node-label cache and the capping group lookup are described in
  [Admission Controller and Updater](#admission-controller-and-updater).

### Admission Controller and Updater

Selection logic (shared capping / recommendation path):

1. If not a scoped DaemonSet (gate off, empty scope, or non-DaemonSet): use
   `status.recommendation` as today.
2. Resolve the pod's node:
   - use `pod.spec.nodeName` if it is already set;
   - otherwise read the node name from the `metadata.name` matchField in the
     pod's required `nodeAffinity`, which the DaemonSet controller injects
     (since Kubernetes 1.12 DaemonSet pods are scheduled by the default
     scheduler and have no `spec.nodeName` at admission time).
3. Resolve the scope value:
   - if `spec.scope` is the hostname label (`kubernetes.io/hostname`), the
     scope value is the node name itself (fast path, no Node GET);
   - otherwise GET (or read from cache) the Node object and use
     `labels[spec.scope]`, or `__absent__` if the key is missing.
4. Look up `status.recommendationGroups` by `scopeValue`.
5. On miss: fall back to the global `status.recommendation`. If that is also
   empty, return an empty container recommendation set (no speculative
   per-group value is invented).

Node label resolution is cached in admission (short TTL) to avoid a Node GET
per pod on every webhook call. The capping group lookup is keyed by the VPA's
`resourceVersion` and scope, so it is rebuilt when the VPA changes.

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

### Missing Scope Label and Cold Start

"Missing label" here means the node does not carry the label key named by
`spec.scope`. Scope resolution and cold start behave as follows:

| Case                          | Scope value used   | Recommendation applied                  |
| ----------------------------- | ------------------ | --------------------------------------- |
| Label present                 | label value string | Matching `recommendationGroups` entry   |
| Label missing on the node     | `__absent__`       | Matching `__absent__` group if present  |
| Never-seen value / no samples | (no group yet)     | Global `status.recommendation` fallback |
| Pod not bound to a node yet   | empty / unresolved | Global `status.recommendation` fallback |

For a brand-new node or a never-seen label value, VPA does not invent a
per-group recommendation from other groups (that is a non-goal). Until the
group has enough samples, the pod uses the global recommendation fallback, or —
if no recommendation exists at all yet — the requests from the DaemonSet
manifest.

### Checkpoints and Restart Behavior

Live aggregation is partitioned per scope value via the synthetic label, and
checkpoints are **scope-aware** so this partitioning survives recommender
restarts and rollbacks:

- The checkpoint writer persists one checkpoint per
  `(VPA, containerName, scopeValue)` **in addition to** the global
  `(VPA, containerName)` checkpoint. The global checkpoint backs
  `status.recommendation`; the per-scope checkpoints back
  `status.recommendationGroups`.
- Checkpoint object names embed a short hash of the scope value so arbitrary
  label values stay within Kubernetes object-name limits, and the empty scope
  value is stored with a distinct sentinel so it is not confused with the
  non-scoped checkpoint.

Garbage collection:

- When a VPA is deleted, the existing checkpoint garbage collector removes all
  of its checkpoints — global and per-scope — because they share the same
  `spec.vpaObjectName`.
- A **stale scope value** whose checkpoint outlives the label value in the
  cluster (while the VPA still exists — for example a node label value that no
  longer occurs anywhere) is not actively removed. Such a checkpoint is inert
  and at worst reappears as an empty group after a restart. This is consistent
  with how VPA already retains per-container checkpoints and is a general
  checkpoint garbage-collection concern rather than something specific to this
  feature; a dedicated cleanup, if ever wanted, would be a separate enhancement
  and needs no API change here.

### Performance Considerations

The number of groups is controlled by the operator's choice of label. A poor
choice — especially `kubernetes.io/hostname` — increases recommender CPU and
memory and the size of the VPA `status` object, because it creates one group
per node. Kubernetes supports clusters of up to ~5000 nodes, so a hostname
scope can produce up to ~5000 groups for a single VPA. Operators should prefer
coarse-grained labels (GPU present, instance type, node pool) and reserve
hostname scope for cases where load is genuinely per-node.

Design decisions that keep this workable:

- **Compact group status:** groups store only the effective `target`, keeping
  the `status` object small even with many groups.
- **Admission node-label cache:** node label lookups are cached (short TTL) to
  avoid a Node GET per pod.
- **Benchmarks:** the `ProcessVPAUpdate` and aggregation paths are benchmarked
  at both 1000 and 5000 groups to confirm the cost scales roughly linearly with
  the number of groups and stays acceptable at the maximum supported cluster
  size.

We measure recommender loop duration, allocations on the scoped
`ProcessVPAUpdate` / aggregation paths, memory growth, and `status` object size
as a function of the number of groups. The target is that overhead scales
roughly linearly with the number of groups and remains acceptable for hostname
scope on large clusters.

### Test Plan

**Unit**

- `IsScopedDaemonSet` respects the feature gate.
- Feeder injects synthetic labels; uses `__absent__` when the node label is
  missing.
- Recommender populates `status.recommendationGroups` and keeps the global
  `status.recommendation` as a fallback.
- Capping / recommendation provider selects the correct group by node label,
  uses the hostname fast path, and falls back to the global recommendation on a
  miss or when the gate is disabled.
- Checkpoint writer persists per-scope checkpoints and the cluster feeder
  restores them into the correct per-scope aggregate state.
- Validation matrix for gate / kind / grandfathering.

**Benchmarks**

- DaemonSet scope with 1000 and 5000 groups (`ProcessVPAUpdate` and aggregation).

**e2e (alpha)**

- Kind cluster with three node classes for the scope label — two distinct
  label values plus one node without the label — and a DaemonSet VPA scoped on
  that label; assert distinct recommendation groups (including `__absent__` for
  the unlabeled node) and that the global recommendation is kept.
- Gate-off rejection of `spec.scope` on create.

### Feature Enablement and Rollback

- **Enable:** set `DaemonSetScope=true` on admission, recommender, updater;
  create scoped VPAs.
- **Disable (rollback) does not lose data:** the recommender stops publishing
  `recommendationGroups`, and admission/updater fall back to the global
  `status.recommendation`, which is always kept for scoped VPAs. Validation
  blocks new scoped specs while existing fields remain stored. Per-scope
  history is preserved in scope-aware checkpoints, so re-enabling the gate
  restores per-group recommendations without a cold start.
- **Re-enable:** the recommender rebuilds groups from live metrics and the
  restored per-scope checkpoints.

### Graduation Criteria

The feature gate controls rollout, not maturity. The whole feature ships in
alpha. Beta and GA only change the gate's default and then lock it; they add no
functionality. Out-of-scope work (see [Non-Goals](#non-goals)) would be a
separate AEP with its own feature gate.

**Alpha** — gate defaults to `false` (opt-in). Delivered in full:

- Feature gate, `spec.scope` and `status.recommendationGroups` API, the
  `VerticalPodAutoscalerCheckpoint` `scopeValue` field, admission validation,
  and the global-recommendation fallback.
- Scope-aware checkpoints (per `(VPA, container, scopeValue)`).
- Unit coverage for validation, feeder, recommender, capping/admission selection, and checkpoints.
- Benchmarks for 1000 and 5000 groups.
- e2e coverage for scoped grouping (including `__absent__`).
- User-facing documentation and examples in [github.com/kubernetes/autoscaler](https://github.com/kubernetes/autoscaler).

**Beta** — gate defaults to `true` (users can still opt out). No functional or
API changes; this is only the default-on rollout step.

**GA** — gate is locked to `true`. No functional or API changes; the API
(`spec.scope`, `status.recommendationGroups`, checkpoint `scopeValue`) is
declared stable.

### Version Skew

| Recommender | Admission  | Effect                                                                      |
| ----------- | ---------- | --------------------------------------------------------------------------- |
| scoped on   | scoped on  | Full feature                                                                |
| scoped on   | scoped off | Groups written but ignored; admission uses the global recommendation (safe) |
| scoped off  | scoped on  | No groups; admission uses the global recommendation (safe)                  |

Because scoped mode keeps the global `status.recommendation`, any skew degrades
gracefully to the aggregate recommendation instead of leaving pods without one.

### Kubernetes Version Compatibility

Depends only on core Pod/Node APIs. No minimum Kubernetes feature gate beyond
what VPA already requires.

## Implementation History

- 2026-07-21: Implementation opened as
  [PR #10012](https://github.com/kubernetes/autoscaler/pull/10012).
- 2026-07-22: AEP aligned to the implementation: global `status.recommendation`
  kept as a fallback, `status.recommendationGroups` naming, scope-aware
  checkpoints, e2e and 5000-group benchmarks, and the Falco / GPU-label
  motivating example.

## Alternatives

1. **Multiple DaemonSets / VPAs per node pool** — operational overhead; poor
   fit for dynamic GPU node provisioning.
2. **Only support `kubernetes.io/hostname`** — solves per-node variance but
   forces maximum cardinality even when a 2-way GPU split is enough (Falco).
3. **Fixed enum scopes (`Node`, `NodePool`)** — fights cloud-specific and
   custom labels; rejected in favor of arbitrary node label keys.
4. **Clearing `status.recommendation` for scoped VPAs** — rejected; keeping the
   global recommendation gives clients a safe fallback and makes disabling the
   gate non-destructive.
5. **Shadow VPA objects per scope value in ClusterState** — more churn; the
   synthetic-label approach reuses existing aggregation machinery.
6. **Cross-group initial recommendation** — deferred; risk of unschedulable
   pods on smaller nodes.
