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

All the metrics are prefixed with `cluster_autoscaler_`.

### Cluster state

| Metric name | Metric type | Labels | Description |
| ----------- | ----------- | ------ | ----------- |
| cluster_safe_to_autoscale | Gauge | | Whether or not cluster is healthy enough for autoscaling. 1 if it is, 0 otherwise. |
| nodes_count | Gauge | `state`=&lt;node-state&gt; | Number of nodes in cluster. |
| unschedulable_pods_count | Gauge | | Number of unschedulable ("Pending") pods in the cluster. |
| node_groups_count | Gauge | `node_group_type`=&lt;node-group-type&gt; | Number of node groups managed by CA. |

* `cluster_safe_to_autoscale` indicates whether cluster is healthy enough for autoscaling. CA stops all operations if significant number of nodes are unready (by default 33% as of CA 0.5.4).
* `nodes_count` records the total number of nodes, labeled by node state. Possible
states are `ready`, `unready`, `notStarted`.
* `node_groups_count` records the number of currently managed node groups. It's
  useful when using dynamic configuration or Node Autoprovisioning. Types of
  node group are `autoscaled` (managed by CA but not created by NAP) and `autoprovisioned` (created by NAP and managed by CA).

### Cluster Autoscaler execution
This metrics are refactored from currently existing metrics and track execution
of various parts of Cluster Autoscaler loop.

| Metric name | Metric type | Labels | Description |
| ----------- | ----------- | ------ | ----------- |
| last_activity | Gauge | `activity`=&lt;autoscaler-activity&gt; | Last time certain part of CA logic executed |
| function_duration_seconds | Histogram | `function`=&lt;autoscaler-function&gt; | Time taken by various parts of CA main loop. |

* `last_activity` records last time certain part of cluster autoscaler logic
executed. Represented with unix timestamp. autoscaler-activity values are:
  * `main` - main loop iteration started.
  * `autoscaling` - current state of the cluster has been updated, started autoscaling
logic.
  * `scaleUp` - autoscaler will check if scale up is necessary.
  * `scaleDown` - autoscaler will try to scale down some nodes.

  Not all of the above have to happen in every loop. For example if CA adds a new
node it will completely skip scale down logic in this loop.
* `function_duration_seconds ` summarizes time taken by different functions executed in
  main CA goroutine. Uses the following set of values for autoscaler-function:
  * `main` - duration of the whole iteration of main loop.
  * `updateClusterState` - time used by CA to get node status from API server and
update internal data structures.
  * `scaleUp` - time used to check if new node are necessary and add them.
  * `findUnneeded` - time required to find nodes that are candidates for removal.
  * `scaleDown` - time required to verify unneeded nodes are really unnecessary and
remove them.

New labels may be added to both `last_activity` and `function_duration_seconds` if we add more features or additional logic to Cluster Autoscaler.

### Cluster Autoscaler operations
This metrics describe internal state and actions taken by Cluster Autoscaler.

| Metric name | Metric type | Labels | Description |
| ----------- | ----------- | ------ | ----------- |
| errors_total | Counter | `type`=&lt;error-type&gt; | The number of CA loops failed due to an error. |
| scaled_up_nodes_total | Counter | | Number of nodes added by CA. |
| scaled_down_nodes_total | Counter | `reason`=&lt;scale-down-reason&gt; | Number of nodes removed by CA. |
| failed_scale_ups_total | Counter | `reason`=&lt;failure-reason&gt; | Number of times scale-up operation has failed. |
| evicted_pods_total | Counter | | Number of pods evicted by CA. |
| unneeded_nodes_count | Gauge | | Number of nodes currently considered unneeded by CA. |

* `errors_total` counter increases every time main CA loop encounters an error.
  * Growing `errors_total` count signifies an internal error in CA or a problem
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
  * Possible error types are:
    * `cloudProviderError` - failed to get node group info, set node group size, etc.
    * `apiCallError` - error related to call to k8s api server (ex. get nodes, get
 pods).
    * `internalError` - error in any other part of CA logic.
* `scaled_up_nodes_total` counts the number of nodes successfully added by CA. In this
 context we consider node as successfully added after updating node group size (without
 waiting for actual vm to spin up, run a kubelet, etc).
* `failed_scale_ups_total` counts the number of unsuccessful scale-up
  operations performed by CA. This includes both getting error from cloud
  provider and new nodes failing to boot up and register within timeout. It
  does not include reaching maximum cluster size (as CA doesn't attempt scale-up
  at all in that case).
* `scaled_down_nodes_total` counts the number of nodes removed by CA. Possible
scale down reasons are `empty`, `underutilized`, `unready`.

### Node Autoprovisioning operations

This metrics describe operations and state related to Node Autoprovisioning
feature.

| Metric name | Metric type | Labels | Description |
| ----------- | ----------- | ------ | ----------- |
| nap_enabled | Gauge | | Whether or not Node Autoprovisioning is enabled. 1 if it is, 0 otherwise. |
| created_node_groups_total | Counter | | Number of node groups created by Node Autoprovisioning. |
| deleted_node_groups_total | Counter | | Number of node groups deleted by Node Autoprovisioning. |

