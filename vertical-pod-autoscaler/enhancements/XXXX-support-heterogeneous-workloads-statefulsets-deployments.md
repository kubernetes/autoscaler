# AEP-XXXX: Support for Heterogeneous Workloads in StatefulSets and Deployments

## Summary

This proposal enhances the Vertical Pod Autoscaler to support **Heterogeneous Workloads**—specifically **StatefulSets** and **Deployments** where Pods belonging to the same Controller require distinct resource profiles (e.g., Leader vs. Follower).

To achieve this, we propose extending the **`VPAScope`** API object (introduced in AEP-7942) to support **Pod Label Selectors**. By referencing a `VPAScope` in the VPA specification, users can partition a single Workload Controller into multiple VPA profiles, allowing granular scaling based on dynamic pod roles.

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

You are right, I missed the Motivation and the standard Introduction boilerplate. A KEP/AEP needs to tell the full story from start to finish.

Here is the complete, end-to-end AEP with all the sections we refined (Motivation, Use Cases, VPAScope Proposal, Risks, Testing).

Markdown
# AEP-XXXX: Support for Heterogeneous Workloads in StatefulSets and Deployments

## Summary

This proposal aims to enable Vertical Pod Autoscaler to support **Heterogeneous Workloads**—specifically **StatefulSets** and **Deployments** where Pods belonging to the same Controller perform distinct roles with different resource requirements (e.g., Leader vs. Follower).

To achieve this, we propose extending the **`VPAScope`** API object (introduced in AEP-7942) to support **Pod Label Selectors**. By referencing a `VPAScope` in the VPA specification, users can partition a single Workload Controller into multiple VPA profiles, allowing granular scaling based on dynamic pod roles.

## Motivation

Currently, VPA enforces a "One Size Fits All" model: it aggregates metrics from *all* pods in a `TargetRef` and applies a single recommendation to the entire set.

This works well for stateless web servers (uniform load), but fails for **Stateful** or **Distributed** systems where specific pods experience asymmetric load.
* **The Problem:** In a Leader/Follower database, the Leader might need 8GB RAM while Followers need only 1GB. A standard VPA averages these needs to 4GB, which starves the Leader (Risk of OOM) and wastes resources on the Followers.
* **The Constraint:** Users cannot split these pods into separate Deployments because they must remain in the same StatefulSet ring for data replication and service discovery.



## Use Cases

### 1. Consensus-Based Leaders (Databases)
* **Examples:** Etcd, Redis, Postgres (Patroni), MongoDB.
* **Pattern:** Pods use internal consensus (Raft/Paxos) to elect a Leader.
* **Resource Profile:**
    * **Leader:** High CPU/Memory (Processing all writes + coordination).
    * **Follower:** Medium CPU/Memory (Passive replication + read queries).
* **VPA Role:** Ensures the Leader gets "Peak" resources while Followers get "Baseline" resources to save costs.

### 2. Lease-Based Active/Standby (Controllers)
* **Examples:** Custom Operators, Kubernetes Controllers, Singleton Workers.
* **Pattern:** Pods race to acquire a `coordination.k8s.io/Lease`. The winner becomes "Active"; others wait in "Standby."
* **Resource Profile:**
    * **Active:** 100% utilization (Performing reconciliation).
    * **Standby:** ~0% utilization (Idling/Polling).
* **VPA Role:** Allows the Active pod to scale up to full capacity, while Standby pods remain minimal, maximizing density.

### 3. Data-Based Asymmetry (Hot Shards)
* **Examples:** Elasticsearch, Kafka, Cassandra.
* **Pattern:** No specific "Leader" role, but traffic/data is unevenly distributed across shards.
* **Resource Profile:** Specific pods ("Hot Shards") experience sustained high load due to noisy tenants or large partitions.
* **VPA Role:** Allows operators to target and vertically scale only the specific "Hot Nodes" (e.g., labeled `load=high`) without over-provisioning the entire cluster.

## Proposal

To support heterogeneous workloads, we propose modifying the VPA API to allow partitioning of a Controller's pods. We will achieve this by extending the **`VPAScope`** resource (introduced in AEP-7942) to support pod-level filtering in addition to node-level filtering.

This allows users to define a "Scope" (e.g., "Leader Pods" or "Shard A") and link a VPA object to that scope.

### API Changes

We propose two specific changes to the API surface:

#### 1. VerticalPodAutoscalerSpec
We add an optional `vpaScope` field to the VPA specification. This field references a `VPAScope` object in the same namespace.

```go
type VerticalPodAutoscalerSpec struct {
	// TargetRef points to the controller managing the set of pods.
	// +required
	TargetRef *autoscaling.CrossVersionObjectReference `json:"targetRef"`

	// [PROPOSED]
	// VPAScope points to a VPAScope object that defines the subset of pods 
	// within the TargetRef to be controlled by this VPA.
	//
	// If provided, the VPA manages only the pods that match the criteria
	// defined in the referenced VPAScope.
	//
	// +optional
	VPAScope *VPAScopeReference `json:"vpaScope,omitempty"`
}

// VPAScopeReference contains enough information to locate the referenced scope.
type VPAScopeReference struct {
	// Name of the VPAScope object.
	// The Scope must exist in the same namespace as the VPA.
	// +required
	Name string `json:"name"`
}
```

#### 2. VPAScopeSpec
We extend the VPAScope definition to include a podSelector. This complements the nodeSelector field (added for DaemonSets), allowing the same Scope object to be used for Deployments and StatefulSets.

```go
type VPAScopeSpec struct {
	// [EXISTING - AEP-7942]
	// If specified, the VPA is restricted to pods running on nodes matching this selector.
	// Primarily used for DaemonSet partitioning.
	// +optional
	NodeSelector *metav1.LabelSelector `json:"nodeSelector,omitempty"`

	// [PROPOSED]
	// If specified, the VPA is restricted to pods matching this label selector.
	// Primarily used for Deployment and StatefulSet partitioning (e.g., Leader/Follower).
	//
	// +optional
	PodSelector *metav1.LabelSelector `json:"podSelector,omitempty"`
}
```

#### Mechanism
When a VPA object defines a vpaScope:

The VPA Recommender fetches the referenced VPAScope object.

It identifies the target pods by listing all pods owned by TargetRef.

It filters this list by applying the VPAScope.spec.podSelector (if present).

Only the matching subset of pods are included in the recommendation model.

Pods that do not match the scope are ignored by this specific VPA instance.

### Risks and Mitigations

#### 1. Spec Divergence
**Risk**: A Pod's role may not be known at creation time. Therefore, every new Pod matches the Deployment pod spec and only promotes to "Leader" (and the corresponding VPA profile) after the application starts and wins an election.

**Consequence**: There is an unavoidable window between Election and VPA Actuation where the new Leader is running with default pod spec resources.

##### Mitigation

**Safe Baseline Requests**: Users must configure the Deployment.spec.template.resources.requests to be a "Safe Floor"—sufficient to handle the application's boot sequence and the initial election workload without crashing.

**Reactive Resizing**: The VPA Updater will detect the role=leader label transition and perform an In-Place Update to increase capacity to the "Leader Profile". This feature relies on the **[In-Place Update of Pod Resources](https://github.com/kubernetes/enhancements/issues/1287)** feature. The VPA must be configured to use `updateMode: Auto` with `minReplicas: 1` (or equivalent in-place policy) to ensure the Leader is resized **without restart**, preserving its leadership lease.

#### 2. Conflicting Targets (Overlap & Precedence)
**Context:** Historically, VPA behavior was undefined if multiple VPA objects targeted the same workload. However, with AEP-8026 moving configuration parameters (e.g., Safety Margins, History Length) to the VPA Custom Resource, users will increasingly require multiple VPA objects to apply different scaling policies to different subsets of a single Deployment.

**Risk:** A Pod might match multiple Active VPAs simultaneously. For example, a pod labeled `role=leader` matches both a global VPA (Selector: `nil`) and a specific VPA (Selector: `role=leader`).

**Mitigation: Specificity Precedence**
To support granular configuration without conflict, the VPA Updater will resolve overlaps using **Scope Specificity**, similar to Kubernetes Network Policies or CSS:

1.  **Rule:** If a Pod matches multiple VPAs, the VPA with the **Most Specific Scope** takes precedence.
    * *Level 0 (Global):* `selector: null` (Matches everything in TargetRef).
    * *Level 1 (Refined):* `selector: {key: value}` (Matches specific subset).
2.  **Resolution:** The Updater will respect the recommendations from the *Level 1* VPA and ignore the *Level 0* VPA for that specific pod.
3.  **Tie-Breaking:** If two VPAs have equally specific selectors (e.g., VPA-A selects `role=leader` and VPA-B selects `tier=gold`), and a pod matches *both*, this is considered a misconfiguration. The Updater will deterministically break the tie by choosing the VPA with the oldest `CreationTimestamp` and emit a Warning Event.

#### 3. Recommender Data Sparsity
**Risk**: Partitioning a set of pods reduces the sample size for the histogram (e.g., a single Leader pod).

**Mitigation**: The VPA object acts as a persistent store for the *Role*, rather than the specific Pod instance.
* **Inheritance:** When leadership rotates (e.g., `pod-0` steps down and `pod-1` becomes Leader), `pod-1` immediately inherits the historical resource profile built by `pod-0`.
* **Result:** The VPA aggregates data at the **Scope Level**, ensuring that the "Leader Profile" is continuous and robust over time, even if the specific Pod holding the title changes frequently.

#### Unmanaged Pods
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
