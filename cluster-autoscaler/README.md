# Cluster Autoscaler

# Introduction

Cluster Autoscaler is a tool that automatically adjusts the size of the Kubernetes cluster when:
* there are pods that failed to run in the cluster due to insufficient resources.
* some nodes in the cluster are so underutilized, for an extended period of time,
that they can be deleted and their pods will be easily placed on some other, existing nodes.

# FAQ/Documentation

Is available [HERE](./FAQ.md).

# Releases

We strongly recommend using Cluster Autoscaler with version for which it was meant. We don't
do ANY cross version testing so if you put the newest Cluster Autoscaler on an old cluster
there is a big chance that it won't work as expected.

| Kubernetes Version  | CA Version   |
|--------|--------|
| 1.7.X  | 0.6.X  |
| 1.6.X  | 0.5.X, 0.6.X<sup>*</sup>  |
| 1.5.X  | 0.4.X  |
| 1.4.X  | 0.3.X  |

<sup>*</sup>Cluster Autoscaler 0.5.X is the official version shipped with k8s 1.6. We've done some basic tests using k8s 1.6 / CA 0.6 and we're not aware of any problems with this setup. However, CA internally simulates k8s scheduler and using different versions of scheduler code can lead to subtle issues.

# Notable changes

CA Version 1.0.2:
* Fixes issues with scaling node groups using GPU from 0 to 1 on GKE (https://github.com/kubernetes/autoscaler/pull/401) and AWS (https://github.com/kubernetes/autoscaler/issues/321).
* Fixes a bug where goroutines performing API calls were leaking when using dynamic config on AWS (https://github.com/kubernetes/autoscaler/issues/252).
* Node Autoprovisioning support for GKE (the implementation was included in 1.0.0, but this release includes some bugfixes and introduces metrics and events).

CA Version 1.0.1:
* Fixes a bug in handling nodes that, at the same time, fail to register in Kubernetes and can't be deleted from cloud provider (https://github.com/kubernetes/autoscaler/issues/369).
* Improves estimation of resources available on a node when performing scale-from-0 on GCE (https://github.com/kubernetes/autoscaler/issues/326).
* Bugfixes in the new GKE cloud provider implementation.

CA Version 1.0:

With this release we graduated Cluster Autoscaler to GA.

* Support for 1000 nodes running 30 pods each. See: [Scalability testing  report](https://github.com/kubernetes/autoscaler/blob/master/cluster-autoscaler/proposals/scalability_tests.md)
* Support for 10 min graceful termination.
* Improved eventing and monitoring.
* Node allocatable support.
* Removed Azure support. See: [PR removing support with reasoning behind this decision](https://github.com/kubernetes/autoscaler/pull/229)
* cluster-autoscaler.kubernetes.io/scale-down-disabled` annotation for marking
  nodes that should not be scaled down.
* scale-down-delay-after-delete` and `scale-down-delay-after-failure` flags
    replaced `scale-down-trial-interval`

CA Version 0.6:
* Allows scaling node groups to 0 (currently only in GCE/GKE, other cloud providers are coming). See: [How can I scale a node group to 0?](FAQ.md#how-can-i-scale-a-node-group-to-0)
* Price-based expander (currently only in GCE/GKE, other cloud providers are coming). See: [What are Expanders?](FAQ.md#what-are-expanders)
* Similar node groups are balanced (to be enabled with a flag). See: [I'm running cluster with nodes in multiple zones for HA purposes. Is that supported by Cluster Autoscaler?](FAQ.md#im-running-cluster-with-nodes-in-multiple-zones-for-ha-purposes-is-that-supported-by-cluster-autoscaler)
* It is possible to scale-down nodes with kube-system pods if PodDisruptionBudget is provided. See: [How can I scale my cluster to just 1 node?](FAQ.md#how-can-i-scale-my-cluster-to-just-1-node)
* Automatic node group discovery on AWS (to be enabled with a flag). See: [AWS doc](https://github.com/kubernetes/autoscaler/tree/master/cluster-autoscaler/cloudprovider/aws).
* CA exposes runtime metrics. See: [How can I monitor Cluster Autoscaler?](FAQ.md#how-can-i-monitor-cluster-autoscaler)
* CA exposes an endpoint for liveness probe.
* max-grateful-termination-sec flag renamed to max-graceful-termination-sec.
* Lower AWS API traffic to DescribeAutoscalingGroup.

CA Version 0.5.4:
* Fixes problems with node drain when pods are ignoring SIGTERM.

CA Version 0.5.3:
* Fixes problems with pod anti-affinity in scale up https://github.com/kubernetes/autoscaler/issues/33.

CA Version 0.5.2:
* Fixes problems with pods using persistent volume claims in scale up https://github.com/kubernetes/contrib/issues/2507.

CA Version 0.5.1:
* Fixes problems with slow network route creations on cluster scale up https://github.com/kubernetes/kubernetes/issues/43709.

CA Version 0.5:
* CA continues to operate even if some nodes are unready and is able to scale-down them.
* CA exports its status to kube-system/cluster-autoscaler-status config map.
* CA respects PodDisruptionBudgets.
* Azure support.
* Alpha support for dynamic config changes.
* Multiple expanders to decide which node group to scale up.

CA Version 0.4:
* Bulk empty node deletions.
* Better scale-up estimator based on binpacking.
* Improved logging.

CA Version 0.3:
* AWS support.
* Performance improvements around scale down.

# Deployment

Cluster Autoscaler runs on the Kubernetes master node (at least in the default setup on GCE and GKE).
It is possible to run customized Cluster Autoscaler inside of the cluster but then extra care needs
to be taken to ensure that Cluster Autoscaler is up and running. User can put it into kube-system
namespace (Cluster Autoscaler doesn't scale down node with non-manifest based kube-system pods running
on them) and mark with `scheduler.alpha.kubernetes.io/critical-pod` annotation (so that the rescheduler,
if enabled, will kill other pods to make space for it to run).

Right now it is possible to run Cluster Autoscaler on:
* GCE https://kubernetes.io/docs/concepts/cluster-administration/cluster-management/
* GKE https://cloud.google.com/container-engine/docs/cluster-autoscaler
* AWS https://github.com/kubernetes/autoscaler/blob/master/cluster-autoscaler/cloudprovider/aws/README.md
* Azure



