# One-shot (ephemeral) CapacityBuffers via a refill strategy

#### Author: jicowan

This is a follow-up to the [CapacityBuffer API proposal](./buffers.md) (#8151). It adds one-shot
(ephemeral) buffers for gang / batch workloads. A one-shot buffer is configured with a new
`refillStrategy` field, which controls what happens when a workload consumes the buffer's capacity:

* `recreate` — replace the used capacity (today's buffers)
* `recreateUpToLimit` — replace it up to a limit (capped buffers)
* `none` — don't replace it (one-shot buffers, this proposal)

`refillStrategy` is separate from `provisioningStrategy`, which describes the kind of capacity, so
any kind of buffer can be one-shot. An earlier draft added one-shot as a new `provisioningStrategy`
value instead; this change follows review discussion with @kkonrad and @jbtk.

# Timeline

## Alpha

- [ ] Agree on the refill-strategy field name/values (this proposal, coordinated with capped buffers)
- [ ] [Reference implementation](https://github.com/jicowan/karpenter/tree/ephemeral-capacity-buffers) in a buffer-consuming autoscaler (Karpenter — implementation ready,
      including `CapacityBufferBinding`; the selector is currently an interim annotation and will
      adopt the spec field)
- [ ] Document the one-shot (no-refill) + fill-deadline contract for other consumers (CA, GKE CCC)

## Beta / V1 graduation criteria

- [ ] Second OSS implementation of the one-shot behavior
- [ ] E2e coverage of the fill → stop → no-refill lifecycle (and the fill-deadline path), including
      at scale (the reference implementation tested a run of 200 matching pods, with churn, and overlapping
      buffers)
- [ ] In use for at least one full version behind the CapacityBuffer feature gate

# Summary

The `CapacityBuffer` API (autoscaling.x-k8s.io) today maintains *steady-state* spare capacity:
capacity consumed by real workloads is **recreated** to keep the buffer at its configured size. That
is the right behavior for maintaining headroom in the cluster, but wrong behavior for **gang / batch
workloads**, which want capacity pre-provisioned for a specific run and then provisioning to
**stop** — not a buffer that refills after the run and keeps incurring cost.

This proposal makes the buffer's **refill behavior** an explicit, first-class citizen of the
`CapacityBuffer` spec, orthogonal to `provisioningStrategy` (which describes the *kind* of capacity,
e.g. `active-capacity`, GKE `standby-capacity`). The axis has three values:

| refill behavior | meaning | buffer type |
|---|---|---|
| recreate *(default)* | consumed capacity is recreated to maintain size | current buffers |
| recreate-up-to-limit | recreate, bounded by a limit | capped buffers (in review) |
| **none** | consumed capacity is **not** recreated — one-shot | **ephemeral (this proposal)** |

The `none` value gives batch its **completion semantics**: provision once, let the workload consume
the capacity, then stop. This proposal contributes the `none` value, **monotonic shrink-as-fill** (a
`consumedReplicas` high-water mark so a partially-filled buffer provisions only its unfilled
remainder and never recreates consumed capacity), and an optional **`fillDeadlineSeconds`** which time-bounds how
long an unfilled one-shot buffer will keep trying.

Fill is tracked with a new, small resource, **`CapacityBufferBinding`**: one object per pod credited
toward a buffer, named by the **pod's UID**. The API server therefore guarantees a pod fills _at most one
buffer_, the credit survives the pod and autoscaler restarts. Users can also see which pods
filled a buffer because `consumedReplicas` is derived from a buffer's bindings.

This proposal also defines the selector that names which pods fill a buffer, `matchingPodSelector`.
It was previously deferred to the capped buffers proposal; the field is specified
in this proposal and capped buffers is free to reuse it.

# Motivation

There is demand for batch/Kueue-oriented capacity provisioning
([kubernetes-sigs/karpenter#2571](https://github.com/kubernetes-sigs/karpenter/issues/2571),
[#742](https://github.com/kubernetes-sigs/karpenter/issues/742),
[#3138](https://github.com/kubernetes-sigs/karpenter/issues/3138)). The canonical case is Kueue's
cold-start: a job is admitted, its pods transition to a pending state, the autoscaler reacts, and nodes appear late.
Kueue drives Cluster Autoscaler's `ProvisioningRequest` for this, but that API is not implemented across
all autoscalers; expressing the intent through the portable `CapacityBuffer` API reaches any
buffer-consuming autoscaler.

The base CapacityBuffer proposal already anticipated the workload-filled case under "Out of scope,
may be added as follow up proposals" (*"Buffer dedicated for workload: … count the pods matching the
selector as filling in the buffer space"*). That sketch is steady-state; batch additionally needs
the buffer to **stop** once filled. Framing this as a refill-strategy value (rather than a new
provisioning strategy) lets the same one-shot semantics compose with **any** capacity kind — e.g. a
one-shot GKE `standby-capacity` buffer, not just active capacity.

## Goals

* Add an explicit **refill-strategy** field to `CapacityBufferSpec`, orthogonal to
  `provisioningStrategy`, with a **`none`** value meaning "do not recreate consumed capacity"
  (one-shot).
* Give one-shot buffers **completion semantics** (a terminal state; no refill after fill).
* Add an optional **`fillDeadlineSeconds`** so an unfilled one-shot buffer does not hold provisioning intent
  indefinitely.
* Define how a buffer names the pods that fill it (**`matchingPodSelector`**) and how fill is
  recorded (**`CapacityBufferBinding`**): exclusive per pod, durable across pod deletion and
  autoscaler restarts, and observable.
* Compose with the existing (recreate) and capped (recreate-up-to-limit) behaviors,
  rather than introducing a parallel mechanism.

## Non-Goals

* **Reserving/guaranteeing capacity for specific pods.** Exclusivity remains out-of-band (node
  taints) until the scheduler offers capacity reservation — same as the base proposal.
* **Gang pod scheduling.** This provisions *nodes*; all-or-nothing binding stays with the gang
  scheduler (Kueue, Volcano, YuniKorn, coscheduling).
* **Auto-deleting the buffer object.** A completed one-shot buffer is terminal but not garbage
  collected; cleanup is left to the user or a future TTL follow-up.
* **`scalableRef` (and `percentage`) one-shot buffers — out of v1 scope.** A `scalableRef` buffer
  copies its pod shape from a workload (a Deployment, StatefulSet or ReplicaSet), and `percentage`
  sizes it as a share of that workload's current replica count. That count changes whenever the
  workload scales, so a one-shot buffer would have no fixed target to fill. v1 requires
  `podTemplateRef` + `replicas` (see the next item); `scalableRef` is deferred to a follow-up.
* **Mutable / derived targets — out of v1 scope.** For v1 one-shot, the target chunk count is
  **exactly `spec.replicas`**, and it must be **fixed** for the cycle. v1 requires
  **`podTemplateRef` + `replicas`** and:
  * **`limits` is not supported with `refillStrategy: none`.** The base API lets `limits` cap the
    resolved target below `replicas` (a `replicas: 8` object can resolve to a target of 4), which
    would make the `target − consumedReplicas` remainder ambiguous. A one-shot buffer that also sets
    `limits` should be rejected (or `limits` ignored) in v1.
  * **`spec.replicas` is treated as immutable for the cycle.** A one-shot buffer does not re-arm;
    editing `replicas` mid-cycle is unsupported — recreate the object to start a new cycle.
  Supporting `limits` or a mutable/derived target requires persisting a resolved-target snapshot; see
  [snapshotting](#snapshotting-the-target-for-mutable-target-buffers). That is the forward path, not
  v1.

# Proposal

## The refill-strategy axis

Refill behavior becomes a spec field, orthogonal to `provisioningStrategy`. Naming is an open
question carried over from the PR thread — the two candidates are `refillStrategy` and
`onCapacityConsumption`; this document uses `refillStrategy`, values TBD with
capped buffers and @jbtk:

| `refillStrategy` | alt. (`onCapacityConsumption`) | behavior |
|---|---|---|
| `recreate` *(default)* | `recreate` | current: consumed capacity is recreated to maintain size |
| `recreateUpToLimit` | `recreateUpToLimit` | capped: recreate, bounded (capped buffers proposal) |
| `none` | `doNothing` | one-shot: consumed capacity is **not** recreated (this proposal) |

Choosing `refillStrategy: none` on a buffer of any `provisioningStrategy` makes it one-shot: it
provisions its configured capacity once, the workload consumes it, and the buffer does not recreate
what was consumed. Once fully consumed the buffer is **terminal** (a `Fulfilled` status condition,
below); it will not refill even if the workload later leaves and the nodes are reclaimed by normal
consolidation.

## Shrink-as-fill (monotonic)

A *chunk* is one replica of the buffer: room for one pod built from the buffer's pod template. A
buffer with `replicas: 8` is 8 chunks. Each unfilled chunk is represented by one virtual pod.

A one-shot buffer **shrinks as it fills**: as a matching workload's pods are scheduled onto the
buffer's capacity, the buffer provisions only the *unfilled remainder* rather than its full size.
A status field, `consumedReplicas`, counts how many chunks the workload has used so far. The
autoscaler holds only `replicas − consumedReplicas` chunks. `consumedReplicas` only goes up, even
when the workload's pods leave, so **used capacity is never replaced**. That is what
`refillStrategy: none` means. When `consumedReplicas` reaches `replicas`, the buffer is marked
`Fulfilled` and stays that way.

Without shrink, a partially-filled one-shot buffer would hold its *full* empty size **plus** the
nodes the workload already occupies — over-provisioning until the buffer is fully
filled or the deadline hits. Shrink-as-fill eliminates that: a half-filled buffer of N holds exactly
`N − consumed` empty chunks + the consumed capacity. (Contrast with `recreate`/`recreateUpToLimit`,
where the equivalent accounting is *non*-monotonic — freed capacity is recreated.)

### Snapshotting the target for mutable-target buffers

The number of unfilled chunks a one-shot buffer holds is `target − consumedReplicas`, where the
**target** is the total number of chunks the buffer is supposed to have. The buffer controller
reports the target in `status.replicas`. The target is *not* the number of virtual pods that exist
right now, because that number already goes down as the buffer fills. Subtracting
`consumedReplicas` from it would count the same consumption twice.

In v1 (`podTemplateRef` + `replicas`, no `limits`, `replicas` can't be changed; see Non-Goals),
the target is always `spec.replicas` and will never change during the buffer's life. Consumption doesn't
lower it. The reference implementation reads the target from `status.replicas` and never lets
`consumedReplicas` go above it. Because the target never changes, a re-reconcile or a controller
restart computes the same remainder.

Any shape whose target is **derived and can change** — `percentage` of a `scalableRef`, or a
`limits`-capped target, or an edited `replicas` — is out of v1 scope precisely because the resolved
target could drift mid-cycle and corrupt the `target − consumed` remainder (e.g. an 8→6 target with
`consumed=2` would emit `6−2=4`, double-counting consumption; a restart could latch `Fulfilled`
against the wrong target). To bring those shapes in scope, the consumer MUST **persist the resolved
target snapshot** when the one-shot cycle begins and compute the remainder, the clamp, and the
`Fulfilled` check against that snapshot. **Snapshot lifecycle:** the snapshot is taken once per
cycle; a one-shot buffer never re-arms, so **recreating (delete + create) the `CapacityBuffer`
starts a new cycle** with a fresh snapshot. v1 avoids the problem by
fixing the target (rejecting changes to `limits` and target) rather than persisting a snapshot.

## `fillDeadlineSeconds`

A one-shot buffer that is never filled would hold provisioning intent forever. An optional
`fillDeadlineSeconds` (integer seconds — per Kubernetes API convention, cf. `activeDeadlineSeconds`,
rather than a `metav1.Duration`, which the kube-api-linter forbids) enforces this: if the buffer has
not been filled within the deadline (measured from `status.fillStartTime`, when it became ready for
provisioning), it stops trying — the autoscaler **deletes the buffer's virtual pods** so no further
capacity is provisioned for it. Note this does **not** forcibly delete already-provisioned
capacity; nodes created for the buffer are reclaimed only when they are
empty/underutilized, via normal consolidation. The buffer is marked terminal with the **`Expired`**
condition (reason `FillDeadlineExceeded`) — distinct from `Fulfilled`, so consumers can tell a
timed-out buffer from a filled one (see [Status](#status-terminal-conditions-fulfilled-vs-expired)).

Before setting `Expired`, the autoscaler MUST credit every matching pod it has seen bind before the
deadline, so a pod that bound in time but hasn't been credited yet (for example, because the
watch event is still being processed) still counts. If that crediting fills the buffer, the buffer
becomes `Fulfilled` instead. Once `Expired` is set it is final: a pod that bound before the deadline
but that the autoscaler first sees afterward (because its cache was behind) will not be credited.

`fillDeadlineSeconds` is meaningful only for `refillStrategy: none` (a recreate/capped buffer has no
"unfilled forever" state to bound).

## Status: terminal conditions (`Fulfilled` vs `Expired`)

A one-shot buffer is terminal once it reaches one of **two mutually-exclusive** conditions
(standard `metav1.Condition`s on the existing `Conditions` field — no status-type change). Success
and give-up are kept as *distinct* conditions so a consumer that inspects only condition status can
tell them apart:

* `Fulfilled=True, reason=BufferFilled` — **success**: the buffer's capacity was consumed by the
  matching workload; it will not recreate.
* `Expired=True, reason=FillDeadlineExceeded` — **give-up**: the deadline elapsed before the buffer
  filled; provisioning intent was withdrawn. The buffer was **not** filled.

Both are terminal (the buffer emits zero virtual pods thereafter); `IsTerminal()` is
"`Fulfilled` OR `Expired`". They never both become True.

## Measuring consumption

A buffer names the pods that fill it with `matchingPodSelector` (a `metav1.LabelSelector`, scoped
to the buffer's namespace). A matching pod is **credited** to a buffer when it is **bound**
(`spec.nodeName` set, no scheduling gates, past the readiness boundary below). Crediting creates a
`CapacityBufferBinding` (see [below](#capacitybufferbinding)) that records the pod, the buffer, and
the pod's resource requests. `consumedReplicas` is the number of whole chunks covered by the summed
requests of the buffer's bindings, and never decreases.

**Bound-ness** is the signal deliberately: it is a universal lifecycle fact produced by every
scheduler via the Binding subresource, whereas scheduler-specific conditions (e.g.
`PodScheduled=Unschedulable`) are inconsistent across Volcano/YuniKorn/Kueue.

### Counting is cumulative

Every credited pod counts, including a pod that later finishes or is deleted, and including a pod
that replaces one evicted by consolidation, a node failure, a spot interruption, or a rollout. A
replacement is a new pod with a new UID; Kubernetes does not record which pod it replaces, so it
cannot be told apart from new work without per-workload-kind rules, which would not be
scheduler-agnostic.

This is a deliberate trade-off. Cumulative counting can over-count when pods are replaced while
the buffer is filling, so a buffer can shrink or reach `Fulfilled` early and the last pods of a
workload may wait for a new node (a cold start). The alternative considered — counting the peak
concurrent footprint of live pods (see [Alternatives](#peak-concurrent-consumption)) — never
over-counts, but a workload whose pods finish before the rest arrive never fills the buffer, which
then holds idle capacity until `fillDeadlineSeconds`. For expensive capacity such as GPU instances,
an occasional cold start is preferable to paying for idle nodes, and cumulative counting never
holds capacity longer than the peak-concurrent alternative would.

**Sizing guidance.** Size a one-shot buffer to the workload's concurrent footprint (e.g. a Job's
`parallelism`, a gang's member count), not its total number of completions.

### Multi-resource chunk accounting (deterministic rule)

To count consumed chunks, the autoscaler divides the total requests of the bound pods by one chunk's
requests. Requests cover several resources (CPU, memory, GPUs and other extended resources), and
each resource can give a different count. For example, the pods' CPU might add up to 2 chunks
while their memory adds up to only 1. Every autoscaler MUST use this rule so their counts agree:

```text
consumedChunks = min over each resource r in perChunk, with perChunk[r] > 0, of
                     floor( boundSum[r] / perChunk[r] )
                 clamped to [0, target]   // target = resolved replica count (see below)
```

* **MIN, not MAX** — a chunk is a whole pod-shape and requires *all* its dimensions; the binding
  dimension is the scarcest, so the count is the minimum ratio. (MAX would over-count and latch
  `Fulfilled` too early — a point raised and withdrawn in review.)
* **`floor`** — partial chunks do not count.
* **Zero / missing `perChunk[r]`** — dimensions whose per-chunk request is zero (or absent) are
  skipped, not divided by. If *no* dimension has a positive per-chunk request, the result is 0.
* **Missing `boundSum[r]`** — a resource the chunk requests but the bound pods don't, is counted as
  zero, so the result is 0. For example, pods that request no GPU never fill a GPU-sized chunk.
* **Clamp to `replicas`** — a workload larger than the buffer cannot consume more than the buffer's
  target (see [snapshotting the target](#snapshotting-the-target-for-mutable-target-buffers)).

### Deterministic fill order (overlapping selectors)

A bound pod MUST advance/apply-to at most one buffer's `consumedReplicas`; otherwise a pod matched by two
buffers double-counts. Two buffers matching `animal=cow` and `sound=moo` respectively are disjoint until a pod carrying *both* labels appears — and requiring users to enumerate every label so selectors stay disjoint is unusable (two workloads that share some labels but differ in others may legitimately want to share a buffer). Overlap is therefore **allowed**, and instead the pod→buffer assignment is made deterministic at measurement time.

**Rule.** Each bound pod counts toward only one buffer. To pick it, the autoscaler looks at the
buffers in the pod's namespace that match the pod and are **still filling**, meaning they aren't
`Fulfilled` or `Expired` and have used fewer chunks than their target. It sorts them in the order
below and gives the pod to the first one. The other matching buffers ignore the pod. Every
autoscaler MUST use the same order:

1. **Refill strategy** — `none` (one-shot) before `recreateUpToLimit` (capped) before `recreate`.
2. **Age** — older `creationTimestamp` first.
3. **Name** — lexicographic `namespace/name` as the final tiebreak (`creationTimestamp` has
   one-second resolution, so it alone is not a total order).

**Why one-shot fills first.** The buffers that can reach a terminal state must be credited before the
ones that cannot. A `recreate` buffer refills whatever is consumed regardless of who is credited;
if it were credited for a pod that also matched a one-shot buffer, the `recreate` buffer would
re-provision that capacity anyway *and* the one-shot buffer would never fill — it would hold its
full empty size until `fillDeadlineSeconds`, defeating its completion semantics and double-provisioning.
Crediting the latching buffer first is the only assignment under which `refillStrategy: none` works
in the presence of overlap. The same argument places capped (bounded, can latch) before `recreate`
(unbounded, never latches), and it makes the latching buffers' credit independent of a steady-state
buffer's lifecycle — deleting a `recreate` or capped buffer never changes what a one-shot buffer has
already been credited.

**Stickiness.** Ownership is decided once, when the pod's `CapacityBufferBinding` is created, and
never changes: a pod that already has a binding is never re-evaluated. A pod counted toward a buffer
that then latches terminal **stays counted there** and is not re-credited to a lower-priority
buffer. Because the binding is named by the pod's UID, this is enforced by the API server rather
than by every consumer agreeing on a recomputation, and it holds across restarts.

**Interaction with capped buffers.** This order is defined here for all three refill values so the
capped buffers proposal inherits it rather than inventing a second one; it is the shared exclusive
ownership rule for the `matchingPodSelector` field.

### Readiness boundary

Only pods that became scheduled **at or after** the buffer's fill window, count toward its
fill — a pod counts when its `PodScheduled` condition `lastTransitionTime` is not
before the buffer's `status.fillStartTime` (its first `ReadyForProvisioning` transition, written
once so a readiness flap does not move it), falling back to the pod's creation
time when that condition is absent; when neither is known the pod is counted rather than
under-reporting real capacity. This prevents pods that were already bound when the buffer became
ready from consuming it. It is an *observable* rule,
stronger than merely assuming the buffer is created before its workload. A buffer that does not pass
this boundary for a pod is not a candidate in the fill order above.

## CapacityBufferBinding

A namespaced resource in `autoscaling.x-k8s.io` that records one pod's credit toward one buffer:

```yaml
apiVersion: autoscaling.x-k8s.io/v1beta1
kind: CapacityBufferBinding
metadata:
  name: <pod-uid>                     # must equal spec.podRef.uid
  namespace: <pod's namespace>
  labels:
    autoscaling.x-k8s.io/capacity-buffer-uid: <buffer-uid>
  ownerReferences:
  - {apiVersion: autoscaling.x-k8s.io/v1beta1, kind: CapacityBuffer, name: gang-warmup, uid: <buffer-uid>}
spec:                                 # immutable
  bufferRef: {name: gang-warmup, uid: <buffer-uid>}
  podRef:    {name: trainer-0, uid: <pod-uid>}
  boundTime: "2026-11-02T20:40:00Z"   # the pod's PodScheduled transition
  resources: {cpu: "8", memory: 64Gi, nvidia.com/gpu: "1"}   # the pod's requests when credited
```

* **One buffer per pod.** The name is the pod's UID (enforced by a CEL rule), so a second binding
  for the same pod fails with `AlreadyExists`, whichever buffer or consumer attempts it. Consumers
  treat `AlreadyExists` as "already credited".
* **Write-once.** The spec is immutable; a binding is created once and never updated.
* **Owned by the buffer, not the pod.** A binding outlives its pod, so a pod that ran and completed
  still counts, and all of a buffer's bindings are garbage-collected when the buffer is deleted
  (which also starts a new one-shot cycle cleanly, because the label and `bufferRef` carry the UID).
* **Rebuildable.** `consumedReplicas` can be recomputed from a buffer's bindings at any time, so an
  autoscaler restart, or a different autoscaler taking over, loses nothing. Two rules keep a
  recompute from lowering it: the autoscaler recomputes only after its informer cache has synced,
  and it writes the larger of the stored value and the recomputed value.
* **Observable.** `kubectl get capacitybufferbindings -l autoscaling.x-k8s.io/capacity-buffer-uid=<uid>`
  lists the pods that filled a buffer; `kubectl get capacitybufferbinding <pod-uid>` shows which
  buffer a pod filled.

Consumers create bindings as soon as they observe a matching pod bound (an event-driven watch, not a
periodic scan), so short-lived pods are credited as long as they are observed bound at all. A pod
that binds and is deleted entirely while the consumer is down is not credited.

### Control-plane cost

A binding costs one create (~800 bytes) per credited pod, and nothing for pods that match no
filling buffer. Consumers read bindings from their informer cache and need no live API reads: when
several buffers compete for a pod, a consumer serializes its own crediting and remembers the
bindings it created until its cache observes them. In the reference implementation at 200 matching
pods, binding creates were 0.15–0.25% of all API requests, binding create p99 latency was ≤ 50 ms,
and API Priority and Fairness rejected nothing.

## Interaction with the autoscaler's scheduling simulation

A buffer's placeholder (virtual) pods exist only in the autoscaler's scheduling simulation; real
schedulers never see them. Two rules keep the simulation from provisioning extra nodes while the
workload binds onto the buffer's capacity:

* **Pending matching pods fill placeholders.** A matching pod that is pending is
  about to bind onto the buffer's capacity. For the simulation only, each chunk that pending 
  matching pods will fill (assigned in fill order, counted with the same
  [chunk rule](#multi-resource-chunk-accounting-deterministic-rule) as bound pods, up to the
  buffer's remaining placeholders) drops one placeholder, so the pending pods are scheduled **into** the
  buffer's capacity instead of on top of it. Nothing is written; pods are credited only when they
  bind. Gated pods (e.g. not yet admitted by Kueue) do not count, because they cannot be scheduled
  yet. The autoscaler doesn't check whether a pending pod fits on the buffer's nodes: a matching pod
  uses up a chunk when it binds wherever it lands, so its placeholder is dropped either way.
* **Placeholders match the cluster snapshot.** When building a scheduling pass, the consumer credits
  every matching pod already bound in its cluster-state snapshot before counting placeholders, so
  the snapshot and the placeholder set agree on which chunks are taken.

Without these, a workload binding in a burst (a gang, a rollout) is double-counted — once as real
pods, once as the placeholders for the chunks those pods are taking — and the autoscaler launches
nodes that consolidation later removes.

## Interaction with gang scheduling

The one-shot semantics interact with *how* the workload binds:

* **True gang scheduler** (Volcano, YuniKorn gang, coscheduling): binding is all-or-nothing — no pod
  binds until all can be placed. An undersized buffer therefore does not partially fill or terminate
  early; the scheduler waits until total cluster capacity suffices, then binds atomically.
  Undersizing costs pre-warming benefit, not correctness.
* **Incremental binding** (e.g. Kueue plain quota admission → kube-scheduler): pods bind one at a
  time, so an undersized one-shot buffer can be considered "filled" after only part of the workload
  binds. For strictly all-or-nothing jobs this is a pre-existing property of quota-only admission
  (independent of the buffer). Recommendation: pair one-shot buffers with a gang scheduler for
  all-or-nothing workloads, and size the buffer to the workload.

Note that **shrink-as-fill removes the transient over-provisioning** that incremental binding would
otherwise cause: as each pod binds, the buffer drops the corresponding chunk (`consumedReplicas`
rises), so a correctly-sized buffer holds `replicas − consumed` empty chunks throughout the ramp
rather than its full size plus the occupied nodes.

# User Stories

## Story 1 — Kueue-admitted training gang

A user submits an 8-pod all-or-nothing training Job via Kueue and creates a one-shot buffer sized to
it so the gang schedules immediately on admission rather than cold-starting. Once the gang is
scheduled the buffer is `Fulfilled` and stops; when the job completes the nodes are reclaimed and the
buffer does not refill.

```yaml
apiVersion: autoscaling.x-k8s.io/v1beta1
kind: CapacityBuffer
metadata:
  name: gang-warmup
  namespace: my-namespace
spec:
  # provisioningStrategy left default (active-capacity); the one-shot behavior is the refill axis:
  refillStrategy: none          # do not recreate consumed capacity (one-shot)
  fillDeadlineSeconds: 900      # 15m: give up + withdraw provisioning intent if never filled
  podTemplateRef:
    name: gang-shape            # matches the gang's pod shape
  replicas: 8
  matchingPodSelector:          # the gang's pods fill the buffer as they bind
    matchLabels:
      job-name: llm-train
```

## Story 2 — one-shot standby capacity (composition)

Because refill behavior is orthogonal to capacity kind, a one-shot buffer can use a
provider-specific capacity type. E.g. a GKE `standby-capacity` buffer (suspended nodes) that is also
one-shot: `provisioningStrategy: buffer.gke.io/standby-capacity` + `refillStrategy: none`. This
composition is the main reason to model one-shot on the refill axis rather than as its own
provisioning strategy.

# Design Details

## Proposed API

Adds three `CapacityBufferSpec` fields (the refill axis, the deadline, and the selector), two
`CapacityBufferStatus` fields (the shrink high-water mark and the fill-window start), and the
`CapacityBufferBinding` resource:

The new fields carry the same wire metadata as the existing `CapacityBuffer` fields — JSON tags,
sequential protobuf field numbers, and kubebuilder validation — so generated clients and every
consumer serialize and interpret them identically. Protobuf numbers continue the existing sequence
(spec 1–6 and status 1–5 are taken by the base API); this proposal claims **spec 7–9** and
**status 6–7**. Spec 7–8 and status 6–7 match the Karpenter reference implementation; spec 9
(`matchingPodSelector`) replaces its interim annotation.

```go
// +kubebuilder:validation:XValidation:rule="!has(self.refillStrategy) || self.refillStrategy != 'none' || has(self.matchingPodSelector)",message="matchingPodSelector is required when refillStrategy is none"
type CapacityBufferSpec struct {
    // ... existing fields: provisioningStrategy (1), podTemplateRef (2), scalableRef (3),
    // replicas (4), percentage (5), limits (6) ...

    // refillStrategy controls what happens to buffer capacity consumed by real workloads.
    // "recreate" maintains the buffer size; "recreateUpToLimit" is the capped behavior;
    // "none" is one-shot (consumed capacity is not recreated).
    // An omitted/empty value MUST be treated as "recreate" by consumers — i.e. objects created
    // before this field existed keep today's behavior, so this is a backward-compatible addition.
    // (Field name and values are under discussion; alt. name onCapacityConsumption.)
    // +optional
    // +kubebuilder:validation:Enum=recreate;recreateUpToLimit;none
    // +kubebuilder:default=recreate
    RefillStrategy *string `json:"refillStrategy,omitempty" protobuf:"bytes,7,opt,name=refillStrategy"`
    // NOTE ON THE ENUM: the enum above lists all three shared values. This proposal defines only
    // "none"; "recreateUpToLimit" is defined by the capped-buffers proposal. If the two proposals
    // land separately, the shared schema may accept a value that a given autoscaler doesn't
    // implement yet (Karpenter's reference impl implements recreate and none). An autoscaler that
    // sees a refillStrategy it doesn't implement MUST NOT provision for the buffer (rather than
    // silently treating it as another value) and SHOULD report it with a Warning event, reason
    // UnsupportedRefillStrategy. It does not change ReadyForProvisioning, which belongs to the
    // buffer controller.
    // The default marker uses the +kubebuilder:default form to match the existing
    // CapacityBuffer fields (e.g. provisioningStrategy). Consumers that do not server-side default
    // (e.g. Karpenter, which treats a nil value as "recreate" in code) need no schema default.

    // fillDeadlineSeconds bounds, in seconds, how long a one-shot (refillStrategy=none) buffer
    // keeps trying to provision if it is never filled. On expiry the autoscaler withdraws
    // provisioning intent (deletes the buffer's virtual pods) and marks the buffer Expired;
    // already-provisioned nodes are reclaimed only by normal consolidation when
    // empty/underutilized. Applies only to refillStrategy=none. Integer-seconds (not
    // metav1.Duration) per Kubernetes API convention (cf. activeDeadlineSeconds) — the
    // kube-api-linter forbids Duration fields.
    // +optional
    // +kubebuilder:validation:Minimum=1
    FillDeadlineSeconds *int32 `json:"fillDeadlineSeconds,omitempty" protobuf:"varint,8,opt,name=fillDeadlineSeconds"`

    // matchingPodSelector selects the pods, in the buffer's namespace, that fill the buffer when
    // they bind. Required for refillStrategy=none (enforced by the CEL rule on the spec).
    // Intended to be shared with capped buffers.
    // +optional
    MatchingPodSelector *metav1.LabelSelector `json:"matchingPodSelector,omitempty" protobuf:"bytes,9,opt,name=matchingPodSelector"`
}

type CapacityBufferStatus struct {
    // ... existing fields: podTemplateRef (1), replicas (2), podTemplateGeneration (3),
    // conditions (4), provisioningStrategy (5) ...

    // consumedReplicas is the number of whole chunks covered by the buffer's
    // CapacityBufferBindings; it never decreases. The consumer emits replicas-consumedReplicas
    // chunks (shrink-as-fill) and never recreates consumed capacity; when it reaches replicas the
    // buffer becomes Fulfilled. Set by the consuming autoscaler.
    // +optional
    // +kubebuilder:validation:Minimum=0
    ConsumedReplicas *int32 `json:"consumedReplicas,omitempty" protobuf:"varint,6,opt,name=consumedReplicas"`

    // fillStartTime is when the buffer's fill window opened (its first ReadyForProvisioning
    // transition). Written once, so a transient loss of readiness neither restarts the fill
    // deadline nor changes which pods count. The readiness boundary and fillDeadlineSeconds are
    // measured from it.
    // +optional
    FillStartTime *metav1.Time `json:"fillStartTime,omitempty" protobuf:"bytes,7,opt,name=fillStartTime"`
}

// CapacityBufferBinding records that one bound pod was credited toward one CapacityBuffer.
// Its name is the pod's UID; it lives in the pod's namespace and is owned by the buffer.
// +kubebuilder:resource:path=capacitybufferbindings,scope=Namespaced,shortName=cbb
// +kubebuilder:validation:XValidation:rule="self.metadata.name == self.spec.podRef.uid"
type CapacityBufferBinding struct {
    metav1.TypeMeta   `json:",inline"`
    metav1.ObjectMeta `json:"metadata,omitempty"`
    // +required
    Spec CapacityBufferBindingSpec `json:"spec"`
}

// +kubebuilder:validation:XValidation:rule="self == oldSelf",message="spec is immutable"
type CapacityBufferBindingSpec struct {
    BufferRef BindingObjectRef    `json:"bufferRef"` // the credited buffer
    PodRef    BindingObjectRef    `json:"podRef"`    // the credited pod
    BoundTime metav1.Time         `json:"boundTime"` // the pod's PodScheduled transition
    Resources corev1.ResourceList `json:"resources"` // the pod's requests when credited
}

type BindingObjectRef struct {
    Name string `json:"name"`
    UID  string `json:"uid"`
}
```

Because these fields and the `CapacityBufferBinding` resource are additions to the shared
`autoscaling.x-k8s.io` API, they must land in the schema before a consumer can implement against
them. Existing objects omit `refillStrategy` and so default to `recreate`, preserving current
behavior. Consumers need RBAC to create, get, list, and watch `capacitybufferbindings`. The
Karpenter reference implementation uses an interim `karpenter.sh/buffer-match-selector` annotation
for the selector until the spec field exists, then migrates.

## Responsibility split (controller vs. autoscaler)

Consistent with the base proposal: the **buffer controller** resolves pod spec / replicas and sets
`ReadyForProvisioning`, strategy-agnostic. The **autoscaler (consumer)** honors `refillStrategy`:
for `none`, it creates `CapacityBufferBinding`s as matching pods bind, derives `consumedReplicas`
from them, does not recreate consumed capacity, sets `Fulfilled` when filled (or `Expired` when
`fillDeadlineSeconds` elapses unfilled), and stops contributing virtual pods for a terminal buffer.

Both writers update the same `CapacityBufferStatus`, and each owns its own fields:

| Writer | Fields |
| --- | --- |
| Buffer controller | `podTemplateRef`, `replicas`, `podTemplateGeneration`, `ReadyForProvisioning` condition |
| Autoscaler | `consumedReplicas`, `fillStartTime`, `Fulfilled` and `Expired` conditions |

Each writer MUST change only its own fields, using a status merge patch with an optimistic lock
(`resourceVersion` precondition). On a conflict it re-reads the buffer and retries, so neither
writer overwrites the other's fields from a stale copy. Server-side apply with a separate field
manager per writer is an acceptable alternative. The Karpenter reference implementation uses the
merge patch with an optimistic lock for all of its status writes.

## Why a separate `refillStrategy` field (not a `provisioningStrategy` value)

* `provisioningStrategy` describes the **kind** of capacity (active, GKE standby, provider-specific).
  Refill behavior is a **different question** — what to do when that capacity is consumed — and
  applies to every kind. Encoding "one-shot" as a provisioning strategy would prevent, e.g., a
  one-shot standby buffer.
* It **unifies** the three behaviors (recreate / recreate-up-to-limit / none) on one axis instead of
  scattering them, and gives capped buffers and ephemeral buffers a shared home.

# Alternatives

## A new `provisioningStrategy` value (`buffer.x-k8s.io/ephemeral-capacity`)

The original form of this proposal. Rejected in review: conflates capacity *kind* with refill
*behavior* and cannot compose one-shot with other capacity kinds (e.g. GKE standby). The
refill-strategy axis supersedes it.

## Annotations instead of spec fields

The interim Karpenter implementation uses `karpenter.sh/*` annotations. Rejected as the end state:
opaque, unvalidated, and consumer-specific — the shared contract should be typed spec fields.

## Capped buffers alone

Capped buffers are steady-state (recreate-up-to-limit): they refill after the workload leaves, which
batch does not want. One-shot (`none`) is the missing value on the same axis. Note the tradeoff is
real and not one-directional:

* **Static NodePool** — for stable demand, a fixed floor of always-on nodes is simpler than any
  refilling buffer.
* **Recreate / capped (repeating)** — best for a *bursty-but-frequent* stream of jobs: capacity
  stays warm between runs. A controller can create one-shot buffers per job, which removes authoring
  latency but not provisioning latency, so a warm repeating buffer still wins for dense streams.
* **One-shot (`none`)** — best for occasional / one-off / cost-bounded runs, where paying for
  capacity between runs is not worth it.

Users pick per workload; one-shot is additive, not a replacement.

## Other ways to track which pods filled a buffer

* **Pod UIDs in the buffer's status.** Simple, but status grows with the workload: buffers of
  thousands of replicas hit etcd's object-size limit and turn every bind into a rewrite of a large
  object. Rejected in review.
* **Only a consumed count, plus the timestamp of the last pod processed.** Constant size, but not
  level-based: status cannot be rebuilt from the cluster, a restart can double-count or miss pods,
  and users cannot see which pods filled a buffer. Rejected in review.
* **Counting only pods that exist now** (as capped buffers does). Rebuildable, but a pod that
  completed or was deleted stops counting, so a one-shot buffer could refill. Unsuitable for
  one-shot semantics.
* **Labeling each credited pod with its buffer.** Durable and observable, but the autoscaler would
  write to user workload pods, contending with kubelet status updates.
* **EndpointSlice-style shards** (a slice object per buffer holding up to N pod entries). Fewer
  objects than one binding per pod, and rebuildable. But nothing in the API server stops the same
  pod appearing in two buffers' slices, so exclusive ownership depends on every consumer agreeing,
  and slice packing adds controller complexity. A possible evolution if per-pod object counts become
  a concern; the per-pod binding is preferred for v1.

## Peak-concurrent consumption

Count a buffer's consumption as the peak concurrent footprint of credited pods that are still live
(bound, not terminating, not finished), rather than cumulatively. This never counts a replaced pod
twice and is scheduler-agnostic. It was not chosen because a workload whose pods finish before the
rest arrive never fills the buffer, which then holds idle capacity until `fillDeadlineSeconds`; for
expensive capacity, that cost is worse than an occasional cold start. See
[Counting is cumulative](#counting-is-cumulative).

## Do nothing

Batch users work around the gap today with balloon pods or by over-provisioning NodePools, with the
maintenance and cost drawbacks the base proposal already documents.

# Open Questions

1. **Field name and values:** `refillStrategy` (`recreate` / `recreateUpToLimit` / `none`) vs.
   `onCapacityConsumption` (`recreate` / `recreateUpToLimit` / `doNothing`). To be settled with the
   capped buffers proposal and @jbtk so all three behaviors share one field.
2. **Ownership of the default:** does `refillStrategy` default to `recreate` (preserving current
   behavior), and is the current behavior retconned as `recreate` explicitly?
3. **`fillDeadlineSeconds` scope and clock:** confirm it applies only to `none`, and that the deadline is
   measured from `status.fillStartTime` (the first `ReadyForProvisioning` transition). On expiry:
   withdraw provisioning intent (delete virtual pods) and mark `Expired/FillDeadlineExceeded` (distinct from `Fulfilled`), without force-deleting
   already-provisioned nodes.
4. **`Fulfilled` stickiness:** terminal for the object's lifetime (re-arm by delete/recreate) vs.
   re-armable by spec edit.
5. **Selector ownership:** this proposal now defines `matchingPodSelector` because capped buffers is
   on hold. Confirm the field name and that capped buffers will reuse it.
6. **Fill-order priority:** the order `none` → `recreateUpToLimit` → `recreate` (then age, then name)
   is proposed here per review discussion. Confirm with the capped buffers work that capped is
   comfortable being credited *after* one-shot when both match a pod.
7. **Cumulative vs. peak-concurrent counting:** this proposal chooses cumulative counting (replacement
   pods count again) to favor releasing capacity over holding it idle. Confirm, or make it
   configurable.
8. **Binding cleanup for terminal buffers:** bindings live as long as their buffer, and terminal
   buffers are not garbage-collected (a non-goal). Is a TTL for terminal buffers, and therefore
   their bindings, needed in v1?
9. **Should capped buffers use bindings?** Capped buffers count only pods that exist now, so they do
   not need durable credit; but sharing the binding would give them the same exclusive ownership
   and observability.
