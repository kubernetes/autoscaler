<!--TODO: Remove "previously referred to as master" references from this doc once this terminology is fully removed from k8s-->
# Cluster Autoscaler

# Introduction

Cluster Autoscaler is a tool that automatically adjusts the size of the Kubernetes cluster when one of the following conditions is true:
* there are pods that failed to run in the cluster due to insufficient
  resources.
* there are nodes in the cluster that have been underutilized for an extended period of time and their pods can be placed on other existing nodes.

# FAQ/Documentation

An FAQ is available [HERE](./FAQ.md).

You should also take a look at the notes and "gotchas" for your specific cloud provider:
* [AliCloud](./cloudprovider/alicloud/README.md)
* [AWS](./cloudprovider/aws/README.md)
* [Azure](./cloudprovider/azure/README.md)
* [BaiduCloud](./cloudprovider/baiducloud/README.md)
* [BizflyCloud](./cloudprovider/bizflycloud/README.md)
* [Brightbox](./cloudprovider/brightbox/README.md)
* [CherryServers](./cloudprovider/cherryservers/README.md)
* [Civo](./cloudprovider/civo/README.md)
* [CloudStack](./cloudprovider/cloudstack/README.md)
* [ClusterAPI](./cloudprovider/clusterapi/README.md)
* [DigitalOcean](./cloudprovider/digitalocean/README.md)
* [Exoscale](./cloudprovider/exoscale/README.md)
* [Equinix Metal](cloudprovider/equinixmetal/README.md#notes)
* [External gRPC](./cloudprovider/externalgrpc/README.md)
* [Hetzner](./cloudprovider/hetzner/README.md)
* [HuaweiCloud](./cloudprovider/huaweicloud/README.md)
* [IonosCloud](./cloudprovider/ionoscloud/README.md)
* [Kamatera](./cloudprovider/kamatera/README.md)
* [Kwok](./cloudprovider/kwok/README.md)
* [Linode](./cloudprovider/linode/README.md)
* [Magnum](./cloudprovider/magnum/README.md)
* [OracleCloud](./cloudprovider/oci/README.md)
* [OVHcloud](./cloudprovider/ovhcloud/README.md)
* [Rancher](./cloudprovider/rancher/README.md)
* [Scaleway](./cloudprovider/scaleway/README.md)
* [TencentCloud](./cloudprovider/tencentcloud/README.md)
* [Vultr](./cloudprovider/vultr/README.md)

# Releases

We recommend using Cluster Autoscaler with the Kubernetes control plane (previously referred to as master) version for which it was meant. The below combinations have been tested on GCP. We don't do cross version testing or compatibility testing in other environments. Some user reports indicate successful use of a newer version of Cluster Autoscaler with older clusters, however, there is always a chance that it won't work as expected.

Starting from Kubernetes 1.12, versioning scheme was changed to match Kubernetes minor releases exactly.

| Kubernetes Version | CA Version               | Chart Version |
|--------------------|--------------------------|---------------|
| 1.30.x             | 1.30.x                   |9.37.0+|
| 1.29.X             | 1.29.X                   |9.35.0+|
| 1.28.X             | 1.28.X                   |9.34.0+|
| 1.27.X             | 1.27.X                   |9.29.0+|
| 1.26.X             | 1.26.X                   |9.28.0+|
| 1.25.X             | 1.25.X                   |               |
| 1.24.X             | 1.24.X                   |9.25.0+|
| 1.23.X             | 1.23.X                   |9.14.0+|
| 1.22.X             | 1.22.X                   |               |
| 1.21.X             | 1.21.X                   |9.10.0+|
| 1.20.X             | 1.20.X                   |9.5.0+|
| 1.19.X             | 1.19.X                   |               |
| 1.18.X             | 1.18.X                   |9.0.0+|
| 1.17.X             | 1.17.X                   |               |
| 1.16.X             | 1.16.X                   |               |
| 1.15.X             | 1.15.X                   |               |
| 1.14.X             | 1.14.X                   |               |
| 1.13.X             | 1.13.X                   |               |
| 1.12.X             | 1.12.X                   |               |
| 1.11.X             | 1.3.X                    |               |
| 1.10.X             | 1.2.X                    |               |
| 1.9.X              | 1.1.X                    |               |
| 1.8.X              | 1.0.X                    |               |
| 1.7.X              | 0.6.X                    |               |
| 1.6.X              | 0.5.X, 0.6.X<sup>*</sup> |               |
| 1.5.X              | 0.4.X                    |               |
| 1.4.X              | 0.3.X                    |               |

<sup>*</sup>Cluster Autoscaler 0.5.X is the official version shipped with k8s 1.6. We've done some basic tests using k8s 1.6 / CA 0.6 and we're not aware of any problems with this setup. However, Cluster Autoscaler internally simulates Kubernetes' scheduler and using different versions of scheduler code can lead to subtle issues.

## Schedule

Cluster Autoscaler releases new minor versions shortly after OSS Kubernetes release
and patches for versions corresponding to currently
supported [Kubernetes versions](https://kubernetes.io/releases/) on a roughly 2
month cadence. Currently planned schedule is below. Please note that target
dates listed below are approximate and we expect up to a week difference between
target ETA and the actual releases.

| Date       | Maintainer Preparing Release | Backup Maintainer | Type  |
|------------|------------------------------|-------------------|-------|
| 2024-07-18 | x13n                         | MaciekPytel       | patch |
| 2024-08-21 | MaciekPytel                  | gjtempleton       | 1.31  |
| 2024-09-18 | gjtempleton                  | towca             | patch |
| 2024-11-20 | towca                        | BigDarkClown      | patch |
| 2024-12-18 | BigDarkClown                 | x13n              | 1.32  |
| 2025-01-22 | x13n                         | MaciekPytel       | patch |

Additional patch releases may happen outside of the schedule in case of critical
bugs or vulnerabilities.

# Releases (Gardener/Autoscaler)

Starting with Gardener/Autoscaler v1.20, versioning scheme has changed to match Kubernetes and upstream Autoscaler minor releases exactly. Note that there is a jump of major version as well.

| Kubernetes Version | CA Version | Gardener CA Version | 
|--------------------|------------|---------------------|
| 1.31.X             | 1.31.X     | 1.31.X              |
| 1.30.X             | 1.30.X     | 1.30.X              |
| 1.29.X             | 1.29.X     | 1.29.X              |
| 1.28.X             | 1.28.X     | 1.28.X              |
| 1.27.X             | 1.27.X     | 1.27.X              |
| 1.26.X             | 1.26.X     | 1.26.X              |
| 1.25.X             | 1.25.X     | 1.25.X              |
| 1.24.X             | 1.24.X     | 1.24.X              |
| 1.23.X             | 1.23.X     | 1.23.X              |
| 1.22.X             | 1.22.X     | 1.22.X              |
| 1.21.X             | 1.21.X     | 1.21.X              |
| 1.20.X             | 1.20.X     | 1.20.X              |
| 1.20.X             | 1.20.X     | 0.19.X              |
| 1.19.X             | 1.19.X     | 0.18.X              |
| 1.19.X             | 1.19.X     | 0.17.X              |
| 1.19.X             | 1.19.X     | 0.16.X              |
| 1.19.X             | 1.19.X     | 0.15.X              |
| 1.19.X             | 1.19.X     | 0.14.X              |
| 1.19.X             | 1.19.X     | 0.13.X              |
| 1.19.X             | 1.19.X     | 0.12.X              |
| 1.19.X             | 1.19.X     | 0.11.X              |
| 1.12.X             | 1.12.X     | 0.10.X              |

# Notable changes

For CA 1.1.2 and later, please check [release
notes.](https://github.com/kubernetes/autoscaler/releases)

CA version 1.1.1:
* Fixes around metrics in the multiple kube apiserver configuration.
* Fixes for unready nodes issues when quota is overrun.

CA version 1.1.0:
* Added [Azure support](https://github.com/kubernetes/autoscaler/blob/master/cluster-autoscaler/cloudprovider/azure/README.md).
* Added support for pod priorities. More details [here](https://github.com/kubernetes/autoscaler/blob/master/cluster-autoscaler/FAQ.md#how-does-cluster-autoscaler-work-with-pod-priority-and-preemption).

CA version 1.0.3:
* Adds support for safe-to-evict annotation on pod. Pods with this annotation
  can be evicted even if they don't meet other requirements for it.
* Fixes an issue when too many nodes with GPUs could be added during scale-up
    (https://github.com/kubernetes/kubernetes/issues/54959).

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
* `cluster-autoscaler.kubernetes.io/scale-down-disabled` annotation for marking
  nodes that should not be scaled down.
* `scale-down-delay-after-delete` and `scale-down-delay-after-failure` flags
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

Cluster Autoscaler is designed to run on Kubernetes control plane (previously referred to as master) node. This is the
default deployment strategy on GCP.
It is possible to run a customized deployment of Cluster Autoscaler on worker nodes, but extra care needs
to be taken to ensure that Cluster Autoscaler remains up and running. Users can put it into kube-system
namespace (Cluster Autoscaler doesn't scale down node with non-mirrored kube-system pods running
on them) and set a `priorityClassName: system-cluster-critical` property on your pod spec
(to prevent your pod from being evicted).

Supported cloud providers:
* GCE https://kubernetes.io/docs/concepts/cluster-administration/cluster-management/
* GKE https://cloud.google.com/container-engine/docs/cluster-autoscaler
* AWS https://github.com/kubernetes/autoscaler/blob/master/cluster-autoscaler/cloudprovider/aws/README.md
* Azure https://github.com/kubernetes/autoscaler/blob/master/cluster-autoscaler/cloudprovider/azure/README.md
* AliCloud https://github.com/kubernetes/autoscaler/blob/master/cluster-autoscaler/cloudprovider/alicloud/README.md
* BaiduCloud https://github.com/kubernetes/autoscaler/blob/master/cluster-autoscaler/cloudprovider/baiducloud/README.md
* BizflyCloud https://github.com/kubernetes/autoscaler/blob/master/cluster-autoscaler/cloudprovider/bizflycloud/README.md
* Brightbox https://github.com/kubernetes/autoscaler/blob/master/cluster-autoscaler/cloudprovider/brightbox/README.md
* CherryServers https://github.com/kubernetes/autoscaler/blob/master/cluster-autoscaler/cloudprovider/cherryservers/README.md
* Civo https://github.com/kubernetes/autoscaler/blob/master/cluster-autoscaler/cloudprovider/civo/README.md
* CloudStack https://github.com/kubernetes/autoscaler/blob/master/cluster-autoscaler/cloudprovider/cloudstack/README.md
* ClusterAPI https://github.com/kubernetes/autoscaler/blob/master/cluster-autoscaler/cloudprovider/clusterapi/README.md
* DigitalOcean https://github.com/kubernetes/autoscaler/blob/master/cluster-autoscaler/cloudprovider/digitalocean/README.md
* Exoscale https://github.com/kubernetes/autoscaler/blob/master/cluster-autoscaler/cloudprovider/exoscale/README.md
* Equinix Metal https://github.com/kubernetes/autoscaler/blob/master/cluster-autoscaler/cloudprovider/equinixmetal/README.md
* External gRPC https://github.com/kubernetes/autoscaler/blob/master/cluster-autoscaler/cloudprovider/externalgrpc/README.md
* Hetzner https://github.com/kubernetes/autoscaler/blob/master/cluster-autoscaler/cloudprovider/hetzner/README.md
* HuaweiCloud https://github.com/kubernetes/autoscaler/blob/master/cluster-autoscaler/cloudprovider/huaweicloud/README.md
* IonosCloud https://github.com/kubernetes/autoscaler/blob/master/cluster-autoscaler/cloudprovider/ionoscloud/README.md
* Kamatera https://github.com/kubernetes/autoscaler/blob/master/cluster-autoscaler/cloudprovider/kamatera/README.md
* Linode https://github.com/kubernetes/autoscaler/blob/master/cluster-autoscaler/cloudprovider/linode/README.md
* Magnum https://github.com/kubernetes/autoscaler/blob/master/cluster-autoscaler/cloudprovider/magnum/README.md
* OracleCloud https://github.com/kubernetes/autoscaler/blob/master/cluster-autoscaler/cloudprovider/oci/README.md
* OVHcloud https://github.com/kubernetes/autoscaler/blob/master/cluster-autoscaler/cloudprovider/ovhcloud/README.md
* Rancher https://github.com/kubernetes/autoscaler/blob/master/cluster-autoscaler/cloudprovider/rancher/README.md
* Scaleway https://github.com/kubernetes/autoscaler/blob/master/cluster-autoscaler/cloudprovider/scaleway/README.md
* TencentCloud https://github.com/kubernetes/autoscaler/blob/master/cluster-autoscaler/cloudprovider/tencentcloud/README.md
* Vultr https://github.com/kubernetes/autoscaler/blob/master/cluster-autoscaler/cloudprovider/vultr/README.md
