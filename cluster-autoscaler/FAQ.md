# Frequently Asked Questions

# Older versions
The answers in this FAQ apply to the newest (HEAD) version of Cluster Autoscaler. If
you're using an older version of CA please refer to corresponding version of
this document:

* [Cluster Autoscaler 0.5.X](https://github.com/kubernetes/autoscaler/blob/cluster-autoscaler-release-0.5/cluster-autoscaler/FAQ.md)
* [Cluster Autoscaler 0.6.X](https://github.com/kubernetes/autoscaler/blob/cluster-autoscaler-release-0.6/cluster-autoscaler/FAQ.md)

# Table of Contents:
<!--- TOC BEGIN -->
* [Basics](#basics)
  * [What is Cluster Autoscaler?](#what-is-cluster-autoscaler)
  * [When does Cluster Autoscaler change the size of a cluster?](#when-does-cluster-autoscaler-change-the-size-of-a-cluster)
  * [What types of pods can prevent CA from removing a node?](#what-types-of-pods-can-prevent-ca-from-removing-a-node)
  * [How does Horizontal Pod Autoscaler work with Cluster Autoscaler?](#how-does-horizontal-pod-autoscaler-work-with-cluster-autoscaler)
  * [What are the key best practices for running Cluster Autoscaler?](#what-are-the-key-best-practices-for-running-cluster-autoscaler)
  * [Should I use a CPU-usage-based node autoscaler with Kubernetes?](#should-i-use-a-cpu-usage-based-node-autoscaler-with-kubernetes)
  * [How is Cluster Autoscaler different from CPU-usage-based node autoscalers?](#how-is-cluster-autoscaler-different-from-cpu-usage-based-node-autoscalers)
  * [Is Cluster Autoscaler compatible with CPU-usage-based node autoscalers?](#is-cluster-autoscaler-compatible-with-cpu-usage-based-node-autoscalers)
* [How to?](#how-to)
  * [I'm running cluster with nodes in multiple zones for HA purposes. Is that supported by Cluster Autoscaler?](#im-running-cluster-with-nodes-in-multiple-zones-for-ha-purposes-is-that-supported-by-cluster-autoscaler)
  * [How can I monitor Cluster Autoscaler?](#how-can-i-monitor-cluster-autoscaler)
  * [How can I scale my cluster to just 1 node?](#how-can-i-scale-my-cluster-to-just-1-node)
  * [How can I scale a node group to 0?](#how-can-i-scale-a-node-group-to-0)
* [Internals](#internals)
  * [Are all of the mentioned heuristics and timings final?](#are-all-of-the-mentioned-heuristics-and-timings-final)
  * [How does scale up work?](#how-does-scale-up-work)
  * [How does scale down work?](#how-does-scale-down-work)
  * [Does CA work with PodDisruptionBudget in scale down?](#does-ca-work-with-poddisruptionbudget-in-scale-down)
  * [Does CA respect GracefulTermination in scale down?](#does-ca-respect-gracefultermination-in-scale-down)
  * [How does CA deal with unready nodes in version <= 0.4.0?](#how-does-ca-deal-with-unready-nodes-in-version--040)
  * [How does CA deal with unready nodes in version >=0.5.0 ?](#how-does-ca-deal-with-unready-nodes-in-version-050-)
  * [How fast is Cluster Autoscaler?](#how-fast-is-cluster-autoscaler)
  * [How fast is HPA when combined with CA?](#how-fast-is-hpa-when-combined-with-ca)
  * [Where can I find the designs of the upcoming features?](#where-can-i-find-the-designs-of-the-upcoming-features)
  * [What are Expanders?](#what-are-expanders)
  * [What Expanders are available?](#what-expanders-are-available)
* [Troubleshooting](#troubleshooting)
  * [I have a couple of nodes with low utilization, but they are not scaled down. Why?](#i-have-a-couple-of-nodes-with-low-utilization-but-they-are-not-scaled-down-why)
  * [I have a couple of pending pods, but there was no scale up?](#i-have-a-couple-of-pending-pods-but-there-was-no-scale-up)
  * [CA doesn’t work but it used to work yesterday. Why?](#ca-doesnt-work-but-it-used-to-work-yesterday-why)
  * [How can I check what is going on in CA ?](#how-can-i-check-what-is-going-on-in-ca-)
  * [What events are emitted by CA?](#what-events-are-emitted-by-ca)
  * [What happens in scale up when I have no more quota in the cloud provider?](#what-happens-in-scale-up-when-i-have-no-more-quota-in-the-cloud-provider)
* [Developer](#developer)
  * [How can I run e2e tests?](#how-can-i-run-e2e-tests)
  * [How should I test my code before submitting PR?](#how-should-i-test-my-code-before-submitting-pr)
  * [How can I update CA dependencies (particularly k8s.io/kubernetes)?](#how-can-i-update-ca-dependencies-particularly-k8siokubernetes)
<!--- TOC END -->

# Basics

### What is Cluster Autoscaler?

Cluster Autoscaler is a standalone program that adjusts the size of a Kubernetes cluster to meet the current needs.

### When does Cluster Autoscaler change the size of a cluster?

Cluster Autoscaler increases the size of the cluster when:
* there are pods that failed to schedule on any of the current nodes due to insufficient resources.
* adding a node similar to the nodes currently present in the cluster would help.

Cluster Autoscaler decreases the size of the cluster when some nodes are consistently unneeded for a significant amount of time. A node is unneeded when it has low utilization and all of its important pods can be moved elsewhere.

### What types of pods can prevent CA from removing a node?

* Pods with restrictive PodDisruptionBudget.
* Kube-system pods that:
  * are not run on the node by default,
  * don't have PDB or their PDB is too restrictive (since CA 0.6).
* Pods that are not backed by a controller object (so not created by deployment, replica set, job, stateful set etc).
* Pods with local storage.
* Pods that cannot be moved elsewhere due to various constraints (lack of resources, non-matching node selctors or affinity,
matching anti-affinity, etc)

### How does Horizontal Pod Autoscaler work with Cluster Autoscaler?

Horizontal Pod Autoscaler changes the deployment's or replicaset's number of replicas based on the current
CPU load.
If the load increases HPA will create new replicas for which there may or may not be enough
space in the cluster. If there are not enough resources then CA will try to bring up some nodes so that the
HPA-created pods have a place to run.
If the load decreases, HPA will stop some of the replicas. As a result, some nodes may start to be
underutilized or completely empty and then CA will delete such unneeded nodes.

### What are the key best practices for running Cluster Autoscaler?

* Do not modify the nodes. All nodes within the same node group should have the same capacity, labels and system pods running on them.
* Specify requests for your pods.
* Use PodDisruptionBudgets to prevent pods from being deleted (if needed).
* Check if your cloud provider's quota is big enough before specifying min/max settings for your node pools.
* Do not run any additional node group autoscalers (especially those from your cloud provider).

### Should I use a CPU-usage-based node autoscaler with Kubernetes?

No.

### How is Cluster Autoscaler different from CPU-usage-based node autoscalers?

Cluster Autoscaler makes sure that all of the pods in a cluster have a place to run, no matter if
there is any load in the cluster or not. Moreover it tries to ensure that there are no unneeded nodes
in the cluster.

CPU-usage-based (or any metric-based) cluster/node group autoscalers don't care about pods when scaling up
and down. As a result, they may add a node that will not have any pods, or remove a node that
has some system-critical pods on it, like kube-dns. Usage of these autoscalers with Kubernetes is discouraged.

### Is Cluster Autoscaler compatible with CPU-usage-based node autoscalers?

No. CPU-based (or any metric-based) cluster/node group autoscalers, like
[GCE Instance Group Autoscaler](https://cloud.google.com/compute/docs/autoscaler/), are NOT compatible with CA.
They are also not particularly suited to use with Kubernetes in general.

****************

# How to?

### I'm running cluster with nodes in multiple zones for HA purposes. Is that supported by Cluster Autoscaler?
CA 0.6 introduced `--balance-similar-node-groups` flag to support this use-case. If you set the flag to true
CA will automatically identify node groups using the same instance types and
having the same set of labels (except for automatically added zone labels) and try to
keep the size of those node groups balanced.

This does not guarantee similar node groups will have exactly the same sizes:
* Currently the balancing is only done at scale-up. Cluster Autoscaler will
  still scale-down underutilized nodes regardless of relative size of underlying
  node groups. We plan to take balancing into account in scale-down in the future.
* Cluster Autoscaler will only add as many nodes as required to run all existing
  pods. If the number of nodes is not divisible by number of balanced node
  groups some groups will get 1 more node than others.
* Cluster Autoscaler will only balance between node groups that can support the
  same set of pending pods. If you run pods that can only go to a single node group
  (for example due to nodeSelector on zone label) CA will only add nodes to
  this particular node group.

You can opt-out a node group from being automatically balanced with other node
groups using the same instance type by giving it any custom label.

### How can I monitor Cluster Autoscaler?
Cluster Autoscaler provides metrics and livenessProbe endpoints. By
default they're available on port 8085 (configurable with `--address` flag),
respectively under /metrics and /health-check.

Metrics are provided in Prometheus format and their detailed description is
available [here](https://github.com/kubernetes/autoscaler/blob/master/cluster-autoscaler/proposals/metrics.md).

### How can I scale my cluster to just 1 node?

Prior to version 0.6, Cluster Autoscaler was not touching nodes that were running important
kube-system pods like DNS, Heapster, Dashboard etc. If these pods landed on different nodes, 
CA could not scale the cluster down and the user could end up with a completely empty 
3 node cluser. In 0.6 we added an option to tell CA that some system pods can be moved around.
If a K8S user configure a [PodDisruptionBudget](https://kubernetes.io/docs/concepts/workloads/pods/disruptions/)
for the kube-system pod then the default strategy of not touching the node running this pod
is overwritten with PDB settings. So, to enable kube-system pods migration one should set
[minAvailable](https://kubernetes.io/docs/api-reference/v1.7/#poddisruptionbudgetspec-v1beta1-policy)
to 0 (or <= N if there are N+1 pod replicas).
See also [I have a couple of nodes with low utilization, but they are not scaled down. Why?](#i-have-a-couple-of-nodes-with-low-utilization-but-they-are-not-scaled-down-why)

### How can I scale a node group to 0?

From CA 0.6 for GCE/GKE and CA 0.6.1 for AWS - it is possible to scale a node group to 0 (and obviously from 0), assuming that all scale-down conditions are met.

For AWS if you are using `nodeSelector` you need to tag the ASG with a node-template key `"k8s.io/cluster-autoscaler/node-template/label/"`

For example for a node label of `foo=bar` you would tag the ASG with:

```
{
    "ResourceType": "auto-scaling-group",
    "ResourceId": "foo.example.com",
    "PropagateAtLaunch": true,
    "Value": "bar",
    "Key": "k8s.io/cluster-autoscaler/node-template/label/foo"
}
```
****************

# Internals

### Are all of the mentioned heuristics and timings final?

No. We reserve the right to update them in the future if needed.

### How does scale up work?

Scale up creates a watch on the api server looking for all pods. It checks for any unschedulable
pods every 10 seconds (configurable). A pod is unschedulable when the Kubernetes scheduler is unable
to find a node that can accommodate the pod. For example, a pod can request more CPU that is
available on any of the cluster nodes. Unschedulable pods are recognized by their PodCondition.
Whenever a kubernetes scheduler fails to find a place to run a pod it sets "schedulable"
PodCondition to false and reason to "unschedulable".  If there are any items on the unschedulable
lists Cluster Autoscaler tries to find a new place to run them.

It is assumed that the underlying cluster is run on top of some kind of node group.
Inside a node group all machines have identical capacity and have the same set of assigned labels.
Thus increasing a size of a node pool will bring in new machines that will be similar
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

* All system pods running on the node (except these that run on all nodes by default like
manifest-run pods or pods created by daemonsets) have a PodDisruptionBudget.

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
may still do it, because no one touched Y. So C can be deleted immediately after A, but B may not be
deleted immediately.

Cluster Autoscaler does all of this accounting based on the simulations and memorized new pod location.
They may not always be precise (pods can land elsewhere) but it seems to be a good heuristic so far.

### Does CA work with PodDisruptionBudget in scale down?

From 0.5 CA (K8S 1.6) respects PDB. Before starting to delete a node CA makes sure that there is at least some non-zero PodDisruptionBudget. Then it deletes all pods from a node through the pod eviction api, retrying, if needed, for up to 2 min. During that time other CA activities are stopped. If one of the evictions fails the node is saved and it is not deleted, but another attempt to delete it may be conducted in the near future.

### Does CA respect GracefulTermination in scale down?

CA gives pods at most 1 min graceful termination time. If the pod is not stopped within this 1 min the node is deleted anyway.

### How does CA deal with unready nodes in version <= 0.4.0?

A strict requirement for performing any scale operations is that the size of a node group,
measured on the cloud provider side, matches the number of nodes in Kubernetes that belong to this
node group. If this condition is not met then all scaling operations are postponed until it is
fulfilled.
Also, any scale down will happen only after at least 10 min have passed since the last scale up.

### How does CA deal with unready nodes in version >=0.5.0 ?

From 0.5 CA (K8S 1.6) continues the work even if some (up to 33% or not greater than 3, configurable via flag) percentage of nodes
is unavailable. Once there are more unready nodes in the cluster, CA pauses all operations until the situation
improves. If there are fewer unready nodes but they are concentrated in a particular node group
then this node group may be excluded from scale-ups.
Prior to 0.5, CA stopped all operations when a single node became unready.

### How fast is Cluster Autoscaler?

Scale up (if it is reasonable) is executed up to 10 seconds after some pod is marked as unschedulable.
Scale down is executed (by default) 10 min (or later) after a node becomes unneeded.

### How fast is HPA when combined with CA?

By default, Pod CPU usage is scraped by kubelets every 10 sec, and CPU usage is obtained from kubelets by Heapster every 1 min.
HPA checks cpu load metrics in Heapster every 30 sec, and CA looks for unschedulable pods every 10 sec. So the max reaction
time, measured from the time CPU spikes in the pods to the time CA asks the cloud provider for a new node is 2 min. On average
it should be around 1 min.
The amount of time the cloud provider needs to start a new node and boot it up is measured in minutes. On GCE/GKE it is around 1.5-2 min -
however this depends on the data center location and machine type.
Then it may take up to 30 sec to register the node in the Kubernetes master and finalize all of the necessary network settings.

All in all the total reaction time is around 4 min.

### Where can I find the designs of the upcoming features?

CA team follows the generic Kuberntes process and submits design proposals [HERE](https://github.com/kubernetes/autoscaler/tree/master/cluster-autoscaler/proposals)
before starting any bigger/significant effort.
Some of the not-yet-fully-approved proposals may be hidden among [PRs](https://github.com/kubernetes/autoscaler/pulls).

### What are Expanders?

When Cluster Autoscaler identifies that it needs to scale up a cluster due to unscheduable pods, 
it increases the nodes in a node group. When there is one Node Group, this strategy is trivial.

When there are more than one Node Group, which group should be grown or 'expanded'?

Expanders provide different strategies for selecting which Node Group to grow.

Expanders can be selected by passing the name to the `--expander` flag. i.e. 
`./cluster-autoscaler --expander=random`

### What Expanders are available?

Currently Cluster Autoscaler has 4 expanders:

* `random` - this is the default expander, and should be used when you don't have a particular
need for the node groups to scale differently.

* `most-pods` - selects the node group that would be able to schedule the most pods when scaling
up. This is useful when you are using nodeSelector to make sure certain pods land on certain nodes. 
Note that this won't cause the autoscaler to select bigger nodes vs. smaller, as it can grow multiple
smaller nodes at once.

* `least-waste` - selects the node group that will have the least idle CPU (and if tied, unused Memory) node group
when scaling up. This is useful when you have different classes of nodes, for example, high CPU or high Memory nodes,
and only want to expand those when pods that need those requirements are to be launched.

* `price` - select the node group that will cost the least and, in the same time, whose machines 
would match the cluster size. This expander is described in more details 
[HERE](https://github.com/kubernetes/autoscaler/blob/master/cluster-autoscaler/proposals/pricing.md). Currently
it works only for GCE and GKE.

************

# Troubleshooting:

### I have a couple of nodes with low utilization, but they are not scaled down. Why?

CA doesn't remove nodes if they are running system pods without a PodDisruptionBudget, pods without a controller or pods with 
local storage (see [What types of pods can prevent CA from removing a node?](#what-types-of-pods-can-prevent-ca-from-removing-a-node) )
Also it won't remove a node which has pods that cannot be run elsewhere due to limited resources. Another possibility
is that the corresponding node group already has the minimum size. Finally, CA doesn't scale down if there was a scale up
in the last 10 min.

If the reason your cluster isn't scaled down is due to system pods without a PodDisruptionBudget spread across multiple nodes,
you can manually add PDBs for the pods that can be safely rescheduled elsewhere:

```
kubectl create poddisruptionbudget <pdb name> --namespace=kube-system --selector app:<app name> --max-unavailable 1 
```

Here's how to do it for some common pods:

* kube-dns can safely be rescheduled as long as there are supposed to be at least 2 of these pods. In 1.7, this will always be 
the case. For 1.6 and earlier, edit kube-dns-autoscaler config map as described
[here](https://kubernetes.io/docs/tasks/administer-cluster/dns-horizontal-autoscaling/#tuning-autoscaling-parameters), 
adding preventSinglePointFailure parameter. For example:
```
linear:'{"coresPerReplica":256,"nodesPerReplica":16,"preventSinglePointFailure":true}'
```

* Heapster is best left alone, as restarting it causes the loss of metrics for >1 minute, as well as metrics
in dashboard from the last 15 minutes. Add PDB only if you're sure you don't mind it. App name is k8s-heapster.

### I have a couple of pending pods, but there was no scale up?

CA doesn't scale up the cluster when expansion of any of the node groups (for which it is configured) will not
make the pods schedulable. One of the possible reasons is that the pod has too big requests (ex. 100 cpus) or too specific
requests (like node selector) that cannot be fulfilled with the current nodes. The other reason is that all of the
relevant node groups are at their maximum size.

### CA doesn’t work but it used to work yesterday. Why?

Hopefully it is not a bug in Cluster Autoscaler, but most likely a problem with the cluster.

* Check If cluster autoscaler is up and running. In version 0.5 it periodically publishes the kube-system/cluster-autoscaler-status config map. Check last update time annotation. Should be no more than 3 min (usually 10 sec old).
* Check kube-system/cluster-autoscaler-status if the cluster and node groups are in the healthy state. If not, check the unready nodes.

* If you expect some nodes to be deleted but they are not deleted for a long time check:
    * if they contain pods that prevent the node from being deleted (see the corresponding question in the faq).
    * if the min/max boundaries you declared for a particular node group allow the scale up.
    * the content of /var/log/cluster-autoscaler.log.

* If you expect some nodes to be added to help some pending pods, but they are not added for a long time, check:
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

### What events are emitted by CA?

Whenever Cluster Autoscaler adds or removes nodes it will create events
describing this action. It will also create events for some serious
errors. Below is the non-exhaustive list of events emitted by CA (new events may
be added in future):

* on kube-system/cluster-autoscaler-status config map:
    * ScaledUpGroup - CA increased the size of node group, gives
      both old and new group size.
    * ScaleDownEmpty - CA removed a node with no pods running on it (except
      system pods found on all nodes).
    * ScaleDown - CA decided to remove a node with some pods running on it.
      Event includes names of all pods that will be rescheduled to drain the
      node.
* on nodes:
    * ScaleDown - CA is scaling down the node. Multiple ScaleDown events may be
      recorded on the node, describing status of scale down operation.
    * ScaleDownFailed - CA tried to remove the node, but failed. The event
      includes error message.
* on pods:
    * TriggeredScaleUp - CA decided to scale up cluster to make place for this
      pod.
    * NotTriggerScaleUp - CA couldn't find node group that can be scaled up to
      make this pod schedulable.
    * ScaleDown - CA will try to evict this pod as part of draining the node.

Example event:
```sh
$ kubectl describe pods memory-reservation-73rl0 --namespace e2e-tests-autoscaling-kncnx
Name:   memory-reservation-73rl0

...

Events:
  FirstSeen	LastSeen	Count	From			SubObjectPath	Type		Reason			Message
  ---------	--------	-----	----			-------------	--------	------			-------
  1m		1m		1	cluster-autoscaler			Normal		TriggeredScaleUp	pod triggered scale-up, group: https://content.googleapis.com/compute/v1/projects/maciekpytel-dev-playground/zones/us-central1-b/instanceGroups/e2e-test-maciekpytel-minion-group, sizes (current/new): 3/4
```

### What happens in scale up when I have no more quota in the cloud provider?

Scale up will periodically try to increase the cluster and, once failed, move back to the previous size until the quota arrives or
the scale-up-triggering pods are removed.

# Developer:

### How can I run e2e tests?

1. Set up environment and build e2e.go as described in the [Kubernetes docs](https://github.com/kubernetes/community/blob/master/contributors/devel/e2e-tests.md#building-and-running-the-tests).
2. Set up the following env variables:
    ```sh
    export KUBE_AUTOSCALER_MIN_NODES=3
    export KUBE_AUTOSCALER_MAX_NODES=6
    export KUBE_ENABLE_CLUSTER_AUTOSCALER=true
    ```
    This is the minimum number of nodes required for all e2e tests to pass. The tests should also pass if you set higher quota.
3. Run `go run hack/e2e.go -- -v --up` to bring up your cluster.
4. SSH to the master node and edit `/etc/kubernetes/manifests/cluster-autoscaler.manifest` (you will need sudo for this).
    * If you want to test your custom changes set `image` to point at your own CA image.
    * Make sure `--scale-down-enabled` parameter in `command` is set to `true`.
5. Run CA tests with:
    ```sh
    go run hack/e2e.go -- -v --test --test_args="--ginkgo.focus=\[Feature:ClusterSizeAutoscaling"
    ```
    It will take >1 hour to run the full suite. You may want to redirect output to file, as there will be plenty of it.

    Test runner may be missing default credentials. On GCE they can be provided with:
    ```sh
    gcloud beta auth application-default login
    ```

A few tests are specific to GKE and will be skipped if you're running on a
different provider.

Please open an issue if you find a failing or flaky test (a PR will be even more welcome).

### How should I test my code before submitting PR?
This answer only applies to pull requests containing non-trivial code changes.

Unfortunately we can't automatically run e2e tests on every pull request yet, so
for now we need to follow a few manual steps to test that PR doesn't break
basic Cluster Autoscaler functionality. We don't require you to follow this
whole process for trivial bugfixes or minor changes that don't affect main loop. Just
use common sense to decide what is and what isn't required for your change.

To test your PR:
1. Run Cluster Autoscaler e2e tests if you can. We are running our e2e tests on GCE and we
   can't guarantee the tests are passing on every cloud provider.
2. If you can't run e2e we ask you to do a following manual test at the
minimum, using Cluster-Autoscaler build with your changes and using config
required to activate them:
  i. Create a deployment. Scale it up, so that some pods don't fit onto existing
  nodes. Wait for new nodes to be added by Cluster-Autoscaler and confirm all
  pods have scheduled successfully.
  ii. Scale the deployment back down to a single replica and confirm that the
  cluster scales down.
3. Run a manual test following the basic use-case of your change. Confirm that
   nodes are added or removed as expected. Once again we ask you to use common
   sense to decide what needs to be tested.
4. Describe your testing in PR description or in a separate comment on your PR
   (example:
   https://github.com/kubernetes/autoscaler/pull/74#issuecomment-302434795).

We are aware that this process is tedious and we will work to improve it.

### How can I update CA dependencies (particularly k8s.io/kubernetes)?

CA depends on `k8s.io/kubernetes` internals as well as the "official" k8s.io libs like 
`k8s.io/apimachinery`. However `k8s.io/kubernetes` has its own/newer version of these libraries
(in a `staging` directory) which may not always be compatibile with what has been published.
This leads to various conflicts that are hard to resolve in a "proper" way. So until a better solution 
is proposed (or we stop migrating stuff between `k8s.io/kubernets` and other projects on a daily basis),
the following hack has to be used to make the things easier to handle.

1. Create a new `$GOPATH` directory.
2. Get `k8s.io/kubernetes` and `k8s.io/autoscaler` source code (via `git clone` or `go get`).
3. Make sure that you use the correct branch/tag in `k8s.io/kubernetes`. For example, regular dev updates
   should be done against `k8s.io/kubernetes` HEAD, while updates in CA release branches should be done 
   against the latest release tag of the corresponding `k8s.io/kubernetes` branch.
4. Do `godep restore` in `k8s.io/kubernetes`.
5. Remove Godeps and vendor from `k8s.io/autoscaler/cluster-autoscaler`.
6. Invoke `fix-gopath.sh`. This will update `k8s.io/api`, `k8s.io/apimachinery` etc with the content of 
   `k8s.io/kubernetes/staging` and remove all vendor directories from your gopath.
7. Add some other dependencies, if needed and make sure that the code in `k8s.io/autoscaler/cluster-autoscaler`
   refers to them somehow (may be a blank import).
8. Check if everything compiles with `go test ./...` in `k8s.io/autoscaler/cluster-autoscaler`.
9. `godep save ./...` in `k8s.io/autoscaler/cluster-autoscaler`,
10. Send a PR with 2 commits - one that covers `Godep` and `vendor/` and the other one with all 
   required real code changes.
