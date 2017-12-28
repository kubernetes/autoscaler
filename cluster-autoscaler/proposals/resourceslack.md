# Resource Slack

## Abstract

"Resource slack" is some extra capacity in a k8s cluster kept for upcoming pods so that a newly created pod fit within resource slack does not need to wait until an extra node is started and gets ready.
It enables pods to be scaled out faster while minimizing impact to resource utilization.
This document sketches how the resource slack feature will be implemented in a future release of K8S.

## Motivation

Currently, ClusterAutoscaler tries to maintain necessary and sufficient amount of nodes.
For scaling up, it adds nodes only after some unschedulable pods are observed.
This behavior of scaling up implies that, in a worst case, an unschedulable pod must wait until an eligible node is created, started, and then gets ready before it is finally scheduled.
A node takes a much longer time than that of a pod to start and become ready and therefore it results in pods to occasionally delayed for several minutes to be horizontally scaled out.
A possible work-around can be done today is to lower the threshold of HPA to trigger a scaling up of a replicaset/deployment earlier but it reduces total resource utilization when there are many replicasets and deployments required the work-around.
Having the resource slack feature allows us to scale out pods faster while minimizing impact to resource utilization.

## Design

To achieve Resource Slack, introduce placeholder pods that will live only inside of ClusterAutoscaler.
They are created to virtually request certain amounts of CPU and memory so that they influence all of the scale-up/scale-down decision made in CA, but won't be present in the scheduler running inside an actual K8S cluster, so the new pods will be able to pick the place of the placeholders and push the placeholders elsewhere (and make them trigger the scale-up/down).

### Requirements

Version 1

* [R1] Works on relatively large clusters today's CA used to work
  * dozens of nodes, hundreds of pods
  * In general we don't want to have O(pods * nodes) operations every single loop
  * [R1-1] On large cluster there is a huge number of pods to be scheduled. And the pods are mostly the same. If a pod from one replica set X failed to schedule on node A then there is no point to in trying to fit the next pod from X on A. Thus there should be just a single loop over all nodes (ref https://github.com/kubernetes/autoscaler/pull/56#discussion_r115942312)
  * [R1-2] The scale down performance strongly depends that the pods don't change every iteration. So if you placed placeholder-X-123 on node A then in the next iteration it should also land on A or the scale down performance will be heavily impacted (ref https://github.com/kubernetes/autoscaler/pull/56#discussion_r115942312)
* [R2] Keep the specified amount of extra capacity at minimum
  * If we allowed users to annotate their k8s pods so that they are mirrored to equivalent placeholder pods like https://github.com/kubernetes/autoscaler/pull/77#discussion_r117084488,
    CA may actually produce much more extra capacity than specified by the user.

    Let's say an user specified CA to reserve 10% extra capacity, but also marked a pod which requests cpu and memory more than 10% of current capacity. CA will result in creating a placeholder pod which consumes resource more than 10%.
    So, the "10%" in this case is just a minimum requirement. CA may reserve more if the user instructed CA to do so by annotations. That's why I wrote `at minimum`. It is impossible to specify CA an exact rate of extra capacity when we want to suppoprt [R6] and [R7].
* [R3] Ensure full compatibility with scale down
* [R4] Only expose necessary and sufficient knobs to users

Version 2

contains [R1]-[R4] and also what is below:

* [R5] Don't add unnecessarily large node
  * A node is unnecessarily large when:
    * Any pod that is intended to be scheduled faster by resource slack fit within a smaller node
* [R6] Possible to provide extra space even for the "biggest" pod in term of memory and cpu
  * However, if the biggest pod in the cluster was a database, we won't want to keep a hot standby node for the pod to move to
* [R7] Possible to provide extra space even for every pod which is not biggest but do request some cpu(< max(pod_request_cpu)) and memory(< max(pod_request_mmory))
  * Say, the "biggest pods" request 1000m cpu/1G mem, 10m cpu/10G mem respectively
  * Even if they could fit within existing free space, a 500m(<1000m cpu)/5G mem pod has chances not to fit any node 
* [R8] Reserve space even for the pods having anti pod affinity
* [R9] Reserve space even for the pods having node selector/node affinity

Version 3

contains [R1]-[R9] and also what is below:

* [R10] Works on considerably large clusters
  * To move CA from beta to GA we need to handle 5000 nodes and 150 000 pods.
 
### Specifications

Version 1

* [S1] TBD [R1-1]
* [S2] Retain placeholder pods among iterations of CA and remove/re-assign/add only necessary placeholder pods. [R1-2]
  * More concretely, for each CA iteration:
  * Step 1: Calculate a `list of pushed-out placeholder pods`
    * Iterate over existing placeholder pods to determine which one is "pushed out of the node" by actual pod(s)
       * Iterate over ready nodes and see if `sum(pod requested cpu/mem)` exceeds a node's capacity. If it exceeds, push out placeholders until the sum becomes less than or equal to the node capacity.
    * Disassociate the pushed-out placeholder pods with the formerly associated nodes
  * Step 2: Calculate `desired extra capacity`
  * Step 3: Calculate `desired number of placeholder pods` from `desired extra capacity`
  * Step 4: Remove more placeholder pods if current extra capacity is larger than desired. Add more placeholder pods if desired extra capacity is increased.
  * Step 5: Do usual scale-up/down
  * Step 5: Retain the list of placeholder pods for the next iteration...
* [S3] Every place holder pod should be small enough to fit within at least one of nodes [R2]
  * Otherwise CA will produce no resource slack because it will be unable to add any node
* [S4] Create exactly one placeholder replicaset which looks like managing all the placeholder pods [R3]
* [S5] Provide `--extra-capacity-min-rate=<float>` to specify the rate of extra capacity compared to the total requested capacity [R4]

Version 2

contains [S1]-[S5] and also what is below:

* [S6] Every placeholder pod should have the same or smaller size compared to the pod with most requested resource [R5]
  * Otherwise CA will produce too much resource slack by adding an unnecessarily large node
* [S7] Pods/Deploypments/ReplicaSets/Jobs marked via specific annotations are tried to pre-reserve space for them [R6]-[R9]
  * Those pods are mirrored to equivalent placeholder pods, that can be bigger than standard placeholder pods, so that we can ensure the cluster to have extra space for them

Version 3

contains [S1]-[S7] and also what is below:

* TBD

## Algorithm

### Before the modification

The simplified flow inside the main CA loop looks like this:

* Get all nodes `AN`.
* Get all pods and put it into `AP`.
* Get unschedulable pods from `AP` and put `USP`.
* Try to find some place for `USP` on `AN`. Remove pods `USP` that you can place somewhere.
* Trigger scale up for the remaining `USP`, if needed.
* If no scale up occurred consider some nodes for scale down. See if their pods would be placed on some other nodes if their current node was removed.
* Remove a node, if possible.

### After the modification

Imagine that we add the extra placeholder pods `PP` to `AP` every loop. We also let them go to `USP`, but treat them in a slightly different way. If we can put `PP` on some node w actually put them there and consume the fee capacity for a moment. So the following steps will think that they are they are real, scheduled and running movable pods. If some part of `PP` cannot be scheduled it remains in `USP` among other "real" unschedulable pods. They can trigger scale up.

In the same way, during scale down considerations, cluster autoscaler treats PP as regular, replicated pods, and tries to find a place for them. If it cannot it means that scaling the node down would reduce the extra capacity below the desired level.

The flow would now look like this:

* Get all nodes `AN`.
* Get all pods and put it into `AP`.
* Get unschedulable pods from `AP` and put `USP`.
* **Create all placeholder pods `PP` and append to `USP`.**
* Try to find some place for `USP` on `AN`. Remove pods `USP` that you can place somewhere.
* Trigger scale up for the remaining `USP`, if needed.
* If no scale up occurred consider some nodes for scale down. See if their pods would be placed on some other nodes if their current node was removed.
* Remove a node, if possible.

### How placeholder pods are created

#### In version 1

Have `N` placeholder pods, each requiring `placeholder_cpu` amount of cpu and `placeholder_mem` amount of memory. The variables would be defined as:

Limitation: It does produce `min_extra_capacity_rate` more extra capacity than required but does not always produce space for a "biggest" pod 

```
granularity = 5
min_extra_capacity_rate = 0.1
cluster_cpu_capacity = sum(node_cpu_capacity)
cluster_mem_capacity = sum(node_mem_capacity)
extra_cpu = cluster_cpu_capacity * min_extra_capacity_rate
extra_mem = cluster_mem_capacity * min_extra_capacity_rate
placeholder_cpu = avg(node_cpu_capacity) / granularity
placeholder_mem = avg(node_mem_capacity) / granularity
placeholder_cnt = node_cnt * granularity
```

With that there will be an always present in-cluster-autoscaler need for

```
placeholder_cpu * placeholder_cnt
= avg(node_cpu_capacity) / granularity * node_cnt * granularity
= avg(node_cpu_capacity) * node_cnt
= extra_cpu
```

cpu and

```
placeholder_mem * placeholder_cnt
= avg(node_mem_capacity) / granularity * node_cnt * granularity
= avg(node_mem_capacity) * node_cnt
= extra_mem
```

#### In version 2 (Incomplete Idea)

Have `N` placeholder pods, each requiring `placeholder_cpu` amount of cpu and `placeholder_mem` amount of memory. The variables would be defined as:

Limitation: This does produce extra space like [S7] and [S8] but still doesn't address [R7]

```
granularity = 5
min_extra_capacity_rate = 0.1
cluster_cpu_capacity = sum(node_cpu_capacity)
cluster_mem_capacity = sum(node_mem_capacity)
extra_cpu = cluster_cpu_capacity * min_extra_capacity_rate
extra_mem = cluster_mem_capacity * min_extra_capacity_rate
biggest_cpu_pod_idx = idx where max(pod_requested_cpu(idx)) 
biggest_cpu_pod_cpu = pod_requested_cpu(biggest_cpu_pod_idx)
biggest_cpu_pod_mem = pod_requested_mem(biggest_cpu_pod_idx)
biggest_mem_pod_idx = idx where max(pod_requested_mem(idx))
biggest_mem_pod_cpu = pod_requested_cpu(biggest_mem_pod_idx)
biggest_mem_pod_mem = pod_requested_mem(biggest_mem_pod_idx)
remaining_cpu = extra_cpu - biggest_cpu_pod_cpu - biggest_mem_pod_cpu
remaining_mem = extra_mem - biggest_cpu_pod_mem - biggest_mem_pod_mem
placeholder_cpu = remaining_cpu / placeholder_cnt
placeholder_mem = remaining_mem / placeholder_cnt
placeholder_cnt = cluster_size * granularity
```

With that there will be an always present in-cluster-autoscaler need for

```
biggest_cpu_pod_cpu + biggest_mem_pod_cpu + placeholder_cpu * placeholder_cnt
= biggest_cpu_pod_cpu + biggest_mem_pod_cpu + remaining_cpu / placeholder_cnt * placeholder_cnt
= biggest_cpu_pod_cpu + biggest_mem_pod_cpu + remaining_cpu
= extra_cpu
```

cpu and

```
biggest_cpu_pod_mem + biggest_mem_pod_mem + placeholder_mem * placeholder_cnt
= biggest_cpu_pod_mem + biggest_mem_pod_mem + remaining_mem / placeholder_cnt * placeholder_cnt
= biggest_cpu_pod_mem + biggest_mem_pod_mem + remaining_mem
= extra_mem
```

#### In version 2 (More complete idea)

Have `N` placeholder pods, each requiring `placeholder_cpu` amount of cpu and `placeholder_mem` amount of memory. The variables would be defined as:

Limitation: This produces a minimum possible extra space([R5]) while producing enough space for any pod of any size([R6], [R7])

```
granularity = 5
min_extra_capacity_rate = 0.1
cluster_cpu_capacity = sum(node_cpu_capacity)
cluster_mem_capacity = sum(node_mem_capacity)
extra_cpu = cluster_cpu_capacity * min_extra_capacity_rate
extra_mem = cluster_mem_capacity * min_extra_capacity_rate
max_pod_cpu = max(pod_requested_cpu)
max_pod_mem = max(pod_requested_mem)
remaining_cpu = extra_cpu - max_pod_cpu
remaining_mem = extra_mem - max_pod_mem
placeholder_cpu = remaining_cpu / placeholder_cnt
placeholder_mem = remaining_mem / placeholder_cnt
placeholder_cnt = cluster_size * granularity
```

With that there will be an always present in-cluster-autoscaler need for

```
max_pod_cpu + placeholder_cpu * placeholder_cnt
= max_pod_cpu + remaining_cpu / placeholder_cnt * placeholder_cnt
= max_pod_cpu + remaining_cpu
= extra_cpu
```

cpu and

```
max_pod_mem + placeholder_mem * placeholder_cnt
= max_pod_mem + remaining_mem / placeholder_cnt * placeholder_cnt
= max_pod_mem + remaining_mem
= extra_mem
```

### How To Test

#### E2E

* Provide a document to reproduce the test scenario
  * Provision a k8s cluster in an convenient way in GKE, AWS(w/ kops or kube-aws or else, any feasible way)
  * Build a custom CA docker image containing the resource slack feature
  * `kubectl create -f ca.yaml` and an example of `ca.yaml` per cloud provider, maybe per k8s cluster provisioning tool
  * A command like `./generate-load` to generate some load on a cluster
  * Hold on for a while to watch the cluster scales down

## Next Steps

* Version 2: Add implementations of the specifications [S6], [S7], [S8]
* Version 3: TBD
