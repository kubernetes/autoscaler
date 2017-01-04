### Cluster State Registry
### Handling unready nodes 

### Introduction

Currently ClusterAutoscaler stops working when the number of nodes observed on the cloud provider side doesn’t match to the number of ready nodes on the Kubernetes side. This behavior was introduced in the first version of CA in order to prevent CA from breaking more an already broken cluster. However, client reported feedback indicated that this behavior is oftentimes suboptimal and leads to confusion. This document sketches how the problem of unready nodes will be solved in the next release of K8S.

### Use cases

The number of ready nodes can be different than on the mig side in the following situations:

* [UC1] A new node is being added to the cluster. The node group has been increased but the node has not been created/started/registered in K8S yet. On GCP this usually takes couple minutes. 
Indicating factors:
 -- There was a scale up in the last couple minutes.
 -- The number of missing node is at most the size of executed scale-up.
Suggested action: Continue operations, however include all yet-to-arrive nodes in all scale-up considerations.

* [UC2] A new node is being added to the cluster. The node has registered on the cluster but has not yet switched its state to ready. This should be fixed in couple seconds. Indicating factors:
 -- The unready node is new. CreateTime in the last couple minutes.
Suggested action: Continue operations, however include all yet-to-arrive nodes in all scale-up considerations.

* [UC3] A new node was added to the cluster but failed to start within the reasonable time. There is little chance that it will start anytime soon. Indicating factors:
 -- Node is unready
 -- CreateTime == unready NodeCondition.LastTransitionTime
Suggested action: Continue operations, however do not expand this node pool. The probable scenario is that the node will be picked by scale down soon (after it is unused for long enough - see UC5).

* [UC4] A new node is being added to the cluster. However the cloud provider cannot provision the node within the reasonable time due to either no quota or technical problems. Indicating factors:
 -- The target number of nodes on the cloud provider side is greater than the number of nodes in K8S for the prolonged time (more than couple minutes) and the difference doesn’t change.
 -- There are no new nodes when listing nodes on the cluster provider side.
Suggested action: Reduce the target size of the problematic node group to the current size. 

* [UC5] A new node was provided by the cloud. However, it failed to register. Indicating factors:
  -- There are no new nodes on the cluster provider side that have not appeared in K8S for the long time.
Suggested action: Remove the unregistered nodes one by one.

* [UC6] A node is in an unready state for quite a while (+20min) and the total number of unready/not-present nodes is low (less than XX%). It could either not switched from unready to ready on node registration or something crashed on the node and could not be recovered. Indicating factors:
-- Node condition is unready and last transition time is >= 20 min.
-- The number of TOTAL nodes in K8S is equal to the target number of nodes on the cloud provider side. 
Suggested action: Include the node in scale down, although with greater (configurable) unneeded time.

* [UC7] Some nodes are being removed by cluster autoscaler. Indicating factor:
-- Node is unready and has ToBeRemoved taint.
Suggested action: Continue operations. Nodes should be removed soon.

* [UC8] The number of unjustified (not related to scale-up and scale-down) unready nodes is greater than XX%. Something is broken, possibly due to network partition or generic failure. Indicating factors: 
 -- >XX% of nodes are unready 
Suggested action: halt operations.

### Proposed solution

Introduce a cluster state registry that provides the following information:

* [S1] Is the cluster, in general, in a good enough shape for CA to operate. The cluster is in the good shape if most of the nodes are in the ready state, and the number of nodes that are in the unjustified unready state (not related to scale down or scale up operations) is limited. CA should halt operations if the cluster is unhealthy and alert the system administrator.

* [S2] Is the given Node group, in general, in a good enough shape for CA to operate on it. The NodeGroup is in the good shape if the number of nodes that are unready (but not due to current scale-up/scale-down operations) or not present at all (not yet started by cloud provider) is limited. CA should take extra care about these unhealthy
groups and not scale up them further until the situation improves. 

* [S3] What nodes should soon arrive to the cluster. So that estimator takes them into account and don't ask again for resources for the already handled pods. Also, with that, the estimator won't need to wait for nodes to appear in the cluster.

* [S4] How long the given node group has been missing nodes. If a fixed number of nodes is missing for a long time this may indicate quota problems. Such node groups should be resized to the actual size. 

CA will operate with unready nodes possibly present in the cluster. Such nodes will be picked by scale down as K8S controller manager eventually removes all pods from unready nodes. As the result all of the unready nodes, if not brought back into shape will be removed after being uready for long enough (and possibly replaced by new nodes). 

### Algorithm

The main loop algorithm will look as follows:

1. Get all nodes from the cluster.
2. Check if a cluster is, in general, in good health (most of the nodes are ready). If not, alert the user, skip the iteration and wait 10 seconds. Helps with [UC8], uses [S1]. Also clear the unneeded stats (see 8. and 9.).
3. Check if any of the node groups has nodes that failed to register in K8S. If yes, remove these nodes. Helps with
[UC5].
4. Check if any of the node groups has long-time missing nodes. If yes, reduce the size of the node group by the number
of long-missing nodes. Skip the rest of the iteration. Helps with [UC4], uses [S4].
5. Check if there are any pending pods. Skip pending pods that can be scheduled on the currently available ready nodes (not including nodes that are to be deleted soon [UC7]). 
6. If there are still some pending pods, find which of the node group can be expanded to accommodate them. Skip node groups that are not healthy (contains many unready nodes or nodes that failed to start). Helps with [UC3] uses [S2]. 
7. Estimate the number of needed nodes, account yet-to-come nodes [UC1], [UC2], [S3]. Expand the chosen node group if needed.
8. Calculate the unneeded nodes in the whole cluster, including the unready nodes [UC6]. Unneeded nodes must be monitored every iteration to be sure that they have been unneded for the prolonged time.
9. Try to remove some unneeded node, if there was no recent scale up and the node has been unneeded for more than
10 min. Use higher delay for unready nodes [UC6].

