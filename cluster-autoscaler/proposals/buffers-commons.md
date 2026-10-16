# CapacityBuffer Conditions and Shared API Constants

#### Author: Joshua Guo (joshuakguo)

# Summary

This proposal adds two conditions to `CapacityBuffer`:

*   `Ready` is the summary condition that cluster administrators and tools monitor ([#10080](https://github.com/kubernetes/autoscaler/issues/10080)). It is True when the buffer has no blocking problems. It does not depend on capacity, so it doesn't flap as the buffer's capacity is used and refilled.
*   `Provisioned` is True when all of the buffer's capacity exists on nodes that are ready to run pods ([#10127](https://github.com/kubernetes/autoscaler/issues/10127)). It is expected to change as the buffer's capacity is used and refilled.

While making that change, this proposal also adds a shared `constants.go` file next to the CRD. It defines the condition types, the condition reasons that all implementations share, and the default provisioning strategy, so every node autoscaler implementation can import the same definitions.

# Motivation

The conditions a buffer reports today leave two gaps.

The first is that buffers have no `Ready` condition. Cluster administrators commonly build monitoring and alerts on the `Ready` condition of Kubernetes objects, for example to detect a bad rollout. That tooling doesn't work for buffers, which report their health through `ReadyForProvisioning` instead ([#10080](https://github.com/kubernetes/autoscaler/issues/10080)). The doc comment on `CapacityBufferStatus.Conditions` already describes a `Ready` condition, but no implementation sets it.

That `Ready` condition has to be stable. A buffer's capacity is consumed by design: workloads use it, and the node autoscaler refills it. If `Ready` followed capacity, it would flap during normal operation, and alerts built on it would be noise. This matches how `Ready` works for other policy objects such as Karpenter's NodePool, whose `Ready` depends on `ValidationSucceeded` and `NodeClassReady` but not on how many nodes exist.

The second gap is that no condition answers whether the buffer's capacity is actually there. [#10127](https://github.com/kubernetes/autoscaler/issues/10127)

*   `Provisioning` means different things in each implementation. In Cluster Autoscaler, it means the autoscaler is acting on the buffer. In Karpenter, it means the buffer's capacity already exists.
*   `status.readyReplicas` ([#9238](https://github.com/kubernetes/autoscaler/pull/9238), [#10342](https://github.com/kubernetes/autoscaler/pull/10342)) See [Use `status.readyReplicas` instead of a `Provisioned` condition](#use-statusreadyreplicas-instead-of-a-provisioned-condition).

The `CapacityBuffer` API is designed to be agnostic of both the cloud provider and the node autoscaler implementation. The status doesn't yet: the conditions a buffer reports are defined separately in each node autoscaler. Integrations like Kueue, and any future consumer of the `CapacityBuffer` API, read buffer status to decide what to do. They should be able to read a buffer's `Ready` and `Provisioned` conditions the same way whether the cluster runs Cluster Autoscaler or Karpenter, without having to track each implementation's interpretation.

## Goals

*   Define `Ready` and `Provisioned` conditions with the same meaning across implementations
*   Keep `Ready` stable by keeping capacity out of it
*   Establish a common place for shared constants of the `CapacityBuffer` API
*   Share condition reasons only for cases that every implementation reports for the same condition, status, and cause

## Non-Goals

*   Create a new shared place for the entire `CapacityBuffer` API
*   Remove or rename any existing condition types
*   Replace `status.readyReplicas`
*   Define blocking conditions for future buffer types, such as ephemeral buffers
*   Standardize condition reasons that are specific to a single implementation
*   Propose new provisioning strategies

# Proposal

## Ready Condition

`Ready` rolls up the buffer's blocking conditions. A blocking condition reports a problem that keeps the buffer from working as specified until someone changes its configuration or environment.

*   **True:** no blocking condition reports a problem.
*   **False:** at least one blocking condition reports a problem. The reason and message say which, and why.

Today there are two blocking conditions:

*   `ReadyForProvisioning` reports a problem when it is not True: the buffer's configuration is invalid or couldn't be resolved.
*   `LimitedByQuotas` reports a problem when it is True: the namespace's ResourceQuotas can't fit the buffer, so it provides less capacity than its spec asks for. [#10080](https://github.com/kubernetes/autoscaler/issues/10080)

Future conditions can become blocking without changing what `Ready` means, such as an `Expired` condition for ephemeral buffers or a persistent provisioning failure.

`Ready` does not depend on capacity. A buffer whose capacity is being used or refilled is working as intended, so it stays `Ready`. Capacity is reported by `Provisioned`.

`ReadyForProvisioning` stays as the signal from the buffer controller to the node autoscaler that a buffer is safe to act on.

## Provisioned Condition

`Provisioned` is set by the node autoscaler, because the node autoscaler is the component that knows whether the buffer's capacity exists.

*   **True:** every replica in `status.replicas` fits on existing nodes that are ready to run pods.
*   **False:** some of the buffer's replicas don't fit on nodes that are ready to run pods yet.

`Provisioned` is independent of `Ready` because it is expected to be able to change often. It describes the last valid configuration the node autoscaler acted on. If a user makes a mistake in a buffer's configuration, the buffer becomes not `Ready`, and the node autoscaler stops acting on it. `Provisioned` keeps its last value, because the capacity from the last valid configuration still exists.

`Provisioned` is the condition form of `status.readyReplicas`. Both are measured against `status.replicas`, which is the replica count after quotas are applied. A buffer limited by quotas can be `Provisioned`, because all of its allowed replicas exist, but it is not `Ready`, because `LimitedByQuotas` is blocking. See [Use `status.readyReplicas` instead of a `Provisioned` condition](#use-statusreadyreplicas-instead-of-a-provisioned-condition).

## Shared Constants

A new file `constants.go` is added under `/cluster-autoscaler/apis/capacitybuffer/autoscaling.x-k8s.io` to home shared condition types, shared condition reasons, and the default provisioning strategy. This file should also be the place for any annotations shared across implementations in the future.

Node autoscalers and other implementations or uses of the buffers API will refer to or import these definitions.

### Condition Types

*   Ready (new)
*   Provisioned (new)
*   ReadyForProvisioning
*   Provisioning
*   LimitedByQuotas

The existing types are taken from the [Cluster Autoscaler](https://github.com/kubernetes-sigs/cluster-autoscaler/blob/main/pkg/capacitybuffer/constants.go) and [Karpenter](https://github.com/kubernetes-sigs/karpenter/blob/main/pkg/apis/autoscaling/v1beta1/constants.go) constants.

### Condition Reasons

This list is intentionally narrow. For existing conditions, a reason is shared only when every implementation already reports the same condition and status for the same cause. Implementations remain free to use their own reasons for every other case. The reasons for the new `Ready` and `Provisioned` conditions are shared by definition.

*   BufferReady (`Ready` only)
*   NotReadyForProvisioning (`Ready` and `Provisioning`)
*   ExceedsResourceQuotas (`Ready` only)
*   BufferIsEmpty (`Provisioning` only)
*   CapacityReady (`Provisioned` only)
*   CapacityNotReady (`Provisioned` only)

### Provisioning Strategies

*   ActiveProvisioningStrategy

# Design Details

## Ready Condition

| Status | Reason | When |
|---|---|---|
| True | `BufferReady` | `ReadyForProvisioning` is True and `LimitedByQuotas` is not True. |
| False | `NotReadyForProvisioning` | `ReadyForProvisioning` is not True. The message is copied from `ReadyForProvisioning`, so the error is visible on `Ready` directly. |
| False | `ExceedsResourceQuotas` | `ReadyForProvisioning` is True and `LimitedByQuotas` is True. The message is copied from `LimitedByQuotas`, so the limiting quotas are visible on `Ready` directly. |

The rows are checked in order. An invalid buffer reports `NotReadyForProvisioning` even if `LimitedByQuotas` is also True, because a buffer that hasn't resolved has no `status.replicas` to compare against its namespace's quotas.

The component that sets a blocking condition updates `Ready` at the same time. Today that is the buffer controller, which sets both `ReadyForProvisioning` and `LimitedByQuotas`.

Making another condition blocking changes when `Ready` is False, so it needs its own proposal. That proposal adds a False row with its own reason to the table above.

## Provisioned Condition

| Status | Reason | When |
|---|---|---|
| True | `CapacityReady` | Every replica in `status.replicas` fits on existing, ready nodes. |
| False | `CapacityNotReady` | Some replicas don't fit on ready nodes yet. |

The node autoscaler updates `Provisioned` only while the buffer is `ReadyForProvisioning`. It leaves `Provisioned` unchanged while the buffer is not `ReadyForProvisioning`.

Implementations that set `status.readyReplicas` must keep it consistent with `Provisioned` while the buffer is `ReadyForProvisioning`: when `Provisioned` is True, `status.readyReplicas` equals `status.replicas`.

## Proposed `constants.go`

```go
// Constants shared across implementations of the CapacityBuffer API.
const (
	// ActiveProvisioningStrategy is the default value of spec.provisioningStrategy.
	// A buffer with this strategy, or with no strategy set, is served by the node
	// autoscaler placing placeholder pods in its scheduling simulation. The cluster
	// then scales up so the buffer's capacity stays available.
	ActiveProvisioningStrategy = "buffer.x-k8s.io/active-capacity"

	// ReadyCondition rolls up the buffer's blocking conditions. It is updated by
	// whichever component sets a blocking condition.
	// True: no blocking condition reports a problem.
	// False: at least one blocking condition reports a problem; the reason and
	// message say which, and why.
	// Today the blocking conditions are ReadyForProvisioning (a problem when not
	// True) and LimitedByQuotas (a problem when True). Ready does not depend on
	// whether the buffer's capacity exists; see ProvisionedCondition.
	ReadyCondition = "Ready"

	// ProvisionedCondition is set by the node autoscaler.
	// True: every replica in status.replicas fits on existing nodes that are
	// ready to run pods.
	// False: some of the buffer's replicas don't fit on nodes that are ready to
	// run pods yet.
	// The node autoscaler leaves this condition unchanged while the buffer is not
	// ReadyForProvisioning.
	ProvisionedCondition = "Provisioned"

	// ReadyForProvisioningCondition is set by the buffer controller.
	// True: the buffer's pod template and desired replica count have been resolved
	// into status.podTemplateRef and status.replicas, and a node autoscaler can act on them.
	// False: resolution failed; the message describes the error.
	ReadyForProvisioningCondition = "ReadyForProvisioning"

	// ProvisioningCondition is set by the node autoscaler.
	// Its shared definition is an open question; see Open Questions.
	ProvisioningCondition = "Provisioning"

	// LimitedByQuotasCondition is set by the buffer controller.
	// True: status.replicas is lower than the desired count so that the buffer fits
	// within the namespace's ResourceQuotas; the message lists the limiting quotas.
	// False: the buffer is not limited by quotas.
	// This is a blocking condition: when it is True, Ready is False.
	LimitedByQuotasCondition = "LimitedByQuotas"

	// BufferReadyReason is the reason for Ready=True.
	BufferReadyReason = "BufferReady"

	// NotReadyForProvisioningReason is the reason for Ready=False and
	// Provisioning=False when the buffer's ReadyForProvisioning condition is not True.
	NotReadyForProvisioningReason = "NotReadyForProvisioning"

	// ExceedsResourceQuotasReason is the reason for Ready=False when the buffer
	// is LimitedByQuotas.
	ExceedsResourceQuotasReason = "ExceedsResourceQuotas"

	// BufferIsEmptyReason is the reason for Provisioning=False when the buffer's
	// status.replicas is zero.
	BufferIsEmptyReason = "BufferIsEmpty"

	// CapacityReadyReason is the reason for Provisioned=True.
	CapacityReadyReason = "CapacityReady"

	// CapacityNotReadyReason is the reason for Provisioned=False.
	CapacityNotReadyReason = "CapacityNotReady"
)
```

## Compatibility with Existing Implementations

How each implementation uses these constants today:

| Constant | Cluster Autoscaler | Karpenter |
|---|---|---|
| `Ready` | Not set | Not set |
| `Provisioned` | Not set. `status.readyReplicas` is set, and drops to 0 when the buffer fails to resolve. | Not set. `Provisioning=True` is close, but also counts nodes that have launched and aren't initialized yet. `status.readyReplicas` is not set. |
| `ReadyForProvisioning` | Set by the buffer controller | Same meaning |
| `Provisioning` | The autoscaler is acting on the buffer | The buffer's capacity exists |
| `LimitedByQuotas` | Set by the buffer controller | Defined, never set |
| `ActiveProvisioningStrategy` | Default; only buffers with this strategy are processed | Defined, unused; all buffers are processed |

Implementations of the `CapacityBuffer` API, such as Cluster Autoscaler and Karpenter, will need to change to match the definitions in this proposal.

# Open Questions

*   **What does `Provisioning` mean once `Provisioned` exists?** One option is that the node autoscaler is acting on the buffer.
*   **Should any case where the node autoscaler can't act on a buffer be blocking?** For example, a buffer the node autoscaler can never help, such as one with no compatible node group.
*   **Should `status.readyReplicas` keep its last value while the buffer is not `ReadyForProvisioning`, as `Provisioned` does?**
*   **Is an empty buffer `Provisioned`?** When `status.replicas` is zero, there is no capacity to wait for.
*   **Should `Provisioned=False` distinguish why capacity isn't ready?** For example: nodes are still starting, new capacity is needed, or capacity can't be obtained. [#10127](https://github.com/kubernetes/autoscaler/issues/10127) mentions stockouts and insufficient capacity.

# Alternatives

## Do nothing

Each implementation keeps defining its conditions in its own repository, as Cluster Autoscaler and Karpenter do today, and no `Ready` or `Provisioned` condition is set.

### Why not?

Monitoring built on `Ready` doesn't work for buffers, and consumers have no condition that says the buffer's capacity exists. The implementations have already diverged: `Provisioning` means different things in each, and a change in one repository is never reviewed by the owners of the other. Consumers such as Kueue have to handle each implementation separately.

## Make `Ready` depend on capacity

`Ready` is True only when the buffer is `ReadyForProvisioning` and `Provisioned`.

### Why not?

`Ready` would flap during normal operation, every time workloads use the buffer's capacity and the node autoscaler refills it. Alerts built on `Ready` would fire for a buffer that is working as intended. Consumers that need all of the capacity to be available, such as ephemeral buffers for gang scheduling, can read `Provisioned` instead.

## Replace `ReadyForProvisioning` with `Ready`

`ReadyForProvisioning` is renamed to `Ready`.

### Why not?

Every reader of `ReadyForProvisioning` would need a deprecation period in which both conditions are set. A roll-up gives the same result today without breaking anyone. It also leaves room for future blocking conditions without changing what `Ready` means. `ReadyForProvisioning` stays as the specific signal the node autoscaler waits on.

## Use `status.readyReplicas` instead of a `Provisioned` condition

Consumers compare `status.readyReplicas` to `status.replicas` to decide whether a buffer is provisioned. This was raised on [#10127](https://github.com/kubernetes/autoscaler/issues/10127).

### Why not?

`status.readyReplicas` is a count, not a state. Tools that wait on conditions, such as `kubectl wait --for=condition=Provisioned`, can't use it. A count also has no reason field, so it can't say why capacity is missing. `Provisioned` complements `status.readyReplicas` rather than replacing it.

## Document the shared semantics in the original proposal only

The definitions are written out in [`buffers.md`](buffers.md), and each implementation keeps its own Go constants.

### Why not?

Written definitions can't be imported, so the implementations' constants can still drift apart from the document and from each other. A change to an implementation's constants also doesn't require any change next to the CRD, so it doesn't get reviewed by all owners.
