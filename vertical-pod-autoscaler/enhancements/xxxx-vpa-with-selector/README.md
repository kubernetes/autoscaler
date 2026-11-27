# AEP-XXXX: Node-Scoped VPAs via `spec.selector`

<!-- toc -->
- [Summary](#summary)
- [Motivation](#motivation)
  - [Goals](#goals)
  - [Non-Goals](#non-goals)
- [Proposal](#proposal)
- [Design Details](#design-details)
  - [API Changes](#api-changes)
  - [Component Changes](#component-changes)
    - [Recommender](#recommender)
    - [Updater](#updater)
    - [Admission Controller](#admission-controller)
  - [Choosing the controlling VPA for a pod](#choosing-the-controlling-vpa-for-a-pod)
    - [Overlapping selectors](#overlapping-selectors)
  - [Example](#example)
  - [Design Decisions](#design-decisions)
    - [Disruption budget for selector VPAs](#disruption-budget-for-selector-vpas)
    - [How the admission controller determines the target node](#how-the-admission-controller-determines-the-target-node)
    - [Why startup boost is excluded](#why-startup-boost-is-excluded)
    - [Per-node aggregates only for DaemonSets with a selector VPA](#per-node-aggregates-only-for-daemonsets-with-a-selector-vpa)
  - [Test Plan](#test-plan)
  - [Feature Enablement and Rollback](#feature-enablement-and-rollback)
  - [Graduation Criteria](#graduation-criteria)
  - [Version Skew](#version-skew)
  - [Kubernetes Version Compatibility](#kubernetes-version-compatibility)
- [Implementation History](#implementation-history)
- [Alternatives](#alternatives)
  - [One DaemonSet and VPA per node pool](#one-daemonset-and-vpa-per-node-pool)
  - [<code>nodeSelector</code> as a plain label map](#nodeselector-as-a-plain-label-map)
<!-- /toc -->

## Summary

VPA today computes a single recommendation per VPA object and applies it uniformly to all matching
pods. This works well for interchangeable replicas (e.g. Deployments) but breaks down for
DaemonSets running across heterogeneous nodes. A security agent on a GPU node may consume far more
CPU and memory than the same agent on a CPU-only node, yet VPA produces one recommendation for
both. The result is either over-provisioning on quiet nodes or under-provisioning (throttling, OOM
kills) on busy ones. The same problem appears in active/passive deployments: the active replica
serves all traffic while the passive standbys sit idle, yet all replicas receive one
recommendation, so the standbys are over-provisioned or the active replica is starved.

This AEP adds an optional `spec.selector` field to `VerticalPodAutoscaler`. It narrows the set of
pods, among those controlled by the VPA's `targetRef`, that the VPA applies to. In this first
iteration the selector supports node labels only (`selector.nodeSelector`). The API leaves room to
add a pod label selector later (ref [#8956](https://github.com/kubernetes/autoscaler/pull/8956)). A VPA with a selector is still a regular VPA: it has
its own `status.recommendation`, its own checkpoints, and is applied by the updater and the
admission controller only to the pods it selects. Several VPAs can therefore target the same
DaemonSet, each covering a different set of nodes.

## Motivation

DaemonSets are the primary workload type that runs one pod per node. In clusters with heterogeneous
node pools (different instance types, GPU vs CPU, different kernel configurations), the resource
consumption of a DaemonSet pod varies significantly based on the node it runs on. Examples include:

- **Security agents** (e.g. Falco, Tetragon) that consume more resources on nodes with higher
  workload density or GPU workloads.
- **Log shippers** (e.g. Fluent Bit) that process more data on nodes running data-intensive
  applications.
- **Monitoring exporters** (e.g. node-exporter with GPU metrics) that have higher overhead on
  specialized hardware or on bigger nodes with more traffic.
- **CNI agents** that scale with the number of pods or network throughput on a node.

A related case is **active/passive deployments**, where one replica holds a leader lease or
receives all traffic while the others are warm standbys. The active replica's usage is far higher
than the standbys', but a single VPA cannot tell them apart. Replicas in this setup are
distinguished by pod labels (e.g. a `role: active` label maintained by the application or a
leader-election sidecar) rather than by node, so this case is served by the pod label selector
the API reserves room for (ref #8956), not by the node selector implemented in this iteration.
It is listed here because it shapes the `selector` API: the field is a struct that can hold more
than one kind of selector, so pod-label selection can be added without a second API change.

Without per-subset recommendations, cluster operators must choose between:
1. A single VPA that over-provisions on quiet nodes (wasting resources).
2. A single VPA that under-provisions on busy nodes (causing OOMs or throttling).
3. Manually creating separate DaemonSets and VPAs per node pool (operationally heavy, error-prone,
   and drifts as node pools are added or removed).

### Goals

- Allow a VPA to apply to a subset of its target's pods, selected by the labels of the node each
  pod runs on.
- Allow several VPAs to target the same DaemonSet, each covering a disjoint set of nodes, with a
  deterministic rule for which VPA controls a given pod.
- Keep a VPA without a selector working as it does today, so that it still recommends for all
  pods of its target and acts as a fallback for pods that no selector VPA controls.
- Reuse the existing VPA recommendation, status, and checkpoint machinery unchanged for selector
  VPAs.
- Leave room in the API for pod-label-based selection in the future.

### Non-Goals

- A `VerticalPodAutoscalerSet` (or any other) resource that automatically creates one VPA per
  node label value. This is the subject of a follow-up AEP that builds on `spec.selector`.
- Support for workload types other than DaemonSet in this iteration. The API is shaped so it can
  be extended later, but validation restricts `selector.nodeSelector` to DaemonSet targets.
- Pod label selectors on VPA. The API reserves space for them, but they are not implemented here.
- Changing the format of `VerticalPodAutoscalerCheckpoint`.
- CPU startup boost for pods controlled by a VPA with a selector (not supported in this phase).
- Automatic node pool discovery or integration with cloud provider node group APIs.

## Proposal

A VPA with `spec.selector.nodeSelector` set only applies to pods of its `targetRef` that run on
nodes whose labels match the selector. Everything else about the VPA stays as it is today:

- The recommender computes the VPA's recommendation from the usage of only the selected pods
  and writes it to the VPA's own `status.recommendation`. Checkpoints are written as today.
- The updater applies the VPA's recommendation only to the selected pods, according to the
  configured `updateMode`.
- The admission controller works out which node a new DaemonSet pod is bound to (see
  [How the admission controller determines the target node](#how-the-admission-controller-determines-the-target-node)),
  picks the VPA whose selector matches that node, and injects its recommendation.

Several VPAs can therefore target the same DaemonSet, each covering a different set of nodes. A
VPA without a selector on the same DaemonSet is not affected by the selector VPAs when it comes
to its recommendation: it is still computed from all pods of the target, as today. It only loses
control of the pods that a selector VPA controls, and so acts as a fallback for the pods that no
selector VPA covers. Creating or deleting a selector VPA therefore never changes another VPA's
`status.recommendation`, only which pods each VPA acts on.

## Design Details

### API Changes

```go
// +kubebuilder:validation:XValidation:rule="!has(self.selector) || (!has(self.startupBoost) && (!has(self.resourcePolicy) || !has(self.resourcePolicy.containerPolicies) || self.resourcePolicy.containerPolicies.all(p, !has(p.startupBoost))))",message="startupBoost cannot be set when selector is set"
type VerticalPodAutoscalerSpec struct {
    // ... existing fields ...

    // Selector further restricts the set of pods, among those controlled by
    // TargetRef, that this VPA applies to. When unset, the VPA applies to all
    // pods controlled by TargetRef (current behavior).
    //
    // Recommendations are computed only from the selected pods, and only the
    // selected pods are updated by the updater and admission controller.
    // +optional
    Selector *VerticalPodAutoscalerSelector `json:"selector,omitempty"`
}

// VerticalPodAutoscalerSelector selects a subset of the pods controlled by
// the VPA's TargetRef. Pods must match all of the specified selectors.
// +kubebuilder:validation:MinProperties=1
type VerticalPodAutoscalerSelector struct {
    // NodeSelector selects pods by the labels of the node they run on.
    // A pod is selected if its node's labels match the selector.
    // In this iteration it is only allowed when TargetRef is a DaemonSet.
    // +kubebuilder:validation:MinProperties=1
    // +optional
    NodeSelector *metav1.LabelSelector `json:"nodeSelector,omitempty"`
}
```

`VerticalPodAutoscalerSelector` is a struct so that a pod label selector (for example
`podSelector *metav1.LabelSelector`) can be added later without changing the shape of the field.

A new condition type is added, set by the recommender (see
[Overlapping selectors](#overlapping-selectors)):

```go
var (
    // ... existing condition types ...

    // SelectorOverlap indicates that this VPA matched some nodes but lost them
    // to another VPA through the tie-break (oldest first, then by name), rather
    // than through a precedence rule.
    SelectorOverlap VerticalPodAutoscalerConditionType = "SelectorOverlap"
)
```

Validation (admission controller webhook and CRD validation):
- At least one selector field must be set and non-empty when `selector` is present; an empty
  `selector: {}` is rejected.
- `nodeSelector`, when set, must have at least one `matchLabels` entry or one `matchExpressions`
  entry. An empty `nodeSelector` would match every node while still taking precedence over a VPA
  without a selector (see
  [Choosing the controlling VPA for a pod](#choosing-the-controlling-vpa-for-a-pod)), so it is
  rejected.
- `nodeSelector` must be a valid label selector (checked with `ValidateLabelSelector`).
- `nodeSelector` is only allowed when `targetRef.kind` is `DaemonSet`.
- `nodeSelector` is only allowed when the `VPANodeSelector` feature gate is enabled.
- `selector` cannot be combined with CPU startup boost: `spec.startupBoost` and
  `resourcePolicy.containerPolicies[].startupBoost` must be unset when `selector` is set. This
  is enforced both by the admission controller and by a CEL rule on
  `VerticalPodAutoscalerSpec`, shown above. See
  [Why startup boost is excluded](#why-startup-boost-is-excluded).

### Component Changes

#### Recommender

When the `VPANodeSelector` feature gate is enabled:

1. Load node labels - The recommender has no node lister or informer today, so one is added.
   It stores node labels in the cluster state model, so
   the recommender can tell which node labels apply to a pod. The first version loads all
   nodes; loading only nodes that run pods of DaemonSets targeted by a VPA with a selector can
   be added later as an optimization.

2. Node-aware aggregation - Today pods with identical labels share one aggregate container
   state (`aggregateStateKey` is namespace, container name and pod label set), which for a
   DaemonSet means all its pods share one aggregate. For pods of a DaemonSet that is the target
   of at least one VPA with `spec.selector`, the aggregate state key is extended with the pod's
   node name (`pod.spec.nodeName`), so usage from different nodes is kept in separate
   aggregates. Because a DaemonSet runs at most one pod per node, this yields at most one
   aggregate per node per container. Every sample is still added to exactly one aggregate; the
   per-node aggregates replace the shared one, they do not exist next to it. All other pods,
   including pods of DaemonSets targeted only by VPAs without a selector, keep today's key, so
   enabling the gate adds no aggregates for clusters that do not use selectors.

   The node name is stable for the lifetime of the node and does not depend on which selector
   VPAs exist, so the aggregates of a DaemonSet never need to be rebuilt when a selector VPA is
   added, changed or deleted, or when a node is relabeled. Which VPA an aggregate belongs to is
   decided at match time (step 3) against the node's current labels.

   Transition - When the first selector VPA for a DaemonSet is created, new samples of its pods
   are added under the node-aware key from the next recommender cycle on. The existing shared
   aggregate receives no more samples. It stays linked to the VPA without a selector, if there
   is one, until it expires at the end of the aggregation window, so that VPA keeps its history.
   The per-node aggregates start empty, because the shared histogram cannot be split by node,
   so a new selector VPA has no history on its first cycles. This is a cold start similar to a
   new VPA on a workload the recommender has not seen before, see
   [Per-node aggregates only for DaemonSets with a selector VPA](#per-node-aggregates-only-for-daemonsets-with-a-selector-vpa).
   When the last selector VPA for a DaemonSet is
   deleted, new samples go back to the shared key and the per-node aggregates stay linked to
   the VPA without a selector until they expire, so no history is lost in that direction.

3. Matching aggregates to VPAs - When the recommender decides whether an aggregate belongs to
   a VPA (`matchesAggregation`), it checks the pod label selector derived from `targetRef` (as
   today) and also checks `selector.nodeSelector` against the current labels of the aggregate's
   node. A VPA without a selector matches every aggregate of its target, as today. If a node's
   labels change, the recommender re-checks the aggregates of that node, links them to the
   selector VPAs that now match the node and unlinks them from the selector VPAs that no longer
   do. The aggregate itself is untouched, so a relabeled node changes owner without losing its
   history. This is the same shape as changing `targetRef` on a running VPA today: the VPA is
   relinked to the aggregates that match its new target, and the aggregates stay as they are.

4. Computing a recommendation from several aggregates - As today, `UseAggregationIfMatching`
   links an aggregate to every VPA that matches it, so a node aggregate is typically linked to
   one selector VPA and to the VPA without a selector, if there is one. Each recommender cycle,
   each VPA merges all the aggregates linked to it into one histogram per container
   (`AggregateStateByContainerName`, as today) and computes one recommendation from the
   merged result. A VPA selecting `pool=gpu` over three gpu nodes therefore merges three
   aggregates into one recommendation, and the VPA without a selector merges all the aggregates
   of the DaemonSet, which gives the same result as the single shared aggregate would have.
   Merging is the existing mechanism the recommender uses whenever a VPA's pods do not share one
   label set (for example across a Deployment rollout), and is a sum over fixed-size histograms,
   so the number of aggregates does not change the recommendation, only the memory held by the
   recommender. The recommendation is written to the VPA's status and applied to pods by the
   updater and admission controller; pods never see individual aggregates.

   The aggregate's per-VPA fields (`UpdateMode`, `ScalingMode`, `ControlledResources`) are today
   overwritten by whichever VPA linked last. With the gate enabled they are instead taken from
   the controlling VPA, resolved with the rules in
   [Choosing the controlling VPA for a pod](#choosing-the-controlling-vpa-for-a-pod).

Pods on nodes that no selector VPA selects - If a DaemonSet pod runs on a node that matches none of
the selector VPAs targeting the DaemonSet, that pod does not contribute to any of those VPAs. If
there is also a VPA without a selector targeting the same DaemonSet, that VPA controls the pod (see
[Choosing the controlling VPA for a pod](#choosing-the-controlling-vpa-for-a-pod)). Otherwise the
pod is not managed by VPA and keeps its existing resource requests.

Checkpoints - A selector VPA writes checkpoints named `{vpa-name}-{container-name}` exactly as
today. Nothing in the checkpoint format changes. If `spec.selector` is later removed from a VPA,
its checkpoints remain valid and are reused.

#### Updater

When the `VPANodeSelector` feature gate is enabled, the updater:

1. Keeps a node lister so it can look up the labels of `pod.spec.nodeName` for running pods.
2. Uses the extended matching in
   [Choosing the controlling VPA for a pod](#choosing-the-controlling-vpa-for-a-pod), so each
   pod is matched to the VPA whose selector matches its node.
3. Applies that VPA's recommendation according to its `updateMode`. All existing update modes
   are supported unchanged.
4. Scopes the disruption budget (`minReplicas` and the eviction tolerance) to the pods the VPA
   selects rather than to the whole DaemonSet. See
   [Disruption budget for selector VPAs](#disruption-budget-for-selector-vpas).

#### Admission Controller

Mutation - Although `pod.spec.nodeName` is empty at admission time, the DaemonSet controller
binds each pod to a specific node via node affinity: it sets
`spec.affinity.nodeAffinity.requiredDuringSchedulingIgnoredDuringExecution.nodeSelectorTerms[].matchFields`
with `key: metadata.name` and the target node's name as the value. The admission controller
reads the node name from this field, looks up the node's labels, picks the controlling VPA (see
[Choosing the controlling VPA for a pod](#choosing-the-controlling-vpa-for-a-pod)), and injects
that VPA's recommendation into the pod's resource requests. This makes eviction-based modes
(`Recreate`, `InPlaceOrRecreate`) work as well, because the replacement pod is admitted with the
correct resources from the start.

The admission controller may be unable to look up the target node's labels, either because the
node affinity field is absent or malformed, or because the node is not yet in its node cache.

If a VPA with a selector targets the pod's controller, the admission controller then leaves the
pod unchanged and logs a warning. It does not fall back to a VPA without a selector, because that
could apply the wrong recommendation, or update a pod that a selector VPA in `Off` mode should
control. If no VPA with a selector targets the controller, node labels are not needed and the
controlling VPA is chosen as today.

The updater corrects the pod later. Once the node is in its cache, it finds the pod's controlling
VPA and applies that VPA's `updateMode`, so a recreated pod is admitted with the right
recommendation. Until then, the updater skips the pod in the same way.

Validation - The admission controller enforces the validation rules listed under
[API Changes](#api-changes). It needs the `VPANodeSelector` feature gate to accept VPAs with
`spec.selector.nodeSelector`.

### Choosing the controlling VPA for a pod

Today, when more than one VPA matches a pod, `GetControllingVPAForPod` picks the oldest one (by
creation timestamp, then by name). With selectors, several VPAs targeting the same DaemonSet is an
expected setup, so the rule is extended:

1. A VPA matches a pod if its `targetRef` resolves to the pod's controller (as today) and, if
   `selector` is set, the pod's node labels match `selector.nodeSelector`.
2. Among the matching VPAs, a VPA with a selector takes precedence over a VPA without one. This
   lets users keep a plain VPA as a fallback for pods on nodes that no selector VPA covers.
3. If several VPAs remain, the existing tie-break applies (oldest first, then by name). This can
   happen with hand-written selectors that overlap. The recommender reports it with the
   `SelectorOverlap` condition, see [Overlapping selectors](#overlapping-selectors).

The recommender, updater, and admission controller all use this same function, so they always
agree on which VPA controls a pod.

#### Overlapping selectors

Two selector VPAs overlap on a node when they target the same controller, both `nodeSelector`s
match that node's labels, and the node runs a pod of the controller. This is checked against the
nodes' current labels, so selectors with different keys (such as `pool=gpu` and `zone=a`) can
overlap.

For example, VPA `gpu` selects `pool=gpu` and
matches nodes 1-3, and VPA `zone-a`, created later, selects `zone=a` and matches nodes 3-5. Node 3
matches both, so the tie-break (oldest first) gives its pod to `gpu`, because it was created
first. `zone-a` still controls the pods on nodes
4 and 5.

The recommender warns about this with the `SelectorOverlap` condition (reason `LostTieBreak`) on
the losing VPA, here `zone-a`, with a message such as
`1 of 3 nodes is controlled by VPA default/gpu: node-3`. The condition is removed once the
overlap is gone. The losing VPA keeps its recommendation, which is applied only to the pods it
controls.

The condition is defined by how a VPA lost a node, not by what kind of VPA it is: it is set only
when the node went to another VPA through the step 3 tie-break. The tie-break is the one place
where the outcome depends on creation order rather than on the specs, so it is the one worth
flagging. A VPA that loses a node through step 2 precedence is not reported, because a plain VPA
yielding to a selector VPA is the intended fallback setup, and a condition that is always set on
a fallback VPA would carry no information. As a consequence the winning VPA never gets the
condition, and a VPA without a selector never gets it either, since it can only lose through
step 2.

`updateMode: Off` VPAs take part in the choice. Today the updater and the admission controller
skip Off VPAs before matching pods. With selectors that would let a fallback VPA in `Recreate`
mode hijack the pods of an Off partition, since the Off VPA is never considered and the fallback
is the only match left. Both components therefore resolve the controlling VPA over all VPAs
targeting the controller, including Off ones, and only then check its update mode: if the
winner is Off, the pod is left alone, even if a weaker VPA would have acted on it. The updater,
which today iterates per VPA, applies the same rule by dropping from a VPA's pod list any pod
whose controlling VPA is a different object.

### Example

Given a two-node Kind cluster (`kind-worker`, `kind-worker2`) and the following manifests, with
one hand-written VPA per node:

```yaml
apiVersion: apps/v1
kind: DaemonSet
metadata:
  name: test-daemonset
  labels:
    app: test-daemonset
spec:
  selector:
    matchLabels:
      app: test-daemonset
  template:
    metadata:
      labels:
        app: test-daemonset
    spec:
      containers:
      - name: stress
        image: polinux/stress
        command: ["stress"]
        args: ["--vm", "1", "--vm-bytes", "50M", "--vm-hang", "0"]
        resources:
          requests:
            cpu: "10m"
            memory: "100Mi"
          limits:
            cpu: "500m"
            memory: "500Mi"
---
apiVersion: autoscaling.k8s.io/v1
kind: VerticalPodAutoscaler
metadata:
  name: test-daemonset-kind-worker
spec:
  targetRef:
    apiVersion: apps/v1
    kind: DaemonSet
    name: test-daemonset
  updatePolicy:
    updateMode: "Off"
  selector:
    nodeSelector:
      matchLabels:
        kubernetes.io/hostname: kind-worker
---
apiVersion: autoscaling.k8s.io/v1
kind: VerticalPodAutoscaler
metadata:
  name: test-daemonset-kind-worker2
spec:
  targetRef:
    apiVersion: apps/v1
    kind: DaemonSet
    name: test-daemonset
  updatePolicy:
    updateMode: "Off"
  selector:
    nodeSelector:
      matchLabels:
        kubernetes.io/hostname: kind-worker2
```

Each VPA is a regular VPA with a node selector, and its recommendation is computed from the usage
on that node only:

```yaml
apiVersion: autoscaling.k8s.io/v1
kind: VerticalPodAutoscaler
metadata:
  name: test-daemonset-kind-worker
  namespace: default
spec:
  targetRef:
    apiVersion: apps/v1
    kind: DaemonSet
    name: test-daemonset
  updatePolicy:
    updateMode: "Off"
  selector:
    nodeSelector:
      matchLabels:
        kubernetes.io/hostname: kind-worker
status:
  conditions:
  - lastTransitionTime: "2026-09-01T06:31:38Z"
    status: "True"
    type: RecommendationProvided
  recommendation:
    containerRecommendations:
    - containerName: stress
      lowerBound:
        cpu: 25m
        memory: "317014174"
      target:
        cpu: 25m
        memory: "323522422"
      uncappedTarget:
        cpu: 25m
        memory: "323522422"
      upperBound:
        cpu: 123m
        memory: "3627581199"
---
apiVersion: autoscaling.k8s.io/v1
kind: VerticalPodAutoscaler
metadata:
  name: test-daemonset-kind-worker2
  namespace: default
spec:
  targetRef:
    apiVersion: apps/v1
    kind: DaemonSet
    name: test-daemonset
  updatePolicy:
    updateMode: "Off"
  selector:
    nodeSelector:
      matchLabels:
        kubernetes.io/hostname: kind-worker2
status:
  conditions:
  - lastTransitionTime: "2026-09-01T06:31:38Z"
    status: "True"
    type: RecommendationProvided
  recommendation:
    containerRecommendations:
    - containerName: stress
      lowerBound:
        cpu: 25m
        memory: 250Mi
      target:
        cpu: 25m
        memory: 250Mi
      uncappedTarget:
        cpu: 25m
        memory: 250Mi
      upperBound:
        cpu: 123m
        memory: "877084945"
```

In this example, the pod on `kind-worker` was given an additional `stress --vm-bytes 200M`
workload via `kubectl exec` to simulate higher resource consumption, driving its memory usage
higher than the pod on `kind-worker2`. The recommendations reflect this: `323522422` bytes
(~308Mi) on `kind-worker` vs `250Mi` on `kind-worker2`.

### Design Decisions

#### Disruption budget for selector VPAs

The updater's `GetCreatorMaps` builds per-controller group statistics and, for a DaemonSet, sets
`configured` to `status.desiredNumberScheduled`. A pod is disruptable only if the group's running
count stays above `configured - configured * evictionToleranceFraction`, and the group is skipped
if it has fewer live pods than `minReplicas` (default 2).

For a selector VPA the group holds only the selected pods, so a DaemonSet-wide `configured` is
wrong: with the default tolerance of 0.5, a selector VPA holding half or fewer of the DaemonSet's
pods can never evict or resize, and a one-node selector (such as the hostname-selected
[example](#example) above) is always skipped by `minReplicas`.

The fallback VPA has the same problem. It controls only the pods on nodes that no selector VPA
covers, but its `configured` is still the whole DaemonSet, so it can be blocked too.

The fix: if at least one selector VPA targets a DaemonSet, every VPA on that DaemonSet (selector
VPAs and the fallback) sets `configured` to the number of pods it controls. Terminating pods are
counted too. Tolerance and `minReplicas` then apply per VPA. Each DaemonSet pod serves its own
node, so pods under other VPAs do not help this one.

Terminating pods must be counted because the DaemonSet controller creates the replacement pod only
after the old pod is gone. If only live pods were counted, `configured` would get smaller after
each round of evictions, and the updater could keep evicting more pods.

Known limitation (for alpha): after a pod is fully deleted, there are a few seconds before its replacement is
created. If an updater loop runs in that time, it can evict one more pod.

If no selector VPA targets the DaemonSet, nothing changes: `configured` is still
`status.desiredNumberScheduled`.

`minReplicas` keeps its semantics and default. A selector VPA that can cover a single node must
set `updatePolicy.minReplicas: 1` for eviction-based modes to work, as a single-node DaemonSet
with a plain VPA must today. `--in-place-skip-disruption-budget` is unaffected.

#### How the admission controller determines the target node

When a new DaemonSet pod is admitted, `pod.spec.nodeName` is empty: the scheduler has not yet
assigned the pod to a node. However, the DaemonSet controller binds each pod to a specific node
by setting a node affinity with `matchFields` on `metadata.name`:

```yaml
spec:
  affinity:
    nodeAffinity:
      requiredDuringSchedulingIgnoredDuringExecution:
        nodeSelectorTerms:
        - matchFields:
          - key: metadata.name
            operator: In
            values:
            - <node-name>
```

The admission controller reads the node name from this field, looks up the node's labels in its
cache, and picks the VPA whose selector matches. This makes all update modes work, including
eviction-based modes (`Recreate`, `InPlaceOrRecreate`), because the replacement pod is admitted
with the correct recommendation from the start, which avoids eviction loops.

This is also why `selector.nodeSelector` is limited to DaemonSets in this iteration: for other
workloads the node is not known at admission time.

#### Why startup boost is excluded

CPU startup boost is a non-goal for selector VPAs in this phase. Without the validation rule
forbidding the combination, the updater's unboost path would list pods by the `targetRef` label
selector alone and so size its in-place restrictions over the whole DaemonSet rather than the
selected pods. Rejecting the combination up front is simpler than teaching the unboost path about
selectors, and can be revisited once the selector plumbing is stable.

#### Per-node aggregates only for DaemonSets with a selector VPA

The recommender switches a DaemonSet to per-node aggregates only once a VPA with `spec.selector`
targets it. An alternative is to key every DaemonSet pod by node name whenever the
`VPANodeSelector` feature gate is enabled. That would avoid the transition described in the
[Recommender](#recommender) section and the cold start that comes with it: the per-node
aggregates would already hold history when the first selector VPA is created.

The alternative is rejected because it makes every user pay for the feature whether they use it
or not. The recommender tracks all pods by default (`--memory-saver` is off), so every DaemonSet
in the cluster would hold one aggregate per node per container, including DaemonSets nobody
intends to split. Aggregates of nodes that have left the cluster stay in memory until the end of
the aggregation window, so on clusters with node churn the count is a multiple of the live node
count. Once the gate is locked on at GA there would be no way to turn this off. A feature gate should enable a capability, not change the recommender's memory profile
for workloads that do not use it. Keeping the per-node key behind the presence of a selector VPA
preserves the property that enabling the gate adds no aggregates for clusters that do not use
selectors.

The cost of this choice is the cold start of the first selector VPA on a DaemonSet, which is
accepted for alpha. It is the same situation as a new VPA on a workload whose pods the
recommender has not seen, and the recommendation converges as samples arrive.

### Test Plan

- **Unit tests**:
  - VPA matching with `selector.nodeSelector`, including precedence of selector VPAs over a VPA
    without a selector, the tie-break for overlapping selectors, and that an Off selector VPA
    wins over an active fallback VPA so the pod is not updated.
  - `SelectorOverlap` condition: set only on a VPA that lost a node through the step 3
    tie-break, for partial overlap (the loser still controls its other nodes) and full overlap
    (the loser controls no pods), never set on the winner, never set on a VPA that lost through
    step 2 precedence (so never on a VPA without a selector), ignoring overlap on nodes that run
    no pod of the target, and removed when node labels change or the winning VPA is deleted.
  - Recommender cluster state: a node aggregate is linked to every VPA that matches it, so a VPA
    without a selector still receives all aggregates of its target when selector VPAs exist;
    the aggregate's `UpdateMode`, `ScalingMode` and `ControlledResources` follow the controlling
    VPA and are updated when the controlling VPA changes.
  - Recommender cluster state: node-aware aggregate keys only for pods of DaemonSets targeted by
    a selector VPA (pods of other DaemonSets keep the shared key), re-keying when the first
    selector VPA for a DaemonSet is created and when the last one is deleted, aggregates without
    a node name never matching a selector VPA, matching aggregates to VPAs by node labels, and
    re-matching when node labels change.
  - Updater, with a DaemonSet split across several selector VPAs: each pod is matched to the
    VPA whose `nodeSelector` matches its node's labels, a pod on an uncovered node goes to the
    fallback VPA or is skipped, and a pod is dropped from a VPA's list when its controlling VPA
    is a different object (including an Off one). The disruption budget is computed per VPA from
    the pods it controls, including the fallback VPA and terminating pods.
  - Admission controller: extracting the node name from DaemonSet node affinity `matchFields`.
    When the node labels cannot be resolved (absent `matchFields`, empty `values`, malformed node
    affinity, or a node name missing from the node cache), the pod is admitted unmutated if a
    selector VPA targets its controller, including when a VPA without a selector also targets it,
    and the controlling VPA is chosen as today otherwise.
  - Updater: a pod whose node is missing from the node cache is skipped while selector VPAs
    target its controller, and is updated once the node appears.
  - Validation rules for `spec.selector` (DaemonSet-only, at least one non-empty selector field,
    rejection of `selector: {}` and `nodeSelector: {}`, invalid label selectors, feature gate,
    rejection of startup boost together with a selector), in both the admission controller and
    the CRD CEL rule.
- **E2E tests**:
  - Several selector VPAs on a heterogeneous cluster each receive an independent recommendation
    computed only from the pods on their nodes.
  - In-place updates apply the correct VPA's recommendation to each pod.
  - Eviction-based updates (`Recreate`) evict pods, and the admission controller injects the
    correct recommendation into the replacement pod. The test uses at least three selector VPAs,
    so that each holds a minority of the DaemonSet's pods and a DaemonSet-wide budget would have
    blocked eviction.
  - Pods on nodes that no selector VPA covers fall back to a VPA without a selector, if one
    exists, and are otherwise left untouched. The fallback VPA's recommendation is computed from
    all pods of the DaemonSet and does not change when a selector VPA is created or deleted.
  - Pods under a selector VPA that is `Off` are not evicted or resized by a fallback VPA in
    `Recreate` mode targeting the same DaemonSet.

### Feature Enablement and Rollback

- Feature gate name: `VPANodeSelector`
- Components depending on the feature gate:
    - admission-controller
    - recommender
    - updater

Enabling of feature gate `VPANodeSelector` will cause the following to happen:
- admission-controller will:
    - Accept new VPA objects created with `spec.selector.nodeSelector` configured
    - Resolve the controlling VPA for a new DaemonSet pod by the labels of its target node and
      inject that VPA's recommendation
- recommender will:
    - Keep per-node aggregates for pods of DaemonSets targeted by at least one selector VPA, and
      match them to VPAs by node labels. Pods of other workloads keep today's aggregate keys.
    - Compute a selector VPA's recommendation only from the pods on nodes its selector matches
- updater will:
    - Apply a selector VPA's recommendation only to the pods on nodes its selector matches
    - Compute the disruption budget of each VPA on a DaemonSet with selector VPAs (including the
      fallback VPA) from the pods it controls only

Disabling of feature gate `VPANodeSelector` will cause the following to happen:
- admission-controller will:
    - Reject new VPA objects being created with `spec.selector.nodeSelector` configured
        - A descriptive error message should be returned to the user letting them know that they
          are using a feature gated feature
    - Leave existing VPA objects with `spec.selector` in place
    - Ignore `spec.selector` on existing VPAs when resolving the controlling VPA for a pod
- recommender will:
    - Ignore `spec.selector` and compute every VPA's recommendation from all pods of its target,
      as for a VPA without a selector
- updater will:
    - Ignore `spec.selector` and match every VPA to all pods of its target
    - Fall back to the existing oldest-wins rule when several VPAs target the same DaemonSet, so
      one of them controls all pods

Operators rolling back should therefore delete all but one of the VPAs targeting a DaemonSet and,
if needed, remove `spec.selector` from the remaining one.

Removing `spec.selector` from a VPA makes it cover all pods of its target again. Its existing
checkpoints remain valid and are reused.

### Graduation Criteria

- **Alpha → Beta**:
  - E2E tests are stable for at least one release.
  - No open bugs against the `VPANodeSelector` feature gate.
  - Positive user feedback from at least two production deployments.
- **Beta → GA**:
  - Feature has been stable in beta for at least one release.
  - Consider expanding support beyond DaemonSet and adding pod label selection.

### Version Skew

All three components (recommender, updater, admission controller) need the `VPANodeSelector` gate,
and need to be on a version that understands `spec.selector`, for selector VPAs to work
correctly:

- **New recommender, old updater/admission controller**: The recommender computes per-selector
  recommendations, but the old updater and admission controller ignore `spec.selector`. Every
  selector VPA appears to match every pod of the DaemonSet, so the oldest VPA's recommendation is
  applied to all pods. The result is wrong per-node values, not a crash. Operators should upgrade
  the updater and admission controller first, or use `updateMode: Off` until all components are
  upgraded.
- **New updater/admission controller, old recommender**: The old recommender ignores the
  selector and computes the same recommendation, from all pods, for each selector VPA. The
  updater and admission controller apply it to the selected pods only. This is equivalent to
  having no per-node-group recommendations.
- **Old admission controller**: An admission controller that does not know `spec.selector`
  does not validate it. The CRD validation still applies the structural rules (valid label keys
  and values) and the CEL rule rejecting startup boost together with a selector, but the
  DaemonSet-only rule is only enforced once the admission controller is upgraded.

### Kubernetes Version Compatibility

NONE.

## Implementation History

- 2023-07-05: Original issue for per-node DaemonSet VPA recommendations opened in
  [#5928](https://github.com/kubernetes/autoscaler/issues/5928) by @jcogilvie.
- 2025-03-18: Initial AEP-7942 for DaemonSets with heterogeneous resource requirements proposed in
  [#7942](https://github.com/kubernetes/autoscaler/pull/7942) by @bernot-dev.
- 2025-12-18: AEP for `spec.selector` support on heterogeneous workloads proposed in
  [#8956](https://github.com/kubernetes/autoscaler/pull/8956) by @msudheendra-cflt.
- 2026-07-21: DaemonSet-scoped VPA AEP and prototype proposed in
  [#10012](https://github.com/kubernetes/autoscaler/pull/10012),
  [#10031](https://github.com/kubernetes/autoscaler/issues/10031), and
  [#10032](https://github.com/kubernetes/autoscaler/pull/10032) by @ergoz.
- 2026-09-01: Initial version, as part of AEP-7942 together with `VerticalPodAutoscalerSet`.
  The `spec.selector.nodeSelector` and `VerticalPodAutoscalerSet` design was proposed by
  @adrianmoisey in [#10229](https://github.com/kubernetes/autoscaler/pull/10229#issuecomment-6096510946).
- 2026-10-25: Split out of AEP-7942 into its own AEP. `VerticalPodAutoscalerSet` is deferred to a
  follow-up AEP that builds on `spec.selector`.

## Alternatives

### One DaemonSet and VPA per node pool

Operators can manually create separate DaemonSets (with node selectors) and corresponding VPAs
for each node pool. This works but is operationally heavy: it requires updating DaemonSet
manifests every time node pools are added or removed, risks configuration drift, and scales
poorly in clusters with many node pools. It also multiplies the DaemonSet objects, whereas
`spec.selector` leaves the workload untouched and only splits the autoscaling.

### `nodeSelector` as a plain label map

`nodeSelector` could be a `map[string]string`, like `pod.spec.nodeSelector`. It is simpler, but it
cannot express "all nodes except X" or "any of these pools". Changing the field's type later would
break the API, so `metav1.LabelSelector` is used from the start. The simple case is still simple
with `matchLabels`.
