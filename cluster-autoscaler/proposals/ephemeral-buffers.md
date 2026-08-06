# Ephemeral CapacityBuffers (workload-filled, one-shot)

#### Author: jicowan

A follow-up to the [CapacityBuffer API proposal](./buffers.md) (#8151). It defines a shared
`buffer.x-k8s.io/ephemeral-capacity` provisioning-strategy contract; a reference implementation
exists in Karpenter.

# Timeline

## Alpha

- [ ] Agree on the `buffer.x-k8s.io/ephemeral-capacity` strategy semantics (this proposal)
- [ ] Reference implementation in a buffer-consuming autoscaler (Karpenter — implementation ready)
- [ ] Document the workload-fill + latch contract for other consumers (CA, GKE CCC)

## Beta / V1 graduation criteria

- [ ] Second OSS implementation of the strategy (e.g. cluster-autoscaler)
- [ ] E2e coverage of the fill → latch → no-refill lifecycle
- [ ] In use for at least one full version behind the CapacityBuffer feature gate

# Summary

The `CapacityBuffer` API (autoscaling.x-k8s.io) today defines *steady-state* spare capacity: the
`active-capacity` provisioning strategy continuously maintains a buffer and refills it as workloads
consume it. This proposal adds a second, **one-shot** provisioning strategy,
`buffer.x-k8s.io/ephemeral-capacity`, for **gang / batch workloads**: the buffer pre-provisions a
fixed amount of capacity, is *filled* by a matching workload as its pods are scheduled, and once
filled transitions to a terminal state and **stops provisioning** — it does not refill.

This is the follow-up already anticipated in the CapacityBuffer proposal
([`buffers.md`](./buffers.md)), which lists under "Out of scope, may be added as follow up
proposals":

> *"Buffer dedicated for workload: mark the buffer as dedicated for pods defined by PodSelector. It
> would count the pods matching the selector as filling in the buffer space. Example usecase: I
> want to keep buffer for scaling, but I know the usual maximum needed for my workload."*

This proposal specifies that "filled-by-workload" behavior and adds the one-shot / completion
semantics needed for batch. It builds on — rather than competes with — the in-review **capped
capacity buffers** proposal: ephemeral reuses capped's fill-by-selector accounting but adds a
*terminal* latch that capped deliberately omits. The two are complementary operating points
(repeating capped keeps capacity warm for a stream of jobs; one-shot ephemeral stops paying between
runs) — users pick per workload. See [Alternatives](#use-capped-capacity-buffers-instead).

# Motivation

The `active-capacity` strategy is the right shape for "keep N spare CPUs warm at all times." It is
the wrong shape for **gang-scheduled batch jobs** (ML training, HPC, CI/CD run-to-completion),
where the user wants to pre-provision capacity for *a specific run* so the job schedules promptly,
and then have provisioning **stop** once the job is placed — not a standing buffer that refills
after the job completes and keeps incurring cost.

With `active-capacity` there is no terminal state: a buffer sized for a gang job refills after the
gang finishes and provisions capacity indefinitely for a job that already ran. There is no way to
express "provision X once, for this workload, then stop."

The community has repeatedly asked for batch/Kueue-oriented capacity provisioning
([kubernetes-sigs/karpenter#2571](https://github.com/kubernetes-sigs/karpenter/issues/2571),
[#742](https://github.com/kubernetes-sigs/karpenter/issues/742),
[#3138](https://github.com/kubernetes-sigs/karpenter/issues/3138)), and Kueue's cold-start problem
(admit → pods pend → autoscaler reacts → nodes appear late) is the canonical case. Kueue drives
Cluster Autoscaler's `ProvisioningRequest` for exactly this, but that API is not universally
implemented across autoscalers; expressing the same intent as a CapacityBuffer *strategy* makes it
available to any buffer-consuming autoscaler.

## Goals

* Define a `buffer.x-k8s.io/ephemeral-capacity` provisioning strategy in the shared
  `buffer.x-k8s.io` namespace, with semantics any autoscaler can implement.
* Allow a buffer to be **filled by a matching workload** (pods selected by a label selector count
  as consuming the buffer as they are scheduled).
* Give the buffer **completion semantics**: once filled (or after an optional deadline), it is
  terminal — it stops provisioning and does not refill.
* Reuse the existing `CapacityBuffer` object, status-condition machinery, and consumer integration
  points — no new CRD.

## Non-Goals

* **Reserving/guaranteeing the capacity for specific pods.** Like `active-capacity`, the buffer
  makes capacity *exist*; it does not guarantee only the target workload lands on it. Exclusivity
  is achieved out-of-band (e.g. node taints) until the scheduler offers capacity reservation. This
  mirrors the base proposal's Non-Goal.
* **Gang pod scheduling.** This provisions *nodes*; all-or-nothing *binding* remains the job of a
  gang scheduler (Kueue, Volcano, YuniKorn, kube-scheduler coscheduling).
* **Coupling the buffer to a specific batch scheduler or workload CRD.** Fill is detected from
  generic pod state (see Design Details), not from any scheduler's objects.
* **Auto-deleting the buffer object.** A filled buffer becomes terminal but is not garbage
  collected; cleanup is left to the user or a future TTL follow-up.

# Proposal

A new value for the existing `spec.provisioningStrategy` field:
`buffer.x-k8s.io/ephemeral-capacity`. Because `provisioningStrategy` is already an open string in
the CapacityBuffer API, this adds a strategy without a schema change.

An ephemeral buffer behaves as follows:

1. **Provision once.** The buffer controller resolves the pod spec and replica count exactly as
   today; the autoscaler provisions capacity for the resulting virtual pods.
2. **Fill by matching workload.** Pods selected by a label selector, in the buffer's namespace,
   that are *scheduled* (bound to a node) count as consuming ("filling") the buffer's capacity.
3. **Latch and stop.** When filled matching capacity covers the buffer's intended capacity — or an
   optional fill deadline elapses — the buffer transitions to a terminal `Fulfilled` condition,
   stops contributing capacity, and its now-surplus nodes are reclaimed by normal consolidation.
   It does **not** refill if the workload later leaves.

This differs from the base proposal's "buffer dedicated for workload" follow-up in one way: that
sketch is steady-state (count matching pods as filling, but keep maintaining the buffer). The
ephemeral strategy adds the **terminal / one-shot** semantics that batch requires.

### Selecting the matching workload

The pods that fill the buffer are identified by a label selector scoped to the buffer's own
namespace (cross-namespace matching is intentionally excluded to avoid accidental overlap on common
labels). An optional deadline bounds how long an unfilled buffer holds capacity before giving up.
See Design Details for the concrete surface.

### Why "scheduled" (bound) and not a scheduler-specific signal?

Fill must be detected without depending on any particular gang scheduler. A pod being **bound to a
node** (`spec.nodeName` set) is a universal lifecycle fact produced by every scheduler via the
Binding subresource. Scheduler-specific status (e.g. the `PodScheduled=Unschedulable` condition) is
inconsistent across Volcano/YuniKorn/Kueue and unsuitable as a portable signal. See Design Details.

## Out of scope, may be added as follow up proposals

* **Hold-until-Running:** latch on pods reaching `Running` rather than merely bound (adds image-pull
  latency; bound is the earlier, cheaper signal).
* **Native reservation/exclusivity:** once the scheduler offers capacity reservation, replace the
  taint-based exclusivity workaround.
* **TTL / auto-deletion** of terminal ephemeral buffers.
* **First-class spec fields** for the selector/deadline (initially expressed via annotations by
  consumers that prefer to avoid a shared-schema change).

## User Stories

### Story 1 — Kueue-admitted training gang

A user submits an 8-pod all-or-nothing training Job via Kueue. They create an ephemeral buffer
sized to the gang so that, on admission, the pods schedule immediately onto pre-warmed nodes
instead of waiting for a cold-start scale-up. Once the gang is scheduled the buffer latches
`Fulfilled` and stops provisioning; when the job completes, the nodes are reclaimed and the buffer
does not refill.

```yaml
apiVersion: autoscaling.x-k8s.io/v1alpha1
kind: CapacityBuffer
metadata:
  name: gang-warmup
  namespace: my-namespace
spec:
  provisioningStrategy: buffer.x-k8s.io/ephemeral-capacity
  podTemplateRef:
    name: gang-shape          # matches the gang's pod shape
  replicas: 8
  # matching-workload selector + optional deadline: see Design Details
```

### Story 2 — Periodic batch run

An admin runs a nightly batch job needing N nodes for its duration. An ephemeral buffer
pre-provisions the capacity for that run and is filled by the job's pods; after the run the capacity
is reclaimed. Because the buffer is one-shot, it does not re-provision between runs (a fresh buffer,
or a scheduled buffer follow-up, is used for the next run).

# Design Details

## Proposed API

No changes to the `CapacityBufferSpec` / `CapacityBufferStatus` Go types are required. This proposal
adds:

1. A **provisioning-strategy constant**:

   ```go
   // Provisions capacity once, is filled by a matching workload, then stops
   // provisioning permanently (does not refill). For gang/batch workloads.
   const EphemeralProvisioningStrategy = "buffer.x-k8s.io/ephemeral-capacity"
   ```

2. A **terminal status condition** `Fulfilled` (a standard `metav1.Condition`; the existing
   `Conditions []metav1.Condition` field carries it, so no type change):
   - `Fulfilled=True, reason=BufferFilled` — matching scheduled capacity covered the buffer.
   - `Fulfilled=True, reason=FillDeadlineExceeded` — the optional deadline elapsed unfilled.

3. **Matching-workload configuration.** Two inputs are required: a label selector for the filling
   pods, and an optional fill deadline. This proposal recommends these be expressible **without a
   shared-schema change initially** (e.g. as well-known annotations on the CapacityBuffer), with
   promotion to first-class spec fields (`spec.matchingPodSelector`, `spec.fillDeadline`) as a
   follow-up once semantics are settled. Consumers MUST agree on the annotation keys; a reference
   set (used by the Karpenter implementation) is `buffer.x-k8s.io/match-selector` and
   `buffer.x-k8s.io/fill-deadline`. **Open question: annotations vs. spec fields for v1.**

## Responsibility split (controller vs. autoscaler)

Consistent with the base proposal:

* **Buffer controller** — resolves the pod spec / replica count and sets `ReadyForProvisioning`
  exactly as for `active-capacity`. It is strategy-agnostic.
* **Autoscaler (buffer consumer)** — recognizes the `ephemeral-capacity` strategy and implements the
  fill/latch behavior: sums matching *bound* pod capacity, compares to the buffer's intended
  capacity, sets `Fulfilled`, and stops contributing capacity for a fulfilled buffer.

The **fill signal is `spec.nodeName != ""`** on matching pods (with any scheduling gates cleared),
summed as resource requests and compared against `replicas × pod-spec requests`. Comparison is
capacity-based across all requested resource dimensions (cpu, memory, extended/GPU), not
pod-count-based, so a workload whose pod shape differs from the template still fills correctly on a
resource basis.

## Conditions

Adds to the base proposal's condition set:

* `Fulfilled` — set by the consuming autoscaler for `ephemeral-capacity` buffers; terminal. Once
  true, the buffer contributes no further capacity and does not refill.

## Interaction with gang scheduling (important)

The fill/latch semantics interact with *how* the workload is bound:

* **True gang scheduler** (Volcano, YuniKorn gang, kube-scheduler coscheduling): binding is
  all-or-nothing. An undersized buffer causes the scheduler to bind nothing until total cluster
  capacity suffices; the buffer therefore does not latch prematurely and the workload only runs once
  all capacity is present. Undersizing costs pre-warming benefit, not correctness.
* **Incremental binding** (e.g. Kueue plain quota admission → kube-scheduler): pods bind one at a
  time; an undersized buffer can latch after only part of the workload binds. For strictly
  all-or-nothing jobs this is a pre-existing property of quota-only admission (independent of the
  buffer). Recommendation: pair ephemeral buffers with a gang scheduler for all-or-nothing
  workloads, and set the buffer's pod template + replicas to the workload's true shape and count.

## Scalability

* **New API calls?** No new object types. The consumer lists pods (already required for scheduling)
  and patches CapacityBuffer status (already done for `active-capacity`).
* **New API types?** No.
* **Object growth?** One additional status condition per ephemeral buffer.
* **Steady-state cost?** *Lower* than `active-capacity`: a fulfilled ephemeral buffer stops
  producing virtual pods, so it drops out of the scheduling simulation entirely once latched.

# Alternatives

## Extend the base "buffer dedicated for workload" follow-up as steady-state only

The base proposal's sketch counts matching pods as filling but keeps maintaining the buffer. That
does not serve batch: it never stops. Ephemeral adds the terminal semantics. (This proposal can be
seen as making that follow-up concrete *and* adding a one-shot mode.)

## Use capped capacity buffers instead

The **capped capacity buffers** proposal (in review in `kubernetes/autoscaler`) is the closest
existing work: it adds a `MatchingPodSelector` / `MatchingTargetNodeSelector` and "shrink as
filled" accounting, where matching pods are deducted from the buffer's intended size and the
autoscaler provisions only the remainder. Ephemeral reuses that same fill-accounting idea — so the
two are complementary, not competing, and an ephemeral implementation can build directly on capped
buffers' selector + `IntendedSize` machinery.

Capped buffers are explicitly **steady-state / repeating**. From the capped proposal: *"Once the
workload ends (pods are removed) the buffer will maintain empty capacity again."* Its lifecycle is
`empty(X) → filled → empty(X) → …` forever. Ephemeral's is `empty(X) → filled → done`. Neither is
universally "right" — they serve different operating points, and the choice is a real tradeoff:

* **Capped (repeating) is better for a bursty-but-frequent stream of gang jobs.** If you have many
  gang jobs queued back-to-back, a repeating buffer keeps capacity warm between them: as one job
  finishes and the buffer empties, it re-provisions so the *next* job schedules immediately. With a
  one-shot buffer, each job needs a fresh CapacityBuffer created and provisioned before it can use
  pre-warmed capacity, reintroducing cold-start latency (and node churn) for every job after the
  first. Note the per-job buffer need not be created by a human — a controller (e.g. one watching a
  job queue or a Kueue Workload) can mint an ephemeral buffer per job, which removes the *authoring*
  latency but not the *provisioning* latency: the new nodes still have to come up between jobs. So
  automation narrows the gap but doesn't close it; a repeating buffer that stays warm still wins on
  latency for a dense stream.
* **Ephemeral (one-shot) is better when you don't want to pay between jobs.** For an occasional or
  one-off run — or when idle capacity between runs is not worth its cost — a repeating buffer keeps
  provisioning after the job completes, spending on capacity for a job that already ran. One-shot
  stops, and the nodes are reclaimed.

Note that repeating buffers occupy a *middle* band, and it is a genuinely narrow one. If the job
stream is frequent and steady enough that you want capacity always available, the churn-free answer
is not a buffer at all — it is a **static NodePool** (a fixed floor of always-on nodes). A buffer
that perpetually refills to a constant size is a more complex, less direct way to hold stable
capacity than simply declaring that capacity. So repeating buffers are the right tool specifically
for **bursty-but-frequent** streams — enough churn that a static floor would waste capacity at the
troughs, but frequent enough that per-job one-shot provisioning would pay cold-start too often.
Outside that band the choice collapses cleanly: **static NodePool** for stable demand, **ephemeral
buffer** for occasional / cost-bounded runs.

So the gap capped leaves is not a flaw to fix but a **mode it deliberately omits**: capped has no
way to express "provision once for this run, then stop." Ephemeral supplies that terminal mode. The
delta over capped is small and additive: a `provisioningStrategy` value plus a one-shot `Fulfilled`
condition ("once filled, stop and do not refill"). If the capped proposal merges first, ephemeral
becomes a thin strategy on top of its selector/`IntendedSize` accounting; if not, ephemeral can
define the minimal fill surface it needs directly. Either way, ephemeral is not an alternative
*design* to capped — it is the complementary one-shot mode, and users pick per workload:
**repeating for a warm pipeline of jobs, ephemeral for cost-bounded or one-off runs.**

## Use ProvisioningRequest instead

Cluster Autoscaler's `ProvisioningRequest` (atomic, one-shot, Kueue-integrated) is the closest
existing primitive. It is not implemented by all autoscalers (e.g. Karpenter has no support, and
CA's `best-effort-atomic-scale-up` is unimplemented on some cloud providers), so expressing the
intent as a portable CapacityBuffer *strategy* reaches more of the ecosystem. The two can coexist.

## A separate CRD

Rejected — reuses zero of the CapacityBuffer machinery and fragments the API. `provisioningStrategy`
is designed for exactly this kind of extension.

# Open Questions

1. **Config surface:** annotations (no shared-schema change) vs. first-class spec fields
   (`matchingPodSelector`, `fillDeadline`) for the matching workload + deadline. If annotations,
   agree on well-known keys across consumers.
2. **`Fulfilled` stickiness:** terminal for the object's lifetime (re-arm by delete/recreate) vs.
   re-armable by spec edit.
3. **Undersized-buffer signal:** should a consumer emit a warning when observed matching pods imply
   the workload is larger than the buffer's intended capacity (primarily relevant under incremental
   binding)? The Karpenter reference implementation includes a detailed analysis of this case.
4. **Relationship to a future "scheduled buffer" follow-up** for recurring batch (buffer active only
   during certain windows) — is ephemeral composable with it, or an alternative?
