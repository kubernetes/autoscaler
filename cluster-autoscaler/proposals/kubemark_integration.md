# Cluster Autoscaler on kubemark

## Background

In order to perform scalability testing of Cluster Autoscaler we want to be able to run tests on clusters with thousands of nodes. This is possible relatively cheaply using [kubemark](https://github.com/kubernetes/community/blob/master/contributors/design-proposals/scalability/kubemark.md). Therefore it is desirable to make Cluster Autoscaler work with kubemark clusters.

## Kubemark concepts

Kubemark setup includes two clusters. First - **external cluster** is a regular kubernetes cluster. For our setup let's assume it's running on GCE. On top of that cluster a **kubemark cluster** is run with a separate master machine on GCE and **Hollow Nodes** running as pods in the external cluster. Hollow Node pods are created and controlled via a common Replication Controller.

## Implementation

To be able to launch Cluster Autoscaler on kubemark, we need to add [Cloud Provider Interface](https://github.com/kubernetes/autoscaler/blob/master/cluster-autoscaler/cloudprovider/cloud_provider.go) implementation for kubemark in Cluster Autoscaler. Cloud Provider Interface needs Node and NodeGroup abstractions.

### Node

The Node abstraction will map to a kubemark Hollow Node. kubemark schedules nodes as pods via a Replication Controller. Since we need to have an ability to kill a specific node and a replication controller does not allow it if it hosts multiple pods, we will need to modify kubemark to be able to operate using multiple controllers, each hosting exactly one pod.

### NodeGroup

The NodeGroup abstraction will be mapped to labels on Hollow Nodes. Nodes with same value for label `autoscaling.k8s.io/nodegroup` will belong to the same NodeGroup. This will allow to keep the NodeGroups intact regardless of autoscaler restarts or crashes. Getting all nodes for a node group or checking which node belongs to which node group will be a simple query to informer.

On Cluster Autoscaler startup kubemark Cloud Provider will parse the config passed in by the user (`--nodes={MIN}:{MAX}:{NG_LABEL_VALUE}`) and check if the cluster contains at least {MIN} nodes with `autoscaling.k8s.io/nodegroup={NG_LABEL_VALUE}`. If not, it will create the nodes.

### Important operations on NodeGroup

* `Name()` - 'autoscaling.k8s.io/nodegroup' label value
* `IncreaseSize(delta int)` - creation of #delta singleton Replication Controllers in external cluster with label `'autoscaling.k8s.io/nodegroup'=Name()`
* `DeleteNodes([]*apiv1.Node)` - removal of specified Replication Controllers
* `DecreaseTargetSize(delta int) error` - removal of Replication Controllers that have not yet been created
* `TemplateNodeInfo() (*schedulerframework.NodeInfo, error)` - will return ErrNotImplemented
* `MaxSize()` - specified via config (`--nodes={MIN}:{MAX}:{NG_LABEL_VALUE}`)
* `MinSize()` - specified via config

In order to perform the Replication Controller creation and deletion we will need to talk to the Api Server of the external cluster.
