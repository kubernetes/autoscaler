# CapacityBuffer Conditions and Shared API Constants

#### Author: Joshua Guo (joshuakguo)

# Summary

This proposal adds two conditions to `CapacityBuffer`:

*   `Ready` is the condition that cluster administrators and tools monitor ([#10080](https://github.com/kubernetes/autoscaler/issues/10080)). It is True when the buffer's configuration is valid and a node autoscaler can act on it. It replaces `ReadyForProvisioning`, which this proposal removes when the API graduates to v1. `Ready` does not depend on capacity, so it doesn't flap as the buffer's capacity is used and refilled.
*   `Provisioned` is True when the buffer is `Ready` and all of its capacity exists on nodes that are ready to run pods ([#10127](https://github.com/kubernetes/autoscaler/issues/10127)). It is expected to change as the buffer's capacity is used and refilled.

While making that change, this proposal also adds a shared `constants.go` file next to the CRD. It defines the condition types, the condition reasons that all implementations share, and the default provisioning strategy, so every node autoscaler implementation can import the same definitions.

# Motivation

The conditions a buffer reports today leave two gaps.

The first is that buffers have no `Ready` condition. Cluster administrators commonly build monitoring and alerts on the `Ready` condition of Kubernetes objects, for example to detect a bad rollout. That tooling doesn't work for buffers, which report their health through `ReadyForProvisioning` instead ([#10080](https://github.com/kubernetes/autoscaler/issues/10080)). The doc comment on `CapacityBufferStatus.Conditions` already describes a `Ready` condition, but no implementation sets it, and it defines `Ready` as the buffer being "successfully provisioned and active", which is about capacity. This proposal defines `Ready` differently: it reports whether the buffer is valid, and capacity is reported separately by `Provisioned`.

That `Ready` condition has to be stable. A buffer's capacity is consumed by design: workloads use it, and the node autoscaler refills it. If `Ready` followed capacity, it would flap during normal operation, and alerts built on it would be noise. This matches how `Ready` works for other policy objects such as Karpenter's NodePool, whose `Ready` depends on `ValidationSucceeded` and `NodeClassReady` but not on how many nodes exist.

The second gap is that no condition answers whether the buffer's capacity is actually there. [#10127](https://github.com/kubernetes/autoscaler/issues/10127)

*   `Provisioning` means different things in each implementation. In Cluster Autoscaler, it means the autoscaler is acting on the buffer. In Karpenter, it means the buffer's capacity already exists.
*   `status.readyReplicas` ([#9238](https://github.com/kubernetes/autoscaler/pull/9238), [#10342](https://github.com/kubernetes/autoscaler/pull/10342)) See [Use `status.readyReplicas` instead of a `Provisioned` condition](#use-statusreadyreplicas-instead-of-a-provisioned-condition).

The `CapacityBuffer` API is designed to be agnostic of both the cloud provider and the node autoscaler implementation. The status doesn't yet: the conditions a buffer reports are defined separately in each node autoscaler. Integrations like Kueue, and any future consumer of the `CapacityBuffer` API, read buffer status to decide what to do. They should be able to read a buffer's `Ready` and `Provisioned` conditions the same way whether the cluster runs Cluster Autoscaler or Karpenter, without having to track each implementation's interpretation.

## Goals

*   Define `Ready` and `Provisioned` conditions with the same meaning across implementations
*   Keep `Ready` stable by keeping capacity out of it
*   Replace `ReadyForProvisioning` with `Ready`, and remove `ReadyForProvisioning` when the API graduates to v1
*   Establish a common place for shared constants of the `CapacityBuffer` API
*   Share condition reasons only for cases that every implementation reports for the same condition, status, and cause

## Non-Goals

*   Create a new shared place for the entire `CapacityBuffer` API
*   Remove or rename any existing condition types other than `ReadyForProvisioning`
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

Node autoscalers act only on buffers that are `Ready`. `ReadyForProvisioning` is set alongside `Ready` until it is removed at v1 graduation.

Future proposals may add other cases that make `Ready` False, such as an `Expired` condition for ephemeral buffers. Each needs its own proposal, because it changes when `Ready` is False.

## Provisioned Condition

`Provisioned` is set by the node autoscaler, because the node autoscaler is the component that knows whether the buffer's capacity exists.

*   **True:** the buffer is `Ready`, and every replica in `status.replicas` fits on existing nodes that are ready to run pods.
*   **False:** the buffer is not `Ready`, or some of its replicas don't fit on nodes that are ready to run pods yet.

`Provisioned=True` always implies `Ready=True`. A buffer that isn't `Ready` has no valid configuration to be provisioned against, so `Provisioned` is False.

`Provisioned` is a separate condition from `Ready` because it is expected to change often. `Provisioned` is the condition form of `status.readyReplicas`. Both are measured against `status.replicas`, which is the replica count after quotas are applied. A buffer limited by quotas can be `Provisioned`, because all of its allowed replicas exist. `LimitedByQuotas` reports the shortfall. See [Use `status.readyReplicas` instead of a `Provisioned` condition](#use-statusreadyreplicas-instead-of-a-provisioned-condition).

## Shared Constants

A new file `constants.go` is added under `/cluster-autoscaler/apis/capacitybuffer/autoscaling.x-k8s.io` to home shared condition types, shared condition reasons, and the default provisioning strategy. This file should also be the place for any annotations shared across implementations in the future.

Node autoscalers and other implementations or uses of the buffers API will refer to or import these definitions. The existing types are taken from the [Cluster Autoscaler](https://github.com/kubernetes-sigs/cluster-autoscaler/blob/main/pkg/capacitybuffer/constants.go) and [Karpenter](https://github.com/kubernetes-sigs/karpenter/blob/main/pkg/apis/autoscaling/v1beta1/constants.go) constants.

The list of reasons is intentionally narrow. For existing conditions, a reason is shared only when every implementation already reports the same condition and status for the same cause. Implementations remain free to use their own reasons for every other case. `Ready`'s reasons describe validation errors, which differ by implementation, so they are not shared. The reasons for the new `Provisioned` condition are shared by definition.

The proposed file is in the [Appendix](#appendix-proposed-constantsgo).

# Design Details

## Ready Condition

The buffer controller sets `Ready` wherever it sets `ReadyForProvisioning` today, with the same status, reason, and message.

Each condition type has exactly one writer. The buffer controller writes `Ready`, `ReadyForProvisioning`, and `LimitedByQuotas`. The node autoscaler writes `Provisioned` and `Provisioning`. The node autoscaler reads `Ready` but never writes it. Because `status.conditions` is keyed by condition type, the two components never overwrite each other's conditions.

## Provisioned Condition

| Status | Reason | When |
|---|---|---|
| True | `CapacityReady` | `Ready` is True, and every replica in `status.replicas` fits on existing, ready nodes. |
| False | `BufferNotReady` | `Ready` is not True. |
| False | `CapacityNotReady` | `Ready` is True, and some replicas don't fit on ready nodes yet. |

Implementations that set `status.readyReplicas` must keep it consistent with `Provisioned`: when `Provisioned` is True, `status.readyReplicas` equals `status.replicas`.

## Removing `ReadyForProvisioning` at v1 Graduation

1.  While the API is v1beta1, the buffer controller sets both `Ready` and `ReadyForProvisioning`, with the same status, reason, and message.
2.  Node autoscalers and other consumers switch from reading `ReadyForProvisioning` to reading `Ready`.
3.  `ReadyForProvisioning` is removed when the `CapacityBuffer` API graduates to v1.

## Compatibility with Existing Implementations

What each implementation writes today, compared with what this proposal defines:

| Proposed | Cluster Autoscaler today | Karpenter today |
|---|---|---|
| `Ready` | Not set. `ReadyForProvisioning` has this meaning. | Not set. `ReadyForProvisioning` has this meaning. |
| `ReadyForProvisioning` (removed at v1 graduation) | Set by the buffer controller | Same meaning |
| `Provisioned=True` / `CapacityReady` | Not set. Closest is `status.readyReplicas` equal to `status.replicas`. | Not set. Closest is `Provisioning=True` / `FitsExistingCapacity`, which also counts nodes that have launched and aren't initialized yet. |
| `Provisioned=False` / `CapacityNotReady` | Not set. Closest is `status.readyReplicas` lower than `status.replicas`. | Not set. Closest is `Provisioning=False` / `RequiresNewCapacity`. |
| `Provisioned=False` / `BufferNotReady` | Not set. `status.readyReplicas` drops to 0. | Not set. Closest is `Provisioning=False` / `NotReadyForProvisioning`. |
| `Provisioning` | The autoscaler is acting on the buffer | The buffer's capacity exists. This meaning moves to `Provisioned`. |
| `Provisioning=False` / `BufferIsEmpty` | `BufferIsEmpty` | `BufferEmpty` |
| `LimitedByQuotas` | Set by the buffer controller | Defined, never set |
| `ActiveProvisioningStrategy` | Default; only buffers with this strategy are processed | Defined, unused; all buffers are processed |

Implementations of the `CapacityBuffer` API, such as Cluster Autoscaler and Karpenter, will need to change to match the definitions in this proposal.

# Open Questions

*   **What does `Provisioning` mean once `Provisioned` exists?** One option is that the node autoscaler is acting on the buffer.
*   **Should a buffer the node autoscaler can never help be not `Ready`?** For example, a buffer with no compatible node group.
*   **Is an empty buffer `Provisioned`?** When `status.replicas` is zero, there is no capacity to wait for.
*   **Should `Provisioned=False` distinguish why capacity isn't ready?** For example: nodes are still starting, new capacity is needed, or capacity can't be obtained. [#10127](https://github.com/kubernetes/autoscaler/issues/10127) mentions stockouts and insufficient capacity.

# Alternatives

## Do nothing

Each implementation keeps defining its conditions in its own repository, as Cluster Autoscaler and Karpenter do today, and no `Ready` or `Provisioned` condition is set.

### Why not?

Monitoring built on `Ready` doesn't work for buffers, and consumers have no condition that says the buffer's capacity exists. The implementations have already diverged: `Provisioning` means different things in each, and a change in one repository is never reviewed by the owners of the other. Consumers such as Kueue have to handle each implementation separately.

## Make `Ready` depend on capacity

`Ready` is True only when the buffer is valid and `Provisioned`.

### Why not?

`Ready` would flap during normal operation, every time workloads use the buffer's capacity and the node autoscaler refills it. Alerts built on `Ready` would fire for a buffer that is working as intended. Consumers that need all of the capacity to be available, such as ephemeral buffers for gang scheduling, can read `Provisioned` instead.

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

`status.readyReplicas` is a count, not a state. Tools that wait on conditions, such as `kubectl wait --for=condition=Provisioned`, can't use it. A count also has no reason field, so it can't say why capacity is missing. `Provisioned` complements `status.readyReplicas` rather than replacing it.

## Document the shared semantics in the original proposal only

The definitions are written out in [`buffers.md`](buffers.md), and each implementation keeps its own Go constants.

### Why not?

Written definitions can't be imported, so the implementations' constants can still drift apart from the document and from each other. A change to an implementation's constants also doesn't require any change next to the CRD, so it doesn't get reviewed by all owners.

# Appendix: Proposed `constants.go`

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

	// ProvisionedCondition is set by the node autoscaler.
	// True: the buffer is Ready, and every replica in status.replicas fits on
	// existing nodes that are ready to run pods.
	// False: the buffer is not Ready, or some of its replicas don't fit on nodes
	// that are ready to run pods yet.
	ProvisionedCondition = "Provisioned"

	// ReadyForProvisioningCondition has the same meaning as ReadyCondition.
	// It is removed when the CapacityBuffer API graduates to v1; use ReadyCondition.
	ReadyForProvisioningCondition = "ReadyForProvisioning"

	// ProvisioningCondition is set by the node autoscaler.
	// Its shared definition is an open question; see Open Questions.
	ProvisioningCondition = "Provisioning"

	// LimitedByQuotasCondition is set by the buffer controller.
	// True: status.replicas is lower than the desired count so that the buffer fits
	// within the namespace's ResourceQuotas; the message lists the limiting quotas.
	// False: the buffer is not limited by quotas.
	LimitedByQuotasCondition = "LimitedByQuotas"

	// CapacityReadyReason is the reason for Provisioned=True.
	CapacityReadyReason = "CapacityReady"

	// CapacityNotReadyReason is the reason for Provisioned=False when the buffer
	// is Ready but some of its replicas don't fit on ready nodes yet.
	CapacityNotReadyReason = "CapacityNotReady"

	// BufferNotReadyReason is the reason for Provisioned=False when the buffer is
	// not Ready.
	BufferNotReadyReason = "BufferNotReady"

	// NotReadyForProvisioningReason is the reason for Provisioning=False when the
	// buffer is not Ready.
	NotReadyForProvisioningReason = "NotReadyForProvisioning"

	// BufferIsEmptyReason is the reason for Provisioning=False when the buffer's
	// status.replicas is zero.
	BufferIsEmptyReason = "BufferIsEmpty"
)
```
