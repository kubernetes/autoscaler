<!--- For help refer to https://github.com/kubernetes/kubernetes/blob/master/CHANGELOG/CHANGELOG-1.20.md?plain=1 as example --->

- [v1.23.0](#v1230)
    - [Synced with which upstream CA](#synced-with-which-upstream-ca)
    - [Changes made](#changes-made)
        - [To FAQ](#to-faq)
        - [During merging](#during-merging)
        - [During vendoring k8s](#during-vendoring-k8s)
        - [Others](#others)


# v1.23.0


## Synced with which upstream CA

[v1.23.0](https://github.com/kubernetes/autoscaler/tree/cluster-autoscaler-1.23.0/cluster-autoscaler)

## Changes made

### To FAQ

- FAQ added in how to section - how can I enable Cluster Autoscaler to scale up when Node's max volume count is exceeded (CSI migration enabled)?
- price expander works for Packet as well
- multiple expanders may be passed to the expander flag as a comma separated list
- feature gates flag added


### During merging
- In `cluster-autoscaler/core/utils/utils.go`
    - GetNodeInfosForGroups method is removed from (It finds NodeInfos for all node groups used to manage the given nodes. It also returns a node group to sample node mapping.)
    - Some methods were made exportable
  
- In `cluster-autoscaler/core/utils/utils_test.go`
    - removed test cases for the method removed in `utils.go`
  
- In `cluster-autoscaler/processors/processor.go`
    - two new interfaces added
      - TemplateNodeInfoProvider (used to create the initial nodeInfos set)
      - ActionableClusterProcessor (defining whether the cluster is in an actionable state).

### During vendoring k8s
- mcm v0.44.1 -> v0.45.0
- mcm-provider-aws v0.10.0 -> v0.11.0
- mcm-provider-azure v0.6.0 -> v0.7.0
- matching the `require` section in `cluster-autoscaler/go.mod` with the one in upstream v1.23.0
- adding a replace tag —`replace github.com/Azure/azure-sdk-for-go => github.com/Azure/azure-sdk-for-go v55.8.0+incompatible` in `cluster-autoscaler/go.mod`

### Others
- [Release matrix](../README.md#releases-gardenerautoscaler) of Gardener Autoscaler updated.
