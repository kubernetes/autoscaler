## Cluster Autoscaler Monitoring

## Introduction
Currently the options to monitor Cluster Autoscaler are limited to logs, status
configmap and events. While useful for debugging, none of this options is
particularly practical for monitoring Cluster Autoscaler behaviour over time.
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
| node_groups_count | Gauge | `node_group_type`=&lt;node-group-type&gt; | Number of node groups managed by CA. |
| unschedulable_pods_count | Gauge | `type`=&lt;`unschedulable` or `scheduler_unprocessed`&gt; | Number of unschedulable ("Pending") pods in the cluster. |
| max_nodes_count | Gauge | | Maximum number of nodes in all node groups. |
| cluster_cpu_current_cores | Gauge | | Current number of cores in the cluster, minus deleting nodes. |
| cpu_limits_cores | Gauge | `direction`=&lt;`minimum` or `maximum`&gt; | Minimum and maximum number of cores in the cluster. |
| cluster_memory_current_bytes | Gauge | | Current number of bytes of memory in the cluster, minus deleting nodes. |
| memory_limits_bytes | Gauge | `direction`=&lt;`minimum` or `maximum`&gt; | Minimum and maximum number of bytes of memory in cluster. |

* `cluster_safe_to_autoscale` indicates whether cluster is healthy enough for autoscaling. CA stops all operations if significant number of nodes are unready (by default 33% as of CA 0.5.4).
* `nodes_count` records the total number of nodes, labeled by node state. Possible
states are `ready`, `unready`, `notStarted`.
* `node_groups_count` records the number of currently managed node groups. It's
  useful when using dynamic configuration or Node Autoprovisioning. Types of
  node group are `autoscaled` (managed by CA but not created by NAP) and `autoprovisioned` (created by NAP and managed by CA).

#### Per-node-group cluster state

The following per-node-group metrics are emitted only when the
`--emit-per-nodegroup-metrics` flag is set to `true`.

| Metric name | Metric type | Labels | Description |
| ----------- | ----------- | ------ | ----------- |
| node_group_min_count | Gauge | `node_group` | Minimum number of nodes in the node group. |
| node_group_max_count | Gauge | `node_group` | Maximum number of nodes in the node group. |
| node_group_target_count | Gauge | `node_group` | Target number of nodes in the node group as set by CA. |
| node_group_healthiness | Gauge | `node_group` | Whether the node group is healthy enough for autoscaling. 1 if healthy, 0 otherwise. |
| node_group_backoff_status | Gauge | `node_group`, `reason` | Whether the node group is in backoff (and therefore not autoscaling). 1 if in backoff, 0 otherwise. |

### Cluster Autoscaler execution
This metrics are refactored from currently existing metrics and track execution
of various parts of Cluster Autoscaler loop.

| Metric name | Metric type | Labels | Description |
| ----------- | ----------- | ------ | ----------- |
| last_activity | Gauge | `activity`=&lt;autoscaler-activity&gt; | Last time certain part of CA logic executed |
| function_duration_seconds | Histogram | `function`=&lt;autoscaler-function&gt; | Time taken by various parts of CA main loop. |
| function_duration_quantile_seconds | Summary | `function`=&lt;autoscaler-function&gt; | Quantiles of time taken by various parts of CA main loop, computed over a sliding 1h window. Complementary to `function_duration_seconds` for cases where Histogram quantiles are insufficient. |
| pending_node_deletions | Gauge | | Number of nodes that haven't been removed or aborted after the scale-down phase finished. |

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
| scaled_up_nodes_total | Counter | `gpu_resource_name`=&lt;gpu-resource-name&gt;, `gpu_name`=&lt;gpu-name&gt;, `dra_drivers`=&lt;dra-driver-names&gt; | Number of nodes added by CA. |
| scaled_up_gpu_nodes_total | Counter | `gpu_resource_name`=&lt;gpu-resource-name&gt;, `gpu_name`=&lt;gpu-name&gt; | **Deprecated since 1.36.0**. Number of GPU nodes added by CA, by GPU name. |
| failed_scale_ups_total | Counter | `reason`=&lt;failure-reason&gt;, `gpu_resource_name`=&lt;gpu-resource-name&gt;, `gpu_name`=&lt;gpu-name&gt;, `dra_drivers`=&lt;dra-driver-names&gt; | Number of times scale-up operation has failed. |
| failed_node_creations_total | Counter | `reason`=&lt;failure-reason&gt; | Number of nodes that CA failed to add. Per-node granularity, contrasted with `failed_scale_ups_total` which is per-operation: a single failed scale-up that requested 5 nodes increments this metric by 5 and `failed_scale_ups_total` by 1. |
| failed_gpu_scale_ups_total | Counter | `reason`=&lt;failure-reason&gt;, `gpu_resource_name`=&lt;gpu-resource-name&gt;, `gpu_name`=&lt;gpu-name&gt; | **Deprecated since 1.36.0**. Number of times scale-up operation has failed for GPU node groups. |
| scaled_down_nodes_total | Counter | `reason`=&lt;scale-down-reason&gt;, `gpu_resource_name`=&lt;gpu-resource-name&gt;, `gpu_name`=&lt;gpu-name&gt;, `dra_drivers`=&lt;dra-driver-names&gt; | Number of nodes removed by CA. |
| scaled_down_gpu_nodes_total | Counter | `reason`=&lt;scale-down-reason&gt;, `gpu_resource_name`=&lt;gpu-resource-name&gt;, `gpu_name`=&lt;gpu-name&gt; | **Deprecated since 1.36.0**. Number of GPU nodes removed by CA, by reason and GPU name. |
| evicted_pods_total | Counter | `eviction_result`=&lt;`succeeded` or `failed`&gt; | Number of pods evicted by CA. |
| unneeded_nodes_count | Gauge | | Number of nodes currently considered unneeded by CA. |
| unremovable_nodes_count | Gauge | `reason`=&lt;unremovable-reason&gt; | Number of nodes currently considered unremovable by CA, broken down by reason. |
| scale_down_in_cooldown | Gauge | | Whether scale-down is currently in cooldown. 1 if it is, 0 otherwise. |
| old_unregistered_nodes_removed_count | Counter | | Number of unregistered nodes removed by CA. |
| overflowing_controllers_count | Gauge | | Number of controllers that own a large set of heterogenous pods, preventing CA from treating these pods as equivalent during binpacking. |
| skipped_scale_events_count | Counter | `direction`=&lt;scaling-direction&gt;, `reason`=&lt;skipped-scale-reason&gt; | Number of times scaling has been skipped due to a resource limit being reached, or similar event. |
| created_node_groups_total | Counter | `group_type`=&lt;node-group-type&gt; | Number of node groups created by Node Autoprovisioning. |
| deleted_node_groups_total | Counter | `group_type`=&lt;node-group-type&gt; | Number of node groups deleted by Node Autoprovisioning. |
| node_taints_count | Gauge | `type`=&lt;taint-type&gt; | Number of node taints currently present in the cluster, grouped by taint type. |
| inconsistent_instances_migs_count | Gauge | | Number of migs where instance count according to `InstanceGroupManagers.List()` differs from the results of `Instances.List()`. This can happen when some instances are abandoned or a user edits instance `created-by` metadata. |
| binpacking_heterogeneity | Histogram | `instance_type`, `cpu_count`, `namespace_count` | Number of groups of equivalent pods being processed as a part of the same binpacking simulation. |
| max_node_skip_eval_duration_seconds | Gauge | | Maximum evaluation time of a node being skipped during ScaleDown. |
| node_removal_latency_seconds | Histogram | `deleted`=&lt;`true` or `false`&gt; | Latency from when an unneeded node becomes eligible for scale-down until it is actually removed (`deleted=true`) or becomes needed again (`deleted=false`). |
| dra_node_template_resources_mismatch | Gauge | `driver`=&lt;dra-driver-name&gt;, `mismatch_type`=&lt;`missing` or `extra` or `mismatch` or `unknown`&gt; | Count of resource pool mismatches between ready nodes and the node template that CA used to predict the node, useful for diagnosing scale-up correctness with Dynamic Resource Allocation (DRA). |

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
* `scaled_up_gpu_nodes_total` counts the number of GPU-enabled nodes
  successfully added by CA, similar to `scaled_up_nodes_total`. Additionally
  `gpu_name` specifies name of the GPU (e.g. nvidia-tesla-k80).
* `scaled_down_gpu_nodes_total` counts the number of nodes removed by CA. Scale
  down reasons are identical to `scaled_down_nodes_total`, `gpu_name` to
  `scaled_up_gpu_nodes_total`.
* `skipped_scale_events_count` counts the number of times that the
  autoscaler has declined to scale a node group because of a resource limit being reached or
  similar internal event. Scale direction can be either `up` or `down`, and the reason explains
  why the scaling was skipped (eg `CPULimitReached`, `MemoryLimitReached`). This is
  different than failed scaling events in that the autoscaler is choosing not to perform
  a scaling action.
