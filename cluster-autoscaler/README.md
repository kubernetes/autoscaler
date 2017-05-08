# Cluster Autoscaler

# Introduction

Cluster Autoscaler is a tool that automatically adjusts the size of the Kubernetes cluster when:
* there are pods that failed to run in the cluster due to insufficient resources.
* some nodes in the cluster are so underutilized, for an extended period of time,
that they can be deleted and their pods will be easily placed on some other, existing nodes.

# Releases

We strongly recommend using Cluster Autoscaler with version for which it was meant. We don't
do ANY cross version testing so if you put the newest Cluster Autoscaler on an old cluster
there is a big chance that it won't work as expected.

| Kubernetes Version  | CA Version   |
|--------|--------|
| 1.6.X  | 0.5.X  |
| 1.5.X  | 0.4.X  |
| 1.4.X  | 0.3.X  |

# Notable changes:

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

# FAQ

Is available [HERE](./FAQ.md).


