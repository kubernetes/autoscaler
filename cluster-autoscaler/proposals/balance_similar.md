# Balance similar node groups between zones
##### Author: MaciekPytel

## Introduction
We have multiple requests from people who want to use node groups with the same
instance type, but located in multiple zones for redundancy / HA. Currently
Cluster Autoscaler is randomly adding and deleting nodes in those node groups,
which results in uneven node distribution across different zones. The goal of
this proposal is to introduce mechanism to balance the number of nodes in
similar node groups.

## Constraints
We want this feature to work reasonably well with any currently
supported CA configuration. In particular we want to support both homogenous and heterogenous clusters,
allowing the user to easily choose or implement strategy for defining what kind
of node should be added (i.e. large instance vs several small instances, etc).

Those goals imply a few more specific constraints that we want to keep:
 * We want to avoid balancing size of node groups composed of different instances
(there is no point in trying to balance the number of 2-CPU and 32-CPU
machines).
 * We want to respect size limits of each node group. In particular we shouldn't
 expect that the limits will be the same for corresponding node groups.
 * Some pods may require to run on a specific node group due to using
   labalSelector on zone label, or antiaffinity. In such cases we must scale the
   correct node group.
 * Preferably we should avoid implementing this using existing
   expansion.Strategy interface, as that would likely require non-backward
   compatible refactor of the interface and we want to allow using this feature
   with different strategies.
 * User must be able to disable this logic using a flag.

## General idea
The general idea behind this proposal is to introduce a concept of "Node Group
Set", consisting of one or more of "similar" node groups (the definition of
"similar" is provided in separate section). When scaling up we would split the
nodes between node groups in the same set to make their size as similar as
possible. For example assume node group set made of node groups A (currently 1 node), B (3 nodes), and C (6 nodes).
If we needed to add a new node to the cluster it would go to group A. If we
needed to add 4 nodes, 3 of them would go to group A and 1 to group B.

Note that this does not guarantee that node groups will always have the same
size. Cluster Autoscaler will add exactly as many nodes as are required for
pending pods, which may not be divisible by number of node groups in node group
set. Additionally we scale down underutilized nodes, which may happen to be in
the same node group. Including relative sizes of similar node groups in scale
down logic will be covered by a different proposal later on.

## Implementation proposal
There will be no change to how expansion options are generated in ScaleUp
function. Instead the balancing will be executed after expansion option is
chosen by expansion.Strategy and before node group is resized. The high-level
algorithm will be as follows:
1. During loop generating expansion options create a map {node group -> set of
   pods that pass predicates for this node group}. We already calculate that,
   just need to store it.
2. Take expansion option chosen by expansion.Strategy and call it E. Let NG be node group
   chosen by strategy and K be the number of nodes that need to be added. Let P
   be the set of all pods that will be scheduled thanks to E.
3. Find all node groups "similar" to NG and call them NGS.
4. Check if every pod in P passes scheduler predicates for sample node in every
   node group in NGS (by checking if P is a subset of set we stored in step 1).
   Remove from NGS any node group on which at least one pod from P can't be
   scheduled.
5. Add NG to NGS.
6. Get current and maximum size of all node groups in NGS. Split K between node
   groups as described in example above.
7. Resize all groups in NGS as per result of step 6.

If the user sets the corresponding flag to 'false' we skip step 3,
resulting in a single element in NGS (this makes step 4 no-op and step 6 trivial).

## Similar node groups
We will balance size of similar node groups. We want similar groups to consist
of machine with the same instance type and with the same set of custom labels.
In particular we define "similar" node groups as having:
 * The same Capacity for every resource.
 * Allocatable for each resource within 5% of each other (this number can depend on a
   few different factors and so it's good to have some minor slack).
 * "Free" resources (defined as Allocatable minus resources used by daemonsets and
   kube-proxy) for each resource within 5% of each other (this number can depend on a
   few different factors and so it's good to have some minor slack).
 * The same set of labels, except for zone and hostname labels (defined in
   https://github.com/kubernetes/kube-state-metrics/blob/master/vendor/k8s.io/client-go/pkg/api/unversioned/well_known_labels.go)

---

## Other possible solutions and discussion
There are other ways to implement the general idea than the proposed solution.
This section lists other options that were considered and discusses pros and
cons of each one. Feel free to skip it.

#### [S1]Split selected expansion option
This is the solution described in "Implementation proposal" section.

Pros:
 * Simplest solution.

Cons:
 * If at least a single pending pod uses zone-based scheduling features the
   whole scale-up will likely go to a single node group.
 * We add slightly different nodes than those chosen by expansion.Strategy. In
   particular any zone-based choices made by expansion.Strategy will be
   discarded.
 * We operate on single node groups when creating expansion options, so maximum
   scale-up size is limited by maximum size of a single group.

#### [S2]Update expansion options before choosing them
This idea is somewhat similar to [S1], but the new method would be called
on a set of expansion options before expansion.Strategy chooses one. The new
method could either modify each option to contain a set of scale-ups on similar
node groups.

Pros:
 * Addresses some issues of [S1], but not the issues related to pods using
   zone-based scheduling.

Cons:
 * To fix issues of [S1] the function processing expansion strategies need to be
   more complex.
 * Need to update expansion.Option to be based on multiple NodeGroups, not just
   one. This will make expansion.Strategy considerably more complex and difficult to implement.

#### [S3]Make a wrapper for cloudeprovider.CloudProvider
This solution would work by implementing a NodeGroupSet wrapper implementing
cloudprovider.NodeGroup interface. It would consist of one or more
NodeGroups and internally load balance their sizes.

Pros:
 * We could use an aggregated maximum size for all node groups when creating
   expansion options.
 * We could add additional methods to NodeGroupSet to split pending pods between
   underlying NodeGroups in a smart way, allowing to deal with pod antiaffinity
   and zone based label selectors without completely skipping size balancing.

Cons:
 * NodeGroup contract assumes all nodes in NodeGroup are identical. However we
   want to allow at least different zone labels between node groups in node
   group set.
 * Actually doing the smart stuff with labelSelectors, etc. will be complex and
   hard to implement.

#### [S4]Refactor scale-up logic to include balancing similar node groups
This solution would change how expansion options are generated in
core/scale_up.go. The main ScaleUp function could be largely rewritten to take
balancing node groups into account.

Pros:
 * Most benefits of [S3].
 * Avoids problems caused by breaking NodeGroup interface contract.

Cons:
 * Complex.
 * Large changes to critical parts of code, most likely to cause regression.

#### Discussion
A lot of difficulty of the problem comes from the fact that we can have pods who
can only schedule on some of the node groups in a given node group set.
Such pods require specific config by user (zone-based labelSelector or
antiaffinity) and are likely not very common in most
clusters. Additionally one can argue that having a majority of pods explicitly
specify the zone they want to run in defies the purpose of
automatically balancing the size of node groups between zones in the first place.

If we treat those pods as edge case options [S3] and [S4] don't seem very
attractive. Their main benefit of options [S3] and [S4] is allowing to deal with
such edge cases at the cost of significantly increased complexity.

That leaves options [S1] and [S2]. Once again this is a decision between better
handling of difficult cases versus complexity. This time this tradeoff applies
mostly to expansion.Strategy interface. So far there are no implementations of
this interface that make zone-based decisions and making expansion options more
complex (by consisting of a set of NodeGroups) will make all existing strategies
more complex as well, for no benefit. So it seems that [S1] is the best
available option by virtue of its relative simplicity.
