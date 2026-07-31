# Scale Up Salvo: Multiple Scale-Ups Per Cluster Autoscaler Loop

## Background

Traditionally, the Cluster Autoscaler (CA) runs in a single-threaded, synchronous loop. During each iteration of this loop, CA evaluates the set of unschedulable pods, determines the optimal node group(s) to scale up via its expander, and triggers a single scale-up request (resize) to the cloud provider. After triggering this scale-up, the loop finishes, and CA waits for the next iteration to evaluate any remaining or new unschedulable pods.

However, this single-iteration model introduces significant scale-up latency when handling a large backlog of unschedulable pods. Because standard CA triggers only one scale-up request per loop, a massive backlog—even if homogeneous—often cannot be fully satisfied in a single iteration due to node group size limits or zonal balancing constraints. This forces remaining pods to wait for subsequent loop iterations, introducing artificial delays that could be avoided by executing consecutive scale-ups back-to-back.

---

## High-Level Proposal

We introduce the **Scale Up Salvo** feature. When enabled, this feature allows the Cluster Autoscaler to execute multiple sequential scale-up evaluations and cloud provider resize requests within a single main loop iteration.

Instead of immediately exiting the scale-up phase after the first group is provisioned, CA enters a _salvo loop_ where it:

1. Runs a standard scale-up evaluation (`runSingleScaleUp`) to select and scale up the best node group for a subset of the unschedulable pods.
2. Removes the pods that successfully triggered this scale-up from the active unschedulable pods backlog.
3. Updates the in-memory `ClusterSnapshot` by injecting the faked "upcoming" nodes corresponding to the triggered scale-up, and virtually schedules the triggering pods on them.
4. Evaluates and triggers the next scale-up on the remaining pods, taking into account the updated `ClusterSnapshot` (which ensures scheduling constraints against the newly triggered nodes are correctly respected). This process (repeating steps 1–3) continues until the backlog is cleared or the time budget is exhausted.

This allows CA to trigger multiple distinct node group expansions in a single loop, reducing the time-to-provision for large backlogs.

---

## Configuration and Flags

Two new command-line flags are introduced to configure the salvo behavior:

- **`--salvo-scale-up`** (Default: `false`)
  - Enables or disables the Scale Up Salvo feature. When set to `false` (default), CA preserves its legacy behavior of a single scale-up operation per loop.
- **`--salvo-scale-up-budget`** (Default: `1m`)
  - Specifies the maximum time duration CA is allowed to spend on subsequent scale-ups in a single loop. This prevents the salvo loop from starving other critical autoscaler operations (like scale-down evaluation or status reporting) under heavy pod backlogs. Requires `--salvo-scale-up` to be enabled.

---

## Detailed Design & Implementation

The feature is implemented across `StaticAutoscaler` and the `ScaleUpStatusProcessor` tracking structure.

### 1. Salvo Execution Loop (`runScaleUpSalvo`)

When `--salvo-scale-up` is enabled, the autoscaler calls `runScaleUpSalvo` in `static_autoscaler.go`. The workflow is as follows:

1.  **Track Unschedulable Pods:** Create an in-memory map of all unschedulable pods to help (`podsMap`).
2.  **Time Budget Context:** Initialize a Go `context.Context` with a timeout matching the configured `--salvo-scale-up-budget`.
3.  **Iteration Loop:**
    - Convert the remaining `podsMap` back into a slice of unschedulable pods.
    - Call `runSingleScaleUp` to let the scale-up orchestrator and expander select the best node group and resize it.
    - If the scale-up fails or returns a status indicating that no scale-up was tried or needed, terminate the salvo.
    - Remove the pods that successfully triggered a scale-up in this iteration from the remaining `podsMap`.
    - **Update ClusterSnapshot:** Update the local `ClusterSnapshot` using the helper `addLatestScaleUpResultsToClusterSnapshot`. This makes the new upcoming nodes and virtually scheduled pods visible to subsequent iterations of the salvo loop. (Note: this update is intentionally skipped on the final iteration when the remaining backlog queue becomes empty to avoid unnecessary computations.)
    - **Check Budget:** Verify if `salvoCtx.Err()` is non-nil (deadline exceeded). If so, terminate the salvo.

### 2. Cluster Snapshot Synchronization

To ensure subsequent iterations of the salvo loop do not attempt to scale up again for the same pods (or make incorrect decisions because they don't know about the newly triggered nodes), the `ClusterSnapshot` must be updated dynamically.

The helper method `addLatestScaleUpResultsToClusterSnapshot` handles this update in-place:

1.  **Inject Upcoming Nodes:** For each node group scale-up info in the status, it calculates the delta (`NewSize - CurrentSize`). It clones the node template, marks it with the `autoscaling.k8s.io/upcoming-node` annotation, and adds it to the `ClusterSnapshot` under a unique name suffixed with `salvo-<iteration_index>`.
2.  **Virtual Pod Scheduling:** It calls `SchedulePodOnAnyNodeMatching` on the `ClusterSnapshot` to bind the triggering pods to these newly injected upcoming nodes. This mimics the scheduler and ensures that they do not remain in the "unschedulable pods" queue for the next iteration.

### 3. Preserving Triggering Pods

During standard execution, downstream status processors (such as `ScaleUpStatusProcessor`) process `ScaleUpStatus` and may filter the `ScaleUpStatus.PodsTriggeredScaleUp` slice (e.g., removing certain pods for custom metrics, logging, or special status reporting).

However, the salvo loop requires the **exact, unfiltered set** of pods that actually triggered the cloud-provider scale-up in that iteration to:

- Filter them out of the remaining unschedulable pods list.
- Properly schedule them on the virtual upcoming nodes in the `ClusterSnapshot`.

To resolve this, we preserve the original list by assigning the slice reference (`unfilteredPodsTriggeredScaleUp := scaleUpStatus.PodsTriggeredScaleUp`) immediately after the scale up orchestrator finishes and before any status processors run. This simple reference assignment is sufficient because status processors generally only filter the slice and do not perform in-place modifications on the underlying `Pod` objects, preserving the original list.

The salvo loop then uses the `unfilteredPodsTriggeredScaleUp` list for snapshot synchronization and backlog filtering. This guarantees that both operations process the complete set of triggering pods, even if downstream processors filter the main `PodsTriggeredScaleUp` slice.

---

## Benefits

1.  **Faster Backlog Resolution:** Instead of waiting for subsequent loop iterations (which takes tens of seconds or minutes each), CA can trigger multiple consecutive resizes immediately back-to-back. This is particularly helpful for large homogeneous backlogs that cannot be fully resolved in a single scale-up step due to size or zone-balancing limits.
2.  **Consistent Simulator State:** Updating the local `ClusterSnapshot` dynamically ensures that subsequent iterations in the salvo loop evaluate remaining pods with a highly accurate view of the cluster's future state, respecting capacity limits and pod constraints.
3.  **Operational Safety:** The budget timeout (`--salvo-scale-up-budget`) ensures the salvo loop finishes quickly and never hangs, keeping the main CA loop responsive and preventing starvation of other tasks like scale-down checks.
