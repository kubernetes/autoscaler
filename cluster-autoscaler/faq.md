# Frequently Asked Questions

# Basics

## What is Cluster Autoscaler?

Cluster Autoscaler is a standalone program that adjusts the size of a Kubernetes cluster to the current needs.

## When Cluster Autoscaler changes the size of a cluster?

Cluster Autoscaler increases the size of the cluster when: 
* there are pods that failed to schedule on any of the current nodes due to insufficient resources.
* adding a node similar to the nodes currently present in the cluster would help.

Cluster Autoscaler decreases the size of the cluster when some nodes are consistently unneeded for a significant amount of time. A node is unneeded when it has low utilization and all of its important pod can be moved elsewhere. 

## What types of pods can prevent CA from removing a node?

* Kube-system pods that are not run on the node by default.
* Pods that are not backed by a controller objects (so not created by deployment, replica set, job, stateful set etc).
* Pods with local storage.

## How does CA work with PodDisruptionBudget in scale down?

From 0.5 CA respects PDB. Before starting to delete a node CA makes sure that there is at least some non-zero PodDisruptionBudget. Then it deletes all pods from a node through the pod eviction api, retrying, if needed, for up to 2 min. During that time other CA activities are stopped. If one of the evictions fails the node is saved and it is not deleted, but another attempt to delete it may be conducted in the near future.

## How does CA respect GracefulTermination in scale down?

CA gives pods at most 1 min graceful termination time. If the pod is not stopped within this 1 min the node is deleted anyway.

# Troubleshooting:

## CA doesn’t work but it used to work yesterday. Why?

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

## How can I check what is going on in CA ?

There are three options:

* Logs on the master node, in /var/log/cluster-autoscaler.log.
* kube-system/cluster-autoscaler-status config map.
* Events:
    * on pods (particularly those that cannot be scheduled).
    * on nodes.
    * on kube-system/cluster-autoscaler-status config map.
