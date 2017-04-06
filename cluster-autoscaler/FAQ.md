# Frequently Asked Questions

# Basics

### What is Cluster Autoscaler?

Cluster Autoscaler is a standalone program that adjusts the size of a Kubernetes cluster to the current needs.

### When Cluster Autoscaler changes the size of a cluster?

Cluster Autoscaler increases the size of the cluster when: 
* there are pods that failed to schedule on any of the current nodes due to insufficient resources.
* adding a node similar to the nodes currently present in the cluster would help.

Cluster Autoscaler decreases the size of the cluster when some nodes are consistently unneeded for a significant amount of time. A node is unneeded when it has low utilization and all of its important pod can be moved elsewhere. 

### What types of pods can prevent CA from removing a node?

* Kube-system pods that are not run on the node by default.
* Pods that are not backed by a controller objects (so not created by deployment, replica set, job, stateful set etc).
* Pods with local storage.

****************

# Internals 

### How does scale up work?

Scale up creates a watch on the api server looking for all pods. Every 10 seconds (configurable)
it checks for any unschedulable pods. A pod is unschedulable when the Kubernetes scheduler is unable
to find a node that can accommodate the pod. For example a pod can request more CPU that is 
available on any of the cluster nodes. Unschedulable pods are recognized by their PodCondition. 
Whenever a kubernetes scheduler fails to find a place to run a pod it sets "schedulable" 
PodCondition to false and reason to "unschedulable".  If there are any items on the unschedulable 
lists Cluster Autoscaler tries to find a new place to run them. 

It is assumed that the underlying cluster is run on top of some kind of node groups.
Inside a node group all machines have identical capacity and have the same set of assigned labels. 
Thus increasing a size of a node pool will bring a couple of new machines that will be similar 
to these that are already in the cluster - they will just not have the user-created pods (but 
will have all pods run from the node manifest or daemon sets). 

Based on the above assumption Cluster Autoscaler creates template nodes for each of the 
node groups and checks if any of the unschedulable pods would fit to a brand new node, if created.
While it may sound similar to what the real scheduler does, it is currently quite simplified and 
may require multiple iterations before all of the pods are eventually scheduled.
If there are multiple node groups that, if increased, would help with getting some pods running, 
different strategies can be selected for choosing which node group is increased. The default is
random, but other options include selecting the group that can fit the most unschedulable pods,
or the group that will leave the least amount of CPU or Memory available after the scale up.

It may take some time before the nodes from node group appear in Kubernetes. It almost entirely 
depends on the cloud provider and the speed of node provisioning.

### How does scale down work?
 
Every 10 seconds (configurable) Cluster Autoscaler checks which nodes are not needed and can 
be removed. A node is considered not needed when:

* The sum of cpu and memory requests of all pod running on this node is smaller than 50% of node
capacity.

* All pods running on the node (except these that run on all nodes by default like manifest-run pods
or pods created by daemonsets) can be moved to some other nodes. Stand-alone pods which are not
under control of a deployment, replica set, replication controller or job would not be recreated
if the node is deleted so they make a node needed, even if its utilization is low. While 
checking this condition the new locations of all pods are memorized. With that Cluster Autoscaler
knows where each pod can be moved and which nodes depend on which other nodes in terms of 
pod migration. Of course, it may happen that eventually the scheduler will place the pods 
somewhere else. 

* There are no kube-system pods on the node (except these that run on all nodes by default like 
manifest-run pods or pods created by daemonsets).

* There are no pods with local storage. Applications with local storage would lose their 
data if a node is deleted, even if they are replicated.

If a node is not needed for more than 10 min (configurable) then it can be deleted. Cluster Autoscaler
deletes one node at a time to reduce the risk of creating new unschedulable pods. The next node 
can be deleted when it is also not needed for more than 10 min. It may happen just after
the previous node is fully deleted or after some longer time.

What happens when a node is deleted? As mentioned above, all pods should be migrated elsewhere.
For example if node A is deleted then its pods, consuming 400m CPU, are moved to, let's say, node
X where is 450m CPU available. Ok, but what other nodes that also were eligible for deletion? Well,
it depends. If node B also wanted to move its pods, consuming 350m CPU, to node X then it cannot 
do it anymore as there is almost no capacity left. It has to them somewhere else, and it is not sure that
if A had been deleted much earlier than B, during the last 10 min, would always have a place to
move its pods. So the requirement of being unused for 10 min may not be valid anymore for B. 
But if another node C, in case of deletion, can move its pods to node Y then it 
may still do it, because noone touched Y. So C can be deleted immediately after A. And B not. 

Cluster Autoscaler does all of this accounting based on the simulations and memorized new pod location.
They may not always be precise (pods can land elsewhere) but it seems to be a good heuristic so far.

### Does CA work with PodDisruptionBudget in scale down?

From 0.5 CA respects PDB. Before starting to delete a node CA makes sure that there is at least some non-zero PodDisruptionBudget. Then it deletes all pods from a node through the pod eviction api, retrying, if needed, for up to 2 min. During that time other CA activities are stopped. If one of the evictions fails the node is saved and it is not deleted, but another attempt to delete it may be conducted in the near future.

### Does CA respect GracefulTermination in scale down?

CA gives pods at most 1 min graceful termination time. If the pod is not stopped within this 1 min the node is deleted anyway.

### How does CA deal with unready nodes in version <= 0.4.0?

A strict requirement for performing any scale operations is that the size of a node group,
measured on the cloud provider side, matches the number of nodes in Kubernetes that belong to this 
node group. If this condition is not met then all scaling operations are postponed until it is 
fulfilled. 
Also, any scale down will happen only after at least 10 min after the last scale up.

### How does CA deal with unready nodes in version >=0.5.0 ?

From 0.5 CA continues the work even if some (up to 33% or not greater than 3, configurable via flag) percentage of nodes
is unavailable. Once there is more unready nodes in the cluster CA pauses all operations until the situation
improves. If there is less unready nodes but they are concentrated in a particular node group
then this node group may be excluded from scale-ups. 
Prior to 0.5 CA stopped all operations when a single node became unready.

************

# Troubleshooting:

### CA doesn’t work but it used to work yesterday. Why?

Hopefully it is not a bug in Cluster Autoscaler but most likely a problem with the cluster.

* Check If cluster autoscaler is up and running. In version 0.5 it periodically publishes the kube-system/cluster-autoscaler-status config map. Check last update time annotation. Should be no more than 3 min (usually 10 sec old).
* Check kube-system/cluster-autoscaler-status if the cluster and node groups are in the healthy state. If not, check the unready nodes. 

* If you expect some nodes to be deleted but they are not deleted for a long time check:
    * if they contain pods that prevent the node from being deleted (see the corresponding question in the faq).
    * if min/max boundaries you declared for a particular node group allow the scale up.
    * the content of /var/log/cluster-autoscaler.log.

* If you expect some nodes to be added to help some pending pods, but they are not added for a long time check:
    * if the node groups that could potentially accommodate the pods are on their max size.
    * events added by CA to the pod.
    * events on the kube-system/cluster-autoscaler-status config map. 
    * if you have quota on your cloud provider side. 
    * the content of /var/log/cluster-autoscaler.log.

### How can I check what is going on in CA ?

There are three options:

* Logs on the master node, in /var/log/cluster-autoscaler.log.
* kube-system/cluster-autoscaler-status config map.
* Events:
    * on pods (particularly those that cannot be scheduled).
    * on nodes.
    * on kube-system/cluster-autoscaler-status config map.
