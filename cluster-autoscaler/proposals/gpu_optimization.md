# Cluster Autoscaling Optimization for GPU clusters
##### Author: Jeffwan

## Introduction
Cluster Autoscaler makes it extremely easy to scale kubernetes cluster in response to pod status. At the same time, having it automate Kubernetes cluster to work efficiently.

However, there's still challenges in specific scenarios for accelerator computing clusters. Accelerator computing instances are quite different from normal instance. They are powerful, scalable instances that provide GPU-based parallel computing capabilities. 

Accelerator computing nodes are expensive. (ec2 p3.16xlarge on demand instance cost $24.48/hr). Efficiently scaling the cluster will save a substtantial amount of money. GPU nodes are well suited for tasks that have heavy computation needs like machine learning and HPC applications. They can not be interrupted in some cases and that brings us more challenges in scale down. 

Here's the problems I find in current upstream cluster autoscaler. 

### Problems in cluster autoscaler for GPU node

#### Accelerator Label 
GKE uses gpu label `cloud.google.com/gke-accelerator` to inspect if a node is an accelerator node, other cloud providers don't define their label yet. In order for CA to behavior normally in GPU cluster, it's essential to have this gpu label.

#### Scale too much without GPU label support
Pending pods will trigger scale up decision to bring up number of nodes in one node group. When new node is added, CA knows it is `upcoming` and it knows some pods will go on that node once it boots up. So it won't trigger another scale-up for those pods. Device plugin triggers after the node becomes ready and the extra resource required by pod will show up in node spec even later. CA will see that it was wrong and the pods for which it triggered scale-up can't actually go to a node it added for them (because it doesn't have the special resource yet and CA doesn't know it will be added later). So CA will go and create more nodes.

The problem is GPU node becomes ready first but at that time, there's no allocable GPU resources. The root reason could be  

* Device Plugin is not ready.
* Device Plugin is ready but has not sent list of devices to kubelet. 
* kubelet takes time to advertise resources to API server so that node spec is not up to date.

Pod that requests GPU can not be scheduled during this period. In order to address this issue, we either need to work with Nvidia folks on device plugin to change behavior on upstream, or make CA mark nodes with GPU label and unallocated GPU resource as NonReady.


#### GPU Resource is not given consideration in scaling down
CA filter candidates with low node cpu and memory utilization rate in every loop. 

* If utilization rate is high, even all GPUs are idle, scale-down will not be triggered.  
Consider use case for machine learning inference or some training case that could tolerant failures, we actually like to give a try to see if all workloads on that particular node can be moved to other nodes. Then user can scale down this GPU node and efficiently reduce their cost.

* If utilization rate is low, GPUs are in use, scale-down will be triggered.  
If there's a distributed training task on that node, killing task will lead to entire training job failing. In this case, this node can not be a scale down candicate.

```
// CalculateUtilization calculates utilization of a node, defined as maximum of (cpu, memory) utilization.
// Per resource utilization is the sum of requests for it divided by allocatable. It also returns the individual
// cpu and memory utilization.
func CalculateUtilization(node *apiv1.Node, nodeInfo *schedulercache.NodeInfo, skipDaemonSetPods, skipMirrorPods bool) (utilInfo UtilizationInfo, err error) {
    cpu, err := calculateUtilizationOfResource(node, nodeInfo, apiv1.ResourceCPU, skipDaemonSetPods, skipMirrorPods)
    if err != nil {
        return UtilizationInfo{}, err
    }
    mem, err := calculateUtilizationOfResource(node, nodeInfo, apiv1.ResourceMemory, skipDaemonSetPods, skipMirrorPods)
    if err != nil {
        return UtilizationInfo{}, err
    }
    return UtilizationInfo{CpuUtil: cpu, MemUtil: mem, Utilization: math.Max(cpu, mem)}, nil
}

```
> node utilization calculation logic


### Proposed Solution
1. Either move `utils/gpu` logic to cloud provider or have a new option passed from commandline to indicate gpu node label cloud provider like to use
2. Scale down case is kind of tricky because training and serving seems like different use case and have confliction. Fit GPU resource into utilization formula doesn't solve this issue. Instead, I think it's better to have a flag to indicate if gpu nodes can be scaled down or not.

Here's the pseudocode of updated scale down logic.
```
// IdleGPUNodeScaledDownEnabled - client pass this option into CA to control behavior

// calculate node utilization rate

if (IdleGPUNodeScaledDownEnabled) {    
    // resuse existing logic 

    if (node.label[GPULabel] && node.utilization < threhold) {
        // Try to move pods to other nodes if possible
    }

} else {
    // resuse existing logic 

    if (node.label[GPULabel] && isGPUInUse(node)) {
        // remove from scale down list if exists.
    }
}

```

### Related Issues
#1367
#1135
