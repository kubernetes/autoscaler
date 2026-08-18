# One-shot (ephemeral) CapacityBuffers via a refill strategy

#### Author: jicowan

A follow-up to the [CapacityBuffer API proposal](./buffers.md) (#8151). It adds one-shot
(ephemeral) buffers for gang / batch workloads, expressed as a value on an orthogonal
**refill-strategy** axis that also subsumes the current (recreate) and capped
(recreate-up-to-limit) behaviors — rather than as a new `provisioningStrategy` value. (This
framing reflects review discussion with @kkonrad and @jbtk; an earlier draft used a
`provisioningStrategy` value.)

# Timeline

## Alpha

- [ ] Agree on the refill-strategy field name/values (this proposal, coordinated with capped buffers)
- [ ] Reference implementation in a buffer-consuming autoscaler (Karpenter — implementation ready,
      currently keyed on an interim annotation; will adopt the spec field)
- [ ] Document the one-shot (no-refill) + fill-deadline contract for other consumers (CA, GKE CCC)

## Beta / V1 graduation criteria

- [ ] Second OSS implementation of the one-shot behavior
- [ ] E2e coverage of the fill → stop → no-refill lifecycle (and the fill-deadline path)
- [ ] In use for at least one full version behind the CapacityBuffer feature gate

# Summary

The `CapacityBuffer` API (autoscaling.x-k8s.io) today maintains *steady-state* spare capacity:
capacity consumed by real workloads is **recreated** to keep the buffer at its configured size. That
is the right behavior for keeping headroom warm, but the wrong behavior for **gang / batch
workloads**, which want capacity pre-provisioned for a specific run and then provisioning to
**stop** — not a buffer that refills after the run and keeps incurring cost.

This proposal makes the buffer's **refill behavior** an explicit, first-class axis of the
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
remainder and never recreates consumed capacity), and an optional **`fillDeadline`** bounding how
long an unfilled one-shot buffer keeps trying.

The *selector input* — how a buffer names which pods fill it (e.g. `matchingPodSelector`) — is
**deferred to the capped buffers proposal**, which is already defining that field; one-shot reuses
it. The *consumption result and shrink/latch semantics* built on it are defined here, since they are
what distinguishes one-shot from capped.

# Motivation

Community demand for batch/Kueue-oriented capacity provisioning is well established
([kubernetes-sigs/karpenter#2571](https://github.com/kubernetes-sigs/karpenter/issues/2571),
[#742](https://github.com/kubernetes-sigs/karpenter/issues/742),
[#3138](https://github.com/kubernetes-sigs/karpenter/issues/3138)). The canonical case is Kueue's
cold-start: a job is admitted, its pods pend, the autoscaler reacts, and nodes appear late. Kueue
drives Cluster Autoscaler's `ProvisioningRequest` for this, but that API is not implemented across
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
* Add an optional **`fillDeadline`** so an unfilled one-shot buffer does not hold provisioning intent
  indefinitely.
* Compose with the existing (recreate) and capped (recreate-up-to-limit) behaviors on the same axis,
  rather than introducing a parallel mechanism.

## Non-Goals

* **Defining the shared matching-workload *selector* surface.** The spec field that names which
  pods fill a buffer (e.g. `matchingPodSelector`) should be shared with the capped buffers proposal;
  this proposal defers that field to capped and, in the reference implementation, uses an interim
  annotation. Note this proposal *does* define the **consumption-tracking result** — the monotonic
  `consumedReplicas` high-water and the shrink/latch semantics built on it — since those are what
  distinguish one-shot from capped; only the *selector input* defers to capped.
* **Reserving/guaranteeing capacity for specific pods.** Exclusivity remains out-of-band (node
  taints) until the scheduler offers capacity reservation — same as the base proposal.
* **Gang pod scheduling.** This provisions *nodes*; all-or-nothing binding stays with the gang
  scheduler (Kueue, Volcano, YuniKorn, coscheduling).
* **Auto-deleting the buffer object.** A completed one-shot buffer is terminal but not garbage
  collected; cleanup is left to the user or a future TTL follow-up.

# Proposal

## The refill-strategy axis

Refill behavior becomes a spec field, orthogonal to `provisioningStrategy`. Naming is an open
question carried over from the PR thread — the two candidates are `refillStrategy` and
`onCapacityConsumption`; this document uses `refillStrategy` for concreteness, values TBD with
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

A one-shot buffer **shrinks as it fills**: as a matching workload's pods are scheduled onto the
buffer's capacity, the buffer provisions only the *unfilled remainder* rather than its full size.
This is tracked with a monotonically non-decreasing status field, `consumedReplicas` (a high-water
mark of consumed chunks): the consumer emits `replicas − consumedReplicas` chunks, and because
`consumedReplicas` never decreases, **consumed capacity is never recreated** — the defining property
of `refillStrategy: none`. When `consumedReplicas` reaches `replicas`, the buffer latches
`Fulfilled`.

Without shrink, a partially-filled one-shot buffer would hold its *full* empty size **plus** the
nodes the workload already occupies — transient over-provisioning until the buffer either fully
fills or the deadline hits. Shrink-as-fill eliminates that: a half-filled buffer of N holds exactly
`N − consumed` empty chunks + the consumed capacity. (Contrast with `recreate`/`recreateUpToLimit`,
where the equivalent accounting is *non*-monotonic — freed capacity is recreated.)

## `fillDeadlineSeconds`

A one-shot buffer that is never filled would otherwise hold provisioning intent forever. An optional
`fillDeadlineSeconds` (integer seconds — per Kubernetes API convention, cf. `activeDeadlineSeconds`,
rather than a `metav1.Duration`, which the kube-api-linter forbids) bounds this: if the buffer has
not been filled within the deadline (measured from when it became ready for provisioning), it stops
trying — the autoscaler **deletes the buffer's virtual pods** so no further capacity is provisioned
for it. Note this does **not** forcibly delete
already-provisioned capacity; nodes created for the buffer are reclaimed only when they are
empty/underutilized, via normal consolidation. The buffer is marked terminal with reason
`FillDeadlineExceeded`.

`fillDeadline` is meaningful only for `refillStrategy: none` (a recreate/capped buffer has no
"unfilled forever" state to bound).

## Status: the `Fulfilled` condition

A one-shot buffer sets a terminal `Fulfilled` condition (a standard `metav1.Condition` on the
existing `Conditions` field — no status-type change):

* `Fulfilled=True, reason=BufferFilled` — the buffer's capacity was consumed by the matching
  workload; it will not recreate.
* `Fulfilled=True, reason=FillDeadlineExceeded` — the deadline elapsed unfilled; provisioning intent
  was withdrawn.

## Measuring consumption

The *selector input* (which pods count) defers to capped buffers; the *measurement* into
`consumedReplicas` is defined here. A consumer sums the resource requests of matching pods that are
**bound** (`spec.nodeName` set, no scheduling gates), divides by the per-chunk requests to get whole
chunks consumed, and advances `consumedReplicas` to that value (never decreasing it — the monotonic
high-water). **Bound-ness** is the signal deliberately: it is a universal lifecycle fact produced by
every scheduler via the Binding subresource, whereas scheduler-specific conditions (e.g.
`PodScheduled=Unschedulable`) are inconsistent across Volcano/YuniKorn/Kueue. The shared spec field
for the selector is expected from capped buffers (`matchingPodSelector`); until it lands, the
reference implementation uses an interim `karpenter.sh/*` annotation.

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
apiVersion: autoscaling.x-k8s.io/v1alpha1
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
  # matchingPodSelector (how fill is measured) comes from the capped buffers proposal
```

## Story 2 — one-shot standby capacity (composition)

Because refill behavior is orthogonal to capacity kind, a one-shot buffer can use a
provider-specific capacity type. E.g. a GKE `standby-capacity` buffer (suspended nodes) that is also
one-shot: `provisioningStrategy: buffer.gke.io/standby-capacity` + `refillStrategy: none`. This
composition is the main reason to model one-shot on the refill axis rather than as its own
provisioning strategy.

# Design Details

## Proposed API

Adds two `CapacityBufferSpec` fields (the refill axis + deadline) and one `CapacityBufferStatus`
field (the shrink high-water); reuses the capped proposal's selector for fill measurement:

```go
type CapacityBufferSpec struct {
    // ... existing fields (provisioningStrategy, podTemplateRef, scalableRef,
    // replicas, percentage, limits) ...

    // refillStrategy controls what happens to buffer capacity consumed by real workloads.
    // "recreate" (default) maintains the buffer size; "recreateUpToLimit" is the capped
    // behavior; "none" is one-shot (consumed capacity is not recreated).
    // (Field name and values are under discussion; alt. name onCapacityConsumption.)
    // +optional
    RefillStrategy *string

    // fillDeadlineSeconds bounds, in seconds, how long a one-shot (refillStrategy=none) buffer
    // keeps trying to provision if it is never filled. On expiry the autoscaler withdraws
    // provisioning intent (deletes the buffer's virtual pods); already-provisioned nodes are
    // reclaimed only by normal consolidation when empty/underutilized. Applies only to
    // refillStrategy=none. Integer-seconds (not metav1.Duration) per Kubernetes API convention
    // (cf. activeDeadlineSeconds) — the kube-api-linter forbids Duration fields.
    // +optional
    FillDeadlineSeconds *int32

    // matchingPodSelector — expected to be introduced by the capped buffers proposal; used to
    // measure which pods fill the buffer. Referenced here, not defined here.
}

type CapacityBufferStatus struct {
    // ... existing fields (podTemplateRef, replicas, podTemplateGeneration,
    // conditions, provisioningStrategy) ...

    // consumedReplicas is a monotonically non-decreasing high-water mark of how many chunks a
    // matching workload has consumed. The consumer emits replicas-consumedReplicas chunks
    // (shrink-as-fill) and never recreates consumed capacity; when it reaches replicas the buffer
    // becomes Fulfilled. Set by the consuming autoscaler.
    // +optional
    ConsumedReplicas *int32
}
```

Because `refillStrategy`, `fillDeadline`, and `consumedReplicas` are additions to the shared
`autoscaling.x-k8s.io` CRD, they must land in the schema before a consumer can implement against
them. The Karpenter reference implementation uses interim `karpenter.sh/*` annotations for the
selector until the shared field exists, then migrates.

## Responsibility split (controller vs. autoscaler)

Consistent with the base proposal: the **buffer controller** resolves pod spec / replicas and sets
`ReadyForProvisioning`, strategy-agnostic. The **autoscaler (consumer)** honors `refillStrategy`:
for `none`, it does not recreate consumed capacity, sets `Fulfilled` when filled or when
`fillDeadline` elapses, and stops contributing virtual pods for a terminal buffer.

## Why an orthogonal refill axis (not a provisioningStrategy value)

This is the central revision from the original proposal, per PR review:

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

## Do nothing

Batch users work around the gap today with balloon pods or by over-provisioning NodePools, with the
maintenance and cost drawbacks the base proposal already documents.

# Open Questions

1. **Field name and values:** `refillStrategy` (`recreate` / `recreateUpToLimit` / `none`) vs.
   `onCapacityConsumption` (`recreate` / `recreateUpToLimit` / `doNothing`). To be settled with the
   capped buffers proposal and @jbtk so all three behaviors share one field.
2. **Ownership of the default:** does `refillStrategy` default to `recreate` (preserving current
   behavior), and is the current behavior retconned as `recreate` explicitly?
3. **`fillDeadline` scope and clock:** confirm it applies only to `none`, and that the deadline is
   measured from `ReadyForProvisioning`. On expiry: withdraw provisioning intent (delete virtual
   pods) and mark `Fulfilled/FillDeadlineExceeded`, without force-deleting already-provisioned nodes.
4. **`Fulfilled` stickiness:** terminal for the object's lifetime (re-arm by delete/recreate) vs.
   re-armable by spec edit.
5. **Dependency on consumption tracking:** this proposal assumes the capped buffers proposal lands
   the fill-measurement mechanism (`matchingPodSelector`). If capped stalls, does one-shot need a
   minimal fill surface of its own, or wait?
