# Cluster Autoscaler - min at 0 
### Design Document for Google Cloud Platform
##### Author: mwielgus

# Introduction

One of the common requests for Cluster Autoscaler is the ability to scale some node groups to zero. This would definitely be a very useful feature but the implementation is somewhat problematic in ScaleUP due to couple reasons:

* [1] There is no live example of what a new node would look like if the currently zero-sized node group was expanded. The node shape is defined as:
  * [1A] Node capacity
  * [1B] Node allocatable
  *  [1C] Node labels
* [2] There is no live example of what manifest-run pods would be present on the new node.
* [3] There is no live example of what DaemonSets would be run on the new node.

In general the above can be summarized as that the full definition of a new node needs to be somehow known before the node is actually created in order to decide whether the creation of a new node from a particular node group makes sense or not. Scale down has no issues with min@0.

# Design

Problems 1, 1A, 1B, 1C, 2, 3 needs to be solved. The primary focus of this document is to create a solution for GCE/GKE but the proposed option should be generic enough to allow to expand this feature to other cloud providers.

Each node group in Cluster Autoscaler is on GCP backed by a Managed Instance Group. ManagedInstanceGroup has a node template and in the template there is couple of important information:

```yaml
creationTimestamp: '2017-02-24T06:16:21.761-08:00'
description: ''
id: '7272458143615811290'
kind: compute#instanceTemplate
name: kubernetes-minion-template
properties:
  canIpForward: true
  disks:
  - autoDelete: true
    boot: true
    initializeParams:
      diskSizeGb: '100'
      diskType: pd-standard
      sourceImage: https://www.googleapis.com/compute/v1/projects/google-containers/global/images/container-vm-v20170214
    kind: compute#attachedDisk
    mode: READ_WRITE
    type: PERSISTENT
  machineType: n1-standard-2
  metadata:
    fingerprint: -1QtDcJUvH0=
    items:
    - key: cluster-name
      value: |
        kubernetes
    - key: kube-env
      value:
        ENABLE_NODE_PROBLEM_DETECTOR: 'daemonset'
        ENABLE_RESCHEDULER: 'true'
        LOGGING_DESTINATION: 'gcp'
        ELASTICSEARCH_LOGGING_REPLICAS: '1'
        ENABLE_CLUSTER_DNS: 'true'
        ENABLE_CLUSTER_REGISTRY: 'false'
        CLUSTER_REGISTRY_DISK: 'kubernetes-kube-system-kube-registry'
        CLUSTER_REGISTRY_DISK_SIZE: '200GB'
        DNS_SERVER_IP: '10.0.0.10'
        DNS_DOMAIN: 'cluster.local'
        ENABLE_DNS_HORIZONTAL_AUTOSCALER: 'true'
        [...]
```

Node type and kube-env. These two fields are very useful to estimate the node shape.

### [1A] - Node capacity
To estimate the node capacity we need to decode n1-standard-2 or any other string mentioned in the machineType. The format of the strings is well documented
here: https://cloud.google.com/compute/docs/machine-types

However, except for the earliest prototype, we probably don’t want to manually hardcode all of the values available there. Luckily, the GCP platform provides an api to get the machine description in terms of memory and cpu:

```json
{
 "kind": "compute#machineTypeList",
 "id": "projects/mwielgus-proj/zones/us-central1-b/machineTypes",
 "items": [
  {
   "kind": "compute#machineType",
   "id": "2000",
   "creationTimestamp": "2015-01-16T09:25:43.315-08:00",
   "name": "g1-small",
   "description": "1 vCPU (shared physical core) and 1.7 GB RAM",
   "guestCpus": 1,
   "memoryMb": 1740,
   "maximumPersistentDisks": 16,
   "maximumPersistentDisksSizeGb": "3072",
   "zone": "us-central1-b",
   "selfLink": "https://www.googleapis.com/compute/v1/projects/mwielgus-proj/zones/us-central1-b/machineTypes/g1-small",
   "isSharedCpu": true
  },
  {
   "kind": "compute#machineType",
   "id": "4016",
   "creationTimestamp": "2015-01-16T09:25:43.316-08:00",
   "name": "n1-highcpu-16",
   "description": "16 vCPUs, 14.4 GB RAM",
   "guestCpus": 16,
   "memoryMb": 14746,
   "maximumPersistentDisks": 128,
   "maximumPersistentDisksSizeGb": "65536",
   "zone": "us-central1-b",
   "selfLink": "https://www.googleapis.com/compute/v1/projects/mwielgus-proj/zones/us-central1-b/machineTypes/n1-highcpu-16",
   "isSharedCpu": false
  },
```

So all standard type sizes can be obtained by API. On GCP there is a possibility of a custom machine type. This custom machine type has the name formatted as:

```
custom-<cpu_count>-<memory_in_mb>
```

So it is also quite easy to get all of the capacity information from it.

### [1B] - Node allocatable

In GKE 1.5.6 allocatable for new nodes is equal to capacity, however on GCE there is allocatable memory is a bit smaller than capacity. 
Initially, for simplicity, we can assume that the new node will have -0.1cpu/-200mb of capacity, but we will have to be more precise before the release.
More details of how the allocatables are calculated are available here: https://github.com/kubernetes/kubernetes/blob/c20e63bfb98fecef7461dbaf8ed52e31fe12cd11/pkg/kubelet/cm/node_container_manager.go#L184.
Being wrong or underestimating here is not fatal, most users will probably be OK with this. Once some nodes are present we will have more precise estimates. The worst thing that can happen is that the scale up may not be triggered if the request is exactly at the node capacity - system pods.

### [1C] - Node labels

The basic set of labels in GKE is relatively simple.

```
labels:
      beta.kubernetes.io/arch: amd64
      beta.kubernetes.io/instance-type: n1-standard-1
      beta.kubernetes.io/os: linux
      cloud.google.com/gke-nodepool: default-pool
      failure-domain.beta.kubernetes.io/region: us-central1
      failure-domain.beta.kubernetes.io/zone: us-central1-a
      kubernetes.io/hostname: gke-cluster-1-default-pool-408619fb-0zkd
      name: gke-cluster-1-default-pool-408619fb-0zkd
```

All of them can be “guessed” from the MIG description and template. Some of the labels are conditional like.

```
cloud.google.com/gke-local-ssd: "true"
cloud.google.com/gke-preemptible: "true"
```

There is also an option to specify custom labels for the node pool.
```
a: b
c: d
```
Luckily, all of these labels are mentioned in the NODE_LABELS variable in kube-env metadata of the template.
```
NODE_LABELS: a=b,c=d,cloud.google.com/gke-nodepool=pool-3,cloud.google.com/gke-preemptible=true

NODE_LABELS: cloud.google.com/gke-local-ssd=true,cloud.google.com/gke-nodepool=pool-1
```
The kubelet code that populates labels not available in kube_env is here: 
https://github.com/kubernetes/kubernetes/blob/ceff8d8d4d7ac271cd03dcae73edde048a685df5/pkg/kubelet/kubelet_node_status.go#L196

The bottom line is that all of the labels can be easily obtained.

### [2] - There is no live example of what manifest-run pods would be present on the new node.

In GKE (since 1.6) we run 1 types of pods by default on the node - Kube-proxy that requires only cpu. Unfortunately the amount of cpu is not well-defined and is hidden inside the startup script. https://github.com/kubernetes/kubernetes/blob/6bf9f2f0bbf25c550e9dd93bfa0a3cda4feec954/cluster/gce/gci/configure-helper.sh#L797

The amount is fixed at 100m and I guess it is unlikely to change so it can be probably hardcoded in CA as well.
	
### [3] - There is no live example of what DaemonSets would be run on the new node.

All daemon sets can be listed from apiserver and checked against the node with all the labels, capacities, allocatables and manifest-run pods obtained in previous steps. CA codebase already has the set of predicates imported so checking which pods should run on the node will be relatively easy.

# Solution

Given all the information above it should be relatively simple to write a module that given the access to GCP Api 
and Kubernetes API server. We will expand the NodeGroup interface (https://github.com/kubernetes/autoscaler/blob/master/cluster-autoscaler/cloudprovider/cloud_provider.go#L40)
with a method TemplateNodeInfo, taking no parameters and returning NodeInfo (containing api.Node and all pods running by default on the node) or error if unable to do so.

On GCP the method will get the MIG template and perform analysis needed to estimate the node shape.

The method will be called here: https://github.com/kubernetes/autoscaler/blob/master/cluster-autoscaler/core/utils.go#L132

only if the current size of the node group is 0 or all of the nodes are unready/broken. Otherwise CA will try to estimate the shape of the node using live examples to avoid repeating any mis-estimation errors.

The TemplateNodeInfo will also be run on CA startup to ensure that CA is able to build an example for the node pool if the node group min size was set to 0.
