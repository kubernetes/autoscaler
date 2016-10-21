# Cluster Autoscaler

# Introduction

Cluster Autoscaler is a tool that automatically adjusts the size of the Kubernetes cluster when:
* there is a pod that doesn’t have enough space to run in the cluster
* some nodes in the cluster are so underutilized, for an extended period of time, 
that they can be deleted and their pods will be easily placed on some other, existing nodes.  

# Deployment

Cluster Autoscaler runs on the Kubernetes master node (at least in the default setup on GCE and GKE). 
It is possible to run customized Cluster Autoscaler inside of the cluster but then extra care needs
to be taken to ensure that Cluster Autoscaler is up and running. User can put it into kube-system
namespace (Cluster Autoscaler doesn't scale down node with non-manifest based kube-system pods running
on them) and mark with `scheduler.alpha.kubernetes.io/critical-pod` annotation (so that the rescheduler, 
if enabled, will kill other pods to make space for it to run). 

Right now it is possible to run Cluster Autoscaler on:
* GCE http://kubernetes.io/docs/admin/cluster-management/#cluster-autoscaling
* GKE https://cloud.google.com/container-engine/docs/cluster-autoscaler
* AWS https://github.com/kubernetes/contrib/blob/master/cluster-autoscaler/cloudprovider/aws/README.md

# Scale Up

Scale up creates a watch on the api server looking for all pods. Every 10 seconds (configurable)
it checks for any unschedulable pods. A pod is unschedulable when the Kubernetes scheduler is unable
to find a node that can accomodate the pod. For example a pod can request more CPU that is 
available on any of the cluster nodes. Unschedulable pods are reconginzed by their PodCondition. 
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
While it may sound similar to what the real scheduler does it is currently quite simplified and 
may require multiple iterations before all of the pods are eventually scheduled.
If there are multiple node groups that, if increased, would help with getting some pods running, 
one of them is selected at random. 

It may take some time before the nodes from node group appear in Kubernetes. It almost entirely 
depends on the cloud provider and the speed of node provisioning.

# Scale Down

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
pod migration. Of course, it may happen that eventaully the scheduler will place the pods 
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
For example if node A is deleted then its pods, consumig 400m CPU, are moved to, let's say, node
X where is 450m CPU available. Ok, but what other nodes that also were eligible for deletion? Well,
it depends. If node B also wanted to move its pods, consumig 350m CPU, to node X then it cannot 
do it anymore as there is almost no capacity left. It has to them somewhere else, and it is not sure that
if A had been deleted much earlier then B, during the last 10 min, would always have a place to
move its pods. So the requirement of being unused for 10 min may not be valid anymore for B. 
But if another node C, in case of deletion, can move its pods to node Y then it 
may still do it, because noone touched Y. So C can be deleted immediatelly after A. And B not. 

Cluster Autoscaler does all of this acounting based on the simulations and memorized new pod location.
They may not always be precise (pods can land elswehere) but it seems to be a good heuristic so far.


# When scaling is executed

A strict requirement for performing any scale operations is that the size of a node group,
measured on the cloud provider side, matches the number of nodes in Kubernetes that belong to this 
node group. If this condition is not met then all scaling operations are postponed until it is 
fulfilled. 
Also, any scale down will happen only after at least 10 min after the last scale up.