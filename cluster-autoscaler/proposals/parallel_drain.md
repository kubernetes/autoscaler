# Cluster Autoscaler parallel drain

Author: x13n

## Background

Scale down of non-empty nodes is slow. We are only draining one node at a time.
This is particularly problematic in large clusters. While empty nodes can be
removed at the same time, non-empty nodes being removed sequentially, which can
take hours in thousand node clusters. We want to speed it up by allowing
parallel scale down of multiple non-empty nodes. Ideally, we'd like to drain as
many nodes as possible simultaneously, without leaving workloads hanging.

## High level proposal

### Algorithm triggering

With the old algorithm scale down was disabled if deleting of non-empty node
from previous iteration did not yet complete. For the new algorithm we plan to
relax this.

### Algorithm operation

The algorithm will internally keep track of the state of nodes in the cluster,
which will be updated every loop iteration. The state will allow answering
following questions:

*   What is the set of nodes which can be removed from the cluster right now in
    parallel? There pods from the nodes in the set must be schedulable on other
    nodes in the cluster. The set will be called **candidate\_set** later in the
    document.
*   For how long is given node in the candidate\_set?

The candidate\_set will be built in a greedy manner by iterating over all nodes.
To verify if a given node can be put in the candidate\_set we will use scheduler
simulation to binpack the pods on the nodes starting from the nodes on the other
end of the list (details below). Since the simulation can be time-consuming, it
will be time bound. This may lead to a slower scale down, but will prevent scale
up from starving.

In each iteration the contents of candidate\_set will be updated. Nodes can be
added, removed or stay in candidate\_set. For each node in the candidate\_set we
will keep track of how long it is in there. If node is removed and then re-added
to candidate set the timer is reset.

To trigger node deletion will use the already present ScaleDownUnneededTime and
ScaleDownUnreadyTime parameters. If in given CA loop iteration there are nodes
which have been in candidate\_set for more than ScaleDownUnneededTime (or is not
ready and is in candidate\_set for more than scaleDownUnreadyTime), the
actuation of scaledown is triggered. The number of nodes which are actively
scaled down will be limited by MaxDrainParallelism and MaxScaleDownParallelism
configuration parameters. We will configure separate limits for scaling down
empty and non-empty nodes to allow quick scale-down of empty nodes even if
lengthy drains of non-empty nodes are in progress.

The existing `--max-empty-bulk-delete` flag will be deprecated and eventually
removed in favor of `--max-scale-down-parallelism` and --max-drain-parallelism.

The actuation will be done in a similar fashion as in the old algorithm. For a
given set of nodes to be scaled down we will synchronously taint the nodes. Then
separate goroutines to perform draining and node deletions will be run.

Current scale-down algorithm uses SoftTainting to limit the chance that new pods
will be moved toward scaled down candidates. New algorithm will use the same
mechanism for nodes in the candidate\_set.

The state in the scale-down algorithm will be updated incrementally based on the
changes to the set of nodes and pods in the cluster snapshot.

## Detailed design

### Existing code refactoring

The existing scale down code lives mostly in the ScaleDown object, spanning
scale\_down.go file with 1.5k lines of code. As a part of this effort, ScaleDown
object will undergo refactoring, extracting utils to a separate file. ScaleDown
itself will become an interface with two implementations (both relying on common
utils): existing version and the new algorithm described below. This will allow
easy switching between algorithms with a flag: different flag values will pick
different ScaleDown interface implementations.

As a part of the refactoring, we will combine `FastGetPodsToMove` and
`DetailedGetPodsForMove` into a single `PodsToMove` function. The
`checkReferences` flag will be dropped and we will always do a detailed check.
We are now using listers so doing a detailed check does not add extra API calls
and should not add too much to execution time.

Due to this change, doing the simulation pass twice will no longer make sense.
New implementation of the algorithm will perform it only once, as described in
the section below.

Actuation logic will be separated from ScaleDown decisions, since only the
decision-making process is going to change. In particular, the SoftTaining logic
will be extracted from ScaleDown. Instead, NeededNodes()/UnneededNodes() methods
will be used to fetch nodes for (un)tainting.

### Algorithm State

The algorithm is stateful. On each call to UpdateUnneededNodes (happening every
CA loop iteration) the state will be updated to match the most recent cluster
state snapshot. The most important fields to be held in algorithm state are:

*   deleted\_set: set of names of nodes which are being deleted
    (implementation-wise we will probably use NodeDeletionTracker)
*   candidate\_set: set of names for nodes which can be deleted
*   non\_candidate\_set: set of names of nodes for which we tried to simulate
    drain and it failed. For each node we keep time when the simulation was
    done.
*   pod\_destination\_hints:
    *   stores destination nodes computed during draining simulation. For each
        pod from nodes in candidate\_set and deleted\_set it keeps the
        destination node assigned during simulation.
    *   map: source node names -> pod UID -> destination node names
*   recent\_evictions: set of pods that were recently successfully evicted from
    a node

### UpdateUnneededNodes

UpdateUnneededNodes will be called from RunOnce every CA loop as it is now. On
every call the internal state is updated to match the recent snapshot of the
cluster state.

Single loop iteration of algorithm:

*   Build node\_infos list
*   Build new algorithm state (steps below)

#### State update

*   Verify the pods from nodes currently being scaled down can still be
    rescheduled
    *   Iterate over all the nodes in deleted\_set
    *   Validate that we can still reschedule remaining pods on other nodes in
        the cluster
        *   Implementation-wise this is going to reuse the logic from
            filterOutSchedulablePodListProcessor.
    *   Use pod\_destination\_hints and fallback to linear searching
    *   The general algorithm here follows the same rules as "Scaledown
        simulation loop" described in more details below.
        *   Implementation-wise verification for nodes being scaled down and
            looking for new candidates will be shared code.
        *   Important differences:
            *   We will update the
                [pdbs\_remaining\_disruptions](#pdbs-checking) structure but
                skip PDB checking for simulation for nodes being scaled down
            *   We are not updating candidate\_set
*   Verify recently evicted pods can be scheduled
    ([details](#race-conditions-with-other-controllers))
    *   Iterate over all the pods in recent\_evictions
    *   If a pod was added to the list N or more seconds ago, remove it from
        recent\_evictions
    *   If a pod has a known owner reference, check if parent object has created
        all the replicas
    *   If there was no known owner reference or the parent object doesn't have
        all the replicas, try to schedule the pod
        *   Only do this up to `target\_replicas - current\_replicas` times.
    *   Verification fails if any pod failed to be scheduled
*   If either of the above verifications failed, break the algorithm here;
    candidate\_set is empty
*   Use ScaleDownNodeProcessor.GetScaleDownCandidates to get a list of scale
    down eligible nodes: ones that may end up in candidate\_set.
*   Clone the cluster snapshot to be used for the simulation, so simulation
    results don't leak through it.
*   Set target\_node\_pointer to the last node in node\_infos
*   Scaledown simulation loop:
    *   Fetch next node eligible[^1] for scale down
    *   If we run out of time for simulation break the loop; do not continue
        with next node
    *   Simulate the node scaledown
        *   Fork() cluster snapshot.
        *   List the pods that need to be drained from the node. Logic already
            implemented in GetPodsForDeletionOnNodeDrain.
            *   If the function returns an error add node to non\_candidate\_set
                and continue with the next node.
        *   For each pod on the node run a predicate checker to test if it fits
            one of the nodes in the cluster. Pods should be sorted in some way
            so we limit the chance of making different decisions each loop
            iteration.
            *   First try to use pod\_destination\_hints:
                *   Do predicate checking on the node pointed by
                    pod\_destination\_hints.
                *   If scheduling is possible simulate it
                    (sourceNodeInfo.RemovePod(), targetNodeInfo.AddPod(), update
                    new pod\_destination\_hints, update
                    pdb\_remaining\_disruptions)
            *   If scheduling to node pointed by pod\_destination\_hints is not
                possible try to find other node using FitsAnyNodeMatching
                *   Skip nodes which are part of currently built candidate\_set
                    or deleted\_set
                *   Skip nodes which are not in GetPodDestinationCandidates()
                *   This behavior may be optimized in the future, see [this
                    section](#potential-simulation-optimization).
                *   If currently considered pod can be scheduled, simulate
                    scheduling (sourceNodeInfo.RemovePod(),
                    targetNodeInfo.AddPod(), update pod\_destinations\_hints,
                    pdbs\_remaining\_disruptions) and restart loop for the next
                    pod
                *   If the currently considered pod cannot be scheduled to any
                    node, Revert() the cluster snapshot and start simulation for
                    the next node. Add the current node to the
                    non\_candidate\_set.
        *   If all the pods from the considered node find new homes, mark the
            source node as candidate for scaledown (add to new candidate\_set,
            remove from non\_candidate\_set) and Commit() the cluster snapshot.


#### Caveats

*   The number of iterations will be time bound yet we may require that at least
    a fixed number of nodes is evaluated each run.

### UnneededNodes() / NeededNodes()

Return node lists built based on candidate\_set/non\_candidate\_set,
respectively.

### StartDeletion

The responsibility of StartDeletion is to trigger actual scaledown of nodes in
candidate\_set (or a subset of those). The method will keep track of empty and
non-empty nodes being scaled down (separately). The number of nodes being scaled
down will be bounded by MaxDrainParallelism and MaxScaleDownParallelism options
passed in as CA Flags.

Method steps in pseudocode:

*   Delete N empty nodes, up to MaxScaleDownParallelism, considering nodes
    already in the deleted\_set.
*   Delete min(MaxScaleDownParallelism - N, MaxDrainParallelism) non-empty
    nodes.
    *   synchronously taint the nodes to be scaled down as we currently do
    *   schedule draining and node deletion as a separate go routine for each
        node
    *   move nodes from candidate\_set to deleted\_set

### Potential simulation optimization

The algorithm described above may be suboptimal in its approach to scheduling
pods. In particular, it can lead pods to jump between nodes as they are added to
the candidate\_set, which increases the usage of PreFilters/Filters in the
scheduler framework, which can be costly.

To optimize that, we may use a cyclic pointer that would be used as a starting
point in scheduling simulations. Such approach may limit the number of calls to
the scheduler framework.

Implementation-wise, we could reuse existing SchedulerBasedPredicateChecker. In
order to do this, the algorithm for picking nodes for scheduling would be passed
to SchedulerBasedPredicateChecker as a strategy. By default, it would be the
existing round robin across all nodes, but ScaleDown simulation would inject a
more sophisticated one, relying a pointer managed by the scale down logic.

### PDBs checking

Throughout the loop we will keep the quotas remaining for PDBs in
pdbs\_remaining\_disruptions structure. The structure will be computed at the
beginning of UpdateUnneededNodes and for each PDB it will hold how many Pods
matching this PDB can still be disrupted. Then we decrease the counters as we go
over the nodes and simulate pods rescheduling. The initial computation and drain
simulation takes into account the state of the pod. Specifically if Pod is not
Healthy it will be subtracted from the remaining quota on the initial
computation and then not subtracted again on drain simulation.


### PreFilteringScaleDownNodeProcessor changes

We need to repeat checks from PreFilteringScaleDownNodeProcessor in
UpdateUnneededNodes anyway as the latter simulate multi node deletion so it
seems we can drop PreFilteringScaleDownNodeProcessor if new scale down algorithm
is enabled. Not crucial as checks are not costly.


### Changes to ScaleDownStatus

ScaleDownStatus does not play very well with the concept of multiple nodes being
scheduled down. The existing ScaleDownInProgress value in ScaleDownResult enum
represents a state in which CA decides to skip scale down logic due to ongoing
deletion of a single non-empty node. In the new algorithm, this status will no
longer be emitted. Instead, a new ScaleDownThrottled status will be emitted when
max parallelism for both empty and non-empty nodes was reached and hence no new
deletion can happen.

### Changes to clusterstate

Cluster state holds a map listing unneeded candidates for each node group. We
will keep the map. The semantic of unneeded nodes will change though. With the
old algorithm there was no guarantee that all the "unneeded" nodes can be
removed altogether - the drain simulation is done independently for each
candidate node. The parallel scaledown algorithm validates that all unneeded
nodes can be dropped together (modulo difference in behavior of scheduler
simulation and actual scheduler).

### Race conditions with other controllers

Whenever Cluster Autoscaler evicts a pod, it is expected that some external
controller will create a similar pod elsewhere. However, it takes a non-zero
time for controllers to react and hence we will keep track of all evicted pods
on a dedicated list (recent\_evictions). After each eviction, the pod object
along with the eviction timestamp will be added to the list and kept for a
preconfigured amount of time. The pods from that list will be injected back into
the cluster before scale down simulation as a safety buffer, except when CA is
certain the replacement pods were either already scheduled or don't need
replacing. This can be verified by examining the parent object for the pods
(e.g. ReplicaSet).

In the initial implementation, for the sake of simplicity, parallel drain will
not be triggered as long as there are already any nodes in the deleted\_set.

## Monitoring

The existing set of metrics will suffice for the sake of monitoring performance
of parallel scale down (i.e. scaled\_down\_nodes\_total,
scaled\_down\_gpu\_nodes\_total, function\_duration\_seconds). We may extend the
set of function metric label values for function\_duration\_seconds to get a
better visibility into empty vs. non-empty scale down duration.

## Rollout

Flag controlled: when `--max-drain-parallelism` != 1, the new logic is enabled.
Old logic will stay for a while to make rollback possible. The flag will be
introduced in version 1.25, default it to >1 in 1.26 and eventually drop the old
logic in 1.27 or later.

### Flag deprecation

The existing `--max-empty-bulk-delete` flag will be deprecated and eventually
removed: new flags will no longer refer to empty nodes. During the deprecation
period, `--max-scale-down-parallelism` will default to the value of
`--max-empty-bulk-delete`.

## Notes

[^1]:

    The reasons for node to not be eligible for scale down include:

    *   node is currently being scaled down
    *   node is not in scale down candidates
    *   removing the node would move node pool size below min
    *   removing the node would move cluster resources below min
    *   node has no-scaledown annotation
    *   node utilization is too high
    *   node is already marked as destination

