# CapacityBuffer Conditions and Shared API Constants

#### Author: Joshua Guo (joshuakguo)

# Summary

This proposal adds two conditions to `CapacityBuffer`:

*   `Accepted` says whether a buffer is valid: the implementation serving it supports its provisioning strategy, and the buffer passes its validity checks. It doesn't depend on capacity, so it doesn't flap, and administrators can alert on it ([#10080](https://github.com/kubernetes/autoscaler/issues/10080)). `ReadyForProvisioning` stays, and rolls up into `Accepted`.
*   `Provisioned` says whether all of the capacity the buffer asks for is ready, as defined by its provisioning strategy ([#10127](https://github.com/kubernetes/autoscaler/issues/10127)). It changes as capacity is used and refilled.

The names match `ProvisioningRequest`, in the same API group. The proposal also adds a shared `constants.go` next to the CRD with the condition types, the shared reasons, and the default provisioning strategy.

# Motivation

Buffer status has two gaps.

No top-level condition says whether a buffer is valid. `ReadyForProvisioning` reports the buffer controller's handoff to the node autoscaler, but administrators catching a bad rollout want a condition about the buffer itself ([#10080](https://github.com/kubernetes/autoscaler/issues/10080)). The `CapacityBufferStatus.Conditions` doc comment adds to the confusion by describing a `Ready` condition, in terms of capacity, that no implementation sets. That comment is updated to match this proposal. The new condition also has to be stable: capacity is consumed by design, so a condition that followed it would flap.

No condition says whether the buffer's capacity is there ([#10127](https://github.com/kubernetes/autoscaler/issues/10127)). `Provisioning` means different things in each implementation: in Cluster Autoscaler, the autoscaler is acting on the buffer; in Karpenter, the replicas fit on existing nodes, including uninitialized ones. `status.readyReplicas` ([#9238](https://github.com/kubernetes/autoscaler/pull/9238), [#10342](https://github.com/kubernetes/autoscaler/pull/10342)) is a count that consumers have to compare to `status.replicas` themselves.

The `CapacityBuffer` API is meant to be independent of the cloud provider and the node autoscaler, but its conditions are defined separately in each node autoscaler. Consumers such as Kueue should be able to read `Accepted` and `Provisioned` the same way on Cluster Autoscaler and Karpenter.

## Goals

*   Define `Accepted` and `Provisioned` with the same meaning across implementations
*   Keep capacity out of `Accepted`, so it stays stable
*   Add both without removing any existing condition or changing what it means
*   Add a shared place for `CapacityBuffer` API constants

## Non-Goals

*   A shared repository for the whole `CapacityBuffer` API ([#9857](https://github.com/kubernetes/autoscaler/issues/9857))
*   A `Ready` condition. See [Add a `Ready` condition](#add-a-ready-condition).
*   A shared meaning for `Provisioning`, or replacing `status.readyReplicas`
*   Reserving capacity for a specific workload ([#9774](https://github.com/kubernetes/autoscaler/issues/9774))
*   Standardizing implementation-specific condition reasons, or events such as `TriggeredScaleUp` and `NotTriggerScaleUp`
*   New provisioning strategies
*   Changing the default `kubectl get capacitybuffers` output ([#10384](https://github.com/kubernetes/autoscaler/issues/10384))

# Proposal

## Accepted Condition

`Accepted` is set by the buffer controller of the implementation that serves the buffer, such as Cluster Autoscaler or Karpenter. It rolls up the buffer controller's validity conditions: the shared `ReadyForProvisioning`, and any of the implementation's own conditions, prefixed with its domain, such as Cluster Autoscaler's `cluster-autoscaler.kubernetes.io/scalable-ref-watched` when the kind in `spec.scalableRef` isn't installed. Only conditions the buffer controller writes roll up, so it never waits on another controller. New shared inputs need their own proposal.

*   **True:** every validity condition that rolls up into `Accepted` is True, so a node autoscaler can act on the buffer. Today, that means the pod template and desired replica count have been resolved into `status.podTemplateRef` and `status.replicas`.
*   **False:** at least one of them isn't True. The reason and message say which check failed.

`Accepted` reflects the current spec and the objects it references, so it becomes False if, for example, the referenced workload is deleted. It carries `observedGeneration`, so consumers and `kubectl wait` can tell when the latest spec hasn't been evaluated yet.

`Accepted` doesn't depend on capacity, and only covers what the buffer controller can check. A buffer whose pod template can't run on any node, for example because of a typo in a node selector, stays `Accepted` and reports the problem through `Provisioned`. See [Monitoring Buffers](#monitoring-buffers).

An implementation sets `Accepted` and `Provisioned` only on buffers whose provisioning strategy it supports, so a buffer without `Accepted` hasn't been picked up by any implementation. This assumes at most one implementation serves each strategy in a cluster.

Node autoscalers provision capacity only for `Accepted` buffers, so validity checks added later also stop provisioning. A buffer without `Accepted`, for example right after an upgrade, is still provisioned if `ReadyForProvisioning` is True. This fallback only decides whether to provision: `Provisioned` is evaluated against `Accepted` alone, so it stays False with `BufferNotAccepted` until `Accepted` is set.

## ReadyForProvisioning Condition

`ReadyForProvisioning` keeps its meaning and stays in v1 as an input to `Accepted`. Existing readers keep working, but consumers should read `Accepted`, which also covers checks added later. While it is the only shared input, the buffer controller sets `Accepted` alongside it with the same status, reason, and message. Both carry `observedGeneration`, which Cluster Autoscaler doesn't set on `ReadyForProvisioning` today.

## Provisioned Condition

`Provisioned` is set by the node autoscaler, which knows whether capacity exists. It sets it the first time it sees a buffer whose strategy it supports, even one that isn't `Accepted`, so its presence shows that a node autoscaler has picked the buffer up. Consumers that want to know whether capacity is ready read `Provisioned`, not `Provisioning`.

*   **True:** the buffer is `Accepted`, not `LimitedByQuotas`, and every replica in `status.replicas` fits on ready capacity.
*   **False:** otherwise. The reason says which case applies.

Ready capacity is defined by the provisioning strategy, and matches what `status.readyReplicas` counts. For `buffer.x-k8s.io/active-capacity`, it is existing nodes that are ready to run pods. A strategy that keeps capacity on suspended nodes can count those nodes, even though Kubernetes reports them as not ready.

A buffer limited by quotas isn't `Provisioned`, even if all of its allowed replicas exist, because it has less capacity than its spec asks for. These are the namespace's `ResourceQuota`s, which the buffer controller applies to `status.replicas`. An implementation that doesn't apply them, as Karpenter doesn't today, never sets `LimitedByQuotas`, so its buffers never report `ExceedsResourceQuotas`. A buffer with zero replicas by its own spec is `Provisioned` with `BufferEmpty`.

`Provisioned` carries the `observedGeneration` of the `Accepted` condition it was evaluated against, so `kubectl wait --for=condition=Provisioned` doesn't return early for a spec change that hasn't been evaluated yet.

`Provisioned` reports that capacity exists, not that it is held for anyone: any pod that fits can use it. It is also stricter than `ProvisioningRequest`'s `Provisioned`, which a `best-effort-atomic-scale-up` request can set before its nodes are ready ([#10430](https://github.com/kubernetes/autoscaler/issues/10430)).

## Shared Constants

A new `constants.go` is added to each versioned API package, starting with `cluster-autoscaler/apis/capacitybuffer/autoscaling.x-k8s.io/v1beta1`. It holds the shared condition types, the shared reasons, the default provisioning strategy, and later any shared annotations. The values come from the existing [Cluster Autoscaler](https://github.com/kubernetes-sigs/cluster-autoscaler/blob/main/pkg/capacitybuffer/constants.go) and [Karpenter](https://github.com/kubernetes-sigs/karpenter/blob/main/pkg/apis/autoscaling/v1beta1/constants.go) constants.

Only the `Provisioned` reasons are shared, because the reasons for `Accepted` and `ReadyForProvisioning` describe validation errors that differ by implementation. `Provisioning` isn't in `constants.go` at all; see [Provisioning Condition](#provisioning-condition). The proposed file is in the [Appendix](#appendix-proposed-v1beta1constantsgo).

# Design Details

## Provisioned Condition

| Status | Reason | When |
|---|---|---|
| True | `CapacityReady` | `Accepted` is True, `LimitedByQuotas` is not True, `status.replicas` is greater than 0, and every replica fits on ready capacity. |
| True | `BufferEmpty` | `Accepted` is True, `LimitedByQuotas` is not True, and `status.replicas` is 0. |
| False | `BufferNotAccepted` | `Accepted` is not True. |
| False | `ExceedsResourceQuotas` | `Accepted` is True, and `LimitedByQuotas` is True. |
| False | `ProvisioningInProgress` | `Accepted` is True, `LimitedByQuotas` is not True, some replicas don't fit on ready capacity yet, and the node autoscaler knows that capacity has been requested for all of them. |
| False | `CapacityNotReady` | `Accepted` is True, `LimitedByQuotas` is not True, some replicas don't fit on ready capacity yet, and the node autoscaler doesn't know that capacity has been requested for all of them. |

`CapacityNotReady` is the fallback, so consumers can't conclude from it that no capacity is coming. `ProvisioningInProgress` needs capacity to have been requested for every missing replica, such as nodes requested from the cloud provider that aren't ready yet, so stuck replicas aren't hidden behind ones on their way. It isn't a promise: requested capacity can still fail to arrive. A buffer that no node group or NodePool can run also reports `CapacityNotReady`. Implementation-specific causes, such as a stockout or a node group at its maximum size, go in the message.

Each condition has one writer: the buffer controller writes `Accepted`, `ReadyForProvisioning`, and `LimitedByQuotas`, and the node autoscaler writes `Provisioned`. The node autoscaler writes `Provisioned` in the same update as `status.readyReplicas`, with the `resourceVersion` it read, so the two never disagree, and neither is computed from an outdated `status.replicas` or `Accepted`. Cluster Autoscaler's forced server-side apply of `status.readyReplicas` switches to this update. As of each node autoscaler update, `Provisioned` is True exactly when `Accepted` is True, `LimitedByQuotas` isn't, and `status.readyReplicas` equals `status.replicas`. Between updates, `Provisioned` can lag a change the buffer controller makes, such as a higher `status.replicas` after the referenced workload scales. A consumer that acts once on `Provisioned=True` can confirm it against the same inputs: `Accepted` is True, `LimitedByQuotas` isn't, `status.readyReplicas` equals `status.replicas`, and `Provisioned`'s `observedGeneration` matches `metadata.generation`. The `status.readyReplicas` doc comment is updated to refer to ready capacity as defined by the strategy.

`Provisioned` is rewritten only when its status, reason, or `observedGeneration` changes, or along with a change to `status.readyReplicas`, so a stable buffer causes no writes. The message can include counts, such as `6 of 10 replicas fit on ready capacity`.

## Monitoring Buffers

Administrators can alert on:

*   `Accepted` not True, or missing a few minutes after the buffer is created or changed: the configuration is broken, or no implementation serves the strategy.
*   `Provisioned=False` with `CapacityNotReady` for longer than a node takes to start: capacity can't be obtained, including when no node fits the pod template.

`ProvisioningInProgress` is normal while a buffer refills.

## Provisioning Condition

`Provisioning` isn't changed or removed. Implementations that set it today keep setting it with their current meaning, so existing readers keep working.

It isn't part of the shared contract or `constants.go`, and stays an implementation-specific condition that each implementation defines in its own constants, as Cluster Autoscaler does today. Consumers shouldn't rely on it across implementations, and new implementations don't need to set it. It keeps its unprefixed name for compatibility, while new implementation-specific conditions use a domain prefix.

## Compatibility with Existing Implementations

No implementation sets `Accepted` or `Provisioned` today. The closest existing signals are:

| Proposed | Cluster Autoscaler today | Karpenter today |
|---|---|---|
| `Accepted` | `ReadyForProvisioning` | `ReadyForProvisioning` |
| `Provisioned=True` / `CapacityReady` | `status.readyReplicas` equals `status.replicas` | `Provisioning=True` / `FitsExistingCapacity`, which also counts uninitialized nodes |
| `Provisioned=True` / `BufferEmpty` | `status.readyReplicas` and `status.replicas` are both 0 | None |
| `Provisioned=False` / `ProvisioningInProgress` | A `TriggeredScaleUp` event | `Provisioning=True` / `FitsExistingCapacity` on uninitialized nodes |
| `Provisioned=False` / `CapacityNotReady` | `status.readyReplicas` below `status.replicas`, with a `NotTriggerScaleUp` event when no scale-up was possible | `Provisioning=False` / `RequiresNewCapacity` |
| `Provisioned=False` / `ExceedsResourceQuotas` | `LimitedByQuotas=True` | None; `LimitedByQuotas` is never set |
| `Provisioned=False` / `BufferNotAccepted` | `status.readyReplicas` drops to 0, and a buffer that was `Provisioning=True` gets `Provisioning=False` / `NotReadyForProvisioning` | `Provisioning=False` / `NotReadyForProvisioning` |

Other existing definitions:

| Definition | Cluster Autoscaler today | Karpenter today |
|---|---|---|
| `Provisioning` (unchanged, not shared) | The autoscaler is acting on the buffer | The replicas fit on existing nodes |
| `LimitedByQuotas` | Set by the buffer controller | Defined, never set |
| `ActiveProvisioningStrategy` | Default; only these buffers are processed | Defined, unused; all buffers are processed |
| `cluster-autoscaler.kubernetes.io/scalable-ref-watched` | Set for custom `spec.scalableRef` kinds; may roll up into `Accepted` | Not set |

Both implementations need to change to match this proposal. The condition list in [`buffers.md`](buffers.md#what-conditions-will-be-specified-in-the-buffer-state) is updated to point here.

# Open Questions

*   **Should a buffer that no node can be created for get its own reason?** It reports `CapacityNotReady` today, the same as a stockout. A shared reason such as `NoCompatibleNodes` would separate a broken pod template from a temporary shortage, but Cluster Autoscaler and Karpenter decide that no node fits differently, and the answer changes as node groups and NodePools change.
*   **How is `Provisioned` computed for [capped buffers](https://github.com/kubernetes/autoscaler/issues/9774)?** They are filled by workloads instead of refilled, so `status.readyReplicas` drops while the buffer works as intended, and pods filling the buffer can lower `status.replicas` and set `LimitedByQuotas` through the namespace's `ResourceQuota`s. One option is `status.readyReplicas + consumed chunks >= target chunk count`, with `ExceedsResourceQuotas` and `BufferEmpty` based on the target instead of `status.replicas`. This depends on the capped buffers design.

# Alternatives

## Do nothing

Each implementation keeps its own conditions, and no `Accepted` or `Provisioned` is set.

### Why not?

The implementations have already diverged on `Provisioning`, and changes in one repository aren't reviewed by the other's owners. Consumers such as Kueue would keep handling each implementation separately, with no condition for validity or for capacity.

## Add a `Ready` condition

The buffer controller sets `Ready` instead of `Accepted`, as [#10080](https://github.com/kubernetes/autoscaler/issues/10080) asks, or `Ready` is True when the buffer is both `Accepted` and `Provisioned`.

### Why not?

On `Pod`, `Node`, and `NodeClaim`, `Ready=True` means fully operational. A `Ready` that means valid would be True before any capacity exists, so `kubectl wait --for=condition=Ready` would return immediately. The Gateway API dropped `Ready` for this reason in favor of `Accepted` ([GEP-1364](https://gateway-api.sigs.k8s.io/geps/gep-1364/)), and `ProvisioningRequest` uses `Accepted` and `Provisioned`. A `Ready` that includes capacity would flap whenever workloads use capacity and the node autoscaler refills it, so alerts on it would fire for healthy buffers.

## Keep `ReadyForProvisioning` as the only validity condition

No `Accepted` condition is added.

### Why not?

`ReadyForProvisioning` describes a handoff, not the buffer, and every validity check added later would need its own condition for consumers to learn. Cluster Autoscaler's `scalable-ref-watched` is already one.

## Rename `ReadyForProvisioning` to `Accepted`

`Accepted` replaces `ReadyForProvisioning`.

### Why not?

Every existing reader, including automation, alerts, and integrations, would have to migrate for no change in meaning.

## Roll node autoscaler checks into `Accepted`

`Accepted` is also False when no node can be created for the pod template, so it catches more broken rollouts.

### Why not?

The buffer controller would have to wait on the node autoscaler, and `Accepted` would change when node groups or NodePools change. Whether a node can be created is part of whether capacity can be obtained, which `Provisioned` reports.

## Keep `Provisioned`'s last value while the buffer is not `Accepted`

The node autoscaler leaves `Provisioned` alone when a buffer becomes invalid, since the old capacity still exists.

### Why not?

`Provisioned=True` would claim that a configuration that no longer exists is satisfied, and consumers would have to check `Accepted` before trusting it.

## Leave `Provisioned` unset until the buffer is first `Accepted`

`Provisioned=False` with `BufferNotAccepted` is set only on buffers that were `Accepted` before.

### Why not?

The [API conventions](https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#typical-status-properties) ask controllers to set their conditions the first time they see an object, and treat conditions as observations rather than state machines. Two invalid buffers would otherwise report differently depending on their history.

## Use only `CapacityNotReady` when capacity is missing

One reason covers all missing capacity, and the message says whether it has been requested.

### Why not?

Consumers need to know whether to wait or to act, and messages aren't part of the contract. `Provisioning` can't tell them, because it means something different in each implementation. Every node autoscaler knows when it has requested capacity, such as Cluster Autoscaler's upcoming nodes and Karpenter's launched NodeClaims, so `ProvisioningInProgress` can be shared.

## Name the quota reason `LimitedByQuotas`

The quota reason reuses the condition's name.

### Why not?

"Quotas" is ambiguous next to cloud quotas, which are reported with `CapacityNotReady`. `ExceedsResourceQuotas` names Kubernetes `ResourceQuota`s, and stays accurate for buffer types that compute the quota case differently from `LimitedByQuotas`.

## Use `status.readyReplicas` instead of a `Provisioned` condition

Consumers compare `status.readyReplicas` to `status.replicas`, as raised on [#10127](https://github.com/kubernetes/autoscaler/issues/10127).

### Why not?

`kubectl wait --for=jsonpath` can only compare a field to a fixed value, so it can't wait for one field to reach another, and a count has no reason. `Provisioned` complements `status.readyReplicas`.

## Put `Provisioning` in the shared constants

`Provisioning` gets one shared meaning, instead of a new `Provisioned` condition.

### Why not?

`Provisioning` names an activity, not an outcome. Consumers care whether capacity is ready, not whether a node autoscaler is working on it, and a fully provisioned buffer isn't "provisioning" anything. Giving it one meaning would also change it for at least one implementation, breaking that implementation's existing readers.

## Remove `Provisioning`

`Provisioned` replaces `Provisioning`, which implementations stop setting.

### Why not?

It is a breaking change. Existing alerts, automation, and integrations read `Provisioning`, and without a conversion webhook, removing it from one API version removes it from all of them. Keeping it costs nothing, since it isn't part of the shared contract.

## Document the shared semantics in the original proposal only

The definitions go in [`buffers.md`](buffers.md), and implementations keep their own constants.

### Why not?

Documents can't be imported, so the constants can still drift, and changes to them aren't reviewed next to the CRD.

# Appendix: Proposed `constants.go`

```go
// Constants shared across implementations of the CapacityBuffer API.
// The Provisioning condition is implementation-specific, and is intentionally
// not defined here.
const (
	// ActiveProvisioningStrategy is the default spec.provisioningStrategy. The node
	// autoscaler adds placeholder pods for the buffer to its scheduling simulation,
	// and scales up so the buffer's capacity stays available.
	ActiveProvisioningStrategy = "buffer.x-k8s.io/active-capacity"

	// AcceptedCondition says whether a buffer is valid. It is set, with
	// observedGeneration, by the buffer controller of the implementation that
	// supports the buffer's provisioning strategy, and rolls up
	// ReadyForProvisioningCondition and any implementation-specific validity
	// conditions. It does not depend on capacity.
	// True: every validity condition that rolls up into it is True, so a node
	// autoscaler can act on the buffer. Today, that means status.podTemplateRef
	// and status.replicas are resolved.
	// False: at least one of them isn't True; the reason and message say which.
	AcceptedCondition = "Accepted"

	// ProvisionedCondition says whether the buffer's capacity is ready. It is set by
	// the node autoscaler the first time it sees a buffer whose provisioning
	// strategy it supports, with observedGeneration copied from AcceptedCondition.
	// True: the buffer is Accepted, not LimitedByQuotas, and every replica in
	// status.replicas fits on ready capacity, as defined by the provisioning
	// strategy and counted by status.readyReplicas.
	// False: otherwise; the reason says why.
	ProvisionedCondition = "Provisioned"

	// ReadyForProvisioningCondition is set by the buffer controller, and rolls up
	// into AcceptedCondition. It is True when status.podTemplateRef and
	// status.replicas are resolved. To check whether a buffer is valid, read
	// AcceptedCondition.
	ReadyForProvisioningCondition = "ReadyForProvisioning"

	// LimitedByQuotasCondition is set by the buffer controller. It is True when
	// status.replicas was lowered to fit the namespace's ResourceQuotas; the message
	// lists the limiting quotas.
	LimitedByQuotasCondition = "LimitedByQuotas"

	// CapacityReadyReason is the reason for Provisioned=True when status.replicas
	// is greater than 0.
	CapacityReadyReason = "CapacityReady"

	// BufferEmptyReason is the reason for Provisioned=True when status.replicas is 0.
	BufferEmptyReason = "BufferEmpty"

	// BufferNotAcceptedReason is the reason for Provisioned=False when the buffer is
	// not Accepted.
	BufferNotAcceptedReason = "BufferNotAccepted"

	// ExceedsResourceQuotasReason is the reason for Provisioned=False when the
	// buffer is Accepted but LimitedByQuotas.
	ExceedsResourceQuotasReason = "ExceedsResourceQuotas"

	// ProvisioningInProgressReason is the reason for Provisioned=False when some
	// replicas don't fit on ready capacity yet, and capacity has been requested for
	// all of them. It doesn't promise that the capacity will arrive.
	ProvisioningInProgressReason = "ProvisioningInProgress"

	// CapacityNotReadyReason is the reason for Provisioned=False when some replicas
	// don't fit on ready capacity yet, and the node autoscaler doesn't know that
	// capacity has been requested for all of them. It is the fallback when capacity
	// is missing.
	CapacityNotReadyReason = "CapacityNotReady"
)
```
