# Node Auto-provisioning 
author: mwielgus

# Introduction

Node auto-provisioning (abbreviated as NAP) is a multi-step effort to free Kubernetes users from worrying about how many nodes and in what size should they have in their clusters. So far we have completed two steps:

* Cluster Autoscaler is able to scale a node group to/from 0: [Doc](./min_at_zero_gcp.md)
* Cluster Autoscaler picks reasonable node group when scaling up: [Doc](./pricing.md)

Once these two steps are completed we can focus on providing options for scale up that were not originally created during
the configuration process.

# Changes
Allowing CA to create node pools at will requires multiple changes in various places in CA and around it.

## Configuration 

Previously, when the node groups were fixed, the user had to just specify the minimum and maximum size of each of the node groups to keep the cluster within the allowed budget. With NAP it is becoming more complex. GCP machine prices (and other cloud provides as well) linearly depend on the number of cores and gb of memory (and possibly GPUs in the future) we can just ask user to set the min/max amount for each of the resources. So we will add the following flags to the CA:
* `--node-autoprovisioning` - enables node autoprovisioning in CA
* `--min-cpu` - min number of cpus in the cluster
* `--max-cpu` - max number of cpus in the cluster
* `--min-memory` - min size of memory (in gigabytes) in the cluster
* `--max-memory`- max size of memory (in gigabytes) in the cluster
* `--autoprovisioning-prefix` - node group name prefix by which CA will be able to tell autoprovisioned node group from the non-auoprovisioned. This by default will be equal to “nodeautoprovisioning”

Moreover, the users might want to keep some of the node pools always in the cluster. For that purpose we will allow them to specify the node groups explicitly in the same fashion as in regular CA using:

* `--nodes=min:max:id`

It is assumed that if there is any extra node group/node pool in the cluster, that hasn’t been mentioned in the command line should stay exactly “as-is”. 

While there are two options to express the boundaries for CA operations the precedence order is as follows:

* `--min-cpu`, `--max-cpu`, `--min-memory`, `--max-memory` go first. They are not enforcing. If there is more/less resources in the cluster than desired CA will not immediately start/kill nodes. It will move only towards the expected boundaries when needed/appropriate. The difference between the expected cluster size and the current size will not grow. The flags are optional. If they are not specified then the assumption is that the limit is either 0 or +Inf.

* `--nodes=min:max:id` will come second, when applicable. Nodes in that group may go between min and max only if --min-cpu/max-cpu/min-memory/max-memory constraints are met. 

Example:

There 3 groups in the cluster:

* “main” - not autoprovisioned (no prefix), not autoscaled (not be mentioned in CA’s configuration flags). Contains 2 x n1-standard2.  
* “as” - not autprovisioned  (no prefix), autoscaled (mentioned in CA configuration flags) between 0 and 2 nodes. Contains 1 x n1-standard16.
* “nodeautoprovisioning_n1-highmem4_1234129” - autoprovisioned (has prefix). Currently contains 2 x n1-highmem4. 

* If `--max-cpu=5` then no node can be added to any of the groups. No new groups will be created.
* If `--max-cpu=32` then 1 node might be added to “nodeautoprovisioning_n1-highmem4...” or a new node group, with up to 4 n1-standard1 machines created.
* If `--max-cpu=80` then:
   * 1 node might be added to “as”
     “Nodeautoprovisioning_n1-highmem4_1234129” may grow up to 15 nodes,
   * Some other node groups might be created. 

Similar logic applies to `--min-cpu`. It might be good to set this value relatively low so that CA is able to disband unneeded machines.

To allow power users to have some control over what exactly can be The provided new methods in cloudprovider API autprovisioned there will be an semi-internal flag with a list of all machine types that CA can autoprovision: 
`--machine-types=n1-standard-1,n1-standard-2,n1-standard-4,n1-standard-8,n1-standard-16, n1-highmem-1,n1-highmem-2,...`

Also, for sanity, there will be a flag to limit the total number of node groups in a cluster, set to 50 or so.

This will allow NAP to act as an extension of CA. Node pools created by NAP will not have autoscaling flag enabled. Their min/max size will entirely depend on the global settings at the cluster level. There will be some flags added to gcloud alpha container clusters create/update command with the same semantics as in cluster autoscaler.

* `--enable-node-autoprovisioning`
* `--min-cpu`
* `--max-cpu`
* `--min-memory`
* `--max-memory`

## Cluster Autoscaler Code

Right now the scale up code assumes that all node groups are already known and stable:

https://github.com/kubernetes/autoscaler/blob/master/cluster-autoscaler/core/scale_up.go#L77

The list of current node groups to check will have to be expanded with, probably bigger, list of all potential node groups that could be added to the cluster. CA will analyze what labels and other resources are needed by the pods and calculate a set of all labels and additional resources that are useful (and commonly needed) for most of the pods. 

By default CA will not set any taints on the nodes. Tolerations, set on a pod, are not requirements. 

Then it will add these labels to all machine types available in the cloud provider and evaluate the theoretical node groups along with the node groups that are already in the cluster. If the picked node group doesn’t exist then CA should create it.

To allow this the following extensions will be made in CloudProvider interface:

 * `GetAvilableMachineTypes() ([]string, error)` returns all node types that could be requested from the cloud provider.
 * `NewNodeGroup(name string, machineType string, labels map[string]string, extraResources map[Resource]Quantity) (NodeGroup, error)` builds a theoretical node group based on the node definition provided. The node group is not automatically created on the cloud provider side. The argument list will probably be expanded with GPU specific stuff.
 * `NodeGroups` will only return created node groups. Theoretical/temporary node groups will not be included.

Moreover an extension will be made to node group interface:

 * `Exists() (bool, error)` - checks if the node group really exists on the cloud provider side. Allows to tell the theoretical node group from the real one.
 * `Create() error` - creates the node group on the cloud provider side. 
 * `Delete() error` - deletes the node group on the cloud provider side. This will be executed only for autoprovisioned node groups, once their size drops to 0.

# Calculating best label set for nodes

Assume that PS is a set of pods that would fit on a node of type NG if it had the labels matching to its selector. For each of the machine types we can build a node with no labels and for each pod set the labels according to the pod requirements. If the pod fits to the node it goes to PS. 

Then we calculate the stats of all node selectors of the pods. For each significantly different node selector we calculate the number of pods that has this specific node selector. We pick the most popular one, and then check if this selector is “compatible” with the second most popular, third (and so on) as well as the selected machine type. 

Example:

 * `S1: x = "a" and machine_type="n1-standard-2"
 * `S2: y = "b"
 * `S3: x = "c"
 * `S4: machine_type = "n1-standard-16"

S1 is compatible with S2 and S4. S2 is compatible with S1, S3 and S4. S3 is compatible with S4 and S2.

The label selector that would come from S1, S2 and S4 would be x="a" and y="b" and machine_type = "n1-standard-2", however
depending on popularity, the other option is S2, S3, S4 => x="c", y="b" and machine_type = "n1-standard-16".

# Testing 

The following e2e test scenarios will be created to check whether NAP works as expected:

 * [TC1] An unschedulable pod with big requirements is created. A big node/nodegroup, with no labels/taints, is provided.
 * [TC2] An unschedulable pod with specific label selector is created. A node/nodegroup  with labels is provided.
 * [TC3] 2 unschedulable pods with different/incompatible label selectors are created. 2 different node groups are eventually created.
 * [TC4] 2 unschedulable pods with compatible (but different) label selectors are created:
  * For alpha 2 node groups should be created
  * For beta 1 node group should be created
 * [TC5] An unschedulable pod with a custom toleration is created. A node with taints should be provided.
 * [TC6] 2 pods with different custom tolerations is created. 2 node groups are eventually created.
 * [TC7] Unneeded, autoprovisioned node group is deleted.
 * [TC8] No node group is created if there is no cpu quota.
 * [TC9] Scalability tests.
