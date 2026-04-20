# Fast Incremental Slice-Based Affinity

## Overview
This document describes a high-performance architecture for Pod Affinity and Anti-Affinity predicates, optimized for the scale and volatility of the Cluster Autoscaler (CA). This design replaces traditional $O(\text{PodsOnNode})$ selector matching with incremental indexing, a **Two-Way Label Index**, and extremely fast $O(1)$ slice lookups.

The system is built on three architectural pillars:
1.  **Incremental Indexing (Streaming Snapshot Store):** Internal data structures are updated reactively as pods and nodes change using the **Fort** pipeline library, rather than being re-calculated during the scheduling loop.
2.  **Copy-on-Write (CoW) Semantics:** State is shared across simulation forks using CoW `PatchSet` objects, `BTreeMap` structures, and slice operations to avoid expensive allocations.
3.  **Two-Way Label Indexing & Phased Evaluation:** The system splits computations into a thread-safe **Prepare** phase (serial) and an optimized **Check** phase (parallel), using bi-directional label indexing to eliminate redundant $O(\text{Terms} \times \text{Pods})$ operations.

---

## Pillar 1: Incremental Indexing (Streaming & Fort)

Traditional schedulers often "pull" data (e.g., scanning all pods on a node) during the filtering phase. In this design, we "push" updates through a reactive indexing pipeline implemented in the `StreamingSnapshotStore`.

### The Pipeline Flow
- **Event-Driven Updates:** The `StreamingSnapshotStore` manages a set of `fort.ManualSharedInformer` instances. When pods or nodes are added/removed, they trigger updates in downstream indices.
- **Affinity Term Registry:** All unique `AffinityTerms` are tracked in a global registry and assigned a unique integer ID.
- **Lazy Term Registration:** When a new term is introduced by a pod, the system lazily backfills its presence across the cluster. To avoid $O(\text{Pods})$ overhead, it uses the **Two-Way Label Index** to only scan pods with matching labels.
- **Topology Aggregation:** Pod updates propagate to **Presence Slices** (tracking pods satisfying terms) and **Forbidden Slices** (tracking anti-affinity constraints) for every topology domain (Node, Zone, Region). These are stored in a `fort.BTreeMap` indexed by `TopologyKey \x00 TopologyValue`.

---

## Pillar 2: Copy-on-Write (CoW) Simulation

The Cluster Autoscaler frequently forks the cluster state to simulate "what-if" scenarios.

### Efficient Forking & Zero-Allocation Iteration
- **PatchSets:** Global registries, domain trackers, and label indices use `common.PatchSet`. When `Fork()` is called, these push a new empty layer onto the stack ($O(1)$).
- **Slice Counting:** Topology domains track term presence using dense `[]int32` slices. Cloning a domain for modification is a single `copy()` instruction.
- **Incremental Caching:** Match results are cached in a `PatchSet`-backed **Match Cache**. This cache is preserved across forks and only evaluates the *difference* when new terms are registered, ensuring evaluation remains $O(1)$ during simulation.

---

## Pillar 3: Phased Evaluation (Prepare & Check)

To maximize parallel throughput, evaluation is split into two phases:

### Phase 1: `PreparePod` (Serial PreFilter)
Before entering the parallel per-node loop, `PreparePod` is called once per pod:
1. It registers any novel terms the pod owns.
2. It identifies which existing terms the pod matches (using the **Two-Way Label Index**).
3. It returns a lightweight `PodAffinityContext` containing resolved integer IDs.

### Phase 2: `FastCheckAffinity` (Parallel Filter)
Inside the per-node loop, the evaluation drops to pure integer slice lookups:

1.  **Incoming Anti-Affinity:**
    ```go
    // O(1) lookup per term I own.
    for _, term := range podCtx.ownedAntiAffinity {
        key := term.topologyKey + "\x00" + node.Labels[term.topologyKey]
        if m, ok := presence.Get(key); ok {
            if m[term.id] > 0 { // Simple slice bounds and value check
                return Unschedulable
            }
        }
    }
    ```

2.  **Incoming Affinity:**
    ```go
    // O(1) lookup per term I own.
    for _, term := range podCtx.ownedAffinity {
        key := term.topologyKey + "\x00" + node.Labels[term.topologyKey]
        m, hasPresence := presence.Get(key)
        if !hasPresence || m[term.id] <= 0 {
            if !selfMatches(podCtx.matchIDs, term.id) { // Self-satisfaction fallback
                return Unschedulable
            }
        }
    }
    ```

3.  **Existing Anti-Affinity (Symmetry):**
    ```go
    // Check if WE violate any existing pod's anti-affinity rules.
    for k, v := range node.Labels {
        key := k + "\x00" + v
        if f, ok := forbidden.Get(key); ok {
            for _, matchID := range podCtx.matchIDs {
                if f[matchID] > 0 && terms[matchID].antiAffinity {
                    return Unschedulable
                }
            }
        }
    }
    ```

---

## Comparison with Standard Kubernetes `InterPodAffinity`

| Metric | Standard K8s `InterPodAffinity` | Fast Incremental Slice-Based |
| :--- | :--- | :--- |
| **Namespace Handling** | Correct. | Correct (uses `AffinityTerm`). |
| **Symmetry Check** | $O(\text{PodsOnNode})$ matches. | $O(\text{MatchedTermsInSystem})$ array lookups. |
| **Simulation** | Multi-step state management. | Native CoW via `PatchSet` and Slices. |
| **Two-Way Indexing** | No (scans pods). | Yes (scans label buckets). |
| **Scalability** | Degrades with pod density. | Performance depends on unique term types. |

## Summary
By combining `PatchSet` CoW indices with dense slice tracking, a Two-Way Label Index, and a phased evaluation lifecycle (`PreparePod` / `FastCheckAffinity`), the Cluster Autoscaler decouples the cost of affinity checks from the raw number of pods. It achieves the speed of a bitwise implementation while correctly supporting Kubernetes' complex, cross-namespace `AffinityTerm` logic.

## Credits
The initial prototype for the Fort reactive indexing pipeline was taken from [https://github.com/bwsalmon/kubernetes/pull/3](https://github.com/bwsalmon/kubernetes/pull/3).
