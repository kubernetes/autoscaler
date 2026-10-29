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
  - [Example](#example)
  - [Design Decisions](#design-decisions)
    - [Disruption budget for selector VPAs](#disruption-budget-for-selector-vpas)
    - [How the admission controller determines the target node](#how-the-admission-controller-determines-the-target-node)
    - [Why startup boost is excluded](#why-startup-boost-is-excluded)
  - [Test Plan](#test-plan)
  - [Feature Enablement and Rollback](#feature-enablement-and-rollback)
  - [Graduation Criteria](#graduation-criteria)
  - [Version Skew](#version-skew)
  - [Kubernetes Version Compatibility](#kubernetes-version-compatibility)
- [Implementation History](#implementation-history)
- [Alternatives](#alternatives)
  - [One DaemonSet and VPA per node pool](#one-daemonset-and-vpa-per-node-pool)
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
add a pod label selector later (ref #8956). A VPA with a selector is still a regular VPA: it has
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
type VerticalPodAutoscalerSelector struct {
    // NodeSelector selects pods by the labels of the node they run on.
    // A pod is selected if its node has all of the given labels with the given
    // values (same semantics as pod.spec.nodeSelector).
    // In this iteration it is only allowed when TargetRef is a DaemonSet.
    // +optional
    NodeSelector map[string]string `json:"nodeSelector,omitempty"`
}
```

`VerticalPodAutoscalerSelector` is a struct rather than a bare map so that a pod label selector
(for example `podSelector *metav1.LabelSelector`) can be added later without changing the shape of
the field.

Validation (admission controller webhook and CRD validation):
- At least one selector field must be set when `selector` is present.
- `nodeSelector` keys and values must be valid label keys and values.
- `nodeSelector` is only allowed when `targetRef.kind` is `DaemonSet`.
- `selector` is only allowed when the `VPASelector` feature gate is enabled.
- `selector` cannot be combined with CPU startup boost: `spec.startupBoost` and
  `resourcePolicy.containerPolicies[].startupBoost` must be unset when `selector` is set. See
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
   DaemonSet means all its pods share one aggregate. For DaemonSet pods, the aggregate state key
   is extended with the pod's node name, so usage from different nodes is kept in separate
   aggregates. Because a DaemonSet runs at most one pod per node, this yields at most one
   aggregate per node per container.

3. Matching aggregates to VPAs - When the recommender decides whether an aggregate belongs to
   a VPA (`matchesAggregation`), it checks the pod label selector derived from `targetRef` (as
   today) and also checks `selector.nodeSelector` against the current labels of the aggregate's
   node. A VPA without a selector matches every aggregate of its target, as today. When a node's
   labels change, the recommender re-checks the aggregates of that node, links them to the
   selector VPAs that now match the node and unlinks them from the selector VPAs that no longer
   do.

4. Aggregates linked to several VPAs - As today, `UseAggregationIfMatching` links an aggregate to
   every VPA that matches it, so a node aggregate is typically linked to one selector VPA and to
   the VPA without a selector, if there is one. Each VPA computes its recommendation from all the
   aggregates linked to it, so the VPA without a selector keeps recommending for the whole
   DaemonSet. The aggregate's per-VPA fields (`UpdateMode`, `ScalingMode`, `ControlledResources`)
   are today overwritten by whichever VPA linked last. With the gate enabled they are instead
   taken from the controlling VPA, resolved with the rules in
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

When the `VPASelector` feature gate is enabled, the updater:

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

If the admission controller cannot determine the target node (e.g. the node affinity field is
absent or malformed), VPAs with a node selector are treated as not matching the pod. If no other
VPA matches, the pod is admitted with its original resource requests, and a warning is logged.

Validation - The admission controller enforces the validation rules listed under
[API Changes](#api-changes). It needs the `VPASelector` feature gate to accept VPAs with
`spec.selector`.

### Choosing the controlling VPA for a pod

Today, when more than one VPA matches a pod, `GetControllingVPAForPod` picks the oldest one (by
creation timestamp, then by name). With selectors, several VPAs targeting the same DaemonSet is an
expected setup, so the rule is extended:

1. A VPA matches a pod if its `targetRef` resolves to the pod's controller (as today) and, if
   `selector` is set, the pod's node labels match `selector.nodeSelector`.
2. Among the matching VPAs, a VPA with a selector takes precedence over a VPA without one. This
   lets users keep a plain VPA as a fallback for pods on nodes that no selector VPA covers.
3. If several VPAs remain, the existing tie-break applies (oldest first, then by name). This can
   happen with hand-written selectors that overlap. The recommender reports this with a
   condition on the VPAs that lose the tie-break.

The recommender, updater, and admission controller all use this same function, so they always
agree on which VPA controls a pod.

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

For a VPA with `spec.selector`, `GetCreatorMaps` therefore sets `configured` to the number of live
selected pods, as it already does for Jobs. Tolerance and `minReplicas` then apply per selector
VPA, which is the right unit: each DaemonSet pod serves its own node, so pods under other VPAs add
no availability to this one. Counting live pods instead of a desired count slightly loosens the
budget during rollouts; computing a per-selector desired count would mean re-running DaemonSet
scheduling against the matching nodes, which is not worth it for alpha.

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

### Test Plan

- **Unit tests**:
  - VPA matching with `selector.nodeSelector`, including precedence of selector VPAs over a VPA
    without a selector, the tie-break for overlapping selectors, and that an Off selector VPA
    wins over an active fallback VPA so the pod is not updated.
  - Recommender cluster state: a node aggregate is linked to every VPA that matches it, so a VPA
    without a selector still receives all aggregates of its target when selector VPAs exist;
    the aggregate's `UpdateMode`, `ScalingMode` and `ControlledResources` follow the controlling
    VPA and are updated when the controlling VPA changes.
  - Recommender cluster state: node-aware aggregate keys for DaemonSet pods, matching aggregates
    to VPAs by node labels, and re-matching when node labels change.
  - Updater, with a DaemonSet split across several selector VPAs: each pod is matched to the
    VPA whose `nodeSelector` matches its node's labels, a pod on an uncovered node goes to the
    fallback VPA or is skipped, and a pod is dropped from a VPA's list when its controlling VPA
    is a different object (including an Off one). For the disruption budget, `configured`, the
    eviction tolerance, and the `minReplicas` check are computed from the selected pods only, so
    a selector VPA with a minority of the DaemonSet's pods can still evict and resize, a one-node
    selector VPA with `minReplicas: 1` is updated, and disruptions under one VPA do not consume
    another VPA's budget.
  - Admission controller: extracting the node name from DaemonSet node affinity `matchFields`,
    and falling back correctly when the node cannot be determined (absent `matchFields`, empty
    `values`, malformed node affinity, etc.).
  - Validation rules for `spec.selector` (DaemonSet-only, at least one selector field, valid
    label keys and values, feature gate, rejection of startup boost together with a selector).
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

- Feature gate name: `VPASelector`
- Components depending on the feature gate:
    - admission-controller
    - recommender
    - updater

Enabling of feature gate `VPASelector` will cause the following to happen:
- admission-controller will:
    - Accept new VPA objects created with `spec.selector` configured
    - Resolve the controlling VPA for a new DaemonSet pod by the labels of its target node and
      inject that VPA's recommendation
- recommender will:
    - Keep per-node aggregates for DaemonSet pods and match them to VPAs by node labels
    - Compute a selector VPA's recommendation only from the pods on nodes its selector matches
- updater will:
    - Apply a selector VPA's recommendation only to the pods on nodes its selector matches
    - Compute the disruption budget of a selector VPA from the selected pods only

Disabling of feature gate `VPASelector` will cause the following to happen:
- admission-controller will:
    - Reject new VPA objects being created with `spec.selector` configured
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
  - E2E tests are stable for at least 2 releases.
  - No open bugs against the `VPASelector` feature gate.
  - Positive user feedback from at least two production deployments.
- **Beta → GA**:
  - Feature has been stable in beta for at least 2 releases.
  - Consider expanding support beyond DaemonSet and adding pod label selection.

### Version Skew

All three components (recommender, updater, admission controller) need the `VPASelector` gate,
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
  and values), but the DaemonSet-only and startup-boost rules are only enforced once the admission
  controller is upgraded.

### Kubernetes Version Compatibility

When using `updateMode: Off`, `Recreate`, or `InPlaceOrRecreate` (falling back to eviction),
this feature works on any Kubernetes version; no in-place update support is required.

When using `updateMode: InPlace` or `InPlaceOrRecreate` (for the in-place path), this feature
requires Kubernetes 1.27+ with the `InPlacePodVerticalScaling` feature gate enabled.
On Kubernetes versions without in-place pod updates, the `InPlace` update
mode will not function, and the admission controller rejects VPAs that combine `selector` with
`InPlace` mode when the `InPlace` feature gate is not enabled, exactly as it does for VPAs
without a selector.

## Implementation History

- 2026-09-01: Initial version, as part of AEP-7942 together with `VerticalPodAutoscalerSet`.
- 2026-10-25: Split out of AEP-7942 into its own AEP. `VerticalPodAutoscalerSet` is deferred to a
  follow-up AEP that builds on `spec.selector`.

## Alternatives

### One DaemonSet and VPA per node pool

Operators can manually create separate DaemonSets (with node selectors) and corresponding VPAs
for each node pool. This works but is operationally heavy: it requires updating DaemonSet
manifests every time node pools are added or removed, risks configuration drift, and scales
poorly in clusters with many node pools. It also multiplies the DaemonSet objects, whereas
`spec.selector` leaves the workload untouched and only splits the autoscaling.
