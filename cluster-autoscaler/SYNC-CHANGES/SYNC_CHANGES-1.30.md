<!--- For help refer to https://github.com/kubernetes/kubernetes/blob/master/CHANGELOG/CHANGELOG-1.20.md?plain=1 as example --->

- [v1.30.1](#v1290)
    - [Synced with which upstream CA](#synced-with-which-upstream-ca)
    - [Changes made](#changes-made)
        - [During merging](#during-merging)
        - [During vendoring k8s](#during-vendoring-k8s)
        - [Others](#others)


# v1.30.0


## Synced with which upstream CA

[v1.30.1](https://github.com/kubernetes/autoscaler/releases/tag/cluster-autoscaler-1.30.1)

## Changes made
- See general release notes of 1.30.1: https://github.com/kubernetes/autoscaler/releases/tag/cluster-autoscaler-1.30.1
- New flag added in autoscaling options: `flag.Bool("enable-provisioning-requests", false, "Whether the clusterautoscaler will be handling the ProvisioningRequest CRs.")`.
- New flag added in autoscaling options: `flag.Bool("frequent-loops-enabled", false, "Whether clusterautoscaler triggers new iterations more frequently when it's needed")`.
- New interface method `AtomicIncreaseSize` implemented in mcm cloud provider to satisfy interface : `cloudprovider.NodeGroup`.

### During merging
- import package for `machine-controller-manager-provider-aws` was updated after vendoring latest version.
- import package for `machine-controller-manager-provider-azure` was updated after vendoring latest version.

### During vendoring k8s
- mcm-provider-aws  v0.19.2 -> v0.20.0
- mcm-provider-azure v0.12.1 -> v0.13.0

### Others
- [Release matrix](../README.md#releases-gardenerautoscaler) of Gardener Autoscaler updated.