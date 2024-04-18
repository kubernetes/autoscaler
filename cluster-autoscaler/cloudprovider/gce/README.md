# Cluster Autoscaler on GCP

The cluster autoscaler on GCP dynamically scales Kubernetes worker nodes. It runs as a deployment in your cluster.

## Auto-Discovery Setup

To run a cluster-autoscaler which auto-discovers instance groups, use the `--node-group-auto-discovery` flag. There are 2 auto-discovery options to choose from.

> NOTE - Only one of the 2 options can be used when configuring the `--node-group-auto-discovery` flag for cluster-autoscaler.    
 
### Auto-Discovery by Labels 

For example, `--node-group-auto-discovery=label:cluster-autoscaler-enabled=true,cluster-autoscaler-name=<YOUR CLUSTER NAME>` will find all the instance groups with instance templates that are tagged with those labels containing those values.

---
**NOTE**
* It is recommended to use a second tag like `cluster-autoscaler-name=<YOUR CLUSTER NAME>` when `cluster-autoscaler-enabled=true` is used across many clusters to prevent Instance Groups from different clusters recognized as the node groups
* There are no `--nodes` flags passed to cluster-autoscaler because the node groups are automatically discovered by tags
* No `min/max` values are provided when using this option. cluster-autoscaler will detect the "min" and "max" labels on the Instane Group resource in GCP, adjusting the desired number of nodes within these limits.
* If there are no `min/max` labels on the Instance Group resource, cluster-autoscaler will use the default min/max values of 0 and 1000 respectively.
---

### Auto-Discovery by NamePrefix

For example, `--node-group-auto-discovery=mig:namePrefix=test-lemon-peel-mp,min=2,max=10` will internally use a Regular Expression to find all the instance groups whose name begins with `test-lemon-peel-mp` and set the minimum and maximum number of nodes to 2 and 10 respectively. 

---
**NOTE**
* `Min` and `Max` key/value pairs where `max > min` must be specified when using this option and will not use any defaults. 
* To add more than one instance groups that do not share the same name prefix, use the `--node-group-auto-discovery` flag multiple times. Ex:
```
--node-group-auto-discovery=mig:namePrefix=test-lemon-peel-mp,min=2,max=10
--node-group-auto-discovery=mig:namePrefix=confab-nodes,min=2,max=10
```
* Clearly, the name-prefixes must be statically configured before the initialization of the cluster-autoscaler container which makes this option less flexible.
---