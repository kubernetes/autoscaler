# AEP-XXXX: Support for Heterogeneous Workloads in StatefulSets and Deployments

<!-- toc -->
- [Summary](#summary)
- [Motivation](#motivation)
  - [Use Cases](#use-cases)
    - [1. Lease-Based Active/Standby (Controllers)](#1-lease-based-activestandby-controllers)
    - [2. Consensus-Based Leaders (Databases)](#2-consensus-based-leaders-databases)
    - [3. Data-Based Asymmetry (Hot Shards)](#3-data-based-asymmetry-hot-shards)
- [Proposal](#proposal)
  - [API Changes](#api-changes)
  - [Mechanism](#mechanism)
- [Risks and Mitigations](#risks-and-mitigations)
  - [1. Boot-up Resource Gap](#1-boot-up-resource-gap)
  - [2. Disruptive Updates](#2-disruptive-updates)
  - [3. Conflicting Targets (Overlap &amp; Precedence)](#3-conflicting-targets-overlap--precedence)
  - [4. Recommender Data Sparsity](#4-recommender-data-sparsity)
  - [5. Unmanaged Pods](#5-unmanaged-pods)
- [Future Possibilities: Direct Lease Integration](#future-possibilities-direct-lease-integration)
- [Feature Enablement](#feature-enablement)
- [Graduation Criteria](#graduation-criteria)
- [Test Plan](#test-plan)
- [Implementation History](#implementation-history)
<!-- /toc -->

## Summary

This proposal enhances the Vertical Pod Autoscaler to support **Heterogeneous Workloads**—specifically **StatefulSets** and **Deployments** where Pods belonging to the same Controller require distinct resource profiles (e.g., Leader vs. Follower).

To achieve this, we propose extending the **`VPAScope`** struct (introduced in AEP-7942) to support **Pod Label Selectors**. By adding a `VPAScope` struct in the VPAAutoScalerSpec definition, users can partition a single Workload Controller into multiple VPA profiles, allowing granular scaling based on dynamic pod roles.

## Motivation

Currently, VPA relies exclusively on `targetRef` to identify Pods. This enforces a 1:1 relationship between the VPA and a Workload Controller (Deployment, StatefulSet, etc.).

While this is sufficient for stateless workloads, it fails for **Heterogeneous Stateful Workloads** where pods in the same controller perform different roles with different resource footprints.

**The Problem: Leader vs. Follower**
VPA aggregates metrics from all Pods in the target controller into a single histogram. This averages the usage of "high-utilization" (Leader) and "low-utilization" (Follower) pods.

**The Solution:**
By leveraging the VPAScope object, users can partition a single workload into multiple VPA profiles based on the Pod's dynamic state:

VPA-Leader: Points to a scope selecting role=leader.

VPA-Follower: Points to a scope selecting role=follower.

Dynamic Transition: When a Pod promotes from Follower to Leader (and its label is updated), it automatically "migrates" from the control of VPA-Follower to VPA-Leader. The VPA Recommender immediately begins calculating resources based on the new profile, ensuring the pod receives the "Leader-sized" resources it now requires.

### Use Cases

The primary motivation for this feature is to support **Deployments** and **StatefulSets** where the *Identity* is uniform (single Controller) but the *Resource Need* is skewed based on the Pod's role or data distribution.

#### 1. Lease-Based Active/Standby (Controllers)
* **Pattern:** Pods race to acquire a `coordination.k8s.io/Lease`. The winner becomes "Active"; others wait in "Standby."
* **Resource Profile:**
    * **Active:** 100% utilization (Performing reconciliation).
    * **Standby:** ~0% utilization (Idling/Polling).
* **VPA Role:** Allows the Active pod to scale up to full capacity, while Standby pods remain minimal (often just enough to run the health check), maximizing density.

#### 2. Consensus-Based Leaders (Databases)
* **Resource Profile:**
    * **Leader:** High CPU/Memory (Processing all writes + coordination).
    * **Follower:** Medium CPU/Memory (Passive replication + read queries).
* **VPA Role:** Ensures the Leader gets "Peak" resources while Followers get "Baseline" resources to save costs.


#### 3. Data-Based Asymmetry (Hot Shards)
* **Pattern:** No specific "Leader" role, but traffic/data is unevenly distributed across shards.
* **Resource Profile:** Specific pods ("Hot Shards") experience sustained high load due to noisy tenants or large partitions.
* **VPA Role:** Allows operators to target and vertically scale only the specific "Hot Nodes" (e.g., labeled `load=high`) without over-provisioning the entire cluster.

## Proposal

To support heterogeneous workloads, we propose modifying the VPA API to allow partitioning of a Controller's pods. We will achieve this by extending the **`VPAScope`** resource (introduced in AEP-7942) to support pod-level filtering in addition to node-level filtering.

This allows users to define a "Scope" (e.g., "Leader Pods" or "Shard A") and link a VPA object to that scope.

### API Changes
We propose modifying the VerticalPodAutoscalerSpec to include a VPAScope struct. This consolidates the partitioning logic into a single definition that supports both horizontal (Pod-based) and vertical (Node-based) filtering.

```go
type VerticalPodAutoscalerSpec struct {
	// TargetRef points to the controller managing the set of pods.
	// +required
	TargetRef *autoscaling.CrossVersionObjectReference `json:"targetRef"`

	// [PROPOSED]
	// VPAScope defines criteria to restrict the VPA to a subset of the target pods.
	// If nil, the VPA controls all pods in the TargetRef.
	// +optional
	VPAScope *VPAScope `json:"vpaScope,omitempty"`
}

// VPAScope defines the logic for partitioning the TargetRef.
type VPAScope struct {
	// [Defined in AEP-7942]
	// If specified, the VPA is restricted to pods running on nodes matching this selector.
	// Primarily used for DaemonSet partitioning.
	// +optional
	NodeSelector *metav1.LabelSelector `json:"nodeSelector,omitempty"`

	// [PROPOSED]
	// If specified, the VPA is restricted to pods matching this label selector.
	// Primarily used for Deployment and StatefulSet partitioning.
	// +optional
	PodSelector *metav1.LabelSelector `json:"podSelector,omitempty"`
}
```

### Mechanism
When a VPA object defines a vpaScope:

The VPA Recommender fetches the referenced VPAScope object.

It identifies the target pods by listing all pods owned by TargetRef.

It filters this list by applying the VPAScope.spec.podSelector (if present).

Only the matching subset of pods are included in the recommendation model.

Pods that do not match the scope are ignored by this specific VPA instance.

## Risks and Mitigations

### 1. Boot-up Resource Gap
**Risk**: A Pod's role is unknown at creation time. Consequently, every new Pod boots with the default Deployment spec resources and only becomes a "Leader" (requiring higher resources) after the application starts and wins an election.

**Consequence**: There is an unavoidable time window between the Election (Role Change) and the VPA Actuation where the new Leader runs without leader sized resources.

**Mitigation**: Users must configure the Deployment.spec.template.resources.requests to be a "Safe Floor"—sufficient to handle the application's boot sequence and the initial election workload.

### 2. Disruptive Updates
**Risk**: VPA can apply resource changes by evicting the pod. For a kubernetes lease based leader, eviction causes immediate failover.

**Failure Mode**: If VPA evicts the leader to resize it, the pod loses leadership. A new pod becomes leader, VPA detects it needs resizing, and evicts it. This creates an infinite failover loop.

**Mitigation**: In-Place Updates & Guardrails To support the "Leader" use case, VPA must be configured to resize without restart.

Admission Warnings: The VPA Admission Webhook will validate VPAScope objects upon creation. If a user sets updateMode: Recreate while using a scope (which implies high-availability/singleton targeting), the webhook will return a warning, alerting the user to the risk of failover loops

In-Place Reliance: This architecture depends on the in place updates of pod resources (Currently inPlaceOrRecreate). Once [#8818](https://github.com/kubernetes/autoscaler/pull/8818) is merged the inPlace update mode is the recommended policy for this feature.

Eviction Guardrail (minReplicas): Users must configure minReplicas: 1 (or use a restrictive PodDisruptionBudget). This acts as a safety latch: if an In-Place update is not possible, the VPA is blocked from falling back to eviction, preventing the failover loop.

### 3. Conflicting Targets (Overlap & Precedence)
**Context:** Multiple VPA objects can target the same workload.

**Risk:** A Pod might match multiple Active VPAs simultaneously. For example, a pod labeled `role=leader` matches both a global VPA (Selector: `nil`) and a specific VPA (Selector: `role=leader`).

**Mitigation: Specificity Precedence**
To support granular configuration without conflict, the VPA Updater will resolve overlaps using **Scope Specificity**, similar to Kubernetes Network Policies or CSS:

1.  **Rule:** If a Pod matches multiple VPAs, the VPA with the **Most Specific Scope** takes precedence.
    * *Level 0 (Global):* `selector: null` (Matches everything in TargetRef).
    * *Level 1 (Refined):* `selector: {key: value}` (Matches specific subset).
2.  **Resolution:** The Updater will respect the recommendations from the *Level 1* VPA and ignore the *Level 0* VPA for that specific pod.
3.  **Tie-Breaking:** If two VPAs have equally specific selectors (e.g., VPA-A selects `role=leader` and VPA-B selects `tier=gold`), and a pod matches *both*, this is considered a misconfiguration. The Updater will deterministically break the tie by choosing the VPA with the oldest `CreationTimestamp` and emit a Warning Event.

### 4. Recommender Data Sparsity
**Risk**: Partitioning a set of pods reduces the sample size for the histogram (e.g., a single Leader pod).

**Mitigation**: The VPA object acts as a persistent store for the *Role*, rather than the specific Pod instance.
* **Inheritance:** When leadership rotates (e.g., `pod-0` steps down and `pod-1` becomes Leader), `pod-1` immediately inherits the historical resource profile built by `pod-0`.
* **Result:** The VPA aggregates data at the **Scope Level**, ensuring that the "Leader Profile" is continuous and robust over time, even if the specific Pod holding the title changes frequently.

### 5. Unmanaged Pods
**Risk**: A pod might match the TargetRef but fail to match any *VPAScope*.

**Result**: The pod receives no recommendations.

**Mitigation**: This is acceptable behavior. It allows users to intentionally exclude specific pods from autoscaling.

## Future Possibilities: Direct Lease Integration

We discussed the possibility of VPA directly watching `coordination.k8s.io/Lease` objects to detect Leader/Follower transitions.

This proposal (or something similar) will be required to support such a feature in VPA.

1.  **The Foundation:** To support Leader/Follower scaling, the VPA core must first possess the ability to maintain separate metric histories and generate distinct recommendations for subsets of a single Controller. The VPAScope integration provides this mechanism. 
2.  **The Additive Layer:** Once VPAScope support is added, future enhancements could allow VPA to automatically infer these scopes by watching a Lease, or users can simply use a "Lease-to-Label" controller to bridge the gap without modifying VPA core further.

## Feature Enablement

This feature will be guarded by the `VPASelectorRefinement` feature gate.
* **Default:** `false` (Alpha).
* **Rollback:** If disabling the feature or rolling back the binary, administrators must delete any VPA objects using the vpaScope field first. Older binaries will ignore the field and mistakenly apply the VPA to the entire TargetRef.

## Graduation Criteria

**Alpha → Beta**
* Feature gate enabled by default.
* E2E tests passing consistently in CI.
* User feedback collected from at least one production adopter.

**Beta → GA**
* No critical bugs reported regarding target overlaps or race conditions.
* Feature stable for 2 releases.

## Test Plan

1. Unit Tests
Matcher Logic: Verify that GetControllingVPA correctly resolves the vpaScope reference and filters pods based on VPAScope.spec.podSelector.

Precedence Logic: Create mock VPAs with conflicting scopes (Global vs. Local) and assert that the VPA with the defined scope is chosen as the controller.

Admission Validation: Verify the webhook rejects invalid scope references (e.g., referencing a missing Scope object) or scopes with malformed selectors.

2. End-to-End (E2E) Tests
Scenario: LeaderFollowerScaling

Setup: Deploy a StatefulSet, a VPAScope (matching role=leader), and two VPAs (Base + Leader).

Action: Apply the role=leader label to one Pod.

Assertion: Verify that the "Leader VPA" takes over control of that specific Pod and updates its recommended resources, while the other pods remain under the "Base VPA."

## Implementation History

- **2025-12-26:** Initial AEP draft submitted.
- **2026-01-20:** Updated design to integrate with VPAScope (AEP-7942).
- **2026-XX-XX:** [Pending] Proposal approved by SIG Autoscaling.
