## Using CSINode allocatable limits for scaling nodes that are required for pods that use volumes

## Introduction
Currently cluster-autoscaler doesn’t take into account, volume-attach limit that a node may have when scaling nodes to support unschedulable pods.

This leads to bunch of problems:
- If there are unschedulable pods that require more volume than one supported by newly created nodes, there will still be unschedulable pods left.

- Since a node does not come up with a CSI driver typically, usually too many pods get scheduled on a node, which may not be supportable by the node in the first place. This leads to bunch of pods, just stuck.

## Implementation

Unfortunately, this can’t be fixed in cluster-autoscaler alone. We are proposing changes in both kubernetes/kubernetes and cluster-autoscaler.

## Kubernetes Scheduler change

The first change we propose in Kubernetes scheduler is to not schedule pods that require CSI volumes, if given node is not reporting any installed CSI drivers.

The proposed change is small and a draft PR is available here - https://github.com/kubernetes/kubernetes/pull/130702

This will stop too many pods crowding a node, when a new node is spun up and node is not yet reporting volume limits.

But this alone is not enough to fix the underlying problem. In part-2, cluster-autoscaler must be fixed so as it is aware of attach limits of a node via CSINode object.

## Cluster Autoscaler changes

We can split the implementation in cluster-autoscaler in two parts:
- Scaling a node-group that already has one or more nodes.
- Scaling a node-group that doesn’t have one or more nodes (Scaling from zero).

### Scaling a node-group that already has one or more nodes.

1. We propose a similar label as GPULabel added to the node that is supposed to come up with a CSI driver. This would ensure that, nodes which are supposed to have a certain CSI driver installed aren’t considered ready - https://github.com/kubernetes/autoscaler/blob/master/cluster-autoscaler/core/static_autoscaler.go#L979 until CSI driver is installed there.

However, we also propose that a node will be considered ready as soon as corresponding CSI driver is being reported as installed via corresponding CSINode object.

A node which is ready  but does not have CSI driver installed within certain time limit will be considered as NotReady and removed from the cluster.


2. We propose that, we add volume limits and installed CSI driver information to framework.NodeInfo objects. So -

```
type NodeInfo struct {
....
....
csiDrivers map[string]*DriverAllocatable
..
}

type DriverAllocatable struct {
    Allocatable int32
}
```

3. We propose that, when saving `ClusterState` , we capture and add `csiDrivers` information to all existing nodes.

4. We propose that, when getting nodeInfosForGroups , the return nodeInfo map also contains csidriver information, which can be used later on for scheduling decisions.

```
nodeInfosForGroups, autoscalerError := a.processors.TemplateNodeInfoProvider.Process(autoscalingContext, readyNodes, daemonsets, a.taintConfig, currentTime)
```

Please note that, we will have to handle the case of scaling from 0, separately from
scaling from 1, because in former case - no CSI volume limit information will be available
If no node exists in a NodeGroup.

5. We propose that, when deciding pods that should be considered for scaling nodes in podListProcessor.Process function, we update the hinting_simulator to consider CSI volume limits of existing nodes. This will allow cluster-autoscaler to exactly know, if all unschedulable pods will fit in the recently spun or currently running nodes.

Making aforementioned changes should allow us to handle scaling of nodes from 1.

### Scaling from zero

Scaling from zero should work similar to scaling from 1, but the main problem is - we do not have NodeInfo which can tell us what would be the CSI attach limit on the node which is being spun up in a NodeGroup.

We propose that we introduce similar annotation as CPU, Memory resources in cluster-api to process attach limits available on a node.

We have to introduce similar mechanism in various cloudproviders which return Template objects to incorporate volume limits. This will allow us to handle the case of scaling from zero.

## Alternatives Considered

Certain Kubernetes vendors taint the node when a new node is created and CSI driver has logic to remove the taint when CSI driver starts on the node.
- https://github.com/kubernetes-sigs/azuredisk-csi-driver/pull/2309
