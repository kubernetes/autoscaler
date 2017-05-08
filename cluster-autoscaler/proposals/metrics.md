## Cluster Autoscaler Monitoring

## Introduction
Currently the options to monitor Cluster Autoscaler are limited to logs, status
configmap and events. While useful for debugging, none of this options is
paritcularly practical for monitoring Cluster Autoscaler behaviour over time.
This document describes a set of metrics (in Prometheus format) that will be
added to 0.6 Cluster Autoscaler to allow better monitoring of its operations.

## Current state
Cluster Autoscaler already has a metrics endpoint providing some basic metrics.
This includes default process metrics (number of goroutines, gc duration, cpu
and memory details, etc) as well as some custom metrics related to time taken by
various parts of Cluster Autoscaler main loop. Existing metrics will be renamed
and refactored and new metrics will be added.

## Metrics
Cluster Autoscaler needs to monitor some aspects of cluster state as part of its normal operations. It can therefore provide metrics describing its own operations as well as some metrics related to general state of the cluster.

### Cluster state
* __cluster_healthy (gauge)__ - CA stops all operations if significant number of nodes
are unready (by default 33% as of CA 0.5.4). This metric returns 1 if CA
believes cluster to be healthy enough to continue normal operation, 0 otherwise.
* __nodes_count (gauge)__ - number of nodes in different states in cluster. This
metric is labeled with different node states:
  * ready,
  * unready,
  * notStarted.
* __unschedulable_pods_count (gauge)__ - number of unschedulable (“Pending”) pods in
the cluster.

### Cluster Autoscaler execution
This metrics are refactored from currently existing metrics and track execution
of various parts of Cluster Autoscaler loop.

* __last_activity (gauge)__ - last time certain part of cluster autoscaler logic
executed. Represented with unix timestamp. Uses the following set of labels:
  * main - main loop iteration started.
  * autoscaling - current state of the cluster has been updated, started autoscaling
logic.
  * scaleup - autoscaler will check if scale up is necessary.
  * scaledown - autoscaler will try to scale down some nodes.
  Not all of the above have to happen in every loop. For example if CA adds a new
node it will completely skip scale down logic in this loop.
* __duration_microseconds (summary)__ - time taken by various parts of CA loop. Uses
the following set of labels:
  * main - time of the whole iteration of main loop.
  * updateClusterState - time used by CA to get node status from API server and
update internal data structures.
  * scaleup - time used to check if new node are necessary and add them.
  * findUnneeded - time required to find nodes that are candidates for removal.
  * scaledown - time required to verify unneeded nodes are really unnecessary and
remove them.

New labels may be added if we add more features or additional logic to Cluster
Autoscaler.

### Cluster Autoscaler operations
This metrics describe internal state and actions taken by Cluster Autoscaler.

* __autoscaler_error (counter)__ - This metric is set to 1 if main loop iteration is
 skipped or fails to resize cluster due to any error. Otherwise it’s set to 0.
  * Growing autoscaler_error count signifies an internal error in CA or a problem
 with underlying infrastructure preventing normal CA operation. Example errors include:
    * failed to get list of nodes or pods from API server,
    * failed to retrieve node group size from cloud provider,
    * failed to update node group size,
    * error in CA simulations,
    * error updating internal data structures.
  * Not every condition causing CA to skip part of main loop is an error in this
 context. For example having pending pods after reaching maximum cluster size
 causes ScaleUp function to return error resulting in CA skipping the rest of
 main loop. This is expected behaviour of CA and it should not be
 counted as error.
  * Labeled by high-level error type, using following set of labels:
    * cloudProviderError - failed to get node group info, set node group size, etc.
    * apiCallError - error related to call to k8s api server (ex. get nodes, get
 pods).
    * internalError - error in any other part of CA logic.
* __scale_up_count (counter)__ - number of nodes successfully added by CA. In this
 context we consider node as successfully added after updating node group size (without
 waiting for actual vm to spin up, run a kubelet, etc).
* __scale_down_count (counter)__ - number of nodes removed by CA. Labeled by reason
 for removing node:
  * empty,
  * underutilized,
  * unready.
* __evictions_count (counter)__ -  number of pods evicted by CA.
* __unneeded_nodes_count (gauge)__ - number of nodes that are currently considered
 for scale down by CA.
