<!--- For help refer to https://github.com/kubernetes/kubernetes/blob/master/CHANGELOG/CHANGELOG-1.20.md?plain=1 as example --->

- [v1.32.0](#v1320)
  - [Synced with which upstream CA](#synced-with-which-upstream-ca)
  - [Changes made](#changes-made)
    - [During vendoring k8s](#during-vendoring-k8s)
    - [Others](#others)

# v1.32.0

## Synced with which upstream CA
[v1.32.0](https://github.com/kubernetes/autoscaler/tree/cluster-autoscaler-1.32.0/cluster-autoscaler)

## Changes made
- See general release notes of 1.32.0: https://github.com/kubernetes/autoscaler/releases/tag/cluster-autoscaler-1.32.0
- Vendored modules in `vertical-pod-autoscaler` removed.
- New interface method `ForceDeleteNodes` implemented in mcm cloud provider to satisfy interface : `cloudprovider.NodeGroup`.
- Corresponding to the same, a `--force-delete-long-unregistered-nodes` flag was added. It allows CA to remove long unregistered nodes even if it would break the min size constraints of their node group.
- Changes to flags with regards to removal of legacy scale-down code:
  + `--parallel-drain` flag was removed. To ensure only one node is drained at the same time, use `--max-drain-parallelism=1`.
  + `--max-empty-bulk-delete` flag was deprecated. It will be replaced by `--max-scale-down-parallelism` in a future release.
- Experimental support for DRA autoscaling is implemented, disabled by default. Can be enabled via `--enable-dynamic-resource-allocation` flag. [Reference KEP](https://github.com/kubernetes/enhancements/blob/9de7f62e16fc5c1ea3bd40689487c9edc7fa5057/keps/sig-node/4381-dra-structured-parameters/README.md#summary)
- Ability to set custom lease resource name using the `--lease-resource-name` flag provided.
- Allows CheckCapacity ProvisioningRequests to be processed in batch mode with configurable max batch size and batch timebox, controlled via new flags:
  `--check-capacity-batch-processing`, `--check-capacity-provisioning-request-max-batch-size`, `--check-capacity-provisioning-request-batch-timebox`

### During vendoring k8s
- mcm v0.55.0 -> 0.57.0
- mcm-provider-aws v0.20.0 -> 0.23.0
- mcm-provider-azure v0.13.0 -> 0.15.1
- k8s.io/api v0.31.1 -> v0.32.0
- github.com/onsi/ginkgo/v2 v2.19.0 -> v2.21.0

### Others
- [Release matrix](../README.md#releases-gardenerautoscaler) of Gardener Autoscaler updated.