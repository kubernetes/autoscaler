# Design Blueprint for Karpenter-Cluster-Autoscaler Integration

This document outlines the architectural decisions, data structures, and state mapping logic for the Karpenter Simulator integration inside Cluster Autoscaler (CA). It represents the production-grade target state designed to eliminate edge cases around asymmetric capacity limits and custom label matching loops.

**Note on Integration Scope:** This proposal does not replace or invalidate existing Cluster Autoscaler logic. Instead, it introduces an **opt-in mode of operation** (enabled via feature flags) that supercharges CA's decision-making in supported environments. When disabled, Cluster Autoscaler continues to function using its traditional predicate-based simulation logic.

---

## High-Level Architecture

The Karpenter Simulator acts as an **in-memory, high-fidelity scheduling solver** embedded directly inside the CA scale-up orchestrator.

### 1. Global Opt-In & Package Organization
- **Feature Flag:** The integration is governed by a global `KarpenterSimulatorEnabled` flag in `AutoscalingOptions`.
- **Interface-Driven Design:** To enable seamless swapping of scheduling logic, two primary interfaces are used:
    - **`PodSchedulingSimulator`**: Standardizes the "Can these pods fit on existing nodes?" contract. Used by **Scale-Down** and **Scale-Up Pre-checks**.
    - **`ScaleUpSimulator`**: Standardizes the "What new nodes do we need?" contract. Used by the **Scale-Up Orchestrator**.
- **Decoupled Wiring:** The appropriate implementation (`Native` or `Karpenter`) is injected at initialization time based on the `KarpenterSimulatorEnabled` flag.

### 2. Core Execution Loop (Scale-Up):
1. **Translation:** CA's registered physical node groups are clustered and mapped to Karpenter `NodePool`, `InstanceType`, and `Offering` primitives.
2. **Iterative Simulation:** Karpenter's binpacking solver is executed in-memory on the pending pod list in batches. It leverages a zero-network, mock client (`DirectClient`) populated with the current cluster snapshot.
3. **Resolving, Pruning & Commitment:** Karpenter's virtual scheduling decisions (`NodeClaim`s) are mapped back to real, physical CA `NodeGroup`s. For every successful physical scale-up, the new nodes and pods are **committed back to the in-memory `ClusterSnapshot`**, ensuring cross-batch constraint awareness. Exhausted node groups are pruned from the offering set, and overflow pods are recycled for the next batch.
4. **Execution:** Coordinated multi-group scale-ups are packaged into a `CompositeNodeGroup` for CA's scale-up actuation.

---

## Documented Limitations of the Karpenter-Integrated Mode

The opt-in Karpenter-integrated mode provides significant performance and binpacking improvements but currently operates with the following documented limitations:

1.  **NodePool Fragmentation (DaemonSet Variances):** Because the NodePool grouping key incorporates DaemonSet scheduling signatures (HostPorts and Labels), variances in these configurations between physical node groups will force them into separate virtual NodePools. This fragmentation reduces Karpenter's ability to binpack across those groups, potentially leading to suboptimal scale-up decisions compared to a unified pool.
2.  **Scheduler Fidelity Drift:** There is a risk of "Fidelity Drift" if Karpenter's internal scheduling logic differs from the specific version or custom configuration of the cluster's control-plane scheduler.
3.  **Bypassed Schedulers Restriction:** This mode requires `BypassedSchedulers` to be empty. This is required because if unprocessed pods are sent to the simulator, Karpenter might decide to provision a new node that perfectly matches the pod's soft preferences. In the meantime, the real scheduler might bind that pod to an existing node that satisfies strict requirements but not the soft preferences. This race condition would result in an unneeded physical scale-up (overscaling).
4.  **No Native DRA (Dynamic Resource Allocation) Support:** Karpenter's scheduler requires a `dynamicresources.Allocator` to simulate DRA device allocation. Because this allocator is not wired in the integration (passed as `nil`), pods with DRA requests (referencing `ResourceClaims`) cannot be scheduled by Karpenter. The simulator automatically detects these pods and falls back to CA's native predicate-based simulation for them.

---

## Unified Scheduling Consistency (Scale-Down & Pre-Checks)

To prevent scale-down/scale-up oscillation and ensure high-fidelity filtering, existing-node scheduling checks are unified with the Karpenter core.

### 1. Karpenter-Backed Rescheduling Simulation
- **Decision:** The native CA scale-down `RemovalSimulator` and the `filterOutSchedulable` processor are updated to use the Karpenter solver via the **`PodSchedulingSimulator`** interface.
- **Implementation:** The **`KarpenterReschedulingSimulator`** is injected into these components.
- **Existing-Only Mode:** For these checks, the simulator invokes the Karpenter solver with **zero `NodePools`**.
- **Predicate-Aware Placement (Reconciliation):** To ensure Karpenter respects CA's removal logic (specifically preventing pods from scheduling onto the node being removed):
    - **Topology Visibility:** Karpenter's **Topology Engine** (`state.Cluster`) is initialized with the **full set of nodes** from the `ClusterSnapshot`. This ensures all pod anti-affinity and topology spread constraints are mathematically accurate across the entire cluster.
    - **Filtered Placement Targets:** Karpenter's **Placement Engine** (`Scheduler.stateNodes`) is initialized using a **filtered slice of nodes** that pass the `IsNodeAcceptable` predicate provided in `SchedulingOptions`. This restricts the solver to only consider nodes designated as valid rescheduling targets, preventing it from incorrectly placing pods on nodes marked for removal.
- **Zero-Dependency Simplicity:** Unlike the scale-up simulator, the rescheduling simulator does **not** require the `KarpenterConverter` or `ExpanderStrategy`.
- **Lightweight Performance:** This mode skips heavy binpacking and cost optimization, ensuring that CA's performance during heavy scale-down evaluations remains comparable to native predicate checks.

---

## Primitives Conversion (`KarpenterConverter`)

To optimize solver performance and cost-fidelity, physical NodeGroups are collapsed into virtual Karpenter primitives without losing scheduling constraints.

### 1. NodePool Construction
- Physical `NodeGroup`s are grouped by **strictly taint compatibility**. All node groups sharing identical taints are mapped under a single virtual `NodePool` (named `pool-<taints-hash>`).
- **Heterogeneous Scaling:** This prevents virtual `NodePool` fragmentation. Differences in system daemons or labels are handled at the `InstanceType` level, leveraging Karpenter's ability to binpack across diverse types within a pool.
- All custom labels defined across the constituent node groups are aggregated and declared as permissive requirements on the `NodePool` spec.

### 2. Clustered InstanceTypes (Optimization & Fidelity)
To balance simulation performance with high-fidelity scheduling, physical `NodeGroup`s are clustered into virtual `InstanceType` objects based on their **Scheduling Footprint**.
- **Clustering Key:** Physical groups are collapsed into a single virtual `InstanceType` if they share:
    1.  The same **Instance Type label** (e.g., `m5.large`).
    2.  An identical set of compatible **DaemonSet pods**.
- **Common-Case Optimization:** In typical clusters where system daemons target broad categories (e.g., `linux` or `amd64`), all zonal ASGs of the same family will share the same compatible DS set and collapse into a single virtual `InstanceType`. This keeps the solver search space small and fast.
- **Fidelity for Edge Cases:** If a specific `NodeGroup` has unique labels targeted by a DaemonSet's `nodeSelector`, it will have a different "compatible DS set" than its peers. The converter will correctly detect this and provision a distinct virtual `InstanceType` for that group, ensuring 100% HostPort and resource fidelity even in complex environments.
- **Metadata Extraction:** 
    - **Base Overhead:** Each clustered `InstanceType` sets its `Overhead` (KubeReserved, SystemReserved, EvictionThreshold) based strictly on the values extracted from the physical node template. **DaemonSet resource requests must NOT be baked into this overhead.**
    - **Platform Labels:** Standard labels (Arch, OS, CapacityType) are extracted from the templates with safe fallbacks.
    - **Volume Limits:** Driver-specific limits are extracted from `CSINode` metadata. If limits differ between node groups that would otherwise be clustered, they are separated into distinct `InstanceType`s to ensure binpacking accuracy.

### 3. 1:1 Offering Mapping (Identity & Limits)
Within each clustered `InstanceType`, physical `NodeGroup`s are mapped 1:1 to virtual **`Offering`s**:
- **Identity Retention:** Mapping 1:1 at the `Offering` level allows the resolver to enforce `MaxSize` limits by marking specific physical groups as `Available = false` without disabling the entire instance family cluster.
- **Custom Label Routing:** Each `Offering` inherits the specific identifying labels of its physical `NodeGroup` (e.g., zones, custom ASG tags). Karpenter's solver uses these requirements during binpacking to ensure pods with custom selectors are routed to the correct physical providers, even within a shared `InstanceType`.
- **Accurate Pricing:** Each `Offering` is populated with a precise price from CA's native `PricingModel`, ensuring the solver makes cost-optimized decisions that reflect the real cloud environment.


#### A. Multi-Dimensional Pricing
- **Requirement:** Karpenter's solver is a cost-optimizing binpacker. To ensure simulated decisions match real-world costs, each `Offering` must be populated with a precise price.
- **Pricing Source:** Prices should be retrieved from the CA native pricing model: `autoscalingCtx.CloudProvider.Pricing().NodePrice(node, startTime, endTime)`.
- **Pre-provisioning Price Lookup:** To retrieve a price for an offering before it is provisioned, the converter must construct a **mock `apiv1.Node` template** representing the offering and pass it to the `NodePrice` method. The mock node must have its `Status.Capacity` and `Labels` populated to match the virtual `InstanceType` and `Offering` primitives.
- **Weighted Fallback:** If the pricing model is unavailable, prices must be calculated using a **weighted resource cost function** (e.g., `CPU_Weight * CPU + Mem_Weight * Memory + GPU_Weight * GPU`).
- **Spot Discount Calculation:** For offerings with the `spot` capacity type, the calculated price must be multiplied by a **Spot Discount coefficient of 0.3** (e.g., `Price = Price * 0.3`), representing a 70% discount from the On-Demand price. This coefficient must be mathematically locked to ensure consistent cost-optimization across different instance families.

#### B. Requirement Matching Matrix
Each `Offering` inherits **all labels** from its representing physical `NodeGroup`:
- **Well-Known Labels:** Declared with actual combined values at the NodePool level, and specific values at the Offering level.
- **Custom / Ignored Labels:** 
  - **NodePool Level:** Add a permissive `NotIn [ca-ignore-reserved-value]` requirement for every unique custom label key. Using a syntactically valid but reserved dummy value ensures compatibility with all internal Kubernetes label parsers.
  - **Offering Level:** For NodeGroups possessing the label, use `In [value]`. For those lacking it, use an explicit `DoesNotExist` operator. 

This strategy ensures Karpenter correctly selects compatible offerings for positive, negative, and neutral pod selectors.

---

## In-Memory Simulation State (`runKarpenterSimulation`)

### 1. Zero-Overhead Mock Client (`DirectClient`)
Karpenter's scheduling and state engines depend on `client.Client` to interact with Kubernetes objects. To eliminate network latency and ensure state consistency:
- **Dynamic Facade Architecture:** We implement a lightweight, in-memory `DirectClient` that acts as a **dynamic, read-only facade** over CA's native **`ClusterSnapshot`**.
- **Unified State Source:** All Kubernetes objects used by the autoscaler during simulation must be encapsulated within the `ClusterSnapshot`. This ensures a single source of truth and high-fidelity simulations.
- **Full Metadata Registry:** The `ClusterSnapshot` is extended to track and serve **`PersistentVolume`**, **`PersistentVolumeClaim`**, **`StorageClass`**, and **`CSINode`** objects. 
- **State Synchronization:** The `DirectClient` retrieves these objects directly from the `ClusterSnapshot` on every `Get` or `List` call. This ensures that the simulation is fully self-contained and that any snapshot updates (e.g., between iterative batches) are immediately visible to the Karpenter solver.
- **High-Fidelity Placement Visibility:** Karpenter's solver inherits CA's awareness of all pods, including **nominated pods** and pods on **upcoming nodes**, as they are served by the `DirectClient` directly from the `ClusterSnapshot`.
- It intercepts `Get` and `List` requests from Karpenter's scheduler, mapping them to the current CA snapshot state with zero network or serialization overhead.

### 2. Dynamic DaemonSet Topology & Capacity Evaluation
To ensure Karpenter accurately evaluates HostPorts and PodAntiAffinity (PAA) constraints for system DaemonSets:
- **Topology Evaluation & Template Extraction:** 
  1. The simulator loops over all CA template `nodeInfos` from all physical node groups and identifies DaemonSet or DaemonSet-like pods using CA's native `podutils.IsDaemonSetPod(pod)` utility.
  2. For every matching pod, it creates a template clone, clearing the `Spec.NodeName` field. **Resource requests are preserved in the clone to allow Karpenter to calculate daemon overhead accurately.**
- **Aggregation with Native Partitioning:**
  - The simulator aggregates all unique DaemonSet templates globally across the cluster.
  - These templates are passed directly to `scheduling.NewScheduler` as the `daemonSetPods` slice.
  - **Native InstanceType-Aware Matching:** Leveraging Karpenter's native improvements, the scheduler automatically evaluates the compatibility of each DaemonSet pod against every virtual `InstanceType` within a pool. 
  - The scheduler partitions instance types into "Daemon Overhead Groups," ensuring that HostPorts and resource overheads from a specific DaemonSet only affect compatible instance families (e.g., matching OS, Arch, or custom labels).
- **Karpenter Registration (Not via DirectClient):** 
  - To prevent Karpenter from treating these templates as standard pending/unschedulable user pods, **they are NOT registered with the `DirectClient` cluster snapshot.**
  - Karpenter's scheduler uses the `daemonSetPods` slice to automatically project synthetic pods onto simulated `NodeClaim`s, correctly accounting for their resources and constraints during binpacking.

### 3. Volume Topology Pre-calculation
To ensure pods requesting zonal storage are placed on offerings in the correct zone:
- Before each iteration of the solver loop, the simulator uses Karpenter's `VolumeTopology.GetRequirements(ctx, pod)` utility to calculate the topology requirements imposed by the pod's PVCs.
- These requirements are passed to the `NewScheduler` via the `volumeReqsByPod` map.
- Karpenter's solver intersects these volume requirements with the requirements of candidate `Offering`s during binpacking, ensuring zonal storage compatibility.

### 4. Unified Topology State
- The live cluster topology (including physical nodes, pending unready nodes, and all pods already bound to them) is retrieved via `autoscalingCtx.ClusterSnapshot.ListNodeInfos()`.
- Upcoming nodes currently scaling up are treated identically to real nodes.
- Pending/unschedulable pods passed to the function are treated strictly as the input (`remainingPods`) to Karpenter's scheduler solver.

### 5. Iterative Batch Solving Loop
To solve the **Zonal Capacity Blindness** problem and maintain **cross-batch scheduling constraint awareness**, the simulation executes in an iterative loop:
- **Pre-Simulation Pruning (Performance Fast-Path):** Before the first solver iteration, the converter must automatically disable any virtual `Offering` whose corresponding physical `NodeGroup` has already hit its `MaxSize` limit. This prevents the solver from wasting cycles attempting to binpack pods onto exhausted capacity.
- **Batching:** Pods are processed in batches (e.g., 1000 pods per iteration).
- **Solve & Map:** Karpenter `Solve()` is called for the current batch. The resulting `NodeClaims` are immediately mapped to physical `NodeGroups` via the resolver.
- **Handling Partial Success & State Commitment:**
    - **Physical Balancing:** The resolver calls `BalanceScaleUpBetweenGroups` to generate a physical scale-up plan based on `MaxSize` limits.
    - **Snapshot Commitment:** For every node added to the physical plan, the simulator **updates the in-memory `ClusterSnapshot`** by adding a corresponding "simulated" node and binding its assigned pods to it. 
    - **State Continuity:** This ensures that in the next iteration, Karpenter's scheduler (which pulls from the updated snapshot via `DirectClient`) is aware of the pods scheduled in previous batches, enabling correct evaluation of `podAffinity`, `podAntiAffinity`, and `topologySpreadConstraints`.
    - **Recycling Overflow:** Pods belonging to "overflow" `NodeClaims` (those that couldn't be physically provisioned) are returned to the pending pool for the next iteration.
    - **Offering Pruning:** Physical node groups that hit their limits are disabled for all subsequent iterations in the same simulation cycle.
- **Termination:** The loop continues until all pods are either successfully scheduled or all compatible offerings have been exhausted.

---

## NodeClaim -> NodeGroup Resolving (`mapResultsToOptions`)

Resolving is the process of mapping Karpenter's virtual scheduling decisions (`NodeClaim`s) back to real-world CA physical primitives. To ensure high-fidelity simulation and correct limit enforcement, this resolution happens **within the batch simulation loop** immediately after each solver batch run.

The resolution flow for a batch of `NodeClaims` is as follows:

### 1. Grouping by Constraints
The solver outputs a list of virtual `NodeClaims` for the batch. We group these claims by their exact scheduling `Requirements` into slices of identical claims: `[]*scheduling.NodeClaim`. This preserves the individual claim identity and its associated pods.

### 2. Converting to Expansion Options
For each group of claims (slice) of size `count = len(slice)`:
- We identify the set of physical `NodeGroup`s that match the claim's requirements.
- We partition this set into `expander.Option`s by grouping similar node groups (using `NodeGroupSetProcessor.FindSimilarNodeGroups`).
- If similar groups also match the claim requirements, they are added to the `SimilarNodeGroups` list of the option.
- The `NodeCount` of the option is set to `count`.
- The `Pods` of the option are the consolidated list of all pods assigned to the claims in the slice.

### 3. Selecting the Best Option (Expander)
We use CA's active `ExpanderStrategy` (e.g. price-based) to select the best `expander.Option` for this group of claims. This ensures we prioritize the most cost-effective physical node groups.

### 4. Balancing & Limit Enforcement (Balancer)
The chosen option (with its target `NodeGroup` and `SimilarNodeGroups`) and the target `count` are passed to the limit checker and balancer:
- **Global Quotas:** We apply global resource limits (e.g., CPU/Memory quotas) using the tracker's `CheckDelta` method.
- **Local Limits:** We distribute the capped count across the similar node groups using `BalanceScaleUpBetweenGroups`, which automatically caps the scale-up if any group hits its `MaxSize` limit.
- This yields a physical scale-up plan (`[]ScaleUpInfo`) with a `finalAllowedCount`.

### 5. Capping & Snapshot Commitment
- If `finalAllowedCount < count` (due to limits), we cap the `NodeClaim` slice: `slice = slice[:finalAllowedCount]`.
- The pods from the remaining claims in the capped slice are simulated as scheduled. We commit these simulated nodes and pods to the in-memory `ClusterSnapshot`. This ensures subsequent batches are aware of these scheduled pods (crucial for affinity/anti-affinity rules).
- The pods from the discarded claims (`slice[finalAllowedCount:]`) are recycled and added back to the pending pool to be retried in the next batch.

### 6. Consolidation for Main Loop
All physical scale-up plans generated across all batches are accumulated. At the end of the simulation, they are consolidated, wrapped in a single `CompositeNodeGroup`, and returned as a single `expander.Option` to CA's main loop to ensure they are all executed atomically.


---

## Processor & Pipeline Disablement

To guarantee optimal performance and prevent scheduling loops:

### 1. Disablement of the `filterOutSchedulable` Phase
- **Decision:** The standard CA `filterOutSchedulable` phase must be **disabled**.
- **Rationale:** Karpenter's simulator performs a high-fidelity scheduling simulation over the entire cluster snapshot. This is safe because `Solve()` sees all nodes in the snapshot—including injected "upcoming" nodes—and naturally simulates scheduling pending pods onto them.
- **Fidelity Risk:** While highly efficient, this introduces a risk of **"Fidelity Drift"**. If Karpenter's solver logic differs even slightly from the real Kubernetes scheduler (e.g., regarding specific alpha features or custom scheduler plugins), pods could be incorrectly marked as schedulable or unschedulable.
- **Safety Requirement (BypassedSchedulers):** To prevent overscaling due to races between the real scheduler and the simulator, the **`BypassedSchedulers` map must be empty** when this phase is disabled. This ensures CA only processes pods that the real scheduler has explicitly marked as `Unschedulable`. This is critical because if unprocessed pods are sent to the simulator, Karpenter might decide to provision a new node that perfectly matches the pod's soft preferences. In the meantime, the real scheduler might bind that pod to an existing node that satisfies strict requirements but not the soft preferences. This race condition would result in an unneeded physical scale-up.

### 2. Disablement of ProvisioningRequest Handling
- **Decision:** ProvisioningRequest handling is strictly disabled to prevent conflicts with Karpenter-driven scheduling.
- **Enforcement:** The `AutoscalerBuilder` will fail to initialize and return a configuration error if both `KarpenterSimulatorEnabled` and `ProvisioningRequestEnabled` are enabled.

### 3. Simulation Preemption Model
- **Decision:** The integrated solver performs **Binpacking Only**. It does not simulate preemption of lower-priority pods to make room for higher-priority pods.
- **Rationale:** This is consistent with CA's native scale-up simulation. Preemption victims and nominated node assignments are handled by CA's native pre-processing logic and are reflected in the `ClusterSnapshot` before the solver is invoked.

### 4. Preserved Processors
- All other active CA pod list processors—including **proactive scale-up** and **capacity buffers support**—must remain fully enabled.

---

## Validation Rules & Implementation Workflow

To ensure the technical integrity and performance of the Karpenter integration, all implementation work must adhere to the following three-phase validation lifecycle:

1.  **Component-Level Unit Testing (Mandatory First Step):** Every new component, interface implementation, or logic refinement (e.g., `DirectClient` facade, `KarpenterConverter`, or `OfferingPruningTracker`) must be developed alongside exhaustive unit tests. Implementation is only considered started once these localized tests are passing.
2.  **Aggregate Integration Testing:** Only after all individual components are verified should the entire integration be tested in aggregate. This phase must confirm that the wiring between the Scale-Up Orchestrator, Scale-Down Planner, and the Karpenter solver works seamlessly without regressions in standard CA behavior.
3.  **Performance Benchmarking:** Once functional correctness is confirmed, the implementation must be validated using the project's benchmark suite. The goal is to empirically demonstrate that the Karpenter-integrated mode provides superior binpacking efficiency and/or computational performance compared to the standard predicate-based Cluster Autoscaler.

---

## Verification & Testing Strategy

The unit and integration tests must exhaustively verify the following core scheduling and balancing scenarios:

### 1. Custom Label Routing (Positive Match)
- **Scenario:** NodeGroup A has label `my-label: foo`. NodeGroup B lacks it. Pod requires `my-label: foo`.
- **Expected Outcome:** Karpenter selects Offering A. Scale-up of NodeGroup A only.

### 2. Custom Label Negative Routing (Negative Match)
- **Scenario:** NodeGroup A has label `my-app: target`. NodeGroup B lacks it (mapped with `DoesNotExist`). Pod requires `my-app NotIn [target]`.
- **Expected Outcome:** Karpenter solver pairs pod with Offering B. Scale-up of NodeGroup B only.

### 3. Custom Label Ignored Routing (Neutral Match)
- **Scenario:** NodeGroup A has label `my-label: value`. NodeGroup B lacks it. Pod has no selector for `my-label`.
- **Expected Outcome:** Karpenter can solve and place the pod on either offering.

### 4. Limit-Aware Balancing (Iterative Spillover)
- **Scenario:** NodeGroup A (Cheaper, MaxSize=10) and NodeGroup B (Expensive, MaxSize=1000) are compatible. Both have 0 current nodes. 100 pods are pending.
- **Expected Outcome:** 
    - Iteration 1: Karpenter solves all 100 pods onto NodeGroup A. Resolver caps at 10 nodes. 10 nodes provisioned for Group A. 90 pods recycled. Group A offering pruned.
    - Iteration 2: Karpenter solves remaining 90 pods onto NodeGroup B (as A is now pruned). 90 nodes provisioned for Group B. 
    - Result: Consolidated scale-up of 10 nodes in A and 90 nodes in B.

### 5. Multi-Zone Reliability Balancing (Cross-Group Distribution)
- **Scenario:** Three physical NodeGroups (Zone-A, Zone-B, Zone-C) are identified as "similar" by CA. Karpenter solves a single virtual `NodeClaim` for 30 nodes of that capacity shape.
- **Expected Outcome:** The resolver correctly invokes `BalanceScaleUpBetweenGroups`. Verifies that the final `CompositeNodeGroup` contains a distribution (e.g., 10 nodes in A, 10 in B, 10 in C) instead of all 30 nodes being assigned to a single group. This confirms the integration preserves CA's zonal reliability guarantees.

### 6. Cross-Batch Scheduling Affinity
- **Scenario:** Pod A and Pod B have a strict `podAffinity` with each other. The batch size is 1. Pod A is processed in Batch 1 and placed in `zone-a`. Pod B is processed in Batch 2.
- **Expected Outcome:** Snapshot commitment after Batch 1 ensures the `ClusterSnapshot` contains Pod A on a simulated node in `zone-a`. Karpenter's solver for Batch 2 correctly evaluates the affinity and places Pod B in the same zone.

### 7. Dynamic DaemonSet Topology & Capacity Evaluation
- **Scenario:** NodeGroup templates carry system DaemonSets. A user pod requests a HostPort occupied by the DS, or has an Anti-Affinity against it.
- **Expected Outcome:** Karpenter correctly prevents the user pod from scheduling onto the conflicting offering, while accurately pricing capacity via `InstanceType` overhead.

### 8. Storage Topology Routing
- **Scenario:** Pod has a PVC bound to a PV in `us-east-1a`. 
- **Expected Outcome:** `VolumeTopology` pre-calculation injects `topology.kubernetes.io/zone In [us-east-1a]` into the pod's volume requirements. Karpenter solver restricts binpacking for this pod to offerings in `us-east-1a`.

### 9. Smallest Instance Type Resolution (Cost-Optimization)
- **Scenario:** Karpenter generates a virtual `NodeClaim` that is compatible with multiple physical `NodeGroup`s of different sizes (e.g., `m5.large` and `m5.4xlarge`).
- **Expected Outcome:** The resolver utilizes the `ExpanderStrategy` to select the `m5.large` (cheapest/smallest) physical group, ensuring that Cluster Autoscaler preserves its cost-optimization mandate even when using Karpenter's flexible scheduling outputs.

---

## Implementation Considerations

This section documents the technical constraints and concrete implementation patterns required to ensure a successful, crash-free delivery.

### 1. Physical Implementation Signatures

#### A. PodSchedulingSimulator Interface
Used for existing-node checks (Scale-Down, Pre-checks). Defined in `simulator/clustersnapshot`:
```go
type PodSchedulingSimulator interface {
	TrySchedulePods(snapshot ClusterSnapshot, pods []*apiv1.Pod, breakOnFailure bool, opts SchedulingOptions) ([]Status, int, error)
}
```
**Circular Dependency Resolution:** The **`Status` struct** (containing `Pod *apiv1.Pod` and `NodeName string`) must be **relocated from `simulator/scheduling` to `simulator/clustersnapshot`**. This ensures that the interface can be defined in the neutral `clustersnapshot` package without requiring an import of the `scheduling` package, which already depends on `clustersnapshot`.

#### B. ScaleUpSimulator Interface
Used for binpacking and expansion (Scale-Up). Defined in `core/scaleup/orchestrator`:
```go
type ScaleUpSimulator interface {
	Simulate(
		autoscalingCtx *ca_context.AutoscalingContext,
		podEquivalenceGroups []*equivalence.PodGroup,
		unschedulablePods []*apiv1.Pod,
		nodes []*apiv1.Node,
		nodeGroups []cloudprovider.NodeGroup,
		nodeInfos map[string]*framework.NodeInfo,
		tracker *resourcequotas.Tracker,
		now time.Time,
		allOrNothing bool,
	) ([]expander.Option, map[string]status.Reasons, map[string][]estimator.PodEquivalenceGroup, error)
}
```

#### C. RemovalSimulator Constructor
Updated to accept the `PodSchedulingSimulator` interface. Defined in `simulator/cluster.go`:
```go
func NewRemovalSimulator(
	listers kube_util.ListerRegistry,
	clusterSnapshot clustersnapshot.ClusterSnapshot,
	deleteOptions options.NodeDeleteOptions,
	drainabilityRules rules.Rules,
	persistSuccessfulSimulations bool,
	podSchedulingSimulator clustersnapshot.PodSchedulingSimulator,
) *RemovalSimulator
```

#### D. KarpenterConverter Interface
Only required for the `ScaleUpSimulator`. Defined in `simulator/karpenter`:
```go
type KarpenterConverter interface {
	Convert(nodeGroups []cloudprovider.NodeGroup, nodeInfos map[string]*framework.NodeInfo) ([]*karpenterv1.NodePool, map[string][]*karpentercloudprovider.InstanceType)
	ITNameToNodeGroups() map[string][]cloudprovider.NodeGroup
	ITNameToPool() map[string]string
	GetPhysicalITName(labels map[string]string, defaultName string) string
}
```

#### E. Native CA Balancing Call
The resolver must invoke the following function from `k8s.io/autoscaler/cluster-autoscaler/processors/nodegroupset`:
```go
func (b *BalancingNodeGroupSetProcessor) BalanceScaleUpBetweenGroups(autoscalingCtx *ca_context.AutoscalingContext, groups []cloudprovider.NodeGroup, newNodes int) ([]ScaleUpInfo, errors.AutoscalerError)
```

### 2. Package Organization & Decoupling
To prevent cyclic import dependencies between `core/scaleup/orchestrator` and `simulator/scheduling`, the architectural core is decoupled via the `PodSchedulingSimulator` interface:
- `simulator/clustersnapshot`: Holds the interface and shared options.
- `simulator/scheduling`: Implements `NativeHintingSimulator`.
- `simulator/karpenter`: Implements `KarpenterReschedulingSimulator` and `KarpenterScaleUpSimulator`.
- **Decoupled Dependency Injection:** The **`AutoscalingContext`** (in `context/autoscaling_context.go`) must be updated to include the **`PodSchedulingSimulator`** interface. 
- **Wiring Layer:** During the initialization of the Cluster Autoscaler (in `core/static_autoscaler.go`), the appropriate simulator implementation is instantiated based on the `KarpenterSimulatorEnabled` flag and assigned to the context.
- **Scale-Down Integration:** The `ScaleDownPlanner` (in `core/scaledown/planner/planner.go`) retrieves the simulator from the `AutoscalingContext` and passes it to the `RemovalSimulator` constructor, ensuring zero direct dependency on specific simulator implementations within the scale-down logic.
- **Scale-Up Integration:** The Scale-Up Orchestrator is similarly updated to receive both the `PodSchedulingSimulator` (for pre-checks) and the `ScaleUpSimulator` (for binpacking) via its constructor options.

### 3. ClusterSnapshot Extensions & Metadata Sourcing
To support storage-aware simulation, the `ClusterSnapshot` must be extended to track PV, PVC, StorageClass, and CSINode objects using the `common.PatchSet` pattern:
- **Internal Maps:**
  - `pvs *common.PatchSet[string, *apiv1.PersistentVolume]`
  - `pvcs *common.PatchSet[string, *apiv1.PersistentVolumeClaim]`
  - `storageClasses *common.PatchSet[string, *storagev1.StorageClass]`
  - `csiNodes *common.PatchSet[string, *storagev1.CSINode]`
- **Interface Methods:** The `ClusterSnapshot` interface must be updated with the following methods to enable `DirectClient` facade lookups:
  - `GetPV(name string) (*apiv1.PersistentVolume, error)`
  - `ListPVs() ([]*apiv1.PersistentVolume, error)`
  - `GetPVC(namespace, name string) (*apiv1.PersistentVolumeClaim, error)`
  - `ListPVCs() ([]*apiv1.PersistentVolumeClaim, error)`
  - `GetStorageClass(name string) (*storagev1.StorageClass, error)`
  - `ListStorageClasses() ([]*storagev1.StorageClass, error)`
- **Sourcing:** These objects are encapsulated within the `ClusterSnapshot` during `SetClusterState`.

### 4. Concrete Field Traversals & Fallbacks

#### A. Extracting CSI Volume Limits
To populate `InstanceType.VolumeLimits`, the converter must traverse the `CSINode` metadata in the physical `NodeInfo`:
```go
// Implementation Logic:
if ni.CSINode != nil {
    for _, driver := range ni.CSINode.Spec.Drivers {
        if driver.Allocatable != nil && driver.Allocatable.Count != nil {
            instanceType.Capacity[apiv1.ResourceName(driver.Name)] = *resource.NewQuantity(int64(*driver.Allocatable.Count), resource.DecimalSI)
        }
    }
}
```

#### B. Dynamic Platform Metadata Fallbacks
If node templates lack standard labels, the following default mapping must be applied:
- `kubernetes.io/arch`: Default to `amd64`.
- `kubernetes.io/os`: Default to `linux`.
- `karpenter.sh/capacity-type`: Default to `on-demand`.

### 5. Karpenter State Instantiation & Filtering
Karpenter's solver requires high-level state objects. These must be instantiated using the `DirectClient` facade:
- **state.Cluster:** Initialized via `state.NewCluster(clk, directClient, nil)`. The `CloudProvider` interface is `nil` as state is populated manually.
- **Manual State Population:** The simulator must manually populate the `state.Cluster` by iterating over the `ClusterSnapshot` nodes and calling `cluster.UpdateNode(ctx, node)`. This triggers Karpenter's internal logic to fetch pods and CSI metadata via the `DirectClient` facade.
- **StateNode Collection:** After hydration, the `*state.StateNode` objects must be collected from the `cluster.Nodes()` iterator to form the `stateNodes` slice required by the `NewScheduler` constructor.
- **Predicate-Aware Filtering (Scale-Down):** In the rescheduling simulator, the `stateNodes` slice passed to `NewScheduler` must be filtered by the `IsNodeAcceptable` predicate from `SchedulingOptions`. This ensures that pods being evicted are never "scheduled" back onto nodes marked for removal (or other excluded nodes), while still allowing the `Topology` engine (using the full `cluster`) to account for anti-affinity against all pods.
- **Topology Engine:** Initialized via `scheduling.NewTopology(ctx, directClient, cluster, stateNodes, ...)`.

### 6. DirectClient (client.Client) Safety
The `DirectClient` acts as a **read-only facade**. To ensure the Karpenter solver does not attempt to mutate cluster state:
- **Read Methods (Get, List):** Dynamically query the `ClusterSnapshot`.
- **Write Methods (Create, Update, Delete, Patch):** Must return **`apierrors.NewMethodNotSupported`**.

### 7. Pricing Fallback Weights (Coefficients)
When cloud provider pricing is unavailable, the weighted cost function must use these standard coefficients:
- **CPU_Weight:** `1.0` (Base cost per 1 vCPU)
- **Mem_Weight:** `0.125` (Cost per 1 GiB; standard 1:8 vCPU-to-GiB ratio)
- **GPU_Weight:** `25.0` (Cost per 1 GPU)
- **Spot_Multiplier:** `0.3` (70% discount)

### 8. Hashing & Naming Stability
To generate stable, unique identifiers for virtual primitives while respecting Kubernetes name length limits (63 chars):
- **Algorithm:** Use **`FNV-1a` (32-bit)** hashing on the serialized grouping keys.
- **Formatting:** Hashes must be represented as lowercase hexadecimal strings.
- **Stability:** The grouping keys (Taints, HostPorts, Labels) must be **lexicographically sorted** before hashing.

### 9. Offering Pruning & Identity
- **Pruning Tracker:** A `map[string]*cloudprovider.Offering` (keyed by physical `NodeGroup.Id()`) must live inside the `runKarpenterSimulation` function's scope.
- **Limit Enforcement:** When the resolver maps a virtual claim back to a physical group and detects it has hit `MaxSize`, it marks `Available = false` in this map.

### 10. Simulated Node Identity
- **Naming Convention:** `simulated-node-<nodegroup-id>-<uuid>`.
- **Hostname Alignment:** The simulated node's `kubernetes.io/hostname` label must be explicitly set to match its generated name.

### 11. Error Translation & Partial Success
To integrate Karpenter's results back into CA's status tracking:
- **Partial Success:** Pods that failed to schedule (present in `Results.PodErrors`) must be reported using CA's `status.Reasons`.
- **Mapping:** 
  - If a pod failure is due to a constraint, it should be mapped to `status.BackoffReason`.
  - If the solver returns a terminal `error`, it must be wrapped as a `status.ScaleUpError` with `status.InternalError` type.

---

## Technical Reference & Code Mapping

This appendix provides a direct mapping between the architectural concepts in this design and their physical locations in the codebase.

### 1. Key Data Structures (Karpenter)
| Type | File Path | Usage |
| :--- | :--- | :--- |
| `v1.NodePool` | `vendor/sigs.k8s.io/karpenter/pkg/apis/v1/nodepool.go` | Primary scheduling policy primitive. |
| `cloudprovider.InstanceType` | `vendor/sigs.k8s.io/karpenter/pkg/cloudprovider/types.go` | Represents virtual capacity shape (includes `Overhead`). |
| `cloudprovider.Offering` | `vendor/sigs.k8s.io/karpenter/pkg/cloudprovider/types.go` | 1:1 mapping to CA NodeGroups; holds `Price` and `Requirements`. |
| `scheduling.Results` | `vendor/sigs.k8s.io/karpenter/pkg/controllers/provisioning/scheduling/scheduler.go` | Output of the `Solve()` loop. |

### 2. Primary Interfaces (Cluster Autoscaler)
| Type | File Path | Usage |
| :--- | :--- | :--- |
| `cloudprovider.NodeGroup` | `cloudprovider/cloud_provider.go` | Source of truth for physical limits and templates. |
| `clustersnapshot.ClusterSnapshot` | `simulator/clustersnapshot/clustersnapshot.go` | High-fidelity in-memory cluster state. |
| `expander.Strategy` | `expander/expander.go` | Tie-breaking logic for multi-group resolver. |

### 3. Solver Method Signatures
- **`NewScheduler`** (`vendor/sigs.k8s.io/.../scheduler.go`):
  `func NewScheduler(ctx context.Context, kubeClient client.Client, nodePools []*v1.NodePool, cluster *state.Cluster, stateNodes []*state.StateNode, topology *Topology, instanceTypes map[string][]*cloudprovider.InstanceType, daemonSetPods []*corev1.Pod, recorder events.Recorder, clock clock.Clock, volumeReqsByPod map[types.UID]scheduling.Requirements, opts ...Options) *Scheduler`
- **`Solve`** (`vendor/sigs.k8s.io/.../scheduler.go`):
  `func (s *Scheduler) Solve(ctx context.Context, pods []*corev1.Pod) (Results, error)`

### 4. Integration Entry Points
- **Feature Flag:** `config/autoscaling_options.go` (`AutoscalingOptions.KarpenterSimulatorEnabled`)
- **Scale-Up Orchestrator:** `core/scaleup/orchestrator/orchestrator.go`
- **Scale-Down Simulator:** `simulator/cluster.go` (`RemovalSimulator`)
- **Fidelity Filter:** `core/podlistprocessor/filter_out_schedulable.go`
- **Simulator Core:** `simulator/karpenter/karpenter_simulator.go` (Target location for `runKarpenterSimulation`)


