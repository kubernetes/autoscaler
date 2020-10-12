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
  * [Which version on Cluster Autoscaler should I use in my cluster?](#which-version-on-cluster-autoscaler-should-i-use-in-my-cluster)
  * [Is Cluster Autoscaler an Alpha, Beta or GA product?](#is-cluster-autoscaler-an-alpha-beta-or-ga-product)
  * [What are the Service Level Objectives for Cluster Autoscaler?](#what-are-the-service-level-objectives-for-cluster-autoscaler)
  * [How does Horizontal Pod Autoscaler work with Cluster Autoscaler?](#how-does-horizontal-pod-autoscaler-work-with-cluster-autoscaler)
  * [What are the key best practices for running Cluster Autoscaler?](#what-are-the-key-best-practices-for-running-cluster-autoscaler)
  * [Should I use a CPU-usage-based node autoscaler with Kubernetes?](#should-i-use-a-cpu-usage-based-node-autoscaler-with-kubernetes)
  * [How is Cluster Autoscaler different from CPU-usage-based node autoscalers?](#how-is-cluster-autoscaler-different-from-cpu-usage-based-node-autoscalers)
  * [Is Cluster Autoscaler compatible with CPU-usage-based node autoscalers?](#is-cluster-autoscaler-compatible-with-cpu-usage-based-node-autoscalers)
  * [How does Cluster Autoscaler work with Pod Priority and Preemption?](#how-does-cluster-autoscaler-work-with-pod-priority-and-preemption)
  * [How does Cluster Autoscaler remove nodes?](#how-does-cluster-autoscaler-remove-nodes)
* [How to?](#how-to)
  * [I'm running cluster with nodes in multiple zones for HA purposes. Is that supported by Cluster Autoscaler?](#im-running-cluster-with-nodes-in-multiple-zones-for-ha-purposes-is-that-supported-by-cluster-autoscaler)
  * [How can I monitor Cluster Autoscaler?](#how-can-i-monitor-cluster-autoscaler)
  * [How can I scale my cluster to just 1 node?](#how-can-i-scale-my-cluster-to-just-1-node)
  * [How can I scale a node group to 0?](#how-can-i-scale-a-node-group-to-0)
  * [How can I prevent Cluster Autoscaler from scaling down a particular node?](#how-can-i-prevent-cluster-autoscaler-from-scaling-down-a-particular-node)
  * [How can I configure overprovisioning with Cluster Autoscaler?](#how-can-i-configure-overprovisioning-with-cluster-autoscaler)
* [Internals](#internals)
  * [Are all of the mentioned heuristics and timings final?](#are-all-of-the-mentioned-heuristics-and-timings-final)
  * [How does scale-up work?](#how-does-scale-up-work)
  * [How does scale-down work?](#how-does-scale-down-work)
  * [Does CA work with PodDisruptionBudget in scale-down?](#does-ca-work-with-poddisruptionbudget-in-scale-down)
  * [Does CA respect GracefulTermination in scale-down?](#does-ca-respect-gracefultermination-in-scale-down)
  * [How does CA deal with unready nodes?](#how-does-ca-deal-with-unready-nodes)
  * [How fast is Cluster Autoscaler?](#how-fast-is-cluster-autoscaler)
  * [How fast is HPA when combined with CA?](#how-fast-is-hpa-when-combined-with-ca)
  * [Where can I find the designs of the upcoming features?](#where-can-i-find-the-designs-of-the-upcoming-features)
  * [What are Expanders?](#what-are-expanders)
  * [Does CA respect node affinity when selecting node groups to scale up?](#does-ca-respect-node-affinity-when-selecting-node-groups-to-scale-up)
  * [What are the parameters to CA?](#what-are-the-parameters-to-ca)
* [Troubleshooting](#troubleshooting)
  * [I have a couple of nodes with low utilization, but they are not scaled down. Why?](#i-have-a-couple-of-nodes-with-low-utilization-but-they-are-not-scaled-down-why)
  * [How to set PDBs to enable CA to move kube-system pods?](#how-to-set-pdbs-to-enable-ca-to-move-kube-system-pods)
  * [I have a couple of pending pods, but there was no scale-up?](#i-have-a-couple-of-pending-pods-but-there-was-no-scale-up)
  * [CA doesn’t work, but it used to work yesterday. Why?](#ca-doesnt-work-but-it-used-to-work-yesterday-why)
  * [How can I check what is going on in CA ?](#how-can-i-check-what-is-going-on-in-ca-)
  * [What events are emitted by CA?](#what-events-are-emitted-by-ca)
  * [What happens in scale-up when I have no more quota in the cloud provider?](#what-happens-in-scale-up-when-i-have-no-more-quota-in-the-cloud-provider)
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
  * are not run on the node by default, *
  * don't have a [pod disruption budget](https://kubernetes.io/docs/concepts/workloads/pods/disruptions/#how-disruption-budgets-work) set or their PDB is too restrictive (since CA 0.6).
* Pods that are not backed by a controller object (so not created by deployment, replica set, job, stateful set etc). *
* Pods with local storage. *
* Pods that cannot be moved elsewhere due to various constraints (lack of resources, non-matching node selectors or affinity,
matching anti-affinity, etc)
* Pods that have the following annotation set:
```
"cluster-autoscaler.kubernetes.io/safe-to-evict": "false"
```

<sup>*</sup>Unless the pod has the following annotation (supported in CA 1.0.3 or later):
```
"cluster-autoscaler.kubernetes.io/safe-to-evict": "true"
```

__Or__ you have have overridden this behaviour with one of the relevant flags. [See below for more information on these flags.](#what-are-the-parameters-to-ca)

### Which version on Cluster Autoscaler should I use in my cluster?

See [Cluster Autoscaler Releases](https://github.com/kubernetes/autoscaler/tree/master/cluster-autoscaler#releases)

### Is Cluster Autoscaler an Alpha, Beta or GA product?

Since version 1.0.0 we consider CA as GA. It means that:

 * We have enough confidence that it does what it is expected to do. Each commit goes through a big suite of unit tests
   with more than 75% coverage (on average). We have a series of e2e tests that validate that CA works well on
   [GCE](https://k8s-testgrid.appspot.com/sig-autoscaling#gce-autoscaling)
   and [GKE](https://k8s-testgrid.appspot.com/sig-autoscaling#gke-autoscaling).
   Due to the missing testing infrastructure, AWS (or any other cloud provider) compatibility
   tests are not the part of the standard development or release procedure.
   However there is a number of AWS users who run CA in their production environment and submit new code, patches and bug reports.
 * It was tested that CA scales well. CA should handle up to 1000 nodes running 30 pods each. Our testing procedure is described
   [here](https://github.com/kubernetes/autoscaler/blob/master/cluster-autoscaler/proposals/scalability_tests.md).
 * Most of the pain-points reported by the users (like too short graceful termination support) were fixed, however
   some of the less critical feature requests are yet to be implemented.
 * CA has decent monitoring, logging and eventing.
 * CA tries to handle most of the error situations in the cluster (like cloud provider stockouts, broken nodes, etc). The cases handled can however vary from cloudprovider to cloudprovider.
 * CA developers are committed to maintaining and supporting CA in the foreseeable future.

All of the previous versions (earlier that 1.0.0) are considered beta.

### What are the Service Level Objectives for Cluster Autoscaler?

The main purpose of Cluster Autoscaler is to get pending pods a place to run.
Cluster Autoscaler periodically checks whether there are any pending pods and increases the size of the
cluster if it makes sense and if the scaled up cluster is still within the user-provided constraints.
The time of new node provisioning doesn't depend on CA,
but rather on the cloud provider and other Kubernetes components.

So, the main SLO for CA would be expressed in the latency time measured
from the time a pod is marked as unschedulable (by K8S scheduler) to the time
CA issues scale-up request to the cloud provider (assuming that happens).
During our scalability tests (described
[here](https://github.com/kubernetes/autoscaler/blob/master/cluster-autoscaler/proposals/scalability_tests.md))
we aimed at max 20sec latency, even in the big clusters.
We reach these goals on GCE on our test cases, however in practice, the
performance may differ. Hence, users should expect:

* No more than 30 sec latency on small clusters (less than 100 nodes with up to 30 pods each), with the average latency of about 5 sec.
* No more than 60 sec latency on big clusters (100 to 1000 nodes), with average latency of about 15 sec.

Please note that the above performance can be achieved only if NO pod affinity and anti-affinity is used on any of the pods.
Unfortunately, the current implementation of the affinity predicate in scheduler is about
3 orders of magnitude slower than for all other predicates combined,
and it makes CA hardly usable on big clusters.

It is also important to request full 1 core (or make it available) for CA pod in a bigger clusters.
Putting CA on an overloaded node would not allow to reach the declared performance.

We didn't run any performance tests on clusters bigger than 1000 nodes,
and supporting them was not a goal for 1.0.

More SLOs may be defined in the future.

### How does Horizontal Pod Autoscaler work with Cluster Autoscaler?

Horizontal Pod Autoscaler changes the deployment's or replicaset's number of replicas
based on the current CPU load.
If the load increases, HPA will create new replicas, for which there may or may not be enough
space in the cluster.
If there are not enough resources, CA will try to bring up some nodes, so that the
HPA-created pods have a place to run.
If the load decreases, HPA will stop some of the replicas. As a result, some nodes may become
underutilized or completely empty, and then CA will terminate such unneeded nodes.

### What are the key best practices for running Cluster Autoscaler?

* Do not modify the nodes belonging to autoscaled node groups directly. All nodes within the same node group should have the same capacity, labels and system pods running on them.
* Specify requests for your pods.
* Use PodDisruptionBudgets to prevent pods from being deleted too abruptly (if needed).
* Check if your cloud provider's quota is big enough before specifying min/max settings for your node pools.
* Do not run any additional node group autoscalers (especially those from your cloud provider).

### Should I use a CPU-usage-based node autoscaler with Kubernetes?

No.

### How is Cluster Autoscaler different from CPU-usage-based node autoscalers?

Cluster Autoscaler makes sure that all pods in the cluster have a place to run, no matter if
there is any CPU load or not.
Moreover, it tries to ensure that there are no unneeded nodes in the cluster.

CPU-usage-based (or any metric-based) cluster/node group autoscalers don't care about pods when
scaling up and down. As a result, they may add a node that will not have any pods,
or remove a node that has some system-critical pods on it, like kube-dns.
Usage of these autoscalers with Kubernetes is discouraged.

### Is Cluster Autoscaler compatible with CPU-usage-based node autoscalers?

No. CPU-based (or any metric-based) cluster/node group autoscalers, like
[GCE Instance Group Autoscaler](https://cloud.google.com/compute/docs/autoscaler/), are NOT compatible with CA.
They are also not particularly suited to use with Kubernetes in general.


### How does Cluster Autoscaler work with Pod Priority and Preemption?

Since version 1.1 (to be shipped with Kubernetes 1.9), CA takes pod priorities into account.

Pod Priority and Preemption feature enables scheduling pods based on priorities if there is not enough resources.
On the other hand, Cluster Autoscaler makes sure that there is enough resources to run all pods.
In order to allow users to schedule "best-effort" pods, which shouldn't trigger Cluster Autoscaler
actions, but only run when there are spare resources available, we introduced priority cutoff to
Cluster Autoscaler.

Pods with priority lower than this cutoff:
* don't trigger scale-ups - no new node is added in order to run them,
* don't prevent scale-downs - nodes running such pods can be terminated.

Nothing changes for pods with priority greater or equal to cutoff, and pods without priority.

Default priority cutoff is -10 (since version 1.12, was 0 before that).
It can be changed using `--expendable-pods-priority-cutoff` flag, but we discourage it.
Cluster Autoscaler also doesn't trigger scale-up if an unschedulable pod is already waiting for a lower
priority pod preemption.

Older versions of CA won't take priorities into account.

More about Pod Priority and Preemption:
 * [Priority in Kubernetes API](https://github.com/kubernetes/community/blob/master/contributors/design-proposals/scheduling/pod-priority-api.md),
 * [Pod Preemption in Kubernetes](https://github.com/kubernetes/community/blob/master/contributors/design-proposals/scheduling/pod-preemption.md),
 * [Pod Priority and Preemption tutorial](https://kubernetes.io/docs/concepts/configuration/pod-priority-preemption/).

### How does Cluster Autoscaler remove nodes?

Cluster Autoscaler terminates the underlying instance in a cloud-provider-dependent manner.

It does _not_ delete the [Node object](https://kubernetes.io/docs/concepts/architecture/nodes/#api-object) from Kubernetes. Cleaning up Node objects corresponding to terminated instances is the responsibility of the [cloud node controller](https://kubernetes.io/docs/concepts/architecture/cloud-controller/#node-controller), which can run as part of [kube-controller-manager](https://kubernetes.io/docs/reference/command-line-tools-reference/kube-controller-manager/) or [cloud-controller-manager](https://kubernetes.io/docs/reference/command-line-tools-reference/cloud-controller-manager/).


****************

# How to?

### I'm running cluster with nodes in multiple zones for HA purposes. Is that supported by Cluster Autoscaler?
CA 0.6 introduced `--balance-similar-node-groups` flag to support this use case. If you set the flag to true,
CA will automatically identify node groups with the same instance type and
the same set of labels (except for automatically added zone label) and try to
keep the sizes of those node groups balanced.

This does not guarantee similar node groups will have exactly the same sizes:
* Currently the balancing is only done at scale-up. Cluster Autoscaler will
  still scale down underutilized nodes regardless of the relative sizes of underlying
  node groups. We plan to take balancing into account in scale-down in the future.
* Cluster Autoscaler will only add as many nodes as required to run all existing
  pods. If the number of nodes is not divisible by the number of balanced node
  groups, some groups will get 1 more node than others.
* Cluster Autoscaler will only balance between node groups that can support the
  same set of pending pods. If you run pods that can only go to a single node group
  (for example due to nodeSelector on zone label) CA will only add nodes to
  this particular node group.

You can opt-out a node group from being automatically balanced with other node
groups using the same instance type by giving it any custom label.

### How can I monitor Cluster Autoscaler?
Cluster Autoscaler provides metrics and livenessProbe endpoints. By
default they're available on port 8085 (configurable with `--address` flag),
respectively under `/metrics` and `/health-check`.

Metrics are provided in Prometheus format and their detailed description is
available [here](https://github.com/kubernetes/autoscaler/blob/master/cluster-autoscaler/proposals/metrics.md).

### How can I scale my cluster to just 1 node?

Prior to version 0.6, Cluster Autoscaler was not touching nodes that were running important
kube-system pods like DNS, Metrics Server, Dashboard, etc. If these pods landed on different nodes,
CA could not scale the cluster down and the user could end up with a completely empty
3 node cluster. In 0.6, we added an option to tell CA that some system pods can be moved around.
If the user configures a [PodDisruptionBudget](https://kubernetes.io/docs/concepts/workloads/pods/disruptions/)
for the kube-system pod, then the default strategy of not touching the node running this pod
is overridden with PDB settings. So, to enable kube-system pods migration, one should set
[minAvailable](https://kubernetes.io/docs/api-reference/v1.7/#poddisruptionbudgetspec-v1beta1-policy)
to 0 (or <= N if there are N+1 pod replicas.)
See also [I have a couple of nodes with low utilization, but they are not scaled down. Why?](#i-have-a-couple-of-nodes-with-low-utilization-but-they-are-not-scaled-down-why)

### How can I scale a node group to 0?

From CA 0.6 for GCE/GKE and CA 0.6.1 for AWS, it is possible to scale a node group to 0 (and obviously from 0), assuming that all scale-down conditions are met.

For AWS, if you are using `nodeSelector`, you need to tag the ASG with a node-template key `"k8s.io/cluster-autoscaler/node-template/label/"`.

For example, for a node label of `foo=bar`, you would tag the ASG with:

```
{
    "ResourceType": "auto-scaling-group",
    "ResourceId": "foo.example.com",
    "PropagateAtLaunch": true,
    "Value": "bar",
    "Key": "k8s.io/cluster-autoscaler/node-template/label/foo"
}
```

### How can I prevent Cluster Autoscaler from scaling down a particular node?

From CA 1.0, node will be excluded from scale-down if it has the
annotation preventing scale-down:

```
"cluster-autoscaler.kubernetes.io/scale-down-disabled": "true"
```

It can be added to (or removed from) a node using kubectl:

```
kubectl annotate node <nodename> cluster-autoscaler.kubernetes.io/scale-down-disabled=true
```

### How can I configure overprovisioning with Cluster Autoscaler?

Below solution works since version 1.1 (to be shipped with Kubernetes 1.9).

Overprovisioning can be configured using deployment running pause pods with very low assigned
priority (see [Priority Preemption](https://kubernetes.io/docs/concepts/configuration/pod-priority-preemption/))
which keeps resources that can be used by other pods. If there is not enough resources then pause
pods are preempted and new pods take their place. Next pause pods become unschedulable and force CA
to scale up the cluster.

The size of overprovisioned resources can be controlled by changing the size of pause pods and the
number of replicas. This way you can configure static size of overprovisioning resources (i.e. 2
additional cores). If we want to configure dynamic size (i.e. 20% of resources in the cluster)
then we need to use [Horizontal Cluster Proportional Autoscaler](https://github.com/kubernetes-incubator/cluster-proportional-autoscaler)
which will change number of pause pods depending on the size of the cluster. It will increase the
number of replicas when cluster grows and decrease the number of replicas if cluster shrinks.

Configuration of dynamic overprovisioning:

1. (For 1.10, and below) Enable priority preemption in your cluster.

For GCE, it can be done by exporting following env
variables before executing kube-up (more details [here](https://kubernetes.io/docs/concepts/configuration/pod-priority-preemption/)):
```sh
export KUBE_RUNTIME_CONFIG=scheduling.k8s.io/v1alpha1=true
export ENABLE_POD_PRIORITY=true
```

For AWS using kops, see [this issue](https://github.com/kubernetes/autoscaler/issues/1410#issuecomment-439840945).

2. Define priority class for overprovisioning pods. Priority -1 will be reserved for
overprovisioning pods as it is the lowest priority that triggers scaling clusters. Other pods need
to use priority 0 or higher in order to be able to preempt overprovisioning pods. You can use
following definitions.

**For 1.10, and below:**

```yaml
apiVersion: scheduling.k8s.io/v1alpha1
kind: PriorityClass
metadata:
  name: overprovisioning
value: -1
globalDefault: false
description: "Priority class used by overprovisioning."
```

**For 1.11:**

```yaml
apiVersion: scheduling.k8s.io/v1beta1
kind: PriorityClass
metadata:
  name: overprovisioning
value: -1
globalDefault: false
description: "Priority class used by overprovisioning."
```

3. Change pod priority cutoff in CA to -10 so pause pods are taken into account during scale down
and scale up. Set flag ```expendable-pods-priority-cutoff``` to -10. If you already use priority
preemption then pods with priorities between -10 and -1 won't be best effort anymore.

4. Create service account that will be used by Horizontal Cluster Proportional Autoscaler which needs
specific roles. More details [here](https://github.com/kubernetes-incubator/cluster-proportional-autoscaler/tree/master/examples#rbac-configurations)

5. Create deployments that will reserve resources. "overprovisioning" deployment will reserve
resources and "overprovisioning-autoscaler" deployment will change the size of reserved resources.
You can use following definitions (you need to change service account for "overprovisioning-autoscaler"
deployment to the one created in the previous step):

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: overprovisioning
  namespace: default
spec:
  replicas: 1
  selector:
    matchLabels:
      run: overprovisioning
  template:
    metadata:
      labels:
        run: overprovisioning
    spec:
      priorityClassName: overprovisioning
      containers:
      - name: reserve-resources
        image: k8s.gcr.io/pause
        resources:
          requests:
            cpu: "200m"
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: overprovisioning-autoscaler
  namespace: default
  labels:
    app: overprovisioning-autoscaler
spec:
  selector:
    matchLabels:
      app: overprovisioning-autoscaler
  replicas: 1
  template:
    metadata:
      labels:
        app: overprovisioning-autoscaler
    spec:
      containers:
        - image: k8s.gcr.io/cluster-proportional-autoscaler-amd64:1.1.2
          name: autoscaler
          command:
            - ./cluster-proportional-autoscaler
            - --namespace=default
            - --configmap=overprovisioning-autoscaler
            - --default-params={"linear":{"coresPerReplica":1}}
            - --target=deployment/overprovisioning
            - --logtostderr=true
            - --v=2
      serviceAccountName: cluster-proportional-autoscaler-service-account
```

****************

# Internals

### Are all of the mentioned heuristics and timings final?

No. We reserve the right to update them in the future if needed.

### How does scale-up work?

Scale-up creates a watch on the API server looking for all pods. It checks for any unschedulable
pods every 10 seconds (configurable by `--scan-interval` flag). A pod is unschedulable when the Kubernetes scheduler is unable
to find a node that can accommodate the pod. For example, a pod can request more CPU that is
available on any of the cluster nodes. Unschedulable pods are recognized by their PodCondition.
Whenever a Kubernetes scheduler fails to find a place to run a pod, it sets "schedulable"
PodCondition to false and reason to "unschedulable".  If there are any items in the unschedulable
pods list, Cluster Autoscaler tries to find a new place to run them.

It is assumed that the underlying cluster is run on top of some kind of node groups.
Inside a node group, all machines have identical capacity and have the same set of assigned labels.
Thus, increasing a size of a node group will create a new machine that will be similar
to these already in the cluster - they will just not have any user-created pods running (but
will have all pods run from the node manifest and daemon sets.)

Based on the above assumption, Cluster Autoscaler creates template nodes for each of the
node groups and checks if any of the unschedulable pods would fit on a new node.
While it may sound similar to what the real scheduler does, it is currently quite simplified and
may require multiple iterations before all of the pods are eventually scheduled.
If there are multiple node groups that, if increased, would help with getting some pods running,
different strategies can be selected for choosing which node group is increased. Check [What are Expanders?](#what-are-expanders) section to learn more about strategies.

It may take some time before the created nodes appear in Kubernetes. It almost entirely
depends on the cloud provider and the speed of node provisioning. Cluster
Autoscaler expects requested nodes to appear within 15 minutes
(configured by `--max-node-provision-time` flag.) After this time, if they are
still unregistered, it stops considering them in simulations and may attempt to scale up a
different group if the pods are still pending. It will also attempt to remove
any nodes left unregistered after this time.

### How does scale-down work?

Every 10 seconds (configurable by `--scan-interval` flag), if no scale-up is
needed, Cluster Autoscaler checks which nodes are unneeded. A node is considered for removal when **all** below conditions hold:

* The sum of cpu and memory requests of all pods running on this node is smaller
  than 50% of the node's allocatable. (Before 1.1.0, node capacity was used
  instead of allocatable.) Utilization threshold can be configured using
  `--scale-down-utilization-threshold` flag.

* All pods running on the node (except these that run on all nodes by default, like manifest-run pods
or pods created by daemonsets) can be moved to other nodes. See
[What types of pods can prevent CA from removing a node?](#what-types-of-pods-can-prevent-ca-from-removing-a-node) section for more details on what pods don't fulfill this condition, even if there is space for them elsewhere.
While checking this condition, the new locations of all movable pods are memorized.
With that, Cluster Autoscaler knows where each pod can be moved, and which nodes
depend on which other nodes in terms of pod migration. Of course, it may happen that eventually
the scheduler will place the pods somewhere else.

* It doesn't have scale-down disabled annotation (see [How can I prevent Cluster Autoscaler from scaling down a particular node?](#how-can-i-prevent-cluster-autoscaler-from-scaling-down-a-particular-node))

If a node is unneeded for more than 10 minutes, it will be terminated. (This time can
be configured by flags - please see [I have a couple of nodes with low utilization, but they are not scaled down. Why?](#i-have-a-couple-of-nodes-with-low-utilization-but-they-are-not-scaled-down-why) section for a more detailed explanation.)
Cluster Autoscaler terminates one non-empty node at a time to reduce the risk of
creating new unschedulable pods. The next node may possibly be terminated just after the first one,
if it was also unneeded for more than 10 min and didn't rely on the same nodes
in simulation (see below example scenario), but not together.
Empty nodes, on the other hand, can be terminated in bulk, up to 10 nodes at a time (configurable by `--max-empty-bulk-delete` flag.)

What happens when a non-empty node is terminated? As mentioned above, all pods should be migrated
elsewhere. Cluster Autoscaler does this by evicting them and tainting the node, so they aren't
scheduled there again.

Example scenario:

Nodes A, B, C, X, Y.
A, B, C are below utilization threshold.
In simulation, pods from A fit on X, pods from B fit on X, and pods from C fit
on Y.

Node A was terminated. OK, but what about B and C, which were also eligible for deletion? Well, it depends.

Pods from B may no longer fit on X after pods from A were moved there. Cluster Autoscaler has to find place for them somewhere else, and it is not sure that if A had been terminated much earlier than B, there would always have been a place for them. So the condition of having been unneeded for 10 min may not be true for B anymore.

But for node C, it's still true as long as nothing happened to Y. So C can be terminated immediately after A, but B may not.

Cluster Autoscaler does all of this accounting based on the simulations and memorized new pod location.
They may not always be precise (pods can be scheduled elsewhere in the end), but it seems to be a good heuristic so far.

### Does CA work with PodDisruptionBudget in scale-down?

From 0.5 CA (K8S 1.6) respects PDBs. Before starting to terminate a node, CA makes sure that PodDisruptionBudgets for pods scheduled there allow for removing at least one replica. Then it deletes all pods from a node through the pod eviction API, retrying, if needed, for up to 2 min. During that time other CA activity is stopped. If one of the evictions fails, the node is saved and it is not terminated, but another attempt to terminate it may be conducted in the near future.

### Does CA respect GracefulTermination in scale-down?

CA, from version 1.0, gives pods at most 10 minutes graceful termination time by default (configurable via `--max-graceful-termination-sec`). If the pod is not stopped within these 10 min then the node is terminated anyway. Earlier versions of CA gave 1 minute or didn't respect graceful termination at all.

### How does CA deal with unready nodes?

From 0.5 CA (K8S 1.6) continues to work even if some nodes are unavailable.
The default number of tolerated unready nodes in CA 1.2.1 or earlier is 33% of total nodes in the cluster or up to 3 nodes, whichever is higher.
For CA 1.2.2 and later, it's 45% or 3 nodes.
This is configurable by `--max-total-unready-percentage` and `--ok-total-unready-count` flags.
Once there are more unready nodes in the cluster,
CA stops all operations until the situation improves. If there are fewer unready nodes,
but they are concentrated in a particular node group,
then this node group may be excluded from future scale-ups.

### How fast is Cluster Autoscaler?

By default, scale-up is considered up to 10 seconds after pod is marked as unschedulable, and scale-down 10 minutes after a node becomes unneeded.
There are multiple flags which can be used to configure these thresholds. For example, in some environments, you may wish to give the k8s scheduler
a bit more time to schedule a pod than the CA's scan-interval. One way to do this is by setting `--new-pod-scale-up-delay`, which causes the CA to
ignore unschedulable pods until they are a certain "age", regardless of the scan-interval. If k8s has not scheduled them by the end of that delay,
then they may be considered by the CA for a possible scale-up.

Assuming default settings, [SLOs described here apply](#what-are-the-service-level-objectives-for-cluster-autoscaler).

### How fast is HPA when combined with CA?

When HPA is combined with CA, the total time from increased load to new pods
running is determined by three major factors:

* HPA reaction time,

* CA reaction time,

* node provisioning time.

By default, pods' CPU usage is scraped by kubelet every 10 seconds, and it is obtained from kubelet
by Metrics Server every 1 minute. HPA checks CPU load metrics in Metrics Server every 30 seconds.
However, after changing the number of replicas, HPA backs off for 3 minutes before taking
further action. So it can be up to 3 minutes before pods are added or deleted,
but usually it's closer to 1 minute.

CA should react [as fast as described
here](#what-are-the-service-level-objectives-for-cluster-autoscaler), regardless
of whether it was HPA or the user that modified the number of replicas. For
scale-up, we expect it to be less than 30 seconds in most cases.

Node provisioning time depends mostly on cloud provider. In our experience, on GCE it usually takes 3
to 4 minutes from CA request to when pods can be scheduled on newly created nodes.

Total time is a sum of those steps, and it's usually about 5 minutes. Please note that CA is the
least significant factor here.

On the other hand, for scale-down CA is usually the most significant factor, as
it doesn't attempt to remove nodes immediately, but only after they've been
unneeded for a certain time.

### Where can I find the designs of the upcoming features?

CA team follows the generic Kubernetes process and submits design proposals [HERE](https://github.com/kubernetes/autoscaler/tree/master/cluster-autoscaler/proposals)
before starting any significant effort.
Some of the not-yet-fully-approved proposals may be hidden among [PRs](https://github.com/kubernetes/autoscaler/pulls).

### What are Expanders?

When Cluster Autoscaler identifies that it needs to scale up a cluster due to unschedulable pods,
it increases the number of nodes in some node group. When there is one node group, this strategy is trivial. When there is more than one node group, it has to decide which to expand.

Expanders provide different strategies for selecting the node group to which
new nodes will be added.

Expanders can be selected by passing the name to the `--expander` flag, i.e.
`./cluster-autoscaler --expander=random`.

Currently Cluster Autoscaler has 5 expanders:

* `random` - this is the default expander, and should be used when you don't have a particular
need for the node groups to scale differently.

* `most-pods` - selects the node group that would be able to schedule the most pods when scaling
up. This is useful when you are using nodeSelector to make sure certain pods land on certain nodes.
Note that this won't cause the autoscaler to select bigger nodes vs. smaller, as it can add multiple
smaller nodes at once.

* `least-waste` - selects the node group that will have the least idle CPU (if tied, unused memory)
after scale-up. This is useful when you have different classes of nodes, for example, high CPU or high memory nodes, and only want to expand those when there are pending pods that need a lot of those resources.

* `price` - select the node group that will cost the least and, at the same time, whose machines
would match the cluster size. This expander is described in more details
[HERE](https://github.com/kubernetes/autoscaler/blob/master/cluster-autoscaler/proposals/pricing.md). Currently it works only for GCE and GKE (patches welcome.)

* `priority` - selects the node group that has the highest priority assigned by the user. It's configuration is described in more details [here](expander/priority/readme.md)

### Does CA respect node affinity when selecting node groups to scale up?

CA respects `nodeSelector` and `requiredDuringSchedulingIgnoredDuringExecution` in nodeAffinity given that you have labelled your node groups accordingly. If there is a pod that cannot be scheduled with either `nodeSelector` or `requiredDuringSchedulingIgnoredDuringExecution` specified, CA will only consider node groups that satisfy those requirements for expansion.

However, CA does not consider "soft" constraints like `preferredDuringSchedulingIgnoredDuringExecution` when selecting node groups. That means that if CA has two or more node groups available for expansion, it will not use soft constraints to pick one node group over another.

****************

### What are the parameters to CA?

The following startup parameters are supported for cluster autoscaler:

| Parameter | Description | Default |
| --- | --- | --- |
| `cluster-name` | Autoscaled cluster name, if available | ""
| `address` | The address to expose prometheus metrics | :8085
| `kubernetes` | Kubernetes master location. Leave blank for default | ""
| `kubeconfig` | Path to kubeconfig file with authorization and master location information | ""
| `cloud-config` | The path to the cloud provider configuration file.  Empty string for no configuration file | ""
| `namespace` | Namespace in which cluster-autoscaler run | "kube-system"
| `scale-down-enabled` | Should CA scale down the cluster | true
| `scale-down-delay-after-add` | How long after scale up that scale down evaluation resumes | 10 minutes
| `scale-down-delay-after-delete` | How long after node deletion that scale down evaluation resumes, defaults to scan-interval | scan-interval
| `scale-down-delay-after-failure` | How long after scale down failure that scale down evaluation resumes | 3 minutes
| `scale-down-unneeded-time` | How long a node should be unneeded before it is eligible for scale down | 10 minutes
| `scale-down-unready-time` | How long an unready node should be unneeded before it is eligible for scale down | 20 minutes
| `scale-down-utilization-threshold` | Node utilization level, defined as sum of requested resources divided by capacity, below which a node can be considered for scale down | 0.5
| `scale-down-non-empty-candidates-count` | Maximum number of non empty nodes considered in one iteration as candidates for scale down with drain<br>Lower value means better CA responsiveness but possible slower scale down latency<br>Higher value can affect CA performance with big clusters (hundreds of nodes)<br>Set to non positive value to turn this heuristic off - CA will not limit the number of nodes it considers." | 30
| `scale-down-candidates-pool-ratio` | A ratio of nodes that are considered as additional non empty candidates for<br>scale down when some candidates from previous iteration are no longer valid<br>Lower value means better CA responsiveness but possible slower scale down latency<br>Higher value can affect CA performance with big clusters (hundreds of nodes)<br>Set to 1.0 to turn this heuristics off - CA will take all nodes as additional candidates.  | 0.1
| `scale-down-candidates-pool-min-count` | Minimum number of nodes that are considered as additional non empty candidates<br>for scale down when some candidates from previous iteration are no longer valid.<br>When calculating the pool size for additional candidates we take<br>`max(#nodes * scale-down-candidates-pool-ratio, scale-down-candidates-pool-min-count)` | 50
| `scan-interval` | How often cluster is reevaluated for scale up or down | 10 seconds
| `max-nodes-total` | Maximum number of nodes in all node groups. Cluster autoscaler will not grow the cluster beyond this number. | 0
| `cores-total` | Minimum and maximum number of cores in cluster, in the format <min>:<max>. Cluster autoscaler will not scale the cluster beyond these numbers. | 320000
| `memory-total` | Minimum and maximum number of gigabytes of memory in cluster, in the format <min>:<max>. Cluster autoscaler will not scale the cluster beyond these numbers. | 6400000
| `gpu-total` | Minimum and maximum number of different GPUs in cluster, in the format <gpu_type>:<min>:<max>. Cluster autoscaler will not scale the cluster beyond these numbers. Can be passed multiple times. CURRENTLY THIS FLAG ONLY WORKS ON GKE. | ""
| `cloud-provider` | Cloud provider type. | gce
| `max-empty-bulk-delete` | Maximum number of empty nodes that can be deleted at the same time.  | 10
| `max-graceful-termination-sec` | Maximum number of seconds CA waits for pod termination when trying to scale down a node.  | 600
| `max-total-unready-percentage` | Maximum percentage of unready nodes in the cluster.  After this is exceeded, CA halts operations | 45
| `ok-total-unready-count` | Number of allowed unready nodes, irrespective of max-total-unready-percentage  | 3
| `max-node-provision-time` | Maximum time CA waits for node to be provisioned | 15 minutes
| `nodes` | sets min,max size and other configuration data for a node group in a format accepted by cloud provider. Can be used multiple times. Format: <min>:<max>:<other...> | ""
| `node-group-auto-discovery` | One or more definition(s) of node group auto-discovery.<br>A definition is expressed `<name of discoverer>:[<key>[=<value>]]`<br>The `aws`, `gce`, and `azure` cloud providers are currently supported. AWS matches by ASG tags, e.g. `asg:tag=tagKey,anotherTagKey`<br>GCE matches by IG name prefix, and requires you to specify min and max nodes per IG, e.g. `mig:namePrefix=pfx,min=0,max=10`<br> Azure matches by tags on VMSS, e.g. `label:foo=bar`, and will auto-detect `min` and `max` tags on the VMSS to set scaling limits.<br>Can be used multiple times | ""
| `estimator` | Type of resource estimator to be used in scale up | binpacking
| `expander` | Type of node group expander to be used in scale up.  | random
| `write-status-configmap` | Should CA write status information to a configmap  | true
| `max-inactivity` | Maximum time from last recorded autoscaler activity before automatic restart | 10 minutes
| `max-failing-time` | Maximum time from last recorded successful autoscaler run before automatic restart | 15 minutes
| `balance-similar-node-groups` | Detect similar node groups and balance the number of nodes between them | false
| `balancing-ignore-label` | Define a node label that should be ignored when considering node group similarity. One label per flag occurrence. | ""
| `node-autoprovisioning-enabled` | Should CA autoprovision node groups when needed | false
| `max-autoprovisioned-node-group-count` | The maximum number of autoprovisioned groups in the cluster | 15
| `unremovable-node-recheck-timeout` | The timeout before we check again a node that couldn't be removed before | 5 minutes
| `expendable-pods-priority-cutoff` | Pods with priority below cutoff will be expendable. They can be killed without any consideration during scale down and they don't cause scale up. Pods with null priority (PodPriority disabled) are non expendable | -10
| `regional` | Cluster is regional | false
| `leader-elect` | Start a leader election client and gain leadership before executing the main loop.<br>Enable this when running replicated components for high availability | true
| `leader-elect-lease-duration` | The duration that non-leader candidates will wait after observing a leadership<br>renewal until attempting to acquire leadership of a led but unrenewed leader slot.<br>This is effectively the maximum duration that a leader can be stopped before it is replaced by another candidate.<br>This is only applicable if leader election is enabled | 15 seconds
| `leader-elect-renew-deadline` | The interval between attempts by the acting master to renew a leadership slot before it stops leading.<br>This must be less than or equal to the lease duration.<br>This is only applicable if leader election is enabled | 10 seconds
| `leader-elect-retry-period` | The duration the clients should wait between attempting acquisition and renewal of a leadership.<br>This is only applicable if leader election is enabled | 2 seconds
| `leader-elect-resource-lock` | The type of resource object that is used for locking during leader election.<br>Supported options are `endpoints` (default) and `configmaps` | "endpoints"
| `aws-use-static-instance-list` | Should CA fetch instance types in runtime or use a static list. AWS only | false
| `skip-nodes-with-system-pods` | If true cluster autoscaler will never delete nodes with pods from kube-system (except for DaemonSet or mirror pods) | true
| `skip-nodes-with-local-storage`| If true cluster autoscaler will never delete nodes with pods with local storage, e.g. EmptyDir or HostPath | true
| `min-replica-count` | Minimum number or replicas that a replica set or replication controller should have to allow their pods deletion in scale down | 0

# Troubleshooting:

### I have a couple of nodes with low utilization, but they are not scaled down. Why?

CA doesn't remove underutilized nodes if they are running pods [that it shouldn't evict](#what-types-of-pods-can-prevent-ca-from-removing-a-node). Other possible reasons for not scaling down:

* the node group already has the minimum size,

* node has the scale-down disabled annotation (see [How can I prevent Cluster Autoscaler from scaling down a particular node?](#how-can-i-prevent-cluster-autoscaler-from-scaling-down-a-particular-node))

* node was unneeded for less than 10 minutes (configurable by
  `--scale-down-unneeded-time` flag),

* there was a scale-up in the last 10 min (configurable by `--scale-down-delay-after-add` flag),

* there was a failed scale-down for this group in the last 3 minutes (configurable by `--scale-down-delay-after-failure` flag),

* there was a failed attempt to remove this particular node, in which case Cluster Autoscaler
  will wait for extra 5 minutes before considering it for removal again,

* using large custom value for `--scale-down-delay-after-delete` or `--scan-interval`, which delays CA action.

### How to set PDBs to enable CA to move kube-system pods?

By default, kube-system pods prevent CA from removing nodes on which they are running. Users can manually add PDBs for the kube-system pods that can be safely rescheduled elsewhere:

```
kubectl create poddisruptionbudget <pdb name> --namespace=kube-system --selector app=<app name> --max-unavailable 1
```

Here's how to do it for some common pods:

* kube-dns can safely be rescheduled as long as there are supposed to be at least 2 of these pods. In 1.7, this will always be
the case. For 1.6 and earlier, edit kube-dns-autoscaler config map as described
[here](https://kubernetes.io/docs/tasks/administer-cluster/dns-horizontal-autoscaling/#tuning-autoscaling-parameters),
adding preventSinglePointFailure parameter. For example:
```
linear:'{"coresPerReplica":256,"nodesPerReplica":16,"preventSinglePointFailure":true}'
```

* Metrics Server is best left alone, as restarting it causes the loss of metrics for >1 minute, as well as metrics
in dashboard from the last 15 minutes. Metrics Server downtime also means effective HPA downtime as it relies on metrics. Add PDB for it only if you're sure you don't mind.

### I have a couple of pending pods, but there was no scale-up?

CA doesn't add nodes to the cluster if it wouldn't make a pod schedulable.
It will only consider adding nodes to node groups for which it was configured.
So one of the reasons it doesn't scale up the cluster may be that the pod has too large
(e.g. 100 CPUs), or too specific requests (like node selector), and wouldn't fit on any of the
available node types.
Another possible reason is that all suitable node groups are already at their maximum size.

If the pending pods are in a [stateful set](https://kubernetes.io/docs/concepts/workloads/controllers/statefulset)
and the cluster spans multiple zones, CA may not be able to scale up the cluster,
even if it has not yet reached the upper scaling limit in all zones. Stateful
set pods require an associated Persistent Volume (PV), which is created
before scheduling the pod and CA has no way of influencing the zone choice. The
pending pod has a strict constraint to be scheduled in the same zone that the PV
is in, so if it is a zone that has already reached the upper scaling limit, CA
will not be able to perform a scale-up, even if there are other zones in which
nodes could be added. This will manifest itself by following events on the pod:

```
Events:
  Type     Reason             Age   From                Message
  ----     ------             ----  -------             -------
  Normal   NotTriggerScaleUp  ..    cluster-autoscaler  pod didn't trigger scale-up (it wouldn't fit if a new node is added)
  Warning  FailedScheduling   ..    default-scheduler   No nodes are available that match all of the following predicates:: Insufficient cpu (4), NoVolumeZoneConflict (2)
```

This limitation was solved with
[volume topological scheduling](https://github.com/kubernetes/community/blob/master/contributors/design-proposals/storage/volume-topology-scheduling.md)
introduced as beta in Kubernetes 1.11 and planned for GA in 1.13.
To allow CA to take advantage of topological scheduling, use separate node groups per zone.
This way CA knows exactly which node group will create nodes in the required zone rather than relying on the cloud provider choosing a zone for a new node in a multi-zone node group.
When using separate node groups per zone, the `--balance-similar-node-groups` flag will keep nodes balanced across zones for workloads that dont require topological scheduling.

### CA doesn’t work, but it used to work yesterday. Why?

Most likely it's due to a problem with the cluster. Steps to debug:

* Check if cluster autoscaler is up and running. In version 0.5 and later, it periodically publishes the kube-system/cluster-autoscaler-status config map. Check last update time annotation. It should be no more than 3 min (usually 10 sec old).

* Check in the above config map if cluster and node groups are in the healthy state. If not, check if there are unready nodes.

If both the cluster and CA appear healthy:

* If you expect some nodes to be terminated, but they are not terminated for a long
  time, check
  [I have a couple of nodes with low utilization, but they are not scaled down. Why?](#i-have-a-couple-of-nodes-with-low-utilization-but-they-are-not-scaled-down-why) section.

* If you expect some nodes to be added to make space for pending pods, but they are not added for a long time, check [I have a couple of pending pods, but there was no scale-up?](#i-have-a-couple-of-pending-pods-but-there-was-no-scale-up) section.

* If you have access to the master machine, check Cluster Autoscaler logs in `/var/log/cluster-autoscaler.log`. Cluster Autoscaler logs a lot of useful information, including why it considers a pod unremovable or what was its scale-up plan.

* Check events added by CA to the pod object.

* Check events on the kube-system/cluster-autoscaler-status config map.

* If you see failed attempts to add nodes, check if you have sufficient quota on your cloud provider side. If VMs are created, but nodes fail to register, it may be a symptom of networking issues.

### How can I check what is going on in CA ?

There are three options:

* Logs on the master node, in `/var/log/cluster-autoscaler.log`.
* Cluster Autoscaler 0.5 and later publishes kube-system/cluster-autoscaler-status config map.
  To see it, run `kubectl get configmap cluster-autoscaler-status -n kube-system
  -o yaml`.
* Events:
    * on pods (particularly those that cannot be scheduled, or on underutilized
      nodes),
    * on nodes,
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
      recorded on the node, describing status of scale-down operation.
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
### My cluster is below minimum / above maximum number of nodes, but CA did not fix that! Why?

Cluster Autoscaler will not scale the cluster beyond these limits, but does not enforce them. If your cluster is below the minimum number of nodes configured for Cluster Autoscaler, it will be scaled up *only* in presence of unschedulable pods.

### What happens in scale-up when I have no more quota in the cloud provider?

Cluster Autoscaler will periodically try to increase the cluster and, once failed,
move back to the previous size until the quota arrives or the scale-up-triggering pods are removed.

From version 0.6.2, Cluster Autoscaler backs off from scaling up a node group after failure.
Depending on how long scale-ups have been failing, it may wait up to 30 minutes before next attempt.

# Developer:

### How can I run e2e tests?

1. Set up environment and build e2e.go as described in the [Kubernetes docs](https://github.com/kubernetes/community/blob/master/contributors/devel/sig-testing/e2e-tests.md#building-and-running-the-tests).
2. Set up the following env variables:
    ```sh
    export KUBE_AUTOSCALER_MIN_NODES=3
    export KUBE_AUTOSCALER_MAX_NODES=6
    export KUBE_ENABLE_CLUSTER_AUTOSCALER=true
    export KUBE_AUTOSCALER_ENABLE_SCALE_DOWN=true
    ```
    This is the minimum number of nodes required for all e2e tests to pass. The tests should also pass if you set higher maximum nodes limit.
3. Run `go run hack/e2e.go -- --verbose-commands --up` to bring up your cluster.
4. SSH to the master node and edit `/etc/kubernetes/manifests/cluster-autoscaler.manifest` (you will need sudo for this).
    * If you want to test your custom changes set `image` to point at your own CA image.
    * Make sure `--scale-down-enabled` parameter in `command` is set to `true`.
5. Run CA tests with:
    ```sh
    go run hack/e2e.go -- --verbose-commands --test --test_args="--ginkgo.focus=\[Feature:ClusterSizeAutoscaling"
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
minimum, using Cluster-Autoscaler image containing your changes and using
configuration required to activate them:
  i. Create a deployment. Scale it up, so that some pods don't fit onto existing
  nodes. Wait for new nodes to be added by Cluster Autoscaler and confirm all
  pods have been scheduled successfully.
  ii. Scale the deployment down to a single replica and confirm that the
  cluster scales down.
3. Run a manual test following the basic use case of your change. Confirm that
   nodes are added or removed as expected. Once again, we ask you to use common
   sense to decide what needs to be tested.
4. Describe your testing in PR description or in a separate comment on your PR
   (example:
   https://github.com/kubernetes/autoscaler/pull/74#issuecomment-302434795).

We are aware that this process is tedious and we will work to improve it.

### How can I update CA dependencies (particularly k8s.io/kubernetes)?

Cluster Autoscaler imports a huge chunk of internal k8s code as it calls out to scheduler implementation.
Therefore we want to keep set of libraries used in CA as close to one used by k8s, to avoid
unexpected problems coming from version incompatibilities.

Cluster Autoscaler depends on `go modules` mechanism for dependency management, but do not use it directly
during build process. `go.mod` file is just used to generate the `vendor` directory and further compilation
is run against set of libraries stored in `vendor`. `vendor` directory can be regenerated using [`update-vendor.sh`](https://github.com/kubernetes/autoscaler/blob/master/cluster-autoscaler/hack/update-vendor.sh) script.
The `update-vendor.sh` script is responsible for autogenerating `go.mod` file used by Cluster Autoscaler. The base
of the file is `go.mod` file coming from [kubernetes/kubernetes](https://github.com/kubernetes/kubernetes) repository.
On top of that script adds modifications as defined
locally in [`go.mod-extra`](https://github.com/kubernetes/autoscaler/blob/master/cluster-autoscaler/go.mod-extra) file.

Note: It is important that one should **never manually edit** `go.mod` file as it is regenerated
on each `update-vendor.sh` call. Any extra libraries or version overrides should be put in `go.mod-extra` file (syntax of the file
is same as syntax of `go.mod` file).

Finally `vendor` directry is materialized and validation tests are run.

If everything completes correctly a commit with updated `vendor` directory is created automatically. The pull-request with changed vendor
must be sent out manually. The PR should include the auto-generated commit as well as commits containing any manual changes/fixes that need to
go together.

Execution of `update-vendor.sh` can be parametrized using command line argumets:
 - `-f` - kubernetes/kubernetes fork to use. On `master` it defaults to `git@github.com:kubernetes/kubernetes.git`
 - `-r` - revision in kubernetes/kubernetes which should be used to get base `go.mod` file
 - `-d` - specifies script workdir; useful to speed up execution if script needs to be run multiple times, because updating vendor resulted in some compilation errors on Cluster-Autoscaler side which need to be fixed
 - `-o` - overrides go version check, which may be useful if CA needs to use a different go version than the one in kubernetes go.mod file

Example execution looks like this:
```
./hack/update-vendor.sh -d/tmp/ca-update-vendor.ou1l -fgit@github.com:kubernetes/kubernetes.git -rmaster
```

Caveats:
 - `update-vendor.sh` is called directly in shell (no docker is used) therefore its operation may differ from environment to environment.
 - It is important that go version, which isn in use in the shell in which `update-vendor.sh` is called, matches the `go <version>` directive specified in `go.mod` file
   in `kubernetes/kubernetes` revision against which revendoring is done.
 - `update-vendor.sh` automatically runs unit tests as part of verification process. If one needs to suppress that, it can be done by overriding `VERIFY_COMMAND` variable (`VERIFY_COMMAND=true ./hack/update-vendor.sh ...`)
 - If one wants to only add new libraries to `go.mod-extra`, but not change the base `go.mod`, `-r` should be used with kubernetes/kubernets revision, which was used last time `update-vendor.sh` was called. One can determine that revision by looking at `git log` in Cluster Autoscaler repository. Following command will do the trick `git log | grep "Updating vendor against"`.


