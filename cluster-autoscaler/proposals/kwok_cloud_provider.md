# KWOK (Kubernetes without Kubelet) cloud provider

## Introduction
> [KWOK](https://sigs.k8s.io/kwok) is a toolkit that enables setting up a cluster of thousands of Nodes in seconds. Under the scene, all Nodes are simulated to behave like real ones, so the overall approach employs a pretty low resource footprint that you can easily play around on your laptop.

https://kwok.sigs.k8s.io/

## Problem
### 1. It is hard to reproduce an issue happening at scale on local machine 
e.g., https://github.com/kubernetes/autoscaler/issues/5769

To reproduce such issues, we have the following options today:  
### (a) setup [Kubemark](https://github.com/kubernetes/design-proposals-archive/blob/main/scalability/kubemark.md) on a public cloud provider and try reproducing the issue  
You can [setup Kubemark](https://github.com/kubernetes/community/blob/master/contributors/devel/sig-scalability/kubemark-guide.md) ([related](https://github.com/kubernetes/kubernetes/blob/master/test/kubemark/pre-existing/README.md))  and use the [`kubemark` cloudprovider](https://github.com/kubernetes/autoscaler/tree/master/cluster-autoscaler/cloudprovider/kubemark) (kubemark [proposal](https://github.com/kubernetes/autoscaler/blob/master/cluster-autoscaler/proposals/kubemark_integration.md)) directly or [`cluster-api` cloudprovider with kubemark](https://github.com/kubernetes-sigs/cluster-api-provider-kubemark)

In either case,

> Every running Kubemark setup looks like the following:
> 1) A running Kubernetes cluster pointed to by the local kubeconfig
> 2) A separate VM where the kubemark master is running
> 3) Some hollow-nodes that run on the Kubernetes Cluster from #1
> 4) The hollow-nodes are configured to talk with the kubemark master at #2

https://github.com/kubernetes/kubernetes/blob/master/test/kubemark/pre-existing/README.md#introduction

You need to setup a separate VM (Virtual Machine) with master components to get Kubemark running. 

> Currently we're running HollowNode with a limit of 0.09 CPU core/pod and 220MB of memory. However, if we also take into account the resources absorbed by default cluster addons and fluentD running on the 'external' cluster, this limit becomes ~0.1 CPU core/pod, thus allowing ~10 HollowNodes to run per core (on an "n1-standard-8" VM node).

https://github.com/kubernetes/community/blob/master/contributors/devel/sig-scalability/kubemark-guide.md#starting-a-kubemark-cluster

Kubemark can mimic 10 nodes with 1 CPU core. 

In reality it might be lesser than 10 nodes,
> Using Kubernetes and [kubemark](https://github.com/kubernetes/community/blob/master/contributors/design-proposals/scalability/kubemark.md) on GCP we have created a following 1000 node cluster setup:
>* 1 master - 1-core VM
>* 17 nodes - 8-core VMs, each core running up to 8 Kubemark nodes.
>* 1 Kubemark master - 32-core VM
>* 1 dedicated VM for Cluster Autoscaler 

https://github.com/kubernetes/autoscaler/blob/master/cluster-autoscaler/proposals/scalability_tests.md#test-setup

This is a cheaper option than (c) but if you want to setup Kubemark on your local machine you will need a master node and 1 core per 10 fake nodes i.e., if you want to mimic 100 nodes, that's 10 cores of CPU + extra CPU for master node. Unless you have 10-12 free cores on your local machine, it is hard to run scale tests with Kubemark for nodes > 100.  

### (b) try to get as much information from the issue reporter as possible and try to reproduce the issue by tweaking our code tests  
This works well if the issue is easy to reproduce by tweaking tests e.g., you want to check why scale down is getting blocked on a particular pod. You can do so by mimicing the pod in the tests by adding an entry [here](https://github.com/kubernetes/autoscaler/blob/1009797f5585d7bf778072ba59fd12eb2b8ab83c/cluster-autoscaler/utils/drain/drain_test.go#L878-L887) and running
```
cluster-autoscaler/utils/drain$ go test -run TestDrain
```
But when you want to test an issue related to scale e.g., CA is slow in scaling up, it is hard to do. 
### (c) try reproducing the issue using the same CA setup as user with actual nodes in a public cloud provider 
e.g., if the issue reporter has a 200 node cluster in AWS, try creating a 200 node cluster in AWS and use the same CA flags as the issue reporter.

This is a viable option if you already have a cluster running with a similar size but otherwise creating a big cluster just to reproduce the issue is costly. 

### 2. It is hard to confirm behavior of CA at scale
For example, a user with a big Kubernetes cluster (> 100-200 nodes) wants to check if adding scheduling properties to their workloads (node affinity, pod affinity, node selectors etc.,) leads to better utilization of the nodes (which saves cost). To give a more concrete example, imagine a situation like this:
1. There is a cluster with > 100 nodes. cpu to memory ratio for the nodes is 1:1, 1:2, 1:8 and 1:16
2. It is observed that 1:16 nodes are underutilized on memory
3. It is observed that workloads with cpu to memory ratio of 1:7 are getting scheduled on 1:16 nodes thereby leaving some memory unused
e.g., 
1:16 node looks like this:
CPUs: 8 Cores
Memory: 128Gi

workload (1:7 memory:cpu ratio):
CPUs: 1 Core
Memory: 7 Gi

resources wasted on the node: 8 % 1 CPU(s) + 128 % 7 Gi
= 0 CPUs + 2 Gi memory = 2Gi of wasted memory

1:8 node looks like this:
CPUs: 8 Cores
Memory: 64 Gi

workload (1:7 memory:cpu ratio):
CPUs: 1 Core
Memory: 7 Gi

resources wasted on the node: 8 % 1 CPU(s) + 64 % 7 Gi
= 0 CPUs + 1 Gi memory = 1Gi of wasted memory

If 1:7 can somehow be scheduled on 1:8 node using node selector or required node affinity, the wastage would go down. User wants to add required node affinity on 1:7 workloads and see how CA would behave without creating actual nodes in public cloud provider. The goal here is to see if the theory is true and if there are any side-effects.

This can be done with Kubemark today but a public cloud provider would be needed to mimic the cluster of this size. It can't be done on a local cluster (kind/minikube etc.,). 

## Proposal/Solution

Implement in-tree cloud provider for [Kubernetes WithOut Kubelet](https://kwok.sigs.k8s.io/) (referred as `kwok` provider henceforth) by implementing the [`CloudProvider`](https://github.com/vadasambar/autoscaler/blob/0d167b4e0586b04e773079d81cb9c45bca9ce0d3/cluster-autoscaler/cloudprovider/cloud_provider.go#L98-L149) interface. 

The `kwok` provider can be used by setting `--cloudprovider=kwok` in the CA flags. 

### Function implementations
* `BuildKwokCloudProvider`: 
   * Create a lister to watch for nodes with the annotation `kwok.x-k8s.io/node: fake`
      * The annotation is used to identify nodes created by the `kwok` provider (`kwok` provider creates nodes with this annotation). Limiting the lister `kwok` created nodes ensure we don't touch actual nodes. 
   * Use `KWOK_TEMPLATES_PATH` env variable to fetch user defined template nodes. This is so that the users can specify their own custom template nodes. This can be the yaml result of `kubectl get nodes` from a different cluster with some changes. Or, it can be nodes defined by the user from scratch in yaml. The yaml has to be mounted in the CA pod at the path specified in `KWOK_TEMPLATES_PATH` env variable   
      * Build `nodegroups` based on the annotations specified in the template nodes
         * Following annotations will be supported on the template node:
           * `cluster-autoscaler.kwok.nodegroup/name`: specifies name of the nodegroup this template node belongs to (defaults to a random name if not specified)
           * `cluster-autoscaler.kwok.nodegroup/min-count`: specifies min node count of the nodegroup (defaults to 0 if not specified)
           * `cluster-autoscaler.kwok.nodegroup/max-count`: specifies max node count of the nodegroup (defaults to 200 if not specified)
           * `cluster-autoscaler.kwok.nodegroup/desired-count`: specifies desired node count of the nodegroup (defaults to 0 if not specified)
   * If no `KWOK_TEMPLATES_PATH` env variable is specified,  
     * List all the nodes
     * Figure out `nodegroup` for each node based on the label on the node
     * Build `nodegroups` (rename `nodegroup` name and node template name so that it doesn't conflict with actual `nodegroup` and node). 
       * Example 1: you are on a `kind` cluster 
         * Listing all the nodes gives:
           * `kind-control-plane`
           * `worker-node-1`
         * No `nodegroups` annotation are found on either of the nodes
         * Structs representing `kind-control-plane` and `worker-node-1` are sanitized (`ResourceVersion`, `Generation`, `UID` and `CreationTimestamp` is cleaned up or set to zero value like `""`, `0`, `v1.Time{}`)
         * A new nodegroup `random-name-1` is created which points to the sanitized template node `kind-control-plane`
         * A new nodegroup `random-name-2` is created which points to the sanitized template node `worker-node-1`  
       * Example 2: you are on a GKE (or any other public cloud provider) cluster
         * List all the nodes gives:
           * `gke-cluster-1-default-pool-abc`
           * `gke-cluster-1-default-pool-xyz`
           * `gke-cluster-1-pool-1-lmn`
         * Look for `cloud.google.com/gke-nodepool` annotation to get the nodegroup name. In this case `gke-cluster-1-default-pool-abc` and `gke-cluster-1-default-pool-xyz` belong to the nodegroup `default-pool`. `gke-cluster-1-pool-1-lmn` belongs to the nodegroup `pool-1` 
         * Since these nodegroups actually exist in GCP, create a new nodegroup called `default-pool-kwok` and use the first node `gke-cluster-1-default-pool-abc` as the template node (after sanitizing). Create another nodegroup called `pool-1-kwok` and use `gke-cluster-1-pool-1-lmn` as the template node (after sanitizing)
         * Set min, max and desired size to 0, 200 and 0 respectively

* `(nodeGroup *NodeGroup) IncreaseSize`: create nodes in the cluster based on template node for the nodegroup (every node created will have the name in the format `<nodegroup-name>-<5-letter-random-string>`) using kubernetes client. 
* `(nodeGroup *NodeGroup) DeleteNodes`: delete nodes if the node has `kwok.x-k8s.io/node: fake` annotation (return error otherwise)

You can see a PoC implementation of rest of the functions at https://github.com/kubernetes/autoscaler/pull/5820/files#diff-44474ffb56eda61e9a6b16c3ca66461cdf8e02a7e89b659f5a45ca32f5fa8588 (only supports template nodes passed via env variable as of now)

### Things to note
1. Once the user is done with using the `kwok` provider, `--cloudprovider` flag can be reset to any other cloud provider. Deleting the CA pod will cleanup all the fake nodes created by the `kwok` provider and restore the cluster to its original state. This behavior doesn't seem to be supported as of writing this (based on my tests). 
2. `kwok` controller needs to be deployed alongwith CA pod for `--cloudprovider=kwok` to work correctly. This controller is responsible for making the controlplane think the fake nodes are actual nodes. Official way of deploying `kwok` controller as of writing this is to [use `kustomize`](https://github.com/kubernetes-sigs/kwok/tree/main/kustomize/kwok). We can either 
  * do an upstream contribution to convert the kustomize templates to a helm chart and use it as a dependency in our helm chart
  * or use the output of kustomize and make it a part of our cluster-autoscaler chart
3. `kwok` provider as of now only understands GKE and EKS node labels. The plan is to expand this support for other cloud providers as we go. 
### How does it look in action?
You can check it [here](https://github.com/kubernetes/autoscaler/issues/5769#issuecomment-1590541506).

### Definition of `DONE`
* `kwok` provider is implemented
* relevant unit tests have been added
* `kwok` cloud provider is `demo`ed in one of the sig-autoscaling meetings

### FAQ
1. **Will this be patched back to older releases of Kubernetes?**

    As of writing this, the plan is to release it as a part of Kubernetes 1.28 and patch it back to 1.27 and 1.26. 
2. **Why did we not use GRPC or cluster-api provider to implement this?**  
The idea was to enable users/contributors to be able to scale-test issues around different cloud providers (e.g., https://github.com/kubernetes/autoscaler/issues/5769). Implementing the `kwok` provider in-tree means we are closer to the actual implementation of our most-used cloud providers (adding gRPC communication in between would mean an extra delay which is not there in our in-tree cloud providers). Although only in-tree provider is a part of this proposal, overall plan is to:
    * Implement in-tree provider to cover most of the common use-cases
    * Implement `kwok` provider for `clusterapi` provider so that we can provision `kwok` nodes using `clusterapi` provider ([someone is already working on this](https://kubernetes.slack.com/archives/C8TSNPY4T/p1685648610609449))
    * Implement gRPC provider if there is user demand
3. **How performant is `kwok` provider really compared to `kubemark` provider?**  
`kubemark` provider seems to need 1 core per 8-10 nodes (based on our [last scale tests](https://github.com/kubernetes/autoscaler/blob/master/cluster-autoscaler/proposals/scalability_tests.md#test-setup)). This means we need roughly 10 cores to simulate 100 nodes in `kubemark`.
`kwok` provider can simulate 385 nodes for 122m of CPU and 521Mi of memory. This means, CPU wise `kwok` can simulate 385 / 0.122 =~ 3155 nodes per 1 core of CPU.  
![](images/kwok-provider-grafana.png)
![](images/kwok-provider-in-action.png)
4. **Can I think of `kwok` as a dry-run for my actual `cloudprovider`?**  
That is the goal but note that the definition of what exactly `dry-run` means is not very clear and can mean different things for different users. You can think of it as something similar to a `dry-run`.