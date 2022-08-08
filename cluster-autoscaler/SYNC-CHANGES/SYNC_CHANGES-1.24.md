<!--- For help refer to https://github.com/kubernetes/kubernetes/blob/master/CHANGELOG/CHANGELOG-1.20.md?plain=1 as example --->

- [v1.24.0](#v1240)
    - [Synced with which upstream CA](#synced-with-which-upstream-ca)
    - [Changes made](#changes-made)
        - [To FAQ](#to-faq)
        - [During merging](#during-merging)
        - [During vendoring k8s](#during-vendoring-k8s)
        - [Others](#others)


# v1.24.0


## Synced with which upstream CA

[v1.24.0](https://github.com/kubernetes/autoscaler/tree/cluster-autoscaler-1.24.0/cluster-autoscaler)

## Changes made

### To FAQ

- FAQ added in how to section - How can I prevent Cluster Autoscaler from scaling down non-empty nodes?
- FAQ added in the troubleshooting section -  My cluster is below minimum / above maximum number of nodes, but CA did not fix that! Why?
- FAQ added in the developer section - What go version should be used to compile CA?
- yaml file changes in the FAQ - configuring overprovisioning with CA.
- price expander works for Equinix metal and not Packet.
- some addition to the answer for FAQ - CA used to work yesterday but not today, why?


### During merging
- In `charts/cluster-autoscaler`:
    - some rules added for the following resources in `templates/role.yaml`:\
      a. apiGroup `cluster.x-k8s.io` -> machinedeployments, machines, machinesets, machinedeployments/scale\
      b. apiGroup `coordination.k8s.io` -> leases, cluster-autoscaler\
      namespace is also added for the role.
    - `README.md` updated with config details for clusterapi cloudProvider along with some additions to podSecurityPolicy values
    - `values.yaml` updated with details for clusterapi cloudProvider and topologySpreadConstraints.

- introduced grpcExpander
- `max-pod-eviction-time` flag added (sets the maximum time CA tries to evict a pod before giving up)
- `node-info-cache-expire-time` flag added (Node Info cache expire time for each item. Default value is 10 years.)
- support added for debugging snapshot
- `cluster-autoscaler/util/utils.go` updated to ignore `podspec.Hostname` when comparing two podSpecs to be semantically equal.

- In `cluster-autoscaler/processor`:
    - `nodeinfosprovider` —> A `ttl` field for nodeInfoCache is introduced. Not caching nodeInfo for recently ready nodes(stabilisation delay of 1 min)
    - `processor.go` —> addition of a new processor - ScaleDownSetProcessor(used to make final selection of nodes to scale-down.)

- In `cluster-autoscaler/core`:
    - `cluster-autoscaler/core/filter_out_schedulable.go` —> removed callback to `DisableScaledownForLoop` to allow scaling down in cluster while some pods are waiting for booting up of nodes.
    - `cluster-autoscaler/core/static-autoscaler.go` —> continue CA loop when unregistered nodes are removed. (Return statement removed)
    - `equivalence_groups.go` —> constant `maxEquivalenceGroupsByController` introduced (value is 10)
    - `utils/pod_schedulable.go` —> `maxPodsPerOwnerRef` = 10, PodSchedulableMap is a struct with `OverflowingController` as a new field (stores the number of controllers that had too many different pods to be effectively cached in the form of a map).
      This and the previous point were introduced to limit caching pods per owner reference.

### During vendoring k8s
- mcm v0.45.0 -> 0.46.0
- mcm-provider-aws v0.11.0 -> v0.12.0
- mcm-provider-azure v0.7.0 -> v0.8.0
- replace github.com/Azure/azure-sdk-for-go => github.com/Azure/azure-sdk-for-go v55.8.0+incompatible

### Others
- [Release matrix](../README.md#releases-gardenerautoscaler) of Gardener Autoscaler updated.
