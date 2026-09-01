# AEP-7942: Vertical Pod Autoscaling for DaemonSets with Heterogeneous Resource Requirements

<!-- toc -->
- [Summary](#summary)
- [Motivation](#motivation)
  - [Goals](#goals)
  - [Non-Goals](#non-goals)
- [Proposal](#proposal)
- [Design Details](#design-details)
  - [API Changes](#api-changes)
    - [New field on VerticalPodAutoscalerSpec](#new-field-on-verticalpodautoscalerspec)
    - [New CRD: VerticalPodAutoscalerSlice](#new-crd-verticalpodautoscalerslice)
    - [New CRD: VerticalPodAutoscalerSliceCheckpoint](#new-crd-verticalpodautoscalerslicecheckpoint)
  - [Component Changes](#component-changes)
    - [Recommender](#recommender)
    - [Updater](#updater)
    - [Admission Controller](#admission-controller)
  - [VPASlice Lifecycle](#vpaslice-lifecycle)
  - [Design Decisions](#design-decisions)
    - [Why separate VPASlice CRDs instead of embedding in VPA status](#why-separate-vpaslice-crds-instead-of-embedding-in-vpa-status)
    - [Why separate VPASliceCheckpoint CRDs instead of reusing VPA checkpoints](#why-separate-vpaslicecheckpoint-crds-instead-of-reusing-vpa-checkpoints)
    - [Why InPlace-only](#why-inplace-only)
  - [Test Plan](#test-plan)
  - [Feature Enablement and Rollback](#feature-enablement-and-rollback)
  - [Graduation Criteria](#graduation-criteria)
  - [Version Skew](#version-skew)
  - [Kubernetes Version Compatibility](#kubernetes-version-compatibility)
- [Implementation History](#implementation-history)
- [Alternatives](#alternatives)
<!-- /toc -->

## Summary

VPA today computes a single recommendation per VPA object and applies it uniformly to all matching
pods. This works well for interchangeable replicas (e.g. Deployments) but breaks down for
DaemonSets running across heterogeneous nodes. A security agent on a GPU node may consume far more
CPU and memory than the same agent on a CPU-only node, yet VPA produces one recommendation for
both. The result is either over-provisioning on quiet nodes or under-provisioning (throttling, OOM
kills) on busy ones.

This AEP introduces **VPASlice**, a mechanism that partitions VPA recommendations by node label.
When a VPA targeting a DaemonSet sets `spec.sliceByNodeLabel` to a node label key (e.g.
`node.kubernetes.io/instance-type`), the recommender creates a `VerticalPodAutoscalerSlice` object
for each distinct label value observed across matching pods' nodes. Each VPASlice carries an
independent recommendation computed from the resource usage of pods on that subset of nodes.

The updater and admission controller use VPASlice recommendations to apply the correct
per-node-group resource values to each pod.

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

- Allow VPA to produce independent resource recommendations for subsets of a DaemonSet's pods,
  grouped by a user-specified node label.
- Introduce new CRDs (`VerticalPodAutoscalerSlice` and `VerticalPodAutoscalerSliceCheckpoint`)
  that store per-group recommendations and checkpoints respectively.
- Ensure the recommender, updater, and admission controller all participate in the VPASlice
  lifecycle.
- Gate the feature behind an alpha feature flag (`VPASlice`) so it can be adopted incrementally.

### Non-Goals

- Support for workload types other than DaemonSet. While the VPASlice mechanism could
  conceptually apply to other controllers, this AEP restricts scope to DaemonSets only.
- Support for update modes other than `InPlace`. VPASlice currently requires
  `spec.updatePolicy.updateMode: InPlace` because DaemonSet pods scheduled via node affinity may
  not have `NodeName` set at admission time, preventing the admission controller from matching
  them to the correct slice. InPlace mode avoids eviction loops caused by this limitation.
- CPU startup boost for VPASlice-managed pods.
- Multi-label slicing (slicing by more than one label key simultaneously).
- Automatic node pool discovery or integration with cloud provider node group APIs.

## Proposal

Add an optional field `spec.sliceByNodeLabel` to `VerticalPodAutoscalerSpec`. When this field is
set and the VPA targets a DaemonSet:

1. The **recommender** discovers the distinct values of the specified node label across nodes that
   run pods matched by the VPA. For each distinct value, it creates a
   `VerticalPodAutoscalerSlice` object with an `ownerReference` pointing to the parent VPA (for
   garbage collection). It then computes independent recommendations for each slice and writes
   them to the slice's `status.recommendation`.

2. The **updater** reads VPASlice objects and matches each pod to its correct slice by looking up
   the pod's node labels. It applies in-place updates using the slice-specific recommendation
   rather than the parent VPA's (now empty) recommendation.

3. The **admission controller** similarly matches incoming pods to their VPASlice by resolving the
   pod's node and checking the slice's `nodeSelector`. It injects the slice-specific
   recommendation into the pod's resource requests.

When `spec.sliceByNodeLabel` is set, the parent VPA's `status.recommendation` is intentionally
left empty. Consumers should read per-group recommendations from the associated VPASlice objects.

## Design Details

### API Changes

#### New field on VerticalPodAutoscalerSpec

A new optional field is added to the existing `VerticalPodAutoscalerSpec`:

```go
type VerticalPodAutoscalerSpec struct {
    // ... existing fields ...

    // SliceByNodeLabel enables per-node-group recommendations for workloads
    // (typically DaemonSets) running across heterogeneous nodes. When set,
    // the recommender creates a VerticalPodAutoscalerSlice object for each
    // distinct value of this node label, containing independent recommendations
    // computed from pods on nodes with that label value.
    //
    // For example, setting this to "node.kubernetes.io/instance-type" produces
    // separate recommendations for pods on m5.xlarge nodes vs c5.2xlarge nodes.
    //
    // When set, the VPA's own status.recommendation is left empty — consumers
    // should read recommendations from the associated VPASlice objects instead.
    // +optional
    SliceByNodeLabel *string `json:"sliceByNodeLabel,omitempty"`
}
```

The admission controller validates that:
- The feature gate `VPASlice` is enabled.
- `targetRef.Kind` is `DaemonSet`.
- `updatePolicy.updateMode` is `InPlace`.

#### New CRD: VerticalPodAutoscalerSlice

```go
type VerticalPodAutoscalerSlice struct {
    metav1.TypeMeta   `json:",inline"`
    metav1.ObjectMeta `json:"metadata,omitempty"`
    Spec   VerticalPodAutoscalerSliceSpec   `json:"spec"`
    Status VerticalPodAutoscalerSliceStatus `json:"status,omitempty"`
}

type VerticalPodAutoscalerSliceSpec struct {
    // VPAName is the name of the parent VerticalPodAutoscaler object.
    VPAName string `json:"vpaName"`

    // NodeSelector identifies the set of nodes this slice covers.
    // Typically a single key-value pair derived from the parent VPA's
    // spec.sliceByNodeLabel field.
    NodeSelector map[string]string `json:"nodeSelector"`
}

type VerticalPodAutoscalerSliceStatus struct {
    // Recommendation is the most recently computed resource recommendation.
    // Reuses the same RecommendedPodResources type as VPA status.
    Recommendation *RecommendedPodResources `json:"recommendation,omitempty"`

    // Conditions indicates whether a recommendation could be computed.
    Conditions []VerticalPodAutoscalerCondition `json:"conditions,omitempty"`

    // ObservedGeneration tracks the most recent generation observed.
    ObservedGeneration *int64 `json:"observedGeneration,omitempty"`
}
```

VPASlice objects carry:
- A label `autoscaling.k8s.io/vpa-name` referencing the parent VPA (analogous to
  `kubernetes.io/service-name` on EndpointSlice).
- An `ownerReference` to the parent VPA for automatic garbage collection.

The naming convention for VPASlice objects is `{vpa-name}-{sanitized-label-value}`, where the
label value is lowercased and non-DNS characters are replaced with hyphens, truncated to fit the
253-character Kubernetes name limit.

#### New CRD: VerticalPodAutoscalerSliceCheckpoint

```go
type VerticalPodAutoscalerSliceCheckpoint struct {
    metav1.TypeMeta   `json:",inline"`
    metav1.ObjectMeta `json:"metadata,omitempty"`
    Spec   VerticalPodAutoscalerSliceCheckpointSpec   `json:"spec,omitempty"`
    Status VerticalPodAutoscalerSliceCheckpointStatus `json:"status,omitempty"`
}

type VerticalPodAutoscalerSliceCheckpointSpec struct {
    VPAObjectName string `json:"vpaObjectName,omitempty"`
    VPASliceName  string `json:"vpaSliceName,omitempty"`
}

type VerticalPodAutoscalerSliceCheckpointStatus struct {
    LastUpdateTime       metav1.Time                    `json:"lastUpdateTime,omitempty"`
    Version              string                         `json:"version,omitempty"`
    ContainerCheckpoints []ContainerHistogramCheckpoint  `json:"containerCheckpoints,omitempty"`
}

type ContainerHistogramCheckpoint struct {
    ContainerName   string              `json:"containerName"`
    CPUHistogram    HistogramCheckpoint `json:"cpuHistogram,omitempty"`
    MemoryHistogram HistogramCheckpoint `json:"memoryHistogram,omitempty"`
    FirstSampleStart metav1.Time        `json:"firstSampleStart,omitempty"`
    LastSampleStart  metav1.Time        `json:"lastSampleStart,omitempty"`
    TotalSamplesCount int               `json:"totalSamplesCount,omitempty"`
}
```

Unlike `VerticalPodAutoscalerCheckpoint` (which stores one container per object), a slice
checkpoint packs all containers into a single object, yielding one checkpoint per VPASlice
regardless of container count.

### Component Changes

#### Recommender

The recommender's main loop gains three new steps when the `VPASlice` feature gate is enabled:

1. **LoadVPASlices**: Ensures VPASlice objects exist for each distinct label value observed across
   matching pods' nodes. Creates missing slices with `ownerReferences` to the parent VPA. Then
   loads all VPASlice CRDs into the cluster state model, removing slices whose parent VPA is no
   longer active.

2. **LoadNodes**: Loads node labels into the cluster state so that pod-to-slice association can be
   resolved (pods are assigned to aggregation keys that carry their node's label value).
   Currently this loads all nodes; optimizing to load only nodes that run pods matched by a
   sliced VPA will be addressed as a follow-up feature.

3. **UpdateVPASlices**: For each VPASlice, aggregates container states from matching pods,
   computes recommendations using the same `PodResourceRecommender` as regular VPAs, and patches
   the slice's `status` via JSON Patch. VPAs with `sliceByNodeLabel` set are skipped during the
   regular `UpdateVPAs` step — their `status.recommendation` remains empty.

4. **MaintainCheckpointSlices**: Writes `VerticalPodAutoscalerSliceCheckpoint` objects for each
   VPASlice, persisting per-container histogram data. On startup, checkpoints are loaded back via
   `InitFromCheckpointSlices`.

The cluster state model is extended with:
- `VpaSlice` model type holding aggregation state, conditions, and recommendation per slice.
- `VpaSliceID` type for indexing slices.
- `nodeLabels` map on `clusterState` for resolving pod-to-node-label associations.
- `AddOrUpdateVpaSlice`, `DeleteVpaSlice`, `AddOrUpdateNode`, `DeleteNode` methods on the
  `ClusterState` interface.
- `AggregateStateKey` extended with a `NodeLabelValue()` accessor so aggregations can be scoped
  to a specific slice.

#### Updater

The updater is extended to:

1. **List VPASlice objects** and pair each with its parent VPA and the parent's pod selector
   (via `VpaSliceWithNodeSelector`).
2. **Skip VPAs with `sliceByNodeLabel`** from the normal VPA update loop.
3. **Match pods to slices** by resolving the pod's node labels and comparing against each slice's
   `nodeSelector` (via `GetControllingVPASliceForPod`).
4. **Apply in-place updates** using the slice-specific recommendation. The pod update logic
   reuses the existing in-place update pipeline; the VPASlice's recommendation is overlaid onto a
   copy of the parent VPA to produce a "virtual VPA" with the correct per-slice recommendation.

#### Admission Controller

The admission controller currently supports **validation only** for VPASlice-enabled VPAs. It
does not mutate (inject recommendations into) newly admitted pods because at admission time the
pod has not yet been scheduled — `NodeName` is empty — so there is no way to determine which
node the pod will land on and therefore which VPASlice recommendation to apply. This is the
primary reason `sliceByNodeLabel` requires `InPlace` update mode: once the pod is scheduled and
its `NodeName` is known, the updater applies the correct slice-specific recommendation via
in-place update.

A follow-up effort will explore solving this gap — for example, by simulating the scheduler's
node selection to predict the target node at admission time — which could enable support for
additional update modes beyond `InPlace`.

The admission controller is extended with validation only:

1. Reject VPA objects that set `sliceByNodeLabel` unless the target is a DaemonSet.
2. Reject VPA objects that set `sliceByNodeLabel` unless the update mode is `InPlace`.
3. Reject VPA objects that set `sliceByNodeLabel` unless the `VPASlice` feature gate is enabled.

### VPASlice Lifecycle

1. User creates a VPA with `spec.sliceByNodeLabel: "node.kubernetes.io/instance-type"` targeting
   a DaemonSet, with `updateMode: InPlace`.
2. The recommender's `ensureVPASlices` discovers all distinct values of the label across nodes
   running matched pods (e.g. `m5.xlarge`, `c5.2xlarge`) and creates VPASlice objects:
   - `my-vpa-m5-xlarge` with `nodeSelector: {"node.kubernetes.io/instance-type": "m5.xlarge"}`
   - `my-vpa-c5-2xlarge` with `nodeSelector: {"node.kubernetes.io/instance-type": "c5.2xlarge"}`
3. Each recommender iteration computes independent recommendations per slice and writes them to
   the slice's `status.recommendation`.
4. The updater matches each DaemonSet pod to its slice and applies in-place resource updates.
5. If a new node pool is added with a new instance type, the recommender creates a new VPASlice
   on its next iteration.
6. If the parent VPA is deleted, all VPASlice objects are garbage collected via `ownerReferences`.

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
apiVersion: "autoscaling.k8s.io/v1"
kind: VerticalPodAutoscaler
metadata:
  name: test-daemonset-vpa
spec:
  targetRef:
    apiVersion: "apps/v1"
    kind: DaemonSet
    name: test-daemonset
  updatePolicy:
    updateMode: "Off"
  sliceByNodeLabel: "kubernetes.io/hostname"
```

The recommender automatically creates two VPASlice objects — one per node — each with
independent recommendations derived from the resource usage on that specific node:

```yaml
apiVersion: autoscaling.k8s.io/v1
kind: VerticalPodAutoscalerSlice
metadata:
  name: test-daemonset-vpa-kind-worker
  namespace: default
  ownerReferences:
  - apiVersion: autoscaling.k8s.io/v1
    controller: true
    kind: VerticalPodAutoscaler
    name: test-daemonset-vpa
    uid: 07a3f6e9-f497-4c9d-9f44-22ed88dfec24
spec:
  nodeSelector:
    kubernetes.io/hostname: kind-worker
  vpaName: test-daemonset-vpa
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
kind: VerticalPodAutoscalerSlice
metadata:
  name: test-daemonset-vpa-kind-worker2
  namespace: default
  ownerReferences:
  - apiVersion: autoscaling.k8s.io/v1
    controller: true
    kind: VerticalPodAutoscaler
    name: test-daemonset-vpa
    uid: 07a3f6e9-f497-4c9d-9f44-22ed88dfec24
spec:
  nodeSelector:
    kubernetes.io/hostname: kind-worker2
  vpaName: test-daemonset-vpa
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
workload, driving its memory usage higher than the pod on `kind-worker2`. The VPASlice
recommendations reflect this: `323522422` bytes (~308Mi) on `kind-worker` vs `250Mi` on
`kind-worker2`.

The `kubectl get vpaslice` output confirms the per-node recommendations at a glance:

```
NAMESPACE   NAME                              VPA                  NODESELECTOR                                CPU   MEM         PROVIDED   AGE
default     test-daemonset-vpa-kind-worker    test-daemonset-vpa   {"kubernetes.io/hostname":"kind-worker"}    25m   323522422   True       3h16m
default     test-daemonset-vpa-kind-worker2   test-daemonset-vpa   {"kubernetes.io/hostname":"kind-worker2"}   25m   250Mi       True       3h16m
```

### Design Decisions

#### Why separate VPASlice CRDs instead of embedding in VPA status

etcd enforces a [per-object storage limit](https://etcd.io/docs/v3.8/dev-guide/limit/) (default
1.5 MiB). A VPA's `status` already contains
recommendation data, conditions, and other metadata. When slicing by a label with many distinct
values (e.g. `node.kubernetes.io/instance-type` across dozens of instance types), embedding all
per-group recommendations inside the parent VPA's `status.groups` would push the object toward
or past this limit — especially for workloads with multiple containers, each producing target,
lower-bound, upper-bound, and uncapped-target recommendations.

Separate VPASlice objects avoid this problem by distributing recommendations across many small
objects, each well within etcd limits regardless of how many slices exist. This also provides:
- Independent watch semantics — consumers interested in a single node group can watch one slice
  instead of re-parsing the entire VPA on every change.
- Separate RBAC — operators can grant read access to slices without granting access to the
  parent VPA.
- Automatic garbage collection via `ownerReferences`.
- A pattern consistent with how Kubernetes handles similar fan-out (EndpointSlice).

#### Why separate VPASliceCheckpoint CRDs instead of reusing VPA checkpoints

The existing `VerticalPodAutoscalerCheckpoint` stores histogram data for a single container of a
single VPA. If we reused this type for slices, a VPA sliced across N label values with M
containers would produce N × M checkpoint objects. This multiplicative growth creates pressure on
etcd object count and API server list/watch overhead.

`VerticalPodAutoscalerSliceCheckpoint` solves this by packing all container histograms for a
single slice into one object (a `containerCheckpoints` list), yielding exactly one checkpoint per
VPASlice regardless of container count. This keeps checkpoint count proportional to the number of
slices (N) rather than N × M.

#### Why InPlace-only

When a new DaemonSet pod is admitted, the scheduler has not yet assigned it to a node —
`pod.spec.nodeName` is empty. Without knowing which node the pod will run on, the admission
controller cannot determine which VPASlice recommendation to inject. Allowing eviction-based
update modes (`Recreate`, `InPlaceOrRecreate`) would cause eviction loops: the updater evicts a
pod, the replacement arrives without the correct recommendation (because the admission controller
cannot match it to a slice), and the cycle repeats.

Restricting to `InPlace` avoids this: the pod is scheduled first, and once its `NodeName` is
known the updater applies the correct slice-specific recommendation in place without restarting
the pod. A follow-up effort will explore predicting the target node at admission time (e.g. via
scheduling simulation) to lift this restriction.

### Test Plan

- **Unit tests**: All new functions in the recommender, updater, and admission controller have
  corresponding unit tests, including:
  - VPASlice matching logic (`PodMatchesVPASlice`, `GetControllingVPASliceForPod`).
  - Cluster state slice management (`AddOrUpdateVpaSlice`, `DeleteVpaSlice`, aggregation
    matching).
  - Admission controller slice matching (`matchVPASlice`).
  - Validation rules (DaemonSet-only, InPlace-only).
  - Checkpoint slice writer and reader.
- **E2E tests**: Scenarios to cover:
  - DaemonSet with heterogeneous nodes receives per-node-group recommendations.
  - VPASlice objects are created and garbage collected with the parent VPA.
  - In-place updates apply the correct slice recommendation per pod.

### Feature Enablement and Rollback

- **Feature gate name**: `VPASlice`
- **Components depending on the feature gate**: recommender, updater, admission-controller.
- **When the gate is enabled**: The recommender creates VPASlice and VPASliceCheckpoint objects
  for VPAs with `sliceByNodeLabel` set. The updater and admission controller use VPASlice
  recommendations for matching pods. VPAs with `sliceByNodeLabel` skip the normal recommendation
  pipeline (their `status.recommendation` stays empty).
- **When the gate is disabled after being enabled**: Existing VPASlice and VPASliceCheckpoint
  objects remain in the cluster but are not read or updated by any component. The parent VPA's
  `status.recommendation` will begin to be populated again on the next recommender cycle
  (since the `sliceByNodeLabel` skip is gated). Pods will receive the global recommendation
  rather than per-slice recommendations. The VPASlice CRDs can be manually cleaned up, or will
  be garbage collected when the parent VPA is deleted.
- **When the VPA object with `sliceByNodeLabel` is updated to remove the field**: The recommender
  stops creating new slices and starts writing recommendations to the VPA's `status.recommendation`
  again. Existing VPASlice objects are garbage collected via `ownerReferences` when the parent VPA
  is deleted, or can be manually removed.

### Graduation Criteria

- **Alpha → Beta**:
  - E2E tests are stable for at least 2 releases.
  - No open bugs against the `VPASlice` feature gate.
  - Positive user feedback from at least two production deployments.
  - Address known limitations (admission controller NodeName gap).
- **Beta → GA**:
  - Feature has been stable in beta for at least 2 releases.
  - Consider expanding support beyond DaemonSet and InPlace-only.

### Version Skew

The `VPASlice` feature gate must be enabled on all three components (recommender, updater,
admission-controller) for the feature to work correctly:

- **New recommender, old updater/admission-controller**: The recommender creates VPASlice objects
  and skips writing recommendations to the parent VPA. The old updater and admission controller
  ignore VPASlice objects and see an empty `status.recommendation` on the parent VPA, resulting
  in no updates being applied to DaemonSet pods with `sliceByNodeLabel`.
- **New updater, old recommender**: The updater looks for VPASlice objects but finds none (the
  old recommender doesn't create them). It falls through to normal VPA processing.
- **New admission-controller, old recommender**: The admission controller attempts VPASlice
  matching but finds no slices. It returns `nil` for the matching VPA, and the pod receives no
  VPA-injected resources — the same behavior as when no VPA matches.

The feature gate fully mitigates skew: all components must have the gate enabled for
`sliceByNodeLabel` VPAs to function. Enabling the gate on only some components is safe (no
crashes or data corruption) but results in sliced VPAs being effectively inactive.

### Kubernetes Version Compatibility

This feature requires Kubernetes 1.27+ with the `InPlacePodVerticalScaling` feature gate enabled
(alpha in 1.27, beta in 1.33). When running on an older Kubernetes version that does not support
in-place pod updates, the `InPlace` update mode (which is a prerequisite for `sliceByNodeLabel`)
will not function, and the admission controller will reject VPA objects that specify both
`sliceByNodeLabel` and `InPlace` mode if the `InPlace` feature gate is not enabled.

## Implementation History

- 2026-09-01: Initial AEP version.

## Alternatives

### One DaemonSet and VPA per node pool

Operators can manually create separate DaemonSets (with node selectors) and corresponding VPAs
for each node pool. This works but is operationally heavy: it requires updating DaemonSet
manifests every time node pools are added or removed, risks configuration drift, and scales
poorly in clusters with many node pools.
