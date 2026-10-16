# CapacityBuffer Conditions and Shared API Constants

#### Author: Joshua Guo (joshuakguo)

# Summary

This proposal defines the conditions every implementation of `CapacityBuffer` shares:

*   `Ready` is new. It is the top-level condition for whether a buffer is valid and can be served as configured. It doesn't change as capacity is used and refilled, so administrators can alert on it ([#10080](https://github.com/kubernetes/autoscaler/issues/10080)).
*   `Provisioned` is new, and says whether all of the capacity the buffer asks for is ready: on nodes that are initialized and ready to run pods, as defined by its provisioning strategy ([#10127](https://github.com/kubernetes/autoscaler/issues/10127)). It changes as capacity is used and refilled.

`ReadyForProvisioning` and `Provisioning` aren't part of the shared spec. Cluster Autoscaler keeps setting them. Karpenter deprecates them and stops setting them in v1, which is a breaking change from beta to v1 for its readers. The proposal also adds a shared `constants.go` next to the CRD with the shared condition types, the shared reasons, and the default provisioning strategy.

# Motivation

Buffer status has two gaps.

No shared condition says whether a buffer is valid. Cluster Autoscaler and Karpenter both set `ReadyForProvisioning`, but it reports internal state: the buffer controller's handoff to the node autoscaler. Administrators and tools that check object health look for `Ready` ([#10080](https://github.com/kubernetes/autoscaler/issues/10080)).

No condition says whether the buffer's capacity is there ([#10127](https://github.com/kubernetes/autoscaler/issues/10127)). `Provisioning` means different things in each implementation: in Cluster Autoscaler, the autoscaler is acting on the buffer; in Karpenter, the replicas fit on existing nodes, including uninitialized ones. `status.readyReplicas` ([#9238](https://github.com/kubernetes/autoscaler/pull/9238), [#10342](https://github.com/kubernetes/autoscaler/pull/10342)) is a count that consumers have to compare to `status.replicas` themselves.

The `CapacityBuffer` API is meant to be independent of the cloud provider and the node autoscaler, but its conditions are defined separately in each node autoscaler. Consumers such as Kueue should be able to read `Ready` and `Provisioned` the same way on Cluster Autoscaler and Karpenter.

## Goals

*   Define `Ready` and `Provisioned` with the same meaning across implementations
*   Keep capacity out of `Ready`, so it stays stable
*   Add `Ready` and `Provisioned` without changing what any existing condition means
*   Add a shared place for `CapacityBuffer` API constants
*   Agree on how implementations decide whether a new condition should be shared

## Non-Goals

*   A shared repository for the whole `CapacityBuffer` API ([#9857](https://github.com/kubernetes/autoscaler/issues/9857))
*   A shared meaning for `ReadyForProvisioning` or `Provisioning`, or replacing `status.readyReplicas`
*   Reserving capacity for a specific workload ([#9774](https://github.com/kubernetes/autoscaler/issues/9774))
*   Standardizing implementation-specific condition reasons, or events such as `TriggeredScaleUp` and `NotTriggerScaleUp`
*   New provisioning strategies
*   Changing the default `kubectl get capacitybuffers` output ([#10384](https://github.com/kubernetes/autoscaler/issues/10384))

# Proposal

Some conditions are shared across implementations, and the rest stay specific to one implementation. A shared condition has one meaning, defined here and in `constants.go`, so consumers can read it the same way on any implementation. An implementation-specific condition means what its implementation says it means, so consumers shouldn't rely on it across implementations.

The proposed shared conditions are:

*   `Ready`: the buffer is valid and can be served as configured. See [Ready Condition](#ready-condition).
*   `Provisioned`: all of the capacity the buffer asks for is ready. See [Provisioned Condition](#provisioned-condition).
*   `LimitedByQuotas`: `status.replicas` was lowered to fit the namespace's `ResourceQuota`s.

Implementations can keep their own conditions alongside these, such as Cluster Autoscaler's `ReadyForProvisioning` and `Provisioning`. See [Implementation-Specific Conditions](#implementation-specific-conditions).

## Ready Condition

`Ready` is the top-level condition for whether a buffer is valid and can be served as configured. Its meaning and reasons are shared, but each implementation decides which checks it runs. For example, Karpenter can also check whether any instance type can satisfy the buffer's constraints.

*   **True:** the buffer is valid, and the implementation serving it can act on it.
*   **False:** the implementation found a problem with the buffer. The reason says what kind of problem, and the message gives the details.

When `Ready` is False, the implementation sets the first reason that applies:

| Order | Reason | When |
|---|---|---|
| 1 | `ResolutionFailed` | The pod template or replica count can't be resolved, for example because the pod template, the scalable object, or the kind in `spec.scalableRef` doesn't exist. |
| 2 | `Unsatisfiable` | The implementation knows that no capacity it can create satisfies the buffer's constraints. |

Implementations map their own checks onto these reasons, and put the specific cause in the message. For example, Cluster Autoscaler's `scalable-ref-watched` with `UnknownKind` becomes `ResolutionFailed`. An implementation that can't tell whether a buffer is satisfiable never uses `Unsatisfiable`, and those buffers report `Provisioned=False` with `CapacityNotReady` instead.

`metav1.Condition` reasons can't carry a domain prefix, so implementations don't add their own reasons for False. A new reason goes through the same process as a new condition, described in [Implementation-Specific Conditions](#implementation-specific-conditions).

`Ready` reflects the current spec and the objects it references, so it becomes False if, for example, the referenced workload is deleted. It carries `observedGeneration`, so consumers can tell when the latest spec hasn't been evaluated yet. A buffer can be `Ready` before any of its capacity exists. Temporary problems, such as a stockout, are reported by `Provisioned`. See [Monitoring Buffers](#monitoring-buffers) for which condition to wait on or alert on.

An implementation sets `Ready` and `Provisioned` only on buffers whose provisioning strategy it supports, so a buffer without `Ready` hasn't been picked up by any implementation. This assumes at most one implementation serves each strategy in a cluster.

## Provisioned Condition

`Provisioned` is set by the node autoscaler, which knows whether capacity exists. It sets it the first time it sees a buffer whose strategy it supports, even one that isn't `Ready`. Its presence shows that a node autoscaler has picked the buffer up.

*   **True:** the buffer is `Ready`, not `LimitedByQuotas`, and every replica in `status.replicas` fits on ready capacity. A buffer with no replicas is `Provisioned`.
*   **False:** otherwise. The reason says which case applies.

When `Provisioned` is False, the node autoscaler sets the first reason that applies. The order puts the reason that most needs attention first, so a problem that blocks the buffer comes before progress.

| Order | Reason | When |
|---|---|---|
| 1 | `BufferNotReady` | `Ready` is not True. |
| 2 | `ExceedsResourceQuotas` | `LimitedByQuotas` is True. |
| 3 | `CapacityNotReady` | Some replicas don't fit on ready capacity, and the node autoscaler doesn't know that capacity has been requested for all of them. |
| 4 | `ProvisioningInProgress` | Some replicas don't fit on ready capacity yet, and capacity has been requested for all of them. |

For example, workloads use all 10 replicas of a buffer, and the node autoscaler can only get capacity for 5 of them because a node group is at its maximum size. The buffer reports `CapacityNotReady` the whole time, including while those 5 are being provisioned, so the limit stays visible.

`CapacityNotReady` is also the fallback when the node autoscaler can't tell whether capacity has been requested, so consumers can't conclude from it that no capacity is coming. `ProvisioningInProgress` isn't a promise: requested capacity can still fail to arrive. Implementation-specific causes, such as a stockout or a node group at its maximum size, go in the message.

Ready capacity is defined by the provisioning strategy, and matches what `status.readyReplicas` counts. For `buffer.x-k8s.io/active-capacity`, it is nodes that are initialized and ready to run pods. Nodes that have launched or registered but aren't initialized yet don't count, such as Cluster Autoscaler's upcoming nodes or Karpenter's NodeClaims that aren't `Initialized`. A buffer whose missing replicas are all on such nodes reports `ProvisioningInProgress`. A strategy that keeps capacity on suspended nodes can count those nodes, even though Kubernetes reports them as not ready. The `status.readyReplicas` doc comment is updated to say this.

A buffer limited by quotas isn't `Provisioned`, even if all of its allowed replicas exist, because it has less capacity than its spec asks for. These are the namespace's `ResourceQuota`s. Karpenter doesn't apply them today, so it never sets `LimitedByQuotas`, and its buffers never report `ExceedsResourceQuotas`.

`Provisioned` doesn't reserve capacity: any pod that fits can use it.

The node autoscaler writes `Provisioned` in the same update as `status.readyReplicas`, with the `resourceVersion` it read. The two never disagree, and neither is computed from an outdated `status.replicas` or `Ready`. `Provisioned` carries the `observedGeneration` of the `Ready` condition it was evaluated against.

As of each node autoscaler update, `Provisioned` is True exactly when `Ready` is True, `LimitedByQuotas` isn't, and `status.readyReplicas` equals `status.replicas`. Between updates, `Provisioned` can lag a change to its inputs, such as a higher `status.replicas` after the referenced workload scales. A consumer that acts once on `Provisioned=True` can confirm it against the same inputs, and check that `Provisioned`'s `observedGeneration` matches `metadata.generation`.

`Provisioned` is rewritten only when its status, reason, or `observedGeneration` changes, or along with a change to `status.readyReplicas`, so a stable buffer causes no writes. The message can include counts, such as `6 of 10 replicas fit on ready capacity`.

## Implementation-Specific Conditions

Conditions outside the shared spec, such as `ReadyForProvisioning`, `Provisioning`, and Cluster Autoscaler's `cluster-autoscaler.kubernetes.io/scalable-ref-watched`, are up to each implementation. Cluster Autoscaler keeps setting `ReadyForProvisioning` and `Provisioning`. Karpenter stops setting them in v1, which is a breaking change from beta for anyone who reads them on Karpenter.

Before an implementation adds a new condition, it shares it with the other implementations, for example at SIG Autoscaling, to decide whether it should be shared or reflected in `Ready`.

## Shared Constants

A new `constants.go` is added to each versioned API package, starting with `cluster-autoscaler/apis/capacitybuffer/autoscaling.x-k8s.io/v1beta1`. It holds the shared condition types, the shared reasons, the default provisioning strategy, and later any shared annotations. Values that already exist, such as `ActiveProvisioningStrategy` and `LimitedByQuotas`, come from the [Cluster Autoscaler](https://github.com/kubernetes-sigs/cluster-autoscaler/blob/main/pkg/capacitybuffer/constants.go) and [Karpenter](https://github.com/kubernetes-sigs/karpenter/blob/main/pkg/apis/autoscaling/v1beta1/constants.go) constants.

The shared reasons cover False only. `metav1.Condition` requires a non-empty reason, so when `Ready` or `Provisioned` is True, implementations set one of their choosing, and consumers shouldn't rely on it. The proposed file is in the [Appendix](#appendix-proposed-constantsgo).

## Monitoring Buffers

Administrators can alert on:

*   `Ready` not True, or missing a few minutes after the buffer is created or changed: the buffer is broken or can't be satisfied, or no implementation serves its strategy.
*   `Provisioned=False` with `CapacityNotReady` for longer than a node takes to start: capacity can't be obtained.

`ProvisioningInProgress` is normal while a buffer refills.

To wait for a buffer, wait on the condition that matches what you need:

*   `kubectl wait --for=condition=Ready` returns once the buffer is valid, which can be before any capacity exists.
*   `kubectl wait --for=condition=Provisioned` returns once all of its capacity is ready.

Both conditions carry `observedGeneration`, so neither returns early for a spec change that hasn't been evaluated yet.

## Compatibility with Existing Implementations

No implementation sets `Ready` or `Provisioned` today. The closest existing signals are:

| Proposed | Cluster Autoscaler today | Karpenter today |
|---|---|---|
| `Ready=True` | `ReadyForProvisioning=True` | `ReadyForProvisioning=True` |
| `Ready=False` / `ResolutionFailed` | `ReadyForProvisioning=False` / `error`, or `scalable-ref-watched=False` / `UnknownKind` | `ReadyForProvisioning=False` / `PodTemplateNotFound`, `ScalableRefNotFound`, or `ResolutionFailed` |
| `Ready=False` / `Unsatisfiable` | None | None |
| `Provisioned=True` | `status.readyReplicas` equals `status.replicas` | `Provisioning=True` / `FitsExistingCapacity`, which also counts uninitialized nodes |
| `Provisioned=False` / `BufferNotReady` | `status.readyReplicas` drops to 0, and a buffer that was `Provisioning=True` gets `Provisioning=False` / `NotReadyForProvisioning` | `Provisioning=False` / `NotReadyForProvisioning` |
| `Provisioned=False` / `ExceedsResourceQuotas` | `LimitedByQuotas=True` | None; `LimitedByQuotas` is never set |
| `Provisioned=False` / `CapacityNotReady` | `status.readyReplicas` below `status.replicas`, with a `NotTriggerScaleUp` event when no scale-up was possible | `Provisioning=False` / `RequiresNewCapacity` |
| `Provisioned=False` / `ProvisioningInProgress` | A `TriggeredScaleUp` event | `Provisioning=True` / `FitsExistingCapacity` on uninitialized nodes |

Cluster Autoscaler only processes buffers with `ActiveProvisioningStrategy`, while Karpenter processes buffers of every strategy.

Both implementations need to change to match this proposal. The condition list in [`buffers.md`](buffers.md#what-conditions-will-be-specified-in-the-buffer-state) is updated to point here.

# Open Questions

*   **Should a buffer that no node can be created for always be `Ready=False` with `Unsatisfiable`?** Karpenter can report it through `Ready`, while another implementation might report `Provisioned=False` with `CapacityNotReady`, the same as a stockout. Implementations decide differently when no node fits, and the answer changes as node groups and NodePools change.
*   **How is `Provisioned` computed for [capped buffers](https://github.com/kubernetes/autoscaler/issues/9774)?** They are filled by workloads instead of refilled, so `status.readyReplicas` drops while the buffer works as intended. Pods filling the buffer can also lower `status.replicas` and set `LimitedByQuotas` through the namespace's `ResourceQuota`s. One option is `status.readyReplicas + consumed chunks >= target chunk count`, with `ExceedsResourceQuotas` based on the target instead of `status.replicas`. This depends on the capped buffers design.

# Alternatives

## Do nothing

Each implementation keeps its own conditions, and no `Ready` or `Provisioned` condition is set.

### Why not?

The implementations have already diverged on `Provisioning`, and changes in one repository aren't reviewed by the other's owners. Consumers such as Kueue would keep handling each implementation separately, with no shared condition for validity or for capacity.

## Use `ReadyForProvisioning` or `Accepted` instead of `Ready`

`ReadyForProvisioning` becomes the shared validity condition, or the new condition is called `Accepted`, as in the Gateway API ([GEP-1364](https://gateway-api.sigs.k8s.io/geps/gep-1364/)).

### Why not?

`ReadyForProvisioning` reports internal state: the handoff from the buffer controller to the node autoscaler. That handoff is how Cluster Autoscaler and Karpenter split their components, and another implementation might not have it. Administrators and tools that check object health look for `Ready`, which is what [#10080](https://github.com/kubernetes/autoscaler/issues/10080) asks for. `Ready` can also cover checks beyond the handoff, such as whether the buffer's constraints can be satisfied. `Provisioned` covers capacity, so this proposal states that `Ready` doesn't track it.

## Make `Ready` depend on current capacity

`Ready` is True only when the buffer is valid and `Provisioned`.

### Why not?

It would flap whenever workloads use capacity and the node autoscaler refills it, so alerts on it would fire for healthy buffers. Consumers that need all of the capacity can read `Provisioned`.

## Remove `ReadyForProvisioning` and `Provisioning` from every implementation

Every implementation stops setting them in v1, and `Ready` and `Provisioned` replace them.

### Why not?

Cluster Autoscaler's readers, including alerts, dashboards, and wrapper APIs, depend on them. Without a conversion webhook, removing a condition from one API version removes it from all of them. Because these conditions aren't part of the shared spec, each implementation decides for its own readers: Karpenter drops them in v1, and Cluster Autoscaler keeps them.

## Handle `Provisioned` differently on a buffer that isn't `Ready`

The node autoscaler keeps `Provisioned` at its last value when a buffer becomes invalid, or doesn't set `Provisioned` until the buffer is first `Ready`.

### Why not?

Keeping the last value lets `Provisioned=True` claim that a configuration that no longer exists is satisfied. Leaving it unset goes against the [API conventions](https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#typical-status-properties), which ask controllers to set their conditions the first time they see an object. Two invalid buffers would also report differently depending on their history.

## Leave `Ready` reasons to each implementation

Each implementation uses its own reasons for `Ready`, as for `ReadyForProvisioning` today.

### Why not?

Consumers and alerts couldn't tell an unresolvable buffer from an unsatisfiable one without knowing which implementation serves it. Shared reasons keep the kind of problem the same everywhere, and the message still carries the details.

## Use only `CapacityNotReady` when capacity is missing

One reason covers all missing capacity, and the message says whether it has been requested.

### Why not?

Consumers need to know whether to wait or to act, and messages aren't part of the contract. Every node autoscaler knows when it has requested capacity, such as Cluster Autoscaler's upcoming nodes and Karpenter's launched NodeClaims, so `ProvisioningInProgress` can be shared.

## Name the quota reason `LimitedByQuotas`

The quota reason reuses the condition's name.

### Why not?

"Quotas" is ambiguous next to cloud quotas, which are reported with `CapacityNotReady`. `ExceedsResourceQuotas` names Kubernetes `ResourceQuota`s.

## Use `status.readyReplicas` instead of a `Provisioned` condition

Consumers compare `status.readyReplicas` to `status.replicas`, as raised on [#10127](https://github.com/kubernetes/autoscaler/issues/10127).

### Why not?

`kubectl wait --for=jsonpath` can only compare a field to a fixed value, so it can't wait for one field to reach another, and a count has no reason. `Provisioned` complements `status.readyReplicas`.

## Put `Provisioning` in the shared constants

`Provisioning` gets one shared meaning, instead of a new `Provisioned` condition.

### Why not?

`Provisioning` names an activity, not an outcome. Consumers care whether capacity is ready, and a fully provisioned buffer isn't "provisioning" anything. Giving it one meaning would also change it for at least one implementation, breaking that implementation's existing readers.

## Document the shared semantics in the original proposal only

The definitions go in [`buffers.md`](buffers.md), and implementations keep their own constants.

### Why not?

Documents can't be imported, so the constants can still drift, and changes to them aren't reviewed next to the CRD.

# Appendix: Proposed `constants.go`

```go
// Constants shared across implementations of the CapacityBuffer API.
// ReadyForProvisioning and Provisioning are implementation-specific, and are
// intentionally not defined here.
const (
	// ActiveProvisioningStrategy is the default spec.provisioningStrategy. The node
	// autoscaler adds placeholder pods for the buffer to its scheduling simulation,
	// and scales up so the buffer's capacity stays available.
	ActiveProvisioningStrategy = "buffer.x-k8s.io/active-capacity"

	// ReadyCondition is the top-level condition for whether a buffer is valid and
	// can be served as configured. It is set, with observedGeneration, by the
	// implementation that supports the buffer's provisioning strategy, which decides
	// which checks it covers. It does not change as capacity is used and refilled;
	// see ProvisionedCondition.
	// True: the buffer is valid, and the implementation can act on it.
	// False: the implementation found a problem; the reason and message say which.
	ReadyCondition = "Ready"

	// ProvisionedCondition says whether the buffer's capacity is ready. It is set by
	// the node autoscaler the first time it sees a buffer whose provisioning
	// strategy it supports, with observedGeneration copied from ReadyCondition.
	// True: the buffer is Ready, not LimitedByQuotas, and every replica in
	// status.replicas fits on ready capacity, as defined by the provisioning
	// strategy and counted by status.readyReplicas. For ActiveProvisioningStrategy,
	// ready capacity is nodes that are initialized and ready to run pods.
	// False: otherwise; the reason says why.
	ProvisionedCondition = "Provisioned"

	// LimitedByQuotasCondition is set by the buffer controller. It is True when
	// status.replicas was lowered to fit the namespace's ResourceQuotas; the message
	// lists the limiting quotas.
	LimitedByQuotasCondition = "LimitedByQuotas"

	// Reasons for ReadyCondition=False, in order of precedence. The first one
	// that applies wins. The spec doesn't define reasons for True.

	// ResolutionFailedReason is the reason for Ready=False when the pod template
	// or replica count can't be resolved, including when a referenced object
	// doesn't exist.
	ResolutionFailedReason = "ResolutionFailed"

	// UnsatisfiableReason is the reason for Ready=False when the implementation
	// knows that no capacity it can create satisfies the buffer's constraints.
	UnsatisfiableReason = "Unsatisfiable"

	// Reasons for ProvisionedCondition=False, in order of precedence. The first
	// one that applies wins. The spec doesn't define reasons for True.

	// BufferNotReadyReason is the reason for Provisioned=False when the buffer is
	// not Ready.
	BufferNotReadyReason = "BufferNotReady"

	// ExceedsResourceQuotasReason is the reason for Provisioned=False when the
	// buffer is LimitedByQuotas.
	ExceedsResourceQuotasReason = "ExceedsResourceQuotas"

	// CapacityNotReadyReason is the reason for Provisioned=False when some replicas
	// don't fit on ready capacity, and the node autoscaler doesn't know that
	// capacity has been requested for all of them. It is also the fallback when
	// the node autoscaler can't tell.
	CapacityNotReadyReason = "CapacityNotReady"

	// ProvisioningInProgressReason is the reason for Provisioned=False when some
	// replicas don't fit on ready capacity yet, and capacity has been requested for
	// all of them. It doesn't promise that the capacity will arrive.
	ProvisioningInProgressReason = "ProvisioningInProgress"
)
```
