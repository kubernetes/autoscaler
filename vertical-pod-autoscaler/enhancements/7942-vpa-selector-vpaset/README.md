# AEP-7942: Vertical Pod Autoscaling for DaemonSets with Heterogeneous Resource Requirements

<!-- toc -->
- [Summary](#summary)
- [Motivation](#motivation)
  - [Goals](#goals)
  - [Non-Goals](#non-goals)
- [Proposal](#proposal)
- [Design Details](#design-details)
  - [API Changes](#api-changes)
    - [New field on VerticalPodAutoscalerSpec: selector](#new-field-on-verticalpodautoscalerspec-selector)
    - [New CRD: VerticalPodAutoscalerSet](#new-crd-verticalpodautoscalerset)
  - [Component Changes](#component-changes)
    - [Recommender](#recommender)
    - [VPASet Controller](#vpaset-controller)
    - [Updater](#updater)
    - [Admission Controller](#admission-controller)
  - [Choosing the controlling VPA for a pod](#choosing-the-controlling-vpa-for-a-pod)
  - [VPASet Lifecycle Example](#vpaset-lifecycle-example)
    - [Example](#example)
  - [Design Decisions](#design-decisions)
    - [Disruption budget for selector VPAs](#disruption-budget-for-selector-vpas)
    - [How the admission controller determines the target node](#how-the-admission-controller-determines-the-target-node)
  - [Test Plan](#test-plan)
  - [Feature Enablement and Rollback](#feature-enablement-and-rollback)
  - [Graduation Criteria](#graduation-criteria)
  - [Version Skew](#version-skew)
  - [Kubernetes Version Compatibility](#kubernetes-version-compatibility)
- [Implementation History](#implementation-history)
- [Alternatives](#alternatives)
  - [One DaemonSet and VPA per node pool](#one-daemonset-and-vpa-per-node-pool)
  - [Embedding per-group recommendations in VPA status](#embedding-per-group-recommendations-in-vpa-status)
<!-- /toc -->

## Summary

VPA today computes a single recommendation per VPA object and applies it uniformly to all matching
pods. This works well for interchangeable replicas (e.g. Deployments) but breaks down for
DaemonSets running across heterogeneous nodes. A security agent on a GPU node may consume far more
CPU and memory than the same agent on a CPU-only node, yet VPA produces one recommendation for
both. The result is either over-provisioning on quiet nodes or under-provisioning (throttling, OOM
kills) on busy ones.

This AEP introduces two building blocks:

1. An optional `spec.selector` field on `VerticalPodAutoscaler`. It narrows the set of pods,
   among those controlled by the VPA's `targetRef`, that the VPA applies to. In this first
   iteration the selector supports node labels only (`selector.nodeSelector`). The API leaves
   room to add a pod label selector later (ref #8956). A VPA with a selector is still a regular VPA: it has
   its own `status.recommendation`, its own checkpoints, and is applied by the updater and the
   admission controller only to the pods it selects.

2. A new `VerticalPodAutoscalerSet` (VPASet) resource. Its only job is to watch a configured
   node label key and to keep one VPA per distinct value of that label: it creates a VPA, with
   `spec.selector.nodeSelector` set to that value, when the value appears and deletes it when
   the value disappears. The API leaves room to partition by pod labels in the future.

## Motivation

DaemonSets are the primary workload type that runs one pod per node. In clusters with heterogeneous
node pools (different instance types, GPU vs CPU, different kernel configurations), the resource
consumption of a DaemonSet pod varies significantly based on the node it runs on. Examples include:

- **Security agents** (e.g. Falco, Tetragon) that consume more resources on nodes with higher
  workload density or GPU workloads.
- **Log shippers** (e.g. Fluent Bit) that process more data on nodes running data-intensive
  applications.
- **Monitoring exporters** (e.g. node-exporter with GPU metrics) that have higher overhead on
  specialized hardware.
- **CNI agents** that scale with the number of pods or network throughput on a node.

Without per-node-group recommendations, cluster operators must choose between:
1. A single VPA that over-provisions on quiet nodes (wasting resources).
2. A single VPA that under-provisions on busy nodes (causing OOMs or throttling).
3. Manually creating separate DaemonSets and VPAs per node pool (operationally heavy, error-prone,
   and drifts as node pools are added or removed).

### Goals

- Allow a VPA to apply to a subset of its target's pods, selected by the labels of the node each
  pod runs on.
- Introduce a `VerticalPodAutoscalerSet` resource that automatically keeps one VPA per distinct
  value of a node label key, creating and deleting VPAs as label values appear and disappear.
- Reuse the existing VPA recommendation, status, and checkpoint machinery unchanged for the
  generated VPAs.
- Leave room in both APIs for pod-label-based selection and partitioning in the future.
- Gate both features behind alpha feature flags so they can be adopted incrementally.

### Non-Goals

- Support for workload types other than DaemonSet in this iteration. Both APIs are shaped so they
  can be extended later, but validation restricts `selector.nodeSelector` to DaemonSet targets.
- Pod label selectors on VPA, and partitioning a VPASet by pod labels. The APIs reserve space for
  these, but they are not implemented here.
- Partitioning by more than one label key at once.
- Changing the format of `VerticalPodAutoscalerCheckpoint`.
- CPU startup boost for pods controlled by a VPA with a selector (not supported in this phase).
- Automatic node pool discovery or integration with cloud provider node group APIs.

## Proposal

`spec.selector` on VPA. A VPA with `spec.selector.nodeSelector` set only applies to pods of
its `targetRef` that run on nodes whose labels match the selector. Everything else about the VPA
stays as it is today:

- The recommender computes the VPA's recommendation from the usage of only the selected pods
  and writes it to the VPA's own `status.recommendation`. Checkpoints are written as today.
- The updater applies the VPA's recommendation only to the selected pods, according to the
  configured `updateMode`.
- The admission controller works out which node a new DaemonSet pod is bound to (see
  [How the admission controller determines the target node](#how-the-admission-controller-determines-the-target-node)),
  picks the VPA whose selector matches that node, and injects its recommendation.

Several VPAs can therefore target the same DaemonSet, each covering a different set of nodes.

VPASet - Writing one VPA per node pool by hand works, but it drifts as node pools are added
and removed. A `VerticalPodAutoscalerSet` automates this. It holds a VPA template and a node
label key. The VPASet controller watches nodes and, for every distinct value of that key, makes
sure there is a VPA built from the template with `spec.selector.nodeSelector` set to
`{<key>: <value>}`. When no node carries a value any more, the corresponding VPA is deleted.
Generated VPAs are owned by the VPASet, so deleting the VPASet deletes them all.

## Design Details

### API Changes

#### New field on VerticalPodAutoscalerSpec: selector

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

Validation (admission controller webhook and CRD validation):
- At least one selector field must be set when `selector` is present.
- `nodeSelector` keys and values must be valid label keys and values.
- `nodeSelector` is only allowed when `targetRef.kind` is `DaemonSet`.
- `selector` is only allowed when the `VPASelector` feature gate is enabled.
- `selector` cannot be combined with CPU startup boost: `spec.startupBoost` and
  `resourcePolicy.containerPolicies[].startupBoost` must be unset when `selector` is set. CPU
  startup boost is a non-goal for selector VPAs (see [Non-Goals](#non-goals)), and without this
  rule the updater's unboost path would list pods by the `targetRef` label selector alone and so
  size its in-place restrictions over the whole DaemonSet rather than the selected pods.

#### New CRD: VerticalPodAutoscalerSet

`VerticalPodAutoscalerSet` is served at `autoscaling.k8s.io/v1alpha1`. The `ownerReference` on
every generated VPA uses this apiVersion, so the garbage collector can resolve the owner. The
generated VPAs themselves remain `autoscaling.k8s.io/v1` objects.

```go
type VerticalPodAutoscalerSet struct {
    metav1.TypeMeta   `json:",inline"`
    metav1.ObjectMeta `json:"metadata,omitempty"`
    Spec   VerticalPodAutoscalerSetSpec   `json:"spec"`
    Status VerticalPodAutoscalerSetStatus `json:"status,omitempty"`
}

type VerticalPodAutoscalerSetSpec struct {
    // Partition defines how the set is split into individual VPAs.
    Partition VerticalPodAutoscalerSetPartition `json:"partition"`

    // Template describes the VPAs that are created for each partition.
    // template.spec.selector must be unset; the controller fills it in.
    Template VerticalPodAutoscalerTemplateSpec `json:"template"`
}

// VerticalPodAutoscalerSetPartition defines the source of the values the set
// is partitioned by. Exactly one field must be set.
type VerticalPodAutoscalerSetPartition struct {
    // NodeLabelKey is a node label key. One VPA is created for each distinct
    // value of this label across the nodes in the cluster, with
    // spec.selector.nodeSelector set to {NodeLabelKey: <value>}.
    // Nodes without this label do not produce a VPA.
    // +optional
    NodeLabelKey *string `json:"nodeLabelKey,omitempty"`
}

type VerticalPodAutoscalerTemplateSpec struct {
    // Labels and annotations copied onto every generated VPA.
    // +optional
    metav1.ObjectMeta `json:"metadata,omitempty"`

    // Spec of the generated VPAs. Selector must be unset.
    Spec VerticalPodAutoscalerSpec `json:"spec"`
}

type VerticalPodAutoscalerSetStatus struct {
    // VPACount is the number of VPAs currently managed by this set.
    VPACount int32 `json:"vpaCount"`

    // Conditions describe the state of the set (e.g. Ready, ConflictingVPA).
    Conditions []metav1.Condition `json:"conditions,omitempty"`

    // ObservedGeneration is the most recent generation observed by the controller.
    ObservedGeneration int64 `json:"observedGeneration,omitempty"`
}
```

Validation:
- Exactly one field of `partition` must be set.
- `template.spec.selector` must be unset.
- `template.spec.targetRef` must reference a DaemonSet in this iteration.
- `template.spec.startupBoost` and `template.spec.resourcePolicy.containerPolicies[].startupBoost`
  must be unset, since every generated VPA carries a selector and the rule above applies to it.
- `VerticalPodAutoscalerSet` objects are rejected unless the `VPASet` feature gate is enabled.
- `metadata.name` must be at most 63 characters. The name is copied unchanged into the
  `autoscaling.k8s.io/vpaset-name` label on every generated VPA, and label values are limited to
  63 characters. A name limit is preferred over truncating the label value, because the label is
  the lookup key for a set's children and a truncated value would need its own hash to stay
  unique.

Generated VPAs carry:
- An `ownerReference` to the VPASet.
- A label `autoscaling.k8s.io/vpaset-name: <vpaset-name>` to list the children of a set.
- A label `autoscaling.k8s.io/vpaset-partition-value: <value>` to list the VPAs with a specific value.

The name of a generated VPA is `{vpaset-name}-{sanitized-value}-{hash}`. The value is lowercased
and non-DNS characters are replaced with hyphens, the result is truncated so the whole name is at
most 189 characters, and `{hash}` is a short hash of the raw value. The hash keeps names unique
when two values sanitize to the same string (e.g. `m5.xlarge` and `m5-xlarge`).

The 189-character cap, rather than the 253-character object name limit, comes from checkpoints.
The recommender names each checkpoint `{vpa-name}-{container-name}` and container names can be up
to 63 characters, so a VPA name longer than 253 - 1 - 63 = 189 characters would produce
checkpoint names that the API server rejects. With a 63-character VPASet name and a short hash,
this leaves well over 100 characters for the sanitized value.

### Component Changes

#### Recommender

When the `VPASelector` feature gate is enabled:

1. Load node labels - The recommender keeps a node informer and stores node labels in the
   cluster state model, so it can tell which node labels apply to a pod. Today this loads all
   nodes; loading only nodes that run pods of DaemonSets targeted by a VPA with a selector can
   be added later as an optimization.

2. Node-aware aggregation - Today pods with identical labels share one aggregate container
   state, which for a DaemonSet means all its pods share one aggregate. For DaemonSet pods, the
   aggregate state key is extended with the pod's node name, so usage from different nodes is
   kept in separate aggregates. Because a DaemonSet runs at most one pod per node, this yields
   at most one aggregate per node per container.

3. Matching aggregates to VPAs - When the recommender decides whether an aggregate belongs to
   a VPA, it checks the pod label selector derived from `targetRef` (as today) and
   also checks `selector.nodeSelector` against the current labels of
   the aggregate's node. When a node's labels change, the aggregates on that node are matched
   again so that they move to the VPA that now selects them.

4. One VPA per aggregate - Today `UseAggregationIfMatching` links an aggregate to every VPA
   whose selector matches it. With a selector VPA and a fallback VPA on the same DaemonSet, both
   would match the same node aggregate: the fallback would absorb usage from every node, and the
   aggregate's `UpdateMode` and resource policy would be overwritten by whichever VPA linked
   last. When the gate is enabled, the recommender instead resolves the controlling VPA for each
   DaemonSet node aggregate with the rules in
   [Choosing the controlling VPA for a pod](#choosing-the-controlling-vpa-for-a-pod) and links the
   aggregate only to that VPA. A fallback VPA therefore sees only the nodes no selector VPA
   covers. When the winner changes (a node is relabeled, a selector VPA is created or deleted),
   the aggregate is unlinked from the old VPA and linked to the new one.

Pods on nodes that no VPA selects - If a DaemonSet pod runs on a node that matches none of the
selector VPAs targeting the DaemonSet (for example, the node lacks the label a VPASet partitions
by), that pod does not contribute to any of those VPAs. If there is also a VPA without a selector
targeting the same DaemonSet, the pod falls back to it (see
[Choosing the controlling VPA for a pod](#choosing-the-controlling-vpa-for-a-pod)). Otherwise the
pod is not managed by VPA and keeps its existing resource requests.

#### VPASet Controller

The VPASet controller is a new, standalone component (`vpaset-controller`), shipped as its own
binary alongside the recommender, updater, and admission controller. It is optional.

The controller watches `VerticalPodAutoscalerSet`, `VerticalPodAutoscaler`, and `Node` objects.
For each VPASet it:

1. Computes the set of distinct values of `partition.nodeLabelKey` across all nodes.
2. For each value without a matching child VPA, creates one from `template`, with
   `spec.selector.nodeSelector = {<nodeLabelKey>: <value>}` and the labels and owner reference
   described above.
3. For each existing child VPA, updates its spec to match the current template. The VPASet is the
   source of truth for its children: manual edits to a generated VPA's spec are overwritten.
4. Deletes child VPAs whose value no longer appears on any node.
5. Updates `status.vpaCount`, `status.conditions`, and `status.observedGeneration`.

If a VPA with the intended name already exists and is not owned by the VPASet, the controller does
not adopt or modify it. It sets a `ConflictingVPA` condition on the VPASet and skips that value.

Open question: deletion grace period - Step 4 deletes a child VPA as soon as its label value no
longer appears on any node. A transient change, such as a rolling node upgrade that briefly
removes the last node carrying a value, would delete the VPA and its recommendation history, and
the pods on that value would fall back to a VPA without a selector, or to no VPA at all, until
the value returns. One option is to delay deletion by a grace period (for example, a configurable
duration after the value was last seen), at the cost of keeping unused VPAs for longer. Whether to
add the delay, its default, and whether it is configurable on the VPASet are not yet decided and
must be resolved before graduating to Beta (see [Graduation Criteria](#graduation-criteria)).

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

Validation - The validation rules listed under [API Changes](#api-changes) for both
`spec.selector` and `VerticalPodAutoscalerSet`. The admission controller therefore needs both
feature gates: `VPASelector` to accept VPAs with `spec.selector` (including the VPAs generated
by a VPASet), and `VPASet` to accept `VerticalPodAutoscalerSet` objects.

### Choosing the controlling VPA for a pod

Today, when more than one VPA matches a pod, `GetControllingVPAForPod` picks the oldest one (by
creation timestamp, then by name). With selectors, several VPAs targeting the same DaemonSet is an
expected setup, so the rule is extended:

1. A VPA matches a pod if its `targetRef` resolves to the pod's controller (as today) and, if
   `selector` is set, the pod's node labels match `selector.nodeSelector`.
2. Among the matching VPAs, a VPA with a selector takes precedence over a VPA without one. This
   lets users keep a plain VPA as a fallback for pods on nodes that no selector VPA covers.
3. If several VPAs remain, the existing tie-break applies (oldest first, then by name). With a
   single VPASet this cannot happen, because each node has exactly one value for the partition
   key. It can happen with hand-written selectors that overlap, or with two VPASets on the same
   target. The recommender reports this with a condition on the VPAs that lose the tie-break.

The recommender, updater, and admission controller all use this same function, so they always
agree on which VPA controls a pod.

`updateMode: Off` VPAs take part in the choice. Today the updater and the admission controller
skip Off VPAs before matching pods. With selectors that would let a fallback VPA in `Recreate`
mode hijack the pods of an Off partition, since the Off VPA is never considered and the fallback
is the only match left. Both components therefore resolve the controlling VPA over all VPAs
targeting the controller, including Off ones, and only then check its update mode: if the
winner is Off (and has no startup boost), the pod is left alone, even if a weaker VPA would
have acted on it. The updater, which today iterates per VPA, applies the same rule by dropping
from a VPA's pod list any pod whose controlling VPA is a different object.

### VPASet Lifecycle Example

1. A user creates a VPASet with `partition.nodeLabelKey: "node.kubernetes.io/instance-type"` and
   a template targeting a DaemonSet.
2. The VPASet controller finds the distinct values of the label across nodes (e.g. `m5.xlarge`,
   `c5.2xlarge`) and creates one VPA per value:
   - `my-set-m5-xlarge-<hash>` with `selector.nodeSelector: {"node.kubernetes.io/instance-type": "m5.xlarge"}`
   - `my-set-c5-2xlarge-<hash>` with `selector.nodeSelector: {"node.kubernetes.io/instance-type": "c5.2xlarge"}`
3. The recommender treats each as a regular VPA, computes its recommendation from the pods on the
   matching nodes, writes it to the VPA's `status.recommendation`, and writes checkpoints for it.
4. The updater and admission controller match each DaemonSet pod to the VPA for its node and
   apply updates according to the `updateMode` from the template.
5. When a node pool with a new instance type is added, the controller creates a new VPA.
6. When the last node with a given value is removed, the controller deletes that VPA, and the
   existing checkpoint garbage collection later removes its checkpoints.
7. When the user edits the VPASet template, the controller updates all generated VPAs.
8. When the VPASet is deleted, all generated VPAs are garbage collected through their
   `ownerReferences`, and their checkpoints follow.

#### Example

Given a two-node Kind cluster (`kind-worker`, `kind-worker2`) and the following manifests:

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
apiVersion: "autoscaling.k8s.io/v1alpha1"
kind: VerticalPodAutoscalerSet
metadata:
  name: test-daemonset
spec:
  partition:
    nodeLabelKey: "kubernetes.io/hostname"
  template:
    spec:
      targetRef:
        apiVersion: "apps/v1"
        kind: DaemonSet
        name: test-daemonset
      updatePolicy:
        updateMode: "Off"
```

The VPASet controller creates two VPAs, one per node. Each is a regular VPA with a node selector,
and its recommendation is computed from the usage on that node only:

```yaml
apiVersion: autoscaling.k8s.io/v1
kind: VerticalPodAutoscaler
metadata:
  name: test-daemonset-kind-worker-5f7c2
  namespace: default
  labels:
    autoscaling.k8s.io/vpaset-name: test-daemonset
    autoscaling.k8s.io/vpaset-partition-value: kind-worker
  ownerReferences:
  - apiVersion: autoscaling.k8s.io/v1alpha1
    controller: true
    kind: VerticalPodAutoscalerSet
    name: test-daemonset
    uid: 07a3f6e9-f497-4c9d-9f44-22ed88dfec24
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
  name: test-daemonset-kind-worker2-a91e4
  namespace: default
  labels:
    autoscaling.k8s.io/vpaset-name: test-daemonset
    autoscaling.k8s.io/vpaset-partition-value: kind-worker2
  ownerReferences:
  - apiVersion: autoscaling.k8s.io/v1alpha1
    controller: true
    kind: VerticalPodAutoscalerSet
    name: test-daemonset
    uid: 07a3f6e9-f497-4c9d-9f44-22ed88dfec24
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
wrong: with the default tolerance of 0.5, a partition holding half or fewer of the DaemonSet's
pods can never evict or resize, and a one-node partition (such as the hostname-partitioned
[example](#example) above) is always skipped by `minReplicas`.

For a VPA with `spec.selector`, `GetCreatorMaps` therefore sets `configured` to the number of live
selected pods, as it already does for Jobs. Tolerance and `minReplicas` then apply per partition,
which is the right unit: each DaemonSet pod serves its own node, so other partitions add no
availability to this one. Counting live pods instead of a desired count slightly loosens the
budget during rollouts; computing a per-partition desired count would mean re-running DaemonSet
scheduling against the matching nodes, which is not worth it for alpha.

`minReplicas` keeps its semantics and default. A VPASet whose partitions can be a single node must
set `updatePolicy.minReplicas: 1` in its template for eviction-based modes to work, as a
single-node DaemonSet with a plain VPA must today. The controller copies the template's
`updatePolicy` to every generated VPA, so no special handling is needed.
`--in-place-skip-disruption-budget` is unaffected.

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

### Test Plan

- **Unit tests**:
  - VPA matching with `selector.nodeSelector`,
    including precedence of selector VPAs over a VPA without a selector, the tie-break for
    overlapping selectors, and that an Off selector VPA wins over an active fallback VPA so the
    pod is not updated.
  - Recommender cluster state: a node aggregate is linked to exactly one VPA and the fallback VPA
    only receives aggregates from uncovered nodes; relinking when the winner changes.
  - Recommender cluster state: node-aware aggregate keys for DaemonSet pods, matching aggregates
    to VPAs by node labels, and re-matching when node labels change.
  - Updater, with a DaemonSet split across several selector VPAs: each pod is matched to the
    VPA whose `nodeSelector` matches its node's labels, a pod on an uncovered node goes to the fallback VPA or is skipped, and a pod is dropped from a
    VPA's list when its controlling VPA is a different object (including an Off one). For the
    disruption budget, `configured`, the eviction tolerance, and the `minReplicas` check are
    computed from the selected pods only, so a partition with a minority of the DaemonSet's pods
    can still evict and resize, a one-node partition with `minReplicas: 1` is updated, and
    disruptions in one partition do not consume another partition's budget.
  - Admission controller: extracting the node name from DaemonSet node affinity `matchFields`,
    and falling back correctly when the node cannot be determined (absent `matchFields`, empty
    `values`, malformed node affinity, etc.).
  - Validation rules for `spec.selector` and `VerticalPodAutoscalerSet` (DaemonSet-only, unset
    template selector, exactly one partition field, feature gates, rejection of startup boost
    together with a selector at both the VPA and template level, 63-character VPASet name limit).
  - VPASet controller: creating VPAs for new values, deleting VPAs for vanished values,
    propagating template changes, overwriting manual edits to generated VPAs, name generation
    and collision avoidance (including the 189-character cap and that the resulting checkpoint
    names fit in 253 characters), and refusing to adopt VPAs it does not own.
- **E2E tests**:
  - A VPASet on a heterogeneous cluster creates one VPA per label value, and each receives an
    independent recommendation.
  - Deleting the VPASet garbage collects the generated VPAs and, after that, their checkpoints.
  - In-place updates apply the correct VPA's recommendation to each pod.
  - Eviction-based updates (`Recreate`) evict pods, and the admission controller injects the
    correct recommendation into the replacement pod. The test uses at least three partitions,
    so that each holds a minority of the DaemonSet's pods and a DaemonSet-wide budget would have
    blocked eviction.
  - Pods on nodes without the partition label fall back to a VPA without a selector, if one
    exists, and are otherwise left untouched. The fallback VPA's recommendation reflects only
    those pods.
  - A partition whose VPA is `Off` is not evicted or resized by a fallback VPA in `Recreate`
    mode targeting the same DaemonSet.

### Feature Enablement and Rollback

- **Feature gate names**:
  - `VPASelector`: enables the `spec.selector` field. Used by the recommender (computes a VPA's
    recommendation only from the selected pods), the updater (applies it only to the selected
    pods), and the admission controller (accepts VPAs with `spec.selector` and injects the
    recommendation only into selected pods).
  - `VPASet` (on by default): enables the `VerticalPodAutoscalerSet` resource and its
    controller. Used only by the `vpaset-controller` and the admission controller (validation).
    The recommender and
    updater never read VPASets; they only see the generated VPAs, so they need `VPASelector`
    but not `VPASet`.
- **When the gates are enabled**: VPAs with a selector only apply to the selected pods, and
  VPASets create and maintain their child VPAs.
- **When `VPASet` is disabled after being enabled, or the `vpaset-controller` is removed**: The
  controller stops reconciling. Generated VPAs stay in the cluster and keep working as regular
  VPAs with a selector, but are no longer created, updated, or deleted as nodes change. They are
  still garbage collected if the VPASet is deleted.
- **When `VPASelector` is disabled after being enabled**: Components ignore `spec.selector`. A VPA
  with a selector then behaves like a VPA without one, so several VPAs for the same DaemonSet
  all match every pod, and the existing oldest-wins rule picks one of them for all pods. Operators
  rolling back should delete the VPASets (or the generated VPAs) and, if needed, create a single
  VPA without a selector.
- **When `spec.selector` is removed from a VPA**: The VPA goes back to covering all pods of its
  target. Its existing checkpoints remain valid and are reused.

### Graduation Criteria

- **Alpha → Beta**:
  - E2E tests are stable for at least 2 releases.
  - No open bugs against the `VPASelector` or `VPASet` feature gates.
  - Positive user feedback from at least two production deployments.
  - Open questions resolved, in particular the deletion grace period.
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
- **VPASet without the `vpaset-controller`**: Nothing reconciles the VPASet, so no VPAs are
  created.
- **`vpaset-controller` with components that do not support `VPASelector`**: The controller
  creates selector VPAs, which then behave as described in the cases above. The
  `vpaset-controller` itself only needs the VPA and VPASet CRDs to be installed.

Recommended rollout order: CRDs first, then the admission controller and updater, then the
recommender, and finally the `vpaset-controller`.

### Kubernetes Version Compatibility

When using `updateMode: Off`, `Recreate`, or `InPlaceOrRecreate` (falling back to eviction),
this feature works on any Kubernetes version; no in-place update support is required.

When using `updateMode: InPlace` or `InPlaceOrRecreate` (for the in-place path), this feature
requires Kubernetes 1.27+ with the `InPlacePodVerticalScaling` feature gate enabled (alpha in
1.27, beta in 1.33). On Kubernetes versions without in-place pod updates, the `InPlace` update
mode will not function, and the admission controller rejects VPAs and VPASet templates that
combine `selector` with `InPlace` mode when the `InPlace` feature gate is not enabled.

## Implementation History

- 2026-09-01: Initial AEP version.

## Alternatives

### One DaemonSet and VPA per node pool

Operators can manually create separate DaemonSets (with node selectors) and corresponding VPAs
for each node pool. This works but is operationally heavy: it requires updating DaemonSet
manifests every time node pools are added or removed, risks configuration drift, and scales
poorly in clusters with many node pools.

### Embedding per-group recommendations in VPA status

Storing every group's recommendation in the parent VPA's status (e.g. `status.groups`) avoids new
objects, but the VPA grows with the number of groups and containers and could approach the etcd
per-object size limit. Every consumer would also have to re-read the whole VPA on any group's
change.
