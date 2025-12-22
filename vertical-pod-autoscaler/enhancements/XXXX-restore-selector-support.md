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

#### Spec Divergence
**Risk**: A Pod's role may not be known at creation time. Therefore, every new Pod inherits the Deployment pod spec and only promotes to "Leader" (and the corresponding VPA profile) after the application starts and wins an election.

**Consequence**: There is an unavoidable window between Election and VPA Actuation where the new Leader is running with default pod spec resources.

##### Mitigation

**Safe Baseline Requests**: Users must configure the Deployment.spec.template.resources.requests to be a "Safe Floor"—sufficient to handle the application's boot sequence and the initial election workload without crashing.

**Reactive Resizing**: The VPA Updater will detect the role=leader label transition and perform an In-Place Update to increase capacity to the "Leader Profile."

#### Conflicting Targets
**Risk**: A user might configure two VPAs for the same TargetRef that overlap, causing "Split Brain" scaling where two Recommenders fight over the same Pod.

VPA A: `targetRef: app (No selector = All pods)`

VPA B: `targetRef: app, selector: role=leader`

**Result**: The Leader pod matches both. VPA A wants 1GB, VPA B wants 8GB. The Pod thrashes.

**Mitigation**: The Admission Webhook will be updated to enforce Strict Selector Disjointness.

**Current Rule**: "Reject if more than one VPA points to the same TargetRef."

**New Validation Logic**: "Allow multiple VPAs per TargetRef only if they are proven disjoint."

To ensure safety without complex constraint solving, the Webhook will enforce the "Shared Key Exclusion" heuristic. Two VPAs targeting the same controller are considered disjoint if they filter on the same label key but require different values.

**Validation Algorithm**: When a new VPA (VPA-New) is created/updated:

**Identify Siblings**: Find all existing VPAs (VPA-Existing) pointing to the same TargetRef.

**Iterate & Compare**: For each VPA-Existing:

Case 1: Catch-All Conflict (Reject) If either VPA-New or VPA-Existing has a nil selector, REJECT. (You cannot have a "Default" VPA alongside a "Specific" VPA).

Case 2: Different Keys (Reject - Ambiguous) If VPA-New selects on role=leader and VPA-Existing selects on env=prod, REJECT. Reason: A pod could theoretically have {role: leader, env: prod}. We cannot predict future labels, so we must assume overlap is possible.

Case 3: Same Key, Different Values (Allow) If VPA-New selects role=leader and VPA-Existing selects role=follower, ALLOW. Reason: A single label key cannot hold two values simultaneously. Disjointness is mathematically guaranteed.

#### Recommender Data Sparsity
**Risk**: Partitioning a set of pods reduces the sample size for the histogram (e.g., a single Leader pod).

**Mitigation**: The Recommender must aggregate historical samples at the VPA Object Level.

This ensures that when a Pod promotes to Leader, it inherits the "Leader Profile" history immediately, rather than starting with a cold cache.

#### Unmanaged Pods
**Risk**: A pod might match the TargetRef but fail to match any VPA Selector.

**Result**: The pod receives no recommendations.

**Mitigation**: This is acceptable behavior. It allows users to intentionally exclude specific pods from autoscaling.

## Future Possibilities: Direct Lease Integration

We discussed the possibility of VPA directly watching `coordination.k8s.io/Lease` objects to detect Leader/Follower transitions.

This proposal (or something similar) will be required to support such a feature in VPA.

1.  **The Foundation:** To support Leader/Follower scaling, the VPA core must first possess the ability to maintain separate metric histories and generate distinct recommendations for subsets of a single Controller. The `Selector` field provides this mechanism.
2.  **The Additive Layer:** Once `Selector` support is added, future enhancements could allow VPA to automatically infer these selectors by watching a Lease, or users can simply use a "Lease-to-Label" controller to bridge the gap without modifying VPA core further.