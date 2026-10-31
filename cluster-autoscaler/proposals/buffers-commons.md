# CapacityBuffer Conditions and Shared API Constants

#### Author: Joshua Guo (joshuakguo)

# Summary

This proposal adds two conditions to `CapacityBuffer`:

*   `Ready` is the condition that cluster administrators and tools monitor ([#10080](https://github.com/kubernetes/autoscaler/issues/10080)). It is True when the buffer's configuration is valid and a node autoscaler can act on it. It replaces `ReadyForProvisioning`, which this proposal removes when the API graduates to v1. `Ready` does not depend on capacity, so it doesn't flap as the buffer's capacity is used and refilled.
*   `Provisioned` is True when the buffer is `Ready` and all of the capacity its spec asks for exists on nodes that are ready to run pods ([#10127](https://github.com/kubernetes/autoscaler/issues/10127)). It is expected to change as the buffer's capacity is used and refilled. It replaces `Provisioning`, which this proposal also removes when the API graduates to v1.

While making that change, this proposal also adds a shared `constants.go` file next to the CRD. It defines the condition types, the condition reasons that all implementations share, and the default provisioning strategy, so every node autoscaler implementation can import the same definitions.

# Motivation

The conditions a buffer reports today leave two gaps.

The first is that buffers have no `Ready` condition. Cluster administrators commonly build monitoring and alerts on the `Ready` condition of Kubernetes objects, for example to detect a bad rollout. That tooling doesn't work for buffers, which report their health through `ReadyForProvisioning` instead ([#10080](https://github.com/kubernetes/autoscaler/issues/10080)). The doc comment on `CapacityBufferStatus.Conditions` already describes a `Ready` condition, but no implementation sets it, and it defines `Ready` as the buffer being "successfully provisioned and active", which is about capacity. This proposal defines `Ready` differently: it reports whether the buffer is valid, and capacity is reported separately by `Provisioned`. The doc comment is updated to match.

That `Ready` condition has to be stable. A buffer's capacity is consumed by design: workloads use it, and the node autoscaler refills it. If `Ready` followed capacity, it would flap during normal operation, and alerts built on it would be noise. This matches how `Ready` works for other policy objects such as Karpenter's NodePool, whose `Ready` depends on `ValidationSucceeded` and `NodeClassReady` but not on how many nodes exist.

The second gap is that no condition answers whether the buffer's capacity is actually there. [#10127](https://github.com/kubernetes/autoscaler/issues/10127)

*   `Provisioning` means different things in each implementation. In Cluster Autoscaler, it means the autoscaler is acting on the buffer. In Karpenter, it means the buffer's replicas fit on existing nodes, including nodes that have launched but aren't initialized yet.
*   `status.readyReplicas` ([#9238](https://github.com/kubernetes/autoscaler/pull/9238), [#10342](https://github.com/kubernetes/autoscaler/pull/10342)) is a count, so consumers have to compare it to `status.replicas` themselves. See [Use `status.readyReplicas` instead of a `Provisioned` condition](#use-statusreadyreplicas-instead-of-a-provisioned-condition).

The `CapacityBuffer` API is designed to be agnostic of both the cloud provider and the node autoscaler implementation. The status doesn't yet: the conditions a buffer reports are defined separately in each node autoscaler. Integrations like Kueue, and any future consumer of the `CapacityBuffer` API, read buffer status to decide what to do. They should be able to read a buffer's `Ready` and `Provisioned` conditions the same way whether the cluster runs Cluster Autoscaler or Karpenter, without having to track each implementation's interpretation.

## Goals

*   Define `Ready` and `Provisioned` conditions with the same meaning across implementations
*   Keep `Ready` stable by keeping capacity out of it
*   Replace `ReadyForProvisioning` with `Ready`, and remove `ReadyForProvisioning` when the API graduates to v1
*   Replace `Provisioning` with `Provisioned`, and remove `Provisioning` when the API graduates to v1
*   Establish a common place for shared constants of the `CapacityBuffer` API

## Non-Goals

*   Create a new shared place for the entire `CapacityBuffer` API
*   Remove or rename any existing condition types other than `ReadyForProvisioning` and `Provisioning`
*   Define a shared meaning for `Provisioning`
*   Replace `status.readyReplicas`
*   Define how future buffer types, such as ephemeral buffers, affect `Ready`
*   Standardize condition reasons that are specific to a single implementation
*   Propose new provisioning strategies

# Proposal

## Ready Condition

`Ready` is set by the buffer controller and has the meaning `ReadyForProvisioning` has today.

*   **True:** the buffer's pod template and desired replica count have been resolved into `status.podTemplateRef` and `status.replicas`, and a node autoscaler can act on it.
*   **False:** the buffer's configuration is invalid or couldn't be resolved. The message describes the error.

`Ready` does not depend on capacity. A buffer whose capacity is being used or refilled is working as intended, so it stays `Ready`. Capacity is reported by `Provisioned`.

Node autoscalers provision capacity only for buffers that are `Ready`, but they update `Provisioned` on every buffer whose provisioning strategy they support, so a buffer that stops being `Ready` gets `Provisioned=False` with `BufferNotReady`. `ReadyForProvisioning` is set alongside `Ready` until it is removed at v1 graduation.

Future proposals may add other cases that make `Ready` False, such as an `Expired` condition for ephemeral buffers. Each needs its own proposal, because it changes when `Ready` is False.

## Provisioned Condition

`Provisioned` is set by the node autoscaler, because the node autoscaler is the component that knows whether the buffer's capacity exists. The node autoscaler sets it as soon as it starts acting on a buffer, so the presence of `Provisioned` shows that a node autoscaler has picked the buffer up. This makes `Provisioning` redundant, so it is removed at v1 graduation along with `ReadyForProvisioning`. Until then, `Provisioning` is not part of the shared contract.

*   **True:** the buffer is `Ready`, it is not `LimitedByQuotas`, and every replica in `status.replicas` fits on existing nodes that are ready to run pods.
*   **False:** the buffer is not `Ready`, it is `LimitedByQuotas`, or some of its replicas don't fit on nodes that are ready to run pods yet.

`Provisioned=True` always implies `Ready=True`. A buffer that isn't `Ready` has no valid configuration to be provisioned against, so `Provisioned` is False.

`Provisioned` is a separate condition from `Ready` because it is expected to change often. Like `status.readyReplicas`, `Provisioned` is measured against `status.replicas`, which is the replica count after quotas are applied. Unlike `status.readyReplicas`, a buffer limited by quotas is not `Provisioned`, even if all of its allowed replicas exist, because it has less capacity than its spec asks for. A buffer that is empty by its own spec, with `status.replicas` of zero and not `LimitedByQuotas`, is `Provisioned`, because there is no capacity to wait for. See [Use `status.readyReplicas` instead of a `Provisioned` condition](#use-statusreadyreplicas-instead-of-a-provisioned-condition).

## Shared Constants

A new file `constants.go` is added to each versioned API package, starting with `/cluster-autoscaler/apis/capacitybuffer/autoscaling.x-k8s.io/v1beta1`, to home shared condition types, shared condition reasons, and the default provisioning strategy. The v1 package's `constants.go` won't include `ReadyForProvisioningCondition`, so v1beta1 consumers keep the constant until they move to v1. This file should also be the place for any annotations shared across implementations in the future.

Node autoscalers and other implementations or uses of the buffers API will refer to or import these definitions. The existing types are taken from the [Cluster Autoscaler](https://github.com/kubernetes-sigs/cluster-autoscaler/blob/main/pkg/capacitybuffer/constants.go) and [Karpenter](https://github.com/kubernetes-sigs/karpenter/blob/main/pkg/apis/autoscaling/v1beta1/constants.go) constants.

The list of reasons is intentionally narrow: only the reasons for the new `Provisioned` condition are shared. `Ready`'s reasons describe validation errors, which differ by implementation, so they are not shared. `Provisioning` is not part of the shared contract, so its reasons are left out too. Implementations remain free to use their own reasons for every other case.

The proposed file is in the [Appendix](#appendix-proposed-v1beta1constantsgo).

# Design Details

## Ready Condition

The buffer controller sets `Ready` wherever it sets `ReadyForProvisioning` today, with the same status, reason, and message.

Each condition type has exactly one writer. The buffer controller writes `Ready`, `ReadyForProvisioning`, and `LimitedByQuotas`. The node autoscaler writes `Provisioned`. The node autoscaler reads `Ready` and `LimitedByQuotas` but never writes them. `status.conditions` is a map keyed by condition type, so writers that use server-side apply, or that update with resource-version conflict checks, don't overwrite each other's conditions.

## Provisioned Condition

| Status | Reason | When |
|---|---|---|
| True | `CapacityReady` | `Ready` is True, `LimitedByQuotas` is not True, and every replica in `status.replicas` fits on existing, ready nodes. |
| False | `BufferNotReady` | `Ready` is not True. |
| False | `ExceedsResourceQuotas` | `Ready` is True, and `LimitedByQuotas` is True. |
| False | `CapacityNotReady` | `Ready` is True, `LimitedByQuotas` is not True, and some replicas don't fit on ready nodes yet. |

Implementations that set `status.readyReplicas` must keep it consistent with `Provisioned`: when `Provisioned` is True, `status.readyReplicas` equals `status.replicas`.

## Removing `ReadyForProvisioning` and `Provisioning` at v1 Graduation

1.  While the API is v1beta1, the buffer controller sets both `Ready` and `ReadyForProvisioning`, with the same status, reason, and message. Node autoscalers may keep setting `Provisioning` alongside `Provisioned`.
2.  Node autoscalers and other consumers switch from reading `ReadyForProvisioning` to reading `Ready`, and from reading `Provisioning` to reading `Provisioned`.
3.  `ReadyForProvisioning` and `Provisioning` are removed when the `CapacityBuffer` API graduates to v1. The v1 API and its `constants.go` don't include them.

## Compatibility with Existing Implementations

What each implementation writes today, compared with what this proposal defines:

| Proposed | Cluster Autoscaler today | Karpenter today |
|---|---|---|
| `Ready` | Not set. `ReadyForProvisioning` has this meaning. | Not set. `ReadyForProvisioning` has this meaning. |
| `ReadyForProvisioning` (removed at v1 graduation) | Set by the buffer controller | Same meaning |
| `Provisioned=True` / `CapacityReady` | Not set. Closest is `status.readyReplicas` equal to `status.replicas`. | Not set. Closest is `Provisioning=True` / `FitsExistingCapacity`, which also counts nodes that have launched and aren't initialized yet. |
| `Provisioned=False` / `CapacityNotReady` | Not set. Closest is `status.readyReplicas` lower than `status.replicas`. | Not set. Closest is `Provisioning=False` / `RequiresNewCapacity`. |
| `Provisioned=False` / `ExceedsResourceQuotas` | Not set. Closest is `LimitedByQuotas=True`. | Not set. `LimitedByQuotas` is never set. |
| `Provisioned=False` / `BufferNotReady` | Not set. `status.readyReplicas` drops to 0. | Not set. Closest is `Provisioning=False` / `NotReadyForProvisioning`. |
| `Provisioning` (removed at v1 graduation) | The autoscaler is acting on the buffer | The buffer's replicas fit on existing nodes. This meaning moves to `Provisioned`. |
| `LimitedByQuotas` | Set by the buffer controller | Defined, never set |
| `ActiveProvisioningStrategy` | Default; only buffers with this strategy are processed | Defined, unused; all buffers are processed |

Implementations of the `CapacityBuffer` API, such as Cluster Autoscaler and Karpenter, will need to change to match the definitions in this proposal.

# Open Questions

*   **Should `Provisioned=False` distinguish why capacity isn't ready?** For example: nodes are still starting, new capacity is needed, or capacity can't be obtained. [#10127](https://github.com/kubernetes/autoscaler/issues/10127) mentions stockouts and insufficient capacity.

# Alternatives

## Do nothing

Each implementation keeps defining its conditions in its own repository, as Cluster Autoscaler and Karpenter do today, and no `Ready` or `Provisioned` condition is set.

### Why not?

Monitoring built on `Ready` doesn't work for buffers, and consumers have no condition that says the buffer's capacity exists. The implementations have already diverged: `Provisioning` means different things in each, and a change in one repository isn't necessarily reviewed by the owners of the other. Consumers such as Kueue have to handle each implementation separately.

## Make `Ready` depend on capacity

`Ready` is True only when the buffer is valid and `Provisioned`.

### Why not?

`Ready` would flap during normal operation, every time workloads use the buffer's capacity and the node autoscaler refills it. Alerts built on `Ready` would fire for a buffer that is working as intended. Consumers that need all of the capacity to be available, such as gang-scheduled jobs that use ephemeral buffers, can read `Provisioned` instead.

## Keep `ReadyForProvisioning` alongside `Ready` permanently

`Ready` is added, and `ReadyForProvisioning` is kept in v1.

### Why not?

The two conditions would have the same meaning forever, and every reader would have to know they are equivalent. Setting both through v1beta1 gives readers time to switch before `ReadyForProvisioning` is removed.

## Keep `Provisioned`'s last value while the buffer is not `Ready`

When a buffer becomes invalid, the node autoscaler leaves `Provisioned` unchanged, because the capacity from the last valid configuration still exists.

### Why not?

`Provisioned=True` would claim the buffer's configuration is satisfied when there is no valid configuration. Consumers would have to check `Ready` as well before trusting `Provisioned`.

## Use `status.readyReplicas` instead of a `Provisioned` condition

Consumers compare `status.readyReplicas` to `status.replicas` to decide whether a buffer is provisioned. This was raised on [#10127](https://github.com/kubernetes/autoscaler/issues/10127).

### Why not?

`status.readyReplicas` is a count, not a state. `kubectl wait --for=jsonpath` can only compare a field to a fixed value, so it can't wait for `status.readyReplicas` to reach `status.replicas`, while `kubectl wait --for=condition=Provisioned` waits until `Provisioned` is True. A count also has no reason field, so it can't say why capacity is missing. `Provisioned` complements `status.readyReplicas` rather than replacing it.

## Keep `Provisioning` in v1

`Provisioning` is kept as a condition that each implementation defines for itself.

### Why not?

`Provisioned` already shows that a node autoscaler has picked the buffer up, so `Provisioning` would be redundant. Keeping a condition with no shared meaning, under a name every implementation uses differently, would keep confusing consumers.

## Document the shared semantics in the original proposal only

The definitions are written out in [`buffers.md`](buffers.md), and each implementation keeps its own Go constants.

### Why not?

Written definitions can't be imported, so the implementations' constants can still drift apart from the document and from each other. A change to an implementation's constants also doesn't require any change next to the CRD, so it doesn't get reviewed by all owners.

# Appendix: Proposed `v1beta1/constants.go`

```go
// Constants shared across implementations of the CapacityBuffer API.
const (
	// ActiveProvisioningStrategy is the default value of spec.provisioningStrategy.
	// A buffer with this strategy, or with no strategy set, is served by the node
	// autoscaler placing placeholder pods in its scheduling simulation. The cluster
	// then scales up so the buffer's capacity stays available.
	ActiveProvisioningStrategy = "buffer.x-k8s.io/active-capacity"

	// ReadyCondition is set by the buffer controller.
	// True: the buffer's pod template and desired replica count have been resolved
	// into status.podTemplateRef and status.replicas, and a node autoscaler can act on them.
	// False: the buffer's configuration is invalid or couldn't be resolved; the
	// message describes the error.
	// Ready does not depend on whether the buffer's capacity exists; see
	// ProvisionedCondition.
	ReadyCondition = "Ready"

	// ProvisionedCondition is set by the node autoscaler as soon as it starts
	// acting on a buffer.
	// True: the buffer is Ready, it is not LimitedByQuotas, and every replica in
	// status.replicas fits on existing nodes that are ready to run pods.
	// False: the buffer is not Ready, it is LimitedByQuotas, or some of its
	// replicas don't fit on nodes that are ready to run pods yet.
	ProvisionedCondition = "Provisioned"

	// ReadyForProvisioningCondition has the same meaning as ReadyCondition.
	// It is not included in the v1 API; use ReadyCondition.
	ReadyForProvisioningCondition = "ReadyForProvisioning"

	// LimitedByQuotasCondition is set by the buffer controller.
	// True: status.replicas is lower than the desired count so that the buffer fits
	// within the namespace's ResourceQuotas; the message lists the limiting quotas.
	// False: the buffer is not limited by quotas.
	LimitedByQuotasCondition = "LimitedByQuotas"

	// CapacityReadyReason is the reason for Provisioned=True.
	CapacityReadyReason = "CapacityReady"

	// CapacityNotReadyReason is the reason for Provisioned=False when the buffer
	// is Ready and not LimitedByQuotas, but some of its replicas don't fit on
	// ready nodes yet.
	CapacityNotReadyReason = "CapacityNotReady"

	// BufferNotReadyReason is the reason for Provisioned=False when the buffer is
	// not Ready.
	BufferNotReadyReason = "BufferNotReady"

	// ExceedsResourceQuotasReason is the reason for Provisioned=False when the
	// buffer is Ready but LimitedByQuotas.
	ExceedsResourceQuotasReason = "ExceedsResourceQuotas"
)
```
