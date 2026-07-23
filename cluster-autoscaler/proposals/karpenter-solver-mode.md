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
2. **Single-Engine Simulation:** The simulator evaluates the sliced batch of pending pods. It routes compatible pods to the Karpenter solver (using a zero-overhead `DirectClient` over the `ClusterSnapshot`). Any pods requesting features currently unsupported by Karpenter (such as Dynamic Resource Allocation (DRA)) are skipped during this cycle and remain pending.
3. **Resolving:** Karpenter's virtual `NodeClaim` decisions are mapped back to Equivalence Sets of real, physical CA `NodeGroup`s.
4. **Execution:** The generated scale-up options are returned as a list of decisions (`[][]expander.Option`) to the orchestrator. Each option includes a validation callback (`IsSimilarValid`) for dynamic similarity filtering. The orchestrator selects the best options using the expander, resolves similarity, balances them using the `expectedTargetSizes` map to prevent stale size race conditions, and triggers the physical resizes.

---

## Core Concept: Underspecified NodeClaims and Equivalence Sets

To understand how Karpenter integrates with Cluster Autoscaler's expander and balancing mechanisms, it is essential to understand the concept of **Underspecified NodeClaims** and **Equivalence Sets**.

1. **Underspecified NodeClaims:** Karpenter's solver operates on virtual requirements and outputs `NodeClaim`s. These claims are often *underspecified*—they do not reference a specific physical node group, but rather describe the requirements (e.g., "CPU: 4, Mem: 16GB, Arch: amd64, Zone: us-central1-a, CapacityType: Spot").
2. **Equivalence Sets:** The simulator takes this virtual `NodeClaim` and resolves it against CA's physical `NodeGroup`s. It finds **all** physical node groups that satisfy the claim's requirements. This collection of matching physical groups forms an **Equivalence Set** for that claim.
3. **Expander Selection:** The simulator wraps each matching group in the Equivalence Set into an `expander.Option`, yielding a list of options (`[]expander.Option`) representing the alternatives. Because every option in this list strictly satisfies the `NodeClaim`'s scheduling requirements, they are all **equally valid** from a constraint perspective. CA's native `ExpanderStrategy` (e.g., `price`, `random`) is then used to select the "best" option from this set (e.g., the cheapest one) to reduce the underspecified claim into a concrete group to resize. The expander does *not* override Karpenter's scheduling decisions; it merely chooses the best physical realization of it.
4. **Zonal Balancing:** Similarly, CA's balancing logic only distributes scale-ups between node groups that are **similar and equally valid** for the constraints. The `IsSimilarValid` callback validates that candidate similar groups belong to the same Equivalence Set for the claim. This ensures that balancing (e.g., across zones) only happens between groups that Karpenter's solver would have also accepted as valid for those pods, preventing balancing from violating scheduling constraints.

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
- **Decision:** The native CA scale-down `RemovalSimulator` and the `filterOutSchedulable` processor use the Karpenter solver via the **`PodSchedulingSimulator`** interface.
- **Implementation:** The **`KarpenterReschedulingSimulator`** is injected into these components.
- **Existing-Only Mode:** For these checks, the simulator invokes the Karpenter solver with **zero `NodePools`**.
- **Predicate-Aware Placement (Reconciliation):** To ensure Karpenter respects CA's removal logic (specifically preventing pods from scheduling onto the node being removed):
    - **Topology Visibility:** Karpenter's **Topology Engine** (`state.Cluster`) is initialized with the **full set of nodes** from the `ClusterSnapshot`. This ensures all pod anti-affinity and topology spread constraints are mathematically accurate across the entire cluster.
    - **Filtered Placement Targets:** Karpenter's **Placement Engine** (`Scheduler.stateNodes`) is initialized using a **filtered slice of nodes** that pass the `IsNodeAcceptable` predicate provided in `SchedulingOptions`. This restricts the solver to only consider nodes designated as valid rescheduling targets, preventing it from incorrectly placing pods on nodes marked for removal.
- **Zero-Dependency Simplicity:** Unlike the scale-up simulator, the rescheduling simulator does **not** require the `KarpenterConverter` or `ExpanderStrategy`.
- **Lightweight Performance:** This mode skips heavy binpacking and cost optimization, ensuring that CA's performance during heavy scale-down evaluations remains comparable to native predicate checks.

---

## Primitives Conversion (`KarpenterConverter`)

To optimize solver performance and cost-fidelity, physical NodeGroups are collapsed into virtual Karpenter primitives without losing scheduling constraints. The `KarpenterConverter` interface exposes a single, stateless transformation method `Convert(...) (*ConversionResult, error)` that returns the converted Karpenter objects alongside pre-indexed access and lookup methods, eliminating temporal coupling and mutable converter state.

### 1. NodePool Construction
- Physical `NodeGroup`s are grouped by **strictly taint compatibility**. All node groups sharing identical taints are mapped under a single virtual `NodePool` (named `pool-<taints-hash>`). Taint serialization and grouping heuristics are encapsulated as private implementation details of the converter.
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
- **Spot Discount Calculation:** For offerings with the `spot` capacity type, the calculated price must be multiplied by a **Spot Discount coefficient of 0.3** (e.g., `Price = Price * 0.3`), representing a 70% discount from the On-Demand price. This coefficient must be mathematically locked to ensure consistent cost-optimization across different instance families. **Note:** This discount is only applied to the weighted fallback cost function when the native pricing model is unavailable, and not on top of the native pricing model's output.

#### B. Requirement Matching Matrix
Each `Offering` inherits **all labels** from its representing physical `NodeGroup`:
- **Well-Known Labels:** Declared with actual combined values at the NodePool level, and specific values at the Offering level.
- **Custom / Ignored Labels:** 
  - **NodePool Level:** Add a permissive `NotIn [ca-ignore-reserved-value]` requirement for every unique custom label key. Using a syntactically valid but reserved dummy value ensures compatibility with all internal Kubernetes label parsers.
  - **Offering Level:** For NodeGroups possessing the label, use `In [value]`. For those lacking it, use an explicit `DoesNotExist` operator. 

This strategy ensures Karpenter correctly selects compatible offerings for positive, negative, and neutral pod selectors.

---

## In-Memory Simulation State (`runKarpenterSimulation`)

To guarantee isolation and prevent modifying the actual cluster state during simulation, the entire `runKarpenterSimulation` function runs inside a **forked `ClusterSnapshot`** (using `snapshot.Fork()` at the beginning and `snapshot.Revert()` via `defer` or on exit).

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

### 4. Unified Topology State & Unmanaged Snapshot Nodes
- The live cluster topology (including physical nodes, pending unready nodes, and all pods already bound to them) is retrieved via `autoscalingCtx.ClusterSnapshot.ListNodeInfos()`.
- Upcoming nodes currently scaling up are treated identically to real nodes.
- Pending/unschedulable pods passed to the function are treated strictly as the input (`remainingPods`) to Karpenter's scheduler solver.
- **Natural Unmanaged Node Semantics:** Snapshot nodes from CA do **not** need to be decorated with synthetic Karpenter labels (`karpenter.sh/nodepool` or `karpenter.sh/initialized: "true"`). Standard Karpenter natively ingests nodes lacking `karpenter.sh/nodepool` as unmanaged nodes:
    1.  **State Ingestion:** In `state.Cluster.UpdateNode()`, unmanaged nodes without an explicit `Spec.ProviderID` automatically default `Spec.ProviderID = node.Name`.
    2.  **Scheduling Readiness:** In `StateNode`, unmanaged nodes are automatically treated as registered and initialized (`Registered() == true`, `Initialized() == true`).
    3.  **Capacity Accounting:** Karpenter's scheduler ingests all unmanaged snapshot nodes into `ExistingNode` structures and prioritizes placing pending pods on existing available capacity before evaluating new scale-up claims.
    4.  **Disruption Isolation:** Unmanaged nodes are never subject to Karpenter disruption or consolidation.

### 5. Single-Run Batching & Salvo Coupling
Pod Batching is explicitly coupled with the Salvo Scale-Up Loop feature (`SalvoScaleUpEnabled`):
- **When Salvo is Enabled (`SalvoScaleUpEnabled == true`):**
  - **Batch Slicing:** Before invoking the Karpenter solver, the simulator slices the incoming `unschedulablePods` list to a maximum batch size of `MaxBatchSize = 1000` pods: `slicedPods := unschedulablePods[0:min(MaxBatchSize, len(unschedulablePods))]`.
  - **Outer Iteration (Salvo):** The simulator executes the Karpenter solver (`Solve()`) once for the `slicedPods` in the current invocation. The outer Salvo loop (`runScaleUpSalvo`) executes the chosen scale-up options, updates the `ClusterSnapshot` with upcoming nodes and virtually scheduled pods, and invokes the orchestrator again in the next salvo iteration with the remaining pods.
  - **State Continuity:** Subsequent salvo iterations evaluate remaining pods against a snapshot that already contains the upcoming nodes from previous iterations, preserving topology spread constraints and pod anti-affinities.
- **When Salvo is Disabled (`SalvoScaleUpEnabled == false`):**
  - **Batching Disabled:** Pod batching is automatically **disabled** (all unschedulable pods are evaluated in a single solver pass without slicing to produce a single decision, preventing unactuated batch loss).
  - **Diagnostic Warning:** The autoscaler logs a warning at initialization: `"Karpenter solver mode is enabled without Salvo scale-up loop. Pod batching is disabled; Karpenter solver will evaluate pods in a single pass."`
- **Pre-Simulation Pruning (Performance Fast-Path):** Before running the solver, the converter automatically disables virtual `Offering`s whose corresponding physical `NodeGroup`s have already hit their `MaxSize` limits in the cloud provider.
- **Benchmarking Strategy:** Benchmark test suites test both configurations (`KarpenterSimulatorEnabled` + `SalvoScaleUpEnabled` for batched multi-pass simulation, and `KarpenterSimulatorEnabled` alone for single-pass unbatched simulation).

### 6. Resource Clamping & Pod Pruning (HydrateClusterState)
To prevent performance degradation in large clusters (where evaluating topology spreads and affinities over thousands of irrelevant pods is slow), the simulator implements a pruning and clamping phase during cluster state hydration, borrowing the relevance filter pattern from upstream:
- **Relevance Filter (Pod Pruning):** The hydration phase (`HydrateClusterState`) filters the pods on existing nodes. A pod on an existing node is deemed **relevant** (and thus kept) if and only if it matches one of the following criteria:
    1.  **Pending Selector Match:** The running pod's labels match any of the scheduling selectors (Affinity, Anti-Affinity, or Topology Spread Constraints) defined by the pending pods in the current batch.
    2.  **Running Selector Match:** The running pod has Anti-Affinity constraints whose selector matches the label set of *any* pending pod in the batch. (This is evaluated using the real `labels.Selector` matching logic to ensure negative operators like `NotIn` or `DoesNotExist` are correctly handled when labels are absent from pending pods).
    3.  **TSC Domain Representation:** The node containing the pod is required to represent a unique Topology Spread Constraint domain (e.g. to ensure zone-spread calculations are correct, we preserve at least one representative node per topology domain signature).
- **Resource Clamping:** For each running pod that is pruned (hidden), **all** of its resource requests (including CPU, Memory, GPUs, and custom resources) are summed and subtracted directly from the representing node's `Status.Allocatable` capacity in the snapshot. This ensures the simulator accurately tracks remaining node capacity without needing to load and evaluate the pruned pods.
- **Dynamic $O(N)$ Node Pruning Threshold & Bounded Pareto Minimal Pod Shapes:** A node with no relevant pods is retained in the simulation snapshot if and only if its remaining allocatable capacity ($\text{clampedAllocatable}$) can fit **at least one** candidate pending pod shape: $\exists p \in S_{\text{min}}, \text{remCap}(node) \ge \text{req}(p)$.
  - **Candidate Pod Shapes:** A pod shape is defined as a distinct resource request vector $V = (v_{\text{cpu}}, v_{\text{mem}}, v_{\text{gpu}}, \dots)$ present among the pending pods in the batch.
  - **Pareto Dominance Reduction:** A pod shape $V_A$ is dominated by shape $V_B$ if $V_B \le V_A$ across all resource dimensions ($\forall R, V_B[R] \le V_A[R]$). Dominated (larger) shapes are redundant for capacity checks because any node fitting $V_A$ also fits $V_B$. The minimal set $S_{\text{min}}$ retains only non-dominated shapes.
  - **Incomparable Shapes (Pareto Frontier):** Request shapes with trade-offs like $(1\text{ CPU}, 3\text{Gi Mem})$, $(2\text{ CPU}, 2\text{Gi Mem})$, and $(3\text{ CPU}, 1\text{Gi Mem})$ are mutually incomparable under Pareto ordering ($\nexists V_B \le V_A$). All such non-dominated shapes are retained in $S_{\text{min}}$, guaranteeing that nodes with intermediate free capacities (e.g. $2\text{ CPU}, 2\text{Gi Mem}$) are correctly preserved.
  - **Bounded Pareto Fallback ($\text{MaxShapes} = 10$):** To prevent $O(N \cdot P)$ worst-case complexity under pathological workloads where $M = |S_{\text{min}}|$ is very large, $S_{\text{min}}$ is capped at $\text{MaxShapes} = 10$. If $|S_{\text{min}}| > 10$, $S_{\text{min}}$ collapses to the single global minimum vector ($\text{GlobalMin}[R] = \min_{p} \text{req}(p, R)$). This guarantees a strict $O(10 \cdot N) = O(N)$ time complexity under all workloads.
- **Exemptions (Never Pruned):** Regardless of the relevance filter, pods that use **HostPorts** or **CSI Volumes** are **never pruned**, as their physical presence is required to simulate port conflicts and CSI volume attachment limits.
- **Disabling InterPodAffinity & PodTopologySpread Framework Plugins:** In Karpenter simulator mode, `InterPodAffinity` and `PodTopologySpread` plugins are automatically disabled in CA's framework handle (`NewKarpenterDisabledPluginsSchedulerConfig`).
  - **Rationale:** Karpenter's solver natively evaluates and validates all inter-pod affinities and topology spread constraints during its simulation pass (`ks.Solve()`). Disabling these redundant inter-pod tracking plugins in CA's predicate framework avoids superfluous CPU overhead during `snapshot.SchedulePod` and prevents false-positive predicate rejections due to partial snapshot state.

---

## NodeClaim -> NodeGroup Resolving (`mapResultsToOptions`)

Resolving is the process of mapping Karpenter's virtual scheduling decisions (`NodeClaim`s) back to real-world CA physical primitives. In Option B (Single Salvo), this resolution happens once at the end of the single solver run.

The resolution flow is as follows:

### 1. Grouping by Constraints
The solver outputs a list of virtual `NodeClaims` for the sliced batch. We group these claims by their exact scheduling `Requirements` into slices of identical claims: `[]*scheduling.NodeClaim`. This preserves the individual claim identity and its associated pods.

### 2. Converting to Expansion Options (Alternatives Generation)
For each group of claims (slice) of size `count = len(slice)`:
- We identify the set of physical `NodeGroup`s that match the claim's requirements. This set is pre-filtered using Karpenter's strict requirement matching, ensuring any group violating the claim's constraints (e.g. invalid zone) is excluded.
- For **each** physical `NodeGroup` in this matching set, we generate a distinct `expander.Option`:
    - The `NodeGroup` is set to the matching physical group.
    - The `NodeCount` is set to `count`.
    - The `Pods` are the consolidated list of all pods assigned to the claims in the slice.
    - We define a **validation callback (`IsSimilarValid`)** for this option. This callback captures the `NodeClaim` requirements and this specific node group's hardware template. When called with a candidate node group, it returns true if the candidate is hardware-similar to this option's group AND satisfies the `NodeClaim` requirements.
        *   **Zonal Balancing Limitation:** Because Karpenter's scheduler assigns pods to concrete zones during simulation to evaluate topology constraints, the output `NodeClaim` requirements are locked to the chosen zone. Consequently, the `IsSimilarValid` callback (which validates against these requirements) will only accept similar groups in the same zone. Zonal balancing across different zones is therefore disabled for Karpenter-simulated scale-ups, unless Karpenter's solver itself generates claims in different zones (e.g., to satisfy a `TopologySpreadConstraint` across multiple zones).
    - The `SimilarNodeGroups` field is left empty (or initialized to an empty slice), as similarity filtering will be performed dynamically by the orchestrator using the callback.
- This yields a list of alternative options for this claim: `[]expander.Option`.

### 3. Selecting the Best Option (Orchestrator Selection)
The simulator returns the full list of alternatives for each decision to the orchestrator. The orchestrator (in `prepareScaleUp`) loops over these lists and runs the configured `ExpanderStrategy` to select the final option to execute. This keeps the final choice and limit enforcement in the orchestrator, ensuring consistency with native CA.

### 4. Consolidation
All physical scale-up decisions generated for the batch are returned as a list of lists of options (`[][]expander.Option`) to the scale-up orchestrator. Each decision (representing a single or multi-group scale-up) is wrapped in a single-option list.

### 5. Metadata Population (`populateSchedulablePodGroups`)
Before returning, the simulator populates the `schedulablePodGroups` map. Crucially, it populates entries for **both** the chosen group and all candidate groups that *could* have matched the claim (by running the `IsSimilarValid` callback on all existing groups). This ensures that if the orchestrator recomputes similarity later (e.g., after on-demand creation), CA's native logic will find the matching metadata and include the valid groups while excluding the invalid ones.


---

## Scale-Up Orchestrator Execution Loop (`prepareScaleUp`)

To support multi-decision scale-ups (`[][]expander.Option`), the orchestrator's `prepareScaleUp` method processes and executes multiple independent scale-up decisions in a single cycle.

### 1. Collapsing Options to Decisions
The orchestrator loops over the outer slice of `[][]expander.Option` returned by the `ScaleUpSimulator`. For each inner slice (which represents a single decision point with one or more alternative options), it invokes the configured `ExpanderStrategy` to select the best option.
This yields a list of chosen options: `bestOptions []expander.Option`.

### 2. Multi-Group Actuation Flow
To prevent race conditions and latency issues during sequential planning (both within a single orchestrator run and across consecutive Salvo loop iterations before cloud provider cache updates reflect), the Salvo loop maintains an **`expectedTargetSizes` map** (`map[string]int` keyed by `NodeGroup.Id()`) and passes it to the orchestrator's `ScaleUp` call via the **Functional Options Pattern** (`scaleup.WithExpectedTargetSizes(expectedTargetSizes)`).

For each option in `bestOptions`:
1.  **On-Demand Node Group Creation:** If the option's `NodeGroup` does not exist (`!ng.Exist()`), the orchestrator triggers its creation (`CreateNodeGroup`).
    *   **Post-Creation Infrastructure State:** `CreateNodeGroup` provisions the main group and any extra node groups in the cloud provider, registering them in `CloudProvider.NodeGroups()` with `group.Exist() == true`.
    *   **Schedulable Pod Groups Update:** `processCreateNodeGroupResult` updates `schedulablePodGroups` for the main created group and any extra created groups.
2.  **Unified Similarity Resolution:**
    *   Similarity resolution is performed **post-creation** using a single, unified callback interface on `expander.Option`: `IsSimilarValid func(group cloudprovider.NodeGroup, nodeInfo *framework.NodeInfo) bool`.
    *   **Karpenter Mode:** `IsSimilarValid` checks if a candidate group satisfies the Karpenter solver's `NodeClaimTemplate` requirements (arch, os, instance type, taints, labels).
    *   **Native CA Mode:** `IsSimilarValid` checks if candidate groups can schedule the pending pods (`len(schedulablePodGroups[group.Id()]) > 0`).
    *   **Filtering:** The orchestrator queries `ComputeSimilarNodeGroups` to get candidate groups, and filters them by ensuring both `group.Exist() == true` (filtering out uncreated virtual groups) and `opt.IsSimilarValid(group, nodeInfo) == true`.
3.  **Zonal Balancing & Capping:** We call `balanceScaleUps` passing the resolved `similarGroups` (which contains only existing physical groups) and the capped node count for this option.
    *   **Expected Size Overrides:** `balanceScaleUps` (and the orchestrator's `filterValidScaleUpNodeGroups` validation check) is updated to query `expectedTargetSizes` instead of `ng.TargetSize()` directly. If a group has an expected size recorded in the map, it uses that value as the baseline for capacity and limit evaluations.
    *   `balanceScaleUps` is updated: if `similarNodeGroups` is passed as a non-nil slice (even if empty `[]`), it bypasses recomputation and uses it directly. If `nil`, it falls back to native recomputation.
    *   This generates a `ScaleUpInfo` slice for this option.
4.  **Tracking Expected Sizes:** For each group in the generated `ScaleUpInfo` slice, the orchestrator updates the `expectedTargetSizes` map with the new expected target size. This map is returned to the Salvo loop for the next iteration.
5.  **Aggregation:** We aggregate all `ScaleUpInfo`s across all options into a final scale-up plan. Duplicate node groups (which can occur if different decisions map to the same physical group) must be **merged into a single `ScaleUpInfo` entry** by summing their requested node deltas and adjusting the `NewSize` accordingly. This ensures the plan conforms to CA's parallel executor uniqueness requirements.

### 3. Status Reporting & Struct Updates
To ensure correct status reporting in multi-decision mode:
*   **`scaleUpPlan` Structure:** The internal `scaleUpPlan` struct holds `bestOptions []expander.Option` (instead of a single `bestOption` in upstream) to support multi-group scale-ups.
*   **Status Aggregation:** The orchestrator's status reporting code (populating `status.ScaleUpStatus`) aggregates `PodsTriggeredScaleUp` from all executed `bestOptions`.
*   **Evaluation Tracking:** Status helper methods like `GetPodsAwaitingEvaluation` check for scheduling errors across all chosen node groups in `bestOptions` rather than assuming a single target group.

This sequential execution ensures that each decision is created, filtered, and balanced correctly using its specific validation rules, while still executing the entire multi-group scale-up in a single autoscaling cycle.

---

## Processor & Pipeline Disablement

To guarantee optimal performance and prevent scheduling loops:

### 1. Karpenter-Aware `filterOutSchedulable` Phase & Salvo Candidate Filtering
- **Decision:** The standard CA `filterOutSchedulable` phase is **retained and made Karpenter-aware** using `autoscalingCtx.PodSchedulingSimulator` (`KarpenterReschedulingSimulator` without NodePools).
- **Rationale:** Making `filterOutSchedulable` Karpenter-aware guarantees that downstream consumers of CA code (such as metrics, debugging snapshotters, and custom `PodListProcessor`s) continue to observe standard filtering behavior. The phase tests pending pods against existing cluster snapshot nodes without provisioning new NodePools, filtering out any pods that can already schedule on existing capacity.
- **Scale-Up Candidate Optimization & Node Classification:** Because `filterOutSchedulable` guarantees that remaining pending pods cannot be placed on existing non-Salvo nodes, `ScaleUp` (`KarpenterSimulator.Simulate`) skips non-Salvo nodes as candidate placement nodes (`stateNodes` passed to `NewScheduler`), passing only Salvo nodes created in earlier Salvo iterations. Salvo nodes are robustly identified via the explicit annotation `cluster-autoscaler.k8s.io/salvo-node: "true"` (`annotations.NodeSalvoAnnotation`) attached when upcoming nodes are injected into `ClusterSnapshot`. Non-Salvo nodes remain in `state.Cluster` and `Topology` (via `HydrateClusterState`) to maintain full cluster topology context for pod affinity, anti-affinity, and topology spread constraints.

### 2. Disablement of ProvisioningRequest Handling
- **Decision:** ProvisioningRequest handling is disabled to prevent conflicts with Karpenter-driven scheduling.
- **Enforcement:** The `AutoscalerBuilder` fails to initialize and returns a configuration error if both `KarpenterSimulatorEnabled` and `ProvisioningRequestEnabled` are enabled.

### 3. Simulation Preemption Model
- **Decision:** The integrated solver performs **Binpacking Only**. It does not simulate preemption of lower-priority pods to make room for higher-priority pods.
- **Rationale:** This is consistent with CA's native scale-up simulation. Preemption victims and nominated node assignments are handled by CA's native pre-processing logic and are reflected in the `ClusterSnapshot` before the solver is invoked.

### 4. Preserved Processors
- All other active CA pod list processors—including **proactive scale-up** and **capacity buffers support**—must remain fully enabled.

---

## Validation Rules & Testing Lifecycle

To ensure the technical integrity and performance of the Karpenter integration, the implementation adheres to the following three-phase validation lifecycle:

1.  **Component-Level Unit Testing:** Every new component, interface implementation, or logic refinement (e.g., `DirectClient` facade, `KarpenterConverter`, or `OfferingPruningTracker`) is verified with exhaustive unit tests.
2.  **Aggregate Integration Testing:** The integration is tested in aggregate to confirm that the wiring between the Scale-Up Orchestrator, Scale-Down Planner, and the Karpenter solver works seamlessly without regressions in standard CA behavior.
3.  **Performance Benchmarking:** The implementation is validated using the project's benchmark suite to demonstrate that the Karpenter-integrated mode provides superior binpacking efficiency and/or computational performance compared to the standard predicate-based Cluster Autoscaler.

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
- **Expected Outcome:** The orchestrator correctly invokes `balanceScaleUps` with the pre-filtered similar groups (obtained via the callback). Verifies that the final decisions contain a distribution of scale-ups across the valid similar node groups (e.g. Zone A and B) and excludes the invalid one (Zone C). This confirms the integration preserves CA's zonal reliability guarantees.

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

This section documents the technical constraints and concrete implementation patterns used to ensure a successful, crash-free delivery.

### 1. Physical Implementation Signatures

#### A. PodSchedulingSimulator Interface
Used for existing-node checks (Scale-Down, Pre-checks). Defined in `simulator/clustersnapshot`:
```go
type PodSchedulingSimulator interface {
	TrySchedulePods(snapshot ClusterSnapshot, pods []*apiv1.Pod, breakOnFailure bool, opts SchedulingOptions) ([]Status, int, error)
}
```
**Circular Dependency Resolution:** The **`Status` struct** (containing `Pod *apiv1.Pod` and `NodeName string`) is located in `simulator/clustersnapshot` (instead of `simulator/scheduling` in upstream). This ensures that the interface can be defined in the neutral `clustersnapshot` package without requiring an import of the `scheduling` package, which depends on `clustersnapshot`.

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
	) ([][]expander.Option, map[string]status.Reasons, map[string][]estimator.PodEquivalenceGroup, error)
}
```

#### B2. Option Struct Extension
The `Option` struct in `expander/expander.go` is extended with the `IsSimilarValid` callback field:
```go
type Option struct {
	NodeGroup         cloudprovider.NodeGroup
	SimilarNodeGroups []cloudprovider.NodeGroup
	NodeCount         int
	Debug             string
	Pods              []*apiv1.Pod

	// IsSimilarValid is an optional callback to validate if a candidate node group
	// is similar and valid for this option. Used by Karpenter to avoid zonal matching bugs.
	IsSimilarValid func(group cloudprovider.NodeGroup, nodeInfo *framework.NodeInfo) bool
}
```

#### B3. KarpenterSimulator Constructor
The `KarpenterSimulator` constructor in `core/scaleup/orchestrator/karpenter_simulator.go` is defined to accept the configured estimation thresholds:
```go
func NewKarpenterSimulator(
	defaultSimulator ScaleUpSimulator,
	converter karpenter.KarpenterConverter,
	processor nodegroupset.NodeGroupSetProcessor,
	thresholds []estimator.Threshold, // Enforces capacity limits during simulation
) *KarpenterSimulator
```

#### C. RemovalSimulator Constructor
Accepts the `PodSchedulingSimulator` interface. Defined in `simulator/cluster.go`:
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

#### D. KarpenterConverter Interface & ConversionResult
Only required for the `ScaleUpSimulator`. Defined in `simulator/karpenter`:
```go
type KarpenterConverter interface {
	Convert(nodeGroups []cloudprovider.NodeGroup, nodeInfos map[string]*framework.NodeInfo) (*ConversionResult, error)
}

type ConversionResult struct {
	NodePools          []*karpenterv1.NodePool
	InstanceTypes      map[string][]*karpentercloudprovider.InstanceType
	OfferingMap        map[string]*karpentercloudprovider.Offering
	PoolITToNodeGroups map[string]map[string][]cloudprovider.NodeGroup
	NodeGroupToPool    map[string]string
}

func (res *ConversionResult) NodeGroupsFor(nodePoolName, itName string) []cloudprovider.NodeGroup
func (res *ConversionResult) PoolForNodeGroup(ngId string) string

// Exported Package Utilities for Custom Converters:
func SerializeTaints(taints []apiv1.Taint) string
func InstanceTypeNameFromLabels(labels map[string]string, defaultName string) string
```

#### E. Native CA Balancing Call
The balancer is invoked via the `BalancingNodeGroupSetProcessor` processor:
```go
func (b *BalancingNodeGroupSetProcessor) BalanceScaleUpBetweenGroups(autoscalingCtx *ca_context.AutoscalingContext, groups []cloudprovider.NodeGroup, newNodes int) ([]ScaleUpInfo, errors.AutoscalerError)
```

#### F. balanceScaleUps Signature
The orchestrator's internal balancing method `balanceScaleUps` accepts pre-computed similar groups:
```go
func (o *ScaleUpOrchestrator) balanceScaleUps(
	now time.Time,
	nodeGroup cloudprovider.NodeGroup,
	similarNodeGroups []cloudprovider.NodeGroup,
	newNodes int,
	nodeInfos map[string]*framework.NodeInfo,
	schedulablePodGroups map[string][]estimator.PodEquivalenceGroup,
	tracker *resourcequotas.Tracker,
) ([]nodegroupset.ScaleUpInfo, errors.AutoscalerError)
```
If `similarNodeGroups` is not `nil` (it can be an empty slice `[]`), it uses it directly (concatenated with the main group) instead of calling `ComputeSimilarNodeGroups`. An empty slice explicitly disables balancing, whereas a `nil` value acts as a fallback to trigger native recomputation.

### 2. Package Organization & Decoupling
To prevent cyclic import dependencies between `core/scaleup/orchestrator` and `simulator/scheduling`, the architectural core is decoupled via the `PodSchedulingSimulator` interface:
- `simulator/clustersnapshot`: Holds the interface and shared options.
- `simulator/scheduling`: Implements `NativeHintingSimulator`.
- `simulator/karpenter`: Implements `KarpenterReschedulingSimulator` and `KarpenterScaleUpSimulator`.
- **Decoupled Dependency Injection:** The **`AutoscalingContext`** (in `context/autoscaling_context.go`) includes the **`PodSchedulingSimulator`** interface. 
- **Wiring Layer:** During the initialization of the Cluster Autoscaler (in `core/static_autoscaler.go`), the appropriate simulator implementation is instantiated based on the `KarpenterSimulatorEnabled` flag and assigned to the context.
- **Scale-Down Integration:** The `ScaleDownPlanner` (in `core/scaledown/planner/planner.go`) retrieves the simulator from the `AutoscalingContext` and passes it to the `RemovalSimulator` constructor, ensuring zero direct dependency on specific simulator implementations within the scale-down logic.
- **Scale-Up Integration:** The Scale-Up Orchestrator receives both the `PodSchedulingSimulator` (for pre-checks) and the `ScaleUpSimulator` (for binpacking) via its constructor options.

### 3. ClusterSnapshot Extensions & Metadata Sourcing
To support storage-aware simulation, the `ClusterSnapshot` is extended to track PV, PVC, StorageClass, and CSINode objects using the `common.PatchSet` pattern:
- **Internal Maps:**
  - `pvs *common.PatchSet[string, *apiv1.PersistentVolume]`
  - `pvcs *common.PatchSet[string, *apiv1.PersistentVolumeClaim]`
  - `storageClasses *common.PatchSet[string, *storagev1.StorageClass]`
  - `csiNodes *common.PatchSet[string, *storagev1.CSINode]`
- **Interface Methods:** The `ClusterSnapshot` interface includes the following methods to enable `DirectClient` facade lookups:
  - `GetPV(name string) (*apiv1.PersistentVolume, error)`
  - `ListPVs() ([]*apiv1.PersistentVolume, error)`
  - `GetPVC(namespace, name string) (*apiv1.PersistentVolumeClaim, error)`
  - `ListPVCs() ([]*apiv1.PersistentVolumeClaim, error)`
  - `GetStorageClass(name string) (*storagev1.StorageClass, error)`
  - `ListStorageClasses() ([]*storagev1.StorageClass, error)`
- **Sourcing:** These objects are encapsulated within the `ClusterSnapshot` during `SetClusterState`.

### 4. Concrete Field Traversals & Fallbacks

#### A. Extracting CSI Volume Limits
To populate `InstanceType.VolumeLimits`, the converter traverses the `CSINode` metadata in the physical `NodeInfo`:
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
If node templates lack standard labels, the following default mapping is applied:
- `kubernetes.io/arch`: Default to `amd64`.
- `kubernetes.io/os`: Default to `linux`.
- `karpenter.sh/capacity-type`: Default to `on-demand`.

### 5. Karpenter State Instantiation & Filtering
Karpenter's solver requires high-level state objects. These are instantiated using the `DirectClient` facade:
- **state.Cluster:** Initialized via `state.NewCluster(clk, directClient, nil)`. The `CloudProvider` interface is `nil` as state is populated manually.
- **Manual State Population:** The simulator manually populates the `state.Cluster` by iterating over the `ClusterSnapshot` nodes and calling `cluster.UpdateNode(ctx, node)`. This triggers Karpenter's internal logic to fetch pods and CSI metadata via the `DirectClient` facade.
- **StateNode Collection:** After hydration, the `*state.StateNode` objects are collected from the `cluster.Nodes()` iterator to form the `stateNodes` slice required by the `NewScheduler` constructor.
- **Predicate-Aware Filtering (Scale-Down):** In the rescheduling simulator, the `stateNodes` slice passed to `NewScheduler` is filtered by the `IsNodeAcceptable` predicate from `SchedulingOptions`. This ensures that pods being evicted are never "scheduled" back onto nodes marked for removal (or other excluded nodes), while still allowing the `Topology` engine (using the full `cluster`) to account for anti-affinity against all pods.
- **Topology Engine:** Initialized via `scheduling.NewTopology(ctx, directClient, cluster, stateNodes, ...)`.

### 6. DirectClient (client.Client) Safety
The `DirectClient` acts as a `read-only facade`. To ensure the Karpenter solver does not attempt to mutate cluster state:
- **Read Methods (Get, List):** Dynamically query the `ClusterSnapshot`.
- **Write Methods (Create, Update, Delete, Patch):** Return **`apierrors.NewMethodNotSupported`**.

### 7. Pricing Fallback Weights (Coefficients)
When cloud provider pricing is unavailable, the weighted cost function uses these standard coefficients:
- **CPU_Weight:** `1.0` (Base cost per 1 vCPU)
- **Mem_Weight:** `0.125` (Cost per 1 GiB; standard 1:8 vCPU-to-GiB ratio)
- **GPU_Weight:** `25.0` (Cost per 1 GPU)
- **Spot_Multiplier:** `0.3` (70% discount)

### 8. Hashing & Naming Stability
To generate stable, unique identifiers for virtual primitives while respecting Kubernetes name length limits (63 chars):
- **Algorithm:** Uses **`FNV-1a` (64-bit)** hashing on the serialized grouping keys.
- **Formatting:** Hashes are represented as lowercase hexadecimal strings.
- **Stability:** The grouping keys (Taints, HostPorts, Labels) are **lexicographically sorted** before hashing.

### 9. Offering Pruning & Identity
- **Pruning Tracker:** A `map[string]*cloudprovider.Offering` (keyed by physical `NodeGroup.Id()`) lives inside the `runKarpenterSimulation` function's scope.
- **Limit Enforcement:** When the resolver maps a virtual claim back to a physical group and detects it has hit `MaxSize`, it marks `Available = false` in this map.

### 10. Simulated Node Identity
- **Naming Convention:** `simulated-node-<nodegroup-id>-<uuid>`.
- **Hostname Alignment:** The simulated node's `kubernetes.io/hostname` label is explicitly set to match its generated name.

### 11. Error Translation & Partial Success
To integrate Karpenter's results back into CA's status tracking:
- **Partial Success:** Pods that failed to schedule (present in `Results.PodErrors`) are reported using CA's `status.Reasons`.
- **Mapping:** 
  - If a pod failure is due to a constraint, it is mapped to `status.BackoffReason`.
  - If the solver returns a terminal `error`, it is wrapped as a `status.ScaleUpError` with `status.InternalError` type.

---

## Known Limitations

- **Unsupported Karpenter Features (Fallback Pods):** Pods requesting features not yet supported by the integrated Karpenter scheduler version (such as Dynamic Resource Allocation (DRA)) will be skipped by the simulator and remain pending. They will not trigger scale-ups nor will they be scheduled on simulated nodes until Karpenter natively supports these features.

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


