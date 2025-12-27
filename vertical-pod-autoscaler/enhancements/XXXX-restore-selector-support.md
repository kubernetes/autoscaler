# AEP-XXXX: Support for Heterogeneous Workloads in StatefulSets and Deployments

## Summary

This proposal enhances the VerticalPodAutoscalerSpec to allow the co-existence of TargetRef and Selector (present in v1beta1 but removed in v1). When both are present, the Selector acts as a filter applied to the Pods managed by the TargetRef. This allows multiple VPA objects to manage disjoint subsets of a single Controller's pods (e.g., partitioning a StatefulSet/Deployment into Leaders and Followers).

## Motivation

Currently, VPA relies exclusively on `targetRef` to identify Pods. This enforces a 1:1 relationship between the VPA and a Workload Controller (Deployment, StatefulSet, etc.).

While this is sufficient for stateless workloads, it fails for **Heterogeneous Stateful Workloads** where pods in the same controller perform different roles with different resource footprints.

**The Problem: Leader vs. Follower**
VPA aggregates metrics from all Pods in the target controller into a single histogram. This averages the usage of "high-utilization" (Leader) and "low-utilization" (Follower) pods.

**The Solution:**
By restoring the `selector` field, users can partition a single workload into multiple VPA profiles based on the Pod's current state:
1.  **VPA-Leader:** Selects `role=leader`
2.  **VPA-Follower:** Selects `role=follower`

When a Pod promotes from Follower to Leader, its label changes, and it effectively migrates from the Follower VPA to the Leader VPA instantly.

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

Modify the `VerticalPodAutoscalerSpec` to re-introduce `Selector` as an optional field. This field was present in `v1beta1` but removed in `v1`.

### API Spec

```go
type VerticalPodAutoscalerSpec struct {
	// TargetRef points to the controller managing the set of pods.
	// + required
	TargetRef *autoscaling.CrossVersionObjectReference `json:"targetRef"`

	// [PROPOSED]
	// A label query that further restricts the set of pods controlled by the Autoscaler.
	//
	// If provided, the VPA manages only the subset of pods that:
	// 1. Are owned by the TargetRef Controller, AND
	// 2. Match this Label Selector.
	//
	// +optional
	Selector *metav1.LabelSelector `json:"selector,omitempty"`
}
```

### Behavior
Refinement Mode (TargetRef + Selector):

1. The VPA Updater/Recommender identifies the set of pods managed by the TargetRef (Current behavior).

2. It applies the Selector as a filter to this list.

Result: The VPA only generates recommendations and acts upon pods that pass the filter.

### Risks and Mitigations

#### 1. Spec Divergence
**Risk**: A Pod's role may not be known at creation time. Therefore, every new Pod inherits the Deployment pod spec and only promotes to "Leader" (and the corresponding VPA profile) after the application starts and wins an election.

**Consequence**: There is an unavoidable window between Election and VPA Actuation where the new Leader is running with default pod spec resources.

##### Mitigation

**Safe Baseline Requests**: Users must configure the Deployment.spec.template.resources.requests to be a "Safe Floor"—sufficient to handle the application's boot sequence and the initial election workload without crashing.

**Reactive Resizing**: The VPA Updater will detect the role=leader label transition and perform an In-Place Update to increase capacity to the "Leader Profile."

#### 2. Conflicting Targets (Overlap & Precedence)
**Context:** Historically, VPA behavior was undefined if multiple VPA objects targeted the same workload. However, with AEP-8026 moving configuration parameters (e.g., Safety Margins, History Length) to the VPA Custom Resource, users will increasingly require multiple VPA objects to apply different scaling policies to different subsets of a single Deployment.

**Risk:** A Pod might match multiple Active VPAs simultaneously. For example, a pod labeled `role=leader` matches both a global VPA (Selector: `nil`) and a specific VPA (Selector: `role=leader`).

**Mitigation: Specificity Precedence**
To support granular configuration without conflict, the VPA Updater will resolve overlaps using **Selector Specificity**, similar to Kubernetes Network Policies or CSS:

1.  **Rule:** If a Pod matches multiple VPAs, the VPA with the **Most Specific Selector** takes precedence.
    * *Level 0 (Global):* `selector: null` (Matches everything in TargetRef).
    * *Level 1 (Refined):* `selector: {key: value}` (Matches specific subset).
2.  **Resolution:** The Updater will respect the recommendations from the *Level 1* VPA and ignore the *Level 0* VPA for that specific pod.
3.  **Tie-Breaking:** If two VPAs have equally specific selectors (e.g., VPA-A selects `role=leader` and VPA-B selects `tier=gold`), and a pod matches *both*, this is considered a misconfiguration. The Updater will deterministically break the tie by choosing the VPA with the oldest `CreationTimestamp` and emit a Warning Event.

#### 3. Recommender Data Sparsity
**Risk**: Partitioning a set of pods reduces the sample size for the histogram (e.g., a single Leader pod).

**Mitigation**: The VPA object acts as a persistent store for the *Role*, rather than the specific Pod instance.
* **Inheritance:** When leadership rotates (e.g., `pod-0` steps down and `pod-1` becomes Leader), `pod-1` immediately inherits the historical resource profile built by `pod-0`.
* **Result:** The VPA aggregates data at the **Selector Level**, ensuring that the "Leader Profile" is continuous and robust over time, even if the specific Pod holding the title changes frequently.

#### Unmanaged Pods
**Risk**: A pod might match the TargetRef but fail to match any VPA Selector.

**Result**: The pod receives no recommendations.

**Mitigation**: This is acceptable behavior. It allows users to intentionally exclude specific pods from autoscaling.

## Future Possibilities: Direct Lease Integration

We discussed the possibility of VPA directly watching `coordination.k8s.io/Lease` objects to detect Leader/Follower transitions.

This proposal (or something similar) will be required to support such a feature in VPA.

1.  **The Foundation:** To support Leader/Follower scaling, the VPA core must first possess the ability to maintain separate metric histories and generate distinct recommendations for subsets of a single Controller. The `Selector` field provides this mechanism.
2.  **The Additive Layer:** Once `Selector` support is added, future enhancements could allow VPA to automatically infer these selectors by watching a Lease, or users can simply use a "Lease-to-Label" controller to bridge the gap without modifying VPA core further.

## Feature Enablement

This feature will be guarded by the `VPASelectorRefinement` feature gate.
* **Default:** `false` (Alpha).
* **Rollback:** If disabling the feature or rolling back the binary, administrators **must** delete any VPA objects using the `selector` field first. Older binaries will ignore the field and mistakenly apply the VPA to the entire TargetRef.

## Graduation Criteria

**Alpha → Beta**
* Feature gate enabled by default.
* E2E tests passing consistently in CI.
* User feedback collected from at least one production adopter.

**Beta → GA**
* No critical bugs reported regarding target overlaps or race conditions.
* Feature stable for 2 releases.

## Test Plan

### 1. Unit Tests
* **Matcher Logic:** Verify that `GetControllingVPA` correctly filters pods based on the new `selector` field.
* **Precedence Logic:** create mock VPAs with conflicting selectors and assert that the most specific one is chosen.
* **Admission Validation:** Verify the webhook rejects invalid selectors (e.g. malformed strings).

### 2. End-to-End (E2E) Tests
* **Scenario:** `LeaderFollowerScaling`
* **Setup:** Deploy a StatefulSet and two VPAs (Base + Leader).
* **Action:** Apply the `role=leader` label to one Pod.
* **Assertion:** Verify that the "Leader VPA" takes over control of that specific Pod and updates its recommended resources, while the other pods remain under the "Base VPA."

## Implementation History

- **2025-12-26:** Initial AEP draft submitted.
- **2025-XX-XX:** [Pending] Proposal approved by SIG Autoscaling.
